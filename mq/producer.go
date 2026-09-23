package mq

import (
	"context"
	"log"
	"log/slog"
	"sync"

	"github.com/Fulim13/lottery/database"
	"github.com/bytedance/sonic"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// The exchange cancel-order messages are published to. It is declared as an
	// x-delayed-message exchange, a type the rabbitmq_delayed_message_exchange plugin adds:
	// it holds a message back for the number of milliseconds in the message's x-delay header
	// and only routes it on once that time has passed.
	EXCHANGE = "CANCEL_ORDER"
	// The queue bound to that exchange. Every consumer reading this queue shares the messages
	// on it
	QUEUE = "lottery"
	// Key the queue is bound with; the exchange routes on it like a direct exchange would.
	ROUTING_KEY = "cancel_order"
)

var (
	// AMQP address of the broker, built from conf/rabbitmq.yaml by InitMQ
	endPoint     string
	producerConn *amqp.Connection
	producer     *amqp.Channel
	ponce        sync.Once
	// A channel must not be published on by two goroutines at once, so publishes line up
	// behind this instead of each request getting a channel of its own.
	pmu sync.Mutex
)

func GetProducer() *amqp.Channel {
	ponce.Do(func() {
		if producer != nil {
			return
		}
		var err error
		// Connect to the broker
		producerConn, err = amqp.Dial(endPoint)
		if err != nil {
			log.Fatal(err)
		}
		// Create Producer
		producer, err = producerConn.Channel()
		if err != nil {
			log.Fatal(err)
		}
		// The exchange has to exist before the first publish
		if err = declareDelayedExchange(producer); err != nil {
			log.Fatal(err)
		}
	})
	return producer
}

// Send a cancel order message with a delay
func SendCancelOrder(order database.Order, delay int) error {
	content, err := sonic.Marshal(order)
	if err != nil {
		return err
	}
	producer := GetProducer()
	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // written to disk, so a broker restart doesn't drop it
		Body:         content,
	}
	// The plugin keeps the message in the exchange for this many milliseconds before routing it
	// to the queue, which is what SetDelayTimestamp did on RocketMQ.
	msg.Headers = amqp.Table{"x-delay": int32(delay * 1000)}

	pmu.Lock()
	defer pmu.Unlock()
	err = producer.PublishWithContext(context.Background(), EXCHANGE, ROUTING_KEY, false, false, msg)
	if err != nil {
		return err
	}
	return nil
}

func StopProducter() {
	if producer != nil {
		producer.Close()
		slog.Info("stop producer")
	}
	if producerConn != nil {
		producerConn.Close()
	}
}
