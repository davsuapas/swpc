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

package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/strings"
)

var (
	errKeyNotFound = errors.New("Key Not found into JWK")
)

const (
	errConvertJWT    = "Converting JWT token"
	errParseJWT      = "Parsing JWT token"
	errDecodeE       = "Decoding E"
	errDecodeN       = "Decoding N"
	errGetJWK        = "Fetching JWK"
	errReadJWK       = "Reading JWK body"
	errUnmarshallJWK = "Unmarshalling JWK body"
	errGetKeyJWK     = "Getting JWK Key"
)

// JWKKey key structure
type JWKKey struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
}

type JWK struct {
	Keys []JWKKey `json:"keys"`
}

// JWKFetch allows to obtain a JWK given a url from a provider.
// This structure is intended to be stored throughout
// the life cycle of the application as a cache.
// Operations are protected against concurrency
type JWKFetch struct {
	url string
	jwk JWK

	lock sync.RWMutex
}

// NewJWKFetch creates JWKFetch
func NewJWKFetch(url string) *JWKFetch {
	return &JWKFetch{
		url:  url,
		jwk:  JWK{},
		lock: sync.RWMutex{},
	}
}

// JWK gets JWK. it is protected against concurrency
func (k *JWKFetch) JWK() JWK {
	k.lock.RLock()
	jwk := k.jwk
	k.lock.RUnlock()

	return jwk
}

// JWKKey searches the key by kid and if it is not found,
// get the JWK from the provider again (it may have been rotated),
// if it is not found again, it will give an error.
// it is protected against concurrency
func (k *JWKFetch) JWKKey(kid string) (JWKKey, error) {
	findKey := func(kid string) (bool, JWKKey) {
		jwk := k.JWK()

		for _, v := range jwk.Keys {
			if v.Kid == kid {
				return true, v
			}
		}

		return false, JWKKey{}
	}

	if found, key := findKey(kid); found {
		return key, nil
	}

	if err := k.Fetch(); err != nil {
		return JWKKey{}, errors.Wrap(err, errGetKeyJWK)
	}

	if found, key := findKey(kid); found {
		return key, nil
	}

	return JWKKey{}, errKeyNotFound
}

// Fetch gets the JWK from the provider and stores it
func (k *JWKFetch) Fetch() error {
	ctx := context.TODO()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.url, nil)
	if err != nil {
		return errors.Wrap(err, errGetJWK)
	}

	req.Header.Add("Accept", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, errGetJWK)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(strings.Concat(errGetJWK, ", StatusCode: ", resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, errReadJWK)
	}

	jwk := JWK{}

	err = json.Unmarshal(body, &jwk)
	if err != nil {
		return errors.Wrap(err, errUnmarshallJWK)
	}

	k.lock.Lock()
	k.jwk = jwk
	k.lock.Unlock()

	return nil
}

// JWT manages JWT operations
type JWT struct {
	JWKFetch *JWKFetch
}

// GetKey gets the public key
func (k *JWT) GetKey(token *jwt.Token) (interface{}, error) {
	jwkKey, err := k.JWKFetch.JWKKey(token.Header["kid"].(string))
	if err != nil {
		return nil, err
	}

	key, err := convertKey(jwkKey)
	if err != nil {
		return nil, errors.Wrap(err, errConvertJWT)
	}

	return &key, nil
}

// ParseJWT parse the token given a JWK service
func (k *JWT) ParseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, k.GetKey)

	if err != nil {
		return token, errors.Wrap(err, errParseJWT)
	}

	return token, nil
}

func convertKey(jwkKey JWKKey) (rsa.PublicKey, error) {
	rawE := jwkKey.E
	rawN := jwkKey.N

	decodedE, err := base64.RawURLEncoding.DecodeString(rawE)
	if err != nil {
		return rsa.PublicKey{}, errors.Wrap(err, errDecodeE)
	}

	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}

	pubKey := rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE)),
	}

	decodedN, err := base64.RawURLEncoding.DecodeString(rawN)
	if err != nil {
		return rsa.PublicKey{}, errors.Wrap(err, errDecodeN)
	}

	pubKey.N.SetBytes(decodedN)

	return pubKey, nil
}
