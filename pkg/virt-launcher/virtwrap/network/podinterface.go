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

//go:generate mockgen -source $GOFILE -package=$GOPACKAGE -destination=generated_mock_$GOFILE

package network

import (
	"fmt"
	"net"
	"syscall"

	"github.com/docker/go-connections/proxy"
	"github.com/vishvananda/netlink"

	apiv1 "k8s.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/log"
	"kubevirt.io/kubevirt/pkg/precond"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
	//"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/network/proxy"
)

var CommonFakeVMIP = "192.168.0.2/24"

var bridgeFakeIP = "169.254.75.86/32"

type BindMechanism interface {
	discoverPodNetworkInterface() error
	preparePodNetworkInterfaces() error
	startDHCPServer()
	decorateInterfaceConfig(ifconf *api.Interface)
}

type PodInterface struct{}

func (l *PodInterface) Unplug() {}

func (l *PodInterface) Plug(iface *v1.Interface, network *v1.Network, domain *api.Domain) error {
	precond.MustNotBeNil(domain)
	initHandler()

	driver, err := getBinding(iface)
	if err != nil {
		return err
	}

	// There should alway be a pre-configured interface for the default pod interface.
	defaultIconf := domain.Spec.Devices.Interfaces[0]

	ifconf, err := getCachedInterface()
	if err != nil {
		return err
	}

	if ifconf == nil {
		err := driver.discoverPodNetworkInterface()
		if err != nil {
			return err
		}
		if err := driver.preparePodNetworkInterfaces(); err != nil {
			log.Log.Reason(err).Critical("failed to prepared pod networking")
			panic(err)
		}

		driver.startDHCPServer()

		// After the network is configured, cache the result
		// in case this function is called again.
		driver.decorateInterfaceConfig(&defaultIconf)
		err = setCachedInterface(&defaultIconf)
		if err != nil {
			panic(err)
		}
		ifconf = &defaultIconf
	}

	// TODO:(vladikr) Currently we support only one interface per vm.
	// Improve this once we'll start supporting more.
	if len(domain.Spec.Devices.Interfaces) == 0 {
		domain.Spec.Devices.Interfaces = append(domain.Spec.Devices.Interfaces, *ifconf)
	} else {
		domain.Spec.Devices.Interfaces[0] = *ifconf
	}

	return nil
}

func getBinding(iface *v1.Interface) (BindMechanism, error) {
	vif := &VIF{Name: podInterface}
	if iface.Bridge != nil {
		return &BridgePodInterface{iface: iface, vif: vif}, nil
	} else if iface.Proxy != nil {
		intr := &ProxyPodInterface{}
		intr.iface = iface
		intr.vif = vif
		return intr, nil
	}
	return nil, fmt.Errorf("Not implemented")
}

type BridgePodInterface struct {
	vif        *VIF
	iface      *v1.Interface
	podNicLink netlink.Link
}

func (b *BridgePodInterface) discoverPodNetworkInterface() error {
	link, err := Handler.LinkByName(podInterface)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get a link for interface: %s", podInterface)
		return err
	}
	b.podNicLink = link

	// get IP address
	addrList, err := Handler.AddrList(b.podNicLink, netlink.FAMILY_V4)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get an ip address for %s", podInterface)
		return err
	}
	if len(addrList) == 0 {
		return fmt.Errorf("No IP address found on %s", podInterface)
	}
	b.vif.IP = addrList[0]

	// Handle interface routes
	if err := b.setInterfaceRoutes(); err != nil {
		return err
	}

	// Get interface MAC address
	mac, err := Handler.GetMacDetails(podInterface)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get MAC for %s", podInterface)
		return err
	}
	b.vif.MAC = mac

	// Get interface MTU
	b.vif.Mtu = uint16(b.podNicLink.Attrs().MTU)
	return nil
}

func (b *BridgePodInterface) preparePodNetworkInterfaces() error {
	// Remove IP from POD interface
	err := Handler.AddrDel(b.podNicLink, &b.vif.IP)

	if err != nil {
		log.Log.Reason(err).Errorf("failed to delete link for interface: %s", podInterface)
		return err
	}

	// Set interface link to down to change its MAC address
	err = Handler.LinkSetDown(b.podNicLink)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to bring link down for interface: %s", podInterface)
		return err
	}

	_, err = Handler.SetRandomMacAddr(podInterface)
	if err != nil {
		return err
	}

	err = Handler.LinkSetUp(b.podNicLink)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to bring link up for interface: %s", podInterface)
		return err
	}

	if err := b.createDefaultBridge(); err != nil {
		return err
	}

	return nil
}

func (b *BridgePodInterface) startDHCPServer() {
	// Start DHCP Server
	fakeServerAddr, _ := netlink.ParseAddr(bridgeFakeIP)
	go Handler.StartDHCP(b.vif, fakeServerAddr)
}

