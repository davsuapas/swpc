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

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/pkg/auth"
)

func TestOAuth2_Token(t *testing.T) {
	t.Parallel()

	type responsejwk struct {
		body       string
		statusCode int
	}

	type responseoauth struct {
		body       string
		statusCode int
	}

	tests := []struct {
		name      string
		respJWK   responsejwk
		respOAuth responseoauth
		wantErr   string
	}{
		{
			name: "Token Ok",
			respJWK: responsejwk{
				body:       "{\"keys\":[{\"alg\":\"RS256\",\"e\":\"AQAB\",\"kid\":\"Jf50SRGXc8qe5vX1zvK2Li3qJU4j8tgTIcKwpa6yTk4=\",\"kty\":\"RSA\",\"n\":\"1et6-M1xXbFvd9dgCZ5zIOCG57stSbrGIEtFeTG9ULkHts3nlIfH6CSYunuHgGhGKUEH4UdcP75QGrlgJN63nTgJkhL5c4j-eBuWbe6JJNXtMqxGXoRLl1TJYJaSP3KotRMhsi3Df-zDG6tbZ_yH7ulbMzu_U5HZzmaY3cwJ5qB31-uOosnPCDh3Be1ay5YdYzdY8cz_OmHXfonIHSGZMtY6Xes5pRpHlz_ZApW9t1T2frhANXv27VAGiD0xpBFuXEYCDZYe8V68BAM_2NwAIYyDbctk3iu_nA1jw-xFRKVyvbrSh1GPQc0JnW0TwE8_43vWybeE3iGg6TQ67d8bDw\",\"use\":\"sig\"},{\"alg\":\"RS256\",\"e\":\"AQAB\",\"kid\":\"4DAErExwVd4GwK0iPC+R5jSR6OWCA6omcNBfth3w45w=\",\"kty\":\"RSA\",\"n\":\"nK12U2xLIGAzQFyuxebztzrUFE9T3vQUCji8PlY-eRt8EOySiKqxUpVjpkwGxY2jFKZvJS9A1MJf9qC3If_zI6d5i4gNwmvl9hIboWIhT2SUCm5S-BzgJl8Zt86ZLYxBoC993vRQT0pYdXu6AB21UxWn6v4Q56_EA9n5asdwI-9hC22dCtGOx-qThxuOdsdS6mrWUsOgHwBx7uLQJEk2ZoeI_h4vupATGQQ4ZWXnYZrkm3MhMwG68J6qTvTvj5d5rdok4qIIg_GQGht_OZorZiJCNQz0c6Zea95JHhGBLfkbuwem5znvpVZUTa2HxN_c1HlY3_kRd0mUomghXdT4Gw\",\"use\":\"sig\"}]}",
				statusCode: http.StatusOK,
			},
			respOAuth: responseoauth{
				body:       "{\"access_token\":\"eyJraWQiOiI0REFFckV4d1ZkNEd3SzBpUEMrUjVqU1I2T1dDQTZvbWNOQmZ0aDN3NDV3PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiJhN2FlNWNlZi0wZGQyLTQyYWItYjRkYy0zMWI5Yzg1N2UzMzEiLCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAuZXUtd2VzdC0xLmFtYXpvbmF3cy5jb21cL2V1LXdlc3QtMV9VWWgzM1dlbGoiLCJ2ZXJzaW9uIjoyLCJjbGllbnRfaWQiOiI2c3RkZ3Y0ZDYxcmJuNmZka3FqYmEwN3ZoNCIsImV2ZW50X2lkIjoiYjJkNjgzNWUtZTMwNS00NmI0LTliNDgtNjU1OTNhNjlhOWQ2IiwidG9rZW5fdXNlIjoiYWNjZXNzIiwic2NvcGUiOiJvcGVuaWQgZW1haWwiLCJhdXRoX3RpbWUiOjE2ODI5MjI1NDksImV4cCI6MTY4MjkyNjE0OSwiaWF0IjoxNjgyOTIyNTQ5LCJqdGkiOiI0N2Y0YmMzMC01ZWU1LTRkMmUtOTZlYS0yNGI0OTdlMDA2ODciLCJ1c2VybmFtZSI6ImRhdnN1YXBhcyJ9.GRvxrannCQpMEMnE1WYXa39BxP62hF5BfnQcvjsBB3odnGCLBIghwE2wKAG0ufHkvM6Z3uoHPDJesleMKkcGny6RPaQNSRfvvgklQpjTXr5WL-lXEOMjTKzNIZdamiwbDnmQ-qF_zQQpVlDvdPp-zpv-ZwwakETFcxSmGLVfux9PJcPZLs-UmB5wn2dPGUDnxY19_uVY8nPoTB9-ZgbCb-fOoMmbmvGKMWHRZpzQS4O0CUsC5anhIAvBAcTlZa85N4v9U_iPfQO_Uvgy2qqSRsx2e3eBqefrFwVXyaB75oe_SiUXmKYfzXywSN3RCTlu9K-tGCRi7dCC4jRQxc_wIQ\",\"expires_in\":123}",
				statusCode: http.StatusOK,
			},
			wantErr: "Token is expired",
		},
		{
			name: "Token unmarshal error",
			respOAuth: responseoauth{
				body:       "{",
				statusCode: http.StatusOK,
			},
			wantErr: "unexpected end of JSON input",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				sc := tt.respJWK.statusCode
				body := tt.respJWK.body

				if r.Method == http.MethodPost {
					sc = tt.respOAuth.statusCode
					body = tt.respOAuth.body
				}

				w.WriteHeader(sc)
				_, _ = w.Write([]byte(body))
			}

			s := httptest.NewServer(http.HandlerFunc(reg))
			defer s.Close()

			o := auth.OAuth2{
				ClientID: "client_id",
				JWT: &auth.JWT{
					JWKFetch: auth.NewJWKFetch(s.URL),
				},
			}

			if _, err := o.Token(auth.OA2TokenInput{URL: s.URL}); err != nil {
				if len(tt.wantErr) == 0 {
					assert.NoError(t, err, "Error")
				}
				assert.ErrorContains(t, err, tt.wantErr, "Error")
			}
		})
	}
}

func TestOAuth2_RevokeToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		respstatus int
		wantErr    string
	}{
		{
			name:       "Revoke Ok",
			respstatus: http.StatusOK,
			wantErr:    "",
		},
		{
			name:       "Revoke with not found status code response",
			respstatus: http.StatusNotFound,
			wantErr:    "StatusCode: 404 Not Found",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.respstatus)
			}

			s := httptest.NewServer(http.HandlerFunc(reg))
			defer s.Close()

			o := auth.OAuth2{
				ClientID: "client_id",
			}

			if err := o.RevokeToken(auth.OA2RevokeTokenInput{URL: s.URL}); err != nil {
				if len(tt.wantErr) == 0 {
					assert.NoError(t, err, "Error")
				}
				assert.ErrorContains(t, err, tt.wantErr, "Error")
			}
		})
	}
}

func TestOauth2URL(t *testing.T) {
	t.Parallel()

	url := auth.Oauth2URL(
		"http://server?client_id=%client_id;redirect_uri=%redirectURL;state=%state",
		"clientID",
		"redirectURL",
		"state")
	assert.Equal(t, "http://server?client_id=clientID;redirect_uri=%redirectURL;state=state", url)
}
