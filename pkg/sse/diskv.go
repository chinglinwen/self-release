// Store backend for diskv.
//
// Using this as the backend of store package,
// If for registering backend only (it can import as blank identifier).
package sse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/chinglinwen/log"

	"github.com/peterbourgon/diskv"
)

// need to init.
var defaultLogsPath string = "projectlogs"

// Init package with logs path for persistent.
func Init(logspath string) {
	log.Printf("init sse logspath: %v\n", logspath)

	if _, err := os.Stat(logspath); os.IsNotExist(err) {
		log.Printf("init sse create logspath: %v dir\n", logspath)
		err = os.MkdirAll(logspath, os.ModePerm)
		if err != nil {
			log.Fatalf("init logspath: %v err: %v\n", logspath, err)
		}
	} else {
		log.Printf("init sse logspath dir: %v exist. skip create\n", logspath)
	}
	defaultLogsPath = logspath
}

var disk = diskv.New(diskv.Options{
	BasePath:     defaultLogsPath,
	CacheSizeMax: 1024 * 1024,
})

// try store the command too, for restart a build ( mostly last time build )
func WriteFile(key string, b *Broker) (err error) {
	s, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return
	}
	return disk.Write(key, s)
}

func ReadFile(key string) (b *Broker, err error) {
	s, err := disk.Read(key)
	if err != nil {
		return
	}
	b = &Broker{}
	err = json.Unmarshal(s, b)
	return
}

func GetBrokersFromDisk() (bs []*Broker, err error) {
	keys, err := readfilenames()
	if err != nil {
		err = fmt.Errorf("readfilenames err %v", err)
		return
	}

	for _, v := range keys {
		b, err := ReadFile(v)
		if err != nil {
			log.Printf("read key: %v err: %v\n", v, err)
			continue
		}
		bs = append(bs, b)
	}

	// sort.Slice(bs, func(i, j int) bool {
	// 	// fmt.Printf("i:%#v\nj:%#v\n", bs[i], bs[j])
	// 	var t1, t2 time.Time
	// 	if bs[i].Event != nil {
	// 		t1 = bs[i].Event.Time
	// 	}
	// 	if bs[j].Event != nil {
	// 		t2 = bs[j].Event.Time
	// 	}
	// 	if bs[i].Event == nil && bs[j].Event == nil {
	// 		return false
	// 	}
	// 	return t1.After(t2) // recent first?
	// })
	return
}

func readfilenames() (keys []string, err error) {
	log.Debug.Printf("read logs from logspath: %v\n", defaultLogsPath)
	files, err := ioutil.ReadDir(defaultLogsPath)
	if err != nil {
		return
	}
	for _, file := range files {
		keys = append(keys, file.Name())
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j] // recent first?
	})
	return
}

var cutoff = 31 * 24 * time.Hour

// how to clean though? run a shell?
func clean() {
	fileInfo, err := ioutil.ReadDir(defaultLogsPath)
	if err != nil {
		log.Printf("==doing clean of logs err: %v\n", err)
		return
	}
	now := time.Now()
	for _, info := range fileInfo {
		if diff := now.Sub(info.ModTime()); diff > cutoff {
			key := info.Name()
			err := disk.Erase(key)
			if err != nil {
				log.Printf("==deleting %v err: %v\n", key, err)
				continue
			}
			// log.Printf("deleted %s which is %s old\n", key, diff)
		}
	}
	log.Println("==done of clean logs")
}

func init() {
	go func() {
		for {
			time.Sleep(1 * 24 * time.Hour)
			clean()
		}
	}()
}
