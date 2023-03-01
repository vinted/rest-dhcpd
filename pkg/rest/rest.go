package rest

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vinted/rest-dhcpd/pkg/configdb"
	"log"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "<html><body>REST-DHCPD<br>https://github.com/vinted/rest-dhcpd</body></html>")
}

func clientConfigShow(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if checkAuthToken(w, r) {
		return
	}
	mac := params.ByName("mac")
	client, _ := SearchForClientByMac(mac)
	fmt.Fprintf(w, "%s\n", configToJson(client))
}

func clientConfigAdd(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if checkAuthToken(w, r) {
		return
	}
	put_request := configdb.Client{}
	err := json.NewDecoder(r.Body).Decode(&put_request)
	if err != nil {
		log.Printf("JSON marashall error: %s", err)
		return
	}
	put_request.MAC = params.ByName("mac")
	if validateClientConfig(w, r, put_request) {
		return
	}
	cl, id := SearchForClientByMac(put_request.MAC)
	if cl.MAC != "" {
		updateClientConfig(id, put_request)
	} else {
		addClientConfig(put_request)
	}
	fmt.Fprintf(w, "Updating config for: %s\n", params.ByName("mac"))
}

func clientConfigDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if checkAuthToken(w, r) {
		return
	}
	mac := params.ByName("mac")
	_, id := SearchForClientByMac(mac)
	if id == -1 {
		fmt.Fprintf(w, "No configuration to delete for: %s\n", mac)
	} else {
		deleteClientConfig(id)
		fmt.Fprintf(w, "Deleting config for MAC: %s, ID: %d\n", mac, id)
	}

}

func clientConfigShowFull(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if checkAuthToken(w, r) {
		return
	}
	clients, err := listAllClients()
	if err != nil {
		log.Printf("Got error listing clients: %s.", err)
	}
	fmt.Fprintf(w, "%s\n", clients)
}

func validateClientConfig(w http.ResponseWriter, r *http.Request, put_request configdb.Client) bool {
	cl, id := searchForClientByHostname(put_request.Hostname)
	if id != -1 && cl.MAC != put_request.MAC {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Hostname \"%s\" is already used by another client!\n", put_request.Hostname)
		return true
	}
	cl, id = searchForClientByIp(put_request.IP)
	if id != -1 && cl.MAC != put_request.MAC {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "IP \"%s\" is already used by another client!\n", put_request.IP)
		return true
	}
	return false
}

func checkAuthToken(w http.ResponseWriter, r *http.Request) bool {
	token := r.Header.Get("REST-DHCPD-Auth-Token")
	if token == configdb.Config.AuthToken {
		return false
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Auth token invalid\n")
	}
	return true
}

func deleteClientConfig(id int) {
	m := configdb.Mu{}
	m.Mu.Lock()
	configdb.DB.Clients = append(configdb.DB.Clients[:id], configdb.DB.Clients[id+1:]...)
	m.Mu.Unlock()
	err := m.Save()
	if err != nil {
		log.Printf("%s", err)
	}
}
func addClientConfig(config configdb.Client) {
	log.Printf("Adding client config for: %s\n", config.MAC)
	m := configdb.Mu{}
	m.Mu.Lock()
	if config.Options == nil {
		config.Options = make(map[string]interface{})
	}
	configdb.DB.Clients = append(configdb.DB.Clients, config)
	m.Mu.Unlock()
	err := m.Save()
	if err != nil {
		log.Printf("%s", err)
	}
}

func updateClientConfig(id int, config configdb.Client) {
	log.Printf("Updating client config for: %s\n", config.MAC)
	m := configdb.Mu{}
	m.Mu.Lock()
	if config.Options == nil {
		config.Options = make(map[string]interface{})
	}
	configdb.DB.Clients[id].IP = config.IP
	configdb.DB.Clients[id].Hostname = config.Hostname
	configdb.DB.Clients[id].Options = config.Options
	m.Mu.Unlock()
	err := m.Save()
	if err != nil {
		log.Printf("%s", err)
	}
}

func SearchForClientByMac(mac string) (configdb.Client, int) {
	for id, client := range configdb.DB.Clients {
		if client.MAC == mac {
			return client, id
		}
	}
	return configdb.Client{}, -1
}

func searchForClientByHostname(hostname string) (configdb.Client, int) {
	for id, client := range configdb.DB.Clients {
		if client.Hostname == hostname {
			return client, id
		}
	}
	return configdb.Client{}, -1
}

func searchForClientByIp(ip string) (configdb.Client, int) {
	for id, client := range configdb.DB.Clients {
		if client.IP == ip {
			return client, id
		}
	}
	return configdb.Client{}, -1
}

func listAllClients() ([]byte, error) {
	cfg, err := json.MarshalIndent(configdb.DB.Clients, "", "  ")
	if err != nil {
		log.Printf("JSON marshall error: %s.", err)
		return nil, err
	}
	return cfg, nil
}

func configToJson(config configdb.Client) []byte {
	cfg, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("JSON marshall error: %s", err)
	}
	return cfg
}

func StartServer() {
	router := httprouter.New()
	router.GET("/", index)
	router.GET("/clients", clientConfigShowFull)
	router.GET("/client/:mac", clientConfigShow)
	router.PUT("/client/:mac", clientConfigAdd)
	router.DELETE("/client/:mac", clientConfigDelete)
	if configdb.Config.TLSEnabled {
		log.Fatal(http.ListenAndServeTLS(configdb.Config.HTTPListenAddress, configdb.Config.TLSCertificateFile, configdb.Config.TLSPrivateKeyFile, router))
	} else {
		log.Fatal(http.ListenAndServe(configdb.Config.HTTPListenAddress, router))
	}
}
