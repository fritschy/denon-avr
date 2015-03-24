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

type DenonCommands struct {
	commands []commandCategory
	tree     *radixNode
	count    uint
}

func make_commands(args ...string) []command {
	c := make([]command, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		c = append(c, command{args[i], args[i+1]})
	}
	return c
}

// Oh this sucks so hard, I don't even...
func init_denon_commands() *DenonCommands {
	q_on_off := make_commands("?", "", "ON", "", "OFF", "")
	denon_commands := &DenonCommands{
		[]commandCategory{
			commandCategory{"PW", "Main Power", make_commands("?", "", "ON", "", "STANDBY", "")},
			commandCategory{"ZM", "Main Zone", append(q_on_off, make_commands("FAVORITE1", "", "FAVORITE2", "", "FAVORITE3", "", "FAVORITE4", "")...)},
			// CommandCategory{"Z2", "Zone 2", q_on_off},
			// CommandCategory{"Z3", "Zone 3", q_on_off},
			commandCategory{"MV", "Master Volume", make_commands("?", "", "UP", "", "DOWN", "")},
			commandCategory{"MU", "Mute", q_on_off},
			commandCategory{"SI", "Select Input", make_commands("?", "", "BD", "", "DVD", "", "TV", "", "MPLAY", "", "NET", "", "GAME", "")},
			commandCategory{"SV", "Select Video", make_commands("?", "", "BD", "", "DVD", "", "TV", "", "MPLAY", "")}, // not more?!
			commandCategory{"MS", "Mode Select", make_commands("?", "", "STEREO", "", "MOVIE", "", "MUSIC", "", "DIRECT", "", "PURE DIRECT", "", "QUICK ?", "",
				"QUICK1", "", "QUICK1 MEMORY", "",
				"QUICK2", "", "QUICK2 MEMORY", "",
				"QUICK3", "", "QUICK3 MEMORY", "",
				"QUICK4", "", "QUICK4 MEMORY", "",
				"QUICK5", "", "QUICK5 MEMORY", "")},
			commandCategory{"PSMULTEQ:", "MultEQ Settings", make_commands(" ?", "", "AUDYSSEY", "", "FLAT", "", "MANUAL", "", "OFF", "")},
			commandCategory{"PSDYNEQ ", "Dynamic EQ Settings", q_on_off},
			commandCategory{"PSREFLEV ", "DynEQ Reference Level", make_commands("?", "", "0", "", "5", "", "10", "", "15", "")},
			commandCategory{"PSDYNVOL ", "Dynamic Volume Settings", make_commands("?", "", "LIT", "", "MED", "", "HEV", "", "OFF", "")},
			commandCategory{"PSLFC ", "Low Frequency Containment", q_on_off},
			commandCategory{"PSCNTAMT ", "LFC Amount", make_commands("UP", "", "DOWN", "", "?", "")},
			commandCategory{"NS", "Player Control", make_commands("E", "Report Display Text (UTF-8)", "A", "Report Display Text (ASCII)", "RND", "Toggle Random", "RPT", "Toggle Repeat", "D", "Direct Text Search ('NSDx, where x = 0-9,A-Z)")},
			commandCategory{"MN", "Menu Control", make_commands("MEN?", "", "CUP", "up", "CDN", "down", "CRT", "right", "CLT", "left", "ENT", "enter", "OPT", "option", "INF", "info", "RTN", "return")},
		},
		nil,
		0,
	}

	denon_commands.tree = &radixNode{}
	for _, cc := range denon_commands.commands {
		for _, c := range cc.suffixes {
			denon_commands.tree.insert(strings.Join([]string{cc.prefix, c[0]}, ""))
			denon_commands.count++
		}
	}

	return denon_commands
}

var commands *DenonCommands

// Return a readline completer, can be used for github.com/bobappleyard/readline
func MakeReadlineCompleter() func(query, ctx string) []string {
	words := make([]string, 0, commands.count)
	return func(query, ctx string) []string {
		words = words[0:0] // clear
		commands.tree.query(0, strings.ToUpper(query), "", &words)
		return words
	}
}

// Mostly useless, prints some commands and descriptions thereof on stdout
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
	commands = init_denon_commands()
}
