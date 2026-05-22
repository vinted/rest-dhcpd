package configdb

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"sync"
)

type Mu struct {
	Mu sync.Mutex
}

type InterfaceConfig struct {
	Name          string
	IP            string
	LeaseDuration int
	Options       interface{}
}

type GlobalOptions struct {
	LeaseDuration      int
	AuthToken          string
	HTTPListenAddress  string
	TLSEnabled         bool
	TLSPrivateKeyFile  string
	TLSCertificateFile string
	Options            interface{}
	Interfaces         []InterfaceConfig

	// Deprecated single-interface fields. Kept for backwards compatibility
	// with pre-multi-interface configs. Use Interfaces[] instead.
	IP              string
	ListenInterface string
}

type Clients struct {
	Clients []Client
}

type Client struct {
	Hostname string
	MAC      string
	IP       string
	Options  interface{}
}

const (
	configFilename  = "rest-dhcpd-config.json"
	clientsFilename = "rest-dhcpd-clients.json"
)

var (
	DB         *Clients
	Config     *GlobalOptions
	ConfigPath string
)

func Init(configPath string) error {
	ConfigPath = configPath
	content := []byte(`{}`)
	dataFile := path.Join(configPath, clientsFilename)
	_, err := os.Stat(dataFile)
	if !os.IsNotExist(err) {
		content, err = os.ReadFile(dataFile)
		if err != nil {
			return err
		}
	}
	if len(content) == 0 {
		content = []byte(`{}`)
	}
	err = json.Unmarshal(content, &DB)
	if err != nil {
		return err
	}

	cfg, err := os.ReadFile(path.Join(configPath, configFilename))
	if err != nil {
		return err
	}
	err = json.Unmarshal(cfg, &Config)
	if err != nil {
		return err
	}
	if err := validateConfig(Config); err != nil {
		return err
	}
	log.Printf("DB init done.")
	return nil
}

func validateConfig(c *GlobalOptions) error {
	legacyPresent := c.ListenInterface != "" || c.IP != ""
	if len(c.Interfaces) > 0 && legacyPresent {
		return fmt.Errorf("config: cannot use both legacy IP/ListenInterface and Interfaces[]; pick one")
	}
	if len(c.Interfaces) == 0 && legacyPresent {
		log.Printf("config: top-level IP/ListenInterface are deprecated; please migrate to Interfaces[]")
		c.Interfaces = []InterfaceConfig{{Name: c.ListenInterface, IP: c.IP}}
	}
	if len(c.Interfaces) == 0 {
		return fmt.Errorf("config: Interfaces must contain at least one entry")
	}
	for i := range c.Interfaces {
		iface := &c.Interfaces[i]
		if iface.Name == "" {
			return fmt.Errorf("config: Interfaces[%d].Name is empty", i)
		}
		if net.ParseIP(iface.IP).To4() == nil {
			return fmt.Errorf("config: Interfaces[%d] (%s) has invalid IPv4 address %q", i, iface.Name, iface.IP)
		}
		if iface.LeaseDuration == 0 {
			iface.LeaseDuration = c.LeaseDuration
		}
		if iface.LeaseDuration == 0 {
			return fmt.Errorf("config: Interfaces[%d] (%s) has no LeaseDuration and no top-level default is set", i, iface.Name)
		}
	}
	return nil
}

func (m *Mu) Save() error {
	content, err := json.MarshalIndent(DB, "", "  ")
	if err != nil {
		log.Printf("%s", err)
	}
	m.Mu.Lock()
	err = os.WriteFile(path.Join(ConfigPath, clientsFilename), content, 0644)
	m.Mu.Unlock()
	return err
}
