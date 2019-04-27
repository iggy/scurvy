package main

import (
	// "bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"encoding/json"

	"github.com/iggy/scurvy/pkg/config"
	"github.com/iggy/scurvy/pkg/errors"
	"github.com/iggy/scurvy/pkg/msgs"
	"github.com/iggy/scurvy/pkg/notify"

	"github.com/nats-io/go-nats"
	// "github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
	var jmsg = msgs.NewDownload{}
	if jerr := json.Unmarshal(m.Data, &jmsg); jerr != nil {
		log.Panicf("fatal error reading json msg from nats: %s", jerr)
	}
	if m.Subject == "scurvy.notify.newdownload" {
		syncCmd := viper.GetString("syncd.newdownload.script")
		log.Printf("%s\n", jmsg.Name)
		log.Printf("Running sync command: %s\n", syncCmd)
		cmd := exec.Command(syncCmd)
		err := cmd.Run()
		if err != nil {
			log.Printf("Command error code: %v\n", err)
		}
	}
	if m.Subject == "scurvy.notify.reportfiles" {
		log.Printf("Reporting current files to master\n")
	}
	// log.Printf("%s - %v - %v", m.Subject, m.Reply, m.Sub)
}

func main() {
	config.ReadConfig()

	hostname, _ := os.Hostname()

	// Setup a ping timer to send periodic pings to the nats mq... another process listens on the
	// server and sends a slack if it hasn't heard from a host in a while
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		for range ticker.C {

			// log.Printf("Pinging nats with hostname: %s (%s)\n", hostname, t)

			msgs.SendNatsPing(hostname)
		}
	}()

	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.UserInfo(
			viper.GetString("mq.user"),
			viper.GetString("mq.password"),
		),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.Printf("Got disconnected!\n")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("Connection closed. Reason: %q\n", nc.LastError())
		}),
	)
	errors.CheckErr(err)
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	errors.CheckErr(err)

	subj, i := "scurvy.notify.newdownload", 0
	c.Subscribe(subj, func(msg *nats.Msg) {
		i++
		printMsg(msg, i)
	})
	c.Flush()

	lerr := nc.LastError()
	errors.CheckErr(lerr)

	msg := fmt.Sprintf("Initializing Scurvy Sync Daemon on host: %s.", hostname)
	notify.SendAdminSlack(msg)

	runtime.Goexit()
}
