package dhcpd

import (
	"bytes"
	dhcp "github.com/krolaw/dhcp4"
	"testing"
)

func TestBuildOptions(t *testing.T) {
	options := map[string]interface{}{}
	options["1"] = "255.255.255.240"
	options["3"] = "192.168.0.1"
	options["6"] = "8.8.8.8"
	options["12"] = "client-name"
	result := BuildOptions(options)
	if !bytes.Equal(result[dhcp.OptionHostName], []byte{99, 108, 105, 101, 110, 116, 45, 110, 97, 109, 101}) {
		t.Fatal("Hostname is invalid")
	}
	if !bytes.Equal(result[dhcp.OptionSubnetMask], []byte{255, 255, 255, 240}) {
		t.Fatal("Netmask is invalid")
	}
	if !bytes.Equal(result[dhcp.OptionRouter], []byte{192, 168, 0, 1}) {
		t.Fatal("Router address is invalid")
	}
	if !bytes.Equal(result[dhcp.OptionDomainNameServer], []byte{8, 8, 8, 8}) {
		t.Fatal("DNS address is invalid")
	}
}
