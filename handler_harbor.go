package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

func harborHandler(c echo.Context) error {
	//may do redirect later?
	r := c.Request()

	// spew.Dump(r.Header)

	fmt.Printf("r: %#v\n", r)
	// spew.Dump(r.Header)

	var buf bytes.Buffer
	// b, err := r.GetBody()
	// if err != nil {
	// 	log.Println("getbody err", err)
	// }
	buf.ReadFrom(r.Body)
	body := buf.String()
	log.Printf("body: %v", body)

	return c.String(http.StatusOK, "ok")
}
