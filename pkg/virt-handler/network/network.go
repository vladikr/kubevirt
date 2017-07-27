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
	"net"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/golang/glog"

	"kubevirt.io/kubevirt/pkg/virt-handler/network/cniproxy"
	//"github.com/vishvananda/netns"
	"github.com/containernetworking/plugins/pkg/ns"
	//"fmt"
	"fmt"
	"strings"

	"github.com/containernetworking/cni/pkg/types/020"
	"github.com/davecgh/go-spew/spew"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	lmf "github.com/subgraph/libmacouflage"
	"github.com/vishvananda/netlink"

	"github.com/jeevatkm/go-model"

	"kubevirt.io/kubevirt/pkg/virt-handler/virtwrap"
)

type VirtualInterface struct {
	cniProxy *cniproxy.CNIProxy
	Name     string
	IPAddr   net.IPNet
	Mac      net.HardwareAddr
}

func GetVirtualInteface() (*VirtualInterface, error) {
	cniProxy, err := cniproxy.GetCNIProxy()
	if err != nil {
		glog.Errorf("Error getting the CNI Proxy: %v", err)
		return nil, err
	}
	return &VirtualInterface{cniProxy: cniProxy}, nil
}

func (vif *VirtualInterface) RemoveInterface() error {
	err := vif.cniProxy.DeleteFromNetwork()
	if err != nil {
		glog.Errorf("Error deleting network: %v", err)
		return err
	}
	return nil
}

func (vif *VirtualInterface) AddInterface() (types.Result, error) {
	res, err := vif.cniProxy.AddToNetwork()
	if err != nil {
		glog.Errorf("Error adding network: %v", err)
		return nil, err
	}
	glog.Infoln(spew.Sdump(res))
	return res, nil
}

func SetupVmNetworkInterface(domainManager virtwrap.DomainManager) (*VirtualInterface, error) {
	var vif VirtualInterface

	netns, err := ns.GetNS("/proc/1/ns/net")
	if err != nil {
		glog.Errorf("Cannot get host net namespace: %v", err)
		return nil, err
	}
	var result types.Result
	netns.Do(func(_ ns.NetNS) error {
		vif, err := GetVirtualInteface()
		if err != nil {
			return err
		}
		res, err := vif.cniProxy.AddToNetwork()
		if err != nil {
			glog.Errorf("Error adding network: %v", err)
			return err
		}
		result = res
		return nil
	})

	r, err := types020.GetResult(result)
	if err != nil {
		return nil, err
	}

	glog.Errorln(r.IP4.IP)
	glog.Errorln(r.IP4.Gateway)
	vif.IPAddr = r.IP4.IP

	//Switch to libvirt namespace
	libvNS, err := cniproxy.GetLibvirtNS()
	if err != nil {
		return nil, err
	}
	libvnetns, err := ns.GetNS(libvNS.Net)
	if err != nil {
		glog.Errorf("Cannot get libvirt net namespace: %v", err)
		return nil, err
	}
	libvnetns.Do(func(_ ns.NetNS) error {
		inter, err := GetInterfaceByIP(r.IP4.IP.String())
		if err != nil {
			glog.Errorf("Cannot get interface by IP: %s, reason: %v", r.IP4.IP.String(), err)
			return err
		}
		vif.Name = inter.Name
		links, err := netlink.LinkList()
		fmt.Println("ip link show: ")
		for _, link := range links {
			fmt.Println(link.Attrs().Name)
		}

		link, err := netlink.LinkByName(inter.Name)
		if err != nil {
			glog.Errorf("Error getting link for iface: %s, reason: %v", inter.Name, err)
			return err
		}
		err = netlink.AddrDel(link, &netlink.Addr{IPNet: &r.IP4.IP})

		if err != nil {
			glog.Errorf("Cannot delete link for iface: %s, reason: %v", vif.Name, err)
			return err
		}
		links, err = netlink.LinkList()
		fmt.Println("ip link show: ")
		for _, link := range links {
			fmt.Println(link.Attrs().Name)
		}
		err = netlink.LinkSetDown(link)
		if err != nil {
			glog.Errorf("Cannot bring link down for iface: %s, reason: %v", vif.Name, err)
			return err
		}

		vif.Mac, err = changeMacAddr(vif.Name)
		if err != nil {
			return err
		}

		return nil
	})

	netName := "net-" + vif.Name
	netXML, err := createNetworkXML(&vif)
	if err != nil {
		return nil, err
	}
	err = domainManager.CreateNetwork(netXML, netName)
	if err != nil {
		return nil, err
	}

	/*libvnetns.Do(func(_ ns.NetNS) error {
		link, err := netlink.LinkByName("br-" + vif.Name + "-nic")
		if err != nil {
			glog.Errorf("Error getting link for iface: %s, reason: %v", "br-"+vif.Name+"-nic", err)
			return err
		}
		//netlink.LinkSetMaster(link)

		link, err = netlink.LinkByName("br-" + vif.Name + "-nic")
		if err != nil {
			glog.Errorf("Error getting link for iface: %s, reason: %v", "br-"+vif.Name+"-nic", err)
			return err
		}
		netlink.LinkSetUp(link)
		return nil
	})*/

	return &vif, nil
}

