package api

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"net/http"

	"github.com/prithvipal/todo-app/data"
	"github.com/prithvipal/todo-app/models"
	log "github.com/sirupsen/logrus"
)

var (
	handlerFunc = map[string]func(w http.ResponseWriter, r *http.Request){
		"GET":    getHandler,
		"POST":   createHandler,
		"DELETE": deleteHandler,
		"PUT":    updateHandler,
		"PATCH":  partialUpdateHandler,
	}
)

type TodoHandler struct {
}

func (th TodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if hanlder, ok := handlerFunc[r.Method]; ok {
		hanlder(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := validateUrlAndExtractParam(r.URL.Path)
	if err != nil {
		log.Println("Could not parse request url", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = data.DeleteTodo(id)
	if err != nil {
		log.Println("key not found in database", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var todo models.Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		log.Println("Could not parse request payload", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if todo.Status != 0 {
		err := fmt.Errorf("status must be NOT_STARTED while creating todo")
		log.Println("error processing request payload", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := data.SaveTodo(todo)
	w.Write([]byte(id))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	if url == "/api/v1/todo/" || strings.HasPrefix(url, "/api/v1/todo?") {
		listHandler(w, r)
		return
	}
	getByIDHandler(w, r)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	inputs, err := validateAndExtractReqParam(r.URL.Query())
	if err != nil {
		log.Println("Could not parse request", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	todos := data.ListTodo(inputs)
	if inputs["sort"] != "" {
		sort.Slice(todos, func(i, j int) bool {
			if inputs["sort"] == "title" {
				c := strings.Compare(todos[i].Title, todos[j].Title)
				return c < 0
			} else if inputs["sort"] == "status" {
				return todos[i].Status < todos[j].Status
			} else if inputs["sort"] == "created_at" {
				return todos[i].CreatedAt.Before(todos[j].CreatedAt)
			} else {
				return todos[i].UpdatedAt.Before(todos[j].UpdatedAt)
			}
		})
	}

	writeJSON(w, todos)
}

func getByIDHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	id, err := validateUrlAndExtractParam(url)
	if err != nil {
		log.Println("Could not parse request", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	todo, err := data.GetTodo(id)
	if err != nil {
		log.Println("Could not parse request", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, todo)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("In updateHandler  ")
	id, err := validateUrlAndExtractParam(r.URL.Path)
	if err != nil {
		log.Println("Could not parse request url", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var todo models.Todo
	err = json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		log.Println("Could not parse request payload", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	todo.Id = id
	err = data.UpdateTodo(todo)

	if err != nil {
		log.Println("Internal error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func partialUpdateHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("In partialUpdateHandler  ")
	id, err := validateUrlAndExtractParam(r.URL.Path)
	if err != nil {
		log.Println("Could not parse request url", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var todo models.Todo
	err = json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		log.Println("Could not parse request payload", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oldTodo, err := data.GetTodo(id)
	if err != nil {
		log.Println("id not found", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	oldTodo.Status = todo.Status
	if todo.Description != "" {
		oldTodo.Description = todo.Description
	}
	fmt.Println(todo.Title)
	if todo.Title != "" {
		oldTodo.Title = todo.Title
	}
	err = data.UpdateTodo(oldTodo)
	if err != nil {
		log.Println("Internal error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
