package network

import (
	"fmt"
	"net"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/020"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/jeevatkm/go-model"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	"github.com/vishvananda/netlink"

	"strings"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/logging"
	"kubevirt.io/kubevirt/pkg/virt-handler/network/cniproxy"
	"kubevirt.io/kubevirt/pkg/virt-handler/network/utils"
	"kubevirt.io/kubevirt/pkg/virt-handler/virtwrap"
)

type PodNetworkInterface struct {
	cniProxy *cniproxy.CNIProxy
	Name     string
	IPAddr   net.IPNet
	Mac      net.HardwareAddr
	Gateway  net.IP
}

func (vif *PodNetworkInterface) Plug(domainManager virtwrap.DomainManager) error {
	var result types.Result

	netns, err := ns.GetNS(HostNetNS)
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msg("Failed to get host net namespace.")
		return err
	}

	iface, err := utils.GenerateRandomTapName()
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msg("Failed to generate random tap name")
		return err
	}

	// Create network interface
	err = netns.Do(func(_ ns.NetNS) error {
		vif.cniProxy, err = GetContainerInteface(iface)
		if err != nil {
			return err
		}
		res, err := vif.Create()
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Println("results: ", result.String())
	r, err := types020.GetResult(result)

	if err != nil {
		return err
	}
	vif.IPAddr = r.IP4.IP
	vif.Gateway = r.IP4.Gateway.To4()
	fmt.Printf("%s\n", r.IP4.Gateway.To4().String())
	fmt.Printf("%s", r.DNS.Domain)

	// Switch to libvirt namespace
	libvNS, err := cniproxy.GetLibvirtNS()
	if err != nil {
		return err
	}
	libvnetns, err := ns.GetNS(libvNS.Net)
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msg("Cannot get libvirt net namespace")
		return err
	}
	libvnetns.Do(func(_ ns.NetNS) error {
		inter, err := utils.GetInterfaceByIP(vif.IPAddr.String())
		if err != nil {
			logging.DefaultLogger().Error().Reason(err).Msgf("Cannot get interface by IP: %s", vif.IPAddr.String())
			return err
		}
		fmt.Println("Found created interface: ", inter.Name)
		vif.Name = inter.Name
		fmt.Println("Vif.name is: ", vif.Name)
		link, err := netlink.LinkByName(inter.Name)
		if err != nil {
			logging.DefaultLogger().Error().Reason(err).Msgf("Error getting link for interface: %s", inter.Name)
			return err
		}
		err = netlink.AddrDel(link, &netlink.Addr{IPNet: &vif.IPAddr})

		if err != nil {
			logging.DefaultLogger().Error().Reason(err).Msgf("Cannot delete link for interface: %s", vif.Name)
			return err
		}

		// Set interface link to down to change its MAC address
		err = netlink.LinkSetDown(link)
		if err != nil {
			logging.DefaultLogger().Error().Reason(err).Msgf("Cannot bring link down for interface: %s", vif.Name)
			return err
		}

		vif.Mac, err = utils.ChangeMacAddr(vif.Name)
		if err != nil {
			return err
		}

		_, err = vif.AddToBridge()
		if err != nil {
			return err
		}

		return nil
	})

	//ifId := strings.Split(vif.Name, "-")
	//netName := "net-" + ifId[1]
	/*
		netXML, err := vif.createPodNetworkBridgeXML()
		if err != nil {
			return err
		}
		fmt.Println("Creating network: ", *netXML, "for interface: ", vif.Name)
		//logging.DefaultLogger().Object(vif).Info().V(3).With("xml", xmlStr).Msgf("Domain XML generated.")
		err = domainManager.CreateNetwork(netXML, netName)
		if err != nil {
			//TODO:(vladikr) Need to rollback here, but the following crashes with null pointer exception from cni
			fmt.Printf("Network  %s creation failed, going to delete interface %s : \n", netName, vif.Name)
			logging.DefaultLogger().Error().Reason(err).Msgf("Failed to create a network: %s", netName)
			netns.Do(func(_ ns.NetNS) error {
				viferr := vif.cniProxy.DeleteFromNetwork()
				if viferr != nil {
					logging.DefaultLogger().Error().Reason(err).Msg("Failed to remove VIF during rollback.")
				}
				return nil
			})

			return err
		}
		fmt.Println("Created network: ", netName)
	*/
	return nil
}

