package main

import (
	"bumbleserver.org/common/key"
	"bumbleserver.org/router"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	NetworkBinding        string
	BumbleName            string
	PrivateKeyFile        string // relative to the config file's location
	SessionTimeout        int    // disconnected session timeout
	AuthenticationTimeout int    // connected time allowed while not authenticated
	KeyFile               string // filename of key (only required if wanting TLS)
	CertFile              string // filename of cert (only required if wanting TLS)
	StatHatKey            string // if using StatHat, this is the EZ key
	MaxCPUs               int    // how many CPUs or cores can we use?
}

func main() {
	configFileName := "bumble.conf"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		fmt.Println("[ERROR] " + err.Error())
		fmt.Println("For your happiness an example config file is provided in the repository.")
		os.Exit(1)
	}
	configDecoder := json.NewDecoder(configFile)
	config := new(Config)
	err = configDecoder.Decode(config)
	if err != nil {
		fmt.Println("[CONFIG FILE FORMAT ERROR] " + err.Error())
		fmt.Println("Please ensure that your config file is in valid JSON format.")
		os.Exit(1)
	}

	runtime.GOMAXPROCS(config.MaxCPUs)

	switch {
	case config.NetworkBinding == "":
		fmt.Println("[CONFIG ERROR] NetworkBinding is missing.")
		os.Exit(1)
	case config.BumbleName == "":
		fmt.Println("[CONFIG ERROR] BumbleName is missing.")
		os.Exit(1)
	case config.PrivateKeyFile == "":
		fmt.Println("[CONFIG ERROR] PrivateKeyFile is missing.")
		os.Exit(1)
	}

	privateKeyFile := filepath.Clean(config.PrivateKeyFile)
	if !filepath.IsAbs(privateKeyFile) {
		privateKeyFile = filepath.Clean(filepath.Join(filepath.Dir(configFileName), privateKeyFile))
	}

	privateKey, err := key.PrivateKeyFromPEMFile(privateKeyFile)
	if err != nil {
		fmt.Printf("[PRIVATEKEY ERROR] %s\n", err)
		os.Exit(1)
	}

	routerConfig := router.Config{
		NetBind:               config.NetworkBinding,
		Name:                  config.BumbleName,
		PrivateKey:            privateKey,
		SessionTimeout:        config.SessionTimeout,
		AuthenticationTimeout: config.AuthenticationTimeout,
		KeyFile:               config.KeyFile,
		CertFile:              config.CertFile,
		StatHatKey:            config.StatHatKey,
	}
	router.RouterStart(routerConfig)
}
