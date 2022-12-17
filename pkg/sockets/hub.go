/*
 *   Copyright (c) 2022 CARISA
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package sockets

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/arrays"
	"github.com/swpoolcontroller/pkg/strings"
)

const (
	errSendingMsg         = "Hub-> Sending a message. The client will be removed"
	errNotify             = "Hub-> Notifying status. The client will be removed"
	errRemovingDiedClient = "Hub-> Removing died client by expiration"
	errClosingClient      = "Hub-> Closing client"
)

const (
	infCheckerDes     = "Hub-> The check timer is deactivated because there are no clients"
	infClientReg      = "Hub-> Client registered"
	infClientUnReg    = "Hub-> Unregistering client"
	infClientUnRegd   = "Hub-> Client unregisted"
	infHubActive      = "Hub-> The hub is set to active"
	infHubInactive    = "Hub-> The hub is set to inactive. Previous status: active"
	infClientDied     = "Hub-> Client died by expiration"
	infArraySize      = "Hub-> Array size after removing expired clients"
	infNotify         = "Hub-> Notifying status"
	infConfigChanged  = "Hub-> The configuration has been changed"
	infClientID       = "ClientID"
	infLength         = "Length"
	infStatus         = "Status"
	infBufferLatency  = "Buffer + Comm latency"
	infLastMsgDate    = "Last message date"
	infExpirationDate = "Expiration date"
	infActualDate     = "Actual date"
	infPrevStatus     = "Previous status"
	infLastNotify     = "Last notification date"
	infConfig         = "Config"
)

// Status are the communication status between the sender and the hub
type Status int

const (
	// Deactivated the hub is in non-transmit mode. There are not clients connected
	Deactivated Status = iota
	// Active is when the hub is in transmit mode. There are clients connected,
	// but there is not transmission from sender
	Active
	// Streaming is when the hub is receiving information from the sender
	Streaming
	// Inactive is when the hub was in streaming mode but there is no transmission from the sender
	Inactive
	// Closed the hub is closed. Do not exit the go routine so as not to cause deadlock,
	// although no request will take effect in this state.
	Closed
)

type Config struct {
	// CommLatency is the time in seconds before the communication goes to the inactive state
	CommLatency time.Duration
	// Buffer is the time in seconds to store metrics before sending to the hub
	Buffer time.Duration
	// TaskTime defines how often the hub makes maintenance task
	TaskTime time.Duration
	// NotificationTime defines how often a notification is sent
	NotificationTime time.Duration
}

// string returns struct as string
func (c *Config) string() string {
	r, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(r)
}

// Client identifies a connection socket.
type Client struct {
	id         string
	conn       *websocket.Conn
	expiration time.Time
}

// NewClient builds client struct. id identifies the session. The client expires
// depending of the expiration parameter
func NewClient(id string, conn *websocket.Conn, expiration time.Duration) Client {
	return Client{
		id:         id,
		conn:       conn,
		expiration: time.Now().Add(expiration),
	}
}

// Hub manages the socket pool. The hub registers clients across of the reg channel, unregisters
// clients, broadcast messages and returns errors to the sender. Also the hub checks communication
// status, socket, etc.
// The life cycle is: Deactivated -> Active; Active -> Deactivated, Streaming;
// Streaming -> Inactive, Deactivated; Inactive -> Streaming, Deactivated
type Hub struct {
	clients []Client

	config Config

	reg     chan Client
	unreg   chan string
	errors  chan []error
	infos   chan []string
	send    chan string
	sconfig chan Config
	statusc chan chan Status
	closec  chan struct{}

	// lastNotification the time when the last notification was sended
	lastNotification time.Time

	// lastMessage the time when the last messages was sended
	lastMessage time.Time
	status      Status
}

// NewHub builds Hub service.
// The config establishes how the hub should behave
// The infos channel receives all infos into the hub
// The errors channel receives all errors into the hub
func NewHub(
	cnf Config,
	infos chan []string,
	errors chan []error) *Hub {
	return &Hub{
		clients:     []Client{},
		config:      cnf,
		reg:         make(chan Client),
		unreg:       make(chan string),
		send:        make(chan string),
		sconfig:     make(chan Config),
		infos:       infos,
		errors:      errors,
		statusc:     make(chan chan Status),
		closec:      make(chan struct{}),
		lastMessage: time.Time{},
		status:      Deactivated,
	}
}

// Register registers client into the hub
func (h *Hub) Register(client Client) {
	h.reg <- client
}

// Unregister unregisters client into the hub
func (h *Hub) Unregister(id string) {
	h.unreg <- id
}

// Send sends message to the all clients into hub
func (h *Hub) Send(message string) {
	h.send <- message
}

// Send the config to the hub
func (h *Hub) Config(cnf Config) {
	h.sconfig <- cnf
}

// Status request hub status via channel
func (h *Hub) Status(resp chan Status) {
	h.statusc <- resp
}

// Stop finishes the hub
func (h *Hub) Stop() {
	close(h.closec)
}

// Run registers and unregisters clients, sends messages and remove died clients. Launches a gouroutine
func (h *Hub) Run() {
	go func() {
		check := time.NewTimer(h.config.TaskTime)

		for {
			select {
			case client := <-h.reg:
				h.register(client, check)
			case id := <-h.unreg:
				h.unregister(id)
			case message := <-h.send:
				h.sendMessageFromDevice(message)
			case countResp := <-h.statusc:
				countResp <- h.status
			case cnf := <-h.sconfig:
				h.config = cnf
				h.infos <- []string{strings.Format(infConfigChanged, strings.FMTValue(infConfig, h.config.string()))}
			case <-check.C:
				h.controllerStatus()
				h.statusNotification()
				h.removeDeadClient()
				h.tryResetTimer(check)
			case <-h.closec:
				h.close(check)
			}
		}
	}()
}

func (h *Hub) close(check *time.Timer) {
	h.closeh()

	if !check.Stop() {
		<-check.C
	}

	h.status = Closed
}

// tryResetTimer resets timer if there are clients.
// If there aren't clients, the timer is not activated and the CPU is saved.
func (h *Hub) tryResetTimer(check *time.Timer) {
	if len(h.clients) > 0 {
		check.Reset(h.config.TaskTime)
	} else {
		h.infos <- []string{infCheckerDes}
	}
}

func (h *Hub) register(client Client, check *time.Timer) {
	if h.status == Closed {
		return
	}

	h.clients = append(h.clients, client)

	countc := len(h.clients)

	// If there is a client, it means that the timer was not activated before
	// and therefore is activated.
	if countc == 1 {
		check.Reset(h.config.CommLatency)
		// As soon as there is a client, the hub switches to reception mode.
		h.active()
	}

	h.infos <- []string{
		strings.Format(
			infClientReg,
			strings.FMTValue(infClientID, client.id),
			strings.FMTValue(infLength, strconv.Itoa(len(h.clients))),
			strings.FMTValue(infLength, statusString(h.status)))}
}

func (h *Hub) unregister(id string) {
	if h.status == Closed {
		return
	}

	if err := h.removeClient(id); err != nil {
		h.errors <- []error{errors.Wrap(
			err,
			strings.Format(
				infClientUnReg,
				strings.FMTValue(infLength, strconv.Itoa(len(h.clients)))))}
	}

	h.infos <- []string{
		strings.Format(
			infClientUnRegd,
			strings.FMTValue(infClientID, id),
			strings.FMTValue(infLength, strconv.Itoa(len(h.clients))),
			strings.FMTValue(infLength, statusString(h.status)))}
}

// sendMessageFromDevice send message to the all clients registered.
// If sending the message throw a error, the client is removed
func (h *Hub) sendMessageFromDevice(message string) {
	if h.status == Deactivated || h.status == Closed {
		return
	}

	// The message may be changed by sendmessage
	status := h.status

	h.sendMessage([]byte(strings.Concat("1:", message)), errSendingMsg)

	h.lastMessage = time.Now()

	if h.status != Streaming {
		h.infos <- []string{strings.Format(infHubActive, strings.FMTValue(infPrevStatus, statusString(status)))}
		h.status = Streaming
	}
}

// controllerStatus controls the life cycle if the hub
func (h *Hub) controllerStatus() {
	if h.status == Streaming {
		// The idle time is the sum of the time it takes for the sender to create the buffer
		// and a possible latency time
		idleTime := h.config.Buffer + h.config.CommLatency
		if h.lastMessage.Add(idleTime).Before(time.Now()) {
			h.inactive()
			h.infos <- []string{
				strings.Format(
					infHubInactive,
					strings.FMTValue(infBufferLatency, h.lastMessage.String()),
					strings.FMTValue(infLastMsgDate, idleTime.String()))}
		}
	}
}

// statusNotification sends the status to the web client (Only for these states Active or Inactive)
func (h *Hub) statusNotification() {
	if h.status != Active && h.status != Inactive {
		return
	}

	// If the time for the next notification has not elapsed, it does not send the next notification
	if h.lastNotification.Add(h.config.NotificationTime).After(time.Now()) {
		return
	}

	// The message may be changed by sendmessage
	status := h.status

	msg := []byte(strings.Concat("0:", strconv.Itoa(int(h.status))))
	h.sendMessage(msg, errNotify)

	h.infos <- []string{
		strings.Format(
			infNotify,
			strings.FMTValue(infActualDate, time.Now().String()),
			strings.FMTValue(infLastNotify, h.lastNotification.String()),
			strings.FMTValue(infStatus, statusString(status)))}
}

// sendMessage sends messages to the client
func (h *Hub) sendMessage(message []byte, errMessage string) {
	var errs []error

	var brokenclients []uint16

	for i, c := range h.clients {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			brokenclients = append(brokenclients, uint16(i))

			if err := h.closeClient(uint16(i)); err != nil {
				errs = append(errs, errors.Wrap(err, errMessage))
			}

			errs = append(
				errs,
				errors.Wrap(
					err,
					strings.Format(errMessage, strings.FMTValue(infClientID, c.id))))
		}
	}

	if len(brokenclients) > 0 {
		h.removeClientByPos(brokenclients...)
	}

	if len(errs) > 0 {
		h.errors <- errs
		h.infos <- []string{strings.Format(infArraySize, strings.FMTValue(infLength, strconv.Itoa(len(h.clients))))}
	}
}

func (h *Hub) active() {
	h.status = Active
	h.lastNotification = time.Now()
}

func (h *Hub) inactive() {
	h.status = Inactive
	h.lastNotification = time.Now()
}

// removeDeadClient removes died clients
func (h *Hub) removeDeadClient() {
	var (
		infos []string
		errs  []error
	)

	var deadClients []uint16

	for i, c := range h.clients {
		if c.expiration.Before(time.Now()) {
			deadClients = append(deadClients, uint16(i))

			clientID := h.clients[i].id

			if err := h.closeClient(uint16(i)); err != nil {
				errs = append(
					errs,
					errors.Wrap(
						err,
						strings.Format(errRemovingDiedClient, strings.FMTValue(infClientID, clientID))))
			}

			infos = append(
				infos,
				strings.Format(
					infClientDied,
					strings.FMTValue(infClientID, clientID),
					strings.FMTValue(infExpirationDate, c.expiration.String()),
					strings.FMTValue(infActualDate, time.Now().String())))
		}
	}

	if len(errs) > 0 {
		h.errors <- errs
	}

	if len(deadClients) > 0 {
		h.removeClientByPos(deadClients...)
	}

	if len(infos) > 0 {
		infos = append(
			infos,
			strings.Format(infArraySize, strings.FMTValue(infLength, strconv.Itoa(len(h.clients)))))
		h.infos <- infos
	}
}

// findClient seeks a client by ID. If the client is not found returns -1
func (h *Hub) findClient(clientID string) int {
	for i, c := range h.clients {
		if c.id == clientID {
			return i
		}
	}

	return -1
}

// removeClient removes client by id
func (h *Hub) removeClient(clientID string) error {
	pos := h.findClient(clientID)
	if pos == -1 {
		return nil
	}

	posr := uint16(pos)

	if err := h.closeClient(posr); err != nil {
		return err
	}

	h.removeClientByPos(posr)

	return nil
}

// removeClientByPos remove clients by position
func (h *Hub) removeClientByPos(pos ...uint16) {
	h.clients = arrays.Remove(h.clients, pos...)

	if len(h.clients) == 0 {
		h.status = Deactivated
	}
}

// closeClient closes socket client
func (h *Hub) closeClient(pos uint16) error {
	clientID := h.clients[pos].id

	if err := h.clients[pos].conn.Close(); err != nil {
		return errors.Wrap(err, strings.Format(errClosingClient, strings.FMTValue(infClientID, clientID)))
	}

	return nil
}

// closeh closes all the clients socket
func (h *Hub) closeh() {
	for _, c := range h.clients {
		c.conn.Close()
	}
}

func statusString(s Status) string {
	switch s {
	case Deactivated:
		return "Deactivated"
	case Active:
		return "Active"
	case Streaming:
		return "Streaming"
	case Inactive:
		return "Inactive"
	case Closed:
		return "Closed"
	}

	return "None"
}
