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

package pgwhdl

import (
	"net"

	"github.com/electrocucaracha/pgw/internal/core/domain"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/wmnsk/go-gtp/gtpv1"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

type create struct {
	userPlaneConnection *gtpv1.UPlaneConn

	addedRoutes []*netlink.Route
	addedRules  []*netlink.Rule

	config *domain.Pgw
}

// Handler defines PGW contracts.
type Handler interface {
	Handle(*gtpv2.Conn, net.Addr, message.Message) error
	Close() error
}

// NewCreate creates a PGW handler for creating ISMI Sessions.
func NewCreate(conn *gtpv1.UPlaneConn, config *domain.Pgw) Handler {
	return &create{
		userPlaneConnection: conn,
		config:              config,
		addedRoutes:         []*netlink.Route{},
		addedRules:          []*netlink.Rule{},
	}
}

// Close removes rules and routes added by the Router and closes user plane connnection.
func (h *create) Close() error {
	for _, route := range h.addedRoutes {
		if err := netlink.RouteDel(route); err != nil {
			log.WithError(err).Warn("Route Deletion error")
		}
	}

	for _, rule := range h.addedRules {
		if err := netlink.RuleDel(rule); err != nil {
			log.WithError(err).Warn("Rule Deletion error")
		}
	}

	return nil
}

func validate(request *message.CreateSessionRequest) error {
	if request.IMSI == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.IMSI}
	}

	if request.MSISDN == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.MSISDN}
	}

	if request.MEI == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.MobileEquipmentIdentity}
	}

	if request.APN == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.AccessPointName}
	}

	if request.ServingNetwork == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.ServingNetwork}
	}

	if request.RATType == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.RATType}
	}

	if request.SenderFTEIDC == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.FullyQualifiedTEID}
	}

	if request.BearerContextsToBeCreated == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.BearerContext}
	}

	if request.PAA == nil {
		return &gtpv2.RequiredIEMissingError{Type: ie.PDNAddressAllocation}
	}

	return nil
}

func getSession(sgwAddress net.Addr, request *message.CreateSessionRequest) (*gtpv2.Session, error) {
	var err error

	session := gtpv2.NewSession(sgwAddress, &gtpv2.Subscriber{Location: &gtpv2.Location{}})

	session.IMSI, err = request.IMSI.IMSI()
	if err != nil {
		return session, errors.Wrap(err, "failed to get IMSI from the session request")
	}

	session.MSISDN, err = request.MSISDN.MSISDN()
	if err != nil {
		return session, errors.Wrap(err, "failed to get MSISDN from the session request")
	}

	session.IMEI, err = request.MEI.MobileEquipmentIdentity()
	if err != nil {
		return session, errors.Wrap(err, "failed to get Mobile Equipment Identity from the session request")
	}

	session.MCC, err = request.ServingNetwork.MCC()
	if err != nil {
		return session, errors.Wrap(err, "failed to get MCC from the session request")
	}

	session.MNC, err = request.ServingNetwork.MNC()
	if err != nil {
		return session, errors.Wrap(err, "failed to get MNC from the session request")
	}

	session.RATType, err = request.RATType.RATType()
	if err != nil {
		return session, errors.Wrap(err, "failed to get RAT Type from the session request")
	}

	teid, err := request.SenderFTEIDC.TEID()
	if err != nil {
		return session, errors.Wrap(err, "failed to get TEID from the session request")
	}

	session.AddTEID(gtpv2.IFTypeS5S8SGWGTPC, teid)

	return session, nil
}

func getContextObjects(sender net.Addr, msg message.Message) (*message.CreateSessionRequest,
	*gtpv2.Session, *gtpv2.Bearer, error) {
	// assert type to refer to the struct field specific to the message.
	// in general, no need to check if it can be type-asserted, as long as the MessageType is
	// specified correctly in AddHandler().
	request := msg.(*message.CreateSessionRequest)
	if err := validate(request); err != nil {
		return request, nil, nil, errors.Wrap(err, "failed to get a valid create session request")
	}

	// keep session information retrieved from the message.
	session, err := getSession(sender, request)
	if err != nil {
		return request, session, nil, errors.Wrap(err, "failed to get a new create session")
	}

	bearer := session.GetDefaultBearer()

	bearer.APN, err = request.APN.AccessPointName()
	if err != nil {
		return request, session, bearer, errors.Wrap(err, "failed to get access point name for the bearer object")
	}

	bearer.SubscriberIP, err = request.PAA.IPAddress()
	if err != nil {
		return request, session, bearer, errors.Wrap(err, "failed to get the suscriber IP for the bearer object")
	}

	for _, childIE := range request.BearerContextsToBeCreated.ChildIEs {
		if childIE.Type == ie.EPSBearerID {
			bearer.EBI, err = childIE.EPSBearerID()
			if err != nil {
				return request, session, bearer, errors.Wrapf(err, "failed to get EPSBearerID from %s childIE", childIE)
			}

			break
		}
	}

	return request, session, bearer, nil
}

func removePreviousIMSISession(connection *gtpv2.Conn, imsi string) error {
	// remove previous session for the same subscriber if exists.
	previousSession, err := connection.GetSessionByIMSI(imsi)
	if err == nil {
		connection.RemoveSession(previousSession)
	}

	return nil
}

func getTunnelData(session *gtpv2.Session, childIEs []*ie.IE) (string, uint32, bool) {
	for _, childIE := range childIEs {
		if childIE.Type == ie.FullyQualifiedTEID {
			it, err := childIE.InterfaceType()
			if err != nil {
				log.WithError(err).Warnf("Failed to get InterfaceType from %s childIE", childIE)
			}

			oteiU, err := childIE.TEID()
			if err != nil {
				log.WithError(err).Warnf("Failed to get TEID from %s childIE", childIE)
			}

			session.AddTEID(it, oteiU)

			s5sgwuIP, err := childIE.IPAddress()
			if err != nil {
				log.WithError(err).Warnf("Failed to get IP Address from %s childIE", childIE)
			}

			return s5sgwuIP, oteiU, true
		}
	}

	return "", 0, false
}

