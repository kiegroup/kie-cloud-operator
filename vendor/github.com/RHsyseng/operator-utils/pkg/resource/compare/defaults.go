package compare

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	oappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"reflect"
)

type resourceComparator struct {
	defaultCompareFunc func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
	compareFuncMap     map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
}

func (this *resourceComparator) SetDefaultComparator(compFunc func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	this.defaultCompareFunc = compFunc
}

func (this *resourceComparator) GetDefaultComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return this.defaultCompareFunc
}

func (this *resourceComparator) SetComparator(resourceType reflect.Type, compFunc func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	this.compareFuncMap[resourceType] = compFunc
}

func (this *resourceComparator) GetComparator(resourceType reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return this.compareFuncMap[resourceType]
}

func (this *resourceComparator) Compare(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	compareFunc := this.GetDefaultComparator()
	type1 := reflect.ValueOf(deployed).Elem().Type()
	type2 := reflect.ValueOf(requested).Elem().Type()
	if type1 == type2 {
		if comparator, exists := this.compareFuncMap[type1]; exists {
			compareFunc = comparator
		}
	}
	return compareFunc(deployed, requested)
}

func defaultMap() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	equalsMap := make(map[reflect.Type]func(resource.KubernetesResource, resource.KubernetesResource) bool)
	equalsMap[reflect.TypeOf(oappsv1.DeploymentConfig{})] = equalDeploymentConfigs
	equalsMap[reflect.TypeOf(corev1.Service{})] = equalServices
	equalsMap[reflect.TypeOf(routev1.Route{})] = equalRoutes
	equalsMap[reflect.TypeOf(rbacv1.Role{})] = equalRoles
	equalsMap[reflect.TypeOf(rbacv1.RoleBinding{})] = equalRoleBindings
	equalsMap[reflect.TypeOf(corev1.ServiceAccount{})] = equalServiceAccounts
	equalsMap[reflect.TypeOf(corev1.Secret{})] = equalSecrets
	return equalsMap
}

