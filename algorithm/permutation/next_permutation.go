// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
//
// This file is part of gogo.
//
// gogo is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package permutation

// Transform data to its next permutation in lexical order.
// It returns false if data.Len() == 0 or the permutations are exhausted,
// and true otherwise.
// Time complexity: O(n), where n = data.Len().
func NextPermutation(data Interface) bool {
	if data == nil {
		return false
	}
	i := data.Len() - 2
	for i >= 0 && !data.Less(i, i+1) {
		i--
	}
	if i < 0 {
		return false
	}
	j := i + 1
	k := data.Len()
	m := (j + k) / 2
	for j != m {
		if data.Less(i, m) {
			j = m
		} else {
			k = m
		}
		m = (j + k) / 2
	}
	data.Swap(i, j)
	for j, k = i+1, data.Len()-1; j < k; j, k = j+1, k-1 {
		data.Swap(j, k)
	}
	return true
}
