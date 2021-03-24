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

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

// Handle drops a IMSI Session response.
func Handle(connection *gtpv2.Conn, sender net.Addr, msg message.Message) error {
	// assert type to refer to the struct field specific to the message.
	// in general, no need to check if it can be type-asserted, as long as the MessageType is
	// specified correctly in AddHandler().
	session, err := connection.GetSessionByTEID(msg.TEID(), sender)
	if err != nil {
		response := message.NewDeleteSessionResponse(
			0, 0,
			ie.NewCause(gtpv2.CauseIMSIIMEINotKnown, 0, 0, 0, nil),
		)
		if err := connection.RespondTo(sender, msg, response); err != nil {
			return errors.Wrap(err, "failed to send an error caused by getting session method")
		}

		return errors.Wrap(err, "failed to get a session from TEID")
	}

	teid, err := session.GetTEID(gtpv2.IFTypeS5S8SGWGTPC)
	if err != nil {
		log.WithError(err).Error("Failed to respond to S-GW with Delete Session Response")

		return nil
	}

	response := message.NewDeleteSessionResponse(
		teid, 0,
		ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
	)

	if err := connection.RespondTo(sender, msg, response); err != nil {
		return errors.Wrap(err, "failed to send a delete response message")
	}

	log.WithFields(log.Fields{
		"IMSI": session.IMSI,
	}).Info("Session deleted")
	connection.RemoveSession(session)

	return nil
}
