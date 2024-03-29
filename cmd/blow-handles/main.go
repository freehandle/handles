package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/freehandle/breeze/consensus/chain"
	"github.com/freehandle/breeze/crypto"
	"github.com/freehandle/breeze/middleware/config"
	"github.com/freehandle/breeze/middleware/social"
	"github.com/freehandle/breeze/socket"
	"github.com/freehandle/breeze/util"
	"github.com/freehandle/handles/attorney"
)

const ProtocolPort = 6001

type HandleConfig struct {
	// Token for the node
	Token string // `json:"token"`
	// Port for the admin interface
	CredentialsPath string // `json:"credentialsPath"`
	AdminPort       int    // `json:"adminPort"`
	// Firewall for incoming connections
	Blocks config.FirewallConfig // `json:"blocks"`
	// Number of blocks to keep in memory
	KeepNBlocks int // `json:"keepNBlocks"`
	// Trusted block providers for the node
	TrustedProviders []config.Peer // `json:"trustedProviders"`
	// Number of providers to connect to receive new blocks
	ProvidersSize int // `json:"providersSize"`
	// Path to the notary file (empty for memory)
	NotaryPath string // `json:"notaryPath"`
	// True if the node will initiate a new social chain from genesis
	Genesis bool // `json:"genesis"`
	// Trusted peers for the node to sync state
	TrustedPeers []config.Peer // `json:"trustedPeers"`
}

func (c HandleConfig) Check() error {
	token := crypto.TokenFromString(c.Token)
	if token == crypto.ZeroToken {
		return errors.New("invalid nde token")
	}
	if c.CredentialsPath != "" {
		_, err := config.ParseCredentials(c.CredentialsPath, token)
		if err != nil {
			return fmt.Errorf("could not parse credentials: %v", err)
		}
	}
	if c.AdminPort == ProtocolPort {
		return fmt.Errorf("invalid admin port: %d is reserved for protocol", ProtocolPort)
	}
	if c.KeepNBlocks < 900 {
		return fmt.Errorf("invalid keep n block: %v is less than 900", c.KeepNBlocks)
	}
	if len(c.TrustedProviders) == 0 {
		return fmt.Errorf("no trusted providers")
	}
	if c.ProvidersSize > len(c.TrustedProviders) {
		return fmt.Errorf("providers size %v is greater than trusted providers %v", c.ProvidersSize, len(c.TrustedProviders))
	}
	if (!c.Genesis) && len(c.TrustedPeers) == 0 {
		return fmt.Errorf("no trusted peers for non-genesis node")
	}
	return nil
}

func HandleConfigToConfig(hdl *HandleConfig, pk crypto.PrivateKey) Config {

	cfg := Config{
		Node: social.Configuration{
			Hostname:           "",
			Credentials:        pk,
			AdminPort:          hdl.AdminPort,
			Firewall:           config.FirewallToValidConnections(hdl.Blocks),
			KeepNBlocks:        hdl.KeepNBlocks,
			ParentProtocolCode: 0,
			NodeProtocolCode:   1,
			RootBlockInterval:  time.Second,
			RootChecksumWindow: 900,
			CalculateCheckSum:  true,
			BlocksSourcePort:   5405,
			BlocksTargetPort:   6001,
			TrustedProviders:   config.PeersToTokenAddrWithPort(hdl.TrustedProviders, 5405),
			ProvidersSize:      hdl.ProvidersSize,
			MaxCheckpointLag:   10,
		},
		Genesis:      hdl.Genesis,
		TurstedPeers: config.PeersToTokenAddr(hdl.TrustedPeers),
		NotaryPath:   hdl.NotaryPath,
	}
	return cfg
}

type Config struct {
	Node         social.Configuration
	Genesis      bool
	TurstedPeers []socket.TokenAddr
	NotaryPath   string
}

func launchGenesis(ctx context.Context, cfg Config) chan error {
	genesis := attorney.NewGenesisState(cfg.NotaryPath)
	bytes := []byte{}
	util.PutUint32(cfg.Node.NodeProtocolCode, &bytes)
	util.PutUint32(cfg.Node.ParentProtocolCode, &bytes)
	hash := genesis.Checksum()
	genesisHash := crypto.Hasher(append(bytes, hash[:]...))
	checksum := &social.Checksum[*attorney.Mutations, *attorney.MutatingState]{
		Epoch:         0,
		State:         genesis,
		LastBlockHash: genesisHash,
		Hash:          genesisHash,
	}
	clock := chain.ClockSyncronization{
		Epoch:     0,
		TimeStamp: time.Now(),
	}
	return social.LaunchNodeFromState[*attorney.Mutations, *attorney.MutatingState](ctx, cfg.Node, checksum, clock)
}

func main() {
	specs, err := config.LoadConfig[HandleConfig](os.Args[1])
	if err != nil || specs == nil {
		fmt.Printf("misconfiguarion: %v\n", err)
		os.Exit(1)
	}
	token := crypto.TokenFromString(specs.Token)
	ctx, cancel := context.WithCancel(context.Background())
	var secret crypto.PrivateKey
	if specs.CredentialsPath != "" {
		secret, err = config.ParseCredentials(specs.CredentialsPath, token)
		if err != nil {
			fmt.Printf("could not retrieve credentials from file: %v\n", err)
			cancel()
			os.Exit(1)
		}
	} else {
		keys := config.WaitForRemoteKeysSync(ctx, []crypto.Token{token}, "localhost", specs.AdminPort)
		secret = keys[token]
		if !secret.PublicKey().Equal(token) {
			fmt.Println("could not synchrnize keys")
			os.Exit(1)
		}
	}
	cfg := HandleConfigToConfig(specs, secret)

	var finalize chan error
	if cfg.Genesis {
		finalize = launchGenesis(ctx, cfg)
	} else {
		finalize = social.LaunchSyncNode(ctx, cfg.Node, cfg.TurstedPeers, attorney.NewStateFromBytes(cfg.NotaryPath))
	}

	err = <-finalize
	if err != nil {
		fmt.Printf("service crashed: %v\n", err)
		os.Exit(1)
	}
}
