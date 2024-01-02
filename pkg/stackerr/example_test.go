package stackerr

import (
	"errors"
	"fmt"
)

// Examples are in their own file because their output has hard-coded line
// numbers. This makes it easier to change/add other tests without constantly
// breaking the examples.

func Example_wrapped() {
	// user := "gcurtis"
	wrapped := Errorf("wrong password")
	// err := Errorf("login %q: %w", user, wrapped)
	fmt.Printf("error: %+v\n", wrapped)

	// Output:
	// error: login "gcurtis": wrong password
	// example_test.go:15 login "gcurtis": wrong password
	// example_test.go:14 wrong password
}

func Example_joined() {
	errA := Errorf("error a")
	err1 := Errorf("error 1: %w", errA)
	err2 := Errorf("error 2")
	err3 := Errorf("error 3")
	err := Errorf("joined errors:\n%w", errors.Join(err1, err2, err3))
	fmt.Printf("error: %+v\n", err)

	// Output:
	// error: joined errors:
	// error 1: error a
	// error 2
	// error 3
	// example_test.go:29 "joined errors:\nerror 1: error a\nerror 2\nerror 3"
	// 	[0] example_test.go:26 error 1: error a
	// 	    example_test.go:25 error a
	// 	[1] example_test.go:27 error 2
	// 	[2] example_test.go:28 error 3
}
