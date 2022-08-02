/*
Copyright 2021 The Skupper Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/skupperproject/skupper/pkg/apis/skupper/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeLinkConfigs implements LinkConfigInterface
type FakeLinkConfigs struct {
	Fake *FakeSkupperV1alpha1
	ns   string
}

var linkconfigsResource = schema.GroupVersionResource{Group: "skupper.io", Version: "v1alpha1", Resource: "linkconfigs"}

var linkconfigsKind = schema.GroupVersionKind{Group: "skupper.io", Version: "v1alpha1", Kind: "LinkConfig"}

// Get takes name of the linkConfig, and returns the corresponding linkConfig object, and an error if there is any.
func (c *FakeLinkConfigs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.LinkConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(linkconfigsResource, c.ns, name), &v1alpha1.LinkConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LinkConfig), err
}

// List takes label and field selectors, and returns the list of LinkConfigs that match those selectors.
func (c *FakeLinkConfigs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.LinkConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(linkconfigsResource, linkconfigsKind, c.ns, opts), &v1alpha1.LinkConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.LinkConfigList{ListMeta: obj.(*v1alpha1.LinkConfigList).ListMeta}
	for _, item := range obj.(*v1alpha1.LinkConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested linkConfigs.
func (c *FakeLinkConfigs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(linkconfigsResource, c.ns, opts))

}

// Create takes the representation of a linkConfig and creates it.  Returns the server's representation of the linkConfig, and an error, if there is any.
func (c *FakeLinkConfigs) Create(ctx context.Context, linkConfig *v1alpha1.LinkConfig, opts v1.CreateOptions) (result *v1alpha1.LinkConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(linkconfigsResource, c.ns, linkConfig), &v1alpha1.LinkConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LinkConfig), err
}

// Update takes the representation of a linkConfig and updates it. Returns the server's representation of the linkConfig, and an error, if there is any.
func (c *FakeLinkConfigs) Update(ctx context.Context, linkConfig *v1alpha1.LinkConfig, opts v1.UpdateOptions) (result *v1alpha1.LinkConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(linkconfigsResource, c.ns, linkConfig), &v1alpha1.LinkConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LinkConfig), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeLinkConfigs) UpdateStatus(ctx context.Context, linkConfig *v1alpha1.LinkConfig, opts v1.UpdateOptions) (*v1alpha1.LinkConfig, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(linkconfigsResource, "status", c.ns, linkConfig), &v1alpha1.LinkConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LinkConfig), err
}

// Delete takes name of the linkConfig and deletes it. Returns an error if one occurs.
func (c *FakeLinkConfigs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(linkconfigsResource, c.ns, name), &v1alpha1.LinkConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeLinkConfigs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(linkconfigsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.LinkConfigList{})
	return err
}

// Patch applies the patch and returns the patched linkConfig.
func (c *FakeLinkConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.LinkConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(linkconfigsResource, c.ns, name, pt, data, subresources...), &v1alpha1.LinkConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LinkConfig), err
}
