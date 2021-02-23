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

package ports

import "github.com/gw-tester/pgw/internal/core/domain"

// IPRepository exposes methods to save, get and drop IP address information.
type IPRepository interface {
	Save(id, ip string) error
	Get(id string) (string, error)
}

// PGWService exposes an API to store, retrieve and delete PGW instances.
type PGWService interface {
	Create(pgw domain.Pgw) error
	Get() (*domain.Pgw, error)
}
