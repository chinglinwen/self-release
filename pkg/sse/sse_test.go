package sse

import (
	"fmt"
	"os"
	"testing"

	"github.com/chinglinwen/log"
	"github.com/peterbourgon/diskv"
)

func init() {
	if os.Getenv("GODEBUG") != "" {
		log.SetLevel("debug")
		log.Debug.Println("got debug env, set log level to ", "debug")
	}
}

func init() {
	*logsPath = "../../projectlogs"
	os.MkdirAll(*logsPath, 0755)

	disk = diskv.New(diskv.Options{
		BasePath:     *logsPath,
		CacheSizeMax: 1024 * 1024,
	})
}

func TestParseEventInfoJson(t *testing.T) {

	b := "{\"namespace\":\"demo\",\"project\":\"hello\",\"branch\":\"v1.0.0\",\"Env\":\"\",\"time\":\"0001-01-01T00:00:00Z\"}"

	i, err := ParseEventInfoJson(b)
	if err != nil {
		t.Errorf("ParseEventInfoJson err %v", err)
		return
	}
	fmt.Println("got info: ", i)
}
func TestGetBrokers(t *testing.T) {
	b := New("prjoecta", "brancha")
	fmt.Fprint(b.PWriter, "created ")
	b.Close()

	bs, err := GetBrokers()
	if err != nil {
		t.Errorf("GetBrokers err %v", err)
		return
	}
	for _, v := range bs {
		fmt.Println(v.Key)
		// fmt.Println("logs:", v.ExistMsg)
	}
}

func TestGetBrokerFromPerson(t *testing.T) {
	dev := "robot"
	b, err := GetBrokerFromPerson(dev)
	if err != nil {
		t.Errorf("cant find previous released project, err: %v", err)
		return
	}

	fmt.Printf("project: %v, branch: %v\n", b.Project, b.Branch)
}

func TestGetGetBrokersFromDisk(t *testing.T) {
	bs, err := GetBrokersFromDisk()
	if err != nil {
		t.Error("GetBrokersFromDisk err", err)
		return
	}
	for _, v := range bs {
		fmt.Printf("%v, project: %v, branch: %v\n", v.Event.Time, v.Project, v.Branch)

	}

}

func TestWriteFile(t *testing.T) {
	b := New("prjoecta", "brancha")
	fmt.Fprint(b.PWriter, "created ")
	b.Close()

	err := WriteFile(b.Key, b)
	if err != nil {
		t.Errorf("WriteFile err %v", err)
		return
	}

	b1, err := ReadFile(b.Key)
	if err != nil {
		t.Errorf("ReadFile err %v", err)
		return
	}
	if b1.Key != b.Key || len(b1.ExistMsg) != len(b.ExistMsg) {
		t.Errorf("ReadFile contents err %v", err)
		return
	}

}
