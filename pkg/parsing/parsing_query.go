package parsing

import (
	"fmt"
)

type Location struct {
	Line uint
	Col  uint
}

type Token struct {
	Value string
	Kind  TokenKind
	Loc   Location
}

type SelectStatement struct {
	Item       []*Expression
	From       Token
	Where      []interface{}
	IsAllItems bool
}

type Expression struct {
	Literal *Token
}

//for where
type Conditions struct {
	Literal   Token
	Operation Token
	Value     Token
}

//for where
type Predicate struct {
	Predicate Token
}

func (t *Token) equals(other *Token) bool {
	return t.Value == other.Value && t.Kind == other.Kind
}

func (p *Parser) expectToken(cursor uint, t Token) bool {
	if cursor >= uint(len(p.tokens)) {
		return false
	}

	return t.equals(p.tokens[cursor])
}

func (p *Parser) tokenFromKeyword(k Keyword) Token {
	return Token{
		Kind:  KeywordKind,
		Value: string(k),
	}
}

func (p *Parser) helpMessage(cursor uint, msg string) error {
	var c *Token
	if cursor < uint(len(p.tokens)) {
		c = p.tokens[cursor]
	} else {
		c = p.tokens[cursor-1]
	}

	return fmt.Errorf("%s, got: %s", msg, c.Value)
}

func (p *Parser) parseSelectStatement(initialCursor uint, delimiter Token) (*SelectStatement, bool, error) {
	cursor := initialCursor
	if !p.expectToken(cursor, p.tokenFromKeyword(SelectKeyword)) {
		return nil, false, nil
	}
	cursor++

	slct := SelectStatement{}

	exps, newCursor, ok, isAllItems, err := p.parseExpressions(cursor, []Token{p.tokenFromKeyword(FromKeyword), delimiter})
	if !ok {
		return nil, false, err
	}

	if isAllItems {
		slct.IsAllItems = isAllItems
	} else {
		slct.Item = *exps
	}

	cursor = newCursor

	if p.expectToken(cursor, p.tokenFromKeyword(FromKeyword)) {
		cursor++

		from, newCursor, ok := p.parseToken(cursor, IdentifierKind)
		if !ok {
			err := p.helpMessage(cursor, "Expected FROM token")
			return nil, false, err
		}

		slct.From = *from
		cursor = newCursor
	}

	//for where
	where, ok := p.parseWhere(cursor, p.tokenFromSymbol(semicolonSymbol))
	if !ok {
		return nil, false, nil
	}

	slct.Where = *where

	return &slct, true, nil
}

func (p *Parser) parseToken(initialCursor uint, kind TokenKind) (*Token, uint, bool) {
	cursor := initialCursor

	if cursor >= uint(len(p.tokens)) {
		return nil, initialCursor, false
	}

	current := p.tokens[cursor]
	if current.Kind == kind {
		return current, cursor + 1, true
	}

	return nil, initialCursor, false
}

func (p *Parser) tokenFromSymbol(s symbol) Token {
	return Token{
		Kind:  SymbolKind,
		Value: string(s),
	}
}

func (p *Parser) parseWhere(initialCursor uint, delimiter Token) (*[]interface{}, bool) {
	cursor := initialCursor

	where := make([]interface{}, 0)
	i := 0
	conditions := Conditions{}
	kinds := []TokenKind{KeywordKind, IdentifierKind, NumericKind, StringKind, OperationKind}
outer:
	for {
		if cursor >= uint(len(p.tokens)) {
			return nil, false
		}

		// Look for delimiter
		current := p.tokens[cursor]
		if delimiter.equals(current) {
			break outer
		}

		for _, kind := range kinds {
			t, newCursor, ok := p.parseToken(cursor, kind)
			if ok {
				if kind == KeywordKind {
					if Keyword(t.Value) != WhereKeyword {
						p := Predicate{
							Predicate: *t,
						}
						where = append(where, p)
					}
				} else {
					if i == 0 {
						conditions.Literal = *t
						i++
					} else if i == 1 {
						conditions.Operation = *t
						i++
					} else {
						conditions.Value = *t

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
	return &where, true
}

func (p *Parser) parseExpressions(initialCursor uint, delimiters []Token) (*[]*Expression, uint, bool, bool, error) {
	cursor := initialCursor

	isAllItems := false

	exps := []*Expression{}
outer:
	for {
		if cursor >= uint(len(p.tokens)) {
			return nil, initialCursor, false, isAllItems, nil
		}

		current := p.tokens[cursor]
		if current.Kind == SymbolKind {
			if current.Value == string(allFields) {
				isAllItems = true
				cursor++
				return &exps, cursor, true, isAllItems, nil
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
				err := p.helpMessage(cursor, "Expected comma")
				return nil, initialCursor, false, isAllItems, err
			}

			cursor++
		}

		// Look for expression
		exp, newCursor, ok := p.parseExpression(cursor, p.tokenFromSymbol(commaSymbol))
		if !ok {
			err := p.helpMessage(cursor, "Expected expression")
			return nil, initialCursor, false, isAllItems, err
		}
		cursor = newCursor

		exps = append(exps, exp)
	}

	return &exps, cursor, true, isAllItems, nil
}

func (p *Parser) parseExpression(initialCursor uint, _ Token) (*Expression, uint, bool) {
	cursor := initialCursor

	kinds := []TokenKind{IdentifierKind, NumericKind, StringKind}
	for _, kind := range kinds {
		t, newCursor, ok := p.parseToken(cursor, kind)
		if ok {
			return &Expression{
				Literal: t,
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
	stmt, ok, err := p.parseSelect(cursor)
	if !ok {
		er := p.helpMessage(cursor, "Expected statement")

		if err != nil {
			return nil, fmt.Errorf("messageErrorParseSelect: %v", err)
		}
		return nil, fmt.Errorf("messageError: %v", er)
	}

	return stmt, nil
}

func (p *Parser) parseSelect(initialCursor uint) (*SelectStatement, bool, error) {
	cursor := initialCursor

	semicolonToken := p.tokenFromSymbol(semicolonSymbol)
	slct, ok, err := p.parseSelectStatement(cursor, semicolonToken)
	if ok {
		return slct, true, err
	}

	return nil, false, err
}

type Parser struct {
	tokens []*Token
}

func NewParser() Parser {
	return Parser{}
}
