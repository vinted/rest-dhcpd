package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	dhcpdiscoverRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dhcpddiscover_requests_total",
		Help: "The total number of DHCPDISCOVER requests",
	})
)

var (
	dhcprequestRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dhcpdrequest_requests_total",
		Help: "The total number of DHCPREQUEST requests",
	})
)

var (
	dhcpofferReplies = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dhcpoffer_replies_total",
		Help: "The total number of DHCPOFFER replies",
	})
)

var (
	dhcpackReplies = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dhcpack_replies_total",
		Help: "The total number of ACK replies",
	})
)

var (
	dhcpnakReplies = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dhcpnak_replies_total",
		Help: "The total number of NAK replies",
	})
)

var (
	dhcpnosuchleaseRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dhcp_no_such_lease_requests_total",
		Help: "The total number of unkown lease requests",
	})
)

func UpdateDHCPDiscover() {
	dhcpdiscoverRequests.Inc()
}

func UpdateDHCPRequest() {
	dhcprequestRequests.Inc()
}

func UpdateDHCPOffer() {
	dhcpofferReplies.Inc()
}

func UpdateDHCPACK() {
	dhcpackReplies.Inc()
}

func UpdateDHCPNAK() {
	dhcpnakReplies.Inc()
}

func UpdateDHCPNoSuchLease() {
	dhcpnosuchleaseRequests.Inc()
}
