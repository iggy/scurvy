package notify

import (
	"bytes"

	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
)

// SendGeneralSlack Send a slack message to the general channel
func SendGeneralSlack(msg string) {
	SendSlack(msg, "#general")
}

// SendAdminSlack Send a slack message to the admin channel
func SendAdminSlack(msg string) {
	SendSlack(msg, "#admins")
}

// SendSlack Send a slack message
func SendSlack(msg string, channel string) {
	values := map[string]string{
		"text":     msg,
		"channel":  channel,
		"username": "scurvy",
	}

	jsonValue, _ := json.Marshal(values)

	http.Post(viper.GetString("webhook_address"), "application/json", bytes.NewBuffer(jsonValue))
}
