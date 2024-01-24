## Handles Protocol

Official implementation of the handles social protocol.

This file deals with running a handles protocol validator. For information about breeze protocol see the [protocol presentation](https://github.com/freehandle/handles/blob/main/handles.md)

## Building the source

Building blow requires a Go compiler (1.21 or later). You can install it using your favorite package manager. Once it is installed, run

```
make
```

from the repo root folder. It will build a binary handles on the same folder.

## Executables

Handles provides a validator module and a block database module. 

| Module             | Description                                                                  |
| ------------------ | ---------------------------------------------------------------------------- |
| **`blow-handles`** | sequencer and validator for the handles protocol                             |
| **`echo-handles`** | blocks database and token indexation for the handles protocol                |

## Minimum hardware requirements for running handles

For a standalone validator with no outgoing connections

- CPU with 4 cores

- 16Gb RAM
 
- 20 MBit/sec internet connectivity

For a validator connected to block listeners

- a static IP address

## Running Handles

### Validator

To run a handles protocol validator a json configuration file with the relevant specifications must be provided.

```
blow-handles <path-to-json-config-file>
```

The standard configuration for handles protocol is 

```
{
	"token": "hex 64 char representation of node token",
	"adminPort":  <port other than 6001 for node remote administration>
	"blocks": { firewall config (see bellow) for block listener connections }
	"keepNBlock": <number (>=900) of blocks to keep in memory>
	"trustedProviders": [
        {
            "address": "dns or ip address of a trusted provider for breeze blocks",
            "token": "token associated to that address"
        }, ...
    ]
	"providersSize": <number of different providers to connect to>,
	"notaryPath": "path for local state persistence (empty is memory only)"
	"genesis": <true> for a new chain from genesis, <false> for an existing chain,
	"trustedPeers":  [
           {
            "address": "dns or ip address of a trusted peer to sync state",
            "token": "token associated to that address"
        }, ...

    ]
}
```

Firewall rules follow the breeze convention 

```
{
    "open": [true|false]
    "tokenList": [<token 1>, <token 2>,...] 
 }
```

When "open" is set to __true__ the firewall will by default allow all connections except those blacklisted by the "tokenList". When __false__, the firewall will by default forbid all connections except those whitelisted by the "tokenList". 

If the node starts from genesis, node must be ensured to process the entire history of breeze blockchain. Typically only a sync mode configuration will be used. 

### Block Database

To run a handles protocol default block database a json configuration file with the relevant specifications must be provided.

```
echo-handles <path-to-json-config-file>
```

The standard configuration for the handles protocol is 

```
{
	"token": "hex 64 char representation of node token",
	"port": <port for incoming connections>,.
    "adminPort":  <another port for node remote administration>,
	"firewall": { firewall configuration for incomming connections }
	"trustedProviders": [
        {
            "address": "dns or ip address of a trusted provider for bree blocks",
            "token": "token associated to that address"
        }, ...
    ]
	"providersSize": <number of different providers to connect to>,
	"databasePAth": "path for block history and index storage"
	"indexed": <true> for indexation <false> for block history only,
    "networkID": "underlying breeze network id",
}
```

## Contribution

#### Synergy

[Synergy](https://github.com/freehandle/synergy) protocol was designed as a digital framework for collaboration and collective construction. It runs seamlessly on top of the Breeze protocol and on top of handles social protocol.  

Handles is, itself, an ongoing project inside the Synergy protocol. To collaborate with design decisions on handles, you are welcome to join [Personal Internet Collective](https://freehandle.org/synergy/collective/personal_internet). 

#### Github

Contributions that **do not** change protocol functionalities, such as bug fixes, testing coverage, code refactorings, etc, may be proposed directly as a pull request targeting the main branch of [handles official repository](). 

For contributions that in anyway include protocol change, please join [Synergy's Personal Internet Collective]() and join a previous discussion involving the community, so decisions regarding the changes can be made collectively. 

## License

Handles is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0.txt). 

