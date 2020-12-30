package middleware

import (
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
			handlers.RespondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}
