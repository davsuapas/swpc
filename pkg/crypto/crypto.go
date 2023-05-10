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
	"crypto/rand"
	e "errors"
	"io"

	"github.com/pkg/errors"
)

var (
	errKeyLen = e.New("the key length must be 32 bytes")
)

const (
	errEcrypt    = "Encrypt RSA"
	errDesEcrypt = "Des-encrypt RSA"
)

// encrypt encrypts with RSA
func Encrypt(key []byte, plaintext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errKeyLen
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, errEcrypt)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, errors.Wrap(err, errEcrypt)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, errEcrypt)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt desencrypts with RSA
func Decrypt(key []byte, ciphertext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errKeyLen
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, errDesEcrypt)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, errors.Wrap(err, errDesEcrypt)
	}

	nonceSize := gcm.NonceSize()

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	decryptedText, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, errDesEcrypt)
	}

	return decryptedText, nil
}
