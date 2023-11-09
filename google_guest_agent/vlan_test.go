package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/guest-agent/google_guest_agent/run"
)

type vlanTesting struct{}

var exitCode = 0
var stdOut = ""
var condFail = ""

func (v *vlanTesting) runWithOutput(ctx context.Context, name string, args ...string) *run.Result {
	for _, arg := range args {
		if condFail == arg {
			return &run.Result{ExitCode: 1, StdOut: stdOut}
		}
	}
	return &run.Result{ExitCode: exitCode, StdOut: stdOut}
}

func TestVlanSupported(t *testing.T) {
	ctx := context.Background()
	vlan := &vlan{}
	vt := &vlanTesting{}
	res := vlan.vlanSupported(ctx, vt)
	if !res {
		t.Errorf("vlan.vlanSupported() got: %v, want: true", res)
	}
}

func TestVlanNotSupported(t *testing.T) {
	ctx := context.Background()
	vlan := &vlan{}
	vt := &vlanTesting{}
	exitCode = 1
	res := vlan.vlanSupported(ctx, vt)
	if res {
		t.Errorf("vlan.vlanSupported() got: %v, want: false", res)
	}
}

func TestGetLocalVlanConfig(t *testing.T) {
	var tests = []struct {
		name, out, condFail string
		exit                int
		want                []string
	}{
		{
			name: "standard local config",
			out:  "vlan test",
			// We want double outputs here since getLocalVlanConfig makes
			// two calls to ip to get ipv6 configs
			want: []string{"vlan test", "vlan test"},
		},
		{
			name: "expect error with runWithOutput",
			exit: 1,
		},
		{
			name:     "expect error with -6 runWithOutput",
			condFail: "-6",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("getLocalVlanConfig: %s", tt.name), func(t *testing.T) {
			//TODO Test acutal 'ip link show vlan' outputs
			ctx := context.Background()
			vlan := &vlan{}
			vt := &vlanTesting{}
			exitCode = tt.exit
			stdOut = tt.out
			condFail = tt.condFail
			local, err := vlan.getLocalVlanConfig(ctx, vt)
			if err != nil && tt.want != nil {
				t.Errorf("vlan.getLocalVlanConfig() got an error: %v", err)
			}
			if err == nil && tt.want == nil {
				t.Errorf("vlan.getLocalVlanConfig() expected an error")
			}
			for i, val := range local {
				if val != tt.want[i] {
					t.Errorf("vlan.getLocalVlanConfig()\ngot: %v\nwant: %v", local, tt.want)
				}
			}
		})
	}
}
