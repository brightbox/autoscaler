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
	"strconv"
	"strings"
	"time"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/brightbox/k8ssdk"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

var (
	checkInterval = time.Second * 1
	checkTimeout  = time.Second * 30
)

type brightboxNodeGroup struct {
	id            string
	minSize       int
	maxSize       int
	serverOptions *brightbox.ServerOptions
	*k8ssdk.Cloud
}

// MaxSize returns maximum size of the node group.
func (ng *brightboxNodeGroup) MaxSize() int {
	klog.V(4).Info("MaxSize")
	return ng.maxSize
}

// MinSize returns minimum size of the node group.
func (ng *brightboxNodeGroup) MinSize() int {
	klog.V(4).Info("MinSize")
	return ng.minSize
}

// TargetSize returns the current target size of the node group. It
// is possible that the number of nodes in Kubernetes is different at
// the moment but should be equal to Size() once everything stabilizes
// (new nodes finish startup and registration or removed nodes are deleted
// completely). Implementation required.
func (ng *brightboxNodeGroup) TargetSize() (int, error) {
	klog.V(4).Info("TargetSize")
	group, err := ng.GetServerGroup(ng.Id())
	if err != nil {
		return 0, err
	}
	return len(group.Servers), nil
}

// CurrentSize returns the current actual size of the node group.
func (ng *brightboxNodeGroup) CurrentSize() (int, error) {
	klog.V(4).Info("CurrentSize")
	// The implementation is currently synchronous, so
	// CurrentSize and TargetSize will be identical at all times
	return ng.TargetSize()
}

// IncreaseSize increases the size of the node group. To delete a node
// you need to explicitly name it and use DeleteNode. This function should
// wait until node group size is updated. Implementation required.
func (ng *brightboxNodeGroup) IncreaseSize(delta int) error {
	klog.V(4).Infof("IncreaseSize: %v", delta)
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := ng.TargetSize()
	if err != nil {
		return err
	}
	desiredSize := size + delta
	if desiredSize > ng.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", desiredSize, ng.MaxSize())
	}
	err = ng.createServers(delta)
	if err != nil {
		return err
	}
	return wait.Poll(
		checkInterval,
		checkTimeout,
		func() (bool, error) {
			size, err := ng.TargetSize()
			return err == nil && size >= desiredSize, err
		},
	)
}

// DeleteNodes deletes nodes from this node group. Error is returned
// either on failure or if the given node doesn't belong to this
// node group. This function should wait until node group size is
// updated. Implementation required.
func (ng *brightboxNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(4).Info("DeleteNodes")
	klog.V(4).Infof("Nodes: %+v", nodes)
	for _, node := range nodes {
		size, err := ng.CurrentSize()
		if err != nil {
			return err
		}
		if size <= ng.MinSize() {
			return fmt.Errorf("min size reached, no further nodes will be deleted")
		}
		serverId := k8ssdk.MapProviderIDToServerID(node.Spec.ProviderID)
		err = ng.deleteServerFromGroup(serverId)
		if err != nil {
			return err
		}
	}
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This
// function doesn't permit to delete any existing node and can be used
// only to reduce the request for new nodes that have not been yet
// fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes
// when there is an option to just decrease the target. Implementation
// required.
func (ng *brightboxNodeGroup) DecreaseTargetSize(delta int) error {
	klog.V(4).Infof("DecreaseTargetSize: %v", delta)
	if delta >= 0 {
		return fmt.Errorf("decrease size must be negative")
	}
	size, err := ng.TargetSize()
	if err != nil {
		return err
	}
	nodesize, err := ng.CurrentSize()
	if err != nil {
		return err
	}
	// Group size is synchronous at present, so this always fails
	if size+delta < nodesize {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, nodesize)
	}
	return fmt.Errorf("Shouldn't have got here!")
}

// Id returns an unique identifier of the node group.
func (ng *brightboxNodeGroup) Id() string {
	klog.V(4).Info("Id")
	return ng.id
}

// Debug returns a string containing all information regarding this
// node group.
func (ng *brightboxNodeGroup) Debug() string {
	klog.V(4).Info("Debug")
	return fmt.Sprintf("brightboxNodeGroup %+v", *ng)
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id
// field set.  Other fields are optional.
func (ng *brightboxNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(4).Info("Nodes")
	group, err := ng.GetServerGroup(ng.Id())
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Found %d servers in group", len(group.Servers))
	nodes := make([]cloudprovider.Instance, len(group.Servers))
	for i, server := range group.Servers {
		status := cloudprovider.InstanceStatus{}
		switch server.Status {
		case "active":
			status.State = cloudprovider.InstanceRunning
		case "creating":
			status.State = cloudprovider.InstanceCreating
		case "deleting":
			status.State = cloudprovider.InstanceDeleting
		default:
			errorInfo := cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    server.Status,
				ErrorMessage: server.Status,
			}
			status.ErrorInfo = &errorInfo
		}
		nodes[i] = cloudprovider.Instance{
			Id:     server.Id,
			Status: &status,
		}
	}
	klog.V(4).Infof("Created %d nodes", len(nodes))
	return nodes, nil
}

