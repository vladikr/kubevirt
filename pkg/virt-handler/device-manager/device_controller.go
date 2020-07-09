/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package device_manager

import (
	"fmt"
	"math"
	"os"
	"time"

	"kubevirt.io/client-go/log"
	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
)

const (
	KVMPath      = "/dev/kvm"
	KVMName      = "kvm"
	TunPath      = "/dev/net/tun"
	TunName      = "tun"
	VhostNetPath = "/dev/vhost-net"
	VhostNetName = "vhost-net"
)

type DeviceController struct {
	devicePlugins []GenericDevice
	host          string
	maxDevices    int
	backoff       []time.Duration
}

func NewDeviceController(host string, maxDevices int, clusterConfig *virtconfig.ClusterConfig) *DeviceController {
	controller := &DeviceController{
		devicePlugins: []GenericDevice{
			NewGenericDevicePlugin(KVMName, KVMPath, maxDevices, false),
			NewGenericDevicePlugin(TunName, TunPath, maxDevices, true),
			NewGenericDevicePlugin(VhostNetName, VhostNetPath, maxDevices, true),
		},
		host:       host,
		maxDevices: maxDevices,
		backoff:    []time.Duration{1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second},
	}
	if hostDevs := clusterConfig.GetPermittedHostDevices(); hostDevs != nil {
		if len(hostDevs.PciHostDevices) != 0 {
			supportedPCIDeviceMap := make(map[string]string)
			for _, pciDev := range hostDevs.PciHostDevices {
				if !pciDev.ExternalResourceProvider {
					supportedPCIDeviceMap[pciDev.Selector] = pciDev.ResourceName
				}
			}
			pciHostDevices := discoverPermittedHostPCIDevices(supportedPCIDeviceMap)
			pciDevicePlugins := []GenericDevice{}
			for pciID, pciDevices := range pciHostDevices {
				pciResourceName := supportedPCIDeviceMap[pciID]
				pciDevicePlugins = append(pciDevicePlugins, NewPCIDevicePlugin(pciDevices, resourceName))
			}
			controller.devicePlugins = append(controller.devicePlugins, pciDevicePlugins...)
		}
		if len(hostDevs.MediatedDevices) != 0 {
			supportedMdevsMap := make(map[string]string)
			for _, supportedMdev := range hostDevs.MediatedDevices {
				if !supportedMdev.ExternalResourceProvider {
					supportedMdevsMap[suppotedMdev.ModelSelector] = supportedMdev.ResourceNamespace
				}
			}

			hostMdevs := discoverPermittedHostMediatedDevices(supportedMdevsMap)
			mdevPlugins := []GenericDevice{}
			for mdevTypeName, mdevUUIDs := range hostMdevs {
				mdevResourceNamespace := supportedPCIDeviceMap[mdevTypeName]
				mdevResourceName = fmt.Sprintf("%s/%s", mdevResourceNamespace, mdevTypeName)
				mdevPlugins = append(mdevPlugins, NewMediatedDevicePlugin(mdevUUIDs, mdevResourceName))
			}
			controller.devicePlugins = append(controller.devicePlugins, mdevPlugins...)
		}
	}

	return controller
}

func (c *DeviceController) nodeHasDevice(devicePath string) bool {
	_, err := os.Stat(devicePath)
	// Since this is a boolean question, any error means "no"
	return (err == nil)
}

func (c *DeviceController) startDevicePlugin(dev GenericDevice, stop chan struct{}) {
	logger := log.DefaultLogger()
	deviceName := dev.GetDeviceName()
	retries := 0

	for {
		err := dev.Start(stop)
		if err != nil {
			logger.Reason(err).Errorf("Error starting %s device plugin", deviceName)
			retries = int(math.Min(float64(retries+1), float64(len(c.backoff)-1)))
		} else {
			retries = 0
		}

		select {
		case <-stop:
			// Ok we don't want to re-register
			return
		default:
			// Wait a little bit and re-register
			time.Sleep(c.backoff[retries])
		}
	}
}

func (c *DeviceController) Run(stop chan struct{}) error {
	logger := log.DefaultLogger()
	logger.Info("Starting device plugin controller")

	for _, dev := range c.devicePlugins {
		go c.startDevicePlugin(dev, stop)
	}

	<-stop

	logger.Info("Shutting down device plugin controller")
	return nil
}
