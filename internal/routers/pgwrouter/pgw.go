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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gw-tester/pgw/internal/core/domain"
	"github.com/gw-tester/pgw/internal/handlers/commonhdl"
	"github.com/gw-tester/pgw/internal/handlers/pgwhdl"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/wmnsk/go-gtp/gtpv1"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

type router struct {
	ControlPlane controlPlane
	UserPlane    userPlane

	handlers []pgwhdl.Handler

	errorChan chan error
}

type controlPlane struct {
	Connection *gtpv2.Conn
	Address    string
}

type userPlane struct {
	Connection *gtpv1.UPlaneConn
	Address    string
}

// Router provides a server to process requests.
type Router interface {
	ListenAndServe()
	Close() error
}

func (r *router) registerHandler(messageType uint8, handler pgwhdl.Handler) {
	r.ControlPlane.Connection.AddHandler(messageType, commonhdl.NewLogger(handler.Handle).Log)
	r.handlers = append(r.handlers, handler)
}

// New initialize a router object with user and control plane connections.
func New(config *domain.Pgw) Router {
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

	userPlaneConnection := gtpv1.NewUPlaneConn(userPlaneAddr)

	router := &router{
		ControlPlane: controlPlane{
			Connection: gtpv2.NewConn(controlPlaneAddr, gtpv2.IFTypeS5S8PGWGTPC, 0),
			Address:    controlPlaneAddr.String(),
		},
		UserPlane: userPlane{
			Connection: userPlaneConnection,
			Address:    userPlaneAddr.String(),
		},
		handlers:  []pgwhdl.Handler{},
		errorChan: nil,
	}

	// register handlers for ALL the message you expect remote endpoint to send.
	router.registerHandler(message.MsgTypeCreateSessionRequest, pgwhdl.NewCreate(userPlaneConnection, config))
	router.registerHandler(message.MsgTypeDeleteSessionRequest, pgwhdl.NewDelete())

	if err := router.UserPlane.Connection.EnableKernelGTP("gtp-pgw", gtpv1.RoleGGSN); err != nil {
		log.WithError(err).Error("Enable Kernel GTP error")

		return nil
	}

	return router
}

// ListenAndServe initiates user and control plane connections and waits for incomming requests.
func (r *router) ListenAndServe() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fatalCh := make(chan error)

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
		if err := r.ControlPlane.Connection.ListenAndServe(ctx); err != nil {
			log.WithError(err).Warn("Control Plane Listen and Serve error")

			return
		}

		log.Warn("Control Plane Connection ListenAndServe method exitted")
	}()
	log.WithFields(log.Fields{
		"S5-C": r.ControlPlane.Address,
	}).Info("Started serving S5-C")

	go func() {
		if err := r.UserPlane.Connection.ListenAndServe(ctx); err != nil {
			log.WithError(err).Warn("User Plane Listen and Serve error")

			return
		}

		log.Warn("User Plane Connection ListenAndServe method exitted")
	}()
	log.WithFields(log.Fields{
		"S5-U": r.UserPlane.Address,
	}).Info("Started serving S5-U")
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
