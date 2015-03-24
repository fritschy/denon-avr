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

// Package davr implements channel-centric communication with a Denon AVR
package davr

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
)

type pipeStep func(input, output chan []byte)

func eventAssembler(rawIn, eventOut chan []byte) {
	scratch := make([]byte, 0)
	sep := []byte{0xd}

	for {
		input, ok := <-rawIn

		if !ok {
			close(eventOut)
			break
		}

		scratch = append(scratch, input...)

		for {
			evRest := bytes.SplitN(scratch, sep, 2)

			if len(evRest) != 2 {
				break
			}

			eventOut <- evRest[0]

			if len(evRest) == 2 {
				scratch = evRest[1]
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
func pipe(f pipeStep, in chan []byte) chan []byte {
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

// DAVR represents a Denon AVR connection
type DAVR struct {
	conn      net.Conn
	eventIn   chan []byte /// events from avr can be read from here
	commandIn chan []byte /// commands to the avr can be written here
}

func commandProxy(davr *DAVR) {
	for {
		select {
		case cmd, ok := <-davr.commandIn:
			if !ok {
				return
			}
			davr.conn.Write(cmd)
		}
	}
}

// New creates a Denon AVR connection, or returns an error value
// hostPort is of the form "host[:port]" if :port is omitted, the
// default port 23 is used.
func New(hostPort string) (*DAVR, error) {
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

	commandIn := make(chan []byte)

	// setup event pipe
	eventIn := pipe(eventAssembler,
		producer(func(out chan []byte) {
			readerToChan(c, out)
		}))

	davr := &DAVR{c, eventIn, commandIn}

	go commandProxy(davr)

	return davr, nil
}

// GetCommandChan returns the write only command-channel to the AVR
func (avr *DAVR) GetCommandChan() chan<- []byte {
	return avr.commandIn
}

// GetEventChan seturns the read-only event channel from the AVR
// Events returned through this channel are already assembled
// and guarantieed to be bounded and complete
// When reading from this channel, check the ok-flag in
// in order to determine if the connection is still alive.
func (avr *DAVR) GetEventChan() <-chan []byte {
	return avr.eventIn
}

// Close a Denon AVR connection
func (avr *DAVR) Close() {
	avr.conn.Close()
	close(avr.commandIn)
}
