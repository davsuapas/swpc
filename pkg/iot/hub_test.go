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

package iot_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swpoolcontroller/pkg/iot"
)

type trace struct {
	Traces  []string
	Errors  []string
	CHTrace chan iot.Trace
	CHError chan error
}

func newTrace() *trace {
	t := &trace{
		CHTrace: make(chan iot.Trace),
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
			case i, ok := <-t.CHTrace:
				if ok {
					t.Traces = append(t.Traces, i.Message)
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
			IniSendTime:        "00:01",
			EndSendTime:        "00:02",
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Second,
		NotificationTime: 50 * time.Second,
		HeartbeatConfig: iot.HeartbeatConfig{
			HeartbeatInterval:     10 * time.Second,
			HeartbeatPingTime:     0,
			HeartbeatTimeoutCount: 1,
		},
	}

	trace := newTrace()

	wscs, wscc, err := newWS()
	require.NoError(t, err, "New web client socket")

	defer wscs.Close()
	defer wscc.Close()

	client := iot.NewClient("c1", wscs, 10*time.Minute)

	wsds, wsdc, err := newWS()
	require.NoError(t, err, "New web device socket")

	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, iot.DebugLevel, trace.CHTrace, trace.CHError)
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
	require.NoError(t, err, "New web device socket 1")

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
		trace.Traces,
		[]string{
			"Hub.Client registered (ClientID: c1, Client count: 1, state: Inactive, )",
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
	require.NoError(t, err, "Write device message")

	msgd := readMessages(wsdc, 2)
	msgc := readMessages(wscc, 1)

	if assertState(t, hub, iot.Broadcast) {
		return true
	}

	assert.Equal(
		t,
		iot.DeviceConfigDTO{
			DeviceConfig: iot.DeviceConfig{
				WakeUpTime:         1,
				CollectMetricsTime: 800,
				Buffer:             5,
			},
			HBI:  10,
			HBTC: 1,
		},
		unmarshalHeartbeat(msgd[0]),
		"Broadcast. Device config message")
	assert.Equal(t, "\x011", msgd[1], "Broadcast. Device state message")

	assert.Equal(t, []string{"message"}, msgc, "Client messages received")

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
		iot.DeviceConfigDTO{
			DeviceConfig: iot.DeviceConfig{
				WakeUpTime:         1,
				CollectMetricsTime: 800,
				Buffer:             5,
			},
			HBI:  10,
			HBTC: 1,
		},
		unmarshalHeartbeat(msgd[0]),
		"Sleep. Device sleep config message")
	assert.Equal(t, "\x011", msgd[1], "Sleep. Device state message")
	assert.Equal(t, "\x010", msgd[2], "Sleep. Device state message")

	return false
}

func TestHub_InactiveStateTransmissionTimeWindowsStandbyAction(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             5,
			IniSendTime:        "00:00",
			EndSendTime:        "23:59",
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Second,
		NotificationTime: 50 * time.Second,
		HeartbeatConfig: iot.HeartbeatConfig{
			HeartbeatInterval:     10 * time.Second,
			HeartbeatPingTime:     0,
			HeartbeatTimeoutCount: 1,
		},
	}

	trace := newTrace()

	wsds, wsdc, err := newWS()
	require.NoError(t, err, "New web device socket")

	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, iot.DebugLevel, trace.CHTrace, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterDevice(device)

	msgd := readMessages(wsdc, 2)

	if assertState(t, hub, iot.Inactive) {
		return
	}

	assert.Equal(
		t,
		iot.DeviceConfigDTO{
			DeviceConfig: iot.DeviceConfig{
				WakeUpTime:         1,
				CollectMetricsTime: 800,
				Buffer:             5,
			},
			HBI:  10,
			HBTC: 1,
		},
		unmarshalHeartbeat(msgd[0]),
		"Device config message")
	assert.Equal(t, "\x012", msgd[1], "Device state message")

	assert.Subset(
		t,
		trace.Traces,
		[]string{
			"Hub.The state has been changed (Previous state: Dead, state: " +
				"Inactive, Client count: 0, Device empty: false, Transmission " +
				"time window: true, )",
		},
		"Info")

	assert.Empty(t, trace.Errors, "Errors")
}

