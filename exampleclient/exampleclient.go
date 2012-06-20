package main

import (
	"bumbleserver.org/client"
	"bumbleserver.org/common/envelope"
	"bumbleserver.org/common/key"
	"bumbleserver.org/common/message"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	BumbleName      string
	PrivateKeyFile  string // relative to the config file's location
	IsAuthenticated bool
}

var config *Config = new(Config)

func main() {
	configFileName := "client.conf"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		fmt.Println("[ERROR] " + err.Error())
		fmt.Println("For your happiness an example config file is provided in the 'conf' directory in the repository.")
		os.Exit(1)
	}
	configDecoder := json.NewDecoder(configFile)
	err = configDecoder.Decode(config)
	if err != nil {
		fmt.Println("[CONFIG FILE FORMAT ERROR] " + err.Error())
		fmt.Println("Please ensure that your config file is in valid JSON format.")
		os.Exit(1)
	}

	switch {
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

	clientConfig := &client.Config{
		Name:             config.BumbleName,
		PrivateKey:       privateKey,
		OnConnect:        onConnect,
		OnDisconnect:     onDisconnect,
		OnAuthentication: onAuthentication,
		OnMessage:        onMessage,
	}
	c := client.NewClient(clientConfig)
	go func() {
		for {
			err = c.Connect() // NOTE: client.Connect never returns unless it has an error
			if err != nil {
				if err.Error() == "got disconnected" {
					fmt.Println("XXX")
				} else {
					fmt.Println(err)
					os.Exit(1)
				}
			}
			<-time.NewTimer(time.Duration(5e9)).C // 5 second delay before re-looping
		}
	}()

	tick := time.NewTicker(time.Duration(1e9))
	for {
		<-tick.C
		//fmt.Println("TICK")
		if config.IsAuthenticated {
			msg := message.NewGeneric(message.CODE_GENERICMESSAGE)
			msg.SetTo(c.Myself())
			msg.SetInfo("This is a message to myself.")
			c.OriginateMessage(msg)
		}
	}
}

func onConnect() {
	fmt.Println("We connected.")
}

func onDisconnect() {
	fmt.Println("We disconnected or got disconnected.")
}

func onAuthentication(success bool) {
	fmt.Printf("Did we authenticate?  %t\n", success)
	config.IsAuthenticated = success
}

func onMessage(e *envelope.Envelope, m *message.Header) {
	fmt.Printf("We got a message: %s\n", m)
}
