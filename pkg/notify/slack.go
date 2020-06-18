package notify

import (
	"bytes"
	"log"

	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
)

// SendGeneralSlack Send a slack message to the general channel
func SendGeneralSlack(msg string) {
	SendSlack(msg, viper.GetString("slack.general_channel"))
}

// SendAdminSlack Send a slack message to the admin channel
func SendAdminSlack(msg string) {
	SendSlack(msg, viper.GetString("slack.admin_channel"))
}

// SendSlack Send a slack message
func SendSlack(msg string, channel string) {
	values := map[string]string{
		"text":     msg,
		"channel":  channel,
		"username": "scurvy",
	}

	jsonValue, _ := json.Marshal(values)

	resp, err := http.Post(
		viper.GetString("webhook_address"),
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Println("Error posting to slack", err, resp)
	}
}
