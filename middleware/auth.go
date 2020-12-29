package middleware

import (
	"fmt"
	"net/http"
	"todo-app/handlers"
	"todo-app/util/auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := auth.CheckTokenValidity(r)
		if err != nil {
			fmt.Println(err)
			handlers.RespondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}
