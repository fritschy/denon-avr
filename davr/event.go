// Copyright (c) 2015 Marcus Fritzsch
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package davr

import "bytes"

const (
	DavrEventNSE = iota
	// DavrEventNSA
	DavrEventGeneric
)

// We really have just 2 event types for now
// - NSE (or NSA) events, which are made up of parts
// - Generic events which really are just a buffer of text
type DavrEvent struct {
	data   []byte
	fields [][]byte // actually, slices (parts) of data
	event  uint8
}

func (self *DavrEvent) String() string {
	if self.event == DavrEventGeneric {
		return string(self.data)
	}

	if len(self.data) > 4 {
		// This message is made up of 101 bytes:
		//
		// NSE[0-8]TEXT<CR>
		//
		// Where the first byte of TEXT is a special bitfield for NSE[1-6] and
		// should be handled as such. (it is a normal char for NSE0, NSE7 and NSE8)
		//
		// The bits have the following meaning (Cursor Position):
		//   0: Is Playable Music?
		//   1: Is Directory?
		//   3: Is Selected by Cursor?
		//   6: Is (Has?) a Picture?
		// All non-specified bits should be ignored, that is: 2,4,5 and 7
		//
		// In general TEXT is 96 bytes long, all after and including a NUL byte
		// should be ignored.
		// The event is terminated with a <CR> (0x0d) byte.
		//
		// Additionally, NSE0 seems to be a general satus or title.

		ev := append([]byte{}, self.data...)
		n := int(ev[3]) - int('0')

		var flags uint8

		if n >= 1 && n <= 6 {
			flags = uint8(ev[4])
			// copy(ev[4:], ev[5:])
			copy(ev[1:], ev[:4])
			ev = ev[1:]
		}

		if flags&0x2b != 0 { // 0b101011
			// I feel I should be doing this a little more high-level...
			ev = append(ev, byte(' '))
			ev = append(ev, byte('['))

			if flags&(1<<0) != 0 {
				ev = append(ev, byte('F'))
			} else if flags&(1<<1) != 0 {
				ev = append(ev, byte('D'))
			}
			if flags&(1<<6) != 0 {
				ev = append(ev, byte('P'))
			}
			if flags&(1<<3) != 0 { // want cursor last
				ev = append(ev, byte('C'))
			}

			ev = append(ev, byte(']'))
		}

		return string(ev)
	}

	// ARGH!
	return string(self.data)
}

func CookEvent(ev []byte) DavrEvent {
	if len(ev) > 4 && ev[0] == 'N' && ev[1] == 'S' && (ev[2] == 'E' || ev[2] == 'A') {
		// disregard everything after the first NUL byte
		nulidx := bytes.IndexByte(ev, 0) // be safe...?
		if nulidx != -1 {
			ev = ev[:nulidx]
		}

		return DavrEvent{
			ev, [][]byte{}, DavrEventNSE,
		}
	}

	return DavrEvent{
		ev, [][]byte{},
		DavrEventGeneric,
	}
}
