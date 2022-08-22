package main

import (
	"flag"
	"fmt"

	"github.com/kiegroup/kie-cloud-operator/version"
)

var (
	prior    = flag.Bool("prior", false, "get prior product version")
	csv      = flag.Bool("csv", false, "get csv version")
	csvPrior = flag.Bool("csvPrior", false, "get prior csv version")
)

func main() {
	flag.Parse()
	if *prior {
		fmt.Println(version.PriorVersion)
	} else if *csv {
		fmt.Println(version.CsvVersion)
	} else if *csvPrior {
		fmt.Println(version.CsvPriorVersion)
	} else {
		fmt.Println(version.Version)
	}
}
