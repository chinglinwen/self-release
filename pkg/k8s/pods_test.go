package k8s

import (
	"fmt"
	"testing"
)

func init() {
	Init("")
}

func TestPodListInfo(t *testing.T) {
	ss, err := PodListInfo("haodai/main")
	if err != nil {
		t.Error("PodListInfo err", err)
		return
	}
	// b, _ := json.MarshalIndent(ss, "", "  ")
	// fmt.Println(string(b))
	for _, v := range ss {
		fmt.Println(v.PodName, v.Env)
	}
	// pretty("pods", ss)
}