// Exist checks if the node group really exists on the cloud provider
// side. Allows to tell the theoretical node group from the real
// one. Implementation required.
func (ng *brightboxNodeGroup) Exist() bool {
	klog.V(4).Info("Exist")
	_, err := ng.GetServerGroup(ng.Id())
	return err == nil
}

// TemplateNodeInfo returns a schedulernodeinfo.NodeInfo structure
// of an empty (as if just started) node. This will be used in scale-up
// simulations to predict what would a new node look like if a node group was
// expanded. The returned NodeInfo is expected to have a fully populated Node
// object, with all of the labels, capacity and allocatable information as
// well as all pods that are started on the node by default, using manifest
// (most likely only kube-proxy). Implementation optional.
func (ng *brightboxNodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	klog.V(4).Info("TemplateNodeInfo")
	return nil, cloudprovider.ErrNotImplemented
}

// Create creates the node group on the cloud provider
// side. Implementation optional.
func (ng *brightboxNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("Create")
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once
// their size drops to 0.  Implementation optional.
func (ng *brightboxNodeGroup) Delete() error {
	klog.V(4).Info("Delete")
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled
// to 0.
func (ng *brightboxNodeGroup) Autoprovisioned() bool {
	klog.V(4).Info("Autoprovisioned")
	return false
}

//private

func makeNodeGroupFromApiDetails(
	id string,
	name string,
	description string,
	defaultSize int,
	defaultServerTypeId string,
	defaultImageId string,
	defaultZoneId string,
	defaultMainGroupId string,
	defaultUserData string,
	cloudclient *k8ssdk.Cloud,
) *brightboxNodeGroup {
	klog.V(4).Info("makeNodeGroupFromApiDetails")
	klog.V(4).Infof("Group: %s, Description: %s", id, description)
	klog.V(4).Infof("Default size: %v", defaultSize)
	options := &brightbox.ServerOptions{
		Image:        defaultImageId,
		Name:         &name,
		ServerType:   defaultServerTypeId,
		Zone:         defaultZoneId,
		UserData:     &defaultUserData,
		ServerGroups: []string{defaultMainGroupId, id},
	}
	result := brightboxNodeGroup{
		id:            id,
		minSize:       defaultSize,
		maxSize:       defaultSize,
		serverOptions: options,
		Cloud:         cloudclient,
	}
	sizes := strings.Split(description, ":")
	if len(sizes) == 2 {
		value, err := strconv.Atoi(sizes[0])
		if err == nil {
			result.minSize = value
		}
		value, err = strconv.Atoi(sizes[1])
		if err == nil {
			result.maxSize = value
		}
	}
	klog.V(4).Info(result.Debug())
	return &result
}

func (ng *brightboxNodeGroup) createServers(amount int) error {
	klog.V(4).Infof("createServers: %d", amount)
	for i := 1; i <= amount; i++ {
		_, err := ng.CreateServer(ng.serverOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete the server and wait for the group details to be updated
func (ng *brightboxNodeGroup) deleteServerFromGroup(serverId string) error {
	klog.V(4).Infof("deleteServerFromGroup: %q", serverId)
	serverIdNotInGroup := func() (bool, error) {
		return ng.isMissing(serverId)
	}
	missing, err := serverIdNotInGroup()
	if err != nil {
		return err
	} else if missing {
		return fmt.Errorf("%s belongs to a different group than %s", serverId, ng.Id())
	}
	err = ng.DestroyServer(serverId)
	if err != nil {
		return err
	}
	return wait.Poll(
		checkInterval,
		checkTimeout,
		serverIdNotInGroup,
	)
}

func serverNotFoundError(id string) error {
	klog.V(4).Infof("serverNotFoundError: %q", id)
	return fmt.Errorf("Server %s not found", id)
}

func (ng *brightboxNodeGroup) isMissing(serverId string) (bool, error) {
	klog.V(4).Infof("isMissing: %q from %q", serverId, ng.Id())
	server, err := ng.GetServer(
		context.Background(),
		serverId,
		serverNotFoundError(serverId),
	)
	if err != nil {
		return false, err
	}
	if server.DeletedAt != nil {
		klog.V(4).Info("server deleted")
		return true, nil
	}
	for _, group := range server.ServerGroups {
		if group.Id == ng.Id() {
			return false, nil
		}
	}
	return true, nil
}
