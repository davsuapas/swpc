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

package iot

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/arrays"
	"github.com/swpoolcontroller/pkg/strings"
)

const (
	errlink               = "Link device"
	errSendMsg            = "Sending a message to the client"
	errSendDevice         = "Sending a message to device"
	errNotify             = "Notifying state. The client will be removed"
	errRemovingDiedClient = "Removing died client by expiration"
	errClosingClient      = "Closing client"
	errParseStartTime     = "Parser transmission start time"
	errParseEndTime       = "Parser transmission end time"
)

const (
	infSendStateDesac = "Hub.An attempt has been made to send " +
		"a message but the hub is not in transmission mode"
	infCheckerDes = "The check timer is deactivated " +
		"because there are no clients"
	infDeviceReg      = "Hub.Device iot registered"
	infDeviceExists   = "Hub.Unregistering device when trying to register"
	infClientReg      = "Hub.Client registered"
	infClientExists   = "Hub.Unregister existing client when trying to register"
	infClientUnReg    = "Hub.Unregister client"
	infClientUnRegd   = "Hub.Client unregisted"
	infTransmit       = "Hub.Sending information to the client"
	infHubIdle        = "Hub.Broadcast state but no communication is detected"
	infClientDied     = "Hub.Client died by expiration"
	infArraySize      = "Hub.Array size after removing expired clients"
	infNotify         = "Hub.Notifying state"
	infConfigChanged  = "Hub.The configuration has been changed"
	infStateChanged   = "Hub.The state has been changed"
	infDeviceID       = "DeviceID"
	infClientID       = "ClientID"
	infLength         = "Length"
	infState          = "state"
	infLastMsgDate    = "Last message date"
	infExpirationDate = "Expiration date"
	infActualDate     = "Actual date"
	infPrevState      = "Previous state"
	infLastNotify     = "Last notification date"
	infConfig         = "Config"
	infDeviceEmpty    = "Device empty"
	infClientEmpty    = "Clients empty"
	infTimeWindow     = "Transmission time window"
)

const layaoutTime = "15:04"

// state are the communication state between the device iot and the hub
type State uint8

const (
	// Dead there are no clients and no device independent
	// of the transmission window.
	Dead State = iota
	// Inactive there is not clients or device connected
	Inactive
	// Active the hub is ready to transmit and receive
	// because there are clients and device.
	// You may not be in the transmission time window,
	// but if there are clients and device
	// you can always transmit
	Active
	// Broadcast the information is broadcasting
	// between clients and iot device
	Broadcast
	// Asleep there is no time window for transmitting,
	// and no clients to transmit or receive information.
	// In this case, the device is put to sleep to save energy.
	Asleep
	// The hub communication channel is closed
	Closed
)

// DeviceMessageType is the type that differentiates
// the message that is sent to the iot device
type deviceMessageType uint8

const (
	deviceConfig deviceMessageType = iota
	action
)

// deviceAction are the Actions communication between
// the device and server
type deviceAction uint8

const (
	// Sleep puts the micro controller to sleep
	sleep deviceAction = iota
	// Transmit puts the micro controller to Transmit metrics
	transmit
	// standby puts the micro controller to standby
	// until there are customers
	standby
)

// DeviceConfig is the configuration information
// that is sent to the device.
type DeviceConfig struct {
	// WakeUpTime is the time set to wake up
	// the micro-controller in minutes.
	WakeUpTime uint8 `json:"wut"`
	// CollectMetricsTime defines how often metrics
	// are collected in milliseconds
	CollectMetricsTime int `json:"cmt"`
	// Buffer is the time in seconds to store metrics
	// before sending to the hub
	Buffer uint8 `json:"buffer"`
}

func (mc *DeviceConfig) message() (string, error) {
	m, err := json.Marshal(mc)
	if err != nil {
		return "", errors.Wrap(err, "message")
	}

	return string(m), nil
}

