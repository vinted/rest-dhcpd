package dhcpd

import (
	"fmt"
	dhcp "github.com/krolaw/dhcp4"
	"github.com/krolaw/dhcp4/conn"
	"github.com/samber/lo"
	"github.com/vinted/rest-dhcpd/pkg/configdb"
	"github.com/vinted/rest-dhcpd/pkg/prometheus"
	"github.com/vinted/rest-dhcpd/pkg/rest"
	"log"
	"net"
	"reflect"
	"strconv"
	"time"
)

type DHCPHandler struct {
	ip            net.IP
	leaseDuration time.Duration
	options       dhcp.Options
}

var data_type = map[int]string{
	// List of Options that stores IP address represented in 4 bytes of data (without dots).
	1:  "IP",
	2:  "IP",
	3:  "IP",
	4:  "IP",
	5:  "IP",
	6:  "IP",
	7:  "IP",
	8:  "IP",
	9:  "IP",
	10: "IP",
	11: "IP",
}

func StartServer() {
	lease, _ := time.ParseDuration(strconv.Itoa(configdb.Config.LeaseDuration) + "s")
	handler := &DHCPHandler{
		ip:            net.ParseIP(configdb.Config.IP),
		leaseDuration: lease,
		options:       BuildOptions(configdb.Config.Options),
	}
	cn, err := conn.NewUDP4BoundListener(configdb.Config.ListenInterface, ":67")
	if err != nil {
		log.Fatal("Failed to bind to interface. %w", err)
	}
	log.Fatal(dhcp.Serve(cn, handler))
}

func BuildOptions(options interface{}) dhcp.Options {
	opt := options.(map[string]interface{})
	dhcp_opt := dhcp.Options{}
	for key, value := range opt {
		var val []byte
		id, _ := strconv.Atoi(key)
		if data_type[id] == "IP" {
			if reflect.ValueOf(value).Kind() == reflect.Slice {
				for _, ip := range value.([]interface{}) {
					val = append(val, []byte(net.ParseIP(fmt.Sprint(ip)).To4())...)
				}
			} else {
				val = []byte(net.ParseIP(fmt.Sprint(value)).To4())
			}
		} else {
			val = []byte(fmt.Sprint(value))
		}
		dhcp_opt[dhcp.OptionCode(id)] = []byte(val)
	}
	return dhcp_opt
}

func (h *DHCPHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
	switch msgType {

	case dhcp.Discover:
		log.Printf("DHCPDISCOVER from %s", p.CHAddr().String())
		prometheus.UpdateDHCPDiscover()
		client, id := rest.SearchForClientByMac(p.CHAddr().String())
		if id != -1 { // If client exists in configdb, send a DHCPOFFER
			h.options = lo.Assign(h.options, BuildOptions(client.Options)) // Merge global and client DHCP options
			h.options[dhcp.OptionHostName] = []byte(client.Hostname)       // Set hostname via DHCP options
			log.Printf("Sending DHCPOFFER to %s with IP: %s.", p.CHAddr().String(), client.IP)
			prometheus.UpdateDHCPOffer()
			return dhcp.ReplyPacket(p, dhcp.Offer, h.ip, net.ParseIP(client.IP), h.leaseDuration,
				h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
		} else {
			log.Printf("Server %s is not in configdb.", p.CHAddr().String())
			prometheus.UpdateDHCPNoSuchLease()
		}
	case dhcp.Request:
		prometheus.UpdateDHCPRequest()
		if server, ok := options[dhcp.OptionServerIdentifier]; ok && !net.IP(server).Equal(h.ip) {
			return nil // Message not for this dhcp server
		}
		client, id := rest.SearchForClientByMac(p.CHAddr().String())
		if id != -1 { // If client exists in configdb, send a DHCPACK
			reqIP := net.IP(options[dhcp.OptionRequestedIPAddress])
			if reqIP == nil { // DHCPREQUEST does not have IP address if it's a lease renew. IP address is set in CIADDR space instead.
				reqIP = net.IP(p.CIAddr())
			}
			log.Printf("DHCPREQUEST from %s: IP: %s.", p.CHAddr().String(), reqIP.String())
			if client.IP == reqIP.String() {
				log.Printf("Sending DHCPACK to %s with IP: %s.", p.CHAddr().String(), reqIP.String())
				prometheus.UpdateDHCPACK()
				return dhcp.ReplyPacket(p, dhcp.ACK, h.ip, reqIP, h.leaseDuration,
					h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			} else {
				log.Printf("Requested IP %s from %s does not match configdb, sending NAK.", reqIP.String(), p.CHAddr().String())
				prometheus.UpdateDHCPNAK()
				return dhcp.ReplyPacket(p, dhcp.NAK, h.ip, nil, 0, nil)
			}
		} else {
			log.Printf("Server %s is not in configdb.", p.CHAddr().String())
			prometheus.UpdateDHCPNoSuchLease()
		}
	case dhcp.Release:
		log.Printf("DHCPRELEASE from %s. Ignoring.", p.CHAddr().String())
	case dhcp.Decline:
		log.Printf("DHCPDECLINE from %s. Ignoring.", p.CHAddr().String())
	}
	return nil
}
