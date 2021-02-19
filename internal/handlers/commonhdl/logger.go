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

package commonhdl

import (
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

// Logger is a middleware handler that does request logging.
type Logger struct {
	handle gtpv2.HandlerFunc
}

// Log handles the request by passing it to the real handler and logging the request details.
func (h *Logger) Log(connection *gtpv2.Conn, sender net.Addr, message message.Message) error {
	log.WithFields(log.Fields{
		"messageType":   message.MessageTypeName(),
		"addressSource": sender,
	}).Info("Session received")

	return h.handle(connection, sender, message)
}

// NewLogger constructs a new Logger middleware handler.
func NewLogger(handlerFunc gtpv2.HandlerFunc) *Logger {
	return &Logger{
		handle: handlerFunc,
	}
}
