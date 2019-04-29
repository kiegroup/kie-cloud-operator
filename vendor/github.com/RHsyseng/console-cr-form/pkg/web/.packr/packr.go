package main

import (
	"fmt"

	"github.com/gobuffalo/packr/v2/jam"
)

func main() {
	fmt.Println("Generating packr boxes...")
	_ = jam.Pack(jam.PackOptions{})
}
