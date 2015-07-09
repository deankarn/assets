package css

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1

const (
	itemError   itemType = iota // error occurred; value is text of error
	itemEOF                     // end of file
	itemText                    // plain text
	itemSpace                   // space, trimmed to a single space
	itemComment                 // css comment /* something */
)

const (
	startComment = "/*"
	endComment   = "*/"
)

// itemType identifies the type of lex items.
type itemType int

// // Pos represents a byte position in the original input text from which
// // this template was parsed.
// type Pos int

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos int      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name  string // the name of the input; used only for error reports
	input string // the string being scanned
	// leftDelim  string    // start of action
	// rightDelim string    // end of action
	state   stateFn   // the next lexing function to enter
	pos     int       // current position in the input
	start   int       // start position of this item
	width   int       // width of last rune read from input
	lastPos int       // position of most recent item returned by nextItem
	items   chan item // channel of scanned items
	// parenDepth int       // nesting depth of ( ) exprs
}

// newLexer instantiates a new css lexer instance
func newLexer(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {

	defer close(l.items)

	for l.state = lexText; l.state != nil; {
		l.state = l.state(l)
	}
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}

	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width

	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// emitPrev passes an item 1 position back, to the client.
func (l *lexer) emitPrev(t itemType) {
	l.backup()

	if l.pos > l.start {
		l.items <- item{t, l.start, l.input[l.start:l.pos]}
		l.start = l.pos
	}

}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

func lexText(l *lexer) stateFn {

LOOP:
	for {

		switch l.next() {
		case eof:
			break LOOP
		case '/':
			l.emitPrev(itemText)
			return lexComment
		}
	}

	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemText)
	}

	l.emit(itemEOF)
	return nil
}

func lexComment(l *lexer) stateFn {

	if !strings.HasPrefix(l.input[l.pos:], startComment) {
		return l.errorf("invalid comment")
	}

	i := strings.Index(l.input[l.pos:], endComment)

	if i < 0 {
		return l.errorf("unclosed comment")
	}

	l.pos += i + len(endComment)
	l.emit(itemComment)

	return lexText
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *lexer, nextLex stateFn) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}

	l.emit(itemSpace)
	return nextLex
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
