package sending

import (
	"regexp"
	"strings"
)

type CsvModel struct {
	columnsName []string
	data        [][]string
}

func (c *CsvModel) FindColumnNameFromCsv(name string) bool {
	for _, val := range c.columnsName {
		if val == name {
			return true
		}
	}

	return false
}

func (c *CsvModel) GetIdxColumnName(name string) int {
	for idx := range c.columnsName {
		if c.columnsName[idx] == name {
			return idx
		}
	}

	return -1
}

func (c *CsvModel) checkOnInt() error {
	for idx, str := range c.data {
		for idxStr, val := range str {
			ok, err := regexp.MatchString("\\d+[.]\\d+", val)
			if err != nil {
				return err
			}
			if ok {
				values := strings.Split(val, ".")

				if values[1] == "0" {
					c.data[idx][idxStr] = values[0]
				}
			}
		}
	}

	return nil
}
