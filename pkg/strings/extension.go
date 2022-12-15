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
	"strings"
)

// Concat joins many strings
func Concat(params ...string) string {
	var builder strings.Builder

	builder.Grow(len(params) * 5)

	for _, str := range params {
		builder.WriteString(str)
	}

	return builder.String()
}

// FormatValue defines the variables values to format
type FormatValue struct {
	key   string
	value string
}

// FMTValue builds the variables values to format
func FMTValue(key string, value string) FormatValue {
	return FormatValue{
		key:   key,
		value: value,
	}
}

// Format formats the text with variables values without escape to heap
func Format(text string, fmtValues ...FormatValue) string {
	var builder strings.Builder

	builder.Grow(len(text) + (len(fmtValues) * 5))

	builder.WriteString(text)

	if len(fmtValues) > 0 {
		builder.WriteString(" (")
	}

	for _, fv := range fmtValues {
		builder.WriteString(fv.key)
		builder.WriteString(": ")
		builder.WriteString(fv.value)
		builder.WriteString(", ")
	}

	if len(fmtValues) > 0 {
		builder.WriteString(")")
	}

	return builder.String()
}
