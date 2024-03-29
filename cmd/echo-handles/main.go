package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/freehandle/breeze/crypto"
	"github.com/freehandle/breeze/middleware/blockdb"
	"github.com/freehandle/breeze/middleware/blocks"
	"github.com/freehandle/breeze/middleware/config"
	"github.com/freehandle/handles/attorney"
)

type Config struct {
	// Token for the node
	Token           string // json:"token"
	CredentialsPath string // `json:"credentialsPath"`
	// Port for the admin interface
	AdminPort int // json:"admin_port"
	// Port for the new conenctiosn
	Port int // json:"port"
	// Firewall for incoming connections
	Firewall config.FirewallConfig // json:"firewall"
	// Trusted block providers for the node
	TrustedProviders []config.Peer // json:"trustedProviders"
	// Number of providers to connect to receive new blocks
	ProvidersSize int // json:"providersSize"
	// Path to the notary file (empty for memory)
	DatabasePath string // json:"databasePath"
	// True if the node will initiate a new social chain from genesis
	Indexed bool // json:"indexed"
	// NetworkID
	NetworkID string // json:"networkID"
}

func (c Config) Check() error {
	if token := crypto.TokenFromString(c.Token); token == crypto.ZeroToken {
		return fmt.Errorf("invalid token")
	}
	if c.AdminPort == 0 || c.Port == 0 {
		return fmt.Errorf("invalid ports")
	}
	if len(c.TrustedProviders) == 0 {
		return fmt.Errorf("no trusted providers")
	}
	if c.ProvidersSize > len(c.TrustedProviders) {
		return fmt.Errorf("providers size %v is greater than trusted providers %v", c.ProvidersSize, len(c.TrustedProviders))
	}
	if c.DatabasePath == "" {
		return fmt.Errorf("no database path")
	}
	return nil
}

func ConfigToBlocksConfig(cfg Config, pk crypto.PrivateKey) blocks.Config {
	config := blocks.Config{
		Credentials: pk,
		DB: blockdb.DBConfig{
			Path:           cfg.DatabasePath,
			Indexed:        cfg.Indexed,
			ItemsPerBucket: 60,
			BitsForBucket:  10,
			IndexSize:      8,
		},
		NetworkID: cfg.NetworkID,
		Port:      cfg.Port,
		Firewall:  config.FirewallToValidConnections(cfg.Firewall),
		Sources:   config.PeersToTokenAddr(cfg.TrustedProviders),
		PoolSize:  cfg.ProvidersSize,
		Protocol: &blocks.ProtocolRule{
			Code: 0x01,
		},
	}
	if cfg.Indexed {
		config.Protocol.IndexFn = attorney.GetHashes
	} else {
		config.Protocol.IndexFn = func([]byte) []crypto.Hash {
			return nil
		}
	}
	return config
}

func main() {
	specs, err := config.LoadConfig[Config](os.Args[1])
	if err != nil || specs == nil {
		fmt.Printf("misconfiguarion: %v\n", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	token := crypto.TokenFromString(specs.Token)

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

	cfg := ConfigToBlocksConfig(*specs, secret)
	server := blocks.NewServer(ctx, nil, cfg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
		cancel()
	case err := <-server:
		if err == nil {
			fmt.Println("server exited")
			return
		}
		fmt.Println("server exited with error: ", err)
	}
}
