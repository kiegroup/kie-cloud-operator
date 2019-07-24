package openshift

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("env")

func IsOpenShift(cfg *rest.Config) (bool, error) {
	log.Info("attempting detection of OpenShift platform...")

	if cfg == nil {
		var err error
		cfg, err = config.GetConfig()
		if err != nil {
			log.Error(err, "error in fetching config, returning false")
			return false, err
		}
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		log.Error(err, "error in fetching discovery client, returning false")
		return false, err
	}

	apiList, err := discoveryClient.ServerGroups()
	if err != nil {
		log.Error(err, "error in getting ServerGroups from discovery client, returning false")
		return false, err
	}

	for _, v := range apiList.Groups {
		if v.Name == "route.openshift.io" {
			log.Info("OpenShift route detected in api groups, returning true")
			return true, nil
		}
	}

	log.Info("OpenShift route not found in groups, returning false")
	return false, nil
}
