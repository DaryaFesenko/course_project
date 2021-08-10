package main

import (
	"fmt"
	//"course_project/parsing"
)

const p = "papapa"

func main() {
	t := new(string)
	*t = p

	y := new(string)
	*y = p
	a := t
	b := y

	if *a != *b {
		fmt.Println(*a, *b)
	}
	/*
		// дописать в обработчик обработку 'ыекштп' для where
		sel, err := parsing.Parse("continent='Asia' AND date>'2020-04-14';")

		fmt.Println(sel, err)*/
}
