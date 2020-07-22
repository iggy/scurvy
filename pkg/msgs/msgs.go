package msgs

import (
	"log"

	"github.com/iggy/scurvy/pkg/config"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

// NewDownloadSubject - The subject name for the newdownload messages
const NewDownloadSubject = "scurvy.notify.newdownload"

// ReportFilesSubject = The subject name for the messages that tell syncd to update the list of
// local files to the master
const ReportFilesSubject = "scurvy.notify.reportfiles"

// SendNatsMsg - send a commonly formatted Nats message to the message queue
func SendNatsMsg(Subject string, Msg NatsMsg) {
	config.ReadConfig()

	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.UserInfo(viper.GetString("mq.user"), viper.GetString("mq.password")))
	if err != nil {
		log.Println("failed to connect to nats", err)
		return
	}
	defer nc.Close()
	// c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	// if err != nil { log.Println("failed", err) }

	err = nc.Publish(Subject, Msg.serialize())
	if err != nil {
		log.Println("failed to publish message", err)
		return
	}
	nc.Flush()

	err = nc.LastError()
	if err != nil {
		log.Println("failed lasterror, not sure what this means", err)
		return
	}
	log.Printf("Published [%s] : '%s'\n", Subject, Msg)
}

// SendNatsPing - send a ping message to nats, there's something on the other end that listens and
// alerts if a host doesn't check in for a while
// Nothing in this function should panic... the pings aren't that important
func SendNatsPing(Who string) {
	config.ReadConfig()

	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.UserInfo(viper.GetString("mq.user"), viper.GetString("mq.password")))
	if err != nil {
		log.Println("SendNatsPing: failed to connect to send ping", err)
		return
	}
	defer nc.Close()
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Println("SendNatsPing: failed to setup encoded connection", err)
		return
	}

	err = c.Publish("ping", Who)
	if err != nil {
		log.Println("SendNatsPing: failed to publish ping message", err)
		return
	}
	c.Flush()

	err = nc.LastError()
	if err != nil {
		log.Println("SendNatsPing: failed last error, not sure what this means", err)
		return
	}
	err = nc.Drain()
	if err != nil {
		log.Println("SendNatsPing: failed to drain", err)
	}
	// log.Printf("Published [ping] : '%s'\n", Who)
}