func equalDeploymentConfigs(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	dc1 := deployed.(*oappsv1.DeploymentConfig)
	dc2 := requested.(*oappsv1.DeploymentConfig)

	//Removed generated fields from deployed version, when not specified in requested item
	dc1 = dc1.DeepCopy()
	if dc2.Spec.Strategy.RecreateParams == nil {
		dc1.Spec.Strategy.RecreateParams = nil
	}
	if dc2.Spec.Strategy.ActiveDeadlineSeconds == nil {
		dc1.Spec.Strategy.ActiveDeadlineSeconds = nil
	}
	if dc1.Spec.Strategy.RollingParams != nil && dc2.Spec.Strategy.RollingParams != nil {
		if dc2.Spec.Strategy.RollingParams.UpdatePeriodSeconds == nil {
			dc1.Spec.Strategy.RollingParams.UpdatePeriodSeconds = nil
		}
		if dc2.Spec.Strategy.RollingParams.IntervalSeconds == nil {
			dc1.Spec.Strategy.RollingParams.IntervalSeconds = nil
		}
		if dc2.Spec.Strategy.RollingParams.TimeoutSeconds == nil {
			dc1.Spec.Strategy.RollingParams.TimeoutSeconds = nil
		}
	}
	if dc2.Spec.RevisionHistoryLimit == nil {
		dc1.Spec.RevisionHistoryLimit = nil
	}
	if dc1.Spec.Template != nil && dc2.Spec.Template != nil {
		for i := range dc1.Spec.Template.Spec.Volumes {
			if len(dc2.Spec.Template.Spec.Volumes) <= i {
				return false
			}
			volSrc1 := dc1.Spec.Template.Spec.Volumes[i].VolumeSource
			volSrc2 := dc2.Spec.Template.Spec.Volumes[i].VolumeSource
			if volSrc1.Secret != nil && volSrc2.Secret != nil && volSrc2.Secret.DefaultMode == nil {
				volSrc1.Secret.DefaultMode = nil
			}
		}
		if dc2.Spec.Template.Spec.RestartPolicy == "" {
			dc1.Spec.Template.Spec.RestartPolicy = ""
		}
		if dc2.Spec.Template.Spec.DNSPolicy == "" {
			dc1.Spec.Template.Spec.DNSPolicy = ""
		}
		if dc2.Spec.Template.Spec.DeprecatedServiceAccount == "" {
			dc1.Spec.Template.Spec.DeprecatedServiceAccount = ""
		}
		if dc2.Spec.Template.Spec.SecurityContext == nil {
			dc1.Spec.Template.Spec.SecurityContext = nil
		}
		if dc2.Spec.Template.Spec.SchedulerName == "" {
			dc1.Spec.Template.Spec.SchedulerName = ""
		}
		for i := range dc1.Spec.Template.Spec.Containers {
			if len(dc2.Spec.Template.Spec.Containers) <= i {
				return false
			}
			probe1 := dc1.Spec.Template.Spec.Containers[i].LivenessProbe
			probe2 := dc2.Spec.Template.Spec.Containers[i].LivenessProbe
			if probe1 != nil && probe2 != nil {
				if probe2.FailureThreshold == 0 {
					probe1.FailureThreshold = probe2.FailureThreshold
				}
				if probe2.SuccessThreshold == 0 {
					probe1.SuccessThreshold = probe2.SuccessThreshold
				}
			}
			probe1 = dc1.Spec.Template.Spec.Containers[i].ReadinessProbe
			probe2 = dc2.Spec.Template.Spec.Containers[i].ReadinessProbe
			if probe1 != nil && probe2 != nil {
				if probe2.FailureThreshold == 0 {
					probe1.FailureThreshold = probe2.FailureThreshold
				}
				if probe2.SuccessThreshold == 0 {
					probe1.SuccessThreshold = probe2.SuccessThreshold
				}
			}
			if dc2.Spec.Template.Spec.Containers[i].TerminationMessagePath == "" {
				dc1.Spec.Template.Spec.Containers[i].TerminationMessagePath = ""
			}
			if dc2.Spec.Template.Spec.Containers[i].TerminationMessagePolicy == "" {
				dc1.Spec.Template.Spec.Containers[i].TerminationMessagePolicy = ""
			}
			for j := range dc1.Spec.Template.Spec.Containers[i].Env {
				if len(dc2.Spec.Template.Spec.Containers[i].Env) <= j {
					return false
				}
				valueFrom := dc2.Spec.Template.Spec.Containers[i].Env[j].ValueFrom
				if valueFrom != nil && valueFrom.FieldRef != nil && valueFrom.FieldRef.APIVersion == "" {
					valueFrom1 := dc1.Spec.Template.Spec.Containers[i].Env[j].ValueFrom
					if valueFrom1 != nil && valueFrom1.FieldRef != nil {
						valueFrom1.FieldRef.APIVersion = ""
					}
				}
			}
		}
	}

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{dc1.Name, dc2.Name})
	pairs = append(pairs, [2]interface{}{dc1.Namespace, dc2.Namespace})
	pairs = append(pairs, [2]interface{}{dc1.Labels, dc2.Labels})
	pairs = append(pairs, [2]interface{}{dc1.Annotations, dc2.Annotations})
	pairs = append(pairs, [2]interface{}{dc1.Spec, dc2.Spec})
	return EqualPairs(pairs)
}

func equalServices(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	service1 := deployed.(*corev1.Service)
	service2 := requested.(*corev1.Service)

	//Removed generated fields from deployed version, when not specified in requested item
	service1 = service1.DeepCopy()

	//Removed potentially generated annotations for cert request
	delete(service1.Annotations, "service.alpha.openshift.io/serving-cert-signed-by")
	delete(service1.Annotations, "service.beta.openshift.io/serving-cert-signed-by")
	if service2.Spec.ClusterIP == "" {
		service1.Spec.ClusterIP = ""
	}
	if service2.Spec.Type == "" {
		service1.Spec.Type = ""
	}
	if service2.Spec.SessionAffinity == "" {
		service1.Spec.SessionAffinity = ""
	}
	for _, port2 := range service2.Spec.Ports {
		if found, port1 := findServicePort(port2, service1.Spec.Ports); found {
			if port2.Protocol == "" {
				port1.Protocol = ""
			}
		}
	}

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{service1.Name, service2.Name})
	pairs = append(pairs, [2]interface{}{service1.Namespace, service2.Namespace})
	pairs = append(pairs, [2]interface{}{service1.Labels, service2.Labels})
	pairs = append(pairs, [2]interface{}{service1.Annotations, service2.Annotations})
	pairs = append(pairs, [2]interface{}{service1.Spec, service2.Spec})
	return EqualPairs(pairs)
}

func findServicePort(port corev1.ServicePort, ports []corev1.ServicePort) (bool, *corev1.ServicePort) {
	for index, candidate := range ports {
		if port.Name == candidate.Name {
			return true, &ports[index]
		}
	}
	return false, &corev1.ServicePort{}
}

