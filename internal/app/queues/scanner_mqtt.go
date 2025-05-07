package queues

import (
	"fmt"

	"github.com/agilistikmal/parkingo-core/internal/app/services"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ScannerMQTT struct {
	Client         mqtt.Client
	Token          mqtt.Token
	S3Service      *services.S3Service
	SubscribeTopic string
	PublishTopic   string
}

func NewScannerMQTT(s3Service *services.S3Service) *ScannerMQTT {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", viper.GetString("mqtt.host"), viper.GetInt("mqtt.port")))
	opts.SetClientID("parkingo-core")

	opts.OnConnect = func(c mqtt.Client) {
		logrus.Info("Connected to MQTT broker")
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		logrus.Errorf("Connection lost: %v\n", err)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		logrus.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	return &ScannerMQTT{
		Client:         client,
		Token:          token,
		S3Service:      s3Service,
		SubscribeTopic: viper.GetString("mqtt.topic.subscribe"),
		PublishTopic:   viper.GetString("mqtt.topic.publish"),
	}
}

func (s *ScannerMQTT) Subscribe() {
	if !s.Client.IsConnected() {
		logrus.Warn("MQTT client is not connected")
		return
	}

	token := s.Client.Subscribe(s.SubscribeTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		logrus.Infof("Received message: %s from topic: %s", msg.Payload(), msg.Topic())
	})
	token.Wait()
	if token.Error() != nil {
		logrus.Error("Failed to subscribe to topic %s: %v", s.SubscribeTopic, token.Error())
	}
}
