package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"todo-app/models"
	"todo-app/ws"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warningf("auth error: %s", err.Error())
		RespondError(w, http.StatusUnauthorized, err.Error())
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

	res := Response{"success", "Task Created sucessfully", data}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warningf("auth error: %s", err.Error())
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	v := r.URL.Query()
	completed := v.Get("completed")
	priority := v.Get("priority")

	params := make(map[string]interface{})

	if completed != "" {
		params["completed"] = completed
	}

	if priority != "" {
		params["priority"] = priority
	}

	user := req.Context().Value(KeyUser{}).(*models.User)
	tasks, err := h.store.GetTasks(user, params)
	if err != nil {
		log.Warning("Failed to fetch tasks")
		RespondError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	data := map[string]interface{}{
		"tasks": tasks,
	}

	res := Response{"success", "Fetched tasks successfully", data}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warningf("auth error: %s", err.Error())
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user := req.Context().Value(KeyUser{}).(*models.User)
	vars := mux.Vars(req)
	id := vars["id"]
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Warning("Failed to parse task ID")
		RespondError(w, http.StatusBadRequest, "Invalid Id")
		return
	}

	task, err := h.store.GetTask(user, intID)
	if err != nil {
		log.Warningf("Failed to Get Task: %s", err.Error())
		if err.Error() == "record not found" {
			RespondError(w, http.StatusNotFound, "Task not found")
			return
		}
		RespondError(w, http.StatusUnprocessableEntity, "Failed to Get Task")
		return
	}

	data := map[string]interface{}{
		"task": task,
	}

	res := Response{"success", "Fetched task successfully", data}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) AddUserToTask(w http.ResponseWriter, r *http.Request) {
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warningf("auth error: %s", err.Error())
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	c, err := ws.Connect()
	if err != nil {
		log.Error("dial:", err)
	}

	authUser := req.Context().Value(KeyUser{}).(*models.User)
	vars := mux.Vars(req)
	idUser, idTask := vars["idUser"], vars["idTask"]
	var intIDUser, intIDTask int
	intIDUser, err = strconv.Atoi(idUser)
	intIDTask, err = strconv.Atoi(idTask)
	if err != nil {
		log.Warning("Failed to parse parameter IDs")
		RespondError(w, http.StatusBadRequest, "Invalid User or Task ID")
		return
	}

	user, task, err := h.store.AddUserToTask(&models.User{}, models.Task{}, intIDUser, intIDTask)
	if err != nil {
		if err.Error() == "record not found" {
			RespondError(w, http.StatusNotFound, "Task not found")
			return
		}
		log.Warning(err.Error())
		RespondError(w, http.StatusBadRequest, "Failed to add user to task")
		return
	}

	msg := message{
		Username: authUser.Username,
		Action:   "Add User",
		Message:  fmt.Sprintf("%s added %s to task: %s", authUser.Username, user.Username, task.Title),
		Task:     task.ID,
	}
	go func() {
		err := c.WriteJSON(msg)
		if err != nil {
			log.Warn("write:", err)
			return
		}
	}()

	data := map[string]interface{}{
		"user": user,
	}

	res := Response{"success", "Added user to task successfully", data}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) RemoveUserFromTask(w http.ResponseWriter, r *http.Request) {
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warningf("auth error: %s", err.Error())
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	authUser := req.Context().Value(KeyUser{}).(*models.User)
	vars := mux.Vars(req)
	idUser, idTask := vars["idUser"], vars["idTask"]
	var intIDUser, intIDTask int
	intIDUser, err = strconv.Atoi(idUser)
	intIDTask, err = strconv.Atoi(idTask)
	if err != nil {
		log.Warning("Failed to parse parameter IDs")
		RespondError(w, http.StatusBadRequest, "Invalid User or Task ID")
		return
	}

	user, task, err := h.store.RemoveUserFromTask(&models.User{}, models.Task{}, intIDUser, intIDTask)

	if err != nil {
		log.Warning(err.Error())
		RespondError(w, http.StatusBadRequest, "Failed to remove user from task")
		return
	}

	c, err := ws.Connect()
	if err != nil {
		log.Error("dial Error:", err)
	}

	msg := message{
		Username: authUser.Username,
		Action:   "Remove User",
		Message:  fmt.Sprintf("%s removed %s from task: %s", authUser.Username, user.Username, task.Title),
		Task:     task.ID,
	}
	go func() {
		err := c.WriteJSON(msg)
		if err != nil {
			log.Warn("write:", err)
			return
		}
	}()

	res := Response{"success", "Removed user from task successfully", nil}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warningf("auth error: %s", err.Error())
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var t models.UpdateTask
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

	vars := mux.Vars(req)
	id := vars["id"]
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Warning("Failed to parse task ID")
		RespondError(w, http.StatusBadRequest, "Invalid Id")
		return
	}

	user := req.Context().Value(KeyUser{}).(*models.User)
	task, err := h.store.UpdateTask(user, t, intID)
	if err != nil {
		log.Warningf("Update task error: %s", err.Error())
		if err.Error() == "record not found" {
			RespondError(w, http.StatusNotFound, "Task not found")
			return
		}
		RespondError(w, http.StatusBadRequest, "Failed to update task")
		return
	}

	c, err := ws.Connect()
	if err != nil {
		log.Error("dial Error:", err)
	}

	msg := message{
		Username: user.Username,
		Action:   "Update",
		Message:  fmt.Sprintf("%s updated task: %s", user.Username, task.Title),
		Task:     task.ID,
	}
	go func() {
		err := c.WriteJSON(msg)
		if err != nil {
			log.Warn("write:", err)
			return
		}
	}()

	data := map[string]interface{}{
		"task": task,
	}

	res := Response{"success", "Updated task successfully", data}
	RespondJSON(w, http.StatusOK, &res)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	req, err := h.CheckAuth(r)
	if err != nil {
		log.Warningf("auth error: %s", err.Error())
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	vars := mux.Vars(req)
	id := vars["id"]
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Warning("Failed to parse task ID")
		RespondError(w, http.StatusBadRequest, "Invalid Id")
		return
	}

	user := req.Context().Value(KeyUser{}).(*models.User)
	task, err := h.store.DeleteTask(user, intID)

	if err != nil {
		log.Warningf("Delete task error: %s", err.Error())
		if err.Error() == "record not found" {
			RespondError(w, http.StatusNotFound, "Task not found")
			return
		}
		RespondError(w, http.StatusBadRequest, "Failed to delete task")
		return
	}

	c, err := ws.Connect()
	if err != nil {
		log.Error("dial Error:", err)
	}

	msg := message{
		Username: user.Username,
		Action:   "Delete",
		Message:  fmt.Sprintf("%s deleted task: %s", user.Username, task.Title),
		Task:     task.ID,
	}
	go func() {
		err := c.WriteJSON(msg)
		if err != nil {
			log.Warn("write:", err)
			return
		}
	}()

	res := Response{"success", "Deleted task successfully", nil}
	RespondJSON(w, http.StatusOK, &res)
}
