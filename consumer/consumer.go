package consumer

import (
	"fmt"
	"log"
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

	ch, err := c.conn.Channel()
	if err != nil {
		c.listenErrChan <- err
		return
	}

	for _, target := range c.options.Targets {
		exchange, routingKey := split(target)
		if err := exchangeDeclare(ch, exchange); err != nil {
			c.listenErrChan <- err
			return
		}

		q, err := queueDeclare(ch, queueName(exchange, routingKey))
		if err != nil {
			c.listenErrChan <- err
			return
		}
		c.registeredQueue = append(c.registeredQueue, q.Name)

		err = queueBind(ch, q.Name, exchange, routingKey)
		if err != nil {
			c.listenErrChan <- err
			return
		}

		msgs, err := consume(ch, q.Name)
		if err != nil {
			c.listenErrChan <- err
			return
		}

		go func(msgs <-chan amqp.Delivery, target string) {
			for d := range msgs {
				fmt.Println(target, string(d.Body))
				c.listenMsgChan <- &Message{
					Target: target,
					Body:   d.Body,
				}
			}
		}(msgs, target)
	}
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
