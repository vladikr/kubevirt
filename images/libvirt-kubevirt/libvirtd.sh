#!/bin/bash
#
# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright 2017 Red Hat, Inc.
#

set -xe

# HACK
# Use hosts's /dev to see new devices and allow macvtap
mkdir /dev.container && {
  mount --rbind /dev /dev.container
  mount --rbind /host-dev /dev

  # Keep some devices from the containerinal /dev
  keep() { mount --rbind /dev.container/$1 /dev/$1 ; }
  keep shm
  keep mqueue
  # Keep ptmx/pts for pty creation
  keep pts
  mount --rbind /dev/pts/ptmx /dev/ptmx
  # Use the container /dev/kvm if available
  [[ -e /dev.container/kvm ]] && keep kvm
}

mkdir /sys.net.container && {
  mount --rbind /sys/class/net /sys.net.container
  mount --rbind /host-sys/class/net /sys/class/net
}

mkdir /sys.devices.container && {
  mount --rbind /sys/devices /sys.devices.container
  mount --rbind /host-sys/devices /sys/devices
}

# If no cpuacct,cpu is present, symlink it to cpu,cpuacct
# Otherwise libvirt and our emulator get confused
if [ ! -d "/host-sys/fs/cgroup/cpuacct,cpu" ]; then
  echo "Creating cpuacct,cpu cgroup symlink"
  mount -o remount,rw /host-sys/fs/cgroup
  cd /host-sys/fs/cgroup
  ln -s cpu,cpuacct cpuacct,cpu
  mount -o remount,ro /host-sys/fs/cgroup
fi

mount --rbind /host-sys/fs/cgroup /sys/fs/cgroup

mkdir -p /var/log/kubevirt
touch /var/log/kubevirt/qemu-kube.log
chown qemu:qemu /var/log/kubevirt/qemu-kube.log

# We create the network on a file basis to not
# have to wait for libvirtd to come up
if [[ -n "$LIBVIRTD_DEFAULT_NETWORK_DEVICE" ]]; then
  mkdir -p /etc/libvirt/qemu/networks/autostart
  cat > /etc/libvirt/qemu/networks/default.xml <<EOX
<!-- Generated by libvirtd.sh container script -->
<network>
  <name>default</name>
  <forward mode="bridge">
    <interface dev="$LIBVIRTD_DEFAULT_NETWORK_DEVICE" />
  </forward>
</network>
EOX
  ln -s /etc/libvirt/qemu/networks/default.xml /etc/libvirt/qemu/networks/autostart/default.xml
fi

echo "cgroup_controllers = [ ]" >> /etc/libvirt/qemu.conf
#echo -e "log_level = 1" >> /etc/libvirt/libvirtd.conf

if [[ -n "$LIBVIRTD_DISABLE_TCP" ]]; then
  /usr/sbin/libvirtd
else
  /usr/sbin/libvirtd -l
fi
