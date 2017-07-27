/*
 * This file is part of the kubevirt project
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
 * Copyright 2017 Red Hat, Inc.
 *
 */

package network

import (
	"errors"
	"fmt"
	//"strings"

	"github.com/jeevatkm/go-model"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/logging"
	"kubevirt.io/kubevirt/pkg/virt-handler/network/cniproxy"
	"kubevirt.io/kubevirt/pkg/virt-handler/virtwrap"
)

const HostNetNS = "/proc/1/ns/net"

type VirtualInterface interface {
	Plug(domainManager virtwrap.DomainManager) error
	Unplug(device string, domainManager virtwrap.DomainManager) error
	GetConfig() (*v1.Interface, error)
	DecorateInterfaceStatus() *v1.VMNetworkInterface
	DecorateInterfaceMetadata() *v1.InterfaceMetadata
}

func GetInterfaceType(objName string) (VirtualInterface, error) {
	switch objName {
	case "PodNetworkInterface":
		return new(PodNetworkInterface), nil
	default:
		return nil, errors.New("Invalid Interface type.")
	}
}

func getInterfaceTypeFromVMStatus(vm *v1.VirtualMachine, device string) (VirtualInterface, error) {
	for _, iface := range vm.Status.Interfaces {
		if iface.Device == device {
			obj, err := GetInterfaceType(iface.Type)
			if err != nil {
				return nil, err
			}
			return obj, nil
		}
	}
	return nil, fmt.Errorf("Failed to find device %s in known interfaces", device)
}

func UnPlugNetworkDevices(vm *v1.VirtualMachine, domainManager virtwrap.DomainManager) error {
	/*
		for _, inter := range vm.Spec.Domain.Devices.Interfaces {
			fmt.Println("Found interface type: ", inter.Type, "for VM: ", vm.ObjectMeta.Name)
			if inter.Type == "bridge" {
				s := strings.Split(inter.Source.Bridge, "-")
				device := "keth-" + s[1]
				vif, err := getInterfaceTypeFromVMStatus(vm, device)
				err = vif.Unplug(device, domainManager)
				if err != nil {
					return err
				}
			}

		}
	*/
	for _, inter := range vm.Spec.Domain.Metadata.Interfaces {
		fmt.Println("Found interface type: ", inter.Type, "for VM: ", vm.ObjectMeta.Name)
		fmt.Println("Interface: ", inter.Device)
		vif, err := GetInterfaceType(inter.Type)
		if err != nil {
			return err
		}
		err = vif.Unplug(inter.Device, domainManager)

		if err != nil {
			fmt.Println("Failed to unplug: ", inter.Device, "for VM: ", vm.ObjectMeta.Name)
			fmt.Println("ERROR: ", err)
		}

	}
	return nil
}

func PlugNetworkDevices(vm *v1.VirtualMachine, domainManager virtwrap.DomainManager) (*v1.VirtualMachine, error) {
	vmCopy := &v1.VirtualMachine{}
	model.Copy(vmCopy, vm)

	//TODO:(vladikr) Currently we support only one interface per vm. Improve this once we'll start supporting more.
	if len(vmCopy.Status.Interfaces) != len(vmCopy.Spec.Domain.Devices.Interfaces) {

		for idx, inter := range vmCopy.Spec.Domain.Devices.Interfaces {
			vif, err := GetInterfaceType(inter.Type)
			if err != nil {
				return nil, err
			}
			err = vif.Plug(domainManager)
			if err != nil {
				return nil, err
			}

			// Add VIF to VM config
			ifconf, err := vif.GetConfig()
			if err != nil {
				logging.DefaultLogger().Error().Reason(err).Msg("Failed to get VIF config.")
				return nil, err
			}
			vmCopy.Spec.Domain.Devices.Interfaces[idx] = *ifconf
			ifaceMeta := vif.DecorateInterfaceMetadata()
			vmCopy.Spec.Domain.Metadata.Interfaces = append(vmCopy.Spec.Domain.Metadata.Interfaces, *ifaceMeta)

			ifaceStatus := vif.DecorateInterfaceStatus()
			vmCopy.Status.Interfaces = append(vmCopy.Status.Interfaces, *ifaceStatus)
		}
	}
	return vmCopy, nil
}

func GetContainerInteface(iface string) (*cniproxy.CNIProxy, error) {
	runtime, err := cniproxy.BuildRuntimeConfig(iface)
	if err != nil {
		return nil, err
	}
	cniProxy, err := cniproxy.GetProxy(runtime)
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msg("Failed to get CNI Proxy")
		return nil, err
	}
	return cniProxy, nil
}
