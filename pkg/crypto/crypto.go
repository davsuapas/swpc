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

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"github.com/pkg/errors"
)

// Encrypt encrypts or hide any classified text
func Encrypt(text string, key string) (string, error) {
	var bytes = []byte{35, 26, 17, 44, 85, 35, 25, 74, 87, 65, 88, 98, 68, 32, 44, 05}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", errors.Wrap(err, "New cipher")
	}

	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, bytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)

	return encode(cipherText), nil
}

func encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
