package main

import (
	"fmt"
	"os"

	"github.com/gobuffalo/packr/v2/jam"
)

func main() {
	fmt.Println("Generating packr boxes...")
	err := jam.Pack(jam.PackOptions{})
	if err != nil {
		fmt.Println("Failed to get process packr files. ", err)
		os.Exit(1)
	}
}
