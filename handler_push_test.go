package main

import (
	"fmt"
	"testing"
)

func TestCheckIsHeader(t *testing.T) {
	x := checkIsHeader("<h2>Info</h2>")
	if !x {
		t.Error("checkIsHeader err, should be true, got false")
		return
	}

	x = checkIsHeader("Info</h")
	if x {
		t.Error("checkIsHeader err, should be false, got true")
		return
	}
}

func TestGetProjectURL(t *testing.T) {
	fmt.Printf("url: %v\n", getProjectURL("robot/mileage-planet", "pre"))
	fmt.Printf("url: %v\n", getProjectURL("robot/mileage-planet", "test"))
}
