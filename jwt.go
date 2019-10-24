package main

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
)

func validateJWT(token string) (user, usertoken string, err error) {
	if token == "" {
		err = fmt.Errorf("token is empty")
		return
	}
	t, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		err = fmt.Errorf("token parse err: %v", err)
		return
	}
	if !t.Valid {
		err = fmt.Errorf("token invalid")
		return
	}
	// spew.Dump("t", t)
	if claims, ok := t.Claims.(jwt.MapClaims); ok {
		if user, ok = claims["name"].(string); !ok {
			err = fmt.Errorf("get user name failed")
			return
		}
		if usertoken, ok = claims["token"].(string); !ok {
			err = fmt.Errorf("get user token failed")
			return
		}
	} else {
		err = fmt.Errorf("token clamins invalid")
	}
	return
}
