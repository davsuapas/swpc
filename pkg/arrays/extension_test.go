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

package arrays_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/pkg/arrays"
)

type test struct {
	id uint8
}

func TestArraysHas(t *testing.T) {
	t.Parallel()

	type args struct {
		arr  []uint8
		item uint8
	}

	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "Array 3 items. Has: 1",
			args: args{
				arr:  []uint8{1, 2, 3},
				item: 1,
			},
			expected: true,
		},
		{
			name: "Array 3 items. Has: 3",
			args: args{
				arr:  []uint8{1, 2, 3},
				item: 3,
			},
			expected: true,
		},
		{
			name: "Array 3 items. Not has",
			args: args{
				arr:  []uint8{1, 2, 3},
				item: 6,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := arrays.Has(tt.args.arr, tt.args.item)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestArraysRemove(t *testing.T) {
	t.Parallel()

	type args struct {
		arr []test
		pos []uint64
	}

	tests := []struct {
		name     string
		args     args
		expected []test
	}{
		{
			name: "Array 3 items. Remove: 1",
			args: args{
				arr: []test{{id: 1}, {id: 2}, {id: 3}},
				pos: []uint64{0},
			},
			expected: []test{{id: 2}, {id: 3}},
		},
		{
			name: "Array 3 items. Remove: 2",
			args: args{
				arr: []test{{id: 1}, {id: 2}, {id: 3}},
				pos: []uint64{1},
			},
			expected: []test{{id: 1}, {id: 3}},
		},
		{
			name: "Array 3 items. Remove: 3",
			args: args{
				arr: []test{{id: 1}, {id: 2}, {id: 3}},
				pos: []uint64{2},
			},
			expected: []test{{id: 1}, {id: 2}},
		},
		{
			name: "Array 3 items. Remove: 1, 2",
			args: args{
				arr: []test{{id: 1}, {id: 2}, {id: 3}},
				pos: []uint64{0, 1},
			},
			expected: []test{{id: 3}},
		},
		{
			name: "Array 3 items. Remove: 1, 3",
			args: args{
				arr: []test{{id: 1}, {id: 2}, {id: 3}},
				pos: []uint64{0, 2},
			},
			expected: []test{{id: 2}},
		},
		{
			name: "Array 3 items. Remove: 2, 3",
			args: args{
				arr: []test{{id: 1}, {id: 2}, {id: 3}},
				pos: []uint64{1, 2},
			},
			expected: []test{{id: 1}},
		},
		{
			name: "Array 3 items. Remove: All",
			args: args{
				arr: []test{{id: 1}, {id: 2}, {id: 3}},
				pos: []uint64{0, 1, 2},
			},
			expected: []test{},
		},
		{
			name: "Array 3 items. Remove: None",
			args: args{
				arr: []test{{id: 1}, {id: 2}, {id: 3}},
				pos: []uint64{},
			},
			expected: []test{{id: 1}, {id: 2}, {id: 3}},
		},
		{
			name: "Array 1 item. Remove: 1",
			args: args{
				arr: []test{{id: 1}},
				pos: []uint64{0},
			},
			expected: []test{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := arrays.Remove(tt.args.arr, tt.args.pos...)

			assert.Equal(t, tt.expected, actual)
		})
	}
}
