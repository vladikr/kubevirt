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
 * Copyright 2017, 2018 Red Hat, Inc.
 *
 */

package vmpodtransformer

import (
	_ "embed"
    "encoding/json"
	//"encoding/xml"
	//"errors"
	"fmt"
	//"io/fs"
	"os"
	//"path"
	//"path/filepath"
	//"strconv"
	//"strings"
	"bytes"

	//"sigs.k8s.io/yaml"
	"github.com/spf13/cobra"

	k8sv1 "k8s.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
	//"k8s.io/apimachinery/pkg/api/equality"
	//"k8s.io/apimachinery/pkg/api/resource"
	//k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/tools/cache"
	//"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"
    yml "k8s.io/apimachinery/pkg/util/yaml"
    extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/tools/clientcmd"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/kubevirt/pkg/network/vmispec"
	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
	"kubevirt.io/kubevirt/pkg/virt-api/webhooks"
	"kubevirt.io/kubevirt/pkg/virt-api/webhooks/mutating-webhook/mutators"
	"kubevirt.io/kubevirt/pkg/virt-controller/services"
	"kubevirt.io/kubevirt/pkg/virt-controller/watch"
	"kubevirt.io/kubevirt/pkg/virtctl/templates"

    "kubevirt.io/kubevirt/pkg/testutils"
)
const (
	COMMAND_VM2POD       = "vmToPod"
	filePathArg          = "file"
	filePathArgShort     = "f"
)

var (
	filePath     string
)

func NewVMToPodCommand(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vmToPod (VM)",
		Short:   "Return the Pod object with an embedded VirtualMachineInstance object.",
		Example: usageVmToPod(),
		Args:    cobra.MatchAll(cobra.ExactArgs(0), vmToPodArgs()),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := Command{clientConfig: clientConfig}
			return c.vmToPodRun(args, cmd)
		},
	}
	cmd.Flags().StringVarP(&filePath, filePathArg, filePathArgShort, "", "Path to the Virtual Machine spec.")
	//cmd.MarkFlagsMutuallyExclusive(filePathArg, vmArg)
	cmd.SetUsageTemplate(templates.UsageTemplate())
	return cmd
}

type Command struct {
	clientConfig clientcmd.ClientConfig
	command      string
}

func (o *Command) vmToPodRun(args []string, cmd *cobra.Command) error {

    crdInformer, _ := testutils.NewFakeInformerFor(&extv1.CustomResourceDefinition{})
    kubeVirtInformer, _ := testutils.NewFakeInformerFor(&v1.KubeVirt{}) 


    clusterConfig, err := virtconfig.NewClusterConfig(crdInformer, kubeVirtInformer, "kubevirt")
    if err != nil {
        panic(err) 
    }


    // read the provided VM yaml
    vm, err := readVMFromFile(filePath)
    if err != nil {
        return err
    }

    vm.ObjectMeta.Namespace = "default"
    // set VM defaults and exand it.
    if err := webhooks.SetDefaultVirtualMachine(clusterConfig, vm); err != nil {
        return err
    }

    // generate the VMI object from the VM
    vmi := watch.SetupVMIFromVM(vm)
	err = vmispec.SetDefaultNetworkInterface(clusterConfig, &vmi.Spec)
	if err != nil {
		return err
	}

    mutators.ApplyNewVMIMutations(vmi, clusterConfig)

    // generate the pod object

    resourceQuotaStore := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
    namespaceStore := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
    pvcCache := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, nil)
    var qemuGid int64 = 107
    //var defaultArch = "amd64"
    var virtclient kubecli.KubevirtClient
    svc := services.NewTemplateService("kubevirt/virt-launcher",
        240,
        "/var/run/kubevirt",
        "/var/lib/kubevirt",
        "/var/run/kubevirt-ephemeral-disks",
        "/var/run/kubevirt/container-disks",
        v1.HotplugDiskDir,
        "pull-secret-1",
        pvcCache,
        virtclient,
        clusterConfig,
        qemuGid,
        "kubevirt/vmexport",
        resourceQuotaStore,
        namespaceStore,
    )

    print(vmi)
    formatedOutput1, _ := yaml.Marshal(vmi)

	print(string(formatedOutput1))

//templateService services.TemplateService,
    templatePod, err := svc.RenderLaunchManifest(vmi)
//	if curPodImage != "" && curPodImage != c.templateService.GetLauncherImage() {

    // transform vmi to JSON
	vmiJSON, err := vmiToJSON(vmi)
	if err != nil {
		return err
	}
    
    for idx, container := range templatePod.Spec.Containers {
        if container.Name == "compute" {
            templatePod.Spec.Containers[idx].Env = append(container.Env, k8sv1.EnvVar{Name: "VMI_OBJ", Value: vmiJSON})
            break
        }   
    }
    


	output, err := applyOutputFormat(templatePod)
	if err != nil {
		return err
	}

	cmd.Print(output)
	return nil
}

func usageVmToPod() string {
	return `  #Expand a virtual machine called 'myvm'.
  {{ProgramName}} expand --vm myvm
  
  # Expand a virtual machine from file called myvm.yaml.
  {{ProgramName}} expand --file myvm.yaml

  # Expand a virtual machine called myvm and display output in json format.
  {{ProgramName}} expand --vm myvm --output json
  `
}

func vmToPodArgs() cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if filePath == "" {
			return fmt.Errorf("error invalid arguments - VirtualMachine file must be provided")
		}

		return nil
	}
}


func readVMFromFile(filePath string) (*v1.VirtualMachine, error) {
	vm := &v1.VirtualMachine{}

	readFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %+w", err)
	}

	err = yml.NewYAMLOrJSONDecoder(bytes.NewReader(readFile), 1024).Decode(&vm)
	if err != nil {
		return nil, fmt.Errorf("error decoding VirtualMachine %+w", err)
	}

	return vm, nil
}

func applyOutputFormat(pod *k8sv1.Pod) (string, error) {

    formatedOutput, err := yaml.Marshal(pod)

	if err != nil {
		return "", err
	}

	return string(formatedOutput), nil
}

func vmiToJSON(vmi *v1.VirtualMachineInstance) (string, error) {

    formatedOutput, err := json.MarshalIndent(vmi, "", " ")

	if err != nil {
		return "", err
	}

	return string(formatedOutput), nil
}

