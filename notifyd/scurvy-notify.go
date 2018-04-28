package main

import (
	"fmt"
	"runtime"

	"encoding/json"

	"github.com/iggy/scurvy/config"
	"github.com/iggy/scurvy/msgs"
	"github.com/iggy/scurvy/notify"

	"github.com/nats-io/go-nats"
	// "github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printMsg(m *nats.Msg, i int) {
	fmt.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
	var jmsg = msgs.NewDownload{}
	if jerr := json.Unmarshal(m.Data, &jmsg); jerr != nil {
		panic(fmt.Errorf("fatal error reading json msg from nats: %s", jerr))
	}
	if m.Subject == "scurvy.notify.newdownload" {
		notify.SendGeneralSlack(fmt.Sprintf("Good news everyone! %s was downloaded to %s",
			jmsg.Name, jmsg.Path))
	}
	// fmt.Printf("%s - %v - %v", m.Subject, m.Reply, m.Sub)
}

func main() {
	fmt.Printf("Scurvy Notification Daemon\n\n\n")

	config.ReadConfig()

	// build the slack webhook address here and shove it back into viper for safe keeping
	viper.Set("webhook_address",
		fmt.Sprintf("https://hooks.slack.com/services/%s", viper.GetString("slack.webhook_key")))

	scheme := "nats"
	if viper.GetBool("mq.tls") {
		scheme = "tls"
	}

	connectString := fmt.Sprintf("%s://%s:%s",
		scheme,
		viper.GetString("mq.host"),
		viper.GetString("mq.port"))

	nc, err := nats.Connect(connectString,
		nats.UserInfo(viper.GetString("mq.user"), viper.GetString("mq.password")))
	checkErr(err)
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	checkErr(err)

	subj, i := "scurvy.notify.*", 0
	c.Subscribe(subj, func(msg *nats.Msg) {
		i++
		printMsg(msg, i)
	})
	c.Flush()

	lerr := nc.LastError()
	checkErr(lerr)
	// sendAdminSlack("Initializing Scurvy Notification Daemon.")

	runtime.Goexit()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
