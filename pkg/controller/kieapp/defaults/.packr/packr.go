package main

import (
	"fmt"

	"github.com/gobuffalo/packr/v2/jam"
)

func main() {
	fmt.Println("Generating packr boxes...")
	jam.Pack(jam.PackOptions{})
}
