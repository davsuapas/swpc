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

func TestJWKFetch_Fetch(t *testing.T) {
	t.Parallel()

	type response struct {
		body       string
		statusCode int
	}

	tests := []struct {
		name    string
		resp    response
		want    auth.JWK
		wantErr string
	}{
		{
			name: "Receive JWK",
			resp: response{
				body:       `{"keys":[{"alg":"alg","e":"e","kid":"kid","kty":"kty","n":"n"}]}`,
				statusCode: http.StatusOK,
			},
			want: auth.JWK{
				Keys: []auth.JWKKey{{Alg: "alg", E: "e", Kid: "kid", Kty: "kty", N: "n"}},
			},
			wantErr: "",
		},
		{
			name: "Error 404 Status Code",
			resp: response{
				body:       "",
				statusCode: http.StatusNotFound,
			},
			want:    auth.JWK{},
			wantErr: "Fetching JWK, StatusCode: 404 Not Found",
		},
		{
			name: "Error Unmarshal",
			resp: response{
				body:       "error",
				statusCode: http.StatusOK,
			},
			want:    auth.JWK{},
			wantErr: "Unmarshalling JWK body",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.resp.statusCode)
				_, _ = w.Write([]byte(tt.resp.body))
			}

			s := httptest.NewServer(http.HandlerFunc(reg))
			defer s.Close()

			k := auth.NewJWKFetch(s.URL)

			if err := k.Fetch(); err != nil {
				if len(tt.wantErr) == 0 {
					assert.NoError(t, err, "Error")
				}
				assert.ErrorContains(t, err, tt.wantErr, "Error")
			}

			assert.Equal(t, tt.want, k.JWK())
		})
	}
}

func TestJWKFetch_JWKKey(t *testing.T) {
	t.Parallel()

	type response struct {
		body       string
		statusCode int
	}

	type field struct {
		prefecth bool
		kid      string
	}

	tests := []struct {
		name    string
		field   field
		resp    response
		want    auth.JWKKey
		wantErr string
	}{
		{
			name: "kid found with prefetch",
			field: field{
				prefecth: true,
				kid:      "kid",
			},
			resp: response{
				body:       `{"keys":[{"alg":"alg","e":"e","kid":"kid","kty":"kty","n":"n"}]}`,
				statusCode: http.StatusOK,
			},
			want:    auth.JWKKey{Alg: "alg", E: "e", Kid: "kid", Kty: "kty", N: "n"},
			wantErr: "",
		},
		{
			name: "kid not found with prefetch",
			field: field{
				prefecth: true,
				kid:      "kid1",
			},
			resp: response{
				body:       `{"keys":[{"alg":"alg","e":"e","kid":"kid","kty":"kty","n":"n"}]}`,
				statusCode: http.StatusOK,
			},
			want:    auth.JWKKey{},
			wantErr: "Key Not found into JWK",
		},
		{
			name: "kid found without prefetch",
			field: field{
				prefecth: false,
				kid:      "kid1",
			},
			resp: response{
				body: `{"keys":[
						{"alg":"alg","e":"e","kid":"kid","kty":"kty","n":"n"},
					 	{"alg":"alg1","e":"e1","kid":"kid1","kty":"kty1","n":"n1"}
					]}`,
				statusCode: http.StatusOK,
			},
			want:    auth.JWKKey{Alg: "alg1", E: "e1", Kid: "kid1", Kty: "kty1", N: "n1"},
			wantErr: "",
		},
		{
			name: "kid not found without prefetch",
			field: field{
				prefecth: false,
				kid:      "kid1",
			},
			resp: response{
				body:       `{"keys":[{"alg":"alg","e":"e","kid":"kid","kty":"kty","n":"n"}]}`,
				statusCode: http.StatusOK,
			},
			want:    auth.JWKKey{},
			wantErr: "Key Not found into JWK",
		},
		{
			name: "Error 404 not found",
			field: field{
				prefecth: false,
				kid:      "kid",
			},
			resp: response{
				body:       "",
				statusCode: http.StatusNotFound,
			},
			want:    auth.JWKKey{},
			wantErr: "Fetching JWK, StatusCode: 404 Not Found",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.resp.statusCode)
				_, _ = w.Write([]byte(tt.resp.body))
			}

			s := httptest.NewServer(http.HandlerFunc(reg))
			defer s.Close()

			k := auth.NewJWKFetch(s.URL)

			if tt.field.prefecth {
				_ = k.Fetch()
			}

			jwk, err := k.JWKKey(tt.field.kid)
			if err != nil {
				if len(tt.wantErr) == 0 {
					assert.NoError(t, err, "Error")
				}
				assert.ErrorContains(t, err, tt.wantErr, "Error")
			}

			assert.Equal(t, tt.want, jwk)
		})
	}
}

