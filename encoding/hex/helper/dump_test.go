// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

package helper

import (
	"bytes"
	"testing"

	"github.com/donyori/gogo/encoding/hex"
)

var testExampleDumpOutputs = [...]string{
	`00000000: 4865 6c6c 6f20 776f 726c 6421 20e4 bda0 | "Hello world! 你"
00000010: e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 | "好，世界！H"
00000020: 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 | "ello world! 你\xe5"
00000030: a5bd efbc 8ce4 b896 e795 8cef bc81 4865 | "\xa5\xbd，世界！He"
00000040: 6c6c 6f20 776f 726c 6421 20e4 bda0 e5a5 | "llo world! 你\xe5\xa5"
00000050: bdef bc8c e4b8 96e7 958c efbc 8148 656c | "\xbd，世界！Hel"
00000060: 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd | "lo world! 你好"
00000070: efbc 8ce4 b896 e795 8cef bc81 4865 6c6c | "，世界！Hell"
00000080: 6f20 776f 726c 6421 20e4 bda0 e5a5 bdef | "o world! 你好\xef"
00000090: bc8c e4b8 96e7 958c efbc 8148 656c 6c6f | "\xbc\x8c世界！Hello"
000000a0: 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc | " world! 你好\xef\xbc"
000000b0: 8ce4 b896 e795 8cef bc81 4865 6c6c 6f20 | "\x8c世界！Hello "
000000c0: 776f 726c 6421 20e4 bda0 e5a5 bdef bc8c | "world! 你好，"
000000d0: e4b8 96e7 958c efbc 8148 656c 6c6f 2077 | "世界！Hello w"
000000e0: 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 | "orld! 你好，\xe4"
000000f0: b896 e795 8cef bc81 4865 6c6c 6f20 776f | "\xb8\x96界！Hello wo"
00000100: 726c 6421 20e4 bda0 e5a5 bdef bc8c e4b8 | "rld! 你好，\xe4\xb8"
00000110: 96e7 958c efbc 8148 656c 6c6f 2077 6f72 | "\x96界！Hello wor"
00000120: 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 | "ld! 你好，世"
00000130: e795 8cef bc81 4865 6c6c 6f20 776f 726c | "界！Hello worl"
00000140: 6421 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 | "d! 你好，世\xe7"
00000150: 958c efbc 8148 656c 6c6f 2077 6f72 6c64 | "\x95\x8c！Hello world"
00000160: 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 | "! 你好，世\xe7\x95"
00000170: 8cef bc81 4865 6c6c 6f20 776f 726c 6421 | "\x8c！Hello world!"
00000180: 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 958c | " 你好，世界"
00000190: efbc 8148 656c 6c6f 2077 6f72 6c64 2120 | "！Hello world! "
000001a0: e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef | "你好，世界\xef"
000001b0: bc81 4865 6c6c 6f20 776f 726c 6421 20e4 | "\xbc\x81Hello world! \xe4"
000001c0: bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc | "\xbd\xa0好，世界\xef\xbc"
000001d0: 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd | "\x81Hello world! \xe4\xbd"
000001e0: a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 | "\xa0好，世界！"
000001f0: 4865 6c6c 6f20 776f 726c 6421 20e4 bda0 | "Hello world! 你"
00000200: e5a5 bdef bc8c e4b8 96e7 958c efbc 81   | "好，世界！"
`,

	`00000000: 4865 6c6c 6f20 776f | "Hello wo"
00000008: 726c 6421 20e4 bda0 | "rld! 你"
00000010: e5a5 bdef bc8c e4b8 | "好，\xe4\xb8"
00000018: 96e7 958c efbc 8148 | "\x96界！H"
00000020: 656c 6c6f 2077 6f72 | "ello wor"
00000028: 6c64 2120 e4bd a0e5 | "ld! 你\xe5"
00000030: a5bd efbc 8ce4 b896 | "\xa5\xbd，世"
00000038: e795 8cef bc81 4865 | "界！He"
00000040: 6c6c 6f20 776f 726c | "llo worl"
00000048: 6421 20e4 bda0 e5a5 | "d! 你\xe5\xa5"
00000050: bdef bc8c e4b8 96e7 | "\xbd，世\xe7"
00000058: 958c efbc 8148 656c | "\x95\x8c！Hel"
00000060: 6c6f 2077 6f72 6c64 | "lo world"
00000068: 2120 e4bd a0e5 a5bd | "! 你好"
00000070: efbc 8ce4 b896 e795 | "，世\xe7\x95"
00000078: 8cef bc81 4865 6c6c | "\x8c！Hell"
00000080: 6f20 776f 726c 6421 | "o world!"
00000088: 20e4 bda0 e5a5 bdef | " 你好\xef"
00000090: bc8c e4b8 96e7 958c | "\xbc\x8c世界"
00000098: efbc 8148 656c 6c6f | "！Hello"
000000a0: 2077 6f72 6c64 2120 | " world! "
000000a8: e4bd a0e5 a5bd efbc | "你好\xef\xbc"
000000b0: 8ce4 b896 e795 8cef | "\x8c世界\xef"
000000b8: bc81 4865 6c6c 6f20 | "\xbc\x81Hello "
000000c0: 776f 726c 6421 20e4 | "world! \xe4"
000000c8: bda0 e5a5 bdef bc8c | "\xbd\xa0好，"
000000d0: e4b8 96e7 958c efbc | "世界\xef\xbc"
000000d8: 8148 656c 6c6f 2077 | "\x81Hello w"
000000e0: 6f72 6c64 2120 e4bd | "orld! \xe4\xbd"
000000e8: a0e5 a5bd efbc 8ce4 | "\xa0好，\xe4"
000000f0: b896 e795 8cef bc81 | "\xb8\x96界！"
000000f8: 4865 6c6c 6f20 776f | "Hello wo"
00000100: 726c 6421 20e4 bda0 | "rld! 你"
00000108: e5a5 bdef bc8c e4b8 | "好，\xe4\xb8"
00000110: 96e7 958c efbc 8148 | "\x96界！H"
00000118: 656c 6c6f 2077 6f72 | "ello wor"
00000120: 6c64 2120 e4bd a0e5 | "ld! 你\xe5"
00000128: a5bd efbc 8ce4 b896 | "\xa5\xbd，世"
00000130: e795 8cef bc81 4865 | "界！He"
00000138: 6c6c 6f20 776f 726c | "llo worl"
00000140: 6421 20e4 bda0 e5a5 | "d! 你\xe5\xa5"
00000148: bdef bc8c e4b8 96e7 | "\xbd，世\xe7"
00000150: 958c efbc 8148 656c | "\x95\x8c！Hel"
00000158: 6c6f 2077 6f72 6c64 | "lo world"
00000160: 2120 e4bd a0e5 a5bd | "! 你好"
00000168: efbc 8ce4 b896 e795 | "，世\xe7\x95"
00000170: 8cef bc81 4865 6c6c | "\x8c！Hell"
00000178: 6f20 776f 726c 6421 | "o world!"
00000180: 20e4 bda0 e5a5 bdef | " 你好\xef"
00000188: bc8c e4b8 96e7 958c | "\xbc\x8c世界"
00000190: efbc 8148 656c 6c6f | "！Hello"
00000198: 2077 6f72 6c64 2120 | " world! "
000001a0: e4bd a0e5 a5bd efbc | "你好\xef\xbc"
000001a8: 8ce4 b896 e795 8cef | "\x8c世界\xef"
000001b0: bc81 4865 6c6c 6f20 | "\xbc\x81Hello "
000001b8: 776f 726c 6421 20e4 | "world! \xe4"
000001c0: bda0 e5a5 bdef bc8c | "\xbd\xa0好，"
000001c8: e4b8 96e7 958c efbc | "世界\xef\xbc"
000001d0: 8148 656c 6c6f 2077 | "\x81Hello w"
000001d8: 6f72 6c64 2120 e4bd | "orld! \xe4\xbd"
000001e0: a0e5 a5bd efbc 8ce4 | "\xa0好，\xe4"
000001e8: b896 e795 8cef bc81 | "\xb8\x96界！"
000001f0: 4865 6c6c 6f20 776f | "Hello wo"
000001f8: 726c 6421 20e4 bda0 | "rld! 你"
00000200: e5a5 bdef bc8c e4b8 | "好，\xe4\xb8"
00000208: 96e7 958c efbc 81   | "\x96界！"
`,

	`00000000: 48 65 6c 6c 6f 20 77 6f 72 6c 64 21 20 e4 bd | "Hello world! \xe4\xbd"
0000000f: a0 e5 a5 bd ef bc 8c e4 b8 96 e7 95 8c ef bc | "\xa0好，世界\xef\xbc"
0000001e: 81 48 65 6c 6c 6f 20 77 6f 72 6c 64 21 20 e4 | "\x81Hello world! \xe4"
0000002d: bd a0 e5 a5 bd ef bc 8c e4 b8 96 e7 95 8c ef | "\xbd\xa0好，世界\xef"
0000003c: bc 81 48 65 6c 6c 6f 20 77 6f 72 6c 64 21 20 | "\xbc\x81Hello world! "
0000004b: e4 bd a0 e5 a5 bd ef bc 8c e4 b8 96 e7 95 8c | "你好，世界"
0000005a: ef bc 81 48 65 6c 6c 6f 20 77 6f 72 6c 64 21 | "！Hello world!"
00000069: 20 e4 bd a0 e5 a5 bd ef bc 8c e4 b8 96 e7 95 | " 你好，世\xe7\x95"
00000078: 8c ef bc 81 48 65 6c 6c 6f 20 77 6f 72 6c 64 | "\x8c！Hello world"
00000087: 21 20 e4 bd a0 e5 a5 bd ef bc 8c e4 b8 96 e7 | "! 你好，世\xe7"
00000096: 95 8c ef bc 81 48 65 6c 6c 6f 20 77 6f 72 6c | "\x95\x8c！Hello worl"
000000a5: 64 21 20 e4 bd a0 e5 a5 bd ef bc 8c e4 b8 96 | "d! 你好，世"
000000b4: e7 95 8c ef bc 81 48 65 6c 6c 6f 20 77 6f 72 | "界！Hello wor"
000000c3: 6c 64 21 20 e4 bd a0 e5 a5 bd ef bc 8c e4 b8 | "ld! 你好，\xe4\xb8"
000000d2: 96 e7 95 8c ef bc 81 48 65 6c 6c 6f 20 77 6f | "\x96界！Hello wo"
000000e1: 72 6c 64 21 20 e4 bd a0 e5 a5 bd ef bc 8c e4 | "rld! 你好，\xe4"
000000f0: b8 96 e7 95 8c ef bc 81 48 65 6c 6c 6f 20 77 | "\xb8\x96界！Hello w"
000000ff: 6f 72 6c 64 21 20 e4 bd a0 e5 a5 bd ef bc 8c | "orld! 你好，"
0000010e: e4 b8 96 e7 95 8c ef bc 81 48 65 6c 6c 6f 20 | "世界！Hello "
0000011d: 77 6f 72 6c 64 21 20 e4 bd a0 e5 a5 bd ef bc | "world! 你好\xef\xbc"
0000012c: 8c e4 b8 96 e7 95 8c ef bc 81 48 65 6c 6c 6f | "\x8c世界！Hello"
0000013b: 20 77 6f 72 6c 64 21 20 e4 bd a0 e5 a5 bd ef | " world! 你好\xef"
0000014a: bc 8c e4 b8 96 e7 95 8c ef bc 81 48 65 6c 6c | "\xbc\x8c世界！Hell"
00000159: 6f 20 77 6f 72 6c 64 21 20 e4 bd a0 e5 a5 bd | "o world! 你好"
00000168: ef bc 8c e4 b8 96 e7 95 8c ef bc 81 48 65 6c | "，世界！Hel"
00000177: 6c 6f 20 77 6f 72 6c 64 21 20 e4 bd a0 e5 a5 | "lo world! 你\xe5\xa5"
00000186: bd ef bc 8c e4 b8 96 e7 95 8c ef bc 81 48 65 | "\xbd，世界！He"
00000195: 6c 6c 6f 20 77 6f 72 6c 64 21 20 e4 bd a0 e5 | "llo world! 你\xe5"
000001a4: a5 bd ef bc 8c e4 b8 96 e7 95 8c ef bc 81 48 | "\xa5\xbd，世界！H"
000001b3: 65 6c 6c 6f 20 77 6f 72 6c 64 21 20 e4 bd a0 | "ello world! 你"
000001c2: e5 a5 bd ef bc 8c e4 b8 96 e7 95 8c ef bc 81 | "好，世界！"
000001d1: 48 65 6c 6c 6f 20 77 6f 72 6c 64 21 20 e4 bd | "Hello world! \xe4\xbd"
000001e0: a0 e5 a5 bd ef bc 8c e4 b8 96 e7 95 8c ef bc | "\xa0好，世界\xef\xbc"
000001ef: 81 48 65 6c 6c 6f 20 77 6f 72 6c 64 21 20 e4 | "\x81Hello world! \xe4"
000001fe: bd a0 e5 a5 bd ef bc 8c e4 b8 96 e7 95 8c ef | "\xbd\xa0好，世界\xef"
0000020d: bc 81                                        | "\xbc\x81"
`,

	`00000000: 4865 6c6c 6f20 776f 726c 6421 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 4865 | "Hello world! 你好，世界！Hello world! 你好，世界！He"
00000040: 6c6c 6f20 776f 726c 6421 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 4865 6c6c | "llo world! 你好，世界！Hello world! 你好，世界！Hell"
00000080: 6f20 776f 726c 6421 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 4865 6c6c 6f20 | "o world! 你好，世界！Hello world! 你好，世界！Hello "
000000c0: 776f 726c 6421 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 4865 6c6c 6f20 776f | "world! 你好，世界！Hello world! 你好，世界！Hello wo"
00000100: 726c 6421 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 4865 6c6c 6f20 776f 726c | "rld! 你好，世界！Hello world! 你好，世界！Hello worl"
00000140: 6421 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 4865 6c6c 6f20 776f 726c 6421 | "d! 你好，世界！Hello world! 你好，世界！Hello world!"
00000180: 20e4 bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 4865 6c6c 6f20 776f 726c 6421 20e4 | " 你好，世界！Hello world! 你好，世界！Hello world! \xe4"
000001c0: bda0 e5a5 bdef bc8c e4b8 96e7 958c efbc 8148 656c 6c6f 2077 6f72 6c64 2120 e4bd a0e5 a5bd efbc 8ce4 b896 e795 8cef bc81 4865 6c6c 6f20 776f 726c 6421 20e4 bda0 | "\xbd\xa0好，世界！Hello world! 你好，世界！Hello world! 你"
00000200: e5a5 bdef bc8c e4b8 96e7 958c efbc 81                                                                                                                           | "好，世界！"
`,

	`00000000: 4865 6C6C 6F20 776F 726C 6421 20E4 BDA0 | "Hello world! 你"
00000010: E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 | "好，世界！H"
00000020: 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 | "ello world! 你\xe5"
00000030: A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 | "\xa5\xbd，世界！He"
00000040: 6C6C 6F20 776F 726C 6421 20E4 BDA0 E5A5 | "llo world! 你\xe5\xa5"
00000050: BDEF BC8C E4B8 96E7 958C EFBC 8148 656C | "\xbd，世界！Hel"
00000060: 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD | "lo world! 你好"
00000070: EFBC 8CE4 B896 E795 8CEF BC81 4865 6C6C | "，世界！Hell"
00000080: 6F20 776F 726C 6421 20E4 BDA0 E5A5 BDEF | "o world! 你好\xef"
00000090: BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F | "\xbc\x8c世界！Hello"
000000A0: 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC | " world! 你好\xef\xbc"
000000B0: 8CE4 B896 E795 8CEF BC81 4865 6C6C 6F20 | "\x8c世界！Hello "
000000C0: 776F 726C 6421 20E4 BDA0 E5A5 BDEF BC8C | "world! 你好，"
000000D0: E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 | "世界！Hello w"
000000E0: 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 | "orld! 你好，\xe4"
000000F0: B896 E795 8CEF BC81 4865 6C6C 6F20 776F | "\xb8\x96界！Hello wo"
00000100: 726C 6421 20E4 BDA0 E5A5 BDEF BC8C E4B8 | "rld! 你好，\xe4\xb8"
00000110: 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 | "\x96界！Hello wor"
00000120: 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 | "ld! 你好，世"
00000130: E795 8CEF BC81 4865 6C6C 6F20 776F 726C | "界！Hello worl"
00000140: 6421 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 | "d! 你好，世\xe7"
00000150: 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 | "\x95\x8c！Hello world"
00000160: 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 | "! 你好，世\xe7\x95"
00000170: 8CEF BC81 4865 6C6C 6F20 776F 726C 6421 | "\x8c！Hello world!"
00000180: 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 958C | " 你好，世界"
00000190: EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 | "！Hello world! "
000001A0: E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF | "你好，世界\xef"
000001B0: BC81 4865 6C6C 6F20 776F 726C 6421 20E4 | "\xbc\x81Hello world! \xe4"
000001C0: BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC | "\xbd\xa0好，世界\xef\xbc"
000001D0: 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD | "\x81Hello world! \xe4\xbd"
000001E0: A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 | "\xa0好，世界！"
000001F0: 4865 6C6C 6F20 776F 726C 6421 20E4 BDA0 | "Hello world! 你"
00000200: E5A5 BDEF BC8C E4B8 96E7 958C EFBC 81   | "好，世界！"
`,

	`00000000: 4865 6C6C 6F20 776F | "Hello wo"
00000008: 726C 6421 20E4 BDA0 | "rld! 你"
00000010: E5A5 BDEF BC8C E4B8 | "好，\xe4\xb8"
00000018: 96E7 958C EFBC 8148 | "\x96界！H"
00000020: 656C 6C6F 2077 6F72 | "ello wor"
00000028: 6C64 2120 E4BD A0E5 | "ld! 你\xe5"
00000030: A5BD EFBC 8CE4 B896 | "\xa5\xbd，世"
00000038: E795 8CEF BC81 4865 | "界！He"
00000040: 6C6C 6F20 776F 726C | "llo worl"
00000048: 6421 20E4 BDA0 E5A5 | "d! 你\xe5\xa5"
00000050: BDEF BC8C E4B8 96E7 | "\xbd，世\xe7"
00000058: 958C EFBC 8148 656C | "\x95\x8c！Hel"
00000060: 6C6F 2077 6F72 6C64 | "lo world"
00000068: 2120 E4BD A0E5 A5BD | "! 你好"
00000070: EFBC 8CE4 B896 E795 | "，世\xe7\x95"
00000078: 8CEF BC81 4865 6C6C | "\x8c！Hell"
00000080: 6F20 776F 726C 6421 | "o world!"
00000088: 20E4 BDA0 E5A5 BDEF | " 你好\xef"
00000090: BC8C E4B8 96E7 958C | "\xbc\x8c世界"
00000098: EFBC 8148 656C 6C6F | "！Hello"
000000A0: 2077 6F72 6C64 2120 | " world! "
000000A8: E4BD A0E5 A5BD EFBC | "你好\xef\xbc"
000000B0: 8CE4 B896 E795 8CEF | "\x8c世界\xef"
000000B8: BC81 4865 6C6C 6F20 | "\xbc\x81Hello "
000000C0: 776F 726C 6421 20E4 | "world! \xe4"
000000C8: BDA0 E5A5 BDEF BC8C | "\xbd\xa0好，"
000000D0: E4B8 96E7 958C EFBC | "世界\xef\xbc"
000000D8: 8148 656C 6C6F 2077 | "\x81Hello w"
000000E0: 6F72 6C64 2120 E4BD | "orld! \xe4\xbd"
000000E8: A0E5 A5BD EFBC 8CE4 | "\xa0好，\xe4"
000000F0: B896 E795 8CEF BC81 | "\xb8\x96界！"
000000F8: 4865 6C6C 6F20 776F | "Hello wo"
00000100: 726C 6421 20E4 BDA0 | "rld! 你"
00000108: E5A5 BDEF BC8C E4B8 | "好，\xe4\xb8"
00000110: 96E7 958C EFBC 8148 | "\x96界！H"
00000118: 656C 6C6F 2077 6F72 | "ello wor"
00000120: 6C64 2120 E4BD A0E5 | "ld! 你\xe5"
00000128: A5BD EFBC 8CE4 B896 | "\xa5\xbd，世"
00000130: E795 8CEF BC81 4865 | "界！He"
00000138: 6C6C 6F20 776F 726C | "llo worl"
00000140: 6421 20E4 BDA0 E5A5 | "d! 你\xe5\xa5"
00000148: BDEF BC8C E4B8 96E7 | "\xbd，世\xe7"
00000150: 958C EFBC 8148 656C | "\x95\x8c！Hel"
00000158: 6C6F 2077 6F72 6C64 | "lo world"
00000160: 2120 E4BD A0E5 A5BD | "! 你好"
00000168: EFBC 8CE4 B896 E795 | "，世\xe7\x95"
00000170: 8CEF BC81 4865 6C6C | "\x8c！Hell"
00000178: 6F20 776F 726C 6421 | "o world!"
00000180: 20E4 BDA0 E5A5 BDEF | " 你好\xef"
00000188: BC8C E4B8 96E7 958C | "\xbc\x8c世界"
00000190: EFBC 8148 656C 6C6F | "！Hello"
00000198: 2077 6F72 6C64 2120 | " world! "
000001A0: E4BD A0E5 A5BD EFBC | "你好\xef\xbc"
000001A8: 8CE4 B896 E795 8CEF | "\x8c世界\xef"
000001B0: BC81 4865 6C6C 6F20 | "\xbc\x81Hello "
000001B8: 776F 726C 6421 20E4 | "world! \xe4"
000001C0: BDA0 E5A5 BDEF BC8C | "\xbd\xa0好，"
000001C8: E4B8 96E7 958C EFBC | "世界\xef\xbc"
000001D0: 8148 656C 6C6F 2077 | "\x81Hello w"
000001D8: 6F72 6C64 2120 E4BD | "orld! \xe4\xbd"
000001E0: A0E5 A5BD EFBC 8CE4 | "\xa0好，\xe4"
000001E8: B896 E795 8CEF BC81 | "\xb8\x96界！"
000001F0: 4865 6C6C 6F20 776F | "Hello wo"
000001F8: 726C 6421 20E4 BDA0 | "rld! 你"
00000200: E5A5 BDEF BC8C E4B8 | "好，\xe4\xb8"
00000208: 96E7 958C EFBC 81   | "\x96界！"
`,

	`00000000: 48 65 6C 6C 6F 20 77 6F 72 6C 64 21 20 E4 BD | "Hello world! \xe4\xbd"
0000000F: A0 E5 A5 BD EF BC 8C E4 B8 96 E7 95 8C EF BC | "\xa0好，世界\xef\xbc"
0000001E: 81 48 65 6C 6C 6F 20 77 6F 72 6C 64 21 20 E4 | "\x81Hello world! \xe4"
0000002D: BD A0 E5 A5 BD EF BC 8C E4 B8 96 E7 95 8C EF | "\xbd\xa0好，世界\xef"
0000003C: BC 81 48 65 6C 6C 6F 20 77 6F 72 6C 64 21 20 | "\xbc\x81Hello world! "
0000004B: E4 BD A0 E5 A5 BD EF BC 8C E4 B8 96 E7 95 8C | "你好，世界"
0000005A: EF BC 81 48 65 6C 6C 6F 20 77 6F 72 6C 64 21 | "！Hello world!"
00000069: 20 E4 BD A0 E5 A5 BD EF BC 8C E4 B8 96 E7 95 | " 你好，世\xe7\x95"
00000078: 8C EF BC 81 48 65 6C 6C 6F 20 77 6F 72 6C 64 | "\x8c！Hello world"
00000087: 21 20 E4 BD A0 E5 A5 BD EF BC 8C E4 B8 96 E7 | "! 你好，世\xe7"
00000096: 95 8C EF BC 81 48 65 6C 6C 6F 20 77 6F 72 6C | "\x95\x8c！Hello worl"
000000A5: 64 21 20 E4 BD A0 E5 A5 BD EF BC 8C E4 B8 96 | "d! 你好，世"
000000B4: E7 95 8C EF BC 81 48 65 6C 6C 6F 20 77 6F 72 | "界！Hello wor"
000000C3: 6C 64 21 20 E4 BD A0 E5 A5 BD EF BC 8C E4 B8 | "ld! 你好，\xe4\xb8"
000000D2: 96 E7 95 8C EF BC 81 48 65 6C 6C 6F 20 77 6F | "\x96界！Hello wo"
000000E1: 72 6C 64 21 20 E4 BD A0 E5 A5 BD EF BC 8C E4 | "rld! 你好，\xe4"
000000F0: B8 96 E7 95 8C EF BC 81 48 65 6C 6C 6F 20 77 | "\xb8\x96界！Hello w"
000000FF: 6F 72 6C 64 21 20 E4 BD A0 E5 A5 BD EF BC 8C | "orld! 你好，"
0000010E: E4 B8 96 E7 95 8C EF BC 81 48 65 6C 6C 6F 20 | "世界！Hello "
0000011D: 77 6F 72 6C 64 21 20 E4 BD A0 E5 A5 BD EF BC | "world! 你好\xef\xbc"
0000012C: 8C E4 B8 96 E7 95 8C EF BC 81 48 65 6C 6C 6F | "\x8c世界！Hello"
0000013B: 20 77 6F 72 6C 64 21 20 E4 BD A0 E5 A5 BD EF | " world! 你好\xef"
0000014A: BC 8C E4 B8 96 E7 95 8C EF BC 81 48 65 6C 6C | "\xbc\x8c世界！Hell"
00000159: 6F 20 77 6F 72 6C 64 21 20 E4 BD A0 E5 A5 BD | "o world! 你好"
00000168: EF BC 8C E4 B8 96 E7 95 8C EF BC 81 48 65 6C | "，世界！Hel"
00000177: 6C 6F 20 77 6F 72 6C 64 21 20 E4 BD A0 E5 A5 | "lo world! 你\xe5\xa5"
00000186: BD EF BC 8C E4 B8 96 E7 95 8C EF BC 81 48 65 | "\xbd，世界！He"
00000195: 6C 6C 6F 20 77 6F 72 6C 64 21 20 E4 BD A0 E5 | "llo world! 你\xe5"
000001A4: A5 BD EF BC 8C E4 B8 96 E7 95 8C EF BC 81 48 | "\xa5\xbd，世界！H"
000001B3: 65 6C 6C 6F 20 77 6F 72 6C 64 21 20 E4 BD A0 | "ello world! 你"
000001C2: E5 A5 BD EF BC 8C E4 B8 96 E7 95 8C EF BC 81 | "好，世界！"
000001D1: 48 65 6C 6C 6F 20 77 6F 72 6C 64 21 20 E4 BD | "Hello world! \xe4\xbd"
000001E0: A0 E5 A5 BD EF BC 8C E4 B8 96 E7 95 8C EF BC | "\xa0好，世界\xef\xbc"
000001EF: 81 48 65 6C 6C 6F 20 77 6F 72 6C 64 21 20 E4 | "\x81Hello world! \xe4"
000001FE: BD A0 E5 A5 BD EF BC 8C E4 B8 96 E7 95 8C EF | "\xbd\xa0好，世界\xef"
0000020D: BC 81                                        | "\xbc\x81"
`,

	`00000000: 4865 6C6C 6F20 776F 726C 6421 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 | "Hello world! 你好，世界！Hello world! 你好，世界！He"
00000040: 6C6C 6F20 776F 726C 6421 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 6C6C | "llo world! 你好，世界！Hello world! 你好，世界！Hell"
00000080: 6F20 776F 726C 6421 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 6C6C 6F20 | "o world! 你好，世界！Hello world! 你好，世界！Hello "
000000C0: 776F 726C 6421 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 6C6C 6F20 776F | "world! 你好，世界！Hello world! 你好，世界！Hello wo"
00000100: 726C 6421 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 6C6C 6F20 776F 726C | "rld! 你好，世界！Hello world! 你好，世界！Hello worl"
00000140: 6421 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 6C6C 6F20 776F 726C 6421 | "d! 你好，世界！Hello world! 你好，世界！Hello world!"
00000180: 20E4 BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 6C6C 6F20 776F 726C 6421 20E4 | " 你好，世界！Hello world! 你好，世界！Hello world! \xe4"
000001C0: BDA0 E5A5 BDEF BC8C E4B8 96E7 958C EFBC 8148 656C 6C6F 2077 6F72 6C64 2120 E4BD A0E5 A5BD EFBC 8CE4 B896 E795 8CEF BC81 4865 6C6C 6F20 776F 726C 6421 20E4 BDA0 | "\xbd\xa0好，世界！Hello world! 你好，世界！Hello world! 你"
00000200: E5A5 BDEF BC8C E4B8 96E7 958C EFBC 81                                                                                                                           | "好，世界！"
`,
}

func TestExampleDumpConfig(t *testing.T) {
	b := bytes.NewBuffer(make([]byte, 0, 527))
	for b.Len() < 527 {
		b.WriteString("Hello world! 你好，世界！") // "Hello world! 你好，世界！" contains 31 bytes.
	}
	idx := 0
	for _, upper := range []bool{false, true} {
		for _, bytesPerLine := range []int{-1, 0, 8, 15, 64} {
			cfg := ExampleDumpConfig(upper, bytesPerLine)
			s := hex.DumpToString(b.Bytes(), cfg)
			if s != testExampleDumpOutputs[idx] {
				t.Errorf("upper: %t, bytesPerLine: %d\ngot:\n%s\nwanted:\n%s", upper, bytesPerLine, s, testExampleDumpOutputs[idx])
			}
			if bytesPerLine != -1 {
				idx++
			}
		}
	}
}
