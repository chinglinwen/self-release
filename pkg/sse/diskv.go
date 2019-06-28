// Store backend for diskv.
//
// Using this as the backend of store package,
// If for registering backend only (it can import as blank identifier).
package sse

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/peterbourgon/diskv"
)

var logsPath = flag.String("logsDir", "projectlogs", "build logs dir")

var disk = diskv.New(diskv.Options{
	BasePath:     *logsPath,
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
	return
}

func readfilenames() (keys []string, err error) {
	files, err := ioutil.ReadDir(*logsPath)
	if err != nil {
		return
	}
	for _, file := range files {
		keys = append(keys, file.Name())
	}
	return
}

var cutoff = 31 * 24 * time.Hour

// how to clean though? run a shell?
func clean() {
	fileInfo, err := ioutil.ReadDir(*logsPath)
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