type Config struct {
	DeviceConfig `json:"deviceConfig"`
	// Location is the time zone
	Location *time.Location `json:"location"`
	// CommLatency is the time in seconds
	// before the communication goes to the inactive state
	CommLatency time.Duration `json:"commLatency"`
	// TaskTime defines how often the hub makes maintenance task
	TaskTime time.Duration `json:"taskTime"`
	// NotificationTime defines how often a notification is sent
	NotificationTime time.Duration `json:"notificationTime"`
	// IniSendTime is the range for initiating metric sends.
	// Format HH:mm
	IniSendTime string `json:"iniSendTime"`
	// EndSendTime is the range for ending metric sends
	// Format HH:mm
	EndSendTime string `json:"endSendTime"`
}

func (c *Config) string() string {
	r, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(r)
}

type deviceMessage struct {
	Mtype deviceMessageType `json:"t"`
	Msg   string            `json:"m"`
}

func (d *deviceMessage) message() ([]byte, error) {
	r, err := json.Marshal(d)
	if err != nil {
		return []byte{}, errors.Wrap(err, "message")
	}

	return r, nil
}

type Device struct {
	ID         string
	Connection *websocket.Conn
}

// deviceController manages a socket connection of a iot device.
type deviceController struct {
	Device

	mtx    sync.RWMutex
	closed bool

	onRecieveMessage chan string
	onError          chan error

	closeChannel chan struct{}
}

func newDeviceController() deviceController {
	return deviceController{
		closed:           true,
		onRecieveMessage: make(chan string),
		onError:          make(chan error),
		closeChannel:     make(chan struct{}),
	}
}

// link enganges the iot device connection to the controller
func (d *deviceController) link(device Device) error {
	//
	err := d.close()

	d.ID = device.ID
	d.Connection = device.Connection

	d.setClosed(false)

	go d.recieveMessage()

	return err
}

func (d *deviceController) sendConfig(cnf DeviceConfig) error {
	if d.isClosed() {
		return nil
	}

	msgc, err := cnf.message()
	if err != nil {
		return err
	}

	return d.send(deviceConfig, msgc)
}

func (d *deviceController) sendAction(a deviceAction) error {
	if d.isClosed() {
		return nil
	}

	return d.send(action, strconv.Itoa(int(a)))
}

func (d *deviceController) send(t deviceMessageType, msgc string) error {
	dm := deviceMessage{
		Mtype: t,
		Msg:   msgc,
	}

	msgd, err := dm.message()
	if err != nil {
		return errors.Wrap(err, "deviceMessage")
	}

	return errors.Wrap(
		d.Connection.WriteMessage(websocket.TextMessage, msgd),
		"device")
}

// recieveMessage receives messages and
// if there is an error it closes the connection,
// notifies and exits.
func (d *deviceController) recieveMessage() {
	for {
		select {
		case <-d.closeChannel:
			return
		default:
			t, m, err := d.Connection.ReadMessage()
			if err != nil {
				d.setClosed(true)
				d.Connection.Close()
				d.onError <- err

				return
			}

			if t != websocket.TextMessage {
				continue
			}

			d.onRecieveMessage <- string(m)
		}
	}
}

func (d *deviceController) setClosed(closed bool) {
	d.mtx.Lock()
	d.closed = closed
	d.mtx.Unlock()
}

func (d *deviceController) isClosed() bool {
	d.mtx.RLock()
	c := d.closed
	d.mtx.RUnlock()

	return c
}

func (d *deviceController) close() error {
	if d.Connection != nil && !d.isClosed() {
		d.setClosed(true)
		// Causes the routine for receiving messages to terminate.
		d.closeChannel <- struct{}{}

		return errors.Wrap(d.Connection.Close(), "device")
	}

	return nil
}

// Client manages a socket connection to consume or sending
// the messages from or to a device iot.
type Client struct {
	// id identifies the session
	id         string
	conn       *websocket.Conn
	expiration time.Time
}

func NewClient(
	id string,
	conn *websocket.Conn,
	expiration time.Duration) Client {
	//
	return Client{
		id:         id,
		conn:       conn,
		expiration: time.Now().Add(expiration),
	}
}

