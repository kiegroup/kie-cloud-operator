package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	"github.com/kiegroup/kie-cloud-operator/version"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)
var log = logs.GetLogger("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
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

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error("Failed to get watch namespace. ", err)
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error("Error while getting config. ", err)
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	if err := leader.Become(ctx, "kie-cloud-operator-lock"); err != nil {
		log.Error("Error becoming leader: %v", err)
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	syncPeriod := time.Duration(2) * time.Hour
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error("Error getting Manager. ", err)
		os.Exit(1)
	}

	/*
		// Check for OpenShift cluster
		isOpenShift, err := openshiftutil.IsOpenShift(mgr.GetConfig())
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		if !isOpenShift {
			log.Error("OpenShift not detected, exiting")
			os.Exit(1)
		}
	*/

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error("Error on AddToScheme. ", err)
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error("Error adding controllers to Manager. ", err)
		os.Exit(1)
	}

	// Create Service object to expose the metrics port.
	_, err = metrics.ExposeMetricsPort(ctx, metricsPort)
	if err != nil {
		log.Info("Error exposing metrics. ", err)
	}

	log.Info("Starting the Operator.")

	message := "ConfigMaps not available. Using embedded configs."
	if os.Getenv(constants.NameSpaceEnv) == "" {
		log.Warnf("%s required env %s not set, please configure downward API", message, constants.NameSpaceEnv)
	}
	if os.Getenv(constants.OpNameEnv) == "" {
		log.Warnf("%s required env %s not set, please configure env", message, constants.OpNameEnv)
	}

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error("Manager exited non-zero. ", err)
		os.Exit(1)
	}
}
