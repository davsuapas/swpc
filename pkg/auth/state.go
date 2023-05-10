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

package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/crypto"
)

var errIncosState = errors.New("Invalid state")

const (
	errEncodeState = "Encode state"
	errDecodeState = "Decode state"
)

type stateCodec struct {
	StateEcrypt []byte `json:"se,omitempty"`
	StateOrigin []byte `json:"so,omitempty"`
}

// GenerateState generates encrypt state.
// The state is composed of an encrypted secret
// and the secret encoded in base64
func EncodeState(key []byte, state []byte) (string, error) {
	statee, err := crypto.Encrypt(key, state)
	if err != nil {
		return "", errors.Wrap(err, errEncodeState)
	}

	sc := stateCodec{
		StateEcrypt: statee,
		StateOrigin: state,
	}

	scj, err := json.Marshal(sc)
	if err != nil {
		return "", errors.Wrap(err, errEncodeState)
	}

	return base64.StdEncoding.EncodeToString(scj), nil
}

// DecodeState decodes the state and checks that the encrypted state
// and the plain state match
func DecodeState(key []byte, statec string) ([]byte, error) {
	sd, err := base64.StdEncoding.DecodeString(statec)
	if err != nil {
		return nil, errors.Wrap(err, errDecodeState)
	}

	var sc stateCodec

	if err := json.Unmarshal(sd, &sc); err != nil {
		return nil, errors.Wrap(err, errDecodeState)
	}

	stateDecrypt, err := crypto.Decrypt(key, sc.StateEcrypt)
	if err != nil {
		return nil, errors.Wrap(err, errDecodeState)
	}

	if !bytes.Equal(stateDecrypt, sc.StateOrigin) {
		return nil, errIncosState
	}

	return stateDecrypt, nil
}
