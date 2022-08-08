package server

import (
	"net/http"
)

func payloadLimit(limit uint, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > int64(limit) {
			http.Error(w, "Submitted secret too large", http.StatusRequestEntityTooLarge)
			return
		}

		next(w, r)
	}
}

/*
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(
				http.StatusRequestEntityTooLarge,
				"payload is too large",
			)
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}

*/
