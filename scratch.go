package main

import "fmt"

func main() {
	var str string = "Hello World"

	str2 := &str

	fmt.Println(&str)  // Get the memory address
	fmt.Println(*str2) // Get the actual value by dereferencing the pointer

}
