package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gobuffalo/packr/builder"
)

func main() {
	b := builder.New(context.Background(), os.Args[1])
	// b.Compress = true

	fmt.Println("Generating packr boxes")

	err := b.Run()
	if err != nil {
		log.Fatal(err)
	}
}
