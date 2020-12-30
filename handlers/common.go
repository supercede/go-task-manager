package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"todo-app/database"
	"todo-app/util/auth"

	"github.com/pkg/errors"
)

type Handler struct {
	store *database.Store
	tk    auth.TokenInterface
	au    auth.AuthInterface
}

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type KeyUser struct{}

func NewHandler(s *database.Store, tk auth.TokenInterface, au auth.AuthInterface) *Handler {
	return &Handler{s, tk, au}
}

func (h *Handler) CheckAuth(r *http.Request) (*http.Request, error) {
	metadata, err := h.tk.GetTokenMetadata(r)
	if err != nil {
		return nil, errors.Wrap(err, "Error fetching token metadata")
	}
	strUserId, err := h.au.FetchAuth(metadata.TokenUuid)
	if err != nil {
		return nil, errors.Wrap(err, "Error fetching user auth details from redis store")
	}

	userID, err := strconv.ParseUint(strUserId, 10, 64)
	if err != nil {
		return nil, err
	}

	user, err := h.store.GetUserById(uint(userID))
	if err != nil {
		return nil, err
	}

	// ctx := context.WithValue(r.Context(), "userId", userId)
	ctx := context.WithValue(r.Context(), KeyUser{}, user)
	req := r.WithContext(ctx)

	return req, nil
}

func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

// respondError makes the error response with payload as json format
func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"status": "error", "error": message})
}
