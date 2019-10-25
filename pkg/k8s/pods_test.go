package k8s

import (
	"testing"
)

func init() {
	Init("")
}

func TestPodListInfo(t *testing.T) {
	ss, err := PodListInfo("flow_center/hamburg")
	if err != nil {
		t.Error("ServiceList err", err)
		return
	}
	// b, _ := json.MarshalIndent(ss, "", "  ")
	// fmt.Println(string(b))
	// for _, v := range ss {
	// 	fmt.Println(v.GetMetadata().GetNamespace(), v.GetMetadata().GetName())
	// }
	pretty("pods", ss)
}
