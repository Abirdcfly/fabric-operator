/*
 * Copyright contributors to the Hyperledger Fabric Operator project
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 * 	  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Code generated by client-gen. DO NOT EDIT.

package internalversion

import (
	"context"
	"time"

	v1beta1 "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	scheme "github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// IBPOrderersGetter has a method to return a IBPOrdererInterface.
// A group's client should implement this interface.
type IBPOrderersGetter interface {
	IBPOrderers(namespace string) IBPOrdererInterface
}

// IBPOrdererInterface has methods to work with IBPOrderer resources.
type IBPOrdererInterface interface {
	Create(ctx context.Context, iBPOrderer *v1beta1.IBPOrderer, opts v1.CreateOptions) (*v1beta1.IBPOrderer, error)
	Update(ctx context.Context, iBPOrderer *v1beta1.IBPOrderer, opts v1.UpdateOptions) (*v1beta1.IBPOrderer, error)
	UpdateStatus(ctx context.Context, iBPOrderer *v1beta1.IBPOrderer, opts v1.UpdateOptions) (*v1beta1.IBPOrderer, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.IBPOrderer, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.IBPOrdererList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.IBPOrderer, err error)
	IBPOrdererExpansion
}

// iBPOrderers implements IBPOrdererInterface
type iBPOrderers struct {
	client rest.Interface
	ns     string
}

// newIBPOrderers returns a IBPOrderers
func newIBPOrderers(c *IbpClient, namespace string) *iBPOrderers {
	return &iBPOrderers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the iBPOrderer, and returns the corresponding iBPOrderer object, and an error if there is any.
func (c *iBPOrderers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.IBPOrderer, err error) {
	result = &v1beta1.IBPOrderer{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ibporderers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of IBPOrderers that match those selectors.
func (c *iBPOrderers) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.IBPOrdererList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.IBPOrdererList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ibporderers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested iBPOrderers.
func (c *iBPOrderers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("ibporderers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a iBPOrderer and creates it.  Returns the server's representation of the iBPOrderer, and an error, if there is any.
func (c *iBPOrderers) Create(ctx context.Context, iBPOrderer *v1beta1.IBPOrderer, opts v1.CreateOptions) (result *v1beta1.IBPOrderer, err error) {
	result = &v1beta1.IBPOrderer{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("ibporderers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(iBPOrderer).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a iBPOrderer and updates it. Returns the server's representation of the iBPOrderer, and an error, if there is any.
func (c *iBPOrderers) Update(ctx context.Context, iBPOrderer *v1beta1.IBPOrderer, opts v1.UpdateOptions) (result *v1beta1.IBPOrderer, err error) {
	result = &v1beta1.IBPOrderer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ibporderers").
		Name(iBPOrderer.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(iBPOrderer).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *iBPOrderers) UpdateStatus(ctx context.Context, iBPOrderer *v1beta1.IBPOrderer, opts v1.UpdateOptions) (result *v1beta1.IBPOrderer, err error) {
	result = &v1beta1.IBPOrderer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ibporderers").
		Name(iBPOrderer.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(iBPOrderer).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the iBPOrderer and deletes it. Returns an error if one occurs.
func (c *iBPOrderers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ibporderers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *iBPOrderers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ibporderers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched iBPOrderer.
func (c *iBPOrderers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.IBPOrderer, err error) {
	result = &v1beta1.IBPOrderer{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("ibporderers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
