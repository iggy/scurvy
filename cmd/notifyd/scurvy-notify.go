package main

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"

	"github.com/iggy/scurvy/pkg/config"
	"github.com/iggy/scurvy/pkg/msgs"
	"github.com/iggy/scurvy/pkg/notify"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

func handleNatsMsg(m *nats.Msg) {
	log.Printf("Received on [%s]: '%s'\n", m.Subject, string(m.Data))
	if m.Subject == "scurvy.notify.newdownload" {
		var jmsg = msgs.NewDownload{}
		if jerr := json.Unmarshal(m.Data, &jmsg); jerr != nil {
			log.Panicf("fatal error reading NewDownload json msg from nats: %s", jerr)
		}
		notify.SendGeneralSlack(fmt.Sprintf("Good news everyone! %s was downloaded to %s",
			jmsg.Name, jmsg.Path))
	}
	if m.Subject == "scurvy.notify.faileddownload" {
		var jmsg = msgs.FailedDownload{}
		if jerr := json.Unmarshal(m.Data, &jmsg); jerr != nil {
			log.Panicf("fatal error reading FailedDownload json msg from nats: %s", jerr)
		}
		notify.SendGeneralSlack(fmt.Sprintf("Bad news everyone! %s failed to downloaded to %s",
			jmsg.Name, jmsg.Path))
	}
	if m.Subject == "scurvy.notify.diskfull" {
		var jmsg = msgs.DiskFull{}
		if jerr := json.Unmarshal(m.Data, &jmsg); jerr != nil {
			log.Panicf("fatal error reading DiskFull json msg from nats: %s", jerr)
		}
		notify.SendAdminSlack(fmt.Sprintf("Disk full: %s", jmsg.Message))
	}
}

func main() {
	log.Printf("Scurvy Notification Daemon\n\n\n")

	config.ReadConfig()

	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.UserInfo(viper.GetString("mq.user"), viper.GetString("mq.password")))
	if err != nil {
		log.Panicf("Failed to nats.Connect: %#v", err)
	}
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Panicf("Failed to switch to encoded connection: %#v", err)
	}

	subj := "scurvy.notify.*"
	subs, err := c.Subscribe(subj, handleNatsMsg)
	if err != nil {
		log.Println("Failed to subscribe to nats", err, subj, subs)
	}
	c.Flush()

	err = nc.LastError()
	if err != nil {
		log.Panicf("Failed nc.LastError: %#v", err)
	}
	notify.SendAdminSlack("Initializing Scurvy Notification Daemon.")

	runtime.Goexit()
}
