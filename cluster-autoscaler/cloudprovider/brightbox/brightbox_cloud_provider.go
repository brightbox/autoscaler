/*
Copyright 2020 The Kubernetes Authors.

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
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/brightbox/gobrightbox/status"
	"github.com/brightbox/k8ssdk"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	klog "k8s.io/klog/v2"
)

const (
	// GPULabel is added to nodes with GPU resource
	GPULabel          = "cloud.brightbox.com/gpu-node"
	joinCommandEnvVar = "BRIGHTBOX_KUBE_JOIN_COMMAND"
	k8sVersionEnvVar  = "BRIGHTBOX_KUBE_VERSION"
)

var (
	availableGPUTypes = map[string]struct{}{}
	k8sEnvVars        = []string{
		joinCommandEnvVar,
		k8sVersionEnvVar,
	}
)

// brightboxCloudProvider implements cloudprovider.CloudProvider interface
type brightboxCloudProvider struct {
	resourceLimiter *cloudprovider.ResourceLimiter
	ClusterName     string
	nodeGroups      []cloudprovider.NodeGroup
	nodeMap         map[string]string
	*k8ssdk.Cloud
}

// Name returns name of the cloud provider.
func (b *brightboxCloudProvider) Name() string {
	klog.V(4).Info("Name")
	return cloudprovider.BrightboxProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (b *brightboxCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	klog.V(4).Info("NodeGroups")
	// Duplicate the stored nodegroup elements and return it
	//return append(b.nodeGroups[:0:0], b.nodeGroups...)
	// Or just return the stored nodegroup elements by reference
	return b.nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if
// the node should not be processed by cluster autoscaler, or non-nil
// error if such occurred. Must be implemented.
func (b *brightboxCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("NodeGroupForNode")
	klog.V(4).Infof("Looking for %v", node.Spec.ProviderID)
	groupID, ok := b.nodeMap[k8ssdk.MapProviderIDToServerID(node.Spec.ProviderID)]
	if ok {
		klog.V(4).Infof("Found in group %v", groupID)
		return b.findNodeGroup(groupID), nil
	}
	klog.V(4).Info("Not found")
	return nil, nil
}

// Refresh is before every main loop and can be used to dynamically
// update cloud provider state.
// In particular the list of node groups returned by NodeGroups can
// change as a result of CloudProvider.Refresh().
func (b *brightboxCloudProvider) Refresh() error {
	klog.V(4).Info("Refresh")
	groups, err := b.GetServerGroups()
	if err != nil {
		return err
	}
	clusterSuffix := "." + b.ClusterName
	nodeGroups := make([]cloudprovider.NodeGroup, 0)
	nodeMap := make(map[string]string)
	defaultGroup := fetchDefaultGroup(groups, b.ClusterName)
	for _, group := range groups {
		maxSize, minSize := getSizesFromDescription(group.Description)
		switch {
		case !strings.HasSuffix(group.Name, clusterSuffix):
			klog.V(4).Infof("Group %q: name %q doesn't match suffix %q. Ignoring", group.Id, group.Name, clusterSuffix)
		case len(group.Servers) < 1:
			klog.V(4).Infof("Group %q has no members. Ignoring", group.Id)
		case maxSize == minSize:
			klog.V(4).Infof("Group %q has a fixed size. Ignoring", group.Id)
		default:
			klog.V(4).Infof("Group %q: looking for node defaults", group.Id)
			groupType, groupImage, groupZone, err :=
				b.extractGroupDefaults(group.Servers)
			if err != nil {
				klog.V(4).Infof("Group %q: failed to find node defaults", group.Id)
				return err
			}
			klog.V(4).Infof("Group %q: Node defaults found. Adding to node group list", group.Id)
			newNodeGroup := makeNodeGroupFromAPIDetails(
				group.Id,
				defaultServerName(group.Name),
				group.Description,
				len(group.Servers),
				groupType,
				groupImage,
				groupZone,
				defaultGroup,
				defaultUserData(
					strings.TrimSpace(os.Getenv(k8sVersionEnvVar)),
					strings.TrimSpace(os.Getenv(joinCommandEnvVar)),
				),
				b.Cloud,
			)
			for _, server := range group.Servers {
				nodeMap[server.Id] = group.Id
			}
			nodeGroups = append(nodeGroups, newNodeGroup)
		}
	}
	b.nodeGroups = nodeGroups
	b.nodeMap = nodeMap
	klog.V(4).Infof("Refresh located %v node(s) over %v group(s)", len(nodeMap), len(nodeGroups))
	return nil
}

// Pricing returns pricing model for this cloud provider or error if
// not available.
// Implementation optional.
func (b *brightboxCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	klog.V(4).Info("Pricing")
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested
// from the cloud provider.
// Implementation optional.
func (b *brightboxCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	klog.V(4).Info("GetAvailableMachineTypes")
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node
// definition provided. The node group is not automatically created on
// the cloud provider side. The node group is not returned by NodeGroups()
// until it is created.
// Implementation optional.
func (b *brightboxCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("newNodeGroup")
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (b *brightboxCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	klog.V(4).Info("GetResourceLimiter")
	return b.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (b *brightboxCloudProvider) GPULabel() string {
	klog.V(4).Info("GPULabel")
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider
// supports.
func (b *brightboxCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	klog.V(4).Info("GetAvailableGPUTypes")
	return availableGPUTypes
}

// Cleanup cleans up open resources before the cloud provider is
// destroyed, i.e. go routines etc.
func (b *brightboxCloudProvider) Cleanup() error {
	klog.V(4).Info("Cleanup")
	return nil
}

// BuildBrightbox builds the Brightbox provider
func BuildBrightbox(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	klog.V(4).Info("BuildBrightbox")
	klog.V(4).Infof("Config: %+v", opts)
	klog.V(4).Infof("Discovery Options: %+v", do)
	if opts.CloudConfig != "" {
		klog.Warning("supplied config is not read by this version. Using environment")
	}
	if opts.ClusterName == "" {
		klog.Fatal("Set the cluster name option to the Fully Qualified Internal Domain Name of the cluster")
	}
	validateEnvironment()
	validateJoinCommand()
	newCloudProvider := &brightboxCloudProvider{
		ClusterName:     opts.ClusterName,
		resourceLimiter: rl,
		Cloud:           &k8ssdk.Cloud{},
	}
	_, err := newCloudProvider.CloudClient()
	if err != nil {
		klog.Fatalf("Failed to create Brightbox Cloud Client: %v", err)
	}
	return newCloudProvider
}

//private

func validateEnvironment() {
	klog.V(4).Info("validateEnvironment")
	for _, envVar := range k8sEnvVars {
		if _, ok := os.LookupEnv(envVar); !ok {
			klog.Fatalf("Required Environment Variable %q not set", envVar)
		}
	}
}

func validateJoinCommand() {
	klog.V(4).Info("validateJoinCommand")
	const sanitiseKubeadmCommand string = `^kubeadm +join +[0-9\.]+:[0-9]+ +--token +[a-z0-9]{6}\.[a-z0-9]{16} +--discovery-token-ca-cert-hash +sha256:[0-9a-f]+$`
	re := regexp.MustCompile(sanitiseKubeadmCommand)
	command := strings.TrimSpace(os.Getenv(joinCommandEnvVar))
	if !re.MatchString(command) {
		klog.Infof("Join Command: %q", command)
		klog.Fatalf("Join Command does not match sanitisation pattern")
	}
}

func (b *brightboxCloudProvider) findNodeGroup(groupID string) cloudprovider.NodeGroup {
	klog.V(4).Info("findNodeGroup")
	klog.V(4).Infof("Looking for %q", groupID)
	for _, nodeGroup := range b.nodeGroups {
		if nodeGroup.Id() == groupID {
			return nodeGroup
		}
	}
	return nil
}

func defaultServerName(name string) string {
	klog.V(4).Info("defaultServerName")
	klog.V(4).Infof("group name is %q", name)
	return "auto." + name
}

func fetchDefaultGroup(groups []brightbox.ServerGroup, clusterName string) string {
	klog.V(4).Info("findDefaultGroup")
	klog.V(4).Infof("for cluster %q", clusterName)
	for _, group := range groups {
		if group.Name == clusterName {
			return group.Id
		}
	}
	klog.Warningf("Unable to detect main group for cluster %q", clusterName)
	return ""
}

type idWithStatus struct {
	id     string
	status string
}

func (b *brightboxCloudProvider) extractGroupDefaults(servers []brightbox.Server) (string, string, string, error) {
	klog.V(4).Info("extractGroupDefaults")
	const zoneSentinel string = "dummyValue"
	zoneID := zoneSentinel
	var serverType, image idWithStatus
	for _, serverSummary := range servers {
		server, err := b.GetServer(
			context.Background(),
			serverSummary.Id,
			serverNotFoundError(serverSummary.Id),
		)
		if err != nil {
			return "", "", "", err
		}
		image = checkForChange(image, idWithStatus{server.Image.Id, server.Image.Status}, "Group has multiple Image Ids")
		serverType = checkForChange(serverType, idWithStatus{server.ServerType.Id, server.ServerType.Status}, "Group has multiple ServerType Ids")
		zoneID = checkZoneForChange(zoneID, server.Zone.Id, zoneSentinel)
	}
	switch {
	case serverType.id == "":
		return "", "", "", fmt.Errorf("Unable to determine Server Type details from Group")
	case image.id == "":
		return "", "", "", fmt.Errorf("Unable to determine Image details from Group")
	case zoneID == zoneSentinel:
		return "", "", "", fmt.Errorf("Unable to determine Zone details from Group")
	case image.status == status.Deprecated:
		klog.Warningf("Selected image %q is deprecated. Please update to an available version", image.id)
	}
	return serverType.id, image.id, zoneID, nil
}

func checkZoneForChange(zoneID string, newZoneID string, sentinel string) string {
	klog.V(4).Info("checkZoneForChange")
	klog.V(4).Infof("new %q, existing %q", newZoneID, zoneID)
	switch zoneID {
	case newZoneID, sentinel:
		return newZoneID
	default:
		klog.V(4).Info("Group is zone balanced")
		return ""
	}
}

func checkForChange(current idWithStatus, newDetails idWithStatus, errorMessage string) idWithStatus {
	klog.V(4).Info("checkForChange")
	klog.V(4).Infof("new %v, existing %v", newDetails, current)
	switch {
	case newDetails == current:
		// Skip to end
	case newDetails.status == status.Available:
		if current.id == "" || current.status == status.Deprecated {
			klog.V(4).Infof("Object %q is available. Selecting", newDetails.id)
			return newDetails
		}
		// Multiple ids
		klog.Warning(errorMessage)
	case newDetails.status == status.Deprecated:
		if current.id == "" {
			klog.V(4).Infof("Object %q is deprecated, but selecting anyway", newDetails.id)
			return newDetails
		}
		// Multiple ids
		klog.Warning(errorMessage)
	default:
		klog.Warningf("Object %q is no longer available. Ignoring.", newDetails.id)
	}
	return current
}
