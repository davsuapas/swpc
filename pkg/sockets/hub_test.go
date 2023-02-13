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

package sockets_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/pkg/sockets"
)

func TestHub_Register(t *testing.T) {
	t.Parallel()

	type cases struct {
		clientsAlreadyRegistered bool
	}

	type args struct {
		clientID string
	}

	tests := []struct {
		name     string
		cases    cases
		args     args
		expected string
	}{
		{
			name: "Register a client",
			cases: cases{
				clientsAlreadyRegistered: false,
			},
			args: args{
				clientID: "1",
			},
			expected: "Hub-> Client registered (ClientID: 1, Length: 1, Length: Active, )",
		},
		{
			name: "Register a client with other client already registered",
			cases: cases{
				clientsAlreadyRegistered: true,
			},
			args: args{
				clientID: "2",
			},
			expected: "Hub-> Client registered (ClientID: 2, Length: 2, Length: Active, )",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ws, wc, errs := newWS()
			if errs != nil {
				t.Error(errs)

				return
			}
			defer wc.Close()

			info := make(chan string)
			err := make(chan error)

			h := sockets.NewHub(
				sockets.Config{
					TaskTime: 1 * time.Hour,
				},
				info,
				err)
			defer h.Stop(true)

			h.Run()

			if tt.cases.clientsAlreadyRegistered {
				h.Register(sockets.NewClient("0", ws, 10*time.Second))
				<-info
			}

			h.Register(sockets.NewClient(tt.args.clientID, ws, 10*time.Second))

			res := <-info

			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestHub_Unregister(t *testing.T) {
	t.Parallel()

	type cases struct {
		clientsAlreadyRegistered bool
	}

	type args struct {
		clientID string
	}

	tests := []struct {
		name     string
		cases    cases
		args     args
		expected []string
	}{
		{
			name: "Unregister a client",
			cases: cases{
				clientsAlreadyRegistered: false,
			},
			args: args{
				clientID: "1",
			},
			expected: []string{
				"Hub-> The hub is set to deactivated",
				"Hub-> Client unregisted (ClientID: 1, Length: 0, Status: Deactivated, )",
			},
		},
		{
			name: "Unregister a client with other client already registered",
			cases: cases{
				clientsAlreadyRegistered: true,
			},
			args: args{
				clientID: "1",
			},
			expected: []string{
				"Hub-> Client unregisted (ClientID: 1, Length: 1, Status: Active, )",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ws, wc, errs := newWS()
			if errs != nil {
				t.Error(errs)

				return
			}
			defer wc.Close()

			info := make(chan string)
			err := make(chan error)

			h := sockets.NewHub(
				sockets.Config{
					TaskTime: 1 * time.Hour,
				},
				info,
				err)
			defer h.Stop(true)

			h.Run()

			if tt.cases.clientsAlreadyRegistered {
				h.Register(sockets.NewClient("0", ws, 10*time.Second))
				<-info
			}

			h.Register(sockets.NewClient(tt.args.clientID, ws, 10*time.Second))
			<-info

			h.Unregister(tt.args.clientID)

			assertArrayAllEqual(t, tt.expected, info, "Info")
		})
	}
}

func TestHub_Config(t *testing.T) {
	t.Parallel()

	info := make(chan string)
	err := make(chan error)

	h := sockets.NewHub(
		sockets.Config{
			TaskTime: 1 * time.Hour,
		},
		info,
		err)
	defer h.Stop(true)

	h.Run()

	h.Config(sockets.Config{
		CommLatency:      1 * time.Second,
		Buffer:           2 * time.Second,
		TaskTime:         1 * time.Hour,
		NotificationTime: 3 * time.Second,
	})

	assert.Equal(
		t,
		"Hub-> The configuration has been changed (Config: {\"CommLatency\":1000000000,\"Buffer\":2000000000,"+
			"\"TaskTime\":3600000000000,\"NotificationTime\":3000000000}, )",
		<-info)
}

func TestHub_Status(t *testing.T) {
	t.Parallel()

	info := make(chan string)
	err := make(chan error)

	h := sockets.NewHub(
		sockets.Config{
			TaskTime: 1 * time.Hour,
		},
		info,
		err)
	defer h.Stop(true)

	h.Run()

	resp := make(chan sockets.Status)
	h.Status(resp)

	assert.Equal(t, sockets.Deactivated, <-resp)
}

func TestHub_Send(t *testing.T) {
	t.Parallel()

	type cases struct {
		statusPrevious sockets.Status
	}

	type expected struct {
		status  sockets.Status
		message string
	}

	tests := []struct {
		name     string
		cases    cases
		mSend    string
		expected expected
	}{
		{
			name: "Unregister a client with hub deactivated",
			cases: cases{
				statusPrevious: sockets.Deactivated,
			},
			mSend: "message1",
			expected: expected{
				status:  sockets.Deactivated,
				message: "",
			},
		},
		{
			name: "Unregister a client with hub activated",
			cases: cases{
				statusPrevious: sockets.Active,
			},
			mSend: "message2",
			expected: expected{
				status:  sockets.Streaming,
				message: "1:message2",
			},
		},
		{
			name: "Unregister a client with hub in streaming mode",
			cases: cases{
				statusPrevious: sockets.Streaming,
			},
			mSend: "message3",
			expected: expected{
				status:  sockets.Streaming,
				message: "1:message3",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ws, wc, errs := newWS()
			if errs != nil {
				t.Error(errs)

				return
			}
			defer wc.Close()

			info := make(chan string)
			err := make(chan error)

			h := sockets.NewHub(
				sockets.Config{
					TaskTime: 1 * time.Hour,
				},
				info,
				err)
			defer h.Stop(true)

			h.Run()

			if tt.cases.statusPrevious == sockets.Active || tt.cases.statusPrevious == sockets.Streaming {
				h.Register(sockets.NewClient("0", ws, 10*time.Second))
				<-info
			}
			if tt.cases.statusPrevious == sockets.Streaming {
				h.Send("message previous")
				_, _, _ = wc.ReadMessage()
				<-info
			}

			h.Send(tt.mSend)

			if tt.cases.statusPrevious == sockets.Active || tt.cases.statusPrevious == sockets.Streaming {
				_, mactual, errm := wc.ReadMessage()
				if errm != nil {
					assert.Error(t, errm, "Error reading websocket message")

					return
				}

				assert.Equal(t, tt.expected.message, string(mactual))
			}

			if tt.cases.statusPrevious != sockets.Streaming {
				<-info
			}

			resp := make(chan sockets.Status)
			h.Status(resp)

			assert.Equal(t, tt.expected.status, <-resp)
		})
	}
}

func TestHub_IdleController(t *testing.T) {
	t.Parallel()

	type cases struct {
		statusPrevious sockets.Status
		idle           bool
	}

	tests := []struct {
		name     string
		cases    cases
		config   sockets.Config
		expected sockets.Status
	}{
		{
			name: "Idle with hub deactivated",
			cases: cases{
				statusPrevious: sockets.Deactivated,
				idle:           false,
			},
			config: sockets.Config{
				Buffer:      1 * time.Millisecond,
				CommLatency: 1 * time.Millisecond,
				TaskTime:    1 * time.Millisecond,
			},
			expected: sockets.Deactivated,
		},
		{
			name: "Idle with hub in streaming mode",
			cases: cases{
				statusPrevious: sockets.Streaming,
				idle:           true,
			},
			config: sockets.Config{
				Buffer:      1 * time.Millisecond,
				CommLatency: 1 * time.Millisecond,
				TaskTime:    10 * time.Millisecond,
			},
			expected: sockets.Inactive,
		},
		{
			name: "No Idle with hub in streaming mode",
			cases: cases{
				statusPrevious: sockets.Streaming,
			},
			config: sockets.Config{
				Buffer:      1 * time.Minute,
				CommLatency: 1 * time.Minute,
				TaskTime:    10 * time.Millisecond,
			},
			expected: sockets.Streaming,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ws, wc, errs := newWS()
			if errs != nil {
				t.Error(errs)

				return
			}
			defer wc.Close()

			info := make(chan string)
			err := make(chan error)

			h := sockets.NewHub(tt.config, info, err)
			defer h.Stop(true)

			h.Run()

			if tt.cases.statusPrevious == sockets.Streaming {
				h.Register(sockets.NewClient("0", ws, 10*time.Second))
				<-info
				h.Send("message previous")
				<-info
			}

			if tt.cases.statusPrevious == sockets.Streaming {
				if tt.cases.idle {
					<-info
					<-info
				}
			} else {
				<-info
			}

			resp := make(chan sockets.Status)
			h.Status(resp)

			assert.Equal(t, tt.expected, <-resp)
		})
	}
}

func TestHub_NotifyStatus(t *testing.T) {
	t.Parallel()

	type expected struct {
		message string
		status  sockets.Status
	}

	tests := []struct {
		name     string
		notify   bool
		config   sockets.Config
		expected expected
	}{
		{
			name:   "No status reported",
			notify: false,
			config: sockets.Config{
				NotificationTime: 1 * time.Minute,
				TaskTime:         1 * time.Millisecond,
			},
			expected: expected{
				message: "",
				status:  sockets.Active,
			},
		},
		{
			name:   "Status reported",
			notify: true,
			config: sockets.Config{
				NotificationTime: 1 * time.Millisecond,
				TaskTime:         1 * time.Millisecond,
			},
			expected: expected{
				message: "0:1",
				status:  sockets.Active,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ws, wc, errs := newWS()
			if errs != nil {
				t.Error(errs)

				return
			}
			defer wc.Close()

			info := make(chan string)
			err := make(chan error)

			h := sockets.NewHub(tt.config, info, err)
			defer h.Stop(true)

			h.Run()

			h.Register(sockets.NewClient("0", ws, 10*time.Second))
			<-info

			time.Sleep(tt.config.TaskTime)
			if tt.notify {
				_, mactual, errm := wc.ReadMessage()
				if errm != nil {
					assert.Error(t, errm, "Error reading websocket message")

					return
				}

				assert.Equal(t, tt.expected.message, string(mactual))
				<-info
			}

			resp := make(chan sockets.Status)
			h.Status(resp)

			assert.Equal(t, tt.expected.status, <-resp)
		})
	}
}

func TestHub_RemoveDeadClient(t *testing.T) {
	t.Parallel()

	ws, wc, errs := newWS()
	if errs != nil {
		t.Error(errs)

		return
	}
	defer wc.Close()

	info := make(chan string)
	err := make(chan error)

	h := sockets.NewHub(
		sockets.Config{
			TaskTime:         1 * time.Millisecond,
			NotificationTime: 1 * time.Hour,
		},
		info,
		err)
	defer h.Stop(true)

	h.Run()

	h.Register(sockets.NewClient("0", ws, 1*time.Millisecond))
	<-info

	<-info
	<-info
	res := <-info

	<-info

	assert.Equal(t, "Hub-> Array size after removing expired clients (Length: 0, )", res)
}

func assertArrayAllEqual(t assert.TestingT, exptected []string, actual chan string, msg string) {
	for _, e := range exptected {
		assert.Equal(t, e, <-actual, msg)
	}
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
