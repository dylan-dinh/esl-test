package user

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"os"
)

const (
	exchangeName          = "user.events"
	UserCreatedRoutingKey = "user.created"
	UserUpdatedRoutingKey = "user.updated"
	UserDeletedRoutingKey = "user.deleted"
	queueName             = "user"
)

type RabbitMQ struct {
	logger   *slog.Logger
	Ch       *amqp.Channel
	queue    amqp.Queue
	confirms chan amqp.Confirmation
}

func NewRabbitMQ(conn *amqp.Connection) (*RabbitMQ, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err = ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	if err = ch.QueueBind(queue.Name, "user.*", exchangeName, false, nil); err != nil {
		return nil, err

	}

	// allow knowing if message are received successfully or not
	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("channel could not be put into confirm mode: %w", err)
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 10))

	logger.Info("RabbitMQ setup complete", "exchange", exchangeName, "queue", queue.Name)

	return &RabbitMQ{
		logger:   logger,
		Ch:       ch,
		queue:    queue,
		confirms: confirms,
	}, nil
}

func (r *RabbitMQ) publishAndConfirm(ctx context.Context, routingKey string, body []byte) error {
	if err := r.Ch.Publish(exchangeName, routingKey, false, false,
		amqp.Publishing{ContentType: "application/json", Body: body},
	); err != nil {
		return err
	}

	select {
	case confirm := <-r.confirms:
		if !confirm.Ack {
			return fmt.Errorf("message NACK : %d", confirm.DeliveryTag)
		} else {
			r.logger.Info("message ACK", "confirm", confirm.DeliveryTag)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *RabbitMQ) UserCreatedEvent(ctx context.Context, u *User) error {
	body, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return r.publishAndConfirm(ctx, UserCreatedRoutingKey, body)
}

func (r *RabbitMQ) UserUpdatedEvent(ctx context.Context, u *User) error {
	body, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return r.publishAndConfirm(ctx, UserUpdatedRoutingKey, body)
}

func (r *RabbitMQ) UserDeletedEvent(ctx context.Context, id string) error {
	body, err := json.Marshal(id) // produces a quoted JSON string
	if err != nil {
		return err
	}
	return r.publishAndConfirm(ctx, UserDeletedRoutingKey, body)
}
