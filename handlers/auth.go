package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"todo-app/models"
	"todo-app/util"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Password string `json:"password" validate:"required"`
	Username string `json:"username" validate:"required,gte=3"`
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var u models.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid JSON provided")
		return
	}

	validate := validator.New()
	err = validate.Struct(u)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	u.Password = string(hashPassword)

	user, err := h.store.AddUser(u)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	res := Response{"success", "Sign up successful", user}
	RespondJSON(w, http.StatusCreated, &res)
}

func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)

	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid JSON provided")
		return
	}

	validate := validator.New()
	err = validate.Struct(creds)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.store.GetUserByUsername(creds.Username)
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		RespondError(w, http.StatusUnauthorized, "Invalid Password")
		return
	}

	tokens, err := h.tk.CreateToken(user.ID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	authError := h.au.CreateAuth(user.ID, tokens)
	if authError != nil {
		RespondError(w, http.StatusInternalServerError, authError.Error())
		return
	}

	data := map[string]interface{}{
		"user":          user,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	}

	res := Response{"success", "Sign in successful", data}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	metadata, err := h.tk.GetTokenMetadata(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if metadata != nil {
		err = h.au.DeleteTokens(metadata)
		if err != nil {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	res := Response{"success", "Logout successful", nil}
	RespondJSON(w, http.StatusOK, &res)
}

// Generate new refresh and access tokens
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	conf, err := util.GetConfig()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to load config")
	}

	type refresh struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	var mapToken refresh
	// mapToken := map[string]string{}

	err = json.NewDecoder(r.Body).Decode(&mapToken)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid JSON provided")
		return
	}

	validate := validator.New()
	err = validate.Struct(mapToken)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// refreshToken := mapToken["refresh_token"]
	refreshToken := mapToken.RefreshToken

	// Verify Token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(conf.RefreshSecret), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refreshUUID, ok := claims["refresh_uuid"].(string)
		if !ok {
			log.Warning("Failed to get refresh UUID")
			RespondError(w, http.StatusUnprocessableEntity, "Unauthorized")
			return
		}

		// userId, idOk := claims["user_id"].(uint)
		floatUserId, userOk := claims["user_id"].(float64)
		userId := uint(floatUserId)
		if !userOk {
			log.Warning("Failed to get user ID")
			RespondError(w, http.StatusUnprocessableEntity, "Unauthorized")
			return
		}
		delErr := h.au.DeleteRefresh(refreshUUID)
		if delErr != nil {
			RespondError(w, http.StatusUnauthorized, delErr.Error())
			return
		}
		//Create new pairs of refresh and access tokens
		// uintUserID, _ := strconv.ParseUint(userId, 10, 32)
		ts, createErr := h.tk.CreateToken(userId)
		if createErr != nil {
			RespondError(w, http.StatusForbidden, createErr.Error())
			return
		}
		// saveErr := h.au.CreateAuth(uint(uintUserID), ts)
		saveErr := h.au.CreateAuth(userId, ts)
		if saveErr != nil {
			RespondError(w, http.StatusForbidden, createErr.Error())
			return
		}

		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}
		res := Response{"success", "Tokens created sucessfully", tokens}
		RespondJSON(w, http.StatusOK, &res)
		return
	} else {
		RespondError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}
}
