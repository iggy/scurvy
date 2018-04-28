package main

import (
	// "bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

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
		syncCmd := viper.GetString("syncd.newdownload.script")
		fmt.Printf("%s\n", jmsg.Name)
		fmt.Printf("Running sync command: %s\n", syncCmd)
		cmd := exec.Command(syncCmd)
		err := cmd.Run()
		fmt.Printf("Command error code: %x\n", err)
	}
	if m.Subject == "scurvy.notify.reportfiles" {
		fmt.Printf("Reporting current files to master\n")
	}
	// fmt.Printf("%s - %v - %v", m.Subject, m.Reply, m.Sub)
}

func main() {
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

	hostname, _ := os.Hostname()

	// Setup a ping timer to send periodic pings to the nats mq... another process listens on the
	// server and sends a slack if it hasn't heard from a host in a while
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		for range ticker.C {

			// fmt.Printf("Pinging nats with hostname: %s (%s)\n", hostname, t)

			msgs.SendNatsPing(hostname)
		}
	}()

	nc, err := nats.Connect(connectString,
		nats.UserInfo(
			viper.GetString("mq.user"),
			viper.GetString("mq.password"),
		),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("Got disconnected!\n")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Printf("Connection closed. Reason: %q\n", nc.LastError())
		}),
	)
	checkErr(err)
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	checkErr(err)

	subj, i := "scurvy.notify.newdownload", 0
	c.Subscribe(subj, func(msg *nats.Msg) {
		i++
		printMsg(msg, i)
	})
	c.Flush()

	lerr := nc.LastError()
	checkErr(lerr)

	msg := fmt.Sprintf("Initializing Scurvy Sync Daemon on host: %s.", hostname)
	notify.SendAdminSlack(msg)

	runtime.Goexit()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
