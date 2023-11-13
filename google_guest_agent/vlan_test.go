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
			data: `[{
    "ifindex": 3,
    "link": "ens4",
    "ifname": "ens4.5",
    "flags": [
      "BROADCAST",
      "MULTICAST"
    ],
    "mtu": 1460,
    "qdisc": "noop",
    "operstate": "DOWN",
    "linkmode": "DEFAULT",
    "group": "default",
    "txqlen": 1000,
    "link_type": "ether",
    "address": "42:01:0a:00:04:02",
    "broadcast": "ff:ff:ff:ff:ff:ff",
    "promiscuity": 0,
    "min_mtu": 0,
    "max_mtu": 65535,
    "linkinfo": {
      "info_kind": "vlan",
      "info_data": {
        "protocol": "802.1Q",
        "id": 5,
        "flags": [
          "REORDER_HDR"
        ]
      }
    },
    "inet6_addr_gen_mode": "eui64",
    "num_tx_queues": 1,
    "num_rx_queues": 1,
    "gso_max_size": 65536,
    "gso_max_segs": 65535
  }
]`,
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
