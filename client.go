package main

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var BULK_WINDOW_MS int64 = 100

type Client struct {
	done   chan struct{}
	ch     chan []Message
	buffer []Message
}

func (c *Client) handleMessage(m Message) {
	c.buffer = append(c.buffer, m)
}
func (c *Client) close() {
	c.done <- struct{}{}
}

// Messages are delivered in bulks to avoid
// ddossing the client (browser) with too many messages produced
// in a very short timespan
func (c *Client) startBufferFlushLoop() {
	for {
		time.Sleep(time.Millisecond * time.Duration(BULK_WINDOW_MS))
		select {
		case <-c.done:
			logger.Debug("Client: received done signal, quitting")
			defer close(c.done)
			defer close(c.ch)
			return
		default:

			if len(c.buffer) == 0 {
				continue
			}

			logger.WithField("count", len(c.buffer)).Debug("Client: Flushing buffer")
			c.ch <- c.buffer
			c.buffer = []Message{}
		}

	}
}

func NewClient() *Client {
	c := &Client{
		done:   make(chan struct{}),
		ch:     make(chan []Message, 100),
		buffer: []Message{},
	}

	go c.startBufferFlushLoop()

	return c
}

type Clients struct {
	mu                 sync.Mutex
	mainChan           <-chan Message
	clients            map[int]*Client
	buffer             []Message
	currentlyConnected int
}

func NewClients(msgs <-chan Message) *Clients {
	return &Clients{
		mu:                 sync.Mutex{},
		mainChan:           msgs,
		clients:            map[int]*Client{},
		currentlyConnected: 0,
		buffer:             []Message{},
	}
}

func (c *Clients) Start() {
	for {
		msg := <-c.mainChan
		c.mu.Lock()
		if c.currentlyConnected == 0 {
			logger.Debug("Received a log message but no client is connected, buffering message")
			c.buffer = append(c.buffer, msg)
		}

		for _, ch := range c.clients {
			ch.handleMessage(msg)
		}
		c.mu.Unlock()
	}
}

func (c *Clients) Join(id int) *Client {

	if _, ok := c.clients[id]; ok {
		panic("Client already exists")
	}

	c.mu.Lock()
	defer func() {
		if len(c.buffer) == 0 {
			return
		}

		logger.WithFields(logrus.Fields{
			"msg_count": len(c.buffer),
		}).Info("Flushing log messages buffer to a recently connected client")
		for _, msg := range c.buffer {
			cl := c.clients[id]
			cl.handleMessage(msg)
		}

		c.buffer = []Message{}
	}()
	defer c.mu.Unlock()

	c.clients[id] = NewClient()
	c.currentlyConnected++
	return c.clients[id]
}

func (c *Clients) Close(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.clients[id]; !ok {
		return
	}

	cl := c.clients[id]
	cl.close()
	delete(c.clients, id)
	c.currentlyConnected--
}