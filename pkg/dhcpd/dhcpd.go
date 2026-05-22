package dhcpd

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"time"

	dhcp "github.com/krolaw/dhcp4"
	"github.com/krolaw/dhcp4/conn"
	"github.com/samber/lo"
	"github.com/vinted/rest-dhcpd/pkg/configdb"
	"github.com/vinted/rest-dhcpd/pkg/prometheus"
	"github.com/vinted/rest-dhcpd/pkg/rest"
)

type DHCPHandler struct {
	interfaceName string
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
	globalOptions := BuildOptions(configdb.Config.Options)
	errCh := make(chan error, len(configdb.Config.Interfaces))

	for _, iface := range configdb.Config.Interfaces {
		lease, _ := time.ParseDuration(strconv.Itoa(iface.LeaseDuration) + "s")
		handler := &DHCPHandler{
			interfaceName: iface.Name,
			ip:            net.ParseIP(iface.IP).To4(),
			leaseDuration: lease,
			options:       lo.Assign(globalOptions, BuildOptions(iface.Options)),
		}
		cn, err := conn.NewUDP4BoundListener(iface.Name, ":67")
		if err != nil {
			log.Fatalf("Failed to bind to interface %s: %v", iface.Name, err)
		}
		log.Printf("[%s] Listening for DHCP requests (server IP %s).", iface.Name, handler.ip)
		go func(name string) {
			errCh <- fmt.Errorf("[%s] serve exited: %w", name, dhcp.Serve(cn, handler))
		}(iface.Name)
	}
	log.Fatal(<-errCh)
}

func BuildOptions(options interface{}) dhcp.Options {
	dhcp_opt := dhcp.Options{}
	if options == nil {
		return dhcp_opt
	}
	opt, ok := options.(map[string]interface{})
	if !ok {
		return dhcp_opt
	}
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
		log.Printf("[%s] DHCPDISCOVER from %s", h.interfaceName, p.CHAddr().String())
		prometheus.UpdateDHCPDiscover()
		client, id := rest.SearchForClientByMac(p.CHAddr().String())
		if id != -1 { // If client exists in configdb, send a DHCPOFFER
			var reqOptions dhcp.Options = lo.Assign(h.options, BuildOptions(client.Options))
			reqOptions[dhcp.OptionHostName] = []byte(client.Hostname)
			log.Printf("[%s] Sending DHCPOFFER to %s with IP: %s.", h.interfaceName, p.CHAddr().String(), client.IP)
			prometheus.UpdateDHCPOffer()
			return dhcp.ReplyPacket(p, dhcp.Offer, h.ip, net.ParseIP(client.IP), h.leaseDuration,
				reqOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
		} else {
			log.Printf("[%s] Client %s is not in configdb.", h.interfaceName, p.CHAddr().String())
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
			log.Printf("[%s] DHCPREQUEST from %s: IP: %s.", h.interfaceName, p.CHAddr().String(), reqIP.String())
			if client.IP == reqIP.String() {
				var reqOptions dhcp.Options = lo.Assign(h.options, BuildOptions(client.Options))
				reqOptions[dhcp.OptionHostName] = []byte(client.Hostname)
				log.Printf("[%s] Sending DHCPACK to %s with IP: %s.", h.interfaceName, p.CHAddr().String(), reqIP.String())
				prometheus.UpdateDHCPACK()
				return dhcp.ReplyPacket(p, dhcp.ACK, h.ip, reqIP, h.leaseDuration,
					reqOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			} else {
				log.Printf("[%s] Requested IP %s from %s does not match configdb, sending NAK.", h.interfaceName, reqIP.String(), p.CHAddr().String())
				prometheus.UpdateDHCPNAK()
				return dhcp.ReplyPacket(p, dhcp.NAK, h.ip, nil, 0, nil)
			}
		} else {
			log.Printf("[%s] Client %s is not in configdb.", h.interfaceName, p.CHAddr().String())
			prometheus.UpdateDHCPNoSuchLease()
		}
	case dhcp.Release:
		log.Printf("[%s] DHCPRELEASE from %s. Ignoring.", h.interfaceName, p.CHAddr().String())
	case dhcp.Decline:
		log.Printf("[%s] DHCPDECLINE from %s. Ignoring.", h.interfaceName, p.CHAddr().String())
	}
	return nil
}
