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
	"math"
	"os"
	"time"
	"strings"

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
	virtConfig    *virtconfig.ClusterConfig
}

func NewDeviceController(host string, maxDevices int, clusterConfig *virtconfig.ClusterConfig) *DeviceController {
	logger := log.DefaultLogger()
	logger.Infof("Device Controller: config: %v", clusterConfig)
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
	controller.virtConfig = clusterConfig

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
	logger.Infof(" device plugins: %v", c.devicePlugins)
	logger.Infof(" device plugins device name: %v", deviceName)
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

func (c *DeviceController) addPermittedHostDevicePlugins() {
	logger := log.DefaultLogger()
	if hostDevs := c.virtConfig.GetPermittedHostDevices(); hostDevs != nil {
		logger.Infof("Device Controller: permitted devices: %v", hostDevs)
		logger.Infof("Device Controller: permitted pci devices: %v", hostDevs.PciHostDevices)
		logger.Infof("Device Controller: mdevs: %v", hostDevs.MediatedDevices)
		supportedPCIDeviceMap := make(map[string]string)
		if len(hostDevs.PciHostDevices) != 0 {
			for _, pciDev := range hostDevs.PciHostDevices {
				logger.Infof("Device Controller: devices: %v, ExternalResourceProvider: %v", pciDev, pciDev.ExternalResourceProvider)
				if !pciDev.ExternalResourceProvider {
					supportedPCIDeviceMap[pciDev.Selector] = pciDev.ResourceName
				}
			}
			logger.Infof("Device Controller: formed a supportedPCIDeviceMap : %v", supportedPCIDeviceMap)
			pciHostDevices := discoverPermittedHostPCIDevices(supportedPCIDeviceMap)
			logger.Infof("Device Controller: discovered pci devices: %v", pciHostDevices)
			pciDevicePlugins := []GenericDevice{}
			for pciID, pciDevices := range pciHostDevices {
				pciResourceName := supportedPCIDeviceMap[pciID]
				logger.Infof("Device Controller: getting pciID: %v, from supportedPCIDeviceMap, got :%v", pciID, pciResourceName)
				pciDevicePlugins = append(pciDevicePlugins, NewPCIDevicePlugin(pciDevices, pciResourceName))
			}
			c.devicePlugins = append(c.devicePlugins, pciDevicePlugins...)
		}
		if len(hostDevs.MediatedDevices) != 0 {
			supportedMdevsMap := make(map[string]string)
			for _, supportedMdev := range hostDevs.MediatedDevices {
				if !supportedMdev.ExternalResourceProvider {
					selector := removeSelectorSpaces(supportedMdev.Selector)
					supportedMdevsMap[selector] = supportedMdev.ResourceName
				}
			}

			logger.Infof("Device Controller: formed a supportedMdevsMap : %v", supportedMdevsMap)
			hostMdevs := discoverPermittedHostMediatedDevices(supportedMdevsMap)
			logger.Infof("Device Controller: discovered mdevs: %v", hostMdevs)
			mdevPlugins := []GenericDevice{}
			for mdevTypeName, mdevUUIDs := range hostMdevs {
				mdevResourceName := supportedMdevsMap[mdevTypeName]
				logger.Infof("Device Controller: create mdev dp for %v", mdevResourceName)
				mdevPlugins = append(mdevPlugins, NewMediatedDevicePlugin(mdevUUIDs, mdevResourceName))
			}
			c.devicePlugins = append(c.devicePlugins, mdevPlugins...)
		}
	}
}

func removeSelectorSpaces(selectorName string) string {
	// The name usually contain spaces which should be replaced with _
	// Such as GRID T4-1Q
	typeNameStr := strings.Replace(string(selectorName), " ", "_", -1)
	typeNameStr = strings.TrimSpace(typeNameStr)
	return typeNameStr

}

func (c *DeviceController) Run(stop chan struct{}) error {
	logger := log.DefaultLogger()
	logger.Info("Starting device plugin controller")
	c.addPermittedHostDevicePlugins()
	logger.Infof(" device plugins: %v", c.devicePlugins)
	for _, dev := range c.devicePlugins {
		logger.Infof(" device plugin dev: %v", dev)
		go c.startDevicePlugin(dev, stop)
	}

	<-stop

	logger.Info("Shutting down device plugin controller")
	return nil
}