func (c *Client) sendMessage(message string) error {
	if err := c.conn.WriteMessage(
		websocket.TextMessage,
		[]byte(message)); err != nil {
		return errors.Wrap(err, "client")
	}

	return nil
}

func (c *Client) expired() bool {
	return c.expiration.Before(time.Now())
}

func (c *Client) close() error {
	if err := c.conn.Close(); err != nil {
		return errors.Wrap(
			err,
			strings.Format(
				errClosingClient,
				strings.FMTValue(infClientID, c.id)))
	}

	return nil
}

// The hub handles bi-directional messaging from an iot device
// to a set of socket clients.
// The hub registers clients and one device across of the reg channel,
// unregisters clients, broadcast messages and returns errors to the sender.
// Also the hub checks communication state, socket, etc.
// The system can set a time window for the transmission of information.
// If there are connected clients, it is always transmitted,
// even if it is not within this window. If there are no connected clients
// and you are not within this window,
// the device goes to sleep until the next check.
// Messages sent to clients have the following simple format:
// "0:state", where "state" can be 0 (inactive) and 1 (active),
// or "1:message", where "message" is the message to client.
// sent by iot device
type Hub struct {
	// clients manages broadcast clients
	clients []Client
	// device manages the iot device
	device deviceController

	config Config

	regd chan Device

	reg   chan Client
	unreg chan string

	err     chan error
	info    chan string
	send    chan string
	sconfig chan DeviceConfig
	statec  chan chan State
	closec  chan struct{}

	// notifySign Controls how often the hub sends
	// notifications to the client due to lack of communication
	notifySign time.Time
	// lastMessage the time when the last messages was sended
	lastMessage time.Time

	state State
}

// NewHub builds hub service. The config establishes
// how the hub should behave. The infos channel receives
// all infos into the hub. The errors channel receives
// all errors into the hub
func NewHub(
	cnf Config,
	info chan string,
	err chan error) *Hub {
	//
	return &Hub{
		clients:     []Client{},
		device:      newDeviceController(),
		config:      cnf,
		regd:        make(chan Device),
		reg:         make(chan Client),
		unreg:       make(chan string),
		send:        make(chan string),
		sconfig:     make(chan DeviceConfig),
		info:        info,
		err:         err,
		statec:      make(chan chan State),
		closec:      make(chan struct{}),
		lastMessage: time.Time{},
		notifySign:  time.Now(),
		state:       Dead,
	}
}

// Register registers iot device into the hub
func (h *Hub) RegisterDevice(d Device) {
	h.regd <- d
}

// Register registers client into the hub
func (h *Hub) RegisterClient(c Client) {
	h.reg <- c
}

// Unregister unregisters client into the hub
func (h *Hub) UnregisterClient(id string) {
	h.unreg <- id
}

// Config sends the config to the hub
func (h *Hub) Config(cnf DeviceConfig) {
	h.sconfig <- cnf
}

// Status request hub state via channel
func (h *Hub) State(ctx context.Context) (State, error) {
	resp := make(chan State)

	select {
	case h.statec <- resp:
	case <-ctx.Done():
		return Inactive, errors.Wrap(ctx.Err(), "request")
	}

	select {
	case state := <-resp:
		return state, nil
	case <-ctx.Done():
		return Inactive, errors.Wrap(ctx.Err(), "resp")
	}
}

// Stop finishes the hub.
// The force param closes all channels and force to exist of the goroutine
func (h *Hub) Stop() {
	h.closec <- struct{}{}
}

// Run manages the lifecycle of the hub,
// controls the associated device and the clients
// consuming the messages from the device
func (h *Hub) Run() { //nolint:cyclop
	go func() {
		check := time.NewTimer(h.config.TaskTime)

		for {
			select {
			case device := <-h.regd:
				h.registerDevice(device, check)
			case client := <-h.reg:
				h.registerClient(client, check)
			case id := <-h.unreg:
				h.unregister(id, check)
			case m := <-h.device.onRecieveMessage:
				h.sendMessageToClients(m)
			case err := <-h.device.onError:
				h.processDeviceError(err)
			case resps := <-h.statec:
				resps <- h.state
			case cnf := <-h.sconfig:
				h.sendConfigMessageToDevice(cnf)
			case <-check.C:
				h.idleBroadcast()
				h.notifyState()
				h.removeDeadClient()
				h.CheckTransWindow()
				h.tryReactiveTimerCheck(check)
			case <-h.closec:
				h.close(check)

				return
			}
		}
	}()
}

