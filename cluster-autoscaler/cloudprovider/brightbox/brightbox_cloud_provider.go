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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// brightboxCloudProvider implements cloudprovider.CloudProvider interface
type brightboxCloudProvider struct {
}

// Name returns name of the cloud provider.
func (b *brightboxCloudProvider) Name() string {
	panic("not implemented") // TODO: Implement
}

// NodeGroups returns all node groups configured for this cloud provider.
func (b *brightboxCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	panic("not implemented") // TODO: Implement
}

// NodeGroupForNode returns the node group for the given node, nil if
// the node should not be processed by cluster autoscaler, or non-nil
// error if such occurred. Must be implemented.
func (b *brightboxCloudProvider) NodeGroupForNode(_ *apiv1.Node) (cloudprovider.NodeGroup, error) {
	panic("not implemented") // TODO: Implement
}

// Pricing returns pricing model for this cloud provider or error if
// not available.
// Implementation optional.
func (b *brightboxCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	panic("not implemented") // TODO: Implement
}

// GetAvailableMachineTypes get all machine types that can be requested
// from the cloud provider.
// Implementation optional.
func (b *brightboxCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	panic("not implemented") // TODO: Implement
}

// NewNodeGroup builds a theoretical node group based on the node
// definition provided. The node group is not automatically created on
// the cloud provider side. The node group is not returned by NodeGroups()
// until it is created.
// Implementation optional.
func (b *brightboxCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	panic("not implemented") // TODO: Implement
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (b *brightboxCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	panic("not implemented") // TODO: Implement
}

// GPULabel returns the label added to nodes with GPU resource.
func (b *brightboxCloudProvider) GPULabel() string {
	panic("not implemented") // TODO: Implement
}

// GetAvailableGPUTypes return all available GPU types cloud provider
// supports.
func (b *brightboxCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	panic("not implemented") // TODO: Implement
}

// Cleanup cleans up open resources before the cloud provider is
// destroyed, i.e. go routines etc.
func (b *brightboxCloudProvider) Cleanup() error {
	panic("not implemented") // TODO: Implement
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state.
// In particular the list of node groups returned by NodeGroups can
// change as a result of CloudProvider.Refresh().
func (b *brightboxCloudProvider) Refresh() error {
	panic("not implemented") // TODO: Implement
}

// BuildBrightbox builds the Brightbox provider
func BuildBrightbox(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	return &brightboxCloudProvider{}
}
