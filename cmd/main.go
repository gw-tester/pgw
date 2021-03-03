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
	"os"

	"github.com/gw-tester/pgw/internal/core/domain"
	"github.com/gw-tester/pgw/internal/core/ports"
	service "github.com/gw-tester/pgw/internal/core/services/pgwsrv"
	"github.com/gw-tester/pgw/internal/pkg/discover"
	"github.com/gw-tester/pgw/internal/pkg/utils"
	repository "github.com/gw-tester/pgw/internal/repositories/pgwrepo"
	router "github.com/gw-tester/pgw/internal/routers/pgwrouter"
	log "github.com/sirupsen/logrus"
)

func getLogLevel() log.Level {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		return log.InfoLevel
	}

	if userLogLevel, err := log.ParseLevel(logLevel); err == nil {
		return userLogLevel
	}

	return log.InfoLevel
}

func getRepository() ports.IPRepository {
	redisURL, ok := os.LookupEnv("REDIS_URL")
	if ok && redisURL != "" {
		return repository.NewRedis()
	}

	etcdURL, ok := os.LookupEnv("ETCD_URL")
	if ok && etcdURL != "" {
		return repository.NewETCD()
	}

	return repository.NewMemKVS()
}

func main() {
	log.SetLevel(getLogLevel())

	service := service.New(getRepository())

	// The discovery process requires specific order
	s5uIP := discover.GetIPFromNetwork(utils.GetEnv("S5U_NETWORK", "172.25.0.0/24"))
	s5cIP := discover.GetIPFromNetwork(utils.GetEnv("S5C_NETWORK", "172.25.1.0/24"))
	sgiLink := utils.GetEnv("SGI_NIC", "eth2")
	sgiSubnet := utils.GetEnv("SGI_SUBNET", "10.0.1.0/24")

	pgw := domain.New(s5cIP.IP.String(), s5uIP.IP.String(), sgiLink, sgiSubnet)

	if err := service.Create(pgw); err != nil {
		log.WithError(err).Panic("Failed to store P-GW information")
	}

	router := router.New(pgw)
	if router == nil {
		log.Panic("Failed to initialize P-GW service")
	}
	defer router.Close()

	router.ListenAndServe()
}
