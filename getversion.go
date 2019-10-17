package main

import (
	"flag"
	"fmt"

	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/version"
)

var (
	operator = flag.Bool("operator", false, "get current operator version")
	product  = flag.Bool("product", false, "get current product version")
)

func main() {
	flag.Parse()
	if !*operator && !*product {
		fmt.Println("Operator version is " + version.Version)
		fmt.Println("Product version is " + constants.CurrentVersion)
	}
	if *operator {
		fmt.Println(version.Version)
	}
	if *product {
		fmt.Println(constants.CurrentVersion)
	}
}
