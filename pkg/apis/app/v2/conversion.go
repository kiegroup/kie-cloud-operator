package v2

import (
	"encoding/json"
	"fmt"

	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// SIG feature details - https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/20190425-crd-conversion-webhook.md

// Actual conversion methods

func convertKieAppV1toV2(kieAppV1 *v1.KieApp) (*KieApp, error) {
	/*
		items := strings.Split(kieAppV1.Spec, " ")
		if len(items) != 5 {
		   return nil, fmt.Errorf("invalid spec string, needs five parts: %s", kieAppV1.Spec)
		}
	*/
	return &KieApp{
		ObjectMeta: kieAppV1.ObjectMeta,
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       kieAppV1.Kind,
		},
		Spec: KieAppSpec{
			Version: kieAppV1.Spec.CommonConfig.Version,
			Upgrades: KieAppUpgrades{
				Enabled: *kieAppV1.Spec.Upgrades.Patch,
				Minor:   kieAppV1.Spec.Upgrades.Minor,
			},
		},
	}, nil
}

func convertKieAppV2toV1(kieAppV2 *KieApp) (*v1.KieApp, error) {
	return &v1.KieApp{
		ObjectMeta: kieAppV2.ObjectMeta,
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       kieAppV2.Kind,
		},
		Spec: v1.KieAppSpec{
			CommonConfig: v1.CommonConfig{
				Version: kieAppV2.Spec.Version,
			},
			Upgrades: v1.KieAppUpgrades{
				Patch: &kieAppV2.Spec.Upgrades.Enabled,
				Minor: kieAppV2.Spec.Upgrades.Minor,
			},
		},
	}, nil
}

func convert(in runtime.RawExtension, version string) (*runtime.RawExtension, error) {
	inAPIVersion, err := extractAPIVersion(in)
	if err != nil {
		return nil, err
	}
	switch inAPIVersion {
	case v1.SchemeGroupVersion.String():
		var kieAppV1 v1.KieApp
		if err := json.Unmarshal(in.Raw, &kieAppV1); err != nil {
			return nil, err
		}
		switch version {
		case v1.SchemeGroupVersion.String():
			// This should not happened as API server will not call the webhook in this case
			return &in, nil
		case SchemeGroupVersion.String():
			kieAppV2, err := convertKieAppV1toV2(&kieAppV1)
			if err != nil {
				return nil, err
			}
			raw, err := json.Marshal(kieAppV2)
			if err != nil {
				return nil, err
			}
			return &runtime.RawExtension{Raw: raw}, nil
		}
	case SchemeGroupVersion.String():
		var kieAppV2 KieApp
		if err := json.Unmarshal(in.Raw, &kieAppV2); err != nil {
			return nil, err
		}
		switch version {
		case SchemeGroupVersion.String():
			// This should not happened as API server will not call the webhook in this case
			return &in, nil
		case v1.SchemeGroupVersion.String():
			kieAppV1, err := convertKieAppV2toV1(&kieAppV2)
			if err != nil {
				return nil, err
			}
			raw, err := json.Marshal(kieAppV1)
			if err != nil {
				return nil, err
			}
			return &runtime.RawExtension{Raw: raw}, nil
		}
	default:
		return nil, fmt.Errorf("invalid conversion fromVersion requested: %s", inAPIVersion)
	}
	return nil, fmt.Errorf("invalid conversion toVersion requested: %s", version)
}

func extractAPIVersion(in runtime.RawExtension) (string, error) {
	object := unstructured.Unstructured{}
	if err := object.UnmarshalJSON(in.Raw); err != nil {
		return "", err
	}
	return object.GetAPIVersion(), nil
}

//  ??? move this code and serve via console pod/service... `pkg/controller/kieapp/deploy_ui.go`
/*
func serveKieAppConversion(w http.ResponseWriter, r *http.Request) {
	request, err := readConversionRequest(r)
	if err != nil {
		reportError(w, err)
	}
	response := ConversionResponse{}
	response.UID = request.UID
	converted, err := convert(request.Object, request.APIVersion)
	if err != nil {
		reportError(w, err)
	}
	response.ConvertedObject = *converted
	writeConversionResponse(w, response)
}
*/

// ADD TO CRD WHEN READY
//
//  conversion:
//    strategy: Webhook
//    webhookClientConfig:
//      url: https://console-cr-form/my-webhook-path
//
