package cellbuf

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// StyleFormatter is a writer that is aware of ANSI style and hyperlink escape
// sequences. It can be used to write styled text to a buffer, and then perform
// operations on that text, such as truncating it to fit within a certain
// width, or wrapping it to fit within a certain width, without breaking escape
// sequences.
type StyleFormatter struct {
	// Method is the ANSI method used to decode escape sequences and determine
	// the mono-width of characters.
	Method ansi.Method

	// Limit is the maximum number of characters per line. Zero means no limit.
	Limit int

	// Style is the last read style from the value string.
	Style Style

	// Link is the last read hyperlink from the value string.
	Link Link

	// PreserveSpace indicates whether spaces at the beginning of a line should
	// be preserved when wrapping.
	PreserveSpace bool

	// Forward the [io.Writer] to write to.
	Forward io.Writer

	// Breakpoints are the characters that are considered breakpoints for word
	// wrapping. A hyphen (-) is always considered a breakpoint.
	Breakpoints []rune
}

// Wrap returns a string that is wrapped to the specified limit applying any
// ANSI escape sequences in the string. It tries to wrap the string at word
// boundaries, but will break words if necessary.
//
// The breakpoints string is a list of characters that are considered
// breakpoints for word wrapping. A hyphen (-) is always considered a
// breakpoint.
//
// Note: breakpoints must be a string of 1-cell wide rune characters.
func Wrap(s string, limit int, breakpoints string) string {
	return StyleFormatter{Limit: limit, Breakpoints: []rune(breakpoints)}.Wrap(s)
}

// Wrap returns a string that is wrapped to the specified limit applying any
// ANSI escape sequences in the string. It tries to wrap the string at word
// boundaries, but will break words if necessary.
//
// The breakpoints string is a list of characters that are considered
// breakpoints for word wrapping. A hyphen (-) is always considered a
// breakpoint.
//
// Note: breakpoints must be a string of 1-cell wide rune characters.
func (s StyleFormatter) Wrap(b string) string {
	if len(b) == 0 {
		return ""
	}

	if s.Limit < 1 {
		return b
	}

	p := ansi.GetParser()
	defer ansi.PutParser(p)

	var (
		buf      bytes.Buffer
		word     bytes.Buffer
		space    bytes.Buffer
		curWidth int
		wordLen  int
	)

	addSpace := func() {
		curWidth += space.Len()
		buf.Write(space.Bytes())
		space.Reset()
	}

	addWord := func() {
		if word.Len() == 0 {
			return
		}

		addSpace()
		curWidth += wordLen
		buf.Write(word.Bytes())
		word.Reset()
		wordLen = 0
	}

	addNewline := func() {
		if !s.Link.Empty() {
			buf.WriteString(ansi.ResetHyperlink())
		}
		if !s.Style.Empty() {
			buf.WriteString(ansi.ResetStyle)
		}

		buf.WriteString("\n")
		curWidth = 0
		if !s.Style.Empty() {
			buf.WriteString(s.Style.Sequence())
		}
		if !s.Link.Empty() {
			buf.WriteString(ansi.SetHyperlink(s.Link.URL, s.Link.Params))
		}
		space.Reset()
	}

	addBreak := func(seq string, width int) {
		addSpace()
		if curWidth+wordLen+width >= s.Limit {
			// We can't fit the breakpoint in the current line, treat
			// it as part of the word.
			word.WriteString(seq)
			wordLen += width
		} else {
			addWord()
			buf.WriteString(seq)
			curWidth += width
		}
	}

	var state byte
	for len(b) > 0 {
		seq, width, n, newState := s.Method.DecodeSequenceInString(b, state, p)

		switch width {
		case 0:
			// Control codes and escape sequences
			switch {
			case ansi.HasCsiPrefix(seq) && p.Command() == 'm':
				// Select Graphic Rendition [ansi.SGR]
				ReadStyle(p.Params(), &s.Style)
			case ansi.HasOscPrefix(seq) && p.Command() == 8:
				// Hyperlinks OSC 8
				ReadLink(p.Data(), &s.Link)
			case len(seq) == 1 && seq[0] == '\n':
				if wordLen == 0 {
					if curWidth+space.Len() > s.Limit {
						curWidth = 0
					} else {
						// preserve whitespaces
						buf.Write(space.Bytes())
					}
					space.Reset()
				}

				addWord()
				addNewline()
			}

			word.WriteString(seq)
		default:
			switch {
			case len(strings.TrimSpace(seq)) == 0:
				addWord()
				if s.PreserveSpace || curWidth != 0 {
					// Preserve spaces at the beginning of a line
					space.WriteString(seq)
				}
			case len(seq) == 1 && seq[0] == '-':
				addBreak(seq, width)
			case utf8.RuneCountInString(seq) == 1:
				r, _ := utf8.DecodeRuneInString(seq)
				if runeContainsAny(r, s.Breakpoints) {
					addBreak(seq, width)
					break
				}

				fallthrough
			default:
				if curWidth > s.Limit {
					addNewline()
				}
				if wordLen+width >= s.Limit {
					// Hardwrap the word if it's too long
					addWord()
				}

				word.WriteString(seq)
				wordLen += width

				if curWidth+wordLen+space.Len() > s.Limit {
					addNewline()
				}
			}
		}

		state = newState
		b = b[n:]
	}

	if wordLen == 0 {
		if curWidth+space.Len() > s.Limit {
			curWidth = 0
		} else {
			// preserve whitespaces
			buf.Write(space.Bytes())
		}
		space.Reset()
	}

	addWord()

	return buf.String()
}

func runeContainsAny[T string | []rune](r rune, s T) bool {
	for _, c := range []rune(s) {
		if c == r {
			return true
		}
	}
	return false
}
