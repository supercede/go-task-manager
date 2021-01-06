package middleware

import (
	"fmt"
	"net/http"
	"todo-app/handlers"
	"todo-app/util/auth"

	log "github.com/sirupsen/logrus"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := auth.CheckTokenValidity(r)
		if err != nil {
			log.Warning(err)
			handlers.RespondError(w, http.StatusUnauthorized, fmt.Sprint("Unauthorized:", err))
			return
		}
		next.ServeHTTP(w, r)
	})
}
