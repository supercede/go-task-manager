package handlers

// func (h *Handler) ValidateUser(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r http.Request) {
// 		var u models.User

// 		err := json.NewDecoder(r.Body).Decode(&u)
// 		if err != nil {
// 			respondError(w, http.StatusBadRequest, err.Error())
// 			return
// 		}
// 	})
// }
