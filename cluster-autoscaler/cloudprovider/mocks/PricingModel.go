/*
Copyright 2017 The Kubernetes Authors.

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

package mocks

import mock "github.com/stretchr/testify/mock"
import time "time"
import v1 "k8s.io/api/core/v1"

// PricingModel is an autogenerated mock type for the PricingModel type
type PricingModel struct {
	mock.Mock
}

// NodePrice provides a mock function with given fields: node, startTime, endTime
func (_m *PricingModel) NodePrice(node *v1.Node, startTime time.Time, endTime time.Time) (float64, error) {
	ret := _m.Called(node, startTime, endTime)

	var r0 float64
	if rf, ok := ret.Get(0).(func(*v1.Node, time.Time, time.Time) float64); ok {
		r0 = rf(node, startTime, endTime)
	} else {
		r0 = ret.Get(0).(float64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Node, time.Time, time.Time) error); ok {
		r1 = rf(node, startTime, endTime)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PodPrice provides a mock function with given fields: pod, startTime, endTime
func (_m *PricingModel) PodPrice(pod *v1.Pod, startTime time.Time, endTime time.Time) (float64, error) {
	ret := _m.Called(pod, startTime, endTime)

	var r0 float64
	if rf, ok := ret.Get(0).(func(*v1.Pod, time.Time, time.Time) float64); ok {
		r0 = rf(pod, startTime, endTime)
	} else {
		r0 = ret.Get(0).(float64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Pod, time.Time, time.Time) error); ok {
		r1 = rf(pod, startTime, endTime)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
