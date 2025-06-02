package http

import (
	"fmt"
	"sync"
	"time"

	"github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/ring"
	"github.com/logdyhq/logdy-core/utils"

	. "github.com/logdyhq/logdy-core/models"
)

var Ch chan models.Message
var Clients *ClientsStruct

var BULK_WINDOW_MS int64 = 100
var FLUSH_BUFFER_SIZE = 1000

type CursorStatus string

const CURSOR_STOPPED CursorStatus = "stopped"
const CURSOR_FOLLOWING CursorStatus = "following"

type Client struct {
	bufferOpMu sync.Mutex
	id         string
	done       chan struct{}
	ch         chan []Message
	buffer     []Message

	cursorStatus   CursorStatus
	cursorPosition string // last delivered message id
}

func (c *Client) handleMessage(m Message, force bool) {
	if !force && c.cursorStatus == CURSOR_STOPPED {
		utils.Logger.Debug("Client: Status stopped discarding message")
		return
	}
	c.buffer = append(c.buffer, m)
}

func (c *Client) flushBuffer() {
	if len(c.buffer) == 0 {
		return
	}

	for i := 0; i < len(c.buffer); i += FLUSH_BUFFER_SIZE {
		end := i + FLUSH_BUFFER_SIZE
		if end > len(c.buffer) {
			end = len(c.buffer)
		}

		batch := c.buffer[i:end]
		c.ch <- batch
	}
}

func (c *Client) clearBuffer() {
	c.buffer = []Message{}
}

func (c *Client) close() {
	c.done <- struct{}{}
}

func (c *Client) waitForBufferDrain() {
	for len(c.buffer) > 0 {
		time.Sleep(5 * time.Millisecond)
	}
}

// Messages are delivered in bulks to avoid
// ddossing the client (browser) with too many messages produced
// in a very short timespan
func (c *Client) startBufferFlushLoop() {
	for {
		time.Sleep(time.Millisecond * time.Duration(BULK_WINDOW_MS))
		select {
		case <-c.done:
			utils.Logger.Debug("Client: received done signal, quitting")
			defer close(c.done)
			defer close(c.ch)
			return
		default:

			if len(c.buffer) == 0 {
				continue
			}

			utils.Logger.WithField("count", len(c.buffer)).Debug("Client: Flushing buffer")
			c.cursorPosition = c.buffer[len(c.buffer)-1].Id
			c.bufferOpMu.Lock()

			c.flushBuffer()
			c.clearBuffer()

			c.bufferOpMu.Unlock()
		}

	}
}

func NewClient() *Client {
	c := &Client{
		bufferOpMu:     sync.Mutex{},
		done:           make(chan struct{}),
		ch:             make(chan []Message, BULK_WINDOW_MS*25),
		cursorStatus:   CURSOR_STOPPED,
		cursorPosition: "",
		id:             utils.RandStringRunes(6),
	}

	go c.startBufferFlushLoop()

	return c
}

type ClientsStruct struct {
	started            bool
	mu                 sync.Mutex
	mainChan           <-chan Message
	clients            map[string]*Client
	ring               *ring.RingQueue[Message]
	currentlyConnected int
	stats              Stats
}

func NewClients(msgs <-chan Message, maxCount int64) *ClientsStruct {
	if maxCount == 0 {
		maxCount = 100_000
	}

	cls := &ClientsStruct{
		mu:                 sync.Mutex{},
		mainChan:           msgs,
		clients:            map[string]*Client{},
		currentlyConnected: 0,
		ring:               ring.NewRingQueue[Message](maxCount),
		stats: Stats{
			MaxCount: maxCount,
			Count:    0,
		},
	}

	go cls.Start()

	return cls
}

func (c *ClientsStruct) GetClient(clientId string) (*Client, bool) {
	cl, ok := c.clients[clientId]
	return cl, ok
}

func (c *ClientsStruct) Load(clientId string, startCount int, count int, includeStart bool) {
	c.PauseFollowing(clientId)
	cl := c.clients[clientId]
	cl.waitForBufferDrain()

	cl.bufferOpMu.Lock()
	defer cl.bufferOpMu.Unlock()

	seen := false
	sent := 0
	c.ring.Scan(func(msg Message, i int) bool {
		if i+1 == startCount {
			seen = true
			if !includeStart {
				return false
			}
		}

		if !seen {
			return false
		}

		sent++
		cl.handleMessage(msg, true)
		cl.cursorPosition = msg.Id

		if count > 0 && sent >= count {
			return true
		}
		return false
	})

	cl.flushBuffer()

}

