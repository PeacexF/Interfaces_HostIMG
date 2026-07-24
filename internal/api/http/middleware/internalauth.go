package middleware

import (
	"crypto/subtle"
	"net/http"
)

func RequireInternalSecret(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if len(provided) <= len(prefix) || provided[:len(prefix)] != prefix {
				writeJSONError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
				return
			}
			token := provided[len(prefix):]

			if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
				writeJSONError(w, http.StatusUnauthorized, "invalid internal secret")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
