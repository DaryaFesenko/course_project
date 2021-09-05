package parsing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOK(t *testing.T) {
	testData := []struct {
		InRequest string
		OutData   OutData
	}{
		{
			InRequest: "select * from table where col1 > 6 and col2 = 'test';",
			OutData:   CreateOutData(true, []string{}, []string{"col1", ">", "6"}, []string{"col2", "=", "test"}, "table", "and"),
		},
		{
			InRequest: "select col1, col2 from table where col1 >= 6;",
			OutData:   CreateOutData(false, []string{"col1", "col2"}, []string{"col1", ">=", "6"}, []string{}, "table", ""),
		},
	}

	p := NewParser()
	for _, data := range testData {
		out, err := p.Parse(data.InRequest)

		assert.Equal(t, err, nil)
		EqualOutData(t, out, data.OutData)
	}
}

func TestParseFAIL(t *testing.T) {
	testData := []struct {
		InRequest string
		Error     error
	}{
		{
			InRequest: "select col col from table where col1 > 6 and col2 = 'test';",
			Error:     fmt.Errorf("messageErrorParseSelect: Expected comma, got: col"),
		},
		{
			InRequest: "select col1, col2 from where col1 >= 6;",
			Error:     fmt.Errorf("messageErrorParseSelect: Expected FROM token, got: where"),
		},
		{
			InRequest: "select col1, from table where col1 > 6 and col2 = 'test';",
			Error:     fmt.Errorf("messageErrorParseSelect: Expected expression, got: from"),
		},
		{
			InRequest: "col1, col2 from table where col1 >= 6;",
			Error:     fmt.Errorf("messageErrorParseSelect: Expected SELECT statement"),
		},
	}

	p := NewParser()
	for _, data := range testData {
		_, err := p.Parse(data.InRequest)

		assert.Equal(t, err, data.Error)
	}
}

//block helpers
func EqualOutData(t *testing.T, sel *SelectStatement, out OutData) {
	assert.Equal(t, out.IsAll, sel.IsAllItems)
	assert.Equal(t, out.Table, sel.From.Value)

	items := make([]string, 0, len(sel.Item))
	for _, val := range sel.Item {
		items = append(items, val.Literal.Value)
	}

	assert.ElementsMatch(t, items, out.Columns)

	var con []string
	for _, val := range sel.Where {
		item, ok := val.(Predicate)

		if ok {
			assert.Equal(t, out.Predicate, item.Predicate.Value)
		} else {
			eq := out.Condition1
			if len(con) != 0 {
				eq = out.Condition2
				con = []string{}
			}

			item, ok := val.(Conditions)

			if ok {
				con = append(con, item.Literal.Value, item.Operation.Value, item.Value.Value)
			}

			assert.ElementsMatch(t, eq, con)
		}
	}
}

type OutData struct {
	IsAll      bool
	Columns    []string
	Table      string
	Condition1 []string
	Predicate  string
	Condition2 []string
}

func CreateOutData(isAll bool, columns, con1, con2 []string, table, predicate string) OutData {
	return OutData{
		IsAll:      isAll,
		Columns:    columns,
		Table:      table,
		Condition1: con1,
		Predicate:  predicate,
		Condition2: con2,
	}
}
