package main

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/kiegroup/kie-cloud-operator/internal/app/handler"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/sirupsen/logrus"
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
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	resource := "kiegroup.org/v1"
	kind := "App"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}
	resyncPeriod := time.Duration(5) * time.Second
	logrus.Infof("Watching %s of type %s in project %s, every %d seconds", resource, kind, namespace, (resyncPeriod / time.Second))
	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Handle(handler.NewHandler())
	sdk.Run(context.TODO())
}
