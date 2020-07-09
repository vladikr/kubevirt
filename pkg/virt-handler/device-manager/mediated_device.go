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
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"kubevirt.io/client-go/log"
	"kubevirt.io/kubevirt/pkg/util"
	pluginapi "kubevirt.io/kubevirt/pkg/virt-handler/device-manager/deviceplugin/v1beta1"
)

const (
	mdevBasePath = "/sys/bus/mdev/devices"
)

type MDEV struct {
	UUID             string
	typeName         string
	parentPciAddress string
	numaNode         int
}

type MediatedDevicePlugin struct {
	devs       []*pluginapi.Device
	server     *grpc.Server
	socketPath string
	stop       chan struct{}
	health     chan string
	unhealthy  chan string
	devicePath string
	deviceName string
	done       chan struct{}
}

func NewMediatedDevicePlugin(mdevs []*MDEV, resourceName string) *MediatedDevicePlugin {
	s := strings.Split(resourceName, "/")
	mdevTypeName := s[1]
	serverSock := SocketPath(mdevTypeName)

	devs := constructDPIdevicesFromMdev(mdevs)
	dpi := &MediatedDevicePlugin{
		devs:       devs,
		socketPath: serverSock,
		health:     make(chan string),
		deviceName: resourceName,
		devicePath: mdevBasePath,
	}

	return dpi
}

func constructDPIdevicesFromMdev(mdevs []*MDEV) (devs []*pluginapi.Device) {
	for _, mdev := range mdevs {
		dpiDev := &pluginapi.Device{
			ID:     string(mdev.UUID),
			Health: pluginapi.Healthy,
		}
		if mdev.numaNode > 0 {
			numaInfo := &pluginapi.NUMANode{
				ID: int64(mdev.numaNode),
			}
			dpiDev.Topology = &pluginapi.TopologyInfo{
				Nodes: []*pluginapi.NUMANode{numaInfo},
			}
		}
		devs = append(devs, dpiDev)
	}
	return

}

func (dpi *MediatedDevicePlugin) GetDevicePath() string {
	return dpi.devicePath
}

func (dpi *MediatedDevicePlugin) GetDeviceName() string {
	return dpi.deviceName
}

// Start starts the device plugin
func (dpi *MediatedDevicePlugin) Start(stop chan struct{}) (err error) {
	logger := log.DefaultLogger()
	dpi.stop = stop
	dpi.done = make(chan struct{})

	err = dpi.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", dpi.socketPath)
	if err != nil {
		return fmt.Errorf("error creating GRPC server socket: %v", err)
	}

	dpi.server = grpc.NewServer([]grpc.ServerOption{}...)
	defer dpi.Stop()

	pluginapi.RegisterDevicePluginServer(dpi.server, dpi)
	err = dpi.Register()
	if err != nil {
		return fmt.Errorf("error registering with device plugin manager: %v", err)
	}

	errChan := make(chan error, 2)

	go func() {
		errChan <- dpi.server.Serve(sock)
	}()

	err = waitForGrpcServer(dpi.socketPath, connectionTimeout)
	if err != nil {
		return fmt.Errorf("error starting the GRPC server: %v", err)
	}

	go func() {
		errChan <- dpi.healthCheck()
	}()

	logger.Infof("%s device plugin started", dpi.deviceName)
	err = <-errChan

	return err
}

// This will need to be extended
func (dpi *MediatedDevicePlugin) healthCheck() error {
	_, err := os.Stat(dpi.socketPath)
	if err != nil {
		return fmt.Errorf("failed to stat the device-plugin socket: %v", err)
	}

	for {
		select {
		case <-dpi.stop:
			return nil
		}
	}
}

