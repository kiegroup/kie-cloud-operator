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
	// TODO uncomment for next 8.x release.
	//if *prior {
	//	fmt.Println(version.PriorVersion)
	//} else if *csv {
	//	fmt.Println(version.CsvVersion)
	//} else if *csvPrior {
	//	fmt.Println(version.CsvPriorVersion)
	//} else {
	//	fmt.Println(version.Version)
	//}

	// TODO remove for next 8.x release.
	if *csv {
		fmt.Println(version.CsvVersion)
	} else {
		fmt.Println(version.Version)
	}
}
