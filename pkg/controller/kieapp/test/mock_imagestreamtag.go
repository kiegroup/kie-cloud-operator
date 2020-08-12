package test

import (
	"context"
	"fmt"

	imagev1 "github.com/openshift/api/image/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MockImageStreamTag struct {
	Tags map[string]*imagev1.ImageStreamTag
}

func (mock *MockImageStreamTag) Create(ctx context.Context, tag *imagev1.ImageStreamTag, options meta_v1.CreateOptions) (*imagev1.ImageStreamTag, error) {
	if mock.Tags == nil {
		mock.Tags = make(map[string]*imagev1.ImageStreamTag)
	}
	name := fmt.Sprintf("%s/%s", tag.ObjectMeta.Namespace, tag.ObjectMeta.Name)
	mock.Tags[name] = tag
	return tag, nil
}

func (mock *MockImageStreamTag) Update(ctx context.Context, tag *imagev1.ImageStreamTag, options meta_v1.UpdateOptions) (*imagev1.ImageStreamTag, error) {
	if mock.Tags == nil {
		mock.Tags = make(map[string]*imagev1.ImageStreamTag)
	}
	name := fmt.Sprintf("%s/%s", tag.ObjectMeta.Namespace, tag.ObjectMeta.Name)
	old := mock.Tags[name]
	mock.Tags[name] = tag
	return old, nil
}

func (mock *MockImageStreamTag) Delete(ctx context.Context, name string, options meta_v1.DeleteOptions) error {
	if mock.Tags == nil {
		return nil
	}
	delete(mock.Tags, name)
	return nil
}

func (mock *MockImageStreamTag) Get(ctx context.Context, name string, options meta_v1.GetOptions) (*imagev1.ImageStreamTag, error) {
	if mock.Tags == nil {
		return nil, nil
	}
	return mock.Tags[name], nil
}

func (mock *MockImageStreamTag) List(ctx context.Context, opts meta_v1.ListOptions) (*imagev1.ImageStreamTagList, error) {
	if mock.Tags == nil {
		return nil, nil
	}
	items := make([]imagev1.ImageStreamTag, 0, len(mock.Tags))
	for _, val := range mock.Tags {
		items = append(items, *val)
	}
	list := &imagev1.ImageStreamTagList{
		Items: items,
	}
	return list, nil
}
