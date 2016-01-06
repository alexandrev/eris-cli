package initialize

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/version"
)

//need to either pull from toadserver & update with strings.Replace on init
//leave here for testing .. ? (or put right into test!)
func DefChainService() string {
	ver := version.VERSION
	return fmt.Sprintf(`
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml
[service]
image          = "quay.io/eris/erisdb:%s"
data_container = true
ports          = [ "1337", "46656", "46657" ]
entry_point    = "erisdb-wrapper"

[dependencies]
services = [ "keys" ]

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"
`, ver)
}

//moved to gh.com/eris-ltd/eris-chains
func DefChainConfig() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

# It is used to initialize a new eris:db node.
# For more info about the various options, 
# see https://github.com/eris-ltd/mint-client (mintgen).

moniker = "defaulttester.com"
seeds = ""
fast_sync = false
db_backend = "leveldb"
log_level = "debug"
node_laddr = "0.0.0.0:46656"
rpc_laddr = "0.0.0.0:46657"
vm_log = false
`
}

func DefChainGen() string {
	return `
{
  "chain_id": "my_tests",
  "accounts": [
    {
      "address": "0000000000000000000000000000000000000001",
      "amount": 690000000000
    },
    {
      "address": "0000000000000000000000000000000000000002",
      "amount": 565000000000
    },
    {
      "address": "0000000000000000000000000000000000000003",
      "amount": 525000000000
    },
    {
      "address": "0000000000000000000000000000000000000004",
      "amount": 110000000000
    },
    {
      "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
      "amount": 999999999999
    }
  ],
  "validators": [
    {
      "pub_key": [
        1,
        "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
      ],
      "amount": 5000000000,
      "unbond_to": [
        {
          "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
          "amount": 5000000000
        }
      ]
    }
  ]
}
`
}

// different from genesis above! -- used for testing
//[zr] leave for testing - for now ?
var DefaultPubKeys = []string{"CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"}

func DefChainCSV() string {
	return fmt.Sprintf("%s,", DefaultPubKeys[0])
}

//use tool to gen one of these...
func DefChainKeys() string {
	return `
{
  "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
  "pub_key": [
    1,
    "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
  ],
  "priv_key": [
    1,
    "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
  ],
  "last_height": 0,
  "last_round": 0,
  "last_step": 0
}
`
}

//moved to gh.com/eris-ltd/eris-chains
func DefChainServConfig() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

[bind]
address=""
port=1337

[TLS]
tls=false
cert_path=""
key_path=""

[CORS]
enable=false
allow_origins=[]
allow_credentials=false
allow_methods=[]
allow_headers=[]
expose_headers=[]
max_age=0

[HTTP]
json_rpc_endpoint="/rpc"

[web_socket]
websocket_endpoint="/socketrpc"
max_websocket_sessions=50
read_buffer_size=2048
write_buffer_size=2048

[logging]
console_log_level="info"
file_log_level="warn"
log_file=""
`
}
