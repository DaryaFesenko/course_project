package parsing

type Keyword string

const (
	SelectKeyword Keyword = "select"
	WhereKeyword  Keyword = "where"
	FromKeyword   Keyword = "from"
	AndKeyword    Keyword = "and"
	OrKeyword     Keyword = "or"
)

type TokenKind uint

const (
	KeywordKind TokenKind = iota
	SymbolKind
	IdentifierKind
	StringKind
	NumericKind
	OperationKind
)

type symbol string

const (
	commaSymbol     symbol = ","
	semicolonSymbol symbol = ";"
	allFields       symbol = "*"
)

type Operation string

const (
	MoreOperation      Operation = ">"
	LessOperation      Operation = "<"
	MoreEqualOperation Operation = ">="
	LessEqualOperation Operation = "<="
	EqualsOperation    Operation = "="
	NotEqualOperation  Operation = "!="
)
