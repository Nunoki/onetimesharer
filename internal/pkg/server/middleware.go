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
