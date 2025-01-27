package netbox_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/netboxlabs/diode/diode-server/netbox"
)

func TestIpamIPAddress_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    *netbox.IpamIPAddress
		wantErr bool
	}{
		{
			name: "valid IP address with interface",
			json: `{
				"id": 1,
				"address": "192.168.1.1",
				"assigned_object": {
					"interface": {
						"id": 123,
						"name": "eth0"
					}
				},
				"status": "active",
				"dns_name": "test.example.com"
			}`,
			want: &netbox.IpamIPAddress{
				ID:      1,
				Address: "192.168.1.1",
				AssignedObject: &netbox.IPAddressInterface{
					Interface: &netbox.DcimInterface{
						ID:   123,
						Name: "eth0",
					},
				},
				Status:  stringPtr("active"),
				DNSName: stringPtr("test.example.com"),
			},
			wantErr: false,
		},
		{
			name: "valid IP address without assigned object",
			json: `{
				"id": 2,
				"address": "192.168.1.2",
				"status": "active"
			}`,
			want: &netbox.IpamIPAddress{
				ID:      2,
				Address: "192.168.1.2",
				Status:  stringPtr("active"),
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			json:    `{"id": 1, "address": }`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got netbox.IpamIPAddress
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, &got)
		})
	}
}

// stringPtr returns a pointer to the string value passed in
func stringPtr(s string) *string {
	return &s
}
