package gogozmq

import (
	"fmt"
	"testing"
)

func TestPushPullChanneler(t *testing.T) {
	endpoint := "tcp://127.0.0.1:9999"

	pull, err := NewPullChanneler(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer pull.Destroy()

	push, err := NewPushChanneler(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer push.Destroy()

	push.SendChan <- []byte("Hello")
	msg, _ := <-pull.RecvChan

	if want, have := "Hello", string(msg); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	fmt.Println("test has completed, but I am blocking somewhere")
}
