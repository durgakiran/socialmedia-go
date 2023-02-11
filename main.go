package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/durgakiran/socialmedia/internal/database"
)

type errorBody struct {
	Error string `json:"error"`
}

type apiConfig struct {
	dbClient database.Client
}

type parameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
}

type postParameters struct {
	UserEmail string `json:"userEmail"`
	Text      string `json:"text"`
}

var client = database.NewClient("db.json")

var apiCfg = apiConfig{dbClient: client}

func testHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, 200, database.User{
		Email: "test@example.com",
	})
}

func respondWithError(w http.ResponseWriter, code int, err error) {
	if err == nil {
		fmt.Println("don't call respondWithError with a nil err!")
		return
	}
	fmt.Println(err)
	respondWithJSON(w, code, errorBody{
		Error: err.Error(),
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	if payload != nil {
		response, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("error marshalling", err)
			w.WriteHeader(500)
			response, _ := json.Marshal(errorBody{
				Error: "error marshalling",
			})
			w.Write(response)
			return
		}
		w.WriteHeader(code)
		w.Write(response)
	}

}

func testErrHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, errors.New("server error"))
}

func (apiCfg apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	_, err = apiCfg.dbClient.CreateUser(params.Email, params.Password, params.Name, params.Age)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (apiCfg apiConfig) handlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	email := strings.TrimPrefix(url, "/users/")
	err := apiCfg.dbClient.DeleteUser(email)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}

func (apiCfg apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	email := strings.TrimPrefix(url, "/users/")
	user, err := apiCfg.dbClient.Getuser(email)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	respondWithJSON(w, http.StatusOK, user)

}

func (apiCfg apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	email := strings.TrimPrefix(url, "/users/")

	type parameters struct {
		Password string `json:"password"`
		Name     string `json:"name"`
		Age      int    `json:"age"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	user, err := apiCfg.dbClient.UpdateUser(email, params.Password, params.Name, params.Age)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (apiCfg apiConfig) endpointUsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// ca get handler
		apiCfg.handlerGetUser(w, r)
	case http.MethodPost:
		// call post handler
		apiCfg.handlerCreateUser(w, r)
	case http.MethodPut:
		// call PUT handler
		apiCfg.handlerUpdateUser(w, r)
	case http.MethodDelete:
		// call DELETE handler
		apiCfg.handlerDeleteUser(w, r)
	default:
		respondWithError(w, 404, errors.New("method not supported"))
	}
}

func (apiCfg apiConfig) handlerCreatePost(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := postParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	_, err = apiCfg.dbClient.CreatePost(params.UserEmail, params.Text)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (apiCfg apiConfig) handlerDeletePost(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	uuid := strings.TrimPrefix(url, "/posts/")
	err := apiCfg.dbClient.DeletePost(uuid)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}

func (apiCfg apiConfig) handlerGetPosts(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	email := strings.TrimPrefix(url, "/posts/")
	posts, err := apiCfg.dbClient.GetPosts(email)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	respondWithJSON(w, http.StatusOK, posts)

}

func (apiCfg apiConfig) endpointPostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// ca get handler
		apiCfg.handlerGetPosts(w, r)
	case http.MethodPost:
		// call post handler
		apiCfg.handlerCreatePost(w, r)
	case http.MethodDelete:
		// call DELETE handler
		apiCfg.handlerDeletePost(w, r)

	default:
		respondWithError(w, 404, errors.New("method not supported"))
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", testHandler)
	mux.HandleFunc("/err", testErrHandler)
	mux.HandleFunc("/users", apiCfg.endpointUsersHandler)
	mux.HandleFunc("/users/", apiCfg.endpointUsersHandler)
	mux.HandleFunc("/posts", apiCfg.endpointPostsHandler)
	mux.HandleFunc("/posts/", apiCfg.endpointPostsHandler)

	const addr = "localhost:8081"
	srv := http.Server{
		Handler:      mux,
		Addr:         addr,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	err := apiCfg.dbClient.EnsureDb()

	if err != nil {
		fmt.Printf("Failed to init database")
	}

	err = srv.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
