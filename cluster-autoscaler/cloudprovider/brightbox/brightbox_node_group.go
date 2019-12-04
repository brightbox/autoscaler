package brightbox

import (
	"fmt"
	"strconv"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

type brightboxNodeGroup struct {
	id           string
	minSize      int
	maxSize      int
	serverTypeId string
	imageId      string
	zoneId       string
}

// MaxSize returns maximum size of the node group.
func (ng *brightboxNodeGroup) MaxSize() int {
	return ng.maxSize
}

// MinSize returns minimum size of the node group.
func (ng *brightboxNodeGroup) MinSize() int {
	return ng.minSize
}

// TargetSize returns the current target size of the node group. It
// is possible that the number of nodes in Kubernetes is different at
// the moment but should be equal to Size() once everything stabilizes
// (new nodes finish startup and registration or removed nodes are deleted
// completely). Implementation required.
func (ng *brightboxNodeGroup) TargetSize() (int, error) {
	panic("not implemented") // TODO: Implement
}

// IncreaseSize increases the size of the node group. To delete a node
// you need to explicitly name it and use DeleteNode. This function should
// wait until node group size is updated. Implementation required.
func (ng *brightboxNodeGroup) IncreaseSize(delta int) error {
	panic("not implemented") // TODO: Implement
}

// DeleteNodes deletes nodes from this node group. Error is returned
// either on failure or if the given node doesn't belong to this
// node group. This function should wait until node group size is
// updated. Implementation required.
func (ng *brightboxNodeGroup) DeleteNodes(_ []*apiv1.Node) error {
	panic("not implemented") // TODO: Implement
}

// DecreaseTargetSize decreases the target size of the node group. This
// function doesn't permit to delete any existing node and can be used
// only to reduce the request for new nodes that have not been yet
// fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes
// when there is an option to just decrease the target. Implementation
// required.
func (ng *brightboxNodeGroup) DecreaseTargetSize(delta int) error {
	panic("not implemented") // TODO: Implement
}

// Id returns an unique identifier of the node group.
func (ng *brightboxNodeGroup) Id() string {
	return ng.id
}

// Debug returns a string containing all information regarding this
// node group.
func (ng *brightboxNodeGroup) Debug() string {
	return fmt.Sprintf("brightboxNodeGroup %+v", *ng)
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id
// field set.  Other fields are optional.
func (ng *brightboxNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	panic("not implemented") // TODO: Implement
}

// Exist checks if the node group really exists on the cloud provider
// side. Allows to tell the theoretical node group from the real
// one. Implementation required.
func (ng *brightboxNodeGroup) Exist() bool {
	panic("not implemented") // TODO: Implement
}

// TemplateNodeInfo returns a schedulernodeinfo.NodeInfo structure
// of an empty (as if just started) node. This will be used in scale-up
// simulations to predict what would a new node look like if a node group was
// expanded. The returned NodeInfo is expected to have a fully populated Node
// object, with all of the labels, capacity and allocatable information as
// well as all pods that are started on the node by default, using manifest
// (most likely only kube-proxy). Implementation optional.
func (ng *brightboxNodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Create creates the node group on the cloud provider
// side. Implementation optional.
func (ng *brightboxNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once
// their size drops to 0.  Implementation optional.
func (ng *brightboxNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled
// to 0.
func (ng *brightboxNodeGroup) Autoprovisioned() bool {
	return false
}

func makeNodeGroupFromApiDetails(
	id string,
	description string,
	defaultSize int,
	defaultServerTypeId string,
	defaultImageId string,
	defaultZoneId string,
) *brightboxNodeGroup {
	klog.V(4).Info("makeNodeGroupFromApiDetails")
	klog.V(4).Infof("Group: %s, Description: %s", id, description)
	klog.V(4).Infof("Default size: %v", defaultSize)
	result := brightboxNodeGroup{
		id:           id,
		minSize:      defaultSize,
		maxSize:      defaultSize,
		serverTypeId: defaultServerTypeId,
		imageId:      defaultImageId,
		zoneId:       defaultZoneId,
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
