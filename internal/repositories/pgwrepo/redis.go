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
	"github.com/go-redis/redis"
	"github.com/gw-tester/pgw/internal/core/ports"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type redisStore struct {
	client *redis.Client
}

// NewRedis creates a new instance to connect to Redis Server.
func NewRedis(url, password string) ports.IPRepository {
	log.WithFields(log.Fields{
		"Redis URL": url,
	}).Debug("Creating Redis client")

	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       0,
	})
	if _, err := client.Ping().Result(); err != nil {
		log.WithError(err).Panic("Error getting response from Redis server")
	}

	return &redisStore{client: client}
}

// Save stores the entry value into a specific id.
func (repo *redisStore) Save(id, ip string) error {
	if err := repo.client.Set(id, ip, 0).Err(); err != nil {
		return errors.Wrap(err, "Error storing Redis value")
	}

	log.WithFields(log.Fields{
		"id": id,
		"ip": ip,
	}).Debug("IP address stored")

	return nil
}

// Get retrieves the value of a specific id entry.
func (repo *redisStore) Get(id string) (string, error) {
	val, err := repo.client.Get(id).Result()
	if err != nil {
		return "", errors.Wrap(err, "Error getting Redis value")
	}

	log.WithFields(log.Fields{
		"id": id,
		"ip": val,
	}).Debug("IP address retrieved")

	return val, nil
}

// Delete removes the given id entry from the datastore.
func (repo *redisStore) Delete(id string) {
	repo.client.Del(id)
}

// Status is used for performing a Redis check against a dependency.
func (repo *redisStore) Status() (interface{}, error) {
	if _, err := repo.client.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "Ping failed")
	}

	return nil, nil
}
