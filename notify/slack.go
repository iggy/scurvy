package notify

import (
	"bytes"

	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
)

func SendGeneralSlack(msg string) {
	SendSlack(msg, "#general")
}

func SendAdminSlack(msg string) {
	SendSlack(msg, "#admins")
}

func SendSlack(msg string, channel string) {
	values := map[string]string{
		"text":     msg,
		"channel":  channel,
		"username": "scurvy",
	}

	jsonValue, _ := json.Marshal(values)

	http.Post(viper.GetString("webhook_address"), "application/json", bytes.NewBuffer(jsonValue))
}
