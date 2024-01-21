/*
 *   Copyright (c) 2022 ELIPCERO
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

package strings_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/pkg/strings"
)

func TestStrings_Concat(t *testing.T) {
	t.Parallel()

	c := strings.Concat("a", "b", "c")
	assert.Equal(t, c, "abc")
}

func TestStrings_fmt(t *testing.T) {
	t.Parallel()

	type args struct {
		text      string
		fmtValues []strings.FormatValue
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Text + Any params",
			args: args{
				text:      "Text",
				fmtValues: []strings.FormatValue{},
			},
			want: "Text",
		},
		{
			name: "Text + only one param",
			args: args{
				text: "Text",
				fmtValues: []strings.FormatValue{
					strings.FMTValue("key1", "Value1"),
				},
			},
			want: "Text (key1: Value1, )",
		},
		{
			name: "Text + Several params",
			args: args{
				text: "Text",
				fmtValues: []strings.FormatValue{
					strings.FMTValue("key1", "Value1"),
					strings.FMTValue("key2", "Value2"),
				},
			},
			want: "Text (key1: Value1, key2: Value2, )",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := strings.Format(tt.args.text, tt.args.fmtValues...)
			assert.Equal(t, tt.want, res)
		})
	}
}
