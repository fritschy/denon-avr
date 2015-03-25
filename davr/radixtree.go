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

import "fmt"

const (
	prefixCharsCount = 26 + 1 + 1 + 1 + 10 // A-Z, space, colon, question, 0-9
	numberIndexBase  = 26
	spaceIndex       = 36
	questionIndex    = 37
	colonIndex       = 38
)

// do not store actual chars
type radixNode struct {
	end  bool /// a string finishes here, too
	next [prefixCharsCount]*radixNode
}

const index2charMap = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 ?:"

func char2index(r byte) int {
	switch {
	case r == '?':
		return questionIndex
	case r == ':':
		return colonIndex
	case r == ' ':
		return spaceIndex
	case r >= 'A' && r <= 'Z':
		return int(r - 'A')
	case r >= '0' && r <= '9':
		return numberIndexBase + int(r-'0')
	default:
		return -1
	}
}

func index2char(i int) byte {
	switch {
	case i < len(index2charMap):
		return byte(index2charMap[i])
	default:
		return 0
	}
}

func (node *radixNode) insert(str string) {
	if len(str) == 0 {
		node.end = true
		return
	}
	r := byte(str[0])
	idx := char2index(r)
	if idx == -1 {
		fmt.Errorf("cannot insert rune '%v' into radixNode\n", r)
		return
	}
	if node.next[idx] == nil {
		node.next[idx] = &radixNode{}
	}
	node.next[idx].insert(str[1:])
}

// char 0x0 denotes root-node
func (node *radixNode) getWords(char byte, cur string, ret *[]string) {
	if char != 0 {
		cur += string(char)
	}

	if node.end {
		*ret = append(*ret, cur)
	}

	for i, n := range node.next {
		if n != nil {
			n.getWords(index2char(i), cur, ret)
		}
	}
}

// Char 0x0 denotes root-node
func (node *radixNode) query(char byte, str string, cur string, ret *[]string) {
	if len(str) == 0 {
		// empty Query, return all next words...
		node.getWords(char, cur, ret)
		return
	}

	r := byte(str[0])
	idx := char2index(r)

	if idx == -1 { // could not map character to index
		fmt.Errorf("cannot map character '%v' to index\n", r)
		return
	}

	if char != 0 {
		cur += string(char)
	}

	if node.end {
		*ret = append(*ret, cur)
	}

	if node.next[idx] != nil {
		node.next[idx].query(index2char(idx), str[1:], cur, ret)
	}
}
