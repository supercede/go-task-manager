package handlers

import (
	"encoding/json"
	"net/http"
	"todo-app/models"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Password string `json:"password" validate:"required"`
	Username string `json:"username" validate:"required,gte=3"`
}

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var u models.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(u)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	u.Password = string(hashPassword)

	user, err := h.store.AddUser(u)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	res := Response{"Sign up successful", user.ReturnUser()}
	respondJSON(w, http.StatusCreated, &res)
}

func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(creds)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.store.GetUserByUsername(creds.Username)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid Password")
		return
	}

	res := Response{"Sign in successful", user.ReturnUser()}
	respondJSON(w, http.StatusOK, &res)
}
