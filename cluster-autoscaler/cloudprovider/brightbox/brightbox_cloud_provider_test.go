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
	"encoding/json"
	"flag"
	"strings"
	"testing"

	"github.com/brightbox/brightbox-cloud-controller-manager/k8ssdk"
	"github.com/brightbox/brightbox-cloud-controller-manager/k8ssdk/mocks"
	brightbox "github.com/brightbox/gobrightbox"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog"
)

const (
	fakeServer      = "srv-testy"
	fakeGroup       = "grp-testy"
	missingServer   = "srv-notty"
	fakeClusterName = "k8s-fake.cluster.local"
)

var (
	fakeNodeMap = map[string]string{
		fakeServer: fakeGroup,
	}
	fakeNodeGroup = &brightboxNodeGroup{
		id: fakeGroup,
	}
	fakeNodeGroups = []cloudprovider.NodeGroup{
		fakeNodeGroup,
	}
)

func init() {
	klog.InitFlags(nil)
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "4")
	flag.Parse()
}

func TestName(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider(nil).Name(), cloudprovider.BrightboxProviderName)
}

func TestGPULabel(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider(nil).GPULabel(), GPULabel)
}

func TestGetAvailableGPUTypes(t *testing.T) {
	assert.Equal(t, makeFakeCloudProvider(nil).GetAvailableGPUTypes(), availableGPUTypes)
}

func TestPricing(t *testing.T) {
	obj, err := makeFakeCloudProvider(nil).Pricing()
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func TestGetAvailableMachineTypes(t *testing.T) {
	obj, err := makeFakeCloudProvider(nil).GetAvailableMachineTypes()
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func TestNewNodeGroup(t *testing.T) {
	obj, err := makeFakeCloudProvider(nil).NewNodeGroup("", nil, nil, nil, nil)
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func TestCleanUp(t *testing.T) {
	assert.Nil(t, makeFakeCloudProvider(nil).Cleanup())
}

func TestResourceLimiter(t *testing.T) {
	client := makeFakeCloudProvider(nil)
	obj, err := client.GetResourceLimiter()
	assert.Equal(t, obj, client.resourceLimiter)
	assert.NoError(t, err)
}

func TestNodeGroups(t *testing.T) {
	client := makeFakeCloudProvider(nil)
	assert.Zero(t, client.NodeGroups())
	client.nodeGroups = make([]cloudprovider.NodeGroup, 0)
	assert.NotZero(t, client.NodeGroups())
	assert.Empty(t, client.NodeGroups())
	nodeGroup := &brightboxNodeGroup{}
	client.nodeGroups = append(client.nodeGroups, nodeGroup)
	newGroups := client.NodeGroups()
	assert.Len(t, newGroups, 1)
	assert.Same(t, newGroups[0], client.nodeGroups[0])
}

func TestNodeGroupForNode(t *testing.T) {
	client := makeFakeCloudProvider(nil)
	client.nodeGroups = fakeNodeGroups
	client.nodeMap = fakeNodeMap
	node := makeNode(fakeServer)
	nodeGroup, err := client.NodeGroupForNode(&node)
	assert.Equal(t, fakeNodeGroup, nodeGroup)
	assert.Nil(t, err)
	node = makeNode(missingServer)
	nodeGroup, err = client.NodeGroupForNode(&node)
	assert.Nil(t, nodeGroup)
	assert.Nil(t, err)
}

func TestBuildBrightBox(t *testing.T) {
	ts := k8ssdk.GetAuthEnvTokenHandler(t)
	defer k8ssdk.ResetAuthEnvironment()
	defer ts.Close()
	rl := cloudprovider.NewResourceLimiter(nil, nil)
	do := cloudprovider.NodeGroupDiscoveryOptions{}
	opts := config.AutoscalingOptions{
		CloudProviderName: cloudprovider.BrightboxProviderName,
		ClusterName:       fakeClusterName,
	}
	cloud := BuildBrightbox(opts, do, rl)
	assert.Equal(t, cloud.Name(), cloudprovider.BrightboxProviderName)
	obj, err := cloud.GetResourceLimiter()
	assert.Equal(t, rl, obj)
	assert.Nil(t, err)
}

func TestRefresh(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	provider := makeFakeCloudProvider(testclient)
	testGroups := fakeGroups()
	mockclient.On("ServerGroups").Return(testGroups, nil)
	err := provider.Refresh()
	assert.NotEmpty(t, provider.nodeGroups)
	assert.NotEmpty(t, provider.nodeMap)
	assert.Nil(t, err)
}

func makeNode(serverID string) v1.Node {
	return v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: k8ssdk.MapServerIDToProviderID(serverID),
		},
	}
}

func makeFakeCloudProvider(brightboxCloudClient *k8ssdk.Cloud) *brightboxCloudProvider {
	return &brightboxCloudProvider{
		resourceLimiter: &cloudprovider.ResourceLimiter{},
		ClusterName:     fakeClusterName,
		Cloud:           brightboxCloudClient,
	}
}

func fakeGroups() []brightbox.ServerGroup {
	const groupjson = `
[{"id": "grp-sda44",
  "resource_type": "server_group",
  "url": "https://api.gb1.brightbox.com/1.0/server_groups/grp-sda44",
  "name": "default",
  "description": "storage.k8s-fake.cluster.local",
  "created_at": "2011-10-01T00:00:00Z",
  "default": true,
  "account":
   {"id": "acc-43ks4",
    "resource_type": "account",
    "url": "https://api.gb1.brightbox.com/1.0/accounts/acc-43ks4",
    "name": "Brightbox",
    "status": "active"},
  "firewall_policy":
   {"id": "fwp-j3654",
    "resource_type": "firewall_policy",
    "url": "https://api.gb1.brightbox.com/1.0/firewall_policies/fwp-j3654",
    "default": true,
    "name": "default",
    "created_at": "2011-10-01T00:00:00Z",
    "description": null},
  "servers":
   [{"id": "srv-lv426",
     "resource_type": "server",
     "url": "https://api.gb1.brightbox.com/1.0/servers/srv-lv426",
     "name": "",
     "status": "active",
     "locked": false,
     "hostname": "srv-lv426",
     "fqdn": "srv-lv426.gb1.brightbox.com",
     "created_at": "2011-10-01T01:00:00Z",
     "started_at": "2011-10-01T01:01:00Z",
     "deleted_at": null}]}]
     `
	var result []brightbox.ServerGroup
	_ = json.NewDecoder(strings.NewReader(groupjson)).Decode(&result)
	return result
}
