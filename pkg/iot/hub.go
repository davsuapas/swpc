/*
 *   Copyright (c) 2022 ELIPCERO
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
	"bytes"
	"context"
	"encoding/json"
	"net"
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
	errHeartbeatTime      = "IOT Device heartbeat timeout"
)

const (
	infSendStateDesac = "Hub.An attempt has been made to send " +
		"a message but the hub is not in transmission mode"
	infCheckerDes = "The check timer is deactivated " +
		"because there are no activity"
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
	infSendAction     = "Hub.Send action to iot device"
	infDeviceID       = "DeviceID"
	infIOTDevice      = "IOT device"
	infClientID       = "ClientID"
	infClientCount    = "Client count"
	infState          = "state"
	infLastMsgDate    = "Last message date"
	infExpirationDate = "Expiration date"
	infActualDate     = "Actual date"
	infPrevState      = "Previous state"
	infLastNotify     = "Last notification date"
	infConfig         = "Config"
	infDeviceEmpty    = "Device empty"
	infTimeWindow     = "Transmission time window"
	infHBTimeoutCount = "Heartbeat timeout count"
	infHBInterval     = "Heartbeat interval"
)

// Time allowed to write a message to the websocket peer.
const writeWait = 2 * time.Second

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

// stateMessageType is the type that differentiates
// the message that is sent to the client
// In this case, send a message with hub state
const stateMessageType = "0"

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
// which can affects on the conduct of the device
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
	// CalibratingORP is the flag to calibrate the ORP
	CalibratingORP bool `json:"cgorp"`
	// TargetORP is the target value for the calibrating ORP
	TargetORP float32 `json:"torp"`
	// CalibrationORP is the value for the calibration ORP
	// When set to calibrate with the flag it will be
	// the initial value of the calibration.
	// When not set to calibration mode, it will be the calibrated value
	// obtained from the calibration.
	CalibrationORP float32 `json:"corp"`
	// StabilizationTimeORP is the time in seconds to stabilize
	// the calibration value
	StabilizationTimeORP int8 `json:"storp"`
	// IniSendTime is the range for initiating metric sends.
	// Format HH:mm
	IniSendTime string `json:"-"`
	// EndSendTime is the range for ending metric sends
	// Format HH:mm
	EndSendTime string `json:"-"`
}

// DeviceConfigDTO is the information
// to send to the device
type DeviceConfigDTO struct {
	DeviceConfig
	HBI  uint8 `json:"hbi"`
	HBTC uint8 `json:"hbtc"`
}

func (mc *DeviceConfigDTO) message() (string, error) {
	m, err := json.Marshal(mc)
	if err != nil {
		return "", errors.Wrap(err, "message")
	}

	return string(m), nil
}

type HeartbeatConfig struct {
	// HeartbeatInterval is the interval that
	// the iot device sends a ping for heartbeat
	// Zero does not check heartbeat
	HeartbeatInterval time.Duration `json:"heartbeatInterval"`
	// HeartbeatPingTime is the additional time it may take
	// for the ping to arrive.
	HeartbeatPingTime time.Duration `json:"heartbeatPingTime"`
	// HeartbeatTimeoutCount is the amount of timeout allowed
	// before closing the connection to the device.
	HeartbeatTimeoutCount uint8 `json:"heartbeatTimeoutCount"`
}

type Config struct {
	HeartbeatConfig `json:"heartbeatConfig"`
	DeviceConfig    `json:"deviceConfig"`
	// Location is the time zone
	Location *time.Location `json:"location"`
	// CommLatency is the time in seconds
	// before the communication goes to the inactive state
	CommLatency time.Duration `json:"commLatency"`
	// TaskTime defines how often the hub makes maintenance task
	TaskTime time.Duration `json:"taskTime"`
	// NotificationTime defines how often a notification is sent
	NotificationTime time.Duration `json:"notificationTime"`
}

func (c *Config) string() string {
	r, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(r)
}

type deviceMessage struct {
	Mtype deviceMessageType
	Msg   string
}

func (d *deviceMessage) message() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte(byte(d.Mtype))
	buffer.WriteString(d.Msg)

	return buffer.Bytes()
}

type Device struct {
	ID         string
	Connection *websocket.Conn
}

// deviceController manages a socket connection of a iot device.
type deviceController struct {
	HeartbeatConfig

	Device

	mtx    sync.RWMutex
	closed bool

	onRecieveMessage chan string
	onError          chan error

	chSetConn   chan *websocket.Conn
	chConnSetup chan struct{}

	chExit chan struct{}
}

func newDeviceController(h HeartbeatConfig) *deviceController {
	//
	d := &deviceController{
		HeartbeatConfig:  h,
		closed:           true,
		onRecieveMessage: make(chan string),
		onError:          make(chan error),
		chSetConn:        make(chan *websocket.Conn),
		chConnSetup:      make(chan struct{}),
		chExit:           make(chan struct{}),
	}

	d.recieveMessage()

	return d
}

// Link enganges the iot device connection to the controller
// If another connection is open, it is closed before
// assigning the new connection.
// Never use if the Stop() method is called,
// makes a new newDeviceController.
// I don't know how to protect it because it is for internal use
func (d *deviceController) Link(device Device) error {
	d.ID = device.ID

	err := d.Close()

	// Assigns the connection and waits for it to be completed
	d.chSetConn <- device.Connection
	<-d.chConnSetup

	return err
}

// SendConfig sends configuration changes
func (d *deviceController) SendConfig(cnf DeviceConfig) error {
	if d.IsClosed() {
		return nil
	}

	cnfdto := DeviceConfigDTO{
		HBI:          uint8(d.HeartbeatInterval.Seconds()),
		HBTC:         d.HeartbeatTimeoutCount,
		DeviceConfig: cnf,
	}

	msgc, err := cnfdto.message()
	if err != nil {
		return err
	}

	return d.send(deviceConfig, msgc)
}

// SendAction sends the next action to be taken by the device
func (d *deviceController) SendAction(a deviceAction) error {
	if d.IsClosed() {
		return nil
	}

	return d.send(action, strconv.Itoa(int(a)))
}

func (d *deviceController) send(t deviceMessageType, msgc string) error {
	dm := deviceMessage{
		Mtype: t,
		Msg:   msgc,
	}

	_ = d.Connection.SetWriteDeadline(time.Now().Add(writeWait))

	if err := d.Connection.WriteMessage(
		websocket.TextMessage,
		dm.message()); err != nil {
		//
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil
		}

		d.Close()

		return errors.Wrap(err, infIOTDevice)
	}

	return nil
}

func (d *deviceController) pingHandle(message string) error {
	err := d.Connection.WriteControl(
		websocket.PongMessage,
		[]byte(message),
		time.Now().Add(writeWait))

	var ne net.Error

	if errors.Is(err, websocket.ErrCloseSent) {
		return nil
	} else if errors.As(err, &ne) {
		return nil
	}

	if err == nil {
		_ = d.readDeadLine()
	}

	return errors.Wrap(err, "pingHandle")
}

// recieveMessage receives messages and
// if there is an error it closes the connection,
// notifies and it is put on hold for another connection.
func (d *deviceController) recieveMessage() {
	go func() {
		for {
			select {
			case <-d.chExit:
				return
			case conn := <-d.chSetConn:
				d.setClosed(false)
				d.Connection = conn
				d.Connection.SetPingHandler(d.pingHandle)
				d.chConnSetup <- struct{}{}

				for {
					if d.readMessage() {
						break
					}
				}
			}
		}
	}()
}

// readMessage reads the message. If an error is caused,
// the connection is closed, except if a timeout occurs
// because a ping is not received, in which case retries
// are made until the number of possible attempts is exceeded.
// If there are no errors, the message is sent through
// the onRecieveMessage channel.
// If true is returned, it exits
func (d *deviceController) readMessage() bool {
	timeout := d.readDeadLine()

	t, m, err := d.Connection.ReadMessage()
	if err != nil {
		errWrap := errors.Wrap(err, infIOTDevice)

		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			// A heartbeatTimeout has occurred
			errWrap = errors.Wrap(
				err,
				strings.Format(
					errHeartbeatTime,
					strings.FMTValue(infHBInterval, timeout.String()),
					strings.FMTValue(infHBTimeoutCount, string(d.HeartbeatTimeoutCount))))
		}

		closedManual := d.IsClosed()
		d.Close()

		// If it was closed because the close() method
		// was manually called, the close is not notified
		// because it is already manually controlled.
		// It is known because previously the close() method
		// sets closed=true.
		if !closedManual {
			d.onError <- errWrap
		}

		return true
	}

	if t == websocket.TextMessage {
		d.onRecieveMessage <- string(m)
	}

	return false
}

func (d *deviceController) readDeadLine() time.Duration {
	timeout := (d.HeartbeatInterval + d.HeartbeatPingTime) *
		time.Duration(d.HeartbeatTimeoutCount)
	_ = d.Connection.SetReadDeadline(time.Now().Add(timeout))

	return timeout
}

func (d *deviceController) setClosed(closed bool) {
	d.mtx.Lock()
	d.closed = closed
	d.mtx.Unlock()
}

func (d *deviceController) IsClosed() bool {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	return d.closed
}

// Close closes the socket and involves linking to another device
func (d *deviceController) Close() error {
	if d.Connection != nil && !d.IsClosed() {
		d.setClosed(true)

		_ = d.Connection.WriteControl(
			websocket.CloseMessage,
			[]byte{},
			time.Now().Add(writeWait))

		return errors.Wrap(d.Connection.Close(), infIOTDevice)
	}

	return nil
}

// Stop stops the controller and involves performing a NewController
func (d *deviceController) Stop() {
	d.Close()

	d.chExit <- struct{}{}

	close(d.chSetConn)
	close(d.chConnSetup)
	close(d.chExit)
	close(d.onError)
	close(d.onRecieveMessage)
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
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

	if err := c.conn.WriteMessage(
		websocket.TextMessage,
		[]byte(message)); err != nil {
		//
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil
		}

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

// TraceLevel is the trace level
type TraceLevel uint8

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production
	DebugLevel TraceLevel = iota
	// InfoLevel is the default logging priority
	InfoLevel
	// WarnLevel logs are more important than Info
	WarnLevel
	// NoneLevel is not applicable
	NoneLevel
)

type Trace struct {
	Level   TraceLevel
	Message string
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
// Messages sent to the clients are in the format where
// the type of message is differentiated by a number followed
// by the message. The hub has a reserved type to send the status.
// It starts with the number 0.
// iot devices can receive messages from the hub defined in deviceMessage
type Hub struct {
	// clients manages broadcast clients
	clients []Client
	// device manages the iot device
	device *deviceController

	levelTrace TraceLevel

	config Config

	regd chan Device

	reg   chan Client
	unreg chan string

	err     chan error
	trace   chan Trace
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
// all errors into the hub.
func NewHub(
	cnf Config,
	levelTrace TraceLevel,
	trace chan Trace,
	err chan error) *Hub {
	//
	return &Hub{
		clients:     []Client{},
		device:      newDeviceController(cnf.HeartbeatConfig),
		config:      cnf,
		regd:        make(chan Device),
		reg:         make(chan Client),
		unreg:       make(chan string),
		send:        make(chan string),
		sconfig:     make(chan DeviceConfig),
		levelTrace:  levelTrace,
		trace:       trace,
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

	if err := h.device.Link(device); err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				errlink,
				strings.FMTValue(infDeviceID, device.ID)))
	}

	if err := h.device.SendConfig(h.config.DeviceConfig); err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				errSendDevice,
				strings.FMTValue(infDeviceID, device.ID)))
	}

	h.setState(true)

	h.sendActionToDevice()

	h.tryReactiveTimerCheck(check)

	h.sendTrace(
		Trace{
			Level: InfoLevel,
			Message: strings.Format(
				infDeviceReg,
				strings.FMTValue(infDeviceID, device.ID)),
		})
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

	h.setState(false)
	h.tryReactiveTimerCheck(check)

	h.sendTrace(
		Trace{
			Level: InfoLevel,
			Message: strings.Format(
				infClientReg,
				strings.FMTValue(infClientID, client.id),
				strings.FMTValue(infClientCount, strconv.Itoa(len(h.clients))),
				strings.FMTValue(infState, StateString(h.state))),
		})
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
				strings.FMTValue(infClientCount, strconv.Itoa(len(h.clients)))))
	}

	h.setState(false)
	h.tryReactiveTimerCheck(check)

	h.sendTrace(
		Trace{
			Level: InfoLevel,
			Message: strings.Format(
				infClientUnRegd,
				strings.FMTValue(infClientID, id),
				strings.FMTValue(infClientCount, strconv.Itoa(len(h.clients))),
				strings.FMTValue(infState, StateString(h.state))),
		})
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

	h.sendMessage(message, errSendMsg)

	h.sendTrace(
		Trace{
			Level: DebugLevel,
			Message: strings.Format(
				infTransmit,
				strings.FMTValue(infState, StateString(h.state))),
		})
}

// sendConfigMessageToDevice sends the configuration
// you have changed to the iot device
func (h *Hub) sendConfigMessageToDevice(cnf DeviceConfig) {
	h.config.DeviceConfig = cnf

	if err := h.device.SendConfig(cnf); err != nil {
		h.err <- errors.Wrap(
			err,
			strings.Format(
				errSendDevice,
				strings.FMTValue(infDeviceID, h.device.ID)))
	}

	h.setState(false)

	h.sendTrace(
		Trace{
			Level: InfoLevel,
			Message: strings.Format(
				infConfigChanged,
				strings.FMTValue(infConfig, h.config.string())),
		})
}

func (h *Hub) processDeviceError(err error) {
	h.err <- err
	h.setState(false)
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
			h.setState(false)

			h.sendTrace(
				Trace{
					Level: InfoLevel,
					Message: strings.Format(
						infHubIdle,
						strings.FMTValue(infActualDate, time.Now().String()),
						strings.FMTValue(infLastMsgDate, h.lastMessage.String())),
				})
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
		strings.Concat(stateMessageType, strconv.Itoa(int(h.state))), errNotify) {
		//
		h.sendTrace(
			Trace{
				Level: InfoLevel,
				Message: strings.Format(
					infNotify,
					strings.FMTValue(infActualDate, time.Now().String()),
					strings.FMTValue(infLastNotify, h.notifySign.String()),
					strings.FMTValue(infState, StateString(state))),
			})
	}

	h.notifySign = time.Now()
}

// sendMessage sends messages to the client,
// If any are sent, returns true
func (h *Hub) sendMessage(message string, errMessage string) bool {
	var brokenclients []uint16

	sent := false

	for i, c := range h.clients {
		//nolint:gosec
		j := uint16(i)

		if err := c.sendMessage(message); err != nil {
			brokenclients = append(brokenclients, j)

			if err := h.closeClient(j); err != nil {
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
		h.sendTrace(
			Trace{
				Level: InfoLevel,
				Message: strings.Format(
					infArraySize,
					strings.FMTValue(infClientCount, strconv.Itoa(len(h.clients)))),
			})
	}

	return sent
}

// CheckTransWindow changes the state
// if there are no clients and there is a device connected
// can change the transmission window, therefore the state.
func (h *Hub) CheckTransWindow() {
	if h.clientsEmpty() && !h.device.IsClosed() {
		h.setState(false)
	}
}

// tryReactiveTimerCheck reactives the check timer
// if the state is not dead
func (h *Hub) tryReactiveTimerCheck(check *time.Timer) {
	if h.state == Dead {
		h.sendTrace(
			Trace{
				Level:   WarnLevel,
				Message: infCheckerDes,
			})

		return
	}

	check.Reset(h.config.TaskTime)
}

// onChangeState is launched when the hub state is changed
func (h *Hub) onChangeState(_ State) {
	h.sendActionToDevice()
}

func (h *Hub) sendActionToDevice() {
	var sent uint8 = 255

	if h.state == Asleep {
		sent = uint8(sleep)
	}

	if h.state == Active {
		sent = uint8(transmit)
	}

	if h.state == Inactive {
		sent = uint8(standby)
	}

	if sent != 255 {
		if err := h.device.SendAction(deviceAction(sent)); err != nil {
			h.setState(false)

			h.err <- errors.Wrap(
				err,
				strings.Format(
					errSendDevice,
					strings.FMTValue(infDeviceID, h.device.ID)))

			return
		}

		h.sendTrace(
			Trace{
				Level: InfoLevel,
				Message: strings.Format(
					infSendAction,
					strings.FMTValue(infState, strconv.Itoa(int(sent)))),
			})
	}
}

func (h *Hub) setBroadcastState() {
	h.sstate(func(_ bool, _ bool) {
		h.state = Broadcast
	}, false)
}

// setState sets the hub state
// The state is configured according to the clients,
// associated device and time window for transmission.
// There is one exception, which is broadast status.
// This state is configured when the information is sent
// to the client.
// If the state changes, the onChangeState() event is fired.
// If you do not want the event to fire set dd = false.
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
func (h *Hub) setState(cancelOnChange bool) {
	h.sstate(func(clientsEmpty bool, transmitWindow bool) {
		if clientsEmpty && h.device.IsClosed() {
			h.state = Dead

			return
		}

		if !clientsEmpty && !h.device.IsClosed() {
			h.state = Active

			return
		}

		if clientsEmpty && !h.device.IsClosed() && !transmitWindow {
			h.state = Asleep

			return
		}

		h.state = Inactive
	}, cancelOnChange)
}

func (h *Hub) sstate(fchangeState func(bool, bool), cancelOnChange bool) {
	previousState := h.state

	clientsEmpty := h.clientsEmpty()
	transmitWindow := h.transmitWindow()

	fchangeState(clientsEmpty, transmitWindow)

	if previousState != h.state {
		h.sendTrace(
			Trace{
				Level: WarnLevel,
				Message: strings.Format(
					infStateChanged,
					strings.FMTValue(infPrevState, StateString(previousState)),
					strings.FMTValue(infState, StateString(h.state)),
					strings.FMTValue(infClientCount, strconv.Itoa(len(h.clients))),
					strings.FMTValue(infDeviceEmpty, strconv.FormatBool(h.device.IsClosed())),
					strings.FMTValue(infTimeWindow, strconv.FormatBool(transmitWindow))),
			})

		if !cancelOnChange {
			h.onChangeState(previousState)
		}
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
	h.closeAllClient()
	h.device.Stop()

	if !check.Stop() {
		<-check.C
	}

	h.state = Closed

	close(h.err)
	close(h.trace)
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
		//nolint:gosec
		j := uint16(i)

		if c.expired() {
			deadClients = append(deadClients, j)

			clientID := h.clients[i].id

			if err := h.closeClient(j); err != nil {
				h.err <- errors.Wrap(
					err,
					strings.Format(
						errRemovingDiedClient,
						strings.FMTValue(infClientID, clientID)))
			}

			h.sendTrace(
				Trace{
					Level: InfoLevel,
					Message: strings.Format(
						infClientDied,
						strings.FMTValue(infClientID, clientID),
						strings.FMTValue(infExpirationDate, c.expiration.String()),
						strings.FMTValue(infActualDate, time.Now().String())),
				})
		}
	}

	if len(deadClients) > 0 {
		h.removeClientByPos(deadClients...)
		h.sendTrace(
			Trace{
				Level: InfoLevel,
				Message: strings.Format(
					infArraySize,
					strings.FMTValue(infClientCount, strconv.Itoa(len(h.clients)))),
			})
	}
}

// removeClient removes client by id
func (h *Hub) removeClient(clientID string) error {
	pos := h.findClient(clientID)
	if pos == 255 {
		return nil
	}

	err := h.closeClient(pos)
	h.removeClientByPos(pos)

	return err
}

// findClient seeks a client by ID. If the client is not found returns -1
func (h *Hub) findClient(clientID string) uint16 {
	for i, c := range h.clients {
		if c.id == clientID {
			//nolint:gosec
			return uint16(i)
		}
	}

	return 255
}

// removeClientByPos remove clients by position
func (h *Hub) removeClientByPos(pos ...uint16) {
	h.clients = arrays.Remove(h.clients, pos...)

	if h.clientsEmpty() {
		h.setState(false)
	}
}

// closeClient closes socket client
func (h *Hub) closeClient(pos uint16) error {
	return h.clients[pos].close()
}

func (h *Hub) clientsEmpty() bool {
	return len(h.clients) == 0
}

// closeAllClient closes all the clients socket
func (h *Hub) closeAllClient() {
	h.device.Close()

	for _, c := range h.clients {
		_ = c.close()
	}

	h.clients = []Client{}
}

func (h *Hub) sendTrace(t Trace) {
	if t.Level >= h.levelTrace {
		h.trace <- t
	}
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
