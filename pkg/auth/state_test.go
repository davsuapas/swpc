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

package auth_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/pkg/auth"
)

var (
	key = []byte("0123456789QAZWSXEDCRFVTGBYHNUJMI")
)

func TestEncodeState(t *testing.T) {
	t.Parallel()

	type errors struct {
		want bool
		msg  string
	}

	type args struct {
		key   []byte
		state []byte
	}

	tests := []struct {
		name string
		args args
		err  errors
	}{
		{
			name: "Encoding ok",
			args: args{
				key:   key,
				state: []byte("state"),
			},
			err: errors{
				want: false,
				msg:  "",
			},
		},
		{
			name: "Error due to key size",
			args: args{
				key:   []byte("123"),
				state: []byte{},
			},
			err: errors{
				want: true,
				msg:  "Encode state: the key length must be 32 bytes",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state, err := auth.EncodeState(tt.args.key, tt.args.state)

			if err != nil && tt.err.want {
				assert.ErrorContains(t, err, tt.err.msg, "Error")

				return
			}

			assert.NotEmpty(t, state)
		})
	}
}

func TestDecodeState(t *testing.T) {
	t.Parallel()

	statec, err := auth.EncodeState(key, []byte("123"))
	if assert.NoError(t, err) {
		stated, err := auth.DecodeState(key, statec)
		if assert.NoError(t, err) {
			assert.Equal(t, []byte{0x31, 0x32, 0x33}, stated)
		}
	}
}

func TestDecodeState_Error(t *testing.T) {
	t.Parallel()

	type args struct {
		key    []byte
		statec string
	}

	tests := []struct {
		name   string
		args   args
		errMsg string
	}{
		{
			name: "Decode error",
			args: args{
				key:    key,
				statec: "123",
			},
			errMsg: "Decode state: illegal base64 data",
		},
		{
			name: "Unmarshal error",
			args: args{
				key:    key,
				statec: base64.StdEncoding.EncodeToString([]byte("123")),
			},
			errMsg: "Decode state: json: cannot unmarshal",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := auth.DecodeState(tt.args.key, tt.args.statec)

			assert.ErrorContains(t, err, tt.errMsg, "Error")
		})
	}
}
