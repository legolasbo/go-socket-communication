package socketcommunication

import (
	"log"
	"net"
	"strings"
	"time"
)

type MessageHandler interface {
	handle(msg string)
}

type Client struct {
	socketFile    string
	connection    net.Conn
	AddHandler    chan MessageHandler
	RemoveHandler chan MessageHandler
	messages      chan string
	handlers      []MessageHandler
}

func NewClient(socketFile string) *Client {
	return &Client{
		socketFile:    socketFile,
		AddHandler:    make(chan MessageHandler),
		RemoveHandler: make(chan MessageHandler),
		messages:      make(chan string),
	}
}

func (c *Client) connect() {
	for {
		if c.connection != nil {
			c.connection.Close()
		}
		conn, err := net.Dial("unix", c.socketFile)
		if err == nil {
			c.connection = conn
			break
		}
		log.Printf("Failed to dial: %s", err)
		time.Sleep(5 * time.Second)
	}
}

func (c *Client) Start() {
	log.Println("Starting client...")
	c.connect()
	log.Printf("Connected through %s", c.socketFile)
	go c.manageHandlers()
	go c.distributeMessages()
	go c.receive()
}

func (c *Client) receive() {
	rec := make(chan []byte)
	go c.handleRec(rec)

	log.Println("Start receiving")
	for {
		buf := make([]byte, 512)
		count, err := c.connection.Read(buf)
		if err != nil {
			log.Println("Disconnect detected... reconnecting...")
			c.connect()
		}
		rec <- buf[:count]
	}
}

func (c *Client) handleRec(rec chan []byte) {
	var msg string
	for {
		r := <-rec
		msg += string(r)
		msgs := strings.Split(msg, "\n")
		msg = msgs[len(msgs)-1]
		for _, m := range msgs[:len(msgs)-1] {
			c.messages <- m
		}
	}
}

func (c *Client) Stop() {
	c.connection.Close()
}

func (c *Client) manageHandlers() {
	for {
		select {
		case h := <-c.RemoveHandler:
			for i, handler := range c.handlers {
				if h == handler {
					c.handlers[i] = c.handlers[len(c.handlers)-1]
					c.handlers = c.handlers[:len(c.handlers)-1]
					break
				}
			}
			break
		case h := <-c.AddHandler:
			c.handlers = append(c.handlers, h)
			break
		default:
		}
	}
}

func (c *Client) distributeMessages() {
	for {
		msg := <-c.messages
		for _, handler := range c.handlers {
			handler.handle(msg)
		}
	}
}
