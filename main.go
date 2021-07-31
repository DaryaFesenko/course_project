package main

import (
	"fmt"
	//"course_project/parsing"
)

func main() {

	t := new(string)
	*t = "papapa"

	y := new(string)
	*y = "papapa"
	var a, b *string
	a = t
	b = y

	if *a != *b {

		fmt.Println(*a, *b)
	}
	/*
		// дописать в обработчик обработку 'ыекштп' для where
		sel, err := parsing.Parse("continent='Asia' AND date>'2020-04-14';")

		fmt.Println(sel, err)*/
}
