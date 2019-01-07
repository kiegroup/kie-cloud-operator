package test

import (
	imagev1 "github.com/openshift/api/image/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MockImageStreamTag struct {
}

func (*MockImageStreamTag) Create(*imagev1.ImageStreamTag) (*imagev1.ImageStreamTag, error) {
	return nil, nil
}

func (*MockImageStreamTag) Update(*imagev1.ImageStreamTag) (*imagev1.ImageStreamTag, error) {
	return nil, nil
}

func (*MockImageStreamTag) Delete(name string, options *meta_v1.DeleteOptions) error {
	return nil
}

func (*MockImageStreamTag) Get(name string, options meta_v1.GetOptions) (*imagev1.ImageStreamTag, error) {
	return nil, nil
}
