package main

import (
	"flag"
	"github.com/vinted/rest-dhcpd/pkg/configdb"
	"github.com/vinted/rest-dhcpd/pkg/dhcpd"
	"github.com/vinted/rest-dhcpd/pkg/rest"
	"log"
	"os"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "configPath", "/etc/rest-dhcpd", "Path to config directory.")
	flag.Parse()
	err := configdb.Init(configPath)
	if err != nil {
		log.Fatal("Failed to read configuration. %w", err)
		os.Exit(1)
	}
	go rest.StartServer()
	dhcpd.StartServer()
}
