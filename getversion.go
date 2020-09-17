package main

import (
	"flag"
	"fmt"

	"github.com/kiegroup/kie-cloud-operator/version"
)

var (
	prior = flag.Bool("prior", false, "get prior product version")
)

func main() {
	flag.Parse()
	if *prior {
		fmt.Println(version.PriorVersion)
	} else {
		fmt.Println(version.Version)
	}
}
