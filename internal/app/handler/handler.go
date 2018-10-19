package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiegroup/kie-cloud-operator/internal/pkg/defaults"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/kieserver"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/rhpamcentr"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/shared"
	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *opv1.App:
		var err error

		if event.Deleted {
			logrus.Infof("Deleting %s %s", o.Name, o.Kind)
			return nil
		}

		// further work required to support CR object updates
		checkUpdateStatus(o)

		if (o.Status != "Installed") && (o.Status != "Error") {
			var objects []runtime.Object
			if o.Spec.Environment != "" {
				logrus.Infof("Will set up %s environment", o.Spec.Environment)

				objects, err = NewEnv(o)
				if err != nil {
					o.Status = "Error"
					logrus.Error(err)
				}
			}

			for _, obj := range objects {
				var err error
				if o.Status == "Updating" {
					// when this is functional, resourceVersion for each object will need to be known/set - maybe attach to CR status as created?
					err = sdk.Update(obj)
				} else if o.Status != "Error" {
					o.Status = "Installed"
					err = sdk.Create(obj)
				}
				if err != nil {
					if errors.IsAlreadyExists(err) {
						logrus.Warnf("%s already exists, will not be created", obj.GetObjectKind().GroupVersionKind().Kind)
					} else {
						logrus.Errorf("Failed to create object %v: %v", obj, err)
						bytes, err1 := json.Marshal(obj)
						if err1 != nil {
							logrus.Infof("Can't serialize", obj)
						} else {
							logrus.Infof("Object is ", string(bytes))
						}
						return nil
					}
				}
			}

			// Update CR
			err := sdk.Update(o)
			if err != nil {
				logrus.Errorf("failed to update %s status: %v", o.Kind, err)
			}
			if o.Status != "Error" {
				logrus.Infof("%s %s is now installed", o.Name, o.Kind)
			}
		}
	}
	return nil
}

func NewEnv(cr *opv1.App) ([]runtime.Object, error) {
	var objs []runtime.Object
	env, password, err := defaults.GetEnvironment(cr)
	if err != nil {
		return []runtime.Object{}, err
	}
	defer shared.Zeroing(password)

	// console keystore generation
	consoleCN := cr.Name
	for _, r := range env.Console.Routes {
		if shared.CheckTLS(r.Spec.TLS) {
			consoleCN = getRouteHost(r, cr)
			// use host of first tls route in env template
			break
		}
	}
	env.Console.Secrets = append(env.Console.Secrets, corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-businesscentral-app-secret", cr.Name),
			Labels: map[string]string{
				"app": cr.Name,
			},
		},
		Data: map[string][]byte{
			"keystore.jks": shared.GenerateKeystore(consoleCN, "jboss", password),
		},
	})

	// server(s) keystore generation
	serverCN := cr.Name
	for i, server := range env.Servers {
		for _, r := range server.Routes {
			if shared.CheckTLS(r.Spec.TLS) {
				serverCN = getRouteHost(r, cr)
				// use host of first tls route in env template
				break
			}
		}
		server.Secrets = append(server.Secrets, corev1.Secret{
			Type: corev1.SecretTypeOpaque,
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-kieserver-%d-app-secret", cr.Name, i),
				Labels: map[string]string{
					"app": cr.Name,
				},
			},
			Data: map[string][]byte{
				"keystore.jks": shared.GenerateKeystore(serverCN, "jboss", password),
			},
		})

		env.Servers[i] = server
	}

	// create object slice for deployment
	objs = shared.ObjectAppend(objs, rhpamcentr.ConstructObject(env.Console, cr), cr)
	for _, s := range env.Servers {
		objs = shared.ObjectAppend(objs, kieserver.ConstructObject(s, cr), cr)
	}
	for _, o := range env.Others {
		objs = shared.ObjectAppend(objs, o, cr)
	}

	objs = shared.SetReferences(objs, cr)

	return objs, err
}

func getRouteHost(route routev1.Route, cr *opv1.App) string {
	route.SetNamespace(cr.Namespace)
	route.SetOwnerReferences(shared.GetOwnerReferences(cr))
	groupVK := routev1.SchemeGroupVersion.WithKind("Route")
	apiVersion, kind := groupVK.ToAPIVersionAndKind()

	routeClient, _, err := k8sclient.GetResourceClient(apiVersion, kind, cr.Namespace)
	if err != nil {
		logrus.Error(err)
	}

	shared.CreateObject(routeClient, route.DeepCopyObject())
	oByte, err := shared.GetObjectByte(routeClient, route.Name)
	if err != nil {
		logrus.Error(err)
	}
	json.Unmarshal(oByte, &route)

	return route.Spec.Host
}

// figure out later how to know if there is an update to CR, and mark it's status as Updated
func checkUpdateStatus(o *opv1.App) {
	/*
		if o.Status != "" {
			if !reflect.DeepEqual(??.Spec, o.Spec) {
			}
			logrus.Infof("Updating %s %s", o.Kind, o.Name)
			o.Status = "Updating"
		}
	*/
}
