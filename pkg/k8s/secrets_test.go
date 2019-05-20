package k8s

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSecretList(t *testing.T) {
	ss, err := SecretList("xindaiquan")
	if err != nil {
		t.Error("secretlist err", err)
		return
	}
	b, _ := json.MarshalIndent(ss, "", "  ")
	fmt.Println(string(b))
}
