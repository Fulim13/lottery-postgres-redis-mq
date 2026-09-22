package mq

import (
	"log"
	"log/slog"
	"sync"

	"github.com/Fulim13/lottery/database"
	"github.com/bytedance/sonic"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	consumerConn   *amqp.Connection
	simpleConsumer *amqp.Channel
	conce          sync.Once
)

// Declare the delayed exchange. Declaring is idempotent, so both the producer and the consumer
// do it on startup and neither has to come up first. This is what the mqadmin commands used to
// take care of on the RocketMQ side.
func declareDelayedExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(EXCHANGE, "x-delayed-message", true, false, false, false, amqp.Table{
		"x-delayed-type": "direct", // how the exchange routes a message once its delay is up
	})
}

func GetConsumer() *amqp.Channel {
	conce.Do(func() {
		if simpleConsumer != nil {
			return
		}
		var err error
		// Connect to the broker
		consumerConn, err = amqp.Dial(END_POINT)
		if err != nil {
			log.Fatal(err)
		}
		// Create consumer
		simpleConsumer, err = consumerConn.Channel()
		if err != nil {
			log.Fatal(err)
		}
		if err = declareDelayedExchange(simpleConsumer); err != nil {
			log.Fatal(err)
		}
		// The queue to read from, durable so it outlives a broker restart
		if _, err = simpleConsumer.QueueDeclare(QUEUE, true, false, false, false, nil); err != nil {
			log.Fatal(err)
		}
		// Subscribe the queue to the exchange
		if err = simpleConsumer.QueueBind(QUEUE, ROUTING_KEY, EXCHANGE, false, nil); err != nil {
			log.Fatal(err)
		}
		// Hold at most this many unacked messages at a time. Without it the broker pushes the
		// whole queue at one consumer, and the others get nothing to do.
		if err = simpleConsumer.Qos(10, 0, false); err != nil {
			log.Fatal(err)
		}
	})
	return simpleConsumer
}

func ReceiveCancelOrder() {
	consumer := GetConsumer()
	// autoAck is off: a message stays on the queue until the stock is back, so nothing is lost
	// if this process dies halfway through
	megs, err := consumer.Consume(QUEUE, "", false, false, false, false, nil)
	if err != nil {
		slog.Error("consume message failed", "error", err)
		return
	}
	// Ranging over the deliveries blocks until StopConsumer is called or the connection drops
	for mg := range megs {
		var order database.Order
		err := sonic.Unmarshal(mg.Body, &order)
		if err == nil {
			gid := database.GetTempOrder(order.UserId)
			// the temporary order is still there, which means the user never paid
			if gid == order.GiftId {
				database.DeleteTempOrder(order.UserId, order.GiftId) // delete the temporary order
				database.IncreaseInventory(order.GiftId)             // put one unit back in stock
				slog.Info("timed out, temporary order deleted", "uid", order.UserId, "gid", order.GiftId)
			}
		}
		mg.Ack(false)
	}
}

func StopConsumer() {
	if simpleConsumer != nil {
		simpleConsumer.Close()
		slog.Info("stop consumer")
	}
	if consumerConn != nil {
		consumerConn.Close()
	}
}
