package notify

import (
	"fmt"
	"testing"
)

func TestSend(t *testing.T) {
	reply, err := Send("wenzhenglin", "hello2")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(reply)
}
