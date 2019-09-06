package k8s

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSecretList(t *testing.T) {
	ss, err := SecretList("")
	if err != nil {
		t.Error("secretlist err", err)
		return
	}
	fmt.Println("got ", len(ss))
	b, _ := json.MarshalIndent(ss, "", "  ")
	fmt.Println(string(b))
}

func TestSecretListAllWithHasKey(t *testing.T) {
	excludens := []string{"default", "cron"}
	ss, err := SecretListAllWithHasKey("", "database", excludens)
	if err != nil {
		t.Error("secretlist err", err)
		return
	}
	fmt.Println("got ", len(ss))
	b, _ := json.MarshalIndent(ss, "", "  ")
	fmt.Println(string(b))
}