// Stop stops the gRPC server
func (dpi *MediatedDevicePlugin) Stop() error {
	defer close(dpi.done)
	dpi.server.Stop()
	return dpi.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (dpi *MediatedDevicePlugin) Register() error {
	conn, err := connect(pluginapi.KubeletSocket, connectionTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(dpi.socketPath),
		ResourceName: dpi.deviceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

func (dpi *MediatedDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	// FIXME: sending an empty list up front should not be needed. This is a workaround for:
	// https://github.com/kubevirt/kubevirt/issues/1196
	// This can safely be removed once supported upstream Kubernetes is 1.10.3 or higher.
	emptyList := []*pluginapi.Device{}
	s.Send(&pluginapi.ListAndWatchResponse{Devices: emptyList})

	s.Send(&pluginapi.ListAndWatchResponse{Devices: dpi.devs})

	for {
		select {
		case <-dpi.stop:
			return nil
		case <-dpi.done:
			return nil
		}
	}
}

func (dpi *MediatedDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	resourceName := dpi.deviceName
	resourceNameEnvVar := util.ResourceNameToEnvvar("MDEV_PCI_RESOURCE", resourceName)
	allocatedDevices := []string{}
	resp := new(pluginapi.AllocateResponse)
	containerResponse := new(pluginapi.ContainerAllocateResponse)

	for _, request := range r.ContainerRequests {
		deviceSpecs := make([]*pluginapi.DeviceSpec, 0)
		for _, devID := range request.DevicesIDs {
			iommuGroup, err := getDeviceIOMMUGroup(mdevBasePath, devID)
			if err != nil {
				continue
			}
			allocatedDevices = append(allocatedDevices, devID)
			deviceSpecs = append(deviceSpecs, formatVFIODeviceSpecs(iommuGroup)...)
		}
		envVar := make(map[string]string)
		envVar[resourceNameEnvVar] = strings.Join(allocatedDevices, ",")
		containerResponse.Envs = envVar
		containerResponse.Devices = deviceSpecs
		resp.ContainerResponses = append(resp.ContainerResponses, containerResponse)
	}
	return resp, nil
}

func (dpi *MediatedDevicePlugin) cleanup() error {
	if err := os.Remove(dpi.socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (dpi *MediatedDevicePlugin) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	options := &pluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}
	return options, nil
}

func (dpi *MediatedDevicePlugin) PreStartContainer(ctx context.Context, in *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	res := &pluginapi.PreStartContainerResponse{}
	return res, nil
}

func discoverPermittedHostMediatedDevices(supportedMdevsMap map[string]string) map[string][]*MDEV {

	mdevsMap := make(map[string][]*MDEV)
	err := filepath.Walk(mdevBasePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		mdevTypeName, err := getMdevTypeName(info.Name())
		if err != nil {
			log.DefaultLogger().Reason(err).Errorf("failed read type name for mdev: %s", info.Name())
			return nil
		}
		if _, supported := supportedMdevsMap[mdevTypeName]; supported {

			mdev := &MDEV{
				typeName: mdevTypeName,
				UUID:     info.Name(),
			}
			parentPCIAddr, err := getMdevParentPCIAddr(info.Name())
			if err != nil {
				log.DefaultLogger().Reason(err).Errorf("failed parent PCI address for mdev: %s", info.Name())
				return nil
			}

			mdev.numaNode = getDeviceNumaNode(pciBasePath, parentPCIAddr)
			mdevsMap[mdevTypeName] = append(mdevsMap[mdevTypeName], mdev)
		}
		return nil
	})
	if err != nil {
		log.DefaultLogger().Reason(err).Errorf("failed to discover mediated devices")
	}
	return mdevsMap
}

// /sys/class/mdev_bus/0000:00:03.0/53764d0e-85a0-42b4-af5c-2046b460b1dc
func getMdevParentPCIAddr(mdevUUID string) (string, error) {
	mdevLink, err := os.Readlink(filepath.Join(mdevBasePath, mdevUUID))
	if err != nil {
		return "", err
	}
	linkParts := strings.Split(mdevLink, "/")
	return linkParts[len(linkParts)-2], nil
}

func getMdevTypeName(mdevUUID string) (string, error) {
	rawName, err := ioutil.ReadFile(filepath.Join(mdevBasePath, mdevUUID, "mdev_type/name"))
	if err != nil {
		return "", err
	}
	// The name usually contain spaces which should be replaced with _
	typeNameStr := strings.Replace(string(rawName), " ", "_", -1)
	typeNameStr = strings.TrimSpace(typeNameStr)
	return typeNameStr, nil
}