func equalRoutes(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	route1 := deployed.(*routev1.Route)
	route2 := requested.(*routev1.Route)
	route1 = route1.DeepCopy()

	//Removed generated fields from deployed version, that are not specified in requested item
	delete(route1.GetAnnotations(), "openshift.io/host.generated")
	if route2.Spec.Host == "" {
		route1.Spec.Host = ""
	}
	if route2.Spec.To.Kind == "" {
		route1.Spec.To.Kind = ""
	}
	if route2.Spec.To.Name == "" {
		route1.Spec.To.Name = ""
	}
	if route2.Spec.To.Weight == nil {
		route1.Spec.To.Weight = nil
	}
	if route2.Spec.WildcardPolicy == "" {
		route1.Spec.WildcardPolicy = ""
	}

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{route1.Name, route2.Name})
	pairs = append(pairs, [2]interface{}{route1.Namespace, route2.Namespace})
	pairs = append(pairs, [2]interface{}{route1.Labels, route2.Labels})
	pairs = append(pairs, [2]interface{}{route1.Annotations, route2.Annotations})
	pairs = append(pairs, [2]interface{}{route1.Spec, route2.Spec})
	return EqualPairs(pairs)
}

func equalRoles(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	role1 := deployed.(*rbacv1.Role)
	role2 := requested.(*rbacv1.Role)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{role1.Name, role2.Name})
	pairs = append(pairs, [2]interface{}{role1.Namespace, role2.Namespace})
	pairs = append(pairs, [2]interface{}{role1.Labels, role2.Labels})
	pairs = append(pairs, [2]interface{}{role1.Annotations, role2.Annotations})
	pairs = append(pairs, [2]interface{}{role1.Rules, role2.Rules})
	return EqualPairs(pairs)
}

func equalServiceAccounts(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	sa1 := deployed.(*corev1.ServiceAccount)
	sa2 := requested.(*corev1.ServiceAccount)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{sa1.Name, sa2.Name})
	pairs = append(pairs, [2]interface{}{sa1.Namespace, sa2.Namespace})
	pairs = append(pairs, [2]interface{}{sa1.Labels, sa2.Labels})
	pairs = append(pairs, [2]interface{}{sa1.Annotations, sa2.Annotations})
	return EqualPairs(pairs)
}

func equalRoleBindings(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	binding1 := deployed.(*rbacv1.RoleBinding)
	binding2 := requested.(*rbacv1.RoleBinding)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{binding1.Name, binding2.Name})
	pairs = append(pairs, [2]interface{}{binding1.Namespace, binding2.Namespace})
	pairs = append(pairs, [2]interface{}{binding1.Labels, binding2.Labels})
	pairs = append(pairs, [2]interface{}{binding1.Annotations, binding2.Annotations})
	pairs = append(pairs, [2]interface{}{binding1.Subjects, binding2.Subjects})
	pairs = append(pairs, [2]interface{}{binding1.RoleRef.Name, binding2.RoleRef.Name})
	return EqualPairs(pairs)
}

func equalSecrets(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	secret1 := deployed.(*corev1.Secret)
	secret2 := requested.(*corev1.Secret)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{secret1.Name, secret2.Name})
	pairs = append(pairs, [2]interface{}{secret1.Namespace, secret2.Namespace})
	pairs = append(pairs, [2]interface{}{secret1.Labels, secret2.Labels})
	pairs = append(pairs, [2]interface{}{secret1.Annotations, secret2.Annotations})
	pairs = append(pairs, [2]interface{}{secret1.Data, secret2.Data})
	pairs = append(pairs, [2]interface{}{secret1.StringData, secret2.StringData})
	return EqualPairs(pairs)
}

func deepEquals(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	struct1 := reflect.ValueOf(deployed).Elem().Type()
	if field1, found1 := struct1.FieldByName("Spec"); found1 {
		struct2 := reflect.ValueOf(requested).Elem().Type()
		if field2, found2 := struct2.FieldByName("Spec"); found2 {
			return Equals(field1, field2)
		}
	}
	return Equals(deployed, requested)
}

func EqualPairs(objects [][2]interface{}) bool {
	for index := range objects {
		if !Equals(objects[index][0], objects[index][1]) {
			return false
		}
	}
	return true
}

func Equals(deployed interface{}, requested interface{}) bool {
	equal := reflect.DeepEqual(deployed, requested)
	if !equal {
		logger.Info("Objects are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}