// Handle creates a IMSI Session request.
func (h *create) Handle(connection *gtpv2.Conn, sender net.Addr, msg message.Message) error {
	request, session, bearer, err := getContextObjects(sender, msg)
	if err != nil {
		return err
	}

	if err := removePreviousIMSISession(connection, session.IMSI); err != nil {
		return err
	}

	s5sgwuIP, oteiU, ok := getTunnelData(session, request.BearerContextsToBeCreated.ChildIEs)
	if !ok {
		return errors.New("failed to get tunnel data")
	}

	s5cFTEID := connection.NewSenderFTEID(h.config.ControlPlane.IP, "").WithInstance(1)
	s5uFTEID := h.userPlaneConnection.NewFTEID(gtpv2.IFTypeS5S8PGWGTPU, h.config.UserPlane.IP, "").WithInstance(2)

	s5sgwTEID, err := session.GetTEID(gtpv2.IFTypeS5S8SGWGTPC)
	if err != nil {
		return errors.Wrap(err, "failed to get TEID from the current session")
	}

	response := message.NewCreateSessionResponse(
		s5sgwTEID, 0,
		ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
		s5cFTEID,
		ie.NewPDNAddressAllocation(bearer.SubscriberIP),
		ie.NewAPNRestriction(gtpv2.APNRestrictionPublic2),
		ie.NewBearerContext(
			ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
			ie.NewEPSBearerID(bearer.EBI),
			s5uFTEID,
			ie.NewChargingID(bearer.ChargingID),
		),
	)

	if request.SGWFQCSID != nil {
		response.PGWFQCSID = ie.NewFullyQualifiedCSID(h.config.ControlPlane.IP, 1)
	}

	session.AddTEID(gtpv2.IFTypeS5S8PGWGTPC, s5cFTEID.MustTEID())
	session.AddTEID(gtpv2.IFTypeS5S8PGWGTPU, s5uFTEID.MustTEID())

	if err := connection.RespondTo(sender, request, response); err != nil {
		return errors.Wrap(err, "failed to send a respond through the control plane connection")
	}

	if err := addSession(session, connection); err != nil {
		return errors.Wrap(err, "failed to activate and add session created to the session list")
	}

	if err := h.setupUserPlane(s5sgwuIP, bearer.SubscriberIP, oteiU, s5uFTEID.MustTEID()); err != nil {
		return errors.Wrap(err, "failed to setup User Plane routes and rules")
	}

	return nil
}

func addSession(session *gtpv2.Session, connection *gtpv2.Conn) error {
	s5pgwTEID, err := session.GetTEID(gtpv2.IFTypeS5S8PGWGTPC)
	if err != nil {
		return errors.Wrap(err, "failed to get TEID form the current session")
	}

	if err := session.Activate(); err != nil {
		return errors.Wrap(err, "failed to activate the current session")
	}

	connection.RegisterSession(s5pgwTEID, session)

	return nil
}

func newRoute(dst *net.IPNet, linkIndex int) *netlink.Route {
	return &netlink.Route{
		Dst:       dst,
		LinkIndex: linkIndex,
		Scope:     netlink.SCOPE_LINK, // scope link
		Protocol:  4,                  // proto static
		Priority:  1,                  // metric 1
	}
}

func (h *create) appendRoute(route *netlink.Route) {
	if err := netlink.RouteReplace(route); err != nil {
		log.WithError(err).Warnf("Failed to add %s route", route)
	}

	log.WithFields(log.Fields{
		"route": route,
	}).Debug("Adding User plane route")

	h.addedRoutes = append(h.addedRoutes, route)
}

func findRule(ms32 *net.IPNet, ifName string) bool {
	// Priority 0 rules
	rules, _ := netlink.RuleList(0)
	for _, rule := range rules {
		if rule.IifName == ifName && rule.Dst == ms32 {
			log.Debugf("%s rule found", rule)

			return true
		}
	}

	return false
}

func (h *create) appendRule(rule *netlink.Rule) {
	if err := netlink.RuleAdd(rule); err != nil {
		log.WithError(err).Warnf("Failed to add %s rule", rule)
	}

	log.WithFields(log.Fields{
		"rule": rule,
	}).Debug("Adding User plane rule")

	h.addedRules = append(h.addedRules, rule)
}

// setupUserPlane configures the routes and rules for User plane traffic.
func (h *create) setupUserPlane(peer, ms string, otei, itei uint32) error {
	peerIP := net.ParseIP(peer)
	msIP := net.ParseIP(ms)

	if err := h.userPlaneConnection.AddTunnelOverride(peerIP, msIP, otei, itei); err != nil {
		return errors.Wrap(err, "failed to add a GTP-U tunnel")
	}

	ms32 := &net.IPNet{
		IP:   msIP,
		Mask: net.CIDRMask(32, 32),
	}
	dlroute := newRoute(ms32, h.userPlaneConnection.KernelGTP.Link.Attrs().Index)
	dlroute.Table = 3001
	h.appendRoute(dlroute)
	h.appendRoute(newRoute(h.config.Sgi.Subnet, h.config.Sgi.Link.Attrs().Index))

	if !findRule(ms32, h.config.Sgi.Link.Attrs().Name) {
		rule := netlink.NewRule()
		rule.IifName = h.config.Sgi.Link.Attrs().Name
		rule.Dst = ms32
		rule.Table = 3001

		h.appendRule(rule)
	}

	return nil
}
