package parsing

import (
	"errors"
	"fmt"
)

type location struct {
	line uint
	col  uint
}

type token struct {
	value string
	kind  tokenKind
	loc   location
}

type SelectStatement struct {
	item  []*expression
	from  token
	where []interface{}
}

type expression struct {
	literal *token
}

//for where
type conditions struct {
	literal   token
	operation token
	value     token
}

//for where
type predicate struct {
	predicate token
}

func (t *token) equals(other *token) bool {
	return t.value == other.value && t.kind == other.kind
}

func expectToken(tokens []*token, cursor uint, t token) bool {
	if cursor >= uint(len(tokens)) {
		return false
	}

	return t.equals(tokens[cursor])
}

func tokenFromKeyword(k keyword) token {
	return token{
		kind:  keywordKind,
		value: string(k),
	}
}

func helpMessage(tokens []*token, cursor uint, msg string) {
	var c *token
	if cursor < uint(len(tokens)) {
		c = tokens[cursor]
	} else {
		c = tokens[cursor-1]
	}

	fmt.Printf("[%d,%d]: %s, got: %s\n", c.loc.line, c.loc.col, msg, c.value)
}

func parseSelectStatement(tokens []*token, initialCursor uint, delimiter token) (*SelectStatement, uint, bool) {
	cursor := initialCursor
	if !expectToken(tokens, cursor, tokenFromKeyword(selectKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	slct := SelectStatement{}

	exps, newCursor, ok := parseExpressions(tokens, cursor, []token{tokenFromKeyword(fromKeyword), delimiter})
	if !ok {
		return nil, initialCursor, false
	}

	slct.item = *exps
	cursor = newCursor

	if expectToken(tokens, cursor, tokenFromKeyword(fromKeyword)) {
		cursor++

		from, newCursor, ok := parseToken(tokens, cursor, identifierKind)
		if !ok {
			helpMessage(tokens, cursor, "Expected FROM token")
			return nil, initialCursor, false
		}

		slct.from = *from
		cursor = newCursor
	}

	//for where
	where, newCursor, ok := parseWhere(tokens, cursor, tokenFromSymbol(semicolonSymbol))
	if !ok {
		return nil, initialCursor, false
	}

	slct.where = *where
	cursor = newCursor

	return &slct, cursor, true
}

func parseToken(tokens []*token, initialCursor uint, kind tokenKind) (*token, uint, bool) {
	cursor := initialCursor

	if cursor >= uint(len(tokens)) {
		return nil, initialCursor, false
	}

	current := tokens[cursor]
	if current.kind == kind {
		return current, cursor + 1, true
	}

	return nil, initialCursor, false
}

func tokenFromSymbol(s symbol) token {
	return token{
		kind:  symbolKind,
		value: string(s),
	}
}

func parseWhere(tokens []*token, initialCursor uint, delimiter token) (*[]interface{}, uint, bool) {
	cursor := initialCursor

	where := make([]interface{}, 0)
	i := 0
	conditions := conditions{}
	kinds := []tokenKind{keywordKind, identifierKind, numericKind, stringKind, operationKind}
outer:
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		// Look for delimiter
		current := tokens[cursor]
		if delimiter.equals(current) {
			break outer
		}

		for _, kind := range kinds {
			t, newCursor, ok := parseToken(tokens, cursor, kind)
			if ok {
				if kind == keywordKind {
					p := predicate{
						predicate: *t,
					}
					where = append(where, p)
				} else {
					if i == 0 {
						conditions.literal = *t
						i++
					} else if i == 1 {
						conditions.operation = *t
						i++
					} else {
						conditions.value = *t

						where = append(where, conditions)
						i = 0
						cursor = newCursor
						continue
					}
				}
				cursor = newCursor
			}
		}
	}
	return &where, cursor, true
}

func parseExpressions(tokens []*token, initialCursor uint, delimiters []token) (*[]*expression, uint, bool) {
	cursor := initialCursor

	exps := []*expression{}
outer:
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		// Look for delimiter
		current := tokens[cursor]
		for _, delimiter := range delimiters {
			if delimiter.equals(current) {
				break outer
			}
		}

		// Look for comma
		if len(exps) > 0 {
			if !expectToken(tokens, cursor, tokenFromSymbol(commaSymbol)) {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}

			cursor++
		}

		// Look for expression
		exp, newCursor, ok := parseExpression(tokens, cursor, tokenFromSymbol(commaSymbol))
		if !ok {
			helpMessage(tokens, cursor, "Expected expression")
			return nil, initialCursor, false
		}
		cursor = newCursor

		exps = append(exps, exp)
	}

	return &exps, cursor, true
}

func parseExpression(tokens []*token, initialCursor uint, _ token) (*expression, uint, bool) {
	cursor := initialCursor

	kinds := []tokenKind{identifierKind, numericKind, stringKind}
	for _, kind := range kinds {
		t, newCursor, ok := parseToken(tokens, cursor, kind)
		if ok {
			return &expression{
				literal: t,
			}, newCursor, true
		}
	}

	return nil, initialCursor, false
}

func Parse(source string) (*SelectStatement, error) {
	tokens, err := lex(source)
	if err != nil {
		return nil, err
	}

	cursor := uint(0)
	stmt, newCursor, ok := parseSelect(tokens, cursor)
	if !ok {
		helpMessage(tokens, cursor, "Expected statement")
		return nil, errors.New("failed to parse, expected statement")
	}
	cursor = newCursor

	atLeastOneSemicolon := false
	for expectToken(tokens, cursor, tokenFromSymbol(semicolonSymbol)) {
		cursor++
		atLeastOneSemicolon = true
	}

	if !atLeastOneSemicolon {
		helpMessage(tokens, cursor, "Expected semi-colon delimiter between statements")
		return nil, errors.New("missing semi-colon between statements")
	}

	return stmt, nil
}

func parseSelect(tokens []*token, initialCursor uint) (*SelectStatement, uint, bool) {
	cursor := initialCursor

	semicolonToken := tokenFromSymbol(semicolonSymbol)
	slct, newCursor, ok := parseSelectStatement(tokens, cursor, semicolonToken)
	if ok {
		return slct, newCursor, true
	}

	return nil, initialCursor, false
}
