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

package discover

import (
	"net"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func getLinksByNetwork(network string) ([]netlink.Link, bool) {
	_, parsedNetwork, err := net.ParseCIDR(network)
	if err != nil {
		log.WithError(err).Errorf("Failed to parse %s network", network)

		return nil, false
	}

	links := []netlink.Link{}
	filter := &netlink.Route{Dst: parsedNetwork}

	routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, filter, netlink.RT_FILTER_DST)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve local routes")

		return nil, false
	}

	if routes == nil {
		return links, false
	}

	log.WithFields(log.Fields{
		"route": routes,
	}).Debugf("Existing local Routes for %s", network)

	for _, route := range routes {
		link, _ := netlink.LinkByIndex(route.LinkIndex)
		links = append(links, link)
	}

	return links, true
}

func waitForNetworkCreation(network string) {
	log.Info("Waiting for creation of the network...")

	routeUpdates := make(chan netlink.RouteUpdate)
	done := make(chan struct{})

	defer close(done)

	if err := netlink.RouteSubscribe(routeUpdates, done); err != nil {
		log.WithError(err).Panic("Failed to susbscribe to local route change event")
	}

	for {
		update, ok := <-routeUpdates
		if !ok {
			panic("route event closed for some unknown reason, re-subscribe")
		}

		if update.Type == syscall.RTM_NEWROUTE {
			log.WithFields(log.Fields{
				"destination": update.Route.Dst,
				"gateway":     update.Route.Gw,
			}).Info("Receive route add event") // sudo ip route add 192.168.1.0/24 via 192.168.0.1 dev eno1

			if _, ok := getLinksByNetwork(network); ok {
				log.Infof("%s network was created", network)

				return
			}
		}
	}
}

func getFirstIPAddress(links []netlink.Link) *net.IPNet {
	if links == nil || len(links) < 1 {
		return nil
	}

	addresses, err := netlink.AddrList(links[0], netlink.FAMILY_V4)
	if err != nil {
		log.WithError(err).Panic("Error getting the IPv4 addresses")
	}

	if addresses == nil || len(addresses) < 1 {
		return nil
	}

	return addresses[0].IPNet
}

// GetIPFromNetwork waits until a specific network is created and returns its first IP address.
func GetIPFromNetwork(net string) *net.IPNet {
	if _, ok := getLinksByNetwork(net); !ok {
		waitForNetworkCreation(net)
	}

	links, _ := getLinksByNetwork(net)

	return getFirstIPAddress(links)
}