func getMacDetails(iface string, prefix string) (net.HardwareAddr, error) {
	currentMac, err := lmf.GetCurrentMac(iface)
	if err != nil {
		glog.Errorf("Cannot get mac information for iface: %s, reason: %v", iface, err)
		return nil, err
	}
	glog.Errorf("%s - %s MacInfo: %s", prefix, iface, currentMac)

	permanentMac, err := lmf.GetPermanentMac(iface)
	if err != nil {
		glog.Errorf("%v", err)
		return nil, err
	}
	glog.Errorf("%s - %s Mac: %s", prefix, iface, permanentMac.String())

	return currentMac, nil
}

func changeMacAddr(iface string) (net.HardwareAddr, error) {
	var mac net.HardwareAddr

	_, err := getMacDetails(iface, "Current")
	if err != nil {
		return nil, err
	}

	changed, err := lmf.SpoofMacRandom(iface, false)
	if err != nil {
		glog.Errorf("Cannot Spoof MAC for iface: %s, reason: %v", iface, err)
		return nil, err
	}

	if changed {
		mac, err = getMacDetails(iface, "Updated")
		if err != nil {
			return nil, err
		}
	}
	return mac, nil
}

func GetInterfaceByIP(ip string) (*net.Interface, error) {

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, inter := range interfaces {
		fmt.Println(inter.Name, inter.HardwareAddr)
		if addrs, err := inter.Addrs(); err == nil {
			for _, addr := range addrs {
				if addr.(*net.IPNet).IP.To4() != nil && strings.Contains(addr.String(), ip) {
					fmt.Println("Found interface: ", inter.Name, " - ", addr.String())
					return &inter, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("Didn't find an interface with ip: %s", ip)
}

/*
<network>
  <name>net-$IFNAME</name>
  <bridge name="br-$IFNAME" />
  <ip address="$DNS_IP" prefix="$NIC_PREFIX" localPtr="yes">
    <dhcp>
      <range start="$NIC_IP" end="$NIC_IP"/>
      <host mac="$NIC_MAC" name="FIXME_HOSTNAME" ip="$NIC_IP"/>
    </dhcp>
  </ip>
</network>
*/

func createNetworkXML(vif *VirtualInterface) (*string, error) {
	ipCopy := &net.IPNet{}
	//	save := vif.IPAddr.IP.To4()[3]

	errs := model.Copy(ipCopy, vif.IPAddr)
	if errs != nil {
		for _, err := range errs {
			glog.Errorf("Cannot copy model, reason: %v", err)
		}
		return nil, errs[0]
	}
	dns := ipCopy.IP.To4()
	fmt.Println(dns)

	fmt.Println(vif.IPAddr.IP.To4().String())
	if dns != nil {
		dns[3] = 254
	}
	fmt.Println(vif.IPAddr.IP.To4().String())
	/*
		if vif.IPAddr.IP.To4()[3] == 254 {
			fmt.Println("AAAAAAAAAAAAAAA!!!!!!!!")
			vif.IPAddr.IP.To4()[3] = save
			model.Copy(ipCopy, &vif.IPAddr.IP)
			dns := *ipCopy
			dns[3] = 254
			if vif.IPAddr.IP.To4()[3] == 254 {
				fmt.Println("WTF!!!!!!!!!!!!!!!!!!!")
			}
		}
	*/
	net := libvirtxml.Network{
		Name: "net-" + vif.Name,
		Bridge: &libvirtxml.NetworkBridge{
			Name: "br-" + vif.Name,
		},
		IPs: []libvirtxml.NetworkIP{
			libvirtxml.NetworkIP{
				Address:  dns.String(),
				Prefix:   "",
				LocalPtr: "yes",
				DHCP: &libvirtxml.NetworkDHCP{
					Ranges: []libvirtxml.NetworkDHCPRange{
						libvirtxml.NetworkDHCPRange{
							Start: vif.IPAddr.IP.To4().String(),
							End:   vif.IPAddr.IP.To4().String(),
						},
					},
					Hosts: []libvirtxml.NetworkDHCPHost{
						libvirtxml.NetworkDHCPHost{
							MAC:  vif.Mac.String(),
							Name: "hostname",
							IP:   vif.IPAddr.IP.To4().String(),
						},
					},
				},
			},
		},
	}
	xml, err := net.Marshal()
	if err != nil {
		return nil, err
	}
	return &xml, nil
}
