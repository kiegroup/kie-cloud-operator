package main

import (
	"flag"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/logs"
	"runtime"

	"github.com/kiegroup/kie-cloud-operator/pkg/ui"
	"github.com/kiegroup/kie-cloud-operator/version"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/spf13/pflag"
)

var log = logs.GetLogger("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Kie Operator Version: %v", version.Version))
	log.Info("")
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	flag.Parse()

	printVersion()

	ui.Listen()
}
