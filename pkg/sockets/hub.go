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
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/strings"
)

type Status int

const (
	// Active the socket is running and the communication between the micro and server is active
	Active Status = iota
	// Break the socket is break
	Break
	// InactiveComm the socket is running but the communication between the micro and server is inactive
	InactiveComm
	// BreakComm the socket is running but the communication between the micro and server is break
	BreakComm
)

// Client identifies a connection socket. If the client expire, the client
// will be removed of the hub
type Client struct {
	id         string
	conn       *websocket.Conn
	status     Status
	expiration time.Time
	// lastMessage the time when the last messages is sended
	lastMessage time.Time
}

// NewClient builds client struct. id identifies the session. The client expires
// depending of the expiration parameter
func NewClient(id string, conn *websocket.Conn, expiration time.Duration) Client {
	return Client{
		id:         id,
		conn:       conn,
		status:     Active,
		expiration: time.Now().Add(expiration),
	}
}

// Hub manages the socket pool. The hub registers clients across of the reg channel, unregisters
// clients, broadcast messages and returns errors to the sender. Also the hub checks communication
// status, socket, etc. The hub works in a goroutine
type Hub struct {
	client []Client

	reg       chan Client
	unreg     chan string
	send      chan []byte
	errors    chan []error
	reqStatus chan string
	resStatus chan Status

	inactiveCommTime time.Duration
	breakComm        time.Duration
}

// NewHub builds Hub struct. The hub checks clients expiration or client status.
// The inactiveCommTime checks each when the communication between
// the micro and server is inactive and breakComm parameter cheks each when the communication
// between the micro and server is break
// The errors channel receives al errors into the hub
func NewHub(
	inactiveCommTime time.Duration,
	breakComm time.Duration,
	errors chan []error) *Hub {
	return &Hub{
		client:           []Client{},
		reg:              make(chan Client),
		unreg:            make(chan string),
		send:             make(chan []byte),
		errors:           errors,
		reqStatus:        make(chan string),
		resStatus:        make(chan Status),
		inactiveCommTime: inactiveCommTime,
		breakComm:        breakComm,
	}
}

// Run starts the hub process in a goroutine
func (h *Hub) Run() {
	go h.run()
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
func (h *Hub) Send(message []byte) {
	h.send <- message
}

// Status returns the client status
func (h *Hub) Status(id string) Status {
	h.reqStatus = id
}

// run registers and unregisters clients, sends messages and remove died clients
func (h *Hub) run() {
	select {
	case client := <-h.reg:
		h.client = append(h.client, client)
	case id := <-h.unreg:
		if err := h.removeClient(id); err != nil {
			h.errors <- []error{errors.Wrap(err, strings.Concat("Removing client: ", id))}
		}
	case message := <-h.send:
		h.sendMessage(message)
	case <-time.After(h.inactiveCommTime):
		h.updateClientStatus()
		h.removeDeadClient()
	}
}

// sendMessage send message to the all clients registered
func (h *Hub) sendMessage(message []byte) {
	var errs []error

	for i, c := range h.client {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			h.client[i].status = Break

			errs = append(errs, errors.Wrap(err, strings.Concat("Sending message to: ", c.id)))

			continue
		}

		h.client[i].lastMessage = time.Now()
	}

	if len(errs) > 0 {
		h.errors <- errs
	}
}

// removeDeadClient removes died clients
func (h *Hub) removeDeadClient() {
	var errs []error

	for _, c := range h.client {
		if c.expiration.After(time.Now()) {
			if err := h.removeClient(c.id); err != nil {
				errs = append(errs, errors.Wrap(err, strings.Concat("Removing died client: ", c.id)))
			}
		}
	}

	if len(errs) > 0 {
		h.errors <- errs
	}
}

// removeClient removes client by id
func (h *Hub) removeClient(clientID string) error {
	pos := -1

	for i, c := range h.client {
		if c.id == clientID {
			pos = i
		}
	}

	err := h.client[pos].conn.Close()

	h.client[pos] = h.client[len(h.client)-1]
	h.client = h.client[:len(h.client)-1]

	return errors.Wrap(err, strings.Concat("Removing client ID:", clientID))
}

// updateClientStatus updates the client status
func (h *Hub) updateClientStatus() {
	for i, c := range h.client {
		if c.lastMessage.Add(h.inactiveCommTime).After(time.Now()) {
			h.client[i].status = InactiveComm

			continue
		}

		if c.lastMessage.Add(h.breakComm).After(time.Now()) {
			h.client[i].status = BreakComm

			continue
		}
	}
}
