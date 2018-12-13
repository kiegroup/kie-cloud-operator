package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gobuffalo/packr/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("kieapp.packr")

func main() {
	b := builder.New(context.Background(), os.Args[1])
	// b.Compress = true

	fmt.Println("Generating packr boxes...")

	err := b.Run()
	if err != nil {
		log.Error(err, "Error running packr builder")
		os.Exit(1)
	}
}
