package main

import (
	"flag"
	"os"
	"runtime"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller"
	"github.com/kiegroup/kie-cloud-operator/version"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func init() {
	// Set log level... override default w/ command-line variable if set.
	levelString := os.Getenv("LOG_LEVEL") // panic, fatal, error, warn, info, debug
	if levelString == "" {
		levelString = "info"
	}
	lev, err := logrus.ParseLevel(levelString)
	if err != nil {
		lev = logrus.InfoLevel
		logrus.Warnf("Defaulting to INFO level logging, %v", err)
	}
	logrus.SetLevel(lev)

	// Log as JSON instead of the default ASCII formatter.
	//logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr can be any io.Writer, see below for File example
	//logrus.SetOutput(os.Stdout)
}

func printVersion() {
	logrus.Printf("Kie Operator Version: %v", version.Version)
	logrus.Printf("Go Version: %s", runtime.Version())
	logrus.Printf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Printf("operator-sdk Version: %v\n\n", sdkVersion.Version)
}

func main() {
	printVersion()
	flag.Parse()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("failed to get watch namespace: %v", err)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Print("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logrus.Fatal(err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		logrus.Fatal(err)
	}

	logrus.Print("Starting the Operator.")

	// Start the Cmd
	logrus.Fatal(mgr.Start(signals.SetupSignalHandler()))
}
