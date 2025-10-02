package userhandler

import (
	"encoding/json"
	"errors"
	"github.com/koyif/gophermart/internal/domain"
	"github.com/koyif/gophermart/pkg/dto"
	"github.com/koyif/gophermart/pkg/logger"
	"io"
	"net/http"
)

type UserService interface {
	Register(username, password string) (string, error)
	Login(login, password string) (string, error)
}

type UserHandler struct {
	srv UserService
}

func New(srv UserService) *UserHandler {

	return &UserHandler{
		srv: srv,
	}
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var auth dto.Auth

	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		logger.Log.Warn("error while decoding a register request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			logger.Log.Error("error while closing request body", logger.Error(err))
			return
		}
	}(r.Body)

	if err := auth.IsValid(); err != nil {
		logger.Log.Warn("invalid auth fields", logger.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := uh.srv.Register(auth.Login, auth.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var auth dto.Auth

	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		logger.Log.Warn("error while decoding a register request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Log.Error("error while closing request body", logger.Error(err))
			return
		}
	}(r.Body)

	if err := auth.IsValid(); err != nil {
		logger.Log.Warn("invalid auth fields", logger.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := uh.srv.Login(auth.Login, auth.Password)
	if err != nil {
		if errors.Is(err, domain.ErrIncorrectCredentials) {
			http.Error(w, "incorrect login or password", http.StatusUnauthorized)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
