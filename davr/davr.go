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

// readerToChan reads from r and assembles messages
func readerToChan(r io.Reader, out chan []byte) {
	buf := make([]byte, 1024)
	assy := new(bytes.Buffer)

	for {
		n, err := r.Read(buf)

		if err != nil {
			close(out)
			return
		}

		if n == 0 {
			continue
		}

		assy.Write(buf[:n])

		for { // assmble commands that are spread over mutliple reads
			ev, e := assy.ReadBytes(0xd)
			if e != nil || len(ev) == 0 {
				break
			}
			out <- ev
		}
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

			if len(cmd) == 0 {
				continue
			}

			// Allow commands that do not have a trailing \r
			if cmd[len(cmd)-1] != 0xd {
				cmd = append(cmd, 0xd)
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

	davr := &DAVR{c, make(chan []byte), make(chan []byte)}

	go readerToChan(c, davr.eventIn)
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
