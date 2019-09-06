package k8s

import (
	"fmt"
	"testing"
)

func TestServiceList(t *testing.T) {
	ss, err := ServiceList("xindaiquan")
	if err != nil {
		t.Error("ServiceList err", err)
		return
	}
	// b, _ := json.MarshalIndent(ss, "", "  ")
	// fmt.Println(string(b))
	for _, v := range ss {
		fmt.Println(v.GetMetadata().GetNamespace(), v.GetMetadata().GetName())
	}
}

func TestServiceListWithLabels(t *testing.T) {
	l := map[string]string{"codis-component": "proxy"}
	ss, err := ServiceListWithLabels("codis-cluster", l)
	if err != nil {
		t.Error("ServiceList err", err)
		return
	}
	// b, _ := json.MarshalIndent(ss, "", "  ")
	// fmt.Println(string(b))
	for _, v := range ss {
		var port int32
		for _, x := range v.GetSpec().GetPorts() {
			if x.GetName() == "proxy" {
				port = x.GetPort()
			}
		}
		fmt.Println(v.GetMetadata().GetNamespace(), v.GetMetadata().GetName(), port)
	}
}

func TestServiceListAll(t *testing.T) {
	ss, err := ServiceListAll()
	if err != nil {
		t.Error("ServiceListAll err", err)
		return
	}
	// b, _ := json.MarshalIndent(ss, "", "  ")
	// fmt.Println(string(b))
	for _, v := range ss {
		fmt.Println(v.GetMetadata().GetNamespace(), v.GetMetadata().GetName())
	}
}
