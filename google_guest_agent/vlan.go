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
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/guest-agent/google_guest_agent/run"
)

type vlanInterface interface {
	runWithOutput(ctx context.Context, name string, args ...string) *run.Result
}

type vlan struct{}

func (v *vlan) vlanSupported(ctx context.Context, vi vlanInterface) bool {
	out := vi.runWithOutput(ctx, "modinfo", "8021q")
	return out.ExitCode == 0
}

func (v *vlan) configure(ctx context.Context) {

}

func (v *vlan) findAdditions(ctx context.Context, vi vlanInterface) []string {
	return nil
}

func (v *vlan) findRemovals(ctx context.Context, vi vlanInterface) []string {
	return nil
}

func (v *vlan) addVlanNic(ctx context.Context, vi vlanInterface) bool {
	return false
}

func (v *vlan) removeVlanNic(ctx context.Context, vi vlanInterface) bool {
	return false
}

func (v *vlan) getLocalVlanConfig(ctx context.Context, vi vlanInterface) ([]string, error) {
	var res []string
	args := fmt.Sprintf("link show vlan")
	out := vi.runWithOutput(ctx, "ip", strings.Split(args, " ")...)
	if out.ExitCode != 0 {
		return nil, error(out)
	}
	args = fmt.Sprintf("-6 link show vlan")
	outIpv6 := vi.runWithOutput(ctx, "ip", strings.Split(args, " ")...)
	if outIpv6.ExitCode != 0 {
		return nil, error(out)
	}
	allOut := fmt.Sprintf("%s\n%s", out.StdOut, outIpv6.StdOut)
	for _, line := range strings.Split(allOut, "\n") {
		if line != "" {
			res = append(res, line)
		}
	}
	return res, nil
}
