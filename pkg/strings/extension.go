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

package strings

import (
	"strconv"
	"strings"
)

// Concat joins many strings
func Concat(params ...string) string {
	var b strings.Builder
	b.Grow(len(params) * 5)
	for _, str := range params {
		b.WriteString(str)
	}
	return b.String()
}

// ConvertBytes converts array of byte to string without escaping to heap
func ConvertBytes(s []byte) string {
	var b strings.Builder
	b.Grow(len(s))
	b.Write(s)
	return b.String()
}

// Convertuint32 converts a uint64 to string
func Convertuint32(u uint32) string {
	return strconv.FormatUint(uint64(u), 10)
}

func Lpad(s1 string, length int, s2 string) string {
	s1l := len(s1)
	if length <= s1l {
		return s1
	}
	return Concat(strings.Repeat(s2, length-s1l), s1)
}
