package msgs

import (
	"log"

	"github.com/iggy/scurvy/pkg/config"
	"github.com/iggy/scurvy/pkg/errors"

	"github.com/nats-io/nats.go"
	// "github.com/spf13/cobra"
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
	errors.CheckErr(err)
	defer nc.Close()
	// c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	// errors.CheckErr(err)

	nc.Publish(Subject, Msg.serialize())
	nc.Flush()

	lerr := nc.LastError()
	errors.CheckErr(lerr)
	log.Printf("Published [%s] : '%s'\n", Subject, Msg)
}

// SendNatsPing - send a ping message to nats, there's something on the other end that listens and
// alerts if a host doesn't check in for a while
func SendNatsPing(Who string) {
	config.ReadConfig()

	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.UserInfo(viper.GetString("mq.user"), viper.GetString("mq.password")))
	errors.CheckErr(err)
	defer nc.Close()
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	errors.CheckErr(err)

	c.Publish("ping", Who)
	c.Flush()

	lerr := nc.LastError()
	errors.CheckErr(lerr)
	// log.Printf("Published [ping] : '%s'\n", Who)
}
