// Copyright 2025 Christophe Eymard
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package relql

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"gitlab.com/tozd/go/errors"
)

/**
Lexer implementation for the relql language.
*/

type TokenType int

const (
	T_ILLEGAL TokenType = iota
	T_EOF

	T_IDENT
	T_STRING
	T_NUMBER
	T_LPAREN
	T_RPAREN
	T_LBRACKET
	T_RBRACKET
	T_LBRACE
	T_RBRACE
	T_COMMA
	T_COLON
	T_SEMICOLON

	// operators
	T_OPERATOR

	T_INVALID
)

var token_names = map[TokenType]string{
	T_IDENT:     "Ident",
	T_STRING:    "String",
	T_NUMBER:    "Number",
	T_LPAREN:    "LeftParen",
	T_RPAREN:    "RightParen",
	T_LBRACKET:  "LeftBracket",
	T_RBRACKET:  "RightBracket",
	T_LBRACE:    "LeftBrace",
	T_RBRACE:    "RightBrace",
	T_COMMA:     "Comma",
	T_COLON:     "Colon",
	T_OPERATOR:  "Operator",
	T_INVALID:   "Invalid",
	T_SEMICOLON: "Semicolon",
}

// TokenDef represents an operator token definition
type TokenDef struct {
	Kind TokenType
	Name string
	Lbp  int
	Rbp  int
}

// T is a helper function to create TokenDef
func T(tokenType TokenType, name string, lbp, rbp int) TokenDef {
	token_names[tokenType] = name

	return TokenDef{
		Kind: tokenType,
		Name: name,
		Lbp:  lbp,
		Rbp:  rbp,
	}
}

type Token struct {
	Kind TokenType

	// The position of the token in the lexer's buffer
	Pos int

	//
	Bytes []byte

	// The next token in the stream
	next *Token
}

func (t *Token) IsUnexpected() bool {
	return t.Kind == T_ILLEGAL || t.Kind == T_EOF
}

func (t *Token) IsEOF() bool {
	return t.Bytes == nil || t.Kind == T_EOF
}

func (t *Token) IsIllegal() bool {
	return t.Kind == T_ILLEGAL
}

func (t *Token) Name() string {
	if name, ok := token_names[t.Kind]; ok {
		return name
	}
	return fmt.Sprintf("T_%d", t.Kind)
}

func (t *Token) String() string {
	return string(t.Bytes)
}

func (t *Token) Error() error {
	return fmt.Errorf("unexpected token: %s (%s)", t.String(), t.Name())
}

func (t *Token) ErrorMessage(message string) error {
	if t == nil {
		return errors.Errorf("at position %d: %s", -1, message)
	}
	return errors.Errorf("at position %d '%s' (%s): %s", t.Pos, t.String(), t.Name(), message)
}

func NewLexer(buf []byte) *Lexer {
	return &Lexer{
		buf:  buf,
		last: nil,
	}
}

type Lexer struct {
	buf  []byte
	last *Token
}

func (l *Lexer) Peek() *Token {
	var tk = nextToken(l.buf, l.last)
	if l.last != nil {
		l.last.next = tk
	}
	return tk
}

func (l *Lexer) PeekString(s string) *Token {
	var tk = l.Peek()
	if tk.String() != s {
		return nil
	}
	return tk
}

func (l *Lexer) PeekByte(b byte) *Token {
	var tk = l.Peek()
	if len(tk.Bytes) != 1 || tk.Bytes[0] != b {
		return nil
	}
	return tk
}

func (l *Lexer) PeekKind(kind TokenType) *Token {
	var tk = l.Peek()
	if tk.Kind != kind {
		return nil
	}
	return tk
}

// Next returns the next token in the buffer
func (l *Lexer) Next() *Token {
	var tk = nextToken(l.buf, l.last)
	l.last = tk
	return tk
}

// Consume returns the next token if it matches the provided kind or doesn't advance the lexer and returns nil if it doesn't
func (l *Lexer) Consume(kind TokenType) *Token {
	var tok = l.Peek()
	if tok.Kind != kind {
		return nil
	}
	l.SetPosition(tok)
	return tok
}

func (l *Lexer) ConsumeByte(b byte) *Token {
	var tok = l.Peek()
	if len(tok.Bytes) != 1 || tok.Bytes[0] != b {
		return nil
	}
	l.SetPosition(tok)
	return tok
}

func (l *Lexer) ConsumeStringIgnoreCase(s ...string) *Token {
	var tok = l.Peek()
	for _, s := range s {
		if strings.EqualFold(tok.String(), s) {
			l.SetPosition(tok)
			return tok
		}
	}
	return nil
}

func (l *Lexer) ConsumeString(s ...string) *Token {
	var tok = l.Peek()
	for _, s := range s {
		if tok.String() == s {
			l.SetPosition(tok)
			return tok
		}
	}
	return nil
}

// Rewind makes the lexer go back to the provided token
func (l *Lexer) SetPosition(tk *Token) {
	l.last = tk
}

var whitespace = [256]bool{
	' ':  true,
	'\t': true,
	'\n': true,
	'\r': true,
	'~':  true,
}

var single_char_tokens = map[byte]TokenType{
	'(': T_LPAREN,
	')': T_RPAREN,
	'[': T_LBRACKET,
	']': T_RBRACKET,
	'{': T_LBRACE,
	'}': T_RBRACE,
	',': T_COMMA,
	'.': T_OPERATOR,
	// ':': T_COLON,
	';': T_SEMICOLON,
}

// skipWhitespace skips whitespace in the buffer, returning the index of the first non-whitespace character.
func skipWhitespace(buf []byte, start int) int {
	for i := start; i < len(buf); i++ {
		if !whitespace[buf[i]] {
			return i
		}
	}
	return len(buf)
}