func TestHub_IdleBroadcast(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             1,
			IniSendTime:        "00:00",
			EndSendTime:        "00:01",
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Millisecond,
		NotificationTime: 50 * time.Second,
		HeartbeatConfig: iot.HeartbeatConfig{
			HeartbeatInterval:     10 * time.Second,
			HeartbeatPingTime:     0,
			HeartbeatTimeoutCount: 1,
		},
	}

	trace := newTrace()

	wscs, wscc, err := newWS()
	require.NoError(t, err, "New web client socket")

	defer wscs.Close()
	defer wscc.Close()

	client := iot.NewClient("c1", wscs, 10*time.Minute)

	wsds, wsdc, err := newWS()
	require.NoError(t, err, "New web device socket")

	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, iot.DebugLevel, trace.CHTrace, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterClient(client)
	hub.RegisterDevice(device)

	err = wsdc.WriteMessage(websocket.TextMessage, []byte("message"))
	require.NoError(t, err, "Write device message")

	time.Sleep(20 * time.Millisecond)

	if assertState(t, hub, iot.Broadcast) {
		return
	}

	time.Sleep(1400 * time.Millisecond)

	if assertState(t, hub, iot.Active) {
		return
	}

	assert.Empty(t, trace.Errors, "Errors")
}

func TestHub_NotifyState(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             1,
			IniSendTime:        "00:00",
			EndSendTime:        "00:01",
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         10 * time.Millisecond,
		NotificationTime: 20 * time.Millisecond,
		HeartbeatConfig: iot.HeartbeatConfig{
			HeartbeatInterval:     10 * time.Second,
			HeartbeatPingTime:     0,
			HeartbeatTimeoutCount: 1,
		},
	}

	trace := newTrace()

	wscs, wscc, err := newWS()
	require.NoError(t, err, "New web client socket")

	defer wscs.Close()
	defer wscc.Close()

	client := iot.NewClient("c1", wscs, 10*time.Minute)

	hub := iot.NewHub(cnf, iot.DebugLevel, trace.CHTrace, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterClient(client)

	msgd := readMessages(wscc, 1)

	assert.Equal(
		t,
		[]string{
			"01"},
		msgd,
		"Notify messages received")

	assert.Empty(t, trace.Errors, "Errors")
}

func TestHub_SendConfigToDevice(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             5,
			IniSendTime:        "00:00",
			EndSendTime:        "00:01",
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Second,
		NotificationTime: 50 * time.Second,
		HeartbeatConfig: iot.HeartbeatConfig{
			HeartbeatInterval:     10 * time.Second,
			HeartbeatPingTime:     1 * time.Second,
			HeartbeatTimeoutCount: 1,
		},
	}

	trace := newTrace()

	wsds, wsdc, err := newWS()
	require.NoError(t, err, "New web device socket")

	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, iot.DebugLevel, trace.CHTrace, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterDevice(device)

	hub.Config(iot.DeviceConfig{
		WakeUpTime:         1,
		CollectMetricsTime: 1,
		Buffer:             1,
		IniSendTime:        "11:00",
		EndSendTime:        "12:00",
	})

	msgd := readMessages(wsdc, 3)

	assert.Equal(
		t,
		iot.DeviceConfigDTO{
			DeviceConfig: iot.DeviceConfig{
				WakeUpTime:         1,
				CollectMetricsTime: 1,
				Buffer:             1,
			},
			HBI:  10,
			HBTC: 1,
		},
		unmarshalHeartbeat(msgd[2]),
		"Device config message")

	assert.Empty(t, trace.Errors, "Errors")
}

