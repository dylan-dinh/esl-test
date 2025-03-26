package notifier

import (
	"fmt"
	"github.com/dylan-dinh/esl-test/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// NewRabbitMQConn create the RabbitMQ connection
func NewRabbitMQConn(conf config.Config) (*amqp.Connection, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:%s/", conf.RabbitHost, conf.RabbitPort))
	if err != nil {
		return nil, err
	}
	return conn, nil
}
