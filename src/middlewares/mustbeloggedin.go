package middlewares

import (
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/services/authentication"
	"github.com/julienschmidt/httprouter"
)

func MustBeLoggedIn(auth authentication.Auth) func(httprouter.Handle) httprouter.Handle {
	return func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			if !auth.IsLoggedIn(r) {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			h(w, r, ps)
		}
	}
}
