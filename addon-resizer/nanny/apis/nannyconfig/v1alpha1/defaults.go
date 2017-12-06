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

package v1alpha1

import (
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig"
)

func addDefaultingFuncs(scheme *kruntime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_NannyConfiguration(obj *NannyConfiguration) {
	if obj.BaseCPU == "" {
		obj.BaseCPU = nannyconfig.NoValue
	}
	if obj.CPUPerNode == "" {
		obj.CPUPerNode = "0"
	}
	if obj.BaseMemory == "" {
		obj.BaseMemory = nannyconfig.NoValue
	}
	if obj.MemoryPerNode == "" {
		obj.MemoryPerNode = "0"
	}
}

func FillInDefaults_NannyConfiguration(obj *NannyConfiguration, defaults *NannyConfiguration) {
	if obj.BaseCPU == nannyconfig.NoValue {
		obj.BaseCPU = defaults.BaseCPU
	}
	if obj.CPUPerNode == "0" {
		obj.CPUPerNode = defaults.CPUPerNode
	}
	if obj.BaseMemory == nannyconfig.NoValue {
		obj.BaseMemory = defaults.BaseMemory
	}
	if obj.MemoryPerNode == "0" {
		obj.MemoryPerNode = defaults.MemoryPerNode
	}
}
