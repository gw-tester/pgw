/*
Copyright 2021
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pgwrouter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/InVisionApp/go-health/v2"
	"github.com/InVisionApp/go-health/v2/handlers"
	"github.com/gw-tester/pgw/internal/core/domain"
	"github.com/gw-tester/pgw/internal/handlers/counterhdl"
	"github.com/gw-tester/pgw/internal/handlers/loggerhdl"
	"github.com/gw-tester/pgw/internal/handlers/pgwhdl"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/wmnsk/go-gtp/gtpv1"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

type router struct {
	mutex           sync.Mutex
	ControlPlane    controlPlane
	UserPlane       userPlane
	ManagementPlane managementPlane

	sessionsProcessed prometheus.Counter
	handlers          []pgwhdl.Handler

	errorChan chan error
}

type controlPlane struct {
	Connection *gtpv2.Conn
	Address    string
	isReady    bool
}

type userPlane struct {
	Connection *gtpv1.UPlaneConn
	Address    string
	isReady    bool
}

type managementPlane struct {
	health *health.Health
}

// ErrPlaneNotReady indicates that user and/or control plane services are not ready yet.
var ErrPlaneNotReady = errors.New("not ready")

// Router provides a server to process requests.
type Router interface {
	ListenAndServe()
	Close() error
}

func (r *router) registerHandlers(config *domain.Pgw) {
	createHdl := pgwhdl.NewCreate(r.UserPlane.Connection, config)
	r.handlers = append(r.handlers, createHdl)

	r.ControlPlane.Connection.AddHandler(message.MsgTypeCreateSessionRequest, loggerhdl.Wrap(counterhdl.Wrap(
		createHdl.Handle, r.sessionsProcessed)))

	r.ControlPlane.Connection.AddHandler(message.MsgTypeDeleteSessionRequest, loggerhdl.Wrap(
		pgwhdl.Handle))

	http.HandleFunc("/healthcheck", handlers.NewJSONHandlerFunc(r.ManagementPlane.health, nil))
	http.Handle("/metrics", promhttp.Handler())
}

// New initialize a router object with user and control plane connections.
func New(config *domain.Pgw, h *health.Health) Router {
	if err := config.Validate(); err != nil {
		log.WithError(err).Error("Invalid PGW domain object")

		return nil
	}

	controlPlaneAddr, err := config.ControlPlane.GetAddress()
	if err != nil {
		log.WithError(err).Error("Control Plane get address error")

		return nil
	}

	userPlaneAddr, err := config.UserPlane.GetAddress()
	if err != nil {
		log.WithError(err).Error("User Plane get address error")

		return nil
	}

	router := &router{
		ControlPlane: controlPlane{
			Connection: gtpv2.NewConn(controlPlaneAddr, gtpv2.IFTypeS5S8PGWGTPC, 0),
			Address:    controlPlaneAddr.String(),
		},
		UserPlane: userPlane{
			Connection: gtpv1.NewUPlaneConn(userPlaneAddr),
			Address:    userPlaneAddr.String(),
		},
		ManagementPlane: managementPlane{
			health: h,
		},
		sessionsProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "sessions_created_total",
			Help: "Create Session Request",
		}),
		handlers:  []pgwhdl.Handler{},
		errorChan: nil,
	}

	if err := h.AddChecks([]*health.Config{
		{
			Name:     "main-check",
			Checker:  router,
			Interval: time.Duration(2) * time.Second,
			Fatal:    true,
		},
	}); err != nil {
		log.WithError(err).Warn("Add main check error")
	}

	router.registerHandlers(config)

	return router
}

// ListenAndServe initiates user and control plane connections and waits for incomming requests.
func (r *router) ListenAndServe() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fatalCh := make(chan error)

	if err := r.UserPlane.Connection.EnableKernelGTP("gtp-pgw", gtpv1.RoleGGSN); err != nil {
		log.WithError(err).Error("Enable Kernel GTP error")

		return
	}

	go func() {
		if err := r.run(ctx); err != nil {
			fatalCh <- err
		}
	}()

	for {
		select {
		case sig := <-sigCh:
			log.Println(sig)

			return
		case err := <-r.errorChan:
			log.WithError(err).Warn("Router channel error")
		case err := <-fatalCh:
			log.WithError(err).Fatal("Fatal channel error")

			return
		}
	}
}

func (r *router) run(ctx context.Context) error {
	go func() {
		r.ControlPlane.isReady = true
		if err := r.ControlPlane.Connection.ListenAndServe(ctx); err != nil {
			log.WithError(err).Warn("Control Plane Listen and Serve error")

			r.ControlPlane.isReady = false

			return
		}

		r.ControlPlane.isReady = false
	}()
	log.WithFields(log.Fields{
		"S5-C": r.ControlPlane.Address,
	}).Info("Started serving S5-C")

	go func() {
		r.UserPlane.isReady = true
		if err := r.UserPlane.Connection.ListenAndServe(ctx); err != nil {
			log.WithError(err).Warn("User Plane Listen and Serve error")

			r.UserPlane.isReady = false

			return
		}

		r.UserPlane.isReady = false
	}()
	log.WithFields(log.Fields{
		"S5-U": r.UserPlane.Address,
	}).Info("Started serving S5-U")

	go func() {
		if err := r.ManagementPlane.health.Start(); err != nil {
			log.WithError(err).Warn("Unable to start healthcheck")
		}

		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.WithError(err).Warn("Management Plane Listen and Serve error")

			return
		}

		log.Warn("Management Plane Connection ListenAndServe method exitted")
	}()

	fmt.Println("P-GW server has started") //nolint:forbidigo

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-r.errorChan:
			log.WithError(err).Warn("PGW router raised an error")

			return err
		}
	}
}

// Close removes rules and routes added by the Router and closes user plane connnection.
func (r *router) Close() error {
	for _, handler := range r.handlers {
		if err := handler.Close(); err != nil {
			log.WithError(err).Warn("Close Handler error")
		}
	}

	if r.UserPlane.Connection != nil {
		if err := netlink.LinkDel(r.UserPlane.Connection.KernelGTP.Link); err != nil {
			log.WithError(err).Warn("Kernel GTP Link Deletion error")
		}

		if err := r.UserPlane.Connection.Close(); err != nil {
			log.WithError(err).Warn("Close User Plane Connection error")
		}
	}

	close(r.errorChan)

	return nil
}

func (r *router) Status() (interface{}, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	response := map[string]bool{"ControlPlane": r.ControlPlane.isReady, "UserPlane": r.UserPlane.isReady}

	if r.ControlPlane.isReady && r.UserPlane.isReady {
		return response, nil
	}

	return response, ErrPlaneNotReady
}
