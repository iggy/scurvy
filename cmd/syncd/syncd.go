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
	"github.com/iggy/scurvy/pkg/msgs"
	"github.com/iggy/scurvy/pkg/notify"

	"github.com/nats-io/nats.go"
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
		stdoe, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Command error code: %v\n", err)
		}
		log.Println(stdoe)
		log.Println("Finished running sync command")
	}
	if m.Subject == "scurvy.notify.reportfiles" {
		// TODO
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
		log.Printf("Pinging nats with hostname: %s\n", hostname)
		for range ticker.C {
			msgs.SendNatsPing(hostname)
		}
	}()

	// nats.RetryOnFailedConnect(true), // Not released yet, maybe not needed
	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.MaxReconnects(300),
		nats.ReconnectWait(time.Second*5),
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
	if err != nil {
		log.Panicln("failed to connect to nats", err)
	}
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Println("failed to set encoded connection", err)
	}

	subj, i := "scurvy.notify.newdownload", 0
	subs, err := c.Subscribe(subj, func(msg *nats.Msg) {
		i++
		printMsg(msg, i)
	})
	if err != nil {
		log.Println("Failed to subscribe to nats", err, subj, subs)
	}
	c.Flush()

	lerr := nc.LastError()
	if lerr != nil {
		log.Println("failed lasterror, not sure what could cause this", lerr)
	}

	msg := fmt.Sprintf("Initializing Scurvy Sync Daemon on host: %s.", hostname)
	notify.SendAdminSlack(msg)

	runtime.Goexit()
}
