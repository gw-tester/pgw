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

package pgwsrv_test

import (
	"github.com/gw-tester/pgw/internal/core/domain"
	"github.com/gw-tester/pgw/internal/core/services/pgwsrv"
	"github.com/gw-tester/pgw/internal/repositories/pgwrepo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service", func() {
	var (
		service *pgwsrv.Service
		pgw     *domain.Pgw
	)

	const (
		s5uIPAddress = "127.0.0.1"
		s5cIPAddress = "127.0.0.2"
	)

	JustBeforeEach(func() {
		repo := pgwrepo.NewMemKVS()
		service = pgwsrv.New(repo)
	})

	Describe("storing user and control plane IP addresses", func() {
		BeforeEach(func() {
			pgw = domain.New(s5cIPAddress, s5uIPAddress, "", "")
		})
		Context("when information is valid", func() {
			It("should store the user and plane information", func() {
				By("Storing the IP addresses into the DB")

				err := service.Create(pgw)
				Expect(err).NotTo(HaveOccurred())

				By("Getting the IP addresses from the DB")
				instance, err := service.Get()
				Expect(err).NotTo(HaveOccurred())
				Expect(instance).NotTo(BeNil())
				Expect(instance.UserPlane.IP).To(Equal(s5uIPAddress))
				Expect(instance.ControlPlane.IP).To(Equal(s5cIPAddress))
			})
		})
	})

	Describe("avoiding to store invalid control plane IP addresses", func() {
		BeforeEach(func() {
			pgw = domain.New("", s5uIPAddress, "", "")
		})
		Context("when control plane is invalid", func() {
			It("should raise an error", func() {
				err := service.Create(pgw)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("avoiding to store invalid user plane IP addresses", func() {
		BeforeEach(func() {
			pgw = domain.New(s5cIPAddress, "", "", "")
		})
		Context("when user plane is invalid", func() {
			It("should raise an error", func() {
				err := service.Create(pgw)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
