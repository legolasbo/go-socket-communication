package socketcommunication

import (
	"log"
	"net"
	"os"
	"time"
)

// Host provides a way to send messages to clients that connect through a socket.
type Host struct {
	connInfo       ConnectionInfo
	listener       net.Listener
	clients        []net.Conn
	lastActivity   time.Time
	activityTimout time.Duration
	AddClient      chan net.Conn
	RemoveClient   chan net.Conn
	SendMessage    chan string
}

// NewHost creates a new host structure.
func NewHost(info ConnectionInfo) *Host {
	return &Host{
		connInfo:     info,
		AddClient:    make(chan net.Conn),
		RemoveClient: make(chan net.Conn),
		SendMessage:  make(chan string),
	}
}

// Starts the host by initialising the socket and then listening for new connections.
// Should always be called as a new goroutine to prevent blocking further execution.
func (h *Host) Start() {
	h.SetActivityTimeout(10 * 365 * 24 * time.Hour)
	address := h.connInfo.Address
	if h.connInfo.Ctype == UNIX {
		_ = os.Remove(address)
	}
	network := h.connInfo.Ctype.toString()
	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("Unable to listen to %s connections on %s: %s", network, address, err)
	}
	h.listener = listener
	go h.listen()
	go h.manageClients()
	go h.messageSender()
	go h.panicOnTimeout()
}

func (h *Host) listen() {
	for {
		conn, err := h.listener.Accept()
		if err != nil {
			log.Fatalf("Error on accept: %s", err)
		}

		h.AddClient <- conn
	}
}

func (h *Host) manageClients() {
	for {
		select {
		case c := <-h.RemoveClient:
			for i, client := range h.clients {
				if c == client {
					_ = c.Close()
					h.clients[i] = h.clients[len(h.clients)-1]
					h.clients = h.clients[:len(h.clients)-1]
					break
				}
			}
			break
		case c := <-h.AddClient:
			h.clients = append(h.clients, c)
			break
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// Stops the host by closing the listener and all open connections.
func (h *Host) Stop() {
	_ = h.listener.Close()
	for _, c := range h.clients {
		h.RemoveClient <- c
	}
}

// MessageSender reads messages from the channel and distributes then to all clients.
func (h *Host) messageSender() {
	for {
		msg := <-h.SendMessage

		h.lastActivity = time.Now()

		if msg[len(msg)-1] != '\n' {
			msg += "\n"
		}

		for _, c := range h.clients {
			err := c.SetWriteDeadline(time.Now().Add(time.Second))
			if err != nil {
				h.RemoveClient <- c
				continue
			}
			_, err = c.Write([]byte(msg))
			if err != nil {
				h.RemoveClient <- c
			}
		}
	}
}

func (h *Host) panicOnTimeout() {
	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		<-ticker.C
		if time.Now().Sub(h.lastActivity) > h.activityTimout {
			panic("It's been too long since the last message was sent!")
		}
	}
}

// Sets the duration after which we panic.
func (h *Host) SetActivityTimeout(duration time.Duration) {
	h.activityTimout = duration
	h.lastActivity = time.Now()
}
