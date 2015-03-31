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
//
//
//
// Denon commands specification can be downloaed on Denon's website, the
// 'implemented' commands below are just a subset and descriptions/names
// may differ from the original specification.

package davr

import (
	"fmt"
	"strings"
)

type command [2]string

type commandCategory struct {
	prefix   string
	name     string
	suffixes []command
}

// DenonCommands holds Denon AVR commands and possibly descriptions
type DenonCommands struct {
	commands []commandCategory
	tree     *radixNode
	count    uint
}

func makeCommands(args ...string) []command {
	c := make([]command, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		c = append(c, command{args[i], args[i+1]})
	}
	return c
}

// Oh this sucks so hard, I don't even...
func initDenonCommands() *DenonCommands {
	qOnOff := makeCommands("?", "", "ON", "", "OFF", "")
	denonCommands := &DenonCommands{
		[]commandCategory{
			commandCategory{"PW", "Main Power", makeCommands("?", "", "ON", "", "STANDBY", "")},
			commandCategory{"ZM", "Main Zone", append(qOnOff, makeCommands("FAVORITE1", "", "FAVORITE2", "", "FAVORITE3", "", "FAVORITE4", "")...)},
			// CommandCategory{"Z2", "Zone 2", qOnOff},
			// CommandCategory{"Z3", "Zone 3", qOnOff},
			commandCategory{"MV", "Master Volume", makeCommands("?", "", "UP", "", "DOWN", "")},
			commandCategory{"MU", "Mute", qOnOff},
			commandCategory{"SI", "Select Input", makeCommands("?", "", "BD", "", "DVD", "", "TV", "", "MPLAY", "", "NET", "", "GAME", "")},
			commandCategory{"SV", "Select Video", makeCommands("?", "", "BD", "", "DVD", "", "TV", "", "MPLAY", "")}, // not more?!
			commandCategory{"MS", "Mode Select", makeCommands("?", "", "STEREO", "", "MOVIE", "", "MUSIC", "", "DIRECT", "", "PURE DIRECT", "", "QUICK ?", "",
				"QUICK1", "", "QUICK1 MEMORY", "",
				"QUICK2", "", "QUICK2 MEMORY", "",
				"QUICK3", "", "QUICK3 MEMORY", "",
				"QUICK4", "", "QUICK4 MEMORY", "",
				"QUICK5", "", "QUICK5 MEMORY", "")},
			commandCategory{"PSMULTEQ:", "MultEQ Settings", makeCommands(" ?", "", "AUDYSSEY", "", "FLAT", "", "MANUAL", "", "OFF", "")},
			commandCategory{"PSDYNEQ ", "Dynamic EQ Settings", qOnOff},
			commandCategory{"PSREFLEV ", "DynEQ Reference Level", makeCommands("?", "", "0", "", "5", "", "10", "", "15", "")},
			commandCategory{"PSDYNVOL ", "Dynamic Volume Settings", makeCommands("?", "", "LIT", "", "MED", "", "HEV", "", "OFF", "")},
			commandCategory{"PSLFC ", "Low Frequency Containment", qOnOff},
			commandCategory{"PSCNTAMT ", "LFC Amount", makeCommands("UP", "", "DOWN", "", "?", "")},
			commandCategory{"NS", "Player Control", makeCommands("E", "Report Display Text (UTF-8)", "A", "Report Display Text (ASCII)", "RND", "Toggle Random", "RPT", "Toggle Repeat", "D", "Direct Text Search ('NSDx, where x = 0-9,A-Z)")},
			commandCategory{"MN", "Menu Control", makeCommands("MEN?", "", "CUP", "up", "CDN", "down", "CRT", "right", "CLT", "left", "ENT", "enter", "OPT", "option", "INF", "info", "RTN", "return")},
			commandCategory{"CV", "Channel Volume Control", makeCommands("?", "", "FL", "Front Left (+-0 at 50)", "FR", "Front Right (+-0 at 50)", "SW", "Subwoofer 1 (+-0 at 50)", "SW2", "Subwoofer 2 (+-0 at 50)")},
		},
		nil,
		0,
	}

	denonCommands.tree = &radixNode{}
	for _, cc := range denonCommands.commands {
		for _, c := range cc.suffixes {
			denonCommands.tree.insert(strings.Join([]string{cc.prefix, c[0]}, ""))
			denonCommands.count++
		}
	}

	return denonCommands
}

var commands *DenonCommands

// MakeReadlineCompleter returns a readline completer that can be used with
// github.com/bobappleyard/readline
func MakeReadlineCompleter() func(query, ctx string) []string {
	words := make([]string, 0, commands.count)
	return func(query, ctx string) []string {
		words = words[0:0] // clear
		commands.tree.query(0, strings.ToUpper(query), "", &words)
		return words
	}
}

// ShowCommandHelp is mostly useless, prints some commands and descriptions thereof on stdout
func ShowCommandHelp() {
	fmt.Println("I know the following commands:\n")

	for _, cc := range commands.commands {
		fmt.Printf("%-20s%s\n", cc.prefix, cc.name)
		for _, c := range cc.suffixes {
			w := strings.Join([]string{cc.prefix, c[0]}, "")
			fmt.Printf("  %-16s", w)
			if len(c[1]) != 0 {
				fmt.Print("  ", c[1])
			} else if c[0][len(c[0])-1] == '?' {
				fmt.Print("  Query status")
			}
			fmt.Println("")
		}
		fmt.Println("")
	}

	fmt.Println("")
}

func init() {
	commands = initDenonCommands()
}
