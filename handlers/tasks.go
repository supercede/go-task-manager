package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"todo-app/models"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warning(err)
		RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid JSON provided")
		return
	}

	validate := validator.New()
	err = validate.Struct(t)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	user := req.Context().Value(KeyUser{}).(*models.User)

	task, err := h.store.CreateTask(user, t)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// h.store.DB.Model(user).Association("Tasks").Append(task)

	data := map[string]interface{}{
		"task": task,
	}

	res := Response{"Task Created sucessfully", data}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warning(err)
		RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user := req.Context().Value(KeyUser{}).(*models.User)
	tasks, err := h.store.GetTasks(user)
	if err != nil {
		log.Warning("Failed to fetch tasks")
		RespondJSON(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	data := map[string]interface{}{
		"tasks": tasks,
	}

	res := Response{"Fetched tasks successfully", data}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warning(err)
		RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user := req.Context().Value(KeyUser{}).(*models.User)
	vars := mux.Vars(req)
	id := vars["id"]
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Warning("Failed to parse task ID")
		RespondJSON(w, http.StatusBadRequest, "Invalid Id")
		return
	}

	task, err := h.store.GetTask(user, intID)
	data := map[string]interface{}{
		"task": task,
	}

	res := Response{"Fetched task successfully", data}
	RespondJSON(w, http.StatusOK, &res)
}
