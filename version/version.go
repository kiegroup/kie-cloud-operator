package version

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
)

var (
	// Version - current version
	Version = constants.CurrentVersion
	// CsvVersion - csv release
	CsvVersion = Version + "-3"
	// PriorVersion - prior version
	PriorVersion = constants.CurrentVersion
	// CsvPriorVersion - prior csv release
	CsvPriorVersion = PriorVersion + "-2"
)
