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

// Implement channel-centric communication with a Denon AVR
package davr

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
)

type pipe_step func(input, output chan []byte)

func event_assembler(raw_in, event_out chan []byte) {
	scratch := make([]byte, 0)
	sep := []byte{0xd}

	for {
		input, ok := <-raw_in

		if !ok {
			close(event_out)
			break
		}

		scratch = append(scratch, input...)

		for {
			ev_rest := bytes.SplitN(scratch, sep, 2)

			if len(ev_rest) != 2 {
				break
			}

			event_out <- ev_rest[0]

			if len(ev_rest) == 2 {
				scratch = ev_rest[1]
			} else {
				// no remaining bytes, clear scratch
				scratch = scratch[0:0]
			}
		}
	}
}

func producer(f func(out chan []byte)) chan []byte {
	out := make(chan []byte)
	go f(out)
	return out
}

// create a pipe-segment using func f
func pipe(f pipe_step, in chan []byte) chan []byte {
	out := make(chan []byte)
	go f(in, out)
	return out
}

func readerToChan(r io.Reader, out chan []byte) {
	var n int
	var err error
	var buf []byte

	for {
		if buf == nil {
			buf = make([]byte, 1024)
		}

		n, err = r.Read(buf)
		if err != nil {
			close(out)
			return
		}

		if n == 0 {
			continue
		}

		out <- buf[:n]
		buf = nil
	}
}

// Represents a Denon AVR connection
type DAVR struct {
	conn       net.Conn
	event_in   chan []byte /// events from avr can be read from here
	command_in chan []byte /// commands to the avr can be written here
}

func run(davr *DAVR) {
	for {
		select {
		case cmd, ok := <-davr.command_in:
			if !ok {
				return
			}
			davr.conn.Write(cmd)
		}
	}
}

// Connect to an AVR
func Connect(hostPort string) (*DAVR, error) {
	var conn string

	if strings.Contains(hostPort, ":") {
		conn = hostPort
	} else {
		conn = fmt.Sprintf("%s:23", hostPort)
	}

	c, e := net.Dial("tcp", conn)
	if e != nil {
		return nil, e
	}

	command_in := make(chan []byte)

	// setup event pipe
	event_in := pipe(event_assembler,
		producer(func(out chan []byte) {
			readerToChan(c, out)
		}))

	davr := &DAVR{c, event_in, command_in}

	go run(davr)

	return davr, nil
}

// Return the write only command-channel to the AVR
func (self *DAVR) GetCommandChannel() chan<- []byte {
	return self.command_in
}

// Return the read-only event channel from the AVR
// Events returned through this channel are already assembled
// and guarantieed to be bounded and complete
// When reading from this channel, check the ok-flag in
// in order to determine if the connection is still alive.
func (self *DAVR) GetEventChannel() <-chan []byte {
	return self.event_in
}

// Close a Denon AVR connection
func (self *DAVR) Close() {
	self.conn.Close()
	close(self.command_in)
}
