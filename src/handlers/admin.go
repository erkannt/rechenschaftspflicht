package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/services/userstore"
	"github.com/julienschmidt/httprouter"
)

type addUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

func AddUserHandler(userStore userstore.UserStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var req addUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Username == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		exists, err := userStore.IsUser(req.Email)
		if err != nil {
			slog.Error("failed to check if user exists", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exists {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if err := userStore.AddUser(req.Email, req.Username); err != nil {
			slog.Error("failed to add user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
