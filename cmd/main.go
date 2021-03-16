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

package main

import (
	"time"

	"github.com/InVisionApp/go-health/v2"
	arg "github.com/alexflint/go-arg"
	"github.com/gw-tester/ip-discover/pkg/discover"
	"github.com/gw-tester/pgw/internal/core/domain"
	"github.com/gw-tester/pgw/internal/core/ports"
	service "github.com/gw-tester/pgw/internal/core/services/pgwsrv"
	repository "github.com/gw-tester/pgw/internal/repositories/pgwrepo"
	router "github.com/gw-tester/pgw/internal/routers/pgwrouter"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type args struct {
	Log           logLevel `arg:"env:LOG_LEVEL" default:"info" help:"Defines the level of logging for this program."`
	RedisURL      string   `arg:"env:REDIS_URL" help:"Specifies the Redis URL connection string."`
	RedisPassword string   `arg:"env:REDIS_PASSWORD" help:"Specifies the Redis user password."`
	EtcdURL       string   `arg:"env:ETCD_URL" help:"Specifies the ETCD URL connection string."`
	S5uNetwork    string   `arg:"env:S5U_NETWORK,required" help:"Defines the S5 User plane network."`
	S5cNetwork    string   `arg:"env:S5C_NETWORK,required" help:"Defines the S5 Control plane network."`
	SgiNic        string   `arg:"env:SGI_NIC,required" help:"Defines the SGi network interface."`
	SgiSubnet     string   `arg:"env:SGI_SUBNET,required" help:"Defines the SGi subnet."`
}

type logLevel struct {
	Level log.Level
}

func (n *logLevel) UnmarshalText(b []byte) error {
	s := string(b)

	logLevel, err := log.ParseLevel(s)
	if err != nil {
		return errors.Wrap(err, "failed to parse the log level")
	}

	n.Level = logLevel

	return nil
}

func getRepository(a args) ports.IPRepository {
	if a.RedisURL != "" {
		return repository.NewRedis(a.RedisURL, a.RedisPassword)
	}

	if a.EtcdURL != "" {
		return repository.NewETCD(a.EtcdURL)
	}

	return repository.NewMemKVS()
}

func (args) Version() string {
	return "pgw 0.0.3"
}

func (args) Description() string {
	return "this program provides PDN Gateway functionality."
}

func main() {
	var args args

	arg.MustParse(&args)
	log.SetLevel(args.Log.Level)
	repository := getRepository(args)
	service := service.New(getRepository(args))

	// The discovery process requires specific order
	s5uIP, err := discover.GetIPFromNetwork(args.S5uNetwork)
	if err != nil {
		log.WithError(err).Panic("Failed to discovery first IP address of S5-U network")
	}

	s5cIP, err := discover.GetIPFromNetwork(args.S5cNetwork)
	if err != nil {
		log.WithError(err).Panic("Failed to discovery first IP address of S5-C network")
	}

	pgw := domain.New(s5cIP.IP.String(), s5uIP.IP.String(), args.SgiNic, args.SgiSubnet)

	if err := service.Create(pgw); err != nil {
		log.WithError(err).Panic("Failed to store P-GW information")
	}
	defer service.Remove()

	h := health.New()
	if err := h.AddChecks([]*health.Config{
		{
			Name:     "datastore-check",
			Checker:  repository,
			Interval: time.Duration(2) * time.Second,
			Fatal:    true,
		},
	}); err != nil {
		log.WithError(err).Warn("Add datastore check error")
	}

	router := router.New(pgw, h)
	if router == nil {
		log.Panic("Failed to initialize P-GW service")
	}
	defer router.Close()

	router.ListenAndServe()
}
