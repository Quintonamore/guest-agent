package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/guest-agent/google_guest_agent/run"
)

type vlanTesting struct{}

var exitCode = 0
var stdOut = ""

// Testing Interface with replacements for link, ifname, mtu
// address, and VLAN ID
var testIfaceJSON = `[{
    "ifindex": 1,
    "link": "%s",
    "ifname": "%s",
    "flags": [
      "BROADCAST",
      "MULTICAST"
    ],
    "mtu": %s,
    "link_type": "ether",
    "address": "%s",
    "linkinfo": {
      "info_kind": "vlan",
      "info_data": {
        "protocol": "802.1Q",
        "id": %s
      }
    }
  }]`

func (v *vlanTesting) runWithOutput(ctx context.Context, name string, args ...string) *run.Result {
	return &run.Result{ExitCode: exitCode, StdOut: stdOut}
}

func TestVlanNotSupportedSuccess(t *testing.T) {
	ctx := context.Background()
	vt := &vlanTesting{}
	res := vlanNotSupported(ctx, vt)
	if res {
		t.Errorf("vlanNotSupported() got: %v, want: true", res)
	}
}

func TestVlanNotSupportedFailure(t *testing.T) {
	ctx := context.Background()
	vt := &vlanTesting{}
	exitCode = 1
	res := vlanNotSupported(ctx, vt)
	if !res {
		t.Errorf("vlanNotSupported() got: %v, want: false", res)
	}
}

func TestGetLocalVlanConfig(t *testing.T) {
	var tests = []struct {
		name, out string
		exit      int
		want      []InterfaceDescriptor
	}{
		{
			name: "standard local config",
			out:  fmt.Sprintf(testIfaceJSON, "ens4", "ens4.4", "1460", "42:01:0a:00:04:02", "4"),
			want: []InterfaceDescriptor{
				InterfaceDescriptor{
					Ifname:  "ens4.4",
					Link:    "ens4",
					Address: "42:01:0a:00:04:02",
					Flags:   []string{"BROADCAST", "MULTICAST"},
					Mtu:     json.Number("1460"),
					LinkInfo: LinkInfo{
						InfoKind: "vlan",
						InfoData: LinkInfoData{
							Protocol: "802.1Q",
							Id:       json.Number("4"),
						},
					},
				},
			},
		},
		{
			name: "max vlan id",
			out:  fmt.Sprintf(testIfaceJSON, "eth0", "eth0.4094", "1460", "42:01:0a:00:04:02", "4094"),
			want: []InterfaceDescriptor{
				InterfaceDescriptor{
					Ifname:  "eth0.4094",
					Link:    "eth0",
					Address: "42:01:0a:00:04:02",
					Flags:   []string{"BROADCAST", "MULTICAST"},
					Mtu:     json.Number("1460"),
					LinkInfo: LinkInfo{
						InfoKind: "vlan",
						InfoData: LinkInfoData{
							Protocol: "802.1Q",
							Id:       json.Number("4094"),
						},
					},
				},
			},
		},
		{
			name: "expect error with runWithOutput",
			exit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("getLocalVlanConfig: %s", tt.name), func(t *testing.T) {
			ctx := context.Background()
			vt := &vlanTesting{}
			exitCode = tt.exit
			stdOut = tt.out
			local, err := getLocalVlanConfig(ctx, vt)
			if err != nil && tt.want != nil {
				t.Errorf("getLocalVlanConfig() got an error: %v", err)
			}
			if err == nil && tt.want == nil {
				t.Errorf("getLocalVlanConfig() expected an error")
			}
			if !reflect.DeepEqual(local, tt.want) {
				t.Errorf("getLocalVlanConfig()\ngot: %v\nwant: %v", local, tt.want)
			}
		})
	}
}

func TestUnmarshalIfaceJSON(t *testing.T) {
	var tests = []struct {
		name, data string
		want       []InterfaceDescriptor
	}{
		{
			name: "Empty data",
			data: "[]",
			want: []InterfaceDescriptor{},
		},
		{
			name: "Unmarshal JSON Error",
			data: "",
		},
		{
			name: "Unmarshal single vlan interface",
			data: fmt.Sprintf(testIfaceJSON, "ens4", "ens4.5", "1460", "42:01:0a:00:04:02", "5"),
			want: []InterfaceDescriptor{
				InterfaceDescriptor{
					Ifname:  "ens4.5",
					Link:    "ens4",
					Address: "42:01:0a:00:04:02",
					Flags:   []string{"BROADCAST", "MULTICAST"},
					Mtu:     json.Number("1460"),
					LinkInfo: LinkInfo{
						InfoKind: "vlan",
						InfoData: LinkInfoData{
							Protocol: "802.1Q",
							Id:       json.Number("5"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalIfaceJSON([]byte(tt.data))
			if err == nil && tt.want == nil {
				t.Errorf("unmarshalIfaceJSON() expected an error")
			}
			if err != nil && tt.want != nil {
				t.Errorf("unmarshalIfaceJSON() got an error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Did not parse expected interface data.\nGot:%v\nWant:%v", got, tt.want)
			}
		})
	}
}