func (h *Hub) registerDevice(device Device, check *time.Timer) {
	if h.state == Closed {
		return
	}

	if err := h.device.link(device); err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				errlink,
				strings.FMTValue(infDeviceID, device.ID)))
	}

	if err := h.device.sendConfig(h.config.DeviceConfig); err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				errSendDevice,
				strings.FMTValue(infDeviceID, device.ID)))
	}

	h.setState()
	h.tryReactiveTimerCheck(check)

	h.info <- strings.Format(
		infDeviceReg,
		strings.FMTValue(infDeviceID, device.ID))
}

func (h *Hub) registerClient(client Client, check *time.Timer) {
	if h.state == Closed {
		return
	}

	// If client exists is removed
	if err := h.removeClient(client.id); err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				infClientExists,
				strings.FMTValue(infClientID, client.id)))
	}

	h.clients = append(h.clients, client)

	h.setState()
	h.tryReactiveTimerCheck(check)

	h.info <- strings.Format(
		infClientReg,
		strings.FMTValue(infClientID, client.id),
		strings.FMTValue(infLength, strconv.Itoa(len(h.clients))),
		strings.FMTValue(infState, StateString(h.state)))
}

func (h *Hub) unregister(id string, check *time.Timer) {
	if h.state == Closed {
		return
	}

	// removeClient change state
	if err := h.removeClient(id); err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				infClientUnReg,
				strings.FMTValue(infLength, strconv.Itoa(len(h.clients)))))
	}

	h.setState()
	h.tryReactiveTimerCheck(check)

	h.info <- strings.Format(
		infClientUnRegd,
		strings.FMTValue(infClientID, id),
		strings.FMTValue(infLength, strconv.Itoa(len(h.clients))),
		strings.FMTValue(infState, StateString(h.state)))
}

// sendMessageToClients send message to the all clients registered.
// If sending the message throw a error, the client is removed
func (h *Hub) sendMessageToClients(message string) {
	if !(h.state == Active || h.state == Broadcast) {
		return
	}

	h.lastMessage = time.Now()
	// Update the time, so if the communication fails,
	// notification will be made according to the preset time
	h.notifySign = time.Now()
	h.setBroadcastState()

	h.sendMessage(strings.Concat("1:", message), errSendMsg)

	h.info <- strings.Format(
		infTransmit,
		strings.FMTValue(infState, StateString(h.state)))
}

// sendConfigMessageToDevice sends the configuration
// you have changed to the iot device
func (h *Hub) sendConfigMessageToDevice(cnf DeviceConfig) {
	h.config.DeviceConfig = cnf

	if err := h.device.sendConfig(cnf); err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				errSendDevice,
				strings.FMTValue(infDeviceID, h.device.ID)))
	}

	h.info <- strings.Format(
		infConfigChanged,
		strings.FMTValue(infConfig, h.config.string()))
}

func (h *Hub) processDeviceError(err error) {
	h.err <- err
	h.setState()
}

// idleBroadcast checks whether there is still activity
// since the last communication
func (h *Hub) idleBroadcast() {
	if h.state == Broadcast {
		// The idle time is the sum of the time it takes
		// for the sender to create the buffer
		// and a possible latency time
		idleTime := time.Duration(h.config.Buffer)*time.Second +
			h.config.CommLatency

		if h.lastMessage.Add(idleTime).Before(time.Now()) {
			h.setState()

			h.info <- strings.Format(
				infHubIdle,
				strings.FMTValue(infActualDate, time.Now().String()),
				strings.FMTValue(infLastMsgDate, h.lastMessage.String()))
		}
	}
}

