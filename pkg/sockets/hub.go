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

type clientStatus int

const (
	// active the socket is running and the communication between the micro and server is active
	active clientStatus = iota
	// inactiveComm the socket is running but the communication between the micro and server is inactive
	inactiveComm
	// breakComm the socket is running but the communication between the micro and server is break
	breakComm
)

// Client identifies a connection socket. If the client expire, the client
// will be removed of the hub
type Client struct {
	id         string
	conn       *websocket.Conn
	expiration time.Time
	// lastMessage the time when the last messages is sended
	lastMessage time.Time
}

// NewClient builds client struct. id identifies the session. The client expires
// depending of the expiration parameter
func NewClient(id string, conn *websocket.Conn, expiration time.Duration) Client {
	return Client{
		id:          id,
		conn:        conn,
		expiration:  time.Now().Add(expiration),
		lastMessage: time.Now(),
	}
}

// Hub manages the socket pool. The hub registers clients across of the reg channel, unregisters
// clients, broadcast messages and returns errors to the sender. Also the hub checks communication
// status, socket, etc. The hub works in a goroutine
type Hub struct {
	client []Client

	reg    chan Client
	unreg  chan string
	send   chan []byte
	info   chan string
	errors chan []error
	close  chan struct{}

	inactiveCommTime time.Duration
	breakComm        time.Duration
}

// NewHub builds Hub service.
// The inactiveCommTime establishes each time the communication between
// the micro and server is inactive and breakComm parameter establishes each time the communication
// between the micro and server is break
// The info channel receives all infos into the hub
// The errors channel receives all errors into the hub
func NewHub(
	inactiveCommTime time.Duration,
	breakCommTime time.Duration,
	info chan string,
	errors chan []error) *Hub {
	return &Hub{
		client:           []Client{},
		reg:              make(chan Client),
		unreg:            make(chan string),
		send:             make(chan []byte),
		info:             info,
		errors:           errors,
		close:            make(chan struct{}),
		inactiveCommTime: inactiveCommTime,
		breakComm:        breakCommTime,
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
func (h *Hub) Send(message []byte) {
	h.send <- message
}

// Stop finishes the hub
func (h *Hub) Stop() {
	close(h.close)
}

// Run registers and unregisters clients, sends messages and remove died clients. Launches a gouroutine
func (h *Hub) Run() {
	go func() {
		for {
			select {
			case client := <-h.reg:
				h.client = append(h.client, client)
				h.info <- strings.Concat("Client registered: ", client.id)
			case id := <-h.unreg:
				if err := h.removeClient(id); err != nil {
					h.errors <- []error{errors.Wrap(err, strings.Concat("Unregistering client: ", id))}
				}
				h.info <- strings.Concat("Client unregisted: ", id)
			case message := <-h.send:
				h.sendMessage(message)
			case <-time.After(h.inactiveCommTime):
				h.sendInactiveClientStatus()
				h.removeDeadClient()
			case <-h.close:
				h.closeh()

				return
			}
		}
	}()
}

// sendMessage send message to the all clients registered. If sending the message throw a error, the client is removed
func (h *Hub) sendMessage(message []byte) {
	var errs []error

	var clientsBreak []int

	for i, c := range h.client {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			clientsBreak = append(clientsBreak, i)

			errs = append(errs, errors.Wrap(err, strings.Concat("Sending message to: ", c.id)))
		} else {
			h.client[i].lastMessage = time.Now()
		}
	}

	for _, poscb := range clientsBreak {
		if err := h.removeClientByPos(poscb); err != nil {
			errs = append(errs, errors.Wrap(err, "Removing client after sending a message"))
		}
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
				errs = append(errs, errors.Wrap(err, strings.Concat("Removing died client by expiration: ", c.id)))
			}
			h.info <- strings.Concat("Client died by expiration: ", c.id)
		}
	}

	if len(errs) > 0 {
		h.errors <- errs
	}
}

// findClient seeks a client by ID. If the client is not found returns -1
func (h *Hub) findClient(clientID string) int {
	for i, c := range h.client {
		if c.id == clientID {
			return i
		}
	}

	return -1
}

// removeClient removes client by id
func (h *Hub) removeClient(clientID string) error {
	pos := h.findClient(clientID)
	if pos != -1 {
		return nil
	}

	return h.removeClientByPos(pos)
}

// removeClientByPos removes client by position
func (h *Hub) removeClientByPos(pos int) error {
	clientID := h.client[pos].id

	err := h.client[pos].conn.Close()

	h.client[pos] = h.client[len(h.client)-1]
	h.client = h.client[:len(h.client)-1]

	if err != nil {
		return errors.Wrap(err, strings.Concat("Removing client ID:", clientID))
	}

	return nil
}

// sendInactiveClientStatus send the client status when is inactive. If sending the message throw a error, the client is removed
func (h *Hub) sendInactiveClientStatus() {
	var errs []error

	var clientsBreak []int

	for i, c := range h.client {
		status := active

		if c.lastMessage.Add(h.inactiveCommTime).After(time.Now()) {
			status = inactiveComm
		} else if c.lastMessage.Add(h.breakComm).After(time.Now()) {
			status = breakComm
		}

		if status != active {
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte{byte(status)}); err != nil {
				clientsBreak = append(clientsBreak, i)

				errs = append(errs, errors.Wrap(err, strings.Concat("Sending inactive client status: ", c.id)))
			} else {
				h.info <- strings.Concat("Sending inactive client status: ", c.id)
			}
		}
	}

	for _, poscb := range clientsBreak {
		if err := h.removeClientByPos(poscb); err != nil {
			errs = append(errs, errors.Wrap(err, "Removing client after sending a inactive client status"))
		}
	}

	if len(errs) > 0 {
		h.errors <- errs
	}
}

// closeh closes all the clients socket
func (h *Hub) closeh() {
	for _, c := range h.client {
		c.conn.Close()
	}
}
