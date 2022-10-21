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

package main

import (
	"flag"
	"fmt"

	cryptoi "github.com/swpoolcontroller/internal/crypto"
	"github.com/swpoolcontroller/pkg/crypto"
)

func main() {
	text := flag.String("text", "", "Text to encrypt")
	flag.Parse()

	if len(*text) == 0 {
		panic("The text must be filled in. Execute the command with -help option")
	}

	r, err := crypto.Encrypt(*text, cryptoi.Key)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(r)
}
