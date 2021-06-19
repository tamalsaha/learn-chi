package main

import (
	"fmt"
	"reflect"
)

type Person struct {
	Name string
}

func main() {
	a := Person{
		Name: "a",
	}
	fmt.Printf("%+v\n", a)

	pt := reflect.TypeOf(a)

	// creates new object of Pointer type
	b_ptr := reflect.New(pt).Interface().(*Person)
	b_ptr.Name = "b"
	fmt.Printf("%+v\n", b_ptr)

	// creates new object of Pointer type. We use .Elem() to get the value type
	b := reflect.New(pt).Elem().Interface().(Person) // .Interface().(*Person)
	b.Name = "b"
	fmt.Printf("%+v\n", b)
}
