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
	"errors"
)

var (
	errEmptyControlPlaneIPAddress = errors.New("no IP address was provided to the control plane")
	errEmptyUserPlaneIPAddress    = errors.New("no IP address was provided to the user plane")
	errNoControlPlane             = errors.New("no control plane data was provided")
	errNoUserPlane                = errors.New("no user plane data was provided")
)

// Validate the IP address value of the Control Plane Network Interface.
func (p *ControlPlane) Validate() error {
	if p.IP == "" {
		return errEmptyControlPlaneIPAddress
	}

	return nil
}

// Validate the IP address value of the User Plane Network Interface.
func (p *UserPlane) Validate() error {
	if p.IP == "" {
		return errEmptyUserPlaneIPAddress
	}

	return nil
}

// Validate checks if the fields don't have empty values assigned.
func (p *Pgw) Validate() error {
	if p.ControlPlane == nil {
		return errNoControlPlane
	}

	if p.UserPlane == nil {
		return errNoUserPlane
	}

	if err := p.ControlPlane.Validate(); err != nil {
		return err
	}

	return p.UserPlane.Validate()
}
