package irc

import (
	"io"
	"bufio"
	"log"
)

type Client struct {
	Err error

	Server io.ReadWriter
	Handlers Handlers

	Fatal chan error
	Warn chan error
	Quit chan int
}

func NewClient (server io.ReadWriter, handlers Handlers) *Client {
	return &Client{
		Server: server,
		Handlers: handlers,
		Quit: make(chan int),
		Fatal : make(chan error, 1),
		Warn : make(chan error, 1),
	}
}

func (c *Client) Run () {

	s := bufio.NewScanner(c.Server)
	s.Split(ScanLines)

LOOP:
	for {
		select {

		default:
			if !s.Scan() {
				if s.Err() != nil {
					c.Err = s.Err()
				}
				break LOOP
			}

			var (
				msg Message
				m = make([]byte, len(s.Bytes()))
			)

			copy(m, s.Bytes())

			msg.Unmarshal(m)

			log.Printf(">> %#v\n", msg)

			hl := <-c.Handlers
			for _, h := range hl {
				if h.Accept(msg) {
					go h.Handle(c, &msg)
				}
			}
			c.Handlers <-hl

		case warn := <-c.Warn:
			log.Println(warn)

		case c.Err = <-c.Fatal:
			break LOOP
		}
	}
	c.Quit <-1
}

func (c *Client) Send (msg *Message) {

	line, err := msg.Marshal()

	defer func () { c.Err = err } ()

	if err != nil {
		return
	}

	log.Printf("<< %s\n", string(line))

	_, err = c.Server.Write(append(line, '\r', '\n'))
}

type Handler struct{
	Accept func (Message) bool
	Handle func (*Client, *Message)
}

type Handlers chan []*Handler

func NewHandlers () Handlers {
	h := make(Handlers, 1)
	h <- nil
	return h
}

func (handlers Handlers) Add (handler *Handler) (Handlers) {
	handlers <- append(<-handlers, handler)
	return handlers
}

func (handlers Handlers) Remove (handler *Handler) {
	a := <-handlers
	z := len(a) - 1
	for i := range a {
		if a[i] == handler {
			a[i] = a[z]
			a[z] = nil
			a = a[:z]
			z--
		}
	}
	handlers <-a
}

