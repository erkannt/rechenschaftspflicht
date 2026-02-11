package middlewares

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func RequireBearerToken(bearerToken string) func(httprouter.Handle) httprouter.Handle {
	return func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if parts[1] != bearerToken {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h(w, r, ps)
		}
	}
}
