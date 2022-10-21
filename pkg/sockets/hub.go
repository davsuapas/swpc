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
	// InactiveComm the socket is running and the communication between the micro and server is inactive
	InactiveComm
	// StoppedComm the socket is running and the communication between the micro and server is temporarily broken
	StoppedComm
	// BreakComm the socket is running and the communication between the micro and server is break
	BreakComm
)

type Client struct {
	id         string
	conn       *websocket.Conn
	status     Status
	expiration time.Time
}

func NewClient(id string, conn *websocket.Conn, expiration time.Duration) Client {
	return Client{
		id:         id,
		conn:       conn,
		status:     Active,
		expiration: time.Now().Add(expiration),
	}
}

type Hub struct {
	checkExp time.Duration
	client   []Client
	reg      chan Client
	unreg    chan string
	send     chan []byte
	errors   chan []error
}

func (h *Hub) Register(client Client) {
	h.reg <- client
}

func (h *Hub) Unregister(id string) {
	h.unreg <- id
}

func (h *Hub) Send(message []byte) {
	h.send <- message
}

func (h *Hub) Run() {
	select {
	case client := <-h.reg:
		h.client = append(h.client, client)
	case id := <-h.unreg:
		if err := h.removeClient(id); err != nil {
			h.errors <- []error{errors.Wrap(err, strings.Concat("Removing client: ", id))}
		}
	case message := <-h.send:
		h.sendMessage(message)
	case <-time.After(h.checkExp):
		h.removeDeadClient()
	}
}

func (h *Hub) sendMessage(message []byte) {
	var errs []error

	for i, c := range h.client {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			h.client[i].status = Break

			errs = append(errs, errors.Wrap(err, strings.Concat("Sending message to: ", c.id)))
		}
	}

	if len(errs) > 0 {
		h.errors <- errs
	}
}

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