func (c *ClientsStruct) PeekLog(idxs []int) []Message {
	msgs := []Message{}

	for _, idx := range idxs {
		if c.ring.Size()-1 < idx {
			continue
		}
		msg, err := c.ring.PeekIdx(idx)
		if err != nil {
			panic(err)
		}
		msgs = append(msgs, msg)
	}

	return msgs
}

func (c *ClientsStruct) Stats() Stats {
	return c.stats
}
func (c *ClientsStruct) ClientStats(clientId string) ClientStats {
	stats := ClientStats{}
	cl, exists := c.GetClient(clientId)
	if !exists {
		return stats
	}

	stats.LastDeliveredId = cl.cursorPosition

	c.ring.Scan(func(m Message, idx int) bool {
		if m.Id == cl.cursorPosition {
			stats.LastDeliveredIdIdx = idx
			return true
		}

		return false
	})

	stats.CountToTail = c.Stats().Count - stats.LastDeliveredIdIdx

	return stats
}

func (c *ClientsStruct) ResumeFollowing(clientId string, sinceCursor bool) {
	//pump back the items until last element seen

	c.clients[clientId].bufferOpMu.Lock()
	if sinceCursor {
		seen := false
		c.ring.Scan(func(msg Message, _ int) bool {
			if msg.Id == c.clients[clientId].cursorPosition {
				seen = true
				return false
			}

			if !seen {
				return false
			}

			c.clients[clientId].handleMessage(msg, true)
			return false
		})

	}
	c.clients[clientId].flushBuffer()
	c.clients[clientId].cursorStatus = CURSOR_FOLLOWING
	c.clients[clientId].bufferOpMu.Unlock()
}

func (c *ClientsStruct) PauseFollowing(clientId string) {
	c.clients[clientId].cursorStatus = CURSOR_STOPPED
	c.clients[clientId].waitForBufferDrain()
}

// starts a delivery channel to all clients
func (c *ClientsStruct) Start() {
	if c.started {
		utils.Logger.Debug("Clients delivery loop already started")
		return
	}

	c.started = true
	for {
		msg := <-c.mainChan
		if c.stats.FirstMessageAt.IsZero() {
			c.stats.FirstMessageAt = time.Now()
		}

		c.ring.PushSafe(msg)
		if c.stats.Count < int(c.stats.MaxCount) {
			c.stats.Count++
		}

		c.stats.LastMessageAt = time.Now()

		for _, ch := range c.clients {
			ch.bufferOpMu.Lock()
			ch.handleMessage(msg, false)
			ch.bufferOpMu.Unlock()
		}
	}
}

func (c *ClientsStruct) Join(tailLen int, shouldFollow bool) *Client {
	cl := NewClient()
	c.clients[cl.id] = cl
	c.currentlyConnected++

	// deliver last N messages from a buffer upon connection
	idx := 0
	if c.ring.Size() > tailLen {
		idx = c.ring.Size() - tailLen
	}
	sl, err := c.ring.PeekSlice(idx)

	if err != nil {
		panic(err)
	}
	for _, msg := range sl {
		cl.handleMessage(msg, true)
	}

	if shouldFollow {
		c.clients[cl.id].cursorStatus = CURSOR_FOLLOWING
	}

	return c.clients[cl.id]
}

func (c *ClientsStruct) Close(id string) {
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

func InitChannel() {
	if Ch != nil {
		return
	}

	Ch = make(chan models.Message, 1000)
}

func InitializeClients(config Config) *ClientsStruct {
	if Clients != nil {
		return Clients
	}

	bts := int64(0)

	if config.AppendToFileRotateMaxSize != "" {
		var err error
		bts, err = utils.ParseRotateSize(config.AppendToFileRotateMaxSize)

		if err != nil {
			panic(fmt.Errorf("file rotate size parse error: %w", err))
		}
	}

	mainChan := utils.ProcessIncomingMessagesWithRotation(Ch, config.AppendToFile, config.AppendToFileRaw, bts, 1000)
	Clients = NewClients(mainChan, config.MaxMessageCount)

	return Clients
}
