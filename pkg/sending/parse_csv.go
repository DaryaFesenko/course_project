package sending

import (
	"course_project/pkg/parsing"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CsvParser struct {
	csvFilePath string
	tableName   string
	csvModel    *CsvModel
}

func New(csv string) (*CsvParser, error) {
	i := strings.LastIndex(csv, ".csv")
	tableName := csv[:i]

	c := &CsvParser{csvFilePath: csv, csvModel: &CsvModel{}, tableName: tableName}
	if err := c.initCsvModel(); err != nil {
		return c, err
	}

	return c, nil
}

func (c *CsvParser) initCsvModel() error {
	f, err := os.Open(c.csvFilePath)

	if err != nil {
		return err
	}

	defer f.Close()

	r := csv.NewReader(f)

	if c.csvModel.columnsName, err = r.Read(); err != nil {
		return err
	}

	c.csvModel.data, err = r.ReadAll()

	if err != nil {
		return err
	}

	for idx := range c.csvModel.columnsName {
		c.csvModel.columnsName[idx] = strings.ToLower(c.csvModel.columnsName[idx])
	}

	return nil
}

func (c *CsvParser) SendRequest(request *parsing.SelectStatement) ([][]string, error) {
	if !request.IsAllItems {
		if val, ok := c.checkExistingColumnName(request); !ok {
			return [][]string{}, fmt.Errorf("no such column name '%s' in csv", val)
		}
	}

	err := c.csvModel.checkOnInt()
	if err != nil {
		return [][]string{}, err
	}
	res, err := c.request(request)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (c *CsvParser) checkExistingColumnName(request *parsing.SelectStatement) (string, bool) {
	for idx := range request.Item {
		item := request.Item[idx]

		if !c.csvModel.FindColumnNameFromCsv(item.Literal.Value) {
			return item.Literal.Value, false
		}
	}

	for _, val := range request.Where {
		item, ok := val.(parsing.Conditions)
		if ok {
			if !c.csvModel.FindColumnNameFromCsv(item.Literal.Value) {
				return item.Literal.Value, false
			}
		}
	}

	return "", true
}

//gocyclo:ignore
func (c *CsvParser) request(request *parsing.SelectStatement) ([][]string, error) {
	var columnInput []int
	var resultData [][]string

	if request.From.Value != c.tableName {
		return resultData, fmt.Errorf("can`t find csv with name '%s'", request.From.Value)
	}

	if !request.IsAllItems {
		var col []string
		for idx := range request.Item {
			item := request.Item[idx]

			columnInput = append(columnInput, c.csvModel.GetIdxColumnName(item.Literal.Value))
			col = append(col, item.Literal.Value)
		}

		resultData = append(resultData, col)
	}

	for _, str := range c.csvModel.data {
		isAdd := false
		isAndPredicate := false
		isOrPredicate := false

		for idx := range request.Where {
			item := request.Where[idx]

			if idx%2 == 0 {
				val, ok := item.(parsing.Conditions)

				if !ok {
					return resultData, fmt.Errorf("incorrect expression 'where' format")
				}

				idxColumn := c.csvModel.GetIdxColumnName(val.Literal.Value)
				if idxColumn == -1 {
					return resultData, fmt.Errorf("")
				}

				ok, err := c.isConditionOperation(str[idxColumn], val.Operation, val.Value)

				if err != nil {
					return resultData, err
				}

				if ok {
					if idx != 0 {
						if isAndPredicate {
							if !isAdd {
								isAdd = false
								break
							}
						}
					}

					isAdd = true
				} else {
					isAdd = false

					if isAndPredicate {
						break
					}
				}
			} else {
				val, ok := item.(parsing.Predicate)

				if !ok {
					return resultData, fmt.Errorf("incorrect expression 'where' format")
				}

				if val.Predicate.Value == string(parsing.AndKeyword) {
					if idx != 1 && isOrPredicate {
						return resultData, fmt.Errorf("cannot combine AND/OR")
					}

					isAndPredicate = true
				} else if val.Predicate.Value == string(parsing.OrKeyword) {
					if idx != 1 && isAndPredicate {
						return resultData, fmt.Errorf("cannot combine AND/OR")
					}

					isOrPredicate = true
				}
			}
		}

		if isAdd || len(request.Where) == 0 {
			if !request.IsAllItems {
				var strInput []string
				for _, idx := range columnInput {
					strInput = append(strInput, str[idx])
				}

				resultData = append(resultData, strInput)
			} else {
				resultData = append(resultData, str)
			}
		}
	}

	return resultData, nil
}

//написать тест
//gocyclo:ignore
func (c *CsvParser) isConditionOperation(data string, operation parsing.Token, value parsing.Token) (bool, error) {
	var dataInt, valueInt int
	var isInt bool
	var err error

	if value.Kind == parsing.NumericKind {
		dataInt, err = strconv.Atoi(data)
		if err != nil {
			return false, fmt.Errorf("types do not match. data %s must be type int. %v", data, err)
		}
		valueInt, err = strconv.Atoi(value.Value)
		if err != nil {
			return false, err
		}
		isInt = true
	}

	switch parsing.Operation(operation.Value) {
	case parsing.EqualsOperation:
		if isInt {
			if dataInt == valueInt {
				return true, nil
			}
		} else {
			if data == value.Value {
				return true, nil
			}
		}
	case parsing.LessEqualOperation:
		if isInt {
			if dataInt <= valueInt {
				return true, nil
			}
		} else {
			if data <= value.Value {
				return true, nil
			}
		}
	case parsing.LessOperation:
		if isInt {
			if dataInt < valueInt {
				return true, nil
			}
		} else {
			if data < value.Value {
				return true, nil
			}
		}
	case parsing.MoreEqualOperation:
		if isInt {
			if dataInt >= valueInt {
				return true, nil
			}
		} else {
			if data >= value.Value {
				return true, nil
			}
		}
	case parsing.MoreOperation:
		if isInt {
			if dataInt > valueInt {
				return true, nil
			}
		} else {
			if data > value.Value {
				return true, nil
			}
		}
	case parsing.NotEqualOperation:
		if isInt {
			if dataInt != valueInt {
				return true, nil
			}
		} else {
			if data != value.Value {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("operation '%s' does not supported", operation.Value)
	}

	return false, nil
}

func (c *CsvParser) GetColumnNames() []string {
	return c.csvModel.columnsName
}
