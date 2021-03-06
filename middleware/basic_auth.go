// Copyright 2017 King Qiu.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
// https://github.com/qjw/kelly

// https://github.com/martini-contrib/auth

package middleware

import (
	"github.com/qjw/kelly"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

// SecureCompare performs a constant time compare of two strings to limit timing attacks.
func secureCompare(given string, actual string) bool {
	givenSha := sha256.Sum256([]byte(given))
	actualSha := sha256.Sum256([]byte(actual))

	return subtle.ConstantTimeCompare(givenSha[:], actualSha[:]) == 1
}

// BasicRealm is used when setting the WWW-Authenticate response header.
var BasicRealm = "Authorization Required"

// Basic returns a Handler that authenticates via Basic Auth. Writes a http.StatusUnauthorized
// if authentication fails.
func Basic(username string, password string) kelly.HandlerFunc {
	var siteAuth = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return func(c *kelly.Context) {
		if auth,err := c.GetHeader("Authorization");err == nil{
			if !secureCompare(auth, "Basic "+siteAuth) {
				unauthorized(c)
				return
			}
			c.Set("user",username)
			c.InvokeNext()
		}else{
			unauthorized(c)
		}
	}
}

// BasicFunc returns a Handler that authenticates via Basic Auth using the provided function.
// The function should return true for a valid username/password combination.
func BasicFunc(authfn func(string, string) bool) kelly.HandlerFunc {
	return func(c *kelly.Context) {
		auth,err := c.GetHeader("Authorization")
		if err != nil{
			unauthorized(c)
			return
		}

		if len(auth) < 6 || auth[:6] != "Basic " {
			unauthorized(c)
			return
		}
		b, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			unauthorized(c)
			return
		}
		tokens := strings.SplitN(string(b), ":", 2)
		if len(tokens) != 2 || !authfn(tokens[0], tokens[1]) {
			unauthorized(c)
			return
		}
		c.Set("user",tokens[0])
		c.InvokeNext()
	}
}

func unauthorized(res http.ResponseWriter) {
	res.Header().Set("WWW-Authenticate", "Basic realm=\""+BasicRealm+"\"")
	http.Error(res, "Not Authorized", http.StatusUnauthorized)
}