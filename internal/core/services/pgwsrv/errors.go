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
	"errors"
)

var (
	// ErrInvalidPGW indicates that an invalid PGW domain object was passed.
	ErrInvalidPGW = errors.New("invalid PGW domain object")
	// ErrSaveUserPlaneIP indicates a database failure during the storing
	// user plane IP address.
	ErrSaveUserPlaneIP = errors.New("fail to store S5-U IP Address")
	// ErrSaveControlPlaneIP indicates a database failure during the storing
	// control plane IP address.
	ErrSaveControlPlaneIP = errors.New("fail to store S5-C IP Address")
	errGetUserPlaneIP     = errors.New("fail to retrieve S5-U IP Address")
	errGetControlPlaneIP  = errors.New("fail to retrieve S5-C IP Address")
)
