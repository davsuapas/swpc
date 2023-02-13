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

func TestEncrypt_Encrypt_Ok(t *testing.T) {
	t.Parallel()

	e, err := crypto.Encrypt("123", "1234567891234567")
	if assert.NoError(t, err) {
		assert.Equal(t, e, "iO7a")
	}
}

func TestEncrypt_Encrypt_Error_Minimal_Key(t *testing.T) {
	t.Parallel()

	_, err := crypto.Encrypt("123", "123")
	assert.Error(t, err)
}