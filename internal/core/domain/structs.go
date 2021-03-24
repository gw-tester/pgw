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

package domain

import (
	"net"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/wmnsk/go-gtp/gtpv2"
)

// ErrInvalidPgw indicates that an invalid PGW domain field was provided.
var ErrInvalidPgw = errors.New("invalid PGW domain")

// Pgw stores User and Control Plane information about PDN Gateway.
type Pgw struct {
	ControlPlane *ControlPlane
	UserPlane    *UserPlane
	Sgi          *Sgi
}

type Sgi struct {
	Link   netlink.Link
	Subnet *net.IPNet
}

// ControlPlane stores information related to Control Plane.
type ControlPlane struct {
	IP string
}

// UserPlane stores information related to User Plane.
type UserPlane struct {
	IP string
}

// GetAddress retrieves the IP address of the Control Plane Network interface.
func (p *ControlPlane) GetAddress() (*net.UDPAddr, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	addr, err := net.ResolveUDPAddr("udp", p.IP+gtpv2.GTPCPort)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a control plane address")
	}

	return addr, nil
}

func (p *ControlPlane) String() string {
	if p == nil {
		return "<nil>"
	}

	addr, err := p.GetAddress()
	if err != nil {
		return ""
	}

	return addr.String()
}

// GetAddress retrieves the IP address of the User Plane Network interface.
func (p *UserPlane) GetAddress() (*net.UDPAddr, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	addr, err := net.ResolveUDPAddr("udp", p.IP+gtpv2.GTPUPort)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a user plane address")
	}

	return addr, nil
}

func (p *UserPlane) String() string {
	if p == nil {
		return "<nil>"
	}

	addr, err := p.GetAddress()
	if err != nil {
		return ""
	}

	return addr.String()
}

// New creates a PGW instance.
func New(s5cIP, s5uIP, sgiLink, sgiSubnet string) (pgw *Pgw) {
	link, err := netlink.LinkByName(sgiLink)
	if err != nil {
		log.WithError(err).Warnf("SGI %s link retrieve error", sgiLink)
	}

	_, subnet, err := net.ParseCIDR(sgiSubnet)
	if err != nil {
		log.WithError(err).Warnf("SGI %s subnet parse error", subnet)
	}

	pgw = &Pgw{
		ControlPlane: &ControlPlane{
			IP: s5cIP,
		},
		UserPlane: &UserPlane{
			IP: s5uIP,
		},
		Sgi: &Sgi{
			Link:   link,
			Subnet: subnet,
		},
	}

	return
}

// Validate the IP address value of the Control Plane Network Interface.
func (p *ControlPlane) Validate() error {
	if p.IP == "" {
		return errors.Wrap(ErrInvalidPgw, "empty control plane IP address")
	}

	return nil
}

// Validate the IP address value of the User Plane Network Interface.
func (p *UserPlane) Validate() error {
	if p.IP == "" {
		return errors.Wrap(ErrInvalidPgw, "empty user plane IP address")
	}

	return nil
}

// Validate checks if the fields don't have empty values assigned.
func (p *Pgw) Validate() error {
	if p.ControlPlane == nil {
		return errors.Wrap(ErrInvalidPgw, "no control plane")
	}

	if p.UserPlane == nil {
		return errors.Wrap(ErrInvalidPgw, "no user plane")
	}

	if err := p.ControlPlane.Validate(); err != nil {
		return err
	}

	return p.UserPlane.Validate()
}
