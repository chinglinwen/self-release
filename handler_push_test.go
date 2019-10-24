package main

import (
	"fmt"
	"testing"
)

/*
existing env

$ awk '{ print $2 }' FS='{{'  a| tr -d '}' | grep -v -e '^$' | sort -n | uniq
 $CI_ENV
 $CI_IMAGE
 $CI_NAMESPACE
 $CI_NAMESPACE ,project=
 $CI_PROJECT_NAME
 $CI_PROJECT_NAME_WITH_ENV
 $CI_REPLICAS
 $CI_TIME
 $CI_USER_NAME
 $NODE_PORT  # ????
$
*/

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
