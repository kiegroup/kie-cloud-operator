package version

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
)

var (
	// Version - current version
	Version = constants.CurrentVersion
	// CsvVersion - csv release
	CsvVersion = Version + "-1"
	// PriorVersion - prior version
	// TODO uncomment for next 8.x release.
	// PriorVersion = constants.PriorVersion
	// CsvPriorVersion - prior csv release
	// TODO uncomment for next 8.x release.
	// CsvPriorVersion = PriorVersion + "-1"
)
