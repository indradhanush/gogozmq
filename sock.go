package gogozmq

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Conn interface {
	net.Conn
	Connect(endpoint string) error
	CurrentIdx() int
	Serve() error
	handleConnection(conn net.Conn) error
}

type Sock struct {
	conns      []net.Conn
	listener   net.Listener
	currentIdx int
	sockType   byte
}

func NewPushConn(endpoints ...string) (Sock, error) {
	var err error
	s := Sock{
		currentIdx: 0,
		sockType:   Push,
	}

	for i := 0; i < len(endpoints); i++ {
		address := endpoints[i]
		addrParts := strings.Split(address, "://")
		if len(addrParts) != 2 {
			return s, fmt.Errorf("malformed address")
		}

		conn, err := net.Dial(addrParts[0], addrParts[1])
		if err != nil {
			return s, err
		}

		zmtpGreetOutgoing := &greeter{
			sockType: s.sockType,
		}

		err = zmtpGreetOutgoing.greet(conn)
		if err != nil {
			return s, err
		}

		s.conns = append(s.conns, conn)
	}
	return s, err
}

func NewPullConn(endpoint string) (Sock, error) {
	var err error
	s := Sock{
		currentIdx: 0,
		sockType:   Pull,
	}

	addrParts := strings.Split(endpoint, "://")
	if len(addrParts) != 2 {
		return s, fmt.Errorf("malformed address")
	}

	s.listener, err = net.Listen(addrParts[0], addrParts[1])
	if err != nil {
		return s, err
	}
	go s.Serve()
	fmt.Println("inside NewPullConn", &s)

	return s, err
}

func (s Sock) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("error in accepting")
			return err
		}

		fmt.Println("got incoming", conn)
		zmtpGreetOutgoing := &greeter{
			sockType: s.sockType,
		}

		err = zmtpGreetOutgoing.greet(conn)
		if err != nil {
			fmt.Println("error in sending greet")
			return err
		}

		s.conns = append(s.conns, conn)
		fmt.Println("addded conn", s.conns)
		go s.handleConnection(conn)
	}
}

func (s Sock) handleConnection(conn net.Conn) error {
	fmt.Println("inside handleConnection", s.conns)
	buffer := make([]byte, 256)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("error in receiving")
		return err
	}

	more, length, msg := buffer[0], buffer[1], buffer[2:n]

	if more != 0 {
		fmt.Errorf("Sorry. I do not support multiframe messages right now.")
	}

	if int(length) != len(buffer)-2 {
		fmt.Errorf("Length frame does not match length of body")
	}

	fmt.Println("read bytes:", n)
	fmt.Println(msg[:n])

	return nil
}

func (s Sock) Connect(endpoint string) error {
	addrParts := strings.Split(endpoint, "://")
	_, err := net.Dial(addrParts[0], addrParts[1])
	if err != nil {
		return err
	}
	return err
}

func (s Sock) CurrentIdx() int {
	return s.currentIdx
}

func (s Sock) Read(b []byte) (n int, err error) {
	n, err = s.conns[s.currentIdx].Read(b)
	if err != nil {
		return 0, err
	}
	return n, err
}

func (s Sock) Write(b []byte) (n int, err error) {
	zmtpMessageOutgoing := &message{}
	zmtpMessageOutgoing.msg = [][]byte{b}
	n, err = zmtpMessageOutgoing.send(s.conns[s.currentIdx])
	s.currentIdx++
	return n, err
}

func (s Sock) Close() error {
	var err error
	for i := 0; i < len(s.conns); i++ {
		err = s.conns[i].Close()
		if err != nil {
			return err
		}
	}
	return err
}

func (s Sock) LocalAddr() net.Addr {
	return s.conns[s.currentIdx].LocalAddr()
}

func (s Sock) RemoteAddr() net.Addr {
	return s.conns[s.currentIdx].RemoteAddr()
}

func (s Sock) SetDeadline(t time.Time) error {
	return fmt.Errorf("I don't do deadlines")
}

func (s Sock) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("I don't do deadlines")
}

func (s Sock) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("I don't do deadlines")
}
