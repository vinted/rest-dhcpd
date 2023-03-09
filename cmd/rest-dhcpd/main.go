package main

import (
	"flag"
	"log"

	"github.com/vinted/rest-dhcpd/pkg/configdb"
	"github.com/vinted/rest-dhcpd/pkg/dhcpd"
	"github.com/vinted/rest-dhcpd/pkg/rest"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "configPath", "/etc/rest-dhcpd", "Path to config directory.")
	flag.Parse()
	err := configdb.Init(configPath)
	if err != nil {
		log.Fatalf("Failed to read configuration. %v", err)
	}
	go rest.StartServer()
	dhcpd.StartServer()
}
