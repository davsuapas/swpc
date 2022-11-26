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
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/arrays"
	"github.com/swpoolcontroller/pkg/strings"
)

type clientStatus int

const (
	// activeComm the socket is running and the communication between the micro and server is activeComm
	activeComm clientStatus = iota
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
	// brokenComm the communication between the micro and server is break
	brokenComm bool
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
	clients []Client

	reg    chan Client
	unreg  chan string
	send   chan []byte
	infos  chan []string
	errors chan []error
	close  chan struct{}

	inactiveCommTime time.Duration
	breakCommTime    time.Duration
}

// NewHub builds Hub service.
// The inactiveCommTime establishes each time the communication between
// the micro and server is inactive and breakComm parameter establishes each time the communication
// between the micro and server is break
// The infos channel receives all infos into the hub
// The errors channel receives all errors into the hub
func NewHub(
	inactiveCommTime time.Duration,
	breakCommTime time.Duration,
	infos chan []string,
	errors chan []error) *Hub {
	return &Hub{
		clients:          []Client{},
		reg:              make(chan Client),
		unreg:            make(chan string),
		send:             make(chan []byte),
		infos:            infos,
		errors:           errors,
		close:            make(chan struct{}),
		inactiveCommTime: inactiveCommTime,
		breakCommTime:    breakCommTime,
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
		check := time.NewTimer(h.inactiveCommTime)

		for {
			select {
			case client := <-h.reg:
				h.registerClient(client, check)
			case id := <-h.unreg:
				if err := h.removeClient(id); err != nil {
					h.errors <- []error{errors.Wrap(
						err, strings.Concat(
							"Hub. Unregistering client: ", id,
							", count: ", strconv.Itoa(len(h.clients))))}
				}
				h.infos <- []string{strings.Concat("Hub. Client unregisted: ", id)}
			case message := <-h.send:
				h.sendMessage(message)
			case <-check.C:
				h.sendInactiveClientStatus()
				h.removeDeadClient()

				// If there are no clients, the timer is not activated and the CPU is saved.
				if len(h.clients) > 0 {
					check.Reset(h.inactiveCommTime)
				} else {
					h.infos <- []string{
						strings.Concat(
							"Hub. The check timer is deactivated because there are no clients")}
				}

			case <-h.close:
				h.closeh()

				if !check.Stop() {
					<-check.C
				}

				return
			}
		}
	}()
}

func (h *Hub) registerClient(client Client, check *time.Timer) {
	h.clients = append(h.clients, client)
	h.infos <- []string{
		strings.Concat(
			"Hub. Client registered: ", client.id,
			", count: ", strconv.Itoa(len(h.clients)))}

	// If there is a customer, it means that the timer was not activated before
	// and therefore I activate it.
	if len(h.clients) == 1 {
		check.Reset(h.inactiveCommTime)
	}
}

// sendMessage send message to the all clients registered. If sending the message throw a error, the client is removed
func (h *Hub) sendMessage(message []byte) {
	var errs []error

	var brokenclients []uint16

	for i, c := range h.clients {
		if c.brokenComm {
			continue
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			brokenclients = append(brokenclients, uint16(i))

			if err := h.closeClient(uint16(i)); err != nil {
				errs = append(errs, errors.Wrap(err, "Hub. Sending a message"))
			}

			errs = append(
				errs,
				errors.Wrap(
					err,
					strings.Concat("Hub. Sending message to: ", c.id, ". The client will be removed")))
		} else {
			h.clients[i].lastMessage = time.Now()
		}
	}

	if len(errs) > 0 {
		h.errors <- errs
	}

	if len(brokenclients) > 0 {
		h.clients = arrays.Remove(h.clients, brokenclients...)
	}
}

// sendInactiveClientStatus send the client status when is inactive.
// If sending the message throw a error, the client is removed
func (h *Hub) sendInactiveClientStatus() {
	var (
		infos []string
		errs  []error
	)

	var brokenclients []uint16

	for i, client := range h.clients {
		if client.brokenComm {
			continue
		}

		status := h.commStatus(client)
		if status == activeComm {
			continue
		}

		if status == breakComm {
			h.clients[i].brokenComm = true
		}

		if err := client.conn.WriteMessage(websocket.TextMessage, []byte{byte(status)}); err != nil {
			brokenclients = append(brokenclients, uint16(i))

			if err := h.closeClient(uint16(i)); err != nil {
				errs = append(errs, errors.Wrap(err, "Hub. Sending inactive/break client status"))
			}

			errs = append(
				errs,
				errors.Wrap(
					err,
					strings.Concat(
						"Hub. Sending inactive/break client status: ", client.id+". The client will be removed")))
		} else {
			infos = append(
				infos,
				strings.Concat(
					"Hub. Sending inactive client status: ", client.id,
					", status: ", strconv.Itoa(int(status)),
					", Last message date: ", client.lastMessage.String(),
					", Actual date: ", time.Now().String()))
		}
	}

	if len(brokenclients) > 0 {
		h.clients = arrays.Remove(h.clients, brokenclients...)
	}

	if len(errs) > 0 {
		infos = append(
			infos,
			strings.Concat("Hub. Size of the client array after sending inactive client status: ",
				strconv.Itoa(len(h.clients))))
		h.errors <- errs
	}

	if len(infos) > 0 {
		h.infos <- infos
	}
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
						strings.Concat(
							"Hub. Removing died client by expiration: ", clientID+". The client will be removed")))
			}

			infos = append(
				infos,
				strings.Concat(
					"Hub. Client died by expiration: ", clientID,
					", Expiration date: ", c.expiration.String(),
					", Actual date: ", time.Now().String()))
		}
	}

	if len(errs) > 0 {
		h.errors <- errs
	}

	if len(deadClients) > 0 {
		h.clients = arrays.Remove(h.clients, deadClients...)
	}

	if len(infos) > 0 {
		infos = append(
			infos,
			strings.Concat(
				"Hub. Size of the client array after removing expired clients: ",
				strconv.Itoa(len(h.clients))))
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
	if pos != -1 {
		return nil
	}

	posr := uint16(pos)

	if err := h.closeClient(posr); err != nil {
		return err
	}

	h.clients = arrays.Remove(h.clients, posr)

	return nil
}

// closeClient closes socket client
func (h *Hub) closeClient(pos uint16) error {
	clientID := h.clients[pos].id

	if err := h.clients[pos].conn.Close(); err != nil {
		return errors.Wrap(err, strings.Concat("Hub. Closing client:", clientID))
	}

	return nil
}

func (h *Hub) commStatus(c Client) clientStatus {
	if c.lastMessage.Add(h.breakCommTime).Before(time.Now()) {
		return breakComm
	} else if c.lastMessage.Add(h.inactiveCommTime).Before(time.Now()) {
		return inactiveComm
	}

	return activeComm
}

// closeh closes all the clients socket
func (h *Hub) closeh() {
	for _, c := range h.clients {
		c.conn.Close()
	}
}
