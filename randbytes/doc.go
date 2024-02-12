// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

// Package randbytes provides interfaces and functions to generate random bytes.
//
// This package is based on the standard library math/rand/v2,
// which is pseudorandom but reproducible.
// This package is for scenarios that require not only random data
// but also the same data next time, such as data for testing.
// Without the need for reproduction, use crypto/rand.Read instead.
//
// For better performance, all functions in this package are unsafe
// for concurrency unless otherwise specified.
package randbytes
