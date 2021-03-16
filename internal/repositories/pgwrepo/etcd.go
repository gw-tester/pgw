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
	"github.com/coreos/go-etcd/etcd"
	"github.com/gw-tester/pgw/internal/core/ports"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type etcdStore struct {
	client *etcd.Client
}

// NewETCD creates a new instance to connect to ETCD Cluster.
func NewETCD(url string) ports.IPRepository {
	log.WithFields(log.Fields{
		"Redis URL": url,
	}).Debug("Creating ETC client")

	client := etcd.NewClient([]string{url})

	return &etcdStore{client: client}
}

// Save stores the entry value into a specific id.
func (repo *etcdStore) Save(id, ip string) error {
	_, err := repo.client.Set("/"+id, ip, 0)
	if err != nil {
		return errors.Wrap(err, "Error storing ETCD value")
	}

	log.WithFields(log.Fields{
		"id": id,
		"ip": ip,
	}).Debug("IP address stored")

	return nil
}

// Get retrieves the value of a specific id entry.
func (repo *etcdStore) Get(id string) (string, error) {
	response, err := repo.client.Get(id, false, false)
	if err != nil {
		return "", errors.Wrap(err, "Error getting ETCD value")
	}

	val := response.Node.Value

	log.WithFields(log.Fields{
		"id": id,
		"ip": val,
	}).Debug("IP address retrieved")

	return val, nil
}

// Delete removes the given id entry from the datastore.
func (repo *etcdStore) Delete(id string) {
	_, err := repo.client.Delete(id, true)
	if err != nil {
		log.WithError(err).Error("Error deleting an ETCD entry")
	}
}

// Status is used for performing a ETCD check against a dependency.
func (repo *etcdStore) Status() (interface{}, error) {
	return nil, nil
}
