package resource

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetResource(t *testing.T) {
	r, err := GetResource("xindaiquan")
	if err != nil {
		t.Error("GetResource err", err)
		return
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(b))
}
