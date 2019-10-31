package read

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var namespace = "ns"

func TestListObjects(t *testing.T) {
	scheme := runtime.NewScheme()
	err := corev1.SchemeBuilder.AddToScheme(scheme)
	assert.Nil(t, err, "Expect no errors building scheme")
	client := fake.NewFakeClientWithScheme(scheme)
	services := getServices(2)
	for index := range services {
		assert.Nil(t, client.Create(context.TODO(), &services[index]), "Expect no errors mock creating objects")
	}
	pods := getPods(3)
	for index := range pods {
		assert.Nil(t, client.Create(context.TODO(), &pods[index]), "Expect no errors mock creating objects")
	}

	reader := New(client).WithNamespace(namespace)
	objectMap, err := reader.ListAll(&corev1.ServiceList{}, &corev1.PodList{})
	assert.Nil(t, err, "Expect no errors listing objects")
	assert.Len(t, objectMap, 2, "Expect two object types found")

	listedServices := objectMap[reflect.TypeOf(corev1.Service{})]
	assert.Len(t, listedServices, 2, "Expect to find 2 services")
	expectedServices := getServices(2)
	assert.Equal(t, &expectedServices[0], listedServices[0])
	assert.Equal(t, &expectedServices[1], listedServices[1])

	listedPods := objectMap[reflect.TypeOf(corev1.Pod{})]
	assert.Len(t, listedPods, 3, "Expect to find 3 pods")
	expectedPods := getPods(3)
	assert.Equal(t, &expectedPods[0], listedPods[0])
	assert.Equal(t, &expectedPods[1], listedPods[1])
	assert.Equal(t, &expectedPods[2], listedPods[2])
}

func TestLoadObject(t *testing.T) {
	scheme := runtime.NewScheme()
	err := corev1.SchemeBuilder.AddToScheme(scheme)
	assert.Nil(t, err, "Expect no errors building scheme")
	client := fake.NewFakeClientWithScheme(scheme)
	service := getServices(1)[0]
	assert.Nil(t, client.Create(context.TODO(), &service), "Expect no errors mock creating object")

	reader := New(client).WithNamespace(namespace)
	found, err := reader.Load(reflect.TypeOf(service), service.Name)
	assert.Equal(t, &service, found)
}

func getServices(count int) []corev1.Service {
	services := make([]corev1.Service, count)
	for index := range services {
		services[index] = corev1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprintf("service-%d", index+1),
				Namespace: namespace,
			},
		}
	}
	return services
}

func getPods(count int) []corev1.Pod {
	pods := make([]corev1.Pod, count)
	for index := range pods {
		pods[index] = corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprintf("pod-%d", index+1),
				Namespace: namespace,
			},
		}
	}
	return pods
}
