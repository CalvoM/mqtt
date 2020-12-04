package main

import (
	"fmt"
	"log"

	"github.com/CalvoM/mqtt"
	"github.com/spf13/viper"
)

func onMessage(data mqtt.MessageData) {
	fmt.Println(data.Topic)
	fmt.Println(data.Payload)
}

func main() {
	viper.SetConfigFile("../.env")
	viper.ReadInConfig()
	host := viper.Get("MQTT_HOST").(string)
	user := viper.Get("MQTT_USER").(string)
	pswd := viper.Get("MQTT_PWD").(string)
	client := mqtt.Client{
		Host: host,
		Port: "1883",
	}
	options := mqtt.ConnectOptions{
		CleanSession: true,
		KeepAlive:    5,
		Password:     pswd,
		Username:     user,
		WillFlag:     true,
		WillTopic:    "v/t/will",
		WillMessage:  "Goodbye",
		WillQOS:      mqtt.QOS2,
	}
	client.Init(&options)
	err := client.Connect()
	if err != nil {
		log.Fatal(err.Error())
	}
	client.OnMessage = onMessage
	fmt.Println("Connection Successful")
	err = client.Subscribe("v/t/test", mqtt.QOS1)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Subscription successful")
	for {

	}
}
