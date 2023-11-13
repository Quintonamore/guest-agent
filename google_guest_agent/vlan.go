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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/guest-agent/google_guest_agent/run"
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

func (v *vlan) vlanSupported(ctx context.Context, vi vlanInterface) bool {
	out := vi.runWithOutput(ctx, "modinfo", "8021q")
	return out.ExitCode == 0
}

// Configure finds VLAN NICs to add and remove from passed metadata.
func (v *vlan) Configure(ctx context.Context) {

}

func (v *vlan) findAdditions(ctx context.Context, vi vlanInterface) []InterfaceDescriptor {
	return nil
}

func (v *vlan) findRemovals(ctx context.Context, vi vlanInterface) []InterfaceDescriptor {
	return nil
}

func (v *vlan) addVlanNic(ctx context.Context, vi vlanInterface) error {
	return nil
}

func (v *vlan) removeVlanNic(ctx context.Context, vi vlanInterface) error {
	return nil
}

func (v *vlan) getLocalVlanConfig(ctx context.Context, vi vlanInterface) ([]InterfaceDescriptor, error) {
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
