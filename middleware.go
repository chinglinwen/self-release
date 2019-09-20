package main

import (
	"fmt"
	"net/http"

	"github.com/chinglinwen/log"
	"github.com/labstack/echo"
)

func loginCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// spew.Dump("resp header", c.Response().Header())

			// r := c.Request()
			// spew.Dump("header", r.Header)

			cookie, err := c.Cookie("token")
			if err != nil {
				err := fmt.Errorf("login required")
				log.Println(err)
				return c.JSONPretty(http.StatusOK, E(-1, err.Error(), "failed"), " ")
			}

			user, usertoken, err := validateJWT(cookie.Value)
			if err != nil {
				err := fmt.Errorf("token invalid")
				log.Println(err)
				return c.JSONPretty(http.StatusOK, E(-2, err.Error(), "failed"), " ")
			}

			if user == "" || usertoken == "" {
				err := fmt.Errorf("user or token is empty")
				log.Println(err)
				return c.JSONPretty(http.StatusOK, E(-3, err.Error(), "failed"), " ")
			}

			c.Request().Header.Set("X-Auth-User", user)
			c.Request().Header.Set("X-Secret", usertoken)

			// user := r.Header.Get("X-Auth-User")
			// log.Printf("got user: %v\n", user)
			// usertoken := r.Header.Get("X-Secret")
			// if user == "" || usertoken == "" {
			// 	err := fmt.Errorf("login required")
			// 	log.Println(err)
			// 	return c.JSONPretty(http.StatusOK, E(-1, err.Error(), "failed"), " ")
			// }
			return next(c)
		}
	}
}
