package parsing

import (
	"fmt"
	"strings"
)

type cursor struct {
	pointer uint
	loc     location
}

type lexer func(string, cursor) (*token, cursor, bool)

func lex(source string) ([]*token, error) {
	tokens := []*token{}
	cur := cursor{}

lex:
	for cur.pointer < uint(len(source)) {
		lexers := []lexer{lexKeyword, lexSymbol, lexString, lexNumeric, lexIdentifier, lexOperation}
		for _, l := range lexers {
			if token, newCursor, ok := l(source, cur); ok {
				cur = newCursor

				// Omit nil tokens for valid, but empty syntax like newlines
				if token != nil {
					tokens = append(tokens, token)
				}

				continue lex
			}
		}

		hint := ""
		if len(tokens) > 0 {
			hint = " after " + tokens[len(tokens)-1].value
		}
		return nil, fmt.Errorf("unable to lex token%s, at %d:%d", hint, cur.loc.line, cur.loc.col)
	}

	return tokens, nil
}

func lexOperation(source string, ic cursor) (*token, cursor, bool) {
	c := source[ic.pointer]
	cur := ic
	// Will get overwritten later if not an ignored syntax
	cur.pointer++
	cur.loc.col++

	switch c {
	// Syntax that should be thrown away
	case '\n':
		cur.loc.line++
		cur.loc.col = 0

		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cur, true
	}

	operations := []operation{
		equalsOperation,
		notEqualOperation,
		moreOperation,
		lessOperation,
	}

	options := make([]string, len(operations))
	for _, o := range operations {
		options = append(options, string(o))
	}

	// Use `ic`, not `cur`
	match := longestMatch(source, ic, options)
	// Unknown character
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	return &token{
		value: match,
		loc:   ic.loc,
		kind:  operationKind,
	}, cur, true
}

func lexNumeric(source string, ic cursor) (*token, cursor, bool) {
	cur := ic

	periodFound := false
	expMarkerFound := false

	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		cur.loc.col++

		isDigit := c >= '0' && c <= '9'
		isPeriod := c == '.'
		isExpMarker := c == 'e'

		// Must start with a digit or period
		if cur.pointer == ic.pointer {
			if !isDigit && !isPeriod {
				return nil, ic, false
			}

			periodFound = isPeriod

			continue
		}

		if isPeriod {
			if periodFound {
				return nil, ic, false
			}

			periodFound = true

			continue
		}

		if isExpMarker {
			if expMarkerFound {
				return nil, ic, false
			}

			// No periods allowed after expMarker
			periodFound = true
			expMarkerFound = true

			// expMarker must be followed by digits
			if cur.pointer == uint(len(source)-1) {
				return nil, ic, false
			}

			cNext := source[cur.pointer+1]
			if cNext == '-' || cNext == '+' {
				cur.pointer++
				cur.loc.col++
			}

			continue
		}

		if !isDigit {
			break
		}
	}

	// No characters accumulated
	if cur.pointer == ic.pointer {
		return nil, ic, false
	}

	return &token{
		value: source[ic.pointer:cur.pointer],
		loc:   ic.loc,
		kind:  numericKind,
	}, cur, true
}

func lexCharacterDelimited(source string, ic cursor, delimiter byte) (*token, cursor, bool) {
	cur := ic

	if len(source[cur.pointer:]) == 0 {
		return nil, ic, false
	}

	if source[cur.pointer] != delimiter {
		return nil, ic, false
	}

	cur.loc.col++
	cur.pointer++

	var value []byte
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]

		if c == delimiter {
			fmt.Println(cur.pointer+1, uint(len(source)), source[cur.pointer+1])
			// SQL escapes are via double characters, not backslash.
			if cur.pointer+1 >= uint(len(source)) || source[cur.pointer+1] != delimiter {
				cur.pointer++

				return &token{
					value: string(value),
					loc:   ic.loc,
					kind:  stringKind,
				}, cur, true
			}
			value = append(value, delimiter)

			cur.pointer++
			cur.loc.col++
		}

		value = append(value, c)
		cur.loc.col++
	}

	return nil, ic, false
}

func lexString(source string, ic cursor) (*token, cursor, bool) {
	return lexCharacterDelimited(source, ic, '\'')
}

func lexSymbol(source string, ic cursor) (*token, cursor, bool) {
	c := source[ic.pointer]
	cur := ic
	// Will get overwritten later if not an ignored syntax
	cur.pointer++
	cur.loc.col++

	switch c {
	// Syntax that should be thrown away
	case '\n':
		cur.loc.line++
		cur.loc.col = 0
		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cur, true
	}

	// Syntax that should be kept
	symbols := []symbol{
		commaSymbol,
		semicolonSymbol,
	}

	options := make([]string, len(symbols))
	for _, s := range symbols {
		options = append(options, string(s))
	}

	// Use `ic`, not `cur`
	match := longestMatch(source, ic, options)
	// Unknown character
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	return &token{
		value: match,
		loc:   ic.loc,
		kind:  symbolKind,
	}, cur, true
}

func lexKeyword(source string, ic cursor) (*token, cursor, bool) {
	cur := ic
	keywords := []keyword{
		selectKeyword,
		whereKeyword,
		fromKeyword,
		andKeyword,
		orKeyword,
	}

	options := make([]string, len(keywords))
	for _, k := range keywords {
		options = append(options, string(k))
	}

	match := longestMatch(source, ic, options)
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	return &token{
		value: match,
		kind:  keywordKind,
		loc:   ic.loc,
	}, cur, true
}

func longestMatch(source string, ic cursor, options []string) string {
	var value []byte
	var skipList []int
	var match string

	cur := ic

	for cur.pointer < uint(len(source)) {
		value = append(value, strings.ToLower(string(source[cur.pointer]))...)
		cur.pointer++

	match:
		for i, option := range options {
			for _, skip := range skipList {
				if i == skip {
					continue match
				}
			}

			// Deal with cases like INT vs INTO
			if option == string(value) {
				skipList = append(skipList, i)
				if len(option) > len(match) {
					match = option
				}

				continue
			}

			sharesPrefix := string(value) == option[:cur.pointer-ic.pointer]
			tooLong := len(value) > len(option)
			if tooLong || !sharesPrefix {
				skipList = append(skipList, i)
			}
		}

		if len(skipList) == len(options) {
			break
		}
	}

	return match
}

func lexIdentifier(source string, ic cursor) (*token, cursor, bool) {
	// Handle separately if is a double-quoted identifier
	if token, newCursor, ok := lexCharacterDelimited(source, ic, '"'); ok {
		return token, newCursor, true
	}

	cur := ic

	c := source[cur.pointer]
	// Other characters count too, big ignoring non-ascii for now
	isAlphabetical := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	if !isAlphabetical {
		return nil, ic, false
	}
	cur.pointer++
	cur.loc.col++

	value := []byte{c}
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c = source[cur.pointer]

		// Other characters count too, big ignoring non-ascii for now
		isAlphabetical := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		isNumeric := c >= '0' && c <= '9'
		if isAlphabetical || isNumeric || c == '$' || c == '_' {
			value = append(value, c)
			cur.loc.col++
			continue
		}

		break
	}

	if len(value) == 0 {
		return nil, ic, false
	}

	return &token{
		// Unquoted dentifiers are case-insensitive
		value: strings.ToLower(string(value)),
		loc:   ic.loc,
		kind:  identifierKind,
	}, cur, true
}
