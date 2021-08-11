package parsing

type keyword string

const (
	selectKeyword keyword = "select"
	whereKeyword  keyword = "where"
	fromKeyword   keyword = "from"
	andKeyword    keyword = "and"
	orKeyword     keyword = "or"
)

type tokenKind uint

const (
	keywordKind tokenKind = iota
	symbolKind
	identifierKind
	stringKind
	numericKind
	operationKind
)

type symbol string

const (
	commaSymbol     symbol = ","
	semicolonSymbol symbol = ";"
)

type operation string

const (
	moreOperation      operation = ">"
	lessOperation      operation = "<"
	moreEqualOperation operation = ">="
	lessEqualOperation operation = "<="
	equalsOperation    operation = "="
	notEqualOperation  operation = "!="
)

const allFields symbol = "*"
