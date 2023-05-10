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

package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/pkg/crypto"
)

var key = []byte("123456789asdfghjklzxcvbnmqwertyu")

func TestEncrypt_Ok(t *testing.T) {
	t.Parallel()

	e, _ := crypto.Encrypt(key, []byte("1234567891234567"))

	assert.Equal(t, len(e), 44)
}

func TestEncrypt_Error_Minimal_Key(t *testing.T) {
	t.Parallel()

	_, err := crypto.Encrypt([]byte("123"), []byte("123"))

	assert.Error(t, err)
}

func TestDecrypt_Ok(t *testing.T) {
	t.Parallel()

	e, _ := crypto.Encrypt(key, []byte("1234567891234567"))
	d, _ := crypto.Decrypt(key, e)

	assert.Equal(t, string(d), "1234567891234567")
}

func TestDecrypt_Error_Minimal_Key(t *testing.T) {
	t.Parallel()

	_, err := crypto.Decrypt([]byte("123"), []byte("123"))
	assert.Error(t, err)
}
