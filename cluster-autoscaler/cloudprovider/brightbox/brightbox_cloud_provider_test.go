/*
Copyright 2019 The Kubernetes Authors.

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

package brightbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func TestName(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider().Name(), cloudprovider.BrightboxProviderName)
}

func TestGPULabel(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider().GPULabel(), GPULabel)
}

func TestGetAvailableGPUTypes(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider().GetAvailableGPUTypes(), availableGPUTypes)
}

func TestPricing(t *testing.T) {
	obj, err := makeFakeCloudProvider().Pricing()
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func makeFakeCloudProvider() *brightboxCloudProvider {
	return &brightboxCloudProvider{}
}
