package notify

import (
	"fmt"
	"testing"
)

func TestSendPerson(t *testing.T) {
	reply, err := SendPerson("hello", "wenzhenglin")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(reply)
}
