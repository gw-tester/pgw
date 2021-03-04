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
	"github.com/gw-tester/pgw/internal/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type etcdStore struct {
	client *etcd.Client
}

// NewETCD creates a new instance to connect to ETCD Cluster.
func NewETCD() ports.IPRepository {
	etcdServerAddr := utils.GetEnv("ETCD_URL", "localhost:2379")

	log.WithFields(log.Fields{
		"Redis URL": etcdServerAddr,
	}).Debug("Creating ETC client")

	client := etcd.NewClient([]string{etcdServerAddr})

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
