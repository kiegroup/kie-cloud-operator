package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gobuffalo/packr/builder"
	"github.com/sirupsen/logrus"
)

func main() {
	b := builder.New(context.Background(), os.Args[1])
	// b.Compress = true

	fmt.Println("Generating packr boxes")

	err := b.Run()
	if err != nil {
		logrus.Fatal(err)
	}
}