// notifyState sends the state to the web client
// (Only for states Inactive or Active)
// When it switches to an active state,
// it starts to count down the time.
// If no information is transmitted for a preset time
// in the configuration,
// the clients shall be notified of the status.
func (h *Hub) notifyState() {
	if !(h.state == Inactive || h.state == Active) {
		return
	}

	timeout := h.notifySign.Add(h.config.NotificationTime)

	// If the time for the next notification has not elapsed,
	// it does not send the next notification
	if timeout.After(time.Now()) {
		return
	}

	// The state may be changed by sendmessage
	state := h.state

	if h.sendMessage(
		strings.Concat("0:", strconv.Itoa(int(h.state))), errNotify) {
		//
		h.info <- strings.Format(
			infNotify,
			strings.FMTValue(infActualDate, time.Now().String()),
			strings.FMTValue(infLastNotify, h.notifySign.String()),
			strings.FMTValue(infState, StateString(state)))
	}

	h.notifySign = time.Now()
}

// sendMessage sends messages to the client,
// If any are sent, returns true
func (h *Hub) sendMessage(message string, errMessage string) bool {
	var brokenclients []uint16

	sent := false

	for i, c := range h.clients {
		if err := c.sendMessage(message); err != nil {
			brokenclients = append(brokenclients, uint16(i))

			if err := h.closeClient(uint16(i)); err != nil {
				h.err <- errors.Wrap(err, errMessage)
			}

			h.err <- errors.Wrap(
				err,
				strings.Format(errMessage, strings.FMTValue(infClientID, c.id)))

			continue
		}

		sent = true
	}

	if len(brokenclients) > 0 {
		h.removeClientByPos(brokenclients...)
		h.info <- strings.Format(
			infArraySize,
			strings.FMTValue(infLength, strconv.Itoa(len(h.clients))))
	}

	return sent
}

// CheckTransWindow changes the state
// if there are no clients and there is a device connected
// can change the transmission window, therefore the state.
func (h *Hub) CheckTransWindow() {
	if h.clientsEmpty() && !h.device.isClosed() {
		h.setState()
	}
}

// tryReactiveTimerCheck reactives the check timer
// if the state is not dead
func (h *Hub) tryReactiveTimerCheck(check *time.Timer) {
	if h.state == Dead {
		h.info <- infCheckerDes

		return
	}

	check.Reset(h.config.TaskTime)
}

// onChangeState is launched when the hub state is changed
func (h *Hub) onChangeState(_ State) {
	var err error

	if h.state == Asleep {
		err = h.device.sendAction(sleep)
	}

	if h.state == Active {
		err = h.device.sendAction(transmit)
	}

	if h.state == Inactive {
		err = h.device.sendAction(standby)
	}

	if err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				errSendDevice,
				strings.FMTValue(infDeviceID, h.device.ID)))
	}
}

func (h *Hub) setBroadcastState() {
	h.sstate(func(_ bool, _ bool) {
		h.state = Broadcast
	})
}

// setState sets the hub state
// The state is configured according to the clients,
// associated device and time window for transmission.
// There is one exception, which is broadast status.
// This state is configured when the information is sent
// to the client.
//
// State table
// -----------
//
// Clients		Device iot		Transmission time Windows			State
// -------------------------------------------------------------
//
//	0						0										0									Dead
//	0						1										0									Asleep
//	1						0										0									Inactive
//	1						1										0									Active
//	0						0										1									Dead
//	0						1										1									Inactive
//	1						0										1									Inactive
//	1						1										1									Active
func (h *Hub) setState() {
	h.sstate(func(clientsEmpty bool, transmitWindow bool) {
		if clientsEmpty && h.device.isClosed() {
			h.state = Dead

			return
		}

		if !clientsEmpty && !h.device.isClosed() {
			h.state = Active

			return
		}

		if clientsEmpty && !h.device.isClosed() && !transmitWindow {
			h.state = Asleep

			return
		}

		h.state = Inactive
	})
}

