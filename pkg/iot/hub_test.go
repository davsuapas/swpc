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

package iot_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/pkg/iot"
)

type trace struct {
	Infos   []string
	Errors  []string
	CHInfo  chan string
	CHError chan error
}

func newTrace() *trace {
	t := &trace{
		CHInfo:  make(chan string),
		CHError: make(chan error),
	}

	t.register()

	return t
}

func (t *trace) register() {
	go func() {
		for {
			select {
			case e, ok := <-t.CHError:
				if ok {
					t.Errors = append(t.Errors, e.Error())
				}
			case i, ok := <-t.CHInfo:
				if ok {
					t.Infos = append(t.Infos, i)
				}
			}
		}
	}()
}

func TestHub_LifecycleNoTransmissionTimeWindows(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             5,
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Second,
		NotificationTime: 50 * time.Second,
		IniSendTime:      "00:01",
		EndSendTime:      "00:02",
	}

	trace := newTrace()

	wscs, wscc, err := newWS()
	if !assert.NoError(t, err, "New web client socket") {
		return
	}
	defer wscs.Close()
	defer wscc.Close()

	client := iot.NewClient("c1", wscs, 10*time.Minute)

	wsds, wsdc, err := newWS()
	if !assert.NoError(t, err, "New web device socket") {
		return
	}
	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, trace.CHInfo, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterClient(client)
	hub.RegisterDevice(device)

	if testHubBroadcastMessageToClientsFromNoTimeWindow(t, wsdc, wscc, hub) {
		return
	}

	if testHubBroadcastToInactiveFromNoTimeWindow(t, wsdc, hub) {
		return
	}

	wsds1, wsdc1, err := newWS()
	if !assert.NoError(t, err, "New web device socket 1") {
		return
	}
	defer wsds1.Close()
	defer wsdc1.Close()

	device1 := iot.Device{
		ID:         "d2",
		Connection: wsds1,
	}
	hub.RegisterDevice(device1)

	if testHubInactiveToActiveToSleepFromNoTimeWindow(t, wsdc1, hub) {
		return
	}

	assert.Subset(
		t,
		trace.Infos,
		[]string{
			"Hub.Client registered (ClientID: c1, Length: 1, state: Inactive, )",
			"Hub.Device iot registered (DeviceID: d1, )",
			"Hub.Sending information to the client (state: Broadcast, )",
			"Hub.Device iot registered (DeviceID: d2, )",
		},
		"Info")

	assert.Len(t, trace.Errors, 1, "Errors")
}

func testHubBroadcastMessageToClientsFromNoTimeWindow(
	t *testing.T,
	wsdc *websocket.Conn,
	wscc *websocket.Conn,
	hub *iot.Hub) bool {
	//
	t.Helper()

	err := wsdc.WriteMessage(websocket.TextMessage, []byte("message"))
	if !assert.NoError(t, err, "Write device message") {
		return true
	}

	msgd := readMessages(wsdc, 2)
	msgc := readMessages(wscc, 1)

	if assertState(t, hub, iot.Broadcast) {
		return true
	}

	assert.Equal(
		t,
		[]string{
			"{\"t\":0,\"m\":\"{\\\"wut\\\":1,\\\"cmt\\\":800,\\\"buffer\\\":5}\"}",
			"{\"t\":1,\"m\":\"1\"}"},
		msgd,
		"Device messages received")

	assert.Equal(t, []string{"1:message"}, msgc, "Client messages received")

	return false
}

func testHubBroadcastToInactiveFromNoTimeWindow(
	t *testing.T,
	wsdc *websocket.Conn,
	hub *iot.Hub) bool {
	//
	t.Helper()

	wsdc.Close()

	time.Sleep(5 * time.Millisecond)

	return assertState(t, hub, iot.Inactive)
}

func testHubInactiveToActiveToSleepFromNoTimeWindow(
	t *testing.T,
	wsdc *websocket.Conn,
	hub *iot.Hub) bool {
	//
	t.Helper()

	hub.UnregisterClient("c1")

	msgd := readMessages(wsdc, 3)

	if assertState(t, hub, iot.Asleep) {
		return true
	}

	assert.Equal(
		t,
		[]string{
			"{\"t\":0,\"m\":\"{\\\"wut\\\":1,\\\"cmt\\\":800,\\\"buffer\\\":5}\"}",
			"{\"t\":1,\"m\":\"1\"}",
			"{\"t\":1,\"m\":\"0\"}"},
		msgd,
		"Device messages sleep received")

	return false
}

func TestHub_InactiveStateTransmissionTimeWindowsStandbyAction(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             5,
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Second,
		NotificationTime: 50 * time.Second,
		IniSendTime:      "00:00",
		EndSendTime:      "23:59",
	}

	trace := newTrace()

	wsds, wsdc, err := newWS()
	if !assert.NoError(t, err, "New web device socket") {
		return
	}
	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, trace.CHInfo, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterDevice(device)

	msgd := readMessages(wsdc, 2)

	if assertState(t, hub, iot.Inactive) {
		return
	}

	assert.Equal(
		t,
		[]string{
			"{\"t\":0,\"m\":\"{\\\"wut\\\":1,\\\"cmt\\\":800,\\\"buffer\\\":5}\"}",
			"{\"t\":1,\"m\":\"2\"}"},
		msgd,
		"Device messages received")

	assert.Subset(
		t,
		trace.Infos,
		[]string{
			"Hub.The state has been changed (Previous state: Dead, state: " +
				"Inactive, Clients empty: true, Device empty: false, Transmission " +
				"time window: true, )",
		},
		"Info")

	assert.Len(t, trace.Errors, 0, "Errors")
}

