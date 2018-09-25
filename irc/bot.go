package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	irc "github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/logging/glog"
	"github.com/iggy/scurvy/common"
	"github.com/iggy/scurvy/config"
	"github.com/iggy/scurvy/msgs"
	"github.com/nats-io/go-nats"
	"github.com/spf13/viper"
)

var ircServername = "irc.oftc.net"
var ircPort = 6697
var ircChannelname = "#testscurvybot"

func main() {
	log.Println("Initializing scurvy ircbot")
	flag.Parse()
	glog.Init()

	config.ReadConfig()

	// config irc connection
	cfg := irc.NewConfig("scurvybot")
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: ircServername}
	cfg.Server = fmt.Sprintf("%s:%d", ircServername, ircPort)
	cfg.NewNick = func(n string) string { return n + "^" }
	c := irc.Client(cfg)
	c.EnableStateTracking()

	// setup handlers
	// join channel on connect
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) { conn.Join(ircChannelname) })

	c.HandleFunc(irc.PRIVMSG, handleprivmsg)

	// signal on disconnect
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// read commands on stdin
	in := make(chan string, 4)
	reallyquit := false
	go func() {
		con := bufio.NewReader(os.Stdin)
		for {
			s, err := con.ReadString('\n')
			if err != nil {
				log.Println("Error on ReadString, not reading from stdin anymore.")
				close(in)
				break
			}
			log.Printf("stdin: %s", s)
			if len(s) > 2 {
				in <- s[0 : len(s)-1]
			}
		}
	}()

	// another goroutine for parsing stdin
	go func() {
		for cmd := range in {
			log.Printf("Parse: %s\n", cmd)
			if cmd[0] == ':' {
				switch idx := strings.Index(cmd, " "); {
				case cmd[1] == 'd':
					log.Print(c.String())
				case cmd[1] == 'n':
					parts := strings.Split(cmd, " ")
					username := strings.TrimSpace(parts[1])
					// channelname := strings.TrimSpace(parts[2])
					_, userIsOn := c.StateTracker().IsOn(ircChannelname, username)
					log.Printf("Checking if %s is in %s: %t\n", username, ircChannelname, userIsOn)
				case idx == -1:
					log.Printf("Unknown command: %s\n", cmd)
					continue
				case cmd[1] == 'q':
					reallyquit = true
					c.Quit(cmd[idx+1:])
				case cmd[1] == 's':
					reallyquit = true
					c.Close()
				case cmd[1] == 'j':
					c.Join(cmd[idx+1:])
				case cmd[1] == 'p':
					c.Part(cmd[idx+1:])
				}
			} else {
				c.Raw(cmd)
			}

		}
	}()

	// setup NATS queue watcher to send IRC message on new download
	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.UserInfo(viper.GetString("mq.user"), viper.GetString("mq.password")))
	common.CheckErr(err)
	natschan, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	common.CheckErr(err)

	subj, i := "scurvy.notify.*", 0
	natschan.Subscribe(subj, func(msg *nats.Msg) {
		i++
		log.Printf("[#%d] Received on [%s]: '%s'\n", i, msg.Subject, string(msg.Data))
		var jmsg = msgs.NewDownload{}
		if jerr := json.Unmarshal(msg.Data, &jmsg); jerr != nil {
			log.Panicf("fatal error reading json msg from nats: %s", jerr)
		}
		if msg.Subject == "scurvy.notify.newdownload" {
			c.Privmsg(ircChannelname, fmt.Sprintf("Good news everyone! %s was downloaded to %s",
				jmsg.Name, jmsg.Path))
		}

	})
	natschan.Flush()

	lerr := nc.LastError()
	common.CheckErr(lerr)

	for !reallyquit {
		// connect
		if err := c.Connect(); err != nil {
			log.Printf("Connection error: %s\n", err.Error())
		}

		// wait for disconnect
		<-quit

	}

}

func handleprivmsg(conn *irc.Conn, line *irc.Line) {
	// This is similar to the inline handler functions, but it's broken out here because of length
	log.Printf("privmsg: %s\n", line)
	log.Printf("args: %s\n", line.Args)
	log.Println(conn)
	switch line.Args[1] {
	case "^search":
		conn.Privmsg(ircChannelname, "Search not implemented yet")
	case "^help":
		conn.Privmsg(ircChannelname, "^help: this message")
		conn.Privmsg(ircChannelname, "^search: search for media")
	default:
		log.Printf("line: %s\n", line)
		log.Printf("args: %s\n", line.Args)
	}
}
