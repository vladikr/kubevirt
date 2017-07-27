package cniproxy

import (
	"fmt"
	"sort"

	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/golang/glog"

	"strings"

	"kubevirt.io/kubevirt/pkg/logging"
	"kubevirt.io/kubevirt/pkg/virt-handler/utils"
)

//TODO: Make it configurable
const (
	CNINetDir     = "/etc/cni/net.d"
	CNIPluginsDir = "/opt/cni/bin"
	LibvirtSocket = "/var/run/libvirt/libvirt-sock"
)

type CNIProxy struct {
	cniConfig   *libcni.CNIConfig
	netConfig   *libcni.NetworkConfig
	runtimeConf *libcni.RuntimeConf
}

func getCNINetworkConfig() (*libcni.NetworkConfig, error) {
	files, err := libcni.ConfFiles(CNINetDir, []string{".conf"})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	for _, confFile := range files {
		conf, err := libcni.ConfFromFile(confFile)
		if err != nil {
			glog.Warningf("Error loading CNI config file %s: %v", confFile, err)
			continue
		}
		return conf, nil
	}
	return nil, fmt.Errorf("No valid networks found in %s", CNINetDir)
}

func GetProxy(runtime *libcni.RuntimeConf) (*CNIProxy, error) {

	conf, err := getCNINetworkConfig()
	if err != nil {
		return nil, err
	}

	cniconf := &libcni.CNIConfig{Path: []string{CNIPluginsDir}}
	cniProxy := &CNIProxy{netConfig: conf, cniConfig: cniconf, runtimeConf: runtime}
	return cniProxy, nil
}

// TODO(vladikr): re-arrange this..
func GetLibvirtNS() (*utils.NSResult, error) {
	pid, err := utils.GetPid(LibvirtSocket)
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msgf("Cannot find libvirt socket in %s", LibvirtSocket)
		return nil, err
	}
	logging.DefaultLogger().Debug().Msgf("Got libvirt pid %d", pid)
	NS := utils.GetNSFromPid(pid)
	return NS, nil
}

func BuildRuntimeConfig(ifname string) (*libcni.RuntimeConf, error) {
	libvNS, err := GetLibvirtNS()
	if err != nil {
		return nil, err
	}
	logging.DefaultLogger().Error().Reason(err).Msgf("Got namespace path from libvirt pid: %s", libvNS.Net)
	randId := strings.Split(ifname, "-")
	return &libcni.RuntimeConf{
		ContainerID: randId[1],
		NetNS:       libvNS.Net,
		IfName:      ifname,
	}, nil
}

func (proxy *CNIProxy) AddToNetwork() (types.Result, error) {
	res, err := proxy.cniConfig.AddNetwork(proxy.netConfig, proxy.runtimeConf)
	if err != nil {
		logging.DefaultLogger().Error().Reason(err).Msgf("Error creating an interface: %s", proxy.runtimeConf.IfName)
		return nil, err
	}

	return res, nil
}

func (proxy *CNIProxy) DeleteFromNetwork() error {
	err := proxy.cniConfig.DelNetwork(proxy.netConfig, proxy.runtimeConf)
	if err != nil {
		glog.Errorf("Error deleting an interface: %v", err)
		return err
	}
	return nil
}
