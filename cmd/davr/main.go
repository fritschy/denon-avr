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

package main

import (
	"flag"
	"fmt"
	"github.com/bobappleyard/readline"
	"github.com/fritschy/denon-avr/davr"
	"os"
	"strings"
	"time"
)

func inputReader() <-chan []byte {
	in := make(chan []byte)

	go func() {
		for {
			s, e := readline.String("Denon> ")
			if e != nil {
				close(in)
				break
			}
			s = strings.TrimSpace(s)
			if len(s) == 0 {
				continue
			}
			in <- []byte(s)
			readline.AddHistory(s)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	return in
}

func main() {
	avr_host := flag.String("host", "avr", "AVR host[:port] to talk to (default: avr[:23])")
	flag.Parse()

	conn, err := davr.New(*avr_host)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	davr.ShowCommandHelp()

	readline.SetWordBreaks("")
	readline.Completer = davr.MakeReadlineCompleter()
	defer readline.Cleanup()
	input := inputReader()

	noRefresh := make(<-chan time.Time)
	refresh := noRefresh

	for {
		select {
		case ev, ok := <-conn.GetEventChan():
			if refresh == noRefresh {
				fmt.Println("")
			}
			if !ok {
				fmt.Errorf("Error reading from ev channel\n")
				return
			}
			cev := davr.CookEvent(ev)
			fmt.Printf("%s\n", &cev)
			refresh = time.After(200 * time.Millisecond)

		case cmd, ok := <-input:
			if !ok {
				if refresh == noRefresh {
					fmt.Println("")
				}
				conn.Close()
				return
			}
			conn.GetCommandChan() <- cmd
			refresh = time.After(200 * time.Millisecond)

		case _ = <-refresh:
			readline.RefreshLine()
			refresh = noRefresh
		}
	}
}
