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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	stringss "github.com/swpoolcontroller/pkg/strings"
)

const (
	errTk           = "OAuth2-> Get token"
	errRevokeTk     = "OAuth2-> Revoke token"
	errJWKCode      = "OAuth2-> Get JWK"
	errRequest      = "OAuth2-> New http request"
	errHTTPPost     = "OAuth2-> Request using post method"
	errHTTPReadBody = "OAuth2-> Read body from response"
	errUnmarshallTk = "OAuth2-> Unmarshall token from response"
	errJWTTk        = "OAuth2-> Get JWT token"
)

// OA2TokenInput defines the parameters of the oauth2 token input
type OA2TokenInput struct {
	URL         string
	Code        string
	RedirectURI string
}

type respToken struct {
	AccessToken string `json:"access_token"` //nolint:golint,tagliatelle
	ExpiresIn   int    `json:"expires_in"`   //nolint:golint,tagliatelle
}

// OA2RevokeTokenInput defines the parameters of the oauth2 input
// to revoke token
type OA2RevokeTokenInput struct {
	URL   string
	Token string
}

// OAuth2 manages the oauth2 infrastructure
type OAuth2 struct {
	ClientID string

	JWT *JWT
}

// Token gets a token given a JWK
// The JWK service must be persisted throughout
// the entire application cycle as a cache.
func (o *OAuth2) Token(params OA2TokenInput) (*jwt.Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", o.ClientID)
	data.Set("code", params.Code)
	data.Set("redirect_uri", params.RedirectURI)

	body, err := post(params.URL, data)
	if err != nil {
		return nil, errors.Wrap(err, errTk)
	}

	var respToken respToken

	if err := json.Unmarshal(body, &respToken); err != nil {
		return nil, errors.Wrap(err, errUnmarshallTk)
	}

	jwtToken, err := o.JWT.ParseJWT(respToken.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, errJWTTk)
	}

	return jwtToken, nil
}

// RevokeToken revokes a token
func (o *OAuth2) RevokeToken(params OA2RevokeTokenInput) error {
	data := url.Values{}
	data.Set("client_id", o.ClientID)
	data.Set("token", params.Token)

	if _, err := post(params.URL, data); err != nil {
		return errors.Wrap(err, errRevokeTk)
	}

	return nil
}

func post(url string, data url.Values) ([]byte, error) {
	ctx := context.TODO()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, errRequest)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errHTTPPost)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errHTTPReadBody)
	}

	if resp.StatusCode != http.StatusOK {
		return body, errors.New(
			stringss.Concat(
				errHTTPPost,
				", StatusCode: ", resp.Status,
				", Response: ", string(body)))
	}

	return body, nil
}

// Oauth2URL returns a valid URL compatible with oauth2
func Oauth2URL(
	url string,
	clientID string,
	redirectURL string,
	state string) string {
	signURL := strings.ReplaceAll(url, "%client_id", clientID)
	signURL = strings.ReplaceAll(signURL, "%redirect_uri", redirectURL)
	signURL = strings.ReplaceAll(signURL, "%state", state)

	return signURL
}
