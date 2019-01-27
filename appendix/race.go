// +build ignore

package main

import "fmt"

// run with command `go run -race race.go`

func main() {
	var data int

	go func() {
		data++
	}()

	if 0 == data {
		fmt.Printf("the value is %v.\n", data)
	}
}