func TestHub_IdleBroadcast(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             1,
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Millisecond,
		NotificationTime: 50 * time.Second,
		IniSendTime:      "00:00",
		EndSendTime:      "00:01",
	}

	trace := newTrace()

	wscs, wscc, err := newWS()
	if !assert.NoError(t, err, "New web client socket") {
		return
	}
	defer wscs.Close()
	defer wscc.Close()

	client := iot.NewClient("c1", wscs, 10*time.Minute)

	wsds, wsdc, err := newWS()
	if !assert.NoError(t, err, "New web device socket") {
		return
	}
	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, trace.CHInfo, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterClient(client)
	hub.RegisterDevice(device)

	err = wsdc.WriteMessage(websocket.TextMessage, []byte("message"))
	if !assert.NoError(t, err, "Write device message") {
		return
	}

	time.Sleep(20 * time.Millisecond)

	if assertState(t, hub, iot.Broadcast) {
		return
	}

	time.Sleep(1 * time.Second)

	if assertState(t, hub, iot.Active) {
		return
	}

	assert.Len(t, trace.Errors, 0, "Errors")
}

func TestHub_NotifyState(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             1,
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         10 * time.Millisecond,
		NotificationTime: 20 * time.Millisecond,
		IniSendTime:      "00:00",
		EndSendTime:      "00:01",
	}

	trace := newTrace()

	wscs, wscc, err := newWS()
	if !assert.NoError(t, err, "New web client socket") {
		return
	}
	defer wscs.Close()
	defer wscc.Close()

	client := iot.NewClient("c1", wscs, 10*time.Minute)

	hub := iot.NewHub(cnf, trace.CHInfo, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterClient(client)

	msgd := readMessages(wscc, 1)

	assert.Equal(
		t,
		[]string{
			"0:1"},
		msgd,
		"Notify messages received")

	assert.Len(t, trace.Errors, 0, "Errors")
}

func TestHub_SendConfigToDevice(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             5,
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Second,
		NotificationTime: 50 * time.Second,
		IniSendTime:      "00:00",
		EndSendTime:      "00:01",
	}

	trace := newTrace()

	wsds, wsdc, err := newWS()
	if !assert.NoError(t, err, "New web device socket") {
		return
	}
	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, trace.CHInfo, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterDevice(device)

	hub.Config(iot.DeviceConfig{
		WakeUpTime:         1,
		CollectMetricsTime: 1,
		Buffer:             1,
	})

	msgd := readMessages(wsdc, 3)

	assert.Equal(
		t,
		"{\"t\":0,\"m\":\"{\\\"wut\\\":1,\\\"cmt\\\":1,\\\"buffer\\\":1}\"}",
		msgd[2],
		"Device messages received")

	assert.Len(t, trace.Errors, 0, "Errors")
}

func TestHub_ClientDead(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             1,
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         10 * time.Millisecond,
		NotificationTime: 50 * time.Second,
		IniSendTime:      "00:00",
		EndSendTime:      "00:01",
	}

	trace := newTrace()

	wscs, wscc, err := newWS()
	if !assert.NoError(t, err, "New web client socket") {
		return
	}
	defer wscs.Close()
	defer wscc.Close()

	client := iot.NewClient("c1", wscs, 10*time.Millisecond)

	hub := iot.NewHub(cnf, trace.CHInfo, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterClient(client)

	time.Sleep(20 * time.Millisecond)

	assert.Subset(
		t,
		trace.Infos,
		[]string{
			"Hub.The state has been changed (Previous state: Inactive, " +
				"state: Dead, Clients empty: true, Device empty: true, " +
				"Transmission time window: false, )",
			"The check timer is deactivated because there are no clients",
		},
		"Info")

	assert.Len(t, trace.Errors, 0, "Errors")
}

func Test_StatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    iot.State
		want string
	}{
		{
			name: "Dead Status",
			s:    iot.Dead,
			want: "Dead",
		},
		{
			name: "Inactive Status",
			s:    iot.Inactive,
			want: "Inactive",
		},
		{
			name: "Active Status",
			s:    iot.Active,
			want: "Active",
		},
		{
			name: "Broadcast Status",
			s:    iot.Broadcast,
			want: "Broadcast",
		},
		{
			name: "Asleep Status",
			s:    iot.Asleep,
			want: "Asleep",
		},
		{
			name: "Closed Status",
			s:    iot.Closed,
			want: "Closed",
		},
		{
			name: "None Status",
			s:    10,
			want: "None",
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := iot.StateString(tt.s)

			assert.Equal(t, tt.want, res)
		})
	}
}

func assertState(t *testing.T, hub *iot.Hub, expected iot.State) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	state, err := hub.State(ctx)
	if !assert.NoError(t, err, "Getting state") {
		return true
	}

	assert.Equal(t, expected, state, "Hub state")

	return false
}

func newWS() (*websocket.Conn, *websocket.Conn, error) {
	connsc := make(chan *websocket.Conn)

	handler := func(w http.ResponseWriter, r *http.Request) {
		u := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conns, _ := u.Upgrade(w, r, nil)
		connsc <- conns
	}

	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	connc, r, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error client connection")
	}
	defer r.Body.Close()

	conns := <-connsc

	if conns == nil {
		return nil, nil, errors.New("Error server connection")
	}

	return conns, connc, nil
}

func readMessages(wsc *websocket.Conn, numMsg int) []string {
	var msgs []string

	for i := 0; i < numMsg; i++ {
		_, m, err := wsc.ReadMessage()
		if err != nil {
			return msgs
		}

		msgs = append(msgs, string(m))
	}

	return msgs
}
