package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	"github.com/kiegroup/kie-cloud-operator/version"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/ready"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var log = logs.GetLogger("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("operator-sdk Version: %v", sdkVersion.Version))
	log.Info(fmt.Sprintf("Kie Operator Version: %v", version.Version))
	log.Info("")
}

func main() {
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

	// Become the leader before proceeding
	leader.Become(context.TODO(), "kie-cloud-operator-lock")

	r := ready.NewFileReady()
	err = r.Set()
	if err != nil {
		log.Error("Error on NewFileReady(). ", err)
		os.Exit(1)
	}
	defer r.Unset()

	// Create a new Cmd to provide shared dependencies and start components
	syncPeriod := time.Duration(2) * time.Hour
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace, SyncPeriod: &syncPeriod})
	if err != nil {
		log.Error("Error getting Manager. ", err)
		os.Exit(1)
	}

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

	log.Info("Starting the Operator.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error("Manager exited non-zero. ", err)
		os.Exit(1)
	}
}
