package socketcommunication

import (
	"log"
	"net"
	"os"
	"time"
)

// Host provides a way to send messages to clients that connect through a socket.
type Host struct {
	socketFile   string
	listener     net.Listener
	clients      []net.Conn
	AddClient    chan net.Conn
	RemoveClient chan net.Conn
	SendMessage  chan string
}

// NewHost creates a new host structure.
func NewHost(socketFile string) *Host {
	return &Host{
		socketFile:   socketFile,
		AddClient:    make(chan net.Conn),
		RemoveClient: make(chan net.Conn),
		SendMessage:  make(chan string),
	}
}

// Starts the host by initialising the socket and then listening for new connections.
// Should always be called as a new goroutine to prevent blocking further execution.
func (h *Host) Start() {
	_ = os.Remove(h.socketFile)
	listener, err := net.Listen("unix", h.socketFile)
	if err != nil {
		log.Fatalf("Unable to listen on socket file %s: %s", h.socketFile, err)
	}
	h.listener = listener
	go h.listen()
	go h.manageClients()
	go h.messageSender()
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
		}
	}
}

// Stops the host by closing the listener and all open connections.
func (h *Host) Stop() {
	h.listener.Close()
	for _, c := range h.clients {
		h.RemoveClient <- c
	}
}

// MessageSender reads messages from the channel and distributes then to all clients.
func (h *Host) messageSender() {
	for {
		msg := <-h.SendMessage

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
