package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// ReadConfig - read/setup the viper config stuffs
func ReadConfig() {
	// Read in config file(s)
	viper.SetConfigName("scurvy")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Panicf("fatal error config file: %s", err)
	}

	// assemble the slack webhook address here and shove it back into viper for safe keeping
	viper.Set("webhook_address",
		fmt.Sprintf("https://hooks.slack.com/services/%s", viper.GetString("slack.webhook_key")))
}

// GetNatsConnString helper to assemble the NATS connect string and return it
func GetNatsConnString() string {
	ReadConfig()

	scheme := "nats"
	if viper.GetBool("mq.tls") {
		scheme = "tls"
	}

	connectString := fmt.Sprintf("%s://%s:%s",
		scheme,
		viper.GetString("mq.host"),
		viper.GetString("mq.port"))

	return connectString
}
