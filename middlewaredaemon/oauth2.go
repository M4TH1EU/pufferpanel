/*
 Copyright 2016 Padduck, LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 	http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package middlewaredaemon

import (
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/pufferpanel/v3"
	"github.com/pufferpanel/pufferpanel/v3/logging"
	"github.com/pufferpanel/pufferpanel/v3/programs"
	"github.com/pufferpanel/pufferpanel/v3/response"
	"net/http"
	"runtime/debug"
	"strings"
)

func OAuth2Handler(requiredScope pufferpanel.Scope, requireServer bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		failure := true
		defer func() {
			if err := recover(); err != nil {
				logging.Error.Printf("Error handling auth check: %s\n%s", err, debug.Stack())
				failure = true
			}
			if failure && !c.IsAborted() {
				c.AbortWithStatus(500)
			}
		}()

		authHeader := c.Request.Header.Get("Authorization")
		var authToken string
		if authHeader == "" {
			authToken = c.Query("accessToken")
			if authToken == "" {
				response.HandleError(c, pufferpanel.ErrMissingAccessToken, http.StatusBadRequest)
				return
			}
		} else {
			authArr := strings.SplitN(authHeader, " ", 2)
			if len(authArr) < 2 || authArr[0] != "Bearer" {
				response.HandleError(c, pufferpanel.ErrNotBearerToken, http.StatusBadRequest)
				return
			}
			authToken = authArr[1]
		}

		//TODO: we need to know what scopes you have....
		var scopes []pufferpanel.Scope
		var serverId string

		if !pufferpanel.ContainsScope(scopes, requiredScope) {
			response.HandleError(c, pufferpanel.CreateErrMissingScope(requiredScope), http.StatusForbidden)
			return
		}

		if requireServer {
			program, _ := programs.Get(serverId)
			if program == nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.Set("server", program)
		}

		c.Set("scopes", scopes)

		failure = false
	}
}