func TestJWT_ParseJWT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		token    string
		respBody string
		wantErr  string
	}{
		{
			name:     "Token expired",
			token:    "eyJraWQiOiI0REFFckV4d1ZkNEd3SzBpUEMrUjVqU1I2T1dDQTZvbWNOQmZ0aDN3NDV3PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiJhN2FlNWNlZi0wZGQyLTQyYWItYjRkYy0zMWI5Yzg1N2UzMzEiLCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAuZXUtd2VzdC0xLmFtYXpvbmF3cy5jb21cL2V1LXdlc3QtMV9VWWgzM1dlbGoiLCJ2ZXJzaW9uIjoyLCJjbGllbnRfaWQiOiI2c3RkZ3Y0ZDYxcmJuNmZka3FqYmEwN3ZoNCIsImV2ZW50X2lkIjoiYjJkNjgzNWUtZTMwNS00NmI0LTliNDgtNjU1OTNhNjlhOWQ2IiwidG9rZW5fdXNlIjoiYWNjZXNzIiwic2NvcGUiOiJvcGVuaWQgZW1haWwiLCJhdXRoX3RpbWUiOjE2ODI5MjI1NDksImV4cCI6MTY4MjkyNjE0OSwiaWF0IjoxNjgyOTIyNTQ5LCJqdGkiOiI0N2Y0YmMzMC01ZWU1LTRkMmUtOTZlYS0yNGI0OTdlMDA2ODciLCJ1c2VybmFtZSI6ImRhdnN1YXBhcyJ9.GRvxrannCQpMEMnE1WYXa39BxP62hF5BfnQcvjsBB3odnGCLBIghwE2wKAG0ufHkvM6Z3uoHPDJesleMKkcGny6RPaQNSRfvvgklQpjTXr5WL-lXEOMjTKzNIZdamiwbDnmQ-qF_zQQpVlDvdPp-zpv-ZwwakETFcxSmGLVfux9PJcPZLs-UmB5wn2dPGUDnxY19_uVY8nPoTB9-ZgbCb-fOoMmbmvGKMWHRZpzQS4O0CUsC5anhIAvBAcTlZa85N4v9U_iPfQO_Uvgy2qqSRsx2e3eBqefrFwVXyaB75oe_SiUXmKYfzXywSN3RCTlu9K-tGCRi7dCC4jRQxc_wIQ",
			respBody: "{\"keys\":[{\"alg\":\"RS256\",\"e\":\"AQAB\",\"kid\":\"Jf50SRGXc8qe5vX1zvK2Li3qJU4j8tgTIcKwpa6yTk4=\",\"kty\":\"RSA\",\"n\":\"1et6-M1xXbFvd9dgCZ5zIOCG57stSbrGIEtFeTG9ULkHts3nlIfH6CSYunuHgGhGKUEH4UdcP75QGrlgJN63nTgJkhL5c4j-eBuWbe6JJNXtMqxGXoRLl1TJYJaSP3KotRMhsi3Df-zDG6tbZ_yH7ulbMzu_U5HZzmaY3cwJ5qB31-uOosnPCDh3Be1ay5YdYzdY8cz_OmHXfonIHSGZMtY6Xes5pRpHlz_ZApW9t1T2frhANXv27VAGiD0xpBFuXEYCDZYe8V68BAM_2NwAIYyDbctk3iu_nA1jw-xFRKVyvbrSh1GPQc0JnW0TwE8_43vWybeE3iGg6TQ67d8bDw\",\"use\":\"sig\"},{\"alg\":\"RS256\",\"e\":\"AQAB\",\"kid\":\"4DAErExwVd4GwK0iPC+R5jSR6OWCA6omcNBfth3w45w=\",\"kty\":\"RSA\",\"n\":\"nK12U2xLIGAzQFyuxebztzrUFE9T3vQUCji8PlY-eRt8EOySiKqxUpVjpkwGxY2jFKZvJS9A1MJf9qC3If_zI6d5i4gNwmvl9hIboWIhT2SUCm5S-BzgJl8Zt86ZLYxBoC993vRQT0pYdXu6AB21UxWn6v4Q56_EA9n5asdwI-9hC22dCtGOx-qThxuOdsdS6mrWUsOgHwBx7uLQJEk2ZoeI_h4vupATGQQ4ZWXnYZrkm3MhMwG68J6qTvTvj5d5rdok4qIIg_GQGht_OZorZiJCNQz0c6Zea95JHhGBLfkbuwem5znvpVZUTa2HxN_c1HlY3_kRd0mUomghXdT4Gw\",\"use\":\"sig\"}]}",
			wantErr:  "token is expired",
		},
		{
			name:     "Kid not found",
			token:    "eyJraWQiOiI0REFFckV4d1ZkNEd3SzBpUEMrUjVqU1I2T1dDQTZvbWNOQmZ0aDN3NDV3PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiJhN2FlNWNlZi0wZGQyLTQyYWItYjRkYy0zMWI5Yzg1N2UzMzEiLCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAuZXUtd2VzdC0xLmFtYXpvbmF3cy5jb21cL2V1LXdlc3QtMV9VWWgzM1dlbGoiLCJ2ZXJzaW9uIjoyLCJjbGllbnRfaWQiOiI2c3RkZ3Y0ZDYxcmJuNmZka3FqYmEwN3ZoNCIsImV2ZW50X2lkIjoiYjJkNjgzNWUtZTMwNS00NmI0LTliNDgtNjU1OTNhNjlhOWQ2IiwidG9rZW5fdXNlIjoiYWNjZXNzIiwic2NvcGUiOiJvcGVuaWQgZW1haWwiLCJhdXRoX3RpbWUiOjE2ODI5MjI1NDksImV4cCI6MTY4MjkyNjE0OSwiaWF0IjoxNjgyOTIyNTQ5LCJqdGkiOiI0N2Y0YmMzMC01ZWU1LTRkMmUtOTZlYS0yNGI0OTdlMDA2ODciLCJ1c2VybmFtZSI6ImRhdnN1YXBhcyJ9.GRvxrannCQpMEMnE1WYXa39BxP62hF5BfnQcvjsBB3odnGCLBIghwE2wKAG0ufHkvM6Z3uoHPDJesleMKkcGny6RPaQNSRfvvgklQpjTXr5WL-lXEOMjTKzNIZdamiwbDnmQ-qF_zQQpVlDvdPp-zpv-ZwwakETFcxSmGLVfux9PJcPZLs-UmB5wn2dPGUDnxY19_uVY8nPoTB9-ZgbCb-fOoMmbmvGKMWHRZpzQS4O0CUsC5anhIAvBAcTlZa85N4v9U_iPfQO_Uvgy2qqSRsx2e3eBqefrFwVXyaB75oe_SiUXmKYfzXywSN3RCTlu9K-tGCRi7dCC4jRQxc_wIQ",
			respBody: "{\"keys\":[{\"alg\":\"RS256\",\"e\":\"AQAB\",\"kid\":\"1=\",\"kty\":\"RSA\",\"n\":\"1et6-M1xXbFvd9dgCZ5zIOCG57stSbrGIEtFeTG9ULkHts3nlIfH6CSYunuHgGhGKUEH4UdcP75QGrlgJN63nTgJkhL5c4j-eBuWbe6JJNXtMqxGXoRLl1TJYJaSP3KotRMhsi3Df-zDG6tbZ_yH7ulbMzu_U5HZzmaY3cwJ5qB31-uOosnPCDh3Be1ay5YdYzdY8cz_OmHXfonIHSGZMtY6Xes5pRpHlz_ZApW9t1T2frhANXv27VAGiD0xpBFuXEYCDZYe8V68BAM_2NwAIYyDbctk3iu_nA1jw-xFRKVyvbrSh1GPQc0JnW0TwE8_43vWybeE3iGg6TQ67d8bDw\",\"use\":\"sig\"},{\"alg\":\"RS256\",\"e\":\"AQAB\",\"kid\":\"2\",\"kty\":\"RSA\",\"n\":\"nK12U2xLIGAzQFyuxebztzrUFE9T3vQUCji8PlY-eRt8EOySiKqxUpVjpkwGxY2jFKZvJS9A1MJf9qC3If_zI6d5i4gNwmvl9hIboWIhT2SUCm5S-BzgJl8Zt86ZLYxBoC993vRQT0pYdXu6AB21UxWn6v4Q56_EA9n5asdwI-9hC22dCtGOx-qThxuOdsdS6mrWUsOgHwBx7uLQJEk2ZoeI_h4vupATGQQ4ZWXnYZrkm3MhMwG68J6qTvTvj5d5rdok4qIIg_GQGht_OZorZiJCNQz0c6Zea95JHhGBLfkbuwem5znvpVZUTa2HxN_c1HlY3_kRd0mUomghXdT4Gw\",\"use\":\"sig\"}]}",
			wantErr:  "Key Not found into JWK",
		},
		{
			name:     "RawE decode Error",
			token:    "eyJraWQiOiI0REFFckV4d1ZkNEd3SzBpUEMrUjVqU1I2T1dDQTZvbWNOQmZ0aDN3NDV3PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiJhN2FlNWNlZi0wZGQyLTQyYWItYjRkYy0zMWI5Yzg1N2UzMzEiLCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAuZXUtd2VzdC0xLmFtYXpvbmF3cy5jb21cL2V1LXdlc3QtMV9VWWgzM1dlbGoiLCJ2ZXJzaW9uIjoyLCJjbGllbnRfaWQiOiI2c3RkZ3Y0ZDYxcmJuNmZka3FqYmEwN3ZoNCIsImV2ZW50X2lkIjoiYjJkNjgzNWUtZTMwNS00NmI0LTliNDgtNjU1OTNhNjlhOWQ2IiwidG9rZW5fdXNlIjoiYWNjZXNzIiwic2NvcGUiOiJvcGVuaWQgZW1haWwiLCJhdXRoX3RpbWUiOjE2ODI5MjI1NDksImV4cCI6MTY4MjkyNjE0OSwiaWF0IjoxNjgyOTIyNTQ5LCJqdGkiOiI0N2Y0YmMzMC01ZWU1LTRkMmUtOTZlYS0yNGI0OTdlMDA2ODciLCJ1c2VybmFtZSI6ImRhdnN1YXBhcyJ9.GRvxrannCQpMEMnE1WYXa39BxP62hF5BfnQcvjsBB3odnGCLBIghwE2wKAG0ufHkvM6Z3uoHPDJesleMKkcGny6RPaQNSRfvvgklQpjTXr5WL-lXEOMjTKzNIZdamiwbDnmQ-qF_zQQpVlDvdPp-zpv-ZwwakETFcxSmGLVfux9PJcPZLs-UmB5wn2dPGUDnxY19_uVY8nPoTB9-ZgbCb-fOoMmbmvGKMWHRZpzQS4O0CUsC5anhIAvBAcTlZa85N4v9U_iPfQO_Uvgy2qqSRsx2e3eBqefrFwVXyaB75oe_SiUXmKYfzXywSN3RCTlu9K-tGCRi7dCC4jRQxc_wIQ",
			respBody: "{\"keys\":[{\"alg\":\"RS256\",\"e\":\"|\",\"kid\":\"Jf50SRGXc8qe5vX1zvK2Li3qJU4j8tgTIcKwpa6yTk4=\",\"kty\":\"RSA\",\"n\":\"1et6-M1xXbFvd9dgCZ5zIOCG57stSbrGIEtFeTG9ULkHts3nlIfH6CSYunuHgGhGKUEH4UdcP75QGrlgJN63nTgJkhL5c4j-eBuWbe6JJNXtMqxGXoRLl1TJYJaSP3KotRMhsi3Df-zDG6tbZ_yH7ulbMzu_U5HZzmaY3cwJ5qB31-uOosnPCDh3Be1ay5YdYzdY8cz_OmHXfonIHSGZMtY6Xes5pRpHlz_ZApW9t1T2frhANXv27VAGiD0xpBFuXEYCDZYe8V68BAM_2NwAIYyDbctk3iu_nA1jw-xFRKVyvbrSh1GPQc0JnW0TwE8_43vWybeE3iGg6TQ67d8bDw\",\"use\":\"sig\"},{\"alg\":\"RS256\",\"e\":\"|\",\"kid\":\"4DAErExwVd4GwK0iPC+R5jSR6OWCA6omcNBfth3w45w=\",\"kty\":\"RSA\",\"n\":\"nK12U2xLIGAzQFyuxebztzrUFE9T3vQUCji8PlY-eRt8EOySiKqxUpVjpkwGxY2jFKZvJS9A1MJf9qC3If_zI6d5i4gNwmvl9hIboWIhT2SUCm5S-BzgJl8Zt86ZLYxBoC993vRQT0pYdXu6AB21UxWn6v4Q56_EA9n5asdwI-9hC22dCtGOx-qThxuOdsdS6mrWUsOgHwBx7uLQJEk2ZoeI_h4vupATGQQ4ZWXnYZrkm3MhMwG68J6qTvTvj5d5rdok4qIIg_GQGht_OZorZiJCNQz0c6Zea95JHhGBLfkbuwem5znvpVZUTa2HxN_c1HlY3_kRd0mUomghXdT4Gw\",\"use\":\"sig\"}]}",
			wantErr:  "Converting JWT token: Decoding E",
		},
		{
			name:     "RawN decode Error",
			token:    "eyJraWQiOiI0REFFckV4d1ZkNEd3SzBpUEMrUjVqU1I2T1dDQTZvbWNOQmZ0aDN3NDV3PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiJhN2FlNWNlZi0wZGQyLTQyYWItYjRkYy0zMWI5Yzg1N2UzMzEiLCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAuZXUtd2VzdC0xLmFtYXpvbmF3cy5jb21cL2V1LXdlc3QtMV9VWWgzM1dlbGoiLCJ2ZXJzaW9uIjoyLCJjbGllbnRfaWQiOiI2c3RkZ3Y0ZDYxcmJuNmZka3FqYmEwN3ZoNCIsImV2ZW50X2lkIjoiYjJkNjgzNWUtZTMwNS00NmI0LTliNDgtNjU1OTNhNjlhOWQ2IiwidG9rZW5fdXNlIjoiYWNjZXNzIiwic2NvcGUiOiJvcGVuaWQgZW1haWwiLCJhdXRoX3RpbWUiOjE2ODI5MjI1NDksImV4cCI6MTY4MjkyNjE0OSwiaWF0IjoxNjgyOTIyNTQ5LCJqdGkiOiI0N2Y0YmMzMC01ZWU1LTRkMmUtOTZlYS0yNGI0OTdlMDA2ODciLCJ1c2VybmFtZSI6ImRhdnN1YXBhcyJ9.GRvxrannCQpMEMnE1WYXa39BxP62hF5BfnQcvjsBB3odnGCLBIghwE2wKAG0ufHkvM6Z3uoHPDJesleMKkcGny6RPaQNSRfvvgklQpjTXr5WL-lXEOMjTKzNIZdamiwbDnmQ-qF_zQQpVlDvdPp-zpv-ZwwakETFcxSmGLVfux9PJcPZLs-UmB5wn2dPGUDnxY19_uVY8nPoTB9-ZgbCb-fOoMmbmvGKMWHRZpzQS4O0CUsC5anhIAvBAcTlZa85N4v9U_iPfQO_Uvgy2qqSRsx2e3eBqefrFwVXyaB75oe_SiUXmKYfzXywSN3RCTlu9K-tGCRi7dCC4jRQxc_wIQ",
			respBody: "{\"keys\":[{\"alg\":\"RS256\",\"e\":\"AQAB\",\"kid\":\"Jf50SRGXc8qe5vX1zvK2Li3qJU4j8tgTIcKwpa6yTk4=\",\"kty\":\"RSA\",\"n\":\"|\",\"use\":\"sig\"},{\"alg\":\"RS256\",\"e\":\"AQAB\",\"kid\":\"4DAErExwVd4GwK0iPC+R5jSR6OWCA6omcNBfth3w45w=\",\"kty\":\"RSA\",\"n\":\"|\",\"use\":\"sig\"}]}",
			wantErr:  "Converting JWT token: Decoding N",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.respBody))
			}

			s := httptest.NewServer(http.HandlerFunc(reg))
			defer s.Close()

			k := &auth.JWT{
				JWKFetch: auth.NewJWKFetch(s.URL),
			}

			if _, err := k.ParseJWT(tt.token); err != nil {
				if len(tt.wantErr) == 0 {
					assert.NoError(t, err, "Error")
				}
				assert.ErrorContains(t, err, tt.wantErr, "Error")
			}
		})
	}
}
