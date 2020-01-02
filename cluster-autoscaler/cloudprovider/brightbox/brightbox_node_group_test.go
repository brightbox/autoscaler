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
	"errors"
	"testing"

	"github.com/brightbox/k8ssdk"
	"github.com/brightbox/k8ssdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

const (
	fakeMaxSize               = 4
	fakeMinSize               = 1
	fakeNodeGroupDescription  = "1:4"
	fakeDefaultSize           = 3
	fakeNodeGroupId           = "grp-sda44"
	fakeNodeGroupName         = "auto.workers.k8s_fake.cluster.local"
	fakeNodeGroupImageId      = "img-testy"
	fakeNodeGroupServerTypeId = "typ-testy"
	fakeNodeGroupZoneId       = "zon-testy"
	fakeNodeGroupMainGroupId  = "grp-y6cai"
	fakeNodeGroupUserData     = "fake userdata"
)

var (
	fakeError     = errors.New("Fake API Error")
	fakeInstances = []cloudprovider.Instance{
		cloudprovider.Instance{
			Id: "srv-rp897",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		cloudprovider.Instance{
			Id: "srv-lv426",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
	}
	fakeTransitionInstances = []cloudprovider.Instance{
		cloudprovider.Instance{
			Id: "srv-rp897",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		cloudprovider.Instance{
			Id: "srv-lv426",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
	}
	fakeErrorInstances = []cloudprovider.Instance{
		cloudprovider.Instance{
			Id: "srv-rp897",
			Status: &cloudprovider.InstanceStatus{
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OtherErrorClass,
					ErrorCode:    "unavailable",
					ErrorMessage: "unavailable",
				},
			},
		},
		cloudprovider.Instance{
			Id: "srv-lv426",
			Status: &cloudprovider.InstanceStatus{
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OtherErrorClass,
					ErrorCode:    "inactive",
					ErrorMessage: "inactive",
				},
			},
		},
	}
)

func TestMaxSize(t *testing.T) {
	assert.Equal(t, makeFakeNodeGroup(nil).MaxSize(), fakeMaxSize)
}

func TestMinSize(t *testing.T) {
	assert.Equal(t, makeFakeNodeGroup(nil).MinSize(), fakeMinSize)
}

func TestSize(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(testclient)
	fakeServerGroup := &fakeGroups()[0]
	t.Run("TargetSize", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(fakeServerGroup, nil).Once()
		size, err := nodeGroup.TargetSize()
		assert.Equal(t, 2, size)
		assert.NoError(t, err)
	})
	t.Run("TargetSizeFail", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(nil, fakeError).Once()
		size, err := nodeGroup.TargetSize()
		assert.Error(t, err)
		assert.Zero(t, size)
	})
	t.Run("CurrentSize", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(fakeServerGroup, nil).Once()
		size, err := nodeGroup.CurrentSize()
		assert.Equal(t, 2, size)
		assert.NoError(t, err)
	})
	t.Run("CurrentSizeFail", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(nil, fakeError).Once()
		size, err := nodeGroup.CurrentSize()
		assert.Error(t, err)
		assert.Zero(t, size)
	})
	mockclient.On("ServerGroup", fakeNodeGroupId).
		Return(fakeServerGroup, nil)
	t.Run("DecreaseTargetSizePositive", func(t *testing.T) {
		err := nodeGroup.DecreaseTargetSize(0)
		assert.Error(t, err)
	})
	t.Run("DecreaseTargetSizeFail", func(t *testing.T) {
		err := nodeGroup.DecreaseTargetSize(-1)
		assert.Error(t, err)
	})
	mockclient.AssertExpectations(t)
}

func TestIncreaseSize(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(testclient)
	t.Run("Creating details set properly", func(t *testing.T) {
		assert.Equal(t, fakeNodeGroupId, nodeGroup.id)
		assert.Equal(t, fakeNodeGroupName, *nodeGroup.serverOptions.Name)
		assert.Equal(t, fakeNodeGroupServerTypeId, nodeGroup.serverOptions.ServerType)
		assert.Equal(t, fakeNodeGroupImageId, nodeGroup.serverOptions.Image)
		assert.Equal(t, fakeNodeGroupZoneId, nodeGroup.serverOptions.Zone)
		assert.ElementsMatch(t, []string{fakeNodeGroupMainGroupId, fakeNodeGroupId}, nodeGroup.serverOptions.ServerGroups)
		assert.Equal(t, fakeNodeGroupUserData, *nodeGroup.serverOptions.UserData)
	})
	t.Run("Require positive delta", func(t *testing.T) {
		err := nodeGroup.IncreaseSize(0)
		assert.Error(t, err)
	})
	fakeServerGroup := &fakeGroups()[0]
	t.Run("Don't exceed max size", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(fakeServerGroup, nil).Once()
		err := nodeGroup.IncreaseSize(4)
		assert.Error(t, err)
	})
	t.Run("Fail to create one new server", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(fakeServerGroup, nil).Once()
		mockclient.On("CreateServer", mock.Anything).
			Return(nil, fakeError).Once()
		err := nodeGroup.IncreaseSize(1)
		assert.Error(t, err)
	})
	t.Run("Create one new server", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(fakeServerGroup, nil).Once()
		mockclient.On("CreateServer", mock.Anything).
			Return(nil, nil).Once()
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(&fakeServerGroupsPlusOne()[0], nil).Once()
		err := nodeGroup.IncreaseSize(1)
		assert.NoError(t, err)
	})
}

