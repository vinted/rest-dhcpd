# REST DHCPD
DHCP server controlled via REST endpoints.

## Configuration

By default configuration files are located in `/etc/rest-dchp` directory.
This can be overridden with `-configPath` command-line argument.
Global server configuration is stored in `rest-dhcpd-config.json` file:

```
{
    "IP": "192.168.100.1",
    "LeaseDuration": 30,
    "AuthToken": "secretToken",
    "ListenInterface": "virbr1",
    "HTTPListenAddress": "127.0.0.1:6767",
    "TLSEnabled": true,
    "TLSPrivateKeyFile": "example.key",
    "TLSCertificateFile": "example.crt",
    "Options": {
      "1": "255.255.255.0",
      "3": "192.168.100.1",
      "6": "10.32.0.3"
    }
}
```
`IP` - IP address of DHCP server.  
`LeaseDuration` - DHCP lease duration in seconds.  
`AuthToken` - authentication token for REST interface.  
`ListenInterface` - network interface to listen fro DHCP requests.  
`HTTPListenAddress` - address to listen fro HTTP requests. To listen on all network addresses use port without any IP address `:6767`.  
`TLSEndabled` - start REST interface with HTTPS support.  
`TLSPrivateKeyFile` - private key file for HTTPS support.  
`TLSCertificateFile` - certificate file for HTTPS support.  
`Options` - list of global DHCP options. Can be overridden by client config.  

### Basics
REST DHCP server does not provide dynamic DHCP leases. It only provides leases to configured clients.
Clients are added, deleted and modified via REST interface. Client configuration is stored in `-configPath` folder, `rest-dhcpd-clients.json` file.

## API

REST API supports following methods:
`GET, PUT, DELETE`

### API endpoints

`/` - supports `GET` method. Displays index page.  
`/clients` - supports `GET` method. Lists all configured clients.  
`/client/AA:BB:CC:DD:EE:FF` - supports:  
		- `GET` - displays configuration of a client defined by `AA:BB:CC:DD:EE:FF` `MAC` address.  
		- `DELETE` - deletes configuration of a client defined by `AA:BB:CC:DD:EE:FF` `MAC` address.  
		- `PUT` -  creates new or update existing configuration of a client defined by `AA:BB:CC:DD:EE:FF` `MAC` address.  

### Examples

#### List all available clients
```
curl http://127.0.0.1:6767/clients -H "REST-DHCPD-Auth-Token: secretToken"
```
##### Display configuration of a specific client
```
curl http://127.0.0.1:6767/client/aa:bb:cc:dd:ee:ff -H "REST-DHCPD-Auth-Token: secretToken"
```
#####  Add or update client
```
curl  -X PUT http://127.0.0.1:6767/client/aa:bb:cc:dd:ee:ff -H "REST-DHCPD-Auth-Token: secretToken" -d '{"Hostname":"test4","IP":"192.168.13.17","Options":{"13":"option13"}}'
```
##### Delete client
```
curl -X DELETE http://127.0.0.1:6767/client/aa:bb:cc:dd:ee:ff -H "REST-DHCPD-Auth-Token: secretToken"
```

#### HTTP return codes
`200` - OK. Request was completed successfully.
`400` - BadRequest. Reports that configuration cannot be added or updated. HTTP body will display more detailed information about error.
`401` - Unauthorized.  `REST-DHCPD-Auth-Token` provided by client does not mach token in configuration.