func TestHub_ClientDead(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             1,
			IniSendTime:        "00:00",
			EndSendTime:        "00:01",
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         10 * time.Millisecond,
		NotificationTime: 50 * time.Second,
		HeartbeatConfig: iot.HeartbeatConfig{
			HeartbeatInterval:     10 * time.Second,
			HeartbeatPingTime:     0,
			HeartbeatTimeoutCount: 1,
		},
	}

	trace := newTrace()

	wscs, wscc, err := newWS()
	require.NoError(t, err, "New web client socket")

	defer wscs.Close()
	defer wscc.Close()

	client := iot.NewClient("c1", wscs, 10*time.Millisecond)

	hub := iot.NewHub(cnf, iot.DebugLevel, trace.CHTrace, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterClient(client)

	time.Sleep(20 * time.Millisecond)

	assert.Subset(
		t,
		trace.Traces,
		[]string{
			"Hub.The state has been changed (Previous state: Inactive, " +
				"state: Dead, Client count: 0, Device empty: true, " +
				"Transmission time window: false, )",
			"The check timer is deactivated because there are no activity",
		},
		"Info")

	assert.Empty(t, trace.Errors, "Errors")
}

func TestHub_Multi_Link_Device(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             1,
			IniSendTime:        "00:00",
			EndSendTime:        "00:01",
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Second,
		NotificationTime: 50 * time.Second,
		HeartbeatConfig: iot.HeartbeatConfig{
			HeartbeatInterval:     10 * time.Second,
			HeartbeatPingTime:     0,
			HeartbeatTimeoutCount: 1,
		},
	}

	trace := newTrace()

	wsds, wsdc, err := newWS()
	require.NoError(t, err, "New web device socket")

	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(cnf, iot.DebugLevel, trace.CHTrace, trace.CHError)
	defer hub.Stop()

	hub.Run()

	hub.RegisterDevice(device)
	time.Sleep(10 * time.Millisecond)

	wsds1, wsdc1, err1 := newWS()
	require.NoError(t, err1, "New web device1 socket")

	defer wsds1.Close()
	defer wsdc1.Close()

	device1 := iot.Device{
		ID:         "d2",
		Connection: wsds1,
	}

	hub.RegisterDevice(device1)
	time.Sleep(10 * time.Millisecond)

	msgd := readMessages(wsdc1, 1)

	assert.Equal(
		t,
		iot.DeviceConfigDTO{
			DeviceConfig: iot.DeviceConfig{
				WakeUpTime:         1,
				CollectMetricsTime: 800,
				Buffer:             1,
			},
			HBI:  10,
			HBTC: 1,
		},
		unmarshalHeartbeat(msgd[0]),
		"Device config message")

	assert.Empty(t, trace.Errors, "Errors")
}

func TestHub_HeartbeatTimeout(t *testing.T) {
	t.Parallel()

	cnf := iot.Config{
		DeviceConfig: iot.DeviceConfig{
			WakeUpTime:         1,
			CollectMetricsTime: 800,
			Buffer:             1,
			IniSendTime:        "00:00",
			EndSendTime:        "00:01",
		},
		CommLatency:      1 * time.Millisecond,
		TaskTime:         50 * time.Second,
		NotificationTime: 50 * time.Second,
		HeartbeatConfig: iot.HeartbeatConfig{
			HeartbeatInterval:     1 * time.Millisecond,
			HeartbeatPingTime:     1 * time.Millisecond,
			HeartbeatTimeoutCount: 2,
		},
	}

	trace := newTrace()

	wsds, wsdc, err := newWS()
	require.NoError(t, err, "New web device socket")

	defer wsds.Close()
	defer wsdc.Close()

	device := iot.Device{
		ID:         "d1",
		Connection: wsds,
	}

	hub := iot.NewHub(
		cnf,
		iot.DebugLevel,
		trace.CHTrace,
		trace.CHError)

	defer hub.Stop()

	hub.Run()

	hub.RegisterDevice(device)
	time.Sleep(10 * time.Millisecond)

	assert.Len(t, trace.Errors, 1, "Errors")
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
	require.NoError(t, err, "Getting state")

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

func unmarshalHeartbeat(data string) iot.DeviceConfigDTO {
	var res iot.DeviceConfigDTO
	_ = json.Unmarshal([]byte(data[1:]), &res)

	return res
}
