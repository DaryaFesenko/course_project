package parsing

import (
	"course_project/app"
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
	item       []*expression
	from       token
	where      []interface{}
	isAllItems bool
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

func (p *Parser) expectToken(cursor uint, t token) bool {
	if cursor >= uint(len(p.tokens)) {
		return false
	}

	return t.equals(p.tokens[cursor])
}

func (p *Parser) tokenFromKeyword(k keyword) token {
	return token{
		kind:  keywordKind,
		value: string(k),
	}
}

func (p *Parser) helpMessage(cursor uint, msg string) {
	var c *token
	if cursor < uint(len(p.tokens)) {
		c = p.tokens[cursor]
	} else {
		c = p.tokens[cursor-1]
	}

	p.app.LogAccess(fmt.Sprintf("[%d,%d]: %s, got: %s\n", c.loc.line, c.loc.col, msg, c.value))
}

func (p *Parser) parseSelectStatement(initialCursor uint, delimiter token) (*SelectStatement, uint, bool) {
	cursor := initialCursor
	if !p.expectToken(cursor, p.tokenFromKeyword(selectKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	slct := SelectStatement{}

	exps, newCursor, ok, isAllItems := p.parseExpressions(cursor, []token{p.tokenFromKeyword(fromKeyword), delimiter})
	if !ok {
		return nil, initialCursor, false
	}

	if isAllItems {
		slct.isAllItems = isAllItems
	} else {
		slct.item = *exps
	}

	cursor = newCursor

	if p.expectToken(cursor, p.tokenFromKeyword(fromKeyword)) {
		cursor++

		from, newCursor, ok := p.parseToken(cursor, identifierKind)
		if !ok {
			p.helpMessage(cursor, "Expected FROM token")
			return nil, initialCursor, false
		}

		slct.from = *from
		cursor = newCursor
	}

	//for where
	where, newCursor, ok := p.parseWhere(cursor, p.tokenFromSymbol(semicolonSymbol))
	if !ok {
		return nil, initialCursor, false
	}

	slct.where = *where
	cursor = newCursor

	return &slct, cursor, true
}

func (p *Parser) parseToken(initialCursor uint, kind tokenKind) (*token, uint, bool) {
	cursor := initialCursor

	if cursor >= uint(len(p.tokens)) {
		return nil, initialCursor, false
	}

	current := p.tokens[cursor]
	if current.kind == kind {
		return current, cursor + 1, true
	}

	return nil, initialCursor, false
}

func (p *Parser) tokenFromSymbol(s symbol) token {
	return token{
		kind:  symbolKind,
		value: string(s),
	}
}

func (p *Parser) parseWhere(initialCursor uint, delimiter token) (*[]interface{}, uint, bool) {
	cursor := initialCursor

	where := make([]interface{}, 0)
	i := 0
	conditions := conditions{}
	kinds := []tokenKind{keywordKind, identifierKind, numericKind, stringKind, operationKind}
outer:
	for {
		if cursor >= uint(len(p.tokens)) {
			return nil, initialCursor, false
		}

		// Look for delimiter
		current := p.tokens[cursor]
		if delimiter.equals(current) {
			break outer
		}

		for _, kind := range kinds {
			t, newCursor, ok := p.parseToken(cursor, kind)
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

func (p *Parser) parseExpressions(initialCursor uint, delimiters []token) (*[]*expression, uint, bool, bool) {
	cursor := initialCursor

	isAllItems := false

	exps := []*expression{}
outer:
	for {
		if cursor >= uint(len(p.tokens)) {
			return nil, initialCursor, false, isAllItems
		}

		current := p.tokens[cursor]
		if current.kind == symbolKind {
			if current.value == string(allFields) {
				isAllItems = true
				cursor++
				return &exps, cursor, true, isAllItems
			}
		}

		// Look for delimiter
		for _, delimiter := range delimiters {
			if delimiter.equals(current) {
				break outer
			}
		}

		// Look for comma
		if len(exps) > 0 {
			if !p.expectToken(cursor, p.tokenFromSymbol(commaSymbol)) {
				p.helpMessage(cursor, "Expected comma")
				return nil, initialCursor, false, isAllItems
			}

			cursor++
		}

		// Look for expression
		exp, newCursor, ok := p.parseExpression(cursor, p.tokenFromSymbol(commaSymbol))
		if !ok {
			p.helpMessage(cursor, "Expected expression")
			return nil, initialCursor, false, isAllItems
		}
		cursor = newCursor

		exps = append(exps, exp)
	}

	return &exps, cursor, true, isAllItems
}

func (p *Parser) parseExpression(initialCursor uint, _ token) (*expression, uint, bool) {
	cursor := initialCursor

	kinds := []tokenKind{identifierKind, numericKind, stringKind}
	for _, kind := range kinds {
		t, newCursor, ok := p.parseToken(cursor, kind)
		if ok {
			return &expression{
				literal: t,
			}, newCursor, true
		}
	}

	return nil, initialCursor, false
}

func (p *Parser) Parse(source string) (*SelectStatement, error) {
	var err error
	p.tokens, err = lex(source)
	if err != nil {
		return nil, err
	}

	cursor := uint(0)
	stmt, newCursor, ok := p.parseSelect(cursor)
	if !ok {
		p.helpMessage(cursor, "Expected statement")
		return nil, errors.New("failed to parse, expected statement")
	}
	cursor = newCursor

	atLeastOneSemicolon := false
	for p.expectToken(cursor, p.tokenFromSymbol(semicolonSymbol)) {
		cursor++
		atLeastOneSemicolon = true
	}

	if !atLeastOneSemicolon {
		p.helpMessage(cursor, "Expected semi-colon delimiter between statements")
		return nil, errors.New("missing semi-colon between statements")
	}

	return stmt, nil
}

func (p *Parser) parseSelect(initialCursor uint) (*SelectStatement, uint, bool) {
	cursor := initialCursor

	semicolonToken := p.tokenFromSymbol(semicolonSymbol)
	slct, newCursor, ok := p.parseSelectStatement(cursor, semicolonToken)
	if ok {
		return slct, newCursor, true
	}

	return nil, initialCursor, false
}

type Parser struct {
	tokens []*token
	app    app.App
}

func NewParser(app *app.App) Parser {
	return Parser{app: *app}
}
