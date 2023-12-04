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

package arrays

// Has seeks if a array has a item
func Has[T comparable](arr []T, item T) bool {
	for _, v := range arr {
		if item == v {
			return true
		}
	}

	return false
}

// Remove removes the elements of an array that match the positions
func Remove[T any, T1 uint8 | uint16 | uint32 | uint64](
	arr []T, pos ...T1) []T {
	//
	pnil := -1

	for i := 0; i < len(arr); i++ {
		if Has(pos, T1(i)) {
			if pnil == -1 {
				pnil = i
			}
		} else {
			if pnil > -1 {
				arr[pnil] = arr[i]
				pnil++
			}
		}
	}

	if pnil == -1 {
		return arr
	}

	return arr[:pnil]
}
