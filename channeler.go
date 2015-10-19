package gogozmq

import (
	"fmt"
	"net"
	"strings"
)

var (
	signature = []byte{0xff, 0, 0, 0, 0, 0, 0, 0, 1, 0x7f}
)

type Channeler struct {
	conn         Conn
	listener     net.Listener
	sockType     byte
	endpoints    string
	subscribe    string
	SendChan     chan<- []byte
	RecvChan     <-chan []byte
	sendDoneChan chan bool
	recvDoneChan chan bool
}

func newChanneler(sockType byte, endpoints, subscribe string) (*Channeler, error) {
	sendChan := make(chan []byte)
	recvChan := make(chan []byte)
	sendDoneChan := make(chan bool)
	recvDoneChan := make(chan bool)

	c := &Channeler{
		sockType:     sockType,
		endpoints:    endpoints,
		subscribe:    subscribe,
		SendChan:     sendChan,
		RecvChan:     recvChan,
		sendDoneChan: sendDoneChan,
		recvDoneChan: recvDoneChan,
	}

	var err error
	switch sockType {
	case Pull:
		addrParts := strings.Split(endpoints, "://")
		if len(addrParts) != 2 {
			return nil, fmt.Errorf("malformed address")
		}

		c.listener, err = net.Listen(addrParts[0], addrParts[1])
		if err != nil {
			return c, err
		}
		go c.recvMessages(recvChan)
	case Push:
		c.conn, err = NewPushConn(endpoints)
		if err != nil {
			return c, err
		}
		go c.sendMessages(sendChan)
	}

	return c, err
}

func NewPushChanneler(endpoints string) (*Channeler, error) {
	c, err := newChanneler(Push, endpoints, "")
	return c, err
}

func NewPullChanneler(endpoints string) (*Channeler, error) {
	return newChanneler(Pull, endpoints, "")
}

func (c *Channeler) Destroy() {
	close(c.SendChan)
	<-c.sendDoneChan
	c.conn.Close()
}

func (c *Channeler) sendMessages(sendChan <-chan []byte) {
	more := true
	var msg []byte
	for more {
		msg, more = <-sendChan
		if more {
			_, err := c.conn.Write(msg)

			if err != nil {
				// reconnection is not handled at the moment
				break
			}
		}
	}

	c.sendDoneChan <- true
}

func (c *Channeler) recvMessages(recvChan chan<- []byte) {
	zmtpMessageIncoming := &message{}
	conn, err := c.listener.Accept()
	if err != nil {
		panic(err)
	}

	zmtpGreetOutgoing := &greeter{
		sockType: c.sockType,
	}
	err = zmtpGreetOutgoing.greet(conn)
	if err != nil {
		fmt.Println("I panicked now")
		panic(err)
	}

	frame, err := zmtpMessageIncoming.recv(conn)
	if err != nil {
		return
	}

	zmtpMessageIncoming.msg = append(zmtpMessageIncoming.msg, frame)
	for _, msg := range zmtpMessageIncoming.msg {
		recvChan <- msg
	}
	c.recvDoneChan <- true
}
