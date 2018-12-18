package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gobuffalo/packr/builder"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
)

var log = logs.GetLogger("kieapp.packr")

func main() {
	b := builder.New(context.Background(), os.Args[1])
	// b.Compress = true

	fmt.Println("Generating packr boxes...")

	err := b.Run()
	if err != nil {
		log.Fatal(err)
	}
}
