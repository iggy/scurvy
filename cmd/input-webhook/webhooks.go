package main

import (
	"fmt"
	"log"

	"net/http"

	"github.com/iggy/scurvy/pkg/config"

	"github.com/spf13/viper"
)

func main() {
	log.Println("scurvy webhook input daemon")

	// Set config defaults - can be overridden in config file
	viper.SetDefault("BindAddress", "127.0.0.1")
	viper.SetDefault("BindPort", "38475")

	config.ReadConfig()

	bindString := fmt.Sprintf("%s:%s",
		viper.GetString("BindAddress"),
		viper.GetString("BindPort"))

	router := NewRouter()

	log.Printf("Attempting to listen on: %s\n", bindString)

	log.Fatal(http.ListenAndServe(bindString, router))
}
