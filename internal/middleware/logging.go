package middleware

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

func LogAccessRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/request") {
			log.Info().
				Str("user", r.Header.Get("X-WebAuth-User")).
				Str("email", r.Header.Get("X-WebAuth-eMail")).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("access request recorded")
		}
		next.ServeHTTP(w, r)
	})
}
