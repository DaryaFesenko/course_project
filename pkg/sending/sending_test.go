package sending

import (
	"course_project/pkg/parsing"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestSendRequestCheckColumnOK(t *testing.T) {
	s := CsvParser{
		csvModel: &CsvModel{
			columnsName: []string{"test1", "test2", "test3", "test4"},
		},
	}

	sel := CreateSelectStatementColumns([]string{"test1", "test2"}, []string{"test3", "test4"})

	_, ok := s.checkExistingColumnName(&sel)

	assert.Equal(t, ok, true)
}

func TestSendRequestCheckColumnFAIL(t *testing.T) {
	s := CsvParser{
		csvModel: &CsvModel{
			columnsName: []string{"test1", "test2", "test3", "test5", "test6"},
		},
	}

	sel := CreateSelectStatementColumns([]string{"test1", "test2"}, []string{"test3", "test4"})

	_, ok := s.checkExistingColumnName(&sel)

	assert.Equal(t, ok, false)
}

//block helpers
func CreateSelectStatementColumns(item, conditionItem []string) parsing.SelectStatement {
	sel := parsing.SelectStatement{}

	for _, val := range item {
		s := parsing.Expression{
			Literal: &parsing.Token{
				Value: val,
			},
		}

		sel.Item = append(sel.Item, &s)
	}

	for _, val := range conditionItem {
		p := parsing.Conditions{
			Literal: parsing.Token{
				Value: val,
			},
		}

		sel.Where = append(sel.Where, p)
	}

	return sel
}
