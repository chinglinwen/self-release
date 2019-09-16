package nfs

import (
	"bufio"
	"strings"
)

type Entry struct {
	Name   string `json:"name,omitempty"`
	Path   string `json:"path,omitempty"`
	Server string `json:"server,omitempty"`
}

func Parse(body, server string) (results []Entry, err error) {
	scanner := bufio.NewScanner(strings.NewReader(body))
	var name string
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		if strings.Contains(text, "#") {
			s := strings.Fields(text)
			if len(s) >= 2 {
				name = s[1]
			}
			continue
		}
		path := strings.Fields(text)[0]
		results = append(results, Entry{
			Name:   name,
			Path:   path,
			Server: server,
		})
	}
	return
}
