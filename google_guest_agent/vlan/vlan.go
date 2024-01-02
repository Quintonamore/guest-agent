//  Copyright 2023 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vlan

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/guest-agent/google_guest_agent/run"
	"github.com/GoogleCloudPlatform/guest-agent/metadata"
	"github.com/GoogleCloudPlatform/guest-logging-go/logger"
)

type vlanInterface interface {
	runWithOutput(ctx context.Context, name string, args ...string) *run.Result
}

type InterfaceDescriptor struct {
	Ifname   string      `json:"ifname"`
	Link     string      `json:"link"`
	Address  string      `json:"address"`
	Flags    []string    `json:"flags"`
	LinkInfo LinkInfo    `json:"linkinfo"`
	Mtu      json.Number `json:"mtu"`
}

type LinkInfo struct {
	InfoKind string       `json:"info_kind"`
	InfoData LinkInfoData `json:"info_data"`
}

type LinkInfoData struct {
	Protocol string      `json:"protocol"`
	Id       json.Number `json:"id"`
}

type vlan struct{}

// Configure finds VLAN NICs to add and remove from passed metadata.
func (v *vlan) Configure(ctx context.Context, vMeta []metadata.VlanNetworkInterfaces, vi vlanInterface) {
	if vlanNotSupported(ctx, vi) {
		logger.Infof("Module 802.1Q is not installed. Skipping VLAN NIC configuration.")
		return
	}
	local, err := getLocalVlanConfig(ctx, vi)
	if err != nil {
		logger.Errorf("getting local VLAN NIC config failed: %v", err)
		return
	}
	vlanNicsToAdd, vlanNicsToRemove := vlanNicsDiff(ctx, local, vMeta, vi)
	for _, vnic := range vlanNicsToAdd {
		addVlanNic(ctx, vnic, vi)
	}
	for _, vnic := range vlanNicsToRemove {
		removeVlanNic(ctx, vnic, vi)
	}
}

func vlanNicsDiff(ctx context.Context, local []InterfaceDescriptor, vMeta []metadata.VlanNetworkInterfaces, vi vlanInterface) (vlanNicsToAdd []metadata.VlanNetworkInterfaces, vlanNicsToRemove []InterfaceDescriptor) {
	ref := make(map[string]bool)
	lMap := make(map[string]InterfaceDescriptor)
	mMap := make(map[string]metadata.VlanNetworkInterfaces)
	for _, nic := range local {
		id := fmt.Sprintf("%s%d", nic.Link, nic.LinkInfo.InfoData.Id)
		lMap[id] = nic
		ref[id] = true
	}
	for _, nic := range vMeta {
		id := fmt.Sprintf("%s%d", nic.ParentInterface, nic.Vlan)
		mMap[id] = nic
		ref[id] = true
	}
	for _, id := range ref {
		l, inLocal := lMap[id]
		m, inMetadata := mMap[id]
		// VLAN exists locally but not in metadata; remove
		if inLocal && !inMetadata {
			vlanNicsToRemove := append(vlanNicsToRemove, l)
		}
		// VLAN exists in metadata but not locally; add
		if !inLocal && inMetadata {
			vlanNicsToAdd := append(vlanNicsToAdd, m)
		}
	}
	return
}

func addVlanNic(ctx context.Context, iface metadata.VlanNetworkInterfaces, vi vlanInterface) error {
	return nil
}

func removeVlanNic(ctx context.Context, iface InterfaceDescriptor, vi vlanInterface) error {
	return nil
}

func vlanNotSupported(ctx context.Context, vi vlanInterface) bool {
	out := vi.runWithOutput(ctx, "modinfo", "8021q")
	return out.ExitCode != 0
}

func getLocalVlanConfig(ctx context.Context, vi vlanInterface) ([]InterfaceDescriptor, error) {
	args := fmt.Sprintf("-d -j -6 link show")
	out := vi.runWithOutput(ctx, "ip", strings.Split(args, " ")...)
	if out.ExitCode != 0 {
		return nil, error(out)
	}
	res, err := unmarshalIfaceJSON([]byte(out.StdOut))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func unmarshalIfaceJSON(data []byte) ([]InterfaceDescriptor, error) {
	var ret []InterfaceDescriptor
	err := json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
