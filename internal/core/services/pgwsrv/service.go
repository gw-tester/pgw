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

package pgwsrv

import (
	"fmt"

	"github.com/gw-tester/pgw/internal/core/domain"
	"github.com/gw-tester/pgw/internal/core/ports"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	s5cIP string = "pgw_s5c_ip"
	s5uIP string = "pgw_s5u_ip"
)

// ErrSaveIP indicates a database failure during the storing IP addresses.
var ErrSaveIP = errors.New("fail to store IP Address")

// Service provides methods to create, retrieve and delete PGW instances.
type Service struct {
	ipRepository ports.IPRepository
}

// New creates PGW service instance.
func New(ipRepository ports.IPRepository) *Service {
	return &Service{
		ipRepository: ipRepository,
	}
}

// Create validates and stores an PGW instance in a given repository.
func (srv *Service) Create(pgw *domain.Pgw) error {
	if err := pgw.Validate(); err != nil {
		log.WithError(err).Error("Invalid PGW domain object")

		return errors.Wrap(err, "invalid PGW domain object")
	}

	if err := srv.ipRepository.Save(s5uIP, pgw.UserPlane.IP); err != nil {
		log.WithError(err).Panic("S5-U IP Address storage error")

		return fmt.Errorf("S5-U IP %q: %w", pgw.UserPlane.IP, ErrSaveIP)
	}

	if err := srv.ipRepository.Save(s5cIP, pgw.ControlPlane.IP); err != nil {
		log.WithError(err).Panic("S5-C IP Address storage error")

		return fmt.Errorf("S5-C IP %q: %w", pgw.ControlPlane.IP, ErrSaveIP)
	}

	return nil
}

// Get retrieves PGW information from the repository.
func (srv *Service) Get() (*domain.Pgw, error) {
	userPlaneIP, err := srv.ipRepository.Get(s5uIP)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get user plane IP address")
	}

	controlPlaneIP, err := srv.ipRepository.Get(s5cIP)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get control plane IP address")
	}

	return domain.New(controlPlaneIP, userPlaneIP, "", ""), nil
}

// Remove deletes the PGW information from the repository.
func (srv *Service) Remove() {
	srv.ipRepository.Delete(s5uIP)
	srv.ipRepository.Delete(s5cIP)
}
