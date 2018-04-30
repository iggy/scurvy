package msgs

import (
	"log"

	"github.com/iggy/scurvy/config"

	"github.com/nats-io/go-nats"
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
	checkErr(err)
	defer nc.Close()
	// c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	// checkErr(err)

	nc.Publish(Subject, Msg.serialize())
	nc.Flush()

	lerr := nc.LastError()
	checkErr(lerr)
	log.Printf("Published [%s] : '%s'\n", Subject, Msg)
}

// SendNatsPing - send a ping message to nats, there's something on the other end that listens and
// alerts if a host doesn't check in for a while
func SendNatsPing(Who string) {
	config.ReadConfig()

	nc, err := nats.Connect(config.GetNatsConnString(),
		nats.UserInfo(viper.GetString("mq.user"), viper.GetString("mq.password")))
	checkErr(err)
	defer nc.Close()
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	checkErr(err)

	c.Publish("ping", Who)
	c.Flush()

	lerr := nc.LastError()
	checkErr(lerr)
	// log.Printf("Published [ping] : '%s'\n", Who)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