func TestDeleteNodes(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(testclient)
	fakeServerGroup := &fakeGroups()[0]
	mockclient.On("ServerGroup", fakeNodeGroupId).
		Return(fakeServerGroup, nil).
		On("Server", fakeServer).
		Return(fakeServertesty(), nil)
	t.Run("Empty Nodes", func(t *testing.T) {
		err := nodeGroup.DeleteNodes(nil)
		assert.NoError(t, err)
	})
	t.Run("Foreign Node", func(t *testing.T) {
		err := nodeGroup.DeleteNodes([]*v1.Node{makeNode(fakeServer)})
		assert.Error(t, err)
	})
	t.Run("Delete Node", func(t *testing.T) {
		mockclient.On("Server", "srv-rp897").
			Return(fakeServerrp897(), nil).Once().
			On("Server", "srv-rp897").
			Return(deletedFakeServer(fakeServerrp897()), nil).
			Once().
			On("DestroyServer", "srv-rp897").
			Return(nil).Once()
		err := nodeGroup.DeleteNodes([]*v1.Node{makeNode("srv-rp897")})
		assert.NoError(t, err)
	})
	t.Run("Delete All Nodes", func(t *testing.T) {
		truncateServers := mocks.ServerListReducer(fakeServerGroup)
		mockclient.On("Server", "srv-rp897").
			Return(fakeServerrp897(), nil).Once().
			On("Server", "srv-rp897").
			Return(deletedFakeServer(fakeServerrp897()), nil).
			Once().
			On("DestroyServer", "srv-rp897").
			Return(nil).Once().Run(truncateServers)
		err := nodeGroup.DeleteNodes([]*v1.Node{
			makeNode("srv-rp897"),
			makeNode("srv-lv426"),
		})
		assert.Error(t, err)
	})

}

func TestExist(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(testclient)
	fakeServerGroup := &fakeGroups()[0]
	t.Run("Find Group", func(t *testing.T) {
		mockclient.On("ServerGroup", nodeGroup.Id()).
			Return(fakeServerGroup, nil).Once()
		assert.True(t, nodeGroup.Exist())
	})
	t.Run("Fail to Find Group", func(t *testing.T) {
		mockclient.On("ServerGroup", nodeGroup.Id()).
			Return(nil, serverNotFoundError(nodeGroup.Id()))
		assert.False(t, nodeGroup.Exist())
	})
	mockclient.AssertExpectations(t)
}

func TestNodes(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(testclient)
	fakeServerGroup := &fakeGroups()[0]
	mockclient.On("ServerGroup", fakeNodeGroupId).
		Return(fakeServerGroup, nil)
	t.Run("Both Active", func(t *testing.T) {
		fakeServerGroup.Servers[0].Status = "active"
		fakeServerGroup.Servers[1].Status = "active"
		nodes, err := nodeGroup.Nodes()
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeInstances, nodes)
	})
	t.Run("Creating and Deleting", func(t *testing.T) {
		fakeServerGroup.Servers[0].Status = "creating"
		fakeServerGroup.Servers[1].Status = "deleting"
		nodes, err := nodeGroup.Nodes()
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeTransitionInstances, nodes)
	})
	t.Run("Inactive and Unavailable", func(t *testing.T) {
		fakeServerGroup.Servers[0].Status = "inactive"
		fakeServerGroup.Servers[1].Status = "unavailable"
		nodes, err := nodeGroup.Nodes()
		require.NoError(t, err)
		assert.ElementsMatch(t, fakeErrorInstances, nodes)
	})
}

func TestTemplateNodeInfo(t *testing.T) {
	obj, err := makeFakeNodeGroup(nil).TemplateNodeInfo()
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func TestCreate(t *testing.T) {
	obj, err := makeFakeNodeGroup(nil).Create()
	assert.Equal(t, err, cloudprovider.ErrNotImplemented)
	assert.Nil(t, obj)
}

func TestDelete(t *testing.T) {
	assert.Equal(t, makeFakeNodeGroup(nil).Delete(), cloudprovider.ErrNotImplemented)
}

func TestAutoprovisioned(t *testing.T) {
	assert.False(t, makeFakeNodeGroup(nil).Autoprovisioned())
}

func makeFakeNodeGroup(brightboxCloudClient *k8ssdk.Cloud) *brightboxNodeGroup {
	return makeNodeGroupFromApiDetails(
		fakeNodeGroupId,
		fakeNodeGroupName,
		fakeNodeGroupDescription,
		fakeDefaultSize,
		fakeNodeGroupServerTypeId,
		fakeNodeGroupImageId,
		fakeNodeGroupZoneId,
		fakeNodeGroupMainGroupId,
		fakeNodeGroupUserData,
		brightboxCloudClient,
	)
}