func (vif *PodNetworkInterface) Unplug(device string, domainManager virtwrap.DomainManager) error {
	netns, err := ns.GetNS(HostNetNS)
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msg("Cannot get host net namespace.")
		return err
	}

	if vif.Name == "" {
		vif.Name = device
	}
	// Remove network interface
	err = netns.Do(func(_ ns.NetNS) error {

		// Look for a bridge link
		ifId := strings.Split(vif.Name, "-")
		bridgeName := "br-" + ifId[1]
		l, err := netlink.LinkByName(bridgeName)
		if err == nil {
			// delete the link
			if err := netlink.LinkDel(l); err != nil {
				logging.DefaultLogger().Error().Reason(err).Msgf("Failed to remove bridge interface %s delete", bridgeName)
			}
		}

		fmt.Println("Unpluging interface: ", vif.Name)
		vif.cniProxy, err = GetContainerInteface(vif.Name)
		if err != nil {
			return err
		}
		err = vif.Remove()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	/*
		ifId := strings.Split(vif.Name, "-")
		netName := "net-" + ifId[1]
		fmt.Println("Destroying network: ", netName)
		err = domainManager.DestroyNetwork(netName)
		if err != nil {
			return err
		}
	*/
	return nil
}

func (vif *PodNetworkInterface) Remove() error {
	err := vif.cniProxy.DeleteFromNetwork()
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msg("Failed to delete container interface")
		return err
	}
	return nil
}

func (vif *PodNetworkInterface) Create() (types.Result, error) {
	res, err := vif.cniProxy.AddToNetwork()
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msg("Failed to create container interface")
		return nil, err
	}
	return res, nil
}

func (vif *PodNetworkInterface) DecorateInterfaceStatus() *v1.VMNetworkInterface {
	fmt.Println("Decorating Vif name: ", vif.Name)
	inter := v1.VMNetworkInterface{
		Type:   "PodNetworkInterface",
		Device: vif.Name,
		IPAddr: vif.IPAddr.String(),
		HWAddr: vif.Mac.String(),
	}
	return &inter
}

func (vif *PodNetworkInterface) DecorateInterfaceMetadata() *v1.InterfaceMetadata {
	fmt.Println("Decorating Vif name: ", vif.Name)
	inter := v1.InterfaceMetadata{
		Type:   "PodNetworkInterface",
		Device: vif.Name,
	}
	return &inter
}

func (vif *PodNetworkInterface) GetConfigBridge() (*v1.Interface, error) {
	inter := v1.Interface{}

	ifId := strings.Split(vif.Name, "-")
	inter.Type = "bridge"
	inter.Source = v1.InterfaceSource{Bridge: "br-" + ifId[1]}
	inter.MAC = &v1.MAC{MAC: vif.Mac.String()}
	inter.Model = &v1.Model{Type: "virtio"}

	return &inter, nil
}

func (vif *PodNetworkInterface) AddToBridge() (string, error) {
	ifId := strings.Split(vif.Name, "-")

	la := netlink.NewLinkAttrs()
	la.Name = "br-" + ifId[1]
	mybridge := &netlink.Bridge{LinkAttrs: la}
	err := netlink.LinkAdd(mybridge)
	if err != nil {
		fmt.Printf("could not add %s: %v\n", la.Name, err)
		return "", fmt.Errorf("Failed to create a bridge %s, reason: %v", la.Name, err)
	}
	iface, _ := netlink.LinkByName(vif.Name)
	netlink.LinkSetMaster(iface, mybridge)
	return la.Name, nil
}

func (vif *PodNetworkInterface) GetConfig() (*v1.Interface, error) {

	/*
		inter := libvirtxml.DomainInterface{
			Type: "direct",
			MAC: &libvirtxml.DomainInterfaceMAC{
				Address: vif.Mac.String(),
			},
			Model: &libvirtxml.DomainInterfaceModel{
				Type: "virtio",
			},
			Source: &libvirtxml.DomainInterfaceSource{
				Dev:  vif.Name,
				Mode: "bridge",
			},
		}
	*/

	inter := v1.Interface{}

	ifId := strings.Split(vif.Name, "-")
	inter.Type = "direct"
	inter.Source = v1.InterfaceSource{Device: "br-" + ifId[1], Mode: "passthrough"}
	inter.MAC = &v1.MAC{MAC: vif.Mac.String()}
	inter.Model = &v1.Model{Type: "virtio"}

	return &inter, nil
}

func (vif *PodNetworkInterface) createPodNetworkBridgeXML() (*string, error) {
	ipCopy := &net.IPNet{}

	errs := model.Copy(ipCopy, vif.IPAddr)
	if errs != nil {
		for _, err := range errs {
			logging.DefaultLogger().Error().Reason(err).Msg("Failed to copy model")
		}
		return nil, errs[0]
	}
	dns := ipCopy.IP.To4()
	if dns != nil {
		dns[3] = 254 - dns[3]
	}

	fmt.Println(vif.IPAddr.IP.To4().String())
	fmt.Println(vif.IPAddr.Mask.String())
	l, u := vif.IPAddr.Mask.Size()
	fmt.Println(l)
	fmt.Println(u)
	ifId := strings.Split(vif.Name, "-")
	net := libvirtxml.Network{
		Name: "net-" + ifId[1],
		Bridge: &libvirtxml.NetworkBridge{
			Name: "br-" + ifId[1],
		},
		IPs: []libvirtxml.NetworkIP{
			libvirtxml.NetworkIP{
				Address:  vif.IPAddr.IP.To4().String(),
				Prefix:   "32",
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
