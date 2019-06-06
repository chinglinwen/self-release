package sse

import (
	"fmt"
	"os"
	"testing"

	"github.com/peterbourgon/diskv"
)

func init() {
	*logsPath = "../../projectlogs"
	os.MkdirAll(*logsPath, 0755)

	disk = diskv.New(diskv.Options{
		BasePath:     *logsPath,
		CacheSizeMax: 1024 * 1024,
	})
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