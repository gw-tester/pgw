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

package pgwrepo

import (
	"github.com/gw-tester/pgw/internal/core/ports"
)

type memkvs struct {
	kvs map[string]string
}

// NewMemKVS creates a new instance for Key/Value store.
func NewMemKVS() ports.IPRepository {
	return &memkvs{kvs: map[string]string{}}
}

// Save stores an IP address with specific Identifier.
func (repo *memkvs) Save(id, ip string) error {
	repo.kvs[id] = ip

	return nil
}

// Get retrieves the value of a specific id entry.
func (repo *memkvs) Get(id string) (string, error) {
	return repo.kvs[id], nil
}

// Delete removes the given id entry from the datastore.
func (repo *memkvs) Delete(id string) {
	delete(repo.kvs, id)
}

// Status is used for performing a MemKV check against a dependency.
func (repo *memkvs) Status() (interface{}, error) {
	return nil, nil
}