func (h *Hub) sstate(fchangeState func(bool, bool)) {
	previousState := h.state

	clientsEmpty := h.clientsEmpty()
	transmitWindow := h.transmitWindow()

	fchangeState(clientsEmpty, transmitWindow)

	if previousState != h.state {
		h.info <- strings.Format(
			infStateChanged,
			strings.FMTValue(infPrevState, StateString(previousState)),
			strings.FMTValue(infState, StateString(h.state)),
			strings.FMTValue(infClientEmpty, strconv.FormatBool(clientsEmpty)),
			strings.FMTValue(infDeviceEmpty, strconv.FormatBool(h.device.isClosed())),
			strings.FMTValue(infTimeWindow, strconv.FormatBool(transmitWindow)))

		h.onChangeState(previousState)
	}
}

// transmitWindow indicates whether you are within
// the time window for transmitting information.
func (h *Hub) transmitWindow() bool {
	iniTime, err := time.Parse(layaoutTime, h.config.IniSendTime)
	if err != nil {
		h.err <- errors.Wrap(err, errParseStartTime)

		return false
	}

	endTime, err := time.Parse(layaoutTime, h.config.EndSendTime)
	if err != nil {
		h.err <- errors.Wrap(err, errParseEndTime)

		return false
	}

	// now is the current date and time within the time zone
	// defined by the configuration
	var n time.Time
	if h.config.Location == nil {
		n = time.Now()
	} else {
		n = time.Now().In(h.config.Location)
	}

	now := time.Date(
		iniTime.Year(),
		iniTime.Month(),
		iniTime.Day(),
		n.Hour(),
		n.Minute(),
		iniTime.Second(),
		iniTime.Nanosecond(),
		iniTime.Location())

	if now.After(iniTime) && now.Before(endTime) {
		return true
	}

	return false
}

func (h *Hub) close(check *time.Timer) {
	h.closeh()

	if !check.Stop() {
		<-check.C
	}

	h.state = Closed

	close(h.err)
	close(h.info)
	close(h.reg)
	close(h.unreg)
	close(h.send)
	close(h.sconfig)
	close(h.statec)
	close(h.closec)
}

// removeDeadClient removes died clients
func (h *Hub) removeDeadClient() {
	var deadClients []uint16

	for i, c := range h.clients {
		if c.expired() {
			deadClients = append(deadClients, uint16(i))

			clientID := h.clients[i].id

			if err := h.closeClient(uint16(i)); err != nil {
				h.err <- errors.Wrap(
					err,
					strings.Format(
						errRemovingDiedClient,
						strings.FMTValue(infClientID, clientID)))
			}

			h.info <- strings.Format(
				infClientDied,
				strings.FMTValue(infClientID, clientID),
				strings.FMTValue(infExpirationDate, c.expiration.String()),
				strings.FMTValue(infActualDate, time.Now().String()))
		}
	}

	if len(deadClients) > 0 {
		h.removeClientByPos(deadClients...)
		h.info <- strings.Format(
			infArraySize,
			strings.FMTValue(infLength, strconv.Itoa(len(h.clients))))
	}
}

// removeClient removes client by id
func (h *Hub) removeClient(clientID string) error {
	pos := h.findClient(clientID)
	if pos == -1 {
		return nil
	}

	posr := uint16(pos)

	err := h.closeClient(posr)
	h.removeClientByPos(posr)

	return err
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

// removeClientByPos remove clients by position
func (h *Hub) removeClientByPos(pos ...uint16) {
	h.clients = arrays.Remove(h.clients, pos...)

	if h.clientsEmpty() {
		h.setState()
	}
}

// closeClient closes socket client
func (h *Hub) closeClient(pos uint16) error {
	return h.clients[pos].close()
}

func (h *Hub) clientsEmpty() bool {
	return len(h.clients) == 0
}

// closeh closes all the clients socket
func (h *Hub) closeh() {
	h.device.close()

	for _, c := range h.clients {
		_ = c.close()
	}

	h.clients = []Client{}
}

func StateString(s State) string {
	switch s {
	case Dead:
		return "Dead"
	case Inactive:
		return "Inactive"
	case Active:
		return "Active"
	case Asleep:
		return "Asleep"
	case Broadcast:
		return "Broadcast"
	case Closed:
		return "Closed"
	}

	return "None"
}
