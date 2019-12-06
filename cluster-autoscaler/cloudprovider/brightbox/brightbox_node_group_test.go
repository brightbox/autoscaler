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
	"errors"
	"testing"

	"github.com/brightbox/brightbox-cloud-controller-manager/k8ssdk"
	"github.com/brightbox/brightbox-cloud-controller-manager/k8ssdk/mocks"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

const (
	fakeMaxSize      = 4
	fakeMinSize      = 1
	fakeDescription  = "1:4"
	fakeDefaultSize  = 3
	fakeNodeGroupId  = "grp-sda44"
	fakeImageId      = "img-testy"
	fakeServerTypeId = "typ-testy"
	fakeZoneId       = "zon-testy"
)

var (
	fakeError = errors.New("Fake API Error")
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
	fakeServerGroup := fakeGroups()[0]
	t.Run("TargetSize", func(t *testing.T) {
		mockclient.On("ServerGroup", fakeNodeGroupId).
			Return(&fakeServerGroup, nil).Once()
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
			Return(&fakeServerGroup, nil).Once()
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
		Return(&fakeServerGroup, nil)
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

func TestDeleteNodes(t *testing.T) {
	mockclient := new(mocks.CloudAccess)
	testclient := k8ssdk.MakeTestClient(mockclient, nil)
	nodeGroup := makeFakeNodeGroup(testclient)
	fakeServerGroup := fakeGroups()[0]
	mockclient.On("ServerGroup", fakeNodeGroupId).
		Return(&fakeServerGroup, nil).
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
		truncateServers := mocks.ServerListReducer(&fakeServerGroup)
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
		fakeDescription,
		fakeDefaultSize,
		fakeServerTypeId,
		fakeImageId,
		fakeZoneId,
		brightboxCloudClient,
	)
}
