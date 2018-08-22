package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/streadway/amqp"
)

type Message struct {
	Target string
	Body   []byte
}

type Consumer struct {
	options         *Options
	conn            *amqp.Connection
	registeredQueue []string
	listenMsgChan   chan *Message
	listenErrChan   chan error
	log             *log.Logger
}

type Options struct {
	Logger  *log.Logger
	DSN     string
	Targets []string
}

func New(o *Options) *Consumer {
	if o.Logger == nil {
		o.Logger = log.New(os.Stderr, "consumer:", log.Ldate|log.Ltime|log.Lshortfile)
	}
	return &Consumer{
		options:       o,
		listenMsgChan: make(chan *Message),
		listenErrChan: make(chan error),
		log:           o.Logger,
	}
}

func (c *Consumer) Run() {
	var err error
	c.conn, err = amqp.Dial(c.options.DSN)
	if err != nil {
		c.listenErrChan <- err
		return
	}

	for _, target := range c.options.Targets {
		if err := c.Consume(target); err != nil {
			c.listenErrChan <- err
		}
	}
}

func (c *Consumer) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(c.registeredQueue)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})
}

func (c *Consumer) Consume(target string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	exchange, routingKey := split(target)
	if err := exchangeDeclare(ch, exchange); err != nil {
		return err
	}

	q, err := queueDeclare(ch, queueName(exchange, routingKey))
	if err != nil {
		return err
	}
	c.registeredQueue = append(c.registeredQueue, q.Name)

	err = queueBind(ch, q.Name, exchange, routingKey)
	if err != nil {
		return err
	}

	msgs, err := consume(ch, q.Name)
	if err != nil {
		return err
	}

	go func(msgs <-chan amqp.Delivery, ch *amqp.Channel, target, qName string) {
		var counter int64
		for d := range msgs {
			c.listenMsgChan <- &Message{
				Target: target,
				Body:   d.Body,
			}
			counter++
			if counter >= 100 {
				break
			}
		}
		ch.QueueDelete(qName, false, false, false)
	}(msgs, ch, target, q.Name)

	return nil
}

func (c *Consumer) Close() {
	if c.conn == nil {
		return
	}

	ch, err := c.conn.Channel()
	if err != nil {
		c.listenErrChan <- err
		return
	}

	for _, queue := range c.registeredQueue {
		_, err = ch.QueueDelete(queue, false, false, false)
		if err != nil {
			c.log.Printf("%s: %s", queue, err.Error())
		}
	}

	err = c.conn.Close()
	if err != nil {
		c.listenErrChan <- err
	}
}

func (c *Consumer) ListenMessage() <-chan *Message {
	return c.listenMsgChan
}
func (c *Consumer) ListenError() <-chan error {
	return c.listenErrChan
}

func queueBind(ch *amqp.Channel, queue, exchange, key string) error {
	return ch.QueueBind(
		queue,    // name of the queue
		key,      // bindingKey
		exchange, // sourceExchange
		false,    // noWait
		nil,      // arguments
	)
}

func exchangeDeclare(ch *amqp.Channel, name string) error {
	return ch.ExchangeDeclare(
		name,    // name of the exchange
		"topic", // type
		true,    // durable
		false,   // delete when complete
		false,   // internal
		false,   // noWait
		nil,     // arguments
	)
}

func queueDeclare(ch *amqp.Channel, name string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // name, leave empty to generate a unique name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
}
func split(s string) (string, string) {
	ss := strings.Split(s, ":")
	return ss[0], ss[1]
}

func queueName(exchange, key string) string {
	return fmt.Sprintf("frog.%s.%s", exchange, strings.Replace(key, "*", "star", -1))
}

func consume(ch *amqp.Channel, queue string) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		queue,
		"",
		true,
		false,
		false,
		false,
		nil)
}
