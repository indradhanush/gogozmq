package gogozmq

import (
	"fmt"
	"testing"
	"time"

	"github.com/zeromq/goczmq"
)

func TestPushSockShortMessage(t *testing.T) {
	endpoint1 := "tcp://127.0.0.1:9998"

	pull1, err := goczmq.NewPull(endpoint1)
	if err != nil {
		t.Fatal(err)
	}
	defer pull1.Destroy()

	push, err := NewPushConn(endpoint1)
	if err != nil {
		t.Fatal(err)
	}

	push.Write([]byte("Hello"))

	msg, more, err := pull1.RecvFrame()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := "Hello", string(msg); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := 5, len(msg); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := 0, more; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func TestPullSockSingleConnection(t *testing.T) {
	endpoint := "tcp://127.0.0.1:9998"

	pull, err := NewPullConn(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("new pull: ", &pull)

	push, err := NewPushConn(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("new push: ", &push)
	// t.Log("*push: ", *push)
	fmt.Printf("type of push: %T\n", push)
	fmt.Printf("ntype of &push: %T\n", &push)
	fmt.Printf("address of push: %p\n", push)
	fmt.Printf("address of &push: %p\n", &push)
	// t.Log("type of *push: %T", *push)

	push.Write([]byte("Hello"))
	time.Sleep(200)
	t.Log("pull.conns: ", pull.conns)

	// for {
	// 	if len(pull.conns) > 0 {
	// 		return
	// 	}
	// }

	// t.Log("pull hello")
	// var msg []byte
	// n, err := pull.Read(msg)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// t.Log("Bytes read: ", n)
}
