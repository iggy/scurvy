package config

import (
	"fmt"
	// "github.com/spf13/cobra"
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
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}
