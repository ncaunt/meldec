package main

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"strconv"
	"time"
)

var client mqtt.Client

func init_mqtt() {
	mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)

	hostname, _ := os.Hostname()
	server := "tcp://127.0.0.1:1883"
	clientid := hostname + strconv.Itoa(time.Now().Second())

	connOpts := mqtt.NewClientOptions().AddBroker(server).SetClientID(clientid).SetCleanSession(true)
	client = mqtt.NewClient(connOpts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return
	}
}

func publish(topic, value string) {
	client.Publish(topic, byte(0), true, value)
	return
}
