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
	"github.com/iggy/scurvy/pkg/config"
	"github.com/iggy/scurvy/pkg/errors"
	"github.com/iggy/scurvy/pkg/msgs"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

var ircServername = "irc.oftc.net"
var ircPort = 6697 // ssl port
var ircChannelname = "#testscurvybot"

var reallyquit = false

func main() {
	log.Println("Initializing scurvy ircbot")
	flag.Parse()

	config.ReadConfig()

	// config irc connection
	cfg := irc.NewConfig("scurvybot")
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: ircServername}
	cfg.Server = fmt.Sprintf("%s:%d", ircServername, ircPort)
	cfg.NewNick = func(n string) string { return n + "^" }
	// different Recover function that exits (vs just logging)
	// cfg.Recover = func(conn *irc.Conn, line *irc.Line) {
	// 	log.Printf("%v\n", conn)
	// 	log.Printf("%v\n", line)
	// 	log.Println("Error in irc handler. Hopefully there's a useful error above.")
	// }
	c := irc.Client(cfg)
	c.EnableStateTracking()

	// setup handlers
	// join channel on connect
	// log.Println("Setting up connect/join handler")
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) { conn.Join(ircChannelname) })

	// log.Println("Setting up privmsg handler")
	c.HandleFunc(irc.PRIVMSG, handleprivmsg)

	// signal on disconnect
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// read commands on stdin
	// log.Println("Setting up stdin reader")
	in := make(chan string, 4)
	stats, statErr := os.Stdin.Stat()
	if statErr != nil {
		fmt.Println("file.Stat()", statErr)
	}

	if stats.Size() > 0 {
		go func() {
			// log.Println("Setting up stdin for loop")

			con := bufio.NewReader(os.Stdin)
			for {
				s, _, err := con.ReadLine()
				if err != nil {
					log.Println("Error on ReadString, not reading from stdin anymore.")
					close(in)
					break
				}
				log.Printf("stdin: %s", s)
				if len(s) > 2 {
					in <- string(s)
				}
			}
		}()
	}
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
	log.Println("Setting up NATS connection")
	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.UserInfo(viper.GetString("mq.user"), viper.GetString("mq.password")))
	if err != nil {
		log.Panicf("Failed to nats.Connect: %#v", err)
	}
	natschan, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Panicf("Failed to switch to encoded connection: %#v", err)
	}

	subj, i := "scurvy.notify.*", 0
	log.Println("Setting up subscription")
	subs, err := natschan.Subscribe(subj, func(msg *nats.Msg) {
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
	if err != nil {
		log.Println("Failed to subscribe to the nats chan", err, subj, subs)
	}
	natschan.Flush()

	lerr := nc.LastError()
	if err != nil {
		log.Panicf("Failed nc.LastError: %#v", err)
	}

	log.Println("Setting up reallyquit loop")
	for !reallyquit {
		// connect
		if err := c.Connect(); err != nil {
			log.Printf("Connection error: %s\n", err.Error())
		}

		// wait for disconnect
		<-quit

	}

}

// parseArgs - parse the irc args into a command and args to that command
func parseArgs(a string) (string, interface{}) {
	if len(a) < 2 {
		log.Printf("Got invalid line: %v\n", a)
		return a, nil
	}

	split := strings.SplitN(a, " ", 2)

	if len(split) < 2 {
		return a, ""
	}

	return split[0], split[1]
}

func handleprivmsg(conn *irc.Conn, line *irc.Line) {
	log.Printf("privmsg: %s\n", line)
	log.Printf("args: %s\n", line.Args)
	// log.Println(conn)

	retChan := line.Args[0]

	command, cargs := parseArgs(line.Args[1])
	switch command {
	case "^search":
		log.Printf("Got search command with args: %v\n", cargs)
		conn.Privmsg(retChan, "Search not implemented yet")
	case "^help":
		conn.Privmsg(retChan, "^help:    this message")
		conn.Privmsg(retChan, "^search:  search for media")
	case "^quit":
		log.Println("Got shutdown command. Bye!")
		conn.Privmsg(retChan, "Fine. I know when I'm not wanted")
		reallyquit = true
		conn.Quit("Got ^quit IRC command")
	default:
		log.Printf("line: %s\n", line)
		log.Printf("args: %s\n", line.Args)
	}
}
