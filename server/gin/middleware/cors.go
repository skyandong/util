package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Cors Web中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		origin := c.Request.Header.Get("Origin")
		headerStr := "access-control-allow-origin, access-control-allow-headers, Connection, User-Agent, Accept, Access-Control-Request-Method, Origin, Access-Control-Request-Headers, Referer, Accept-Encoding, Accept-Language,content-type,Authorization,withcredentials,changeOrigin,cookie"
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", headerStr)
		}

		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}

		c.Next()
	}
}
