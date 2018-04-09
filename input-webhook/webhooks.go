package main

import (
	"fmt"
	"log"

	"net/http"

	"scurvy/config"

	"github.com/spf13/viper"
)

func main() {
	fmt.Printf("scurvy webhook input daemon\n\n\n")

	// Set config defaults - can be overridden in config file
	viper.SetDefault("BindAddress", "127.0.0.1")
	viper.SetDefault("BindPort", "38475")

	config.ReadConfig()

	bindString := fmt.Sprintf("%s:%s",
		viper.GetString("BindAddress"),
		viper.GetString("BindPort"))

	router := NewRouter()

	log.Fatal(http.ListenAndServe(bindString, router))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
