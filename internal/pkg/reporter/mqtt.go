package reporter

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ncaunt/meldec/internal/pkg/decoder"
)

var server = flag.String("mqttserver", "tcp://127.0.0.1:1883", "MQTT server")
var debug = flag.Bool("mqttdebug", false, "Turn on MQTT debugging messages")

type MQTTReporter struct {
	client mqtt.Client
}

func NewMQTTReporter() (r Reporter, err error) {
	if *debug {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
	}
	mqtt.ERROR = log.New(os.Stderr, "", 0)

	hostname, _ := os.Hostname()
	server := *server
	clientid := hostname + strconv.Itoa(time.Now().Second())

	connOpts := mqtt.NewClientOptions().AddBroker(server).SetClientID(clientid).SetCleanSession(true)
	client := mqtt.NewClient(connOpts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return nil, token.Error()
	}

	r = &MQTTReporter{client: client}
	return
}

func (r *MQTTReporter) Publish(s decoder.Stat) {
	r.client.Publish(s.Name, byte(0), true, s.Value)
}
