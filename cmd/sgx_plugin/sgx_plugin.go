// Copyright 2017 Intel Corporation. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	dpapi "github.com/intel/intel-device-plugins-for-kubernetes/pkg/deviceplugin"
	"github.com/klauspost/cpuid"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	// Device plugin settings.
	namespace  = "sgx.intel.com"
	deviceType = "sgx"
	devicePath = "/dev"
)

type devicePlugin struct {
	devfsDir string
}

func newDevicePlugin(devfsDir string) *devicePlugin {
	return &devicePlugin{
		devfsDir: devfsDir,
	}
}

func (dp *devicePlugin) Scan(notifier dpapi.Notifier) error {
	for {
		devTree, err := dp.scan()
		if err != nil {
			return err
		}

		notifier.Notify(devTree)
		time.Sleep(60 * time.Second)
	}
}

func (dp *devicePlugin) scan() (dpapi.DeviceTree, error) {
	devTree := dpapi.NewDeviceTree()

	fmt.Println("SGX available:", cpuid.CPU.SGX.Available)
	fmt.Println("SGX launch control:", cpuid.CPU.SGX.LaunchControl)
	fmt.Println("SGX memory 1:", cpuid.CPU.SGX.MaxEnclaveSize64)
	fmt.Println("SGX memory 2:", cpuid.CPU.SGX.MaxEnclaveSizeNot64)

	sgxEnclavePath := path.Join(dp.devfsDir, "sgx", "enclave")
	sgxProvisionPath := path.Join(dp.devfsDir, "sgx", "provision")
	if _, err := os.Stat(sgxEnclavePath); err != nil {
		fmt.Println("No SGX enclave file available: ", err)
		return devTree, nil
	}
	if _, err := os.Stat(sgxProvisionPath); err != nil {
		fmt.Println("No SGX provision file available: ", err)
		return devTree, nil
	}

	devID := fmt.Sprintf("%s-%d", "sgx", 0) // FIXME
	nodes := []pluginapi.DeviceSpec{
		pluginapi.DeviceSpec{
			HostPath:      sgxEnclavePath,
			ContainerPath: sgxEnclavePath,
			Permissions:   "rw",
		},
		pluginapi.DeviceSpec{
			HostPath:      sgxProvisionPath,
			ContainerPath: sgxProvisionPath,
			Permissions:   "rw",
		},
	}
	devTree.AddDevice(deviceType, devID, dpapi.NewDeviceInfo(pluginapi.Healthy, nodes, nil, nil))

	return devTree, nil
}

func main() {
	var debugEnabled bool

	flag.BoolVar(&debugEnabled, "debug", false, "enable debug output")
	flag.Parse()

	fmt.Println("SGX device plugin started")

	plugin := newDevicePlugin(devicePath)
	manager := dpapi.NewManager(namespace, plugin)
	manager.Run()
}
