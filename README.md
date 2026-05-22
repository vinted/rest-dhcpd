# REST DHCPD
DHCP server controlled via REST endpoints.

## Configuration

By default configuration files are located in `/etc/rest-dhcpd` directory.
This can be overridden with `-configPath` command-line argument.
Global server configuration is stored in `rest-dhcpd-config.json` file:

```
{
    "LeaseDuration": 30,
    "AuthToken": "secretToken",
    "HTTPListenAddress": "127.0.0.1:6767",
    "TLSEnabled": true,
    "TLSPrivateKeyFile": "example.key",
    "TLSCertificateFile": "example.crt",
    "Options": {
      "6": ["1.1.1.1", "8.8.8.8"]
    },
    "Interfaces": [
      {
        "Name": "virbr1",
        "IP": "192.168.100.1",
        "Options": {
          "1": "255.255.255.0",
          "3": "192.168.100.1"
        }
      },
      {
        "Name": "virbr2",
        "IP": "10.0.0.1",
        "LeaseDuration": 3600,
        "Options": {
          "1": "255.255.0.0",
          "3": "10.0.0.1"
        }
      }
    ]
}
```

- `LeaseDuration` - default DHCP lease duration in seconds. Used by any interface that does not set its own.
- `AuthToken` - authentication token for REST interface.
- `HTTPListenAddress` - address to listen for HTTP requests. To listen on all network addresses use port without any IP address `:6767`.
- `TLSEnabled` - start REST interface with HTTPS support.
- `TLSPrivateKeyFile` - private key file for HTTPS support.
- `TLSCertificateFile` - certificate file for HTTPS support.
- `Options` - top-level DHCP options applied to every interface as defaults. Per-interface `Options` override these per option key, and per-client `Options` override both.
- `Interfaces` - list of interfaces to listen on. Must contain at least one entry. Each entry has:
  - `Name` - network interface name (e.g. `virbr1`).
  - `IP` - IP address of the DHCP server on this interface (used as the server identifier in responses).
  - `LeaseDuration` - optional lease duration override for this interface. If omitted, the top-level value is used.
  - `Options` - DHCP options specific to this subnet. Merged on top of the top-level `Options`.

Requests are served by whichever interface received them; the response uses that interface's `IP` and its merged option set.

#### Migration from single-interface configs
If upgrading from a version with top-level `IP` and `ListenInterface`, move those values into a one-element `Interfaces` array along with the subnet-specific entries from `Options`. Any options shared across interfaces (e.g. DNS) can stay at the top level as defaults.

For backwards compatibility, the legacy top-level `IP` and `ListenInterface` fields are still accepted: if `Interfaces` is omitted and these fields are set, a single-entry `Interfaces` array is synthesized from them at startup and a deprecation warning is logged. Configs that set both forms are rejected.

### Basics
REST DHCP server does not provide dynamic DHCP leases. It only provides leases to configured clients.
Clients are added, deleted and modified via REST interface. Client configuration is stored in `-configPath` folder, `rest-dhcpd-clients.json` file.

## API

REST API supports following methods:
`GET, PUT, DELETE`

### API endpoints

`/` - supports `GET` method. Displays index page.  
`/metrics` - supports `GET` method. Displays prometheus metrics.  
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
