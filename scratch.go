package main

import "fmt"

type Point struct {
	x *int
}

type contextKey string

func main() {
	var str string = "Hello World"
	var i *int

	num := Point{x: i}

	fmt.Print(num.x)

	str2 := &str

	fmt.Println(&str)  // Get the memory address
	fmt.Println(*str2) // Get the actual value by dereferencing the pointer

	fmt.Println(contextKey(1))
}