func (b *BridgePodInterface) decorateInterfaceConfig(ifconf *api.Interface) {
	ifconf.MAC = &api.MAC{MAC: b.vif.MAC.String()}
}

func (b *BridgePodInterface) setInterfaceRoutes() error {
	routes, err := Handler.RouteList(b.podNicLink, netlink.FAMILY_V4)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get routes for %s", podInterface)
		return err
	}
	if len(routes) == 0 {
		return fmt.Errorf("No gateway address found in routes for %s", podInterface)
	}
	b.vif.Gateway = routes[0].Gw
	if len(routes) > 1 {
		dhcpRoutes := filterPodNetworkRoutes(routes, b.vif)
		b.vif.Routes = &dhcpRoutes
	}
	return nil
}

func (b *BridgePodInterface) createDefaultBridge() error {
	// Create a bridge
	bridge := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: api.DefaultBridgeName,
		},
	}
	err := Handler.LinkAdd(bridge)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create a bridge")
		return err
	}
	netlink.LinkSetMaster(b.podNicLink, bridge)

	err = Handler.LinkSetUp(bridge)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to bring link up for interface: %s", api.DefaultBridgeName)
		return err
	}

	// set fake ip on a bridge
	fakeaddr, err := Handler.ParseAddr(bridgeFakeIP)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to bring link up for interface: %s", api.DefaultBridgeName)
		return err
	}

	if err := Handler.AddrAdd(bridge, fakeaddr); err != nil {
		log.Log.Reason(err).Errorf("failed to set bridge IP")
		return err
	}

	return nil
}

type ProxyPodInterface struct {
	BridgePodInterface
}

func (b *ProxyPodInterface) setInterfaceRoutes() error {
	ip, _, _ := net.ParseCIDR(bridgeFakeIP)
	b.vif.Gateway = ip
	b.vif.Routes = &[]netlink.Route{}
	return nil
}

func (b *ProxyPodInterface) discoverPodNetworkInterface() error {
	bridgeFakeIP = "192.168.0.1/24"

	link, err := Handler.LinkByName(podInterface)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get a link for interface: %s", podInterface)
		return err
	}

	b.podNicLink = link
	// get IP address
	ip, _ := netlink.ParseAddr(CommonFakeVMIP)
	b.vif.IP = *ip

	// Get interface gateway
	if err := b.setInterfaceRoutes(); err != nil {
		return err
	}

	// generate a random MAC for the VM
	mac, err := Handler.RandomizeMac()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to generate MAC address")
		return err
	}
	b.vif.MAC = mac
	return nil
}

func (b *ProxyPodInterface) setupTapDevice() error {

	tapName := api.DefaultBridgeName + "-nic"

	// Create a tap device
	tap := &netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{
			Name:  tapName,
			Flags: net.FlagUp,
			MTU:   1500,
		},
		Mode: syscall.IFF_TAP,
	}

	err := Handler.LinkAdd(tap)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create a tap device")
		return err
	}

	err = Handler.LinkSetUp(tap)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to bring link up for interface: %s", tapName)
		return err
	}
	b.podNicLink = tap

	return nil
}

func (b *ProxyPodInterface) formatAddresses(port apiv1.ContainerPort) (frontendAddr net.Addr, backendAddr net.Addr, err error) {
	switch port.Protocol {
	case apiv1.ProtocolTCP:
		frontendAddr, err = net.ResolveTCPAddr("tcp",
			fmt.Sprintf(":%d", port.ContainerPort))
		if err != nil {
			return
		}
		backendAddr, err = net.ResolveTCPAddr("tcp",
			fmt.Sprintf("%s:%d", b.vif.IP.IP.String(), port.HostPort))
		if err != nil {
			return
		}
		return
	case apiv1.ProtocolUDP:
		frontendAddr, err = net.ResolveTCPAddr("tcp",
			fmt.Sprintf(":%d", port.ContainerPort))
		if err != nil {
			return
		}
		backendAddr, err = net.ResolveTCPAddr("tcp",
			fmt.Sprintf("%s:%d", b.vif.IP.IP.String(), port.HostPort))
		if err != nil {
			return
		}
		return
	}
	return
}

func (b *ProxyPodInterface) proxyPorts() error {
	for _, port := range b.iface.Proxy.Ports {
		frontendAddr, backendAddr, err := b.formatAddresses(port)
		if err != nil {
			return err
		}

		channel, err := proxy.NewProxy(frontendAddr, backendAddr)
		if err != nil {
			return err
		}

		// start forwarding
		go channel.Run()
	}
	return nil
}

func (b *ProxyPodInterface) preparePodNetworkInterfaces() error {
	if err := b.setupTapDevice(); err != nil {
		return err
	}

	if err := b.createDefaultBridge(); err != nil {
		return err
	}

	// Forward requested ports
	b.proxyPorts()

	return nil
}
