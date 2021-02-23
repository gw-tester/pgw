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

package domain_test

import (
	"strconv"
	"strings"

	"github.com/gw-tester/pgw/internal/core/domain"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wmnsk/go-gtp/gtpv2"
)

var _ = Describe("Pgw", func() {
	var pgw *domain.Pgw

	BeforeEach(func() {
		pgw = domain.New("127.0.0.1", "127.0.0.1", "lo", "10.0.0.0/24")
	})

	Describe("discovering Addresses", func() {
		Context("when user plane address is requested", func() {
			It("should use 2152 port number", func() {
				userPlaneAddress, err := pgw.UserPlane.GetAddress()
				Expect(err).NotTo(HaveOccurred())
				Expect(userPlaneAddress).NotTo(BeNil())
				portExpected, _ := strconv.Atoi(strings.ReplaceAll(gtpv2.GTPUPort, ":", ""))
				Expect(userPlaneAddress.Port).To(Equal(portExpected))
			})
		})
		Context("when control plane address is requested", func() {
			It("should use 2123 port number", func() {
				controlPlane, err := pgw.ControlPlane.GetAddress()
				Expect(err).NotTo(HaveOccurred())
				Expect(controlPlane).NotTo(BeNil())
				portExpected, _ := strconv.Atoi(strings.ReplaceAll(gtpv2.GTPCPort, ":", ""))
				Expect(controlPlane.Port).To(Equal(portExpected))
			})
		})
	})

	Describe("validating control and user plane information", func() {
		Context("when control and user plane IP addresses are provided", func() {
			It("should not error", func() {
				err := pgw.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("when control and user plane IP addresses are provided", func() {
			BeforeEach(func() {
				pgw = domain.New("", "", "", "")
			})
			It("should raise an error", func() {
				err := pgw.Validate()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