func nextToken(buf []byte, last *Token) *Token {
	if last != nil && last.IsIllegal() {
		return last
	}

	if last != nil && last.next != nil {
		return last.next
	}

	var l = len(buf)

	var last_pos = 0
	if last != nil {
		last_pos = last.Pos + len(last.Bytes)
	}

	if last_pos >= l {
		if last != nil && last.IsEOF() {
			return last
		}

		return &Token{
			Kind:  T_EOF,
			Pos:   last_pos,
			Bytes: nil,
		}
	}

	var start_pos = skipWhitespace(buf, last_pos)

	if start_pos >= l {
		return &Token{
			Kind:  T_EOF,
			Pos:   start_pos,
			Bytes: nil,
		}
	}

	var cur = buf[start_pos]

	if kind, ok := single_char_tokens[cur]; ok {
		return &Token{
			Kind:  kind,
			Pos:   start_pos,
			Bytes: buf[start_pos : start_pos+1],
		}
	}

	if cur == '"' || cur == '\'' {
		var end_pos = scanString(l, buf, cur, start_pos+1)

		kind := T_STRING
		if cur == '"' {
			// In postgres, double quoted strings are identifiers
			kind = T_IDENT
		}

		return &Token{
			Kind:  kind,
			Pos:   start_pos,
			Bytes: buf[start_pos:end_pos],
		}
	}

	if num := scanOperator(l, buf, start_pos); num != start_pos {
		return &Token{
			Kind:  T_OPERATOR,
			Pos:   start_pos,
			Bytes: buf[start_pos:num],
		}
	}

	if num := scanNumber(l, buf, start_pos); num != start_pos {
		return &Token{
			Kind:  T_NUMBER,
			Pos:   start_pos,
			Bytes: buf[start_pos:num],
		}
	}

	if ident := scanIdentifier(l, buf, start_pos); ident != start_pos {
		str := strings.ToLower(string(buf[start_pos:ident]))
		kind := T_IDENT

		// special case for operators
		switch str {
		case
			"not",
			"and",
			"or",
			"is",
			"isnull",
			"notnull",
			"between",
			"in",
			"like",
			"ilike",
			"similar",
			"similarto":
			kind = T_OPERATOR
		}

		return &Token{
			Kind:  kind,
			Pos:   start_pos,
			Bytes: buf[start_pos:ident],
		}
	}

	return &Token{
		Kind:  T_ILLEGAL,
		Pos:   start_pos,
		Bytes: buf[start_pos : start_pos+1],
	}
}

// scanNumber scans a number in the buffer, returning the index of the first non-number character.
func scanNumber(l int, buf []byte, pos int) int {
	for i := pos; i < l; i++ {
		var c = buf[i]
		if c < '0' || c > '9' || c == '.' && i == pos {
			return i
		}
	}
	return l
}

/*
+ - * / < > = ~ ! @ # % ^ & | ` ?
*/
func scanOperator(len int, buf []byte, pos int) int {
	cnt := 0
	start := pos
	has_allowed_wonky := false

	if buf[pos] == '.' {
		return pos + 1
	}

	if buf[pos] == ':' {
		if pos+1 < len && buf[pos+1] == ':' {
			return pos + 2
		}
		return pos + 1
	}

	for {
		var c = buf[pos]

		// special cases to avoid comments
		if c == '-' && pos+1 < len && buf[pos+1] == '-' {
			return pos
		}

		if c == '/' && pos+1 < len && buf[pos+1] == '*' {
			return pos
		}

		if operator_allowed_wonky_end[c] {
			has_allowed_wonky = true
		}

		if operator_char[c] {
			pos++
		} else {
			// Postgres has this funky rule where an operator cannot end by '-' or '+' unless it has one of the wonky operator characters somewhere, so that « @- is a valid operator but *- is not. »
			if has_allowed_wonky {
				return pos
			}
			for {
				if pos-1 <= start {
					break
				}
				last := buf[pos-1]
				if (last == '-' || last == '+') && pos-1 > start {
					pos--
				} else {
					break
				}
			}
			return pos
		}

		cnt++
		if cnt >= 63 {
			return pos
		}
	}
}

var operator_allowed_wonky_end = map[byte]bool{
	'~': true,
	'!': true,
	'@': true,
	'#': true,
	'%': true,
	'^': true,
	'&': true,
	'|': true,
}

// + - * / < > = ~ ! @ # % ^ & | ` ?
var operator_char = [256]bool{
	'+': true,
	'-': true,
	'*': true,
	'/': true,
	'<': true,
	'>': true,
	'=': true,
	'~': true,
	'!': true,
	'@': true,
	'#': true,
	'%': true,
	'^': true,
	'&': true,
	'|': true,
	'`': true,
	'?': true,
}

func scanIdentifier(l int, buf []byte, pos int) int {
	var quoted = false

	if pos < l && buf[pos] == '"' {
		quoted = true
		pos++
	}

	var size = 1
	for i := pos; i < l; i += size {
		var c = buf[i]
		if quoted {
			if c == '"' {
				return i
			}
		} else {
			rune, size := utf8.DecodeRune(buf[i:])
			if !unicode.IsLetter(rune) && !unicode.IsDigit(rune) && c != '_' && c != '$' {
				return i + size - 1
			}
		}
	}
	return l
}

func scanString(l int, buf []byte, start byte, pos int) int {
	for i := pos; i < l; i++ {
		var c = buf[i]
		if c == start {
			// double quoting is an escape sequence and is understood as such in pg
			if i+1 < l && buf[i+1] == start {
				i++
				continue
			}
			return i + 1
		}
	}
	return l
}
