package cniproxy

import (
	"fmt"
	"sort"

	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/golang/glog"

	"kubevirt.io/kubevirt/pkg/virt-handler/utils"
	//"golang.org/x/net/proxy"
	//"github.com/vishvananda/netns"
	"io/ioutil"
	"os"
	"path/filepath"
)

//TODO: Make it configurable
const (
	CNINetDir     = "/etc/cni/net.d"
	CNIPluginsDir = "/opt/cni/bin"
	LibvirtSocket = "/var/run/libvirt/libvirt-sock"
	//LibvirtSocketRo = "/var/run/libvirt/libvirt-sock-ro"
	//LibvirtPidFile  = "/var/run/libvirtd.pid"
)

type CNIProxy struct {
	cniConfig   *libcni.CNIConfig
	netConfig   *libcni.NetworkConfig
	runtimeConf *libcni.RuntimeConf
}

func confFiles(dir string, extensions []string) ([]string, error) {
	// In part, adapted from rkt/networking/podenv.go#listFiles
	files, err := ioutil.ReadDir(dir)

	switch {
	case err == nil: // break
		glog.Errorln("err==nil")
	case os.IsNotExist(err):
		glog.Errorln("Doesnot exist: ", err)
		return nil, nil
	default:
		glog.Errorln("default: ", err)
		return nil, err
	}
	for _, f := range files {
		glog.Errorln("Got: ", f.Name())
	}

	confFiles := []string{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fileExt := filepath.Ext(f.Name())
		glog.Errorln("file ext: ", fileExt)
		for _, ext := range extensions {
			glog.Errorln("match ext: ", ext)
			if fileExt == ext {
				glog.Errorln("adding file")
				confFiles = append(confFiles, filepath.Join(dir, f.Name()))
			}
		}
	}
	return confFiles, nil
}

func getCNINetworkConfig() (*libcni.NetworkConfig, error) {
	files, err := libcni.ConfFiles(CNINetDir, []string{".conf"})

	switch {
	case err != nil:
		return nil, err
	case len(files) == 0:
		files, _ := ioutil.ReadDir("/etc/cni/net.d/")
		for _, f := range files {
			glog.Errorln(f.Name())
		}
		fls, _ := confFiles(CNINetDir, []string{})
		glog.Errorln("mutercycle: ", fls)
		return nil, fmt.Errorf("No networks found in %s", CNINetDir)
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

func GetCNIProxy() (*CNIProxy, error) {

	conf, err := getCNINetworkConfig()
	if err != nil {
		return nil, err
	}

	runtime, err := buildRuntimeConfig()
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
		glog.Errorf("Cannot find libvirt socket")
		return nil, err
	}

	glog.Errorf("Got libvirt pid %d", pid)
	NS := utils.GetNSFromPid(pid)

	return NS, nil
}

func buildRuntimeConfig() (*libcni.RuntimeConf, error) {
	/*
		files, _ := ioutil.ReadDir("/var/run/libvirt/")
		for _, f := range files {
			glog.Errorf(f.Name())
		}

		files1, _ := ioutil.ReadDir("/var/run/")
		for _, f := range files1 {
			glog.Errorf(f.Name())
		}

	*/
	/*
		pid, err := utils.GetLibvirtPidFromFile(LibvirtPidFile)
		if err != nil {
			glog.Errorf("Cannot find libvirt pid file: %s", LibvirtPidFile)
			pid, err := utils.GetPid(LibvirtSocket)
			if err != nil {
				glog.Errorf("Cannot find libvirt socket")
			}


			return nil, err
		}
	*/
	libvNS, err := GetLibvirtNS()
	if err != nil {
		return nil, err
	}
	glog.Errorf("Got namespace path from libvirt pid: %s", libvNS.Net)

	return &libcni.RuntimeConf{
		ContainerID: "00dba414cb54b03d74ca3bf216568d0284b51be39c242e0d8a237b7359788740",
		NetNS:       libvNS.Net,
		IfName:      "kveth0",
	}, nil
}

func (proxy *CNIProxy) AddToNetwork() (types.Result, error) {
	res, err := proxy.cniConfig.AddNetwork(proxy.netConfig, proxy.runtimeConf)
	if err != nil {
		glog.Errorf("Error adding network: %v", err)
		return nil, err
	}

	return res, nil
}

func (proxy *CNIProxy) DeleteFromNetwork() error {
	err := proxy.cniConfig.DelNetwork(proxy.netConfig, proxy.runtimeConf)
	if err != nil {
		glog.Errorf("Error deleting network: %v", err)
		return err
	}
	return nil
}
