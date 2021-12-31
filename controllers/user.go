package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ezrod12/go-web-server/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userController struct {
	userIdPattern *regexp.Regexp
	context       context.Context
	collection    *mongo.Collection
}

func (uc userController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/users" {
		switch r.Method {
		case http.MethodGet:
			uc.getAll(w, r)
		case http.MethodPost:
			uc.post(w, r)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	} else {
		matches := uc.userIdPattern.FindStringSubmatch(r.URL.Path)
		if len(matches) == 0 {
			w.WriteHeader(http.StatusNotFound)
		}
		id := matches[1]

		fmt.Println(id)

		switch r.Method {
		case http.MethodGet:
			uc.get(id, w)
		case http.MethodPut:
			uc.put(id, w, r)
		case http.MethodDelete:
			uc.delete(id, w)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}
}

func newUserController() *userController {
	/*
	   Connect to my cluster
	*/
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("library").Collection("users")

	return &userController{
		userIdPattern: regexp.MustCompile(`/users/([A-Za-z0-9\-]+)/?`),
		collection:    collection,
	}
}

func (uc *userController) getAll(w http.ResponseWriter, r *http.Request) {
	encodeResponseAsJson(models.GetUsers(uc.collection, uc.context), w)
}

func (uc *userController) get(id string, w http.ResponseWriter) {
	u, err := models.GetUserById(id, uc.collection, uc.context)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	encodeResponseAsJson(u, w)
}

func (uc *userController) post(w http.ResponseWriter, r *http.Request) {
	u, err := uc.parseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = validateUserEntity(u, true)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	u, err = models.AddUser(u, uc.collection, uc.context)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	encodeResponseAsJson(u, w)
}

func (uc *userController) put(id string, w http.ResponseWriter, r *http.Request) {
	u, err := uc.parseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not parse User object"))
	}
	if id != u.Id {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID of submitted user must math ID in URL"))
	}

	err = validateUserEntity(u, false)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	u, err = models.UpdateUser(u)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	encodeResponseAsJson(u, w)
}

func (uc *userController) delete(id string, w http.ResponseWriter) {
	err := models.RemoveUser(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	w.WriteHeader(http.StatusOK)
}

func (uc *userController) parseRequest(r *http.Request) (models.User, error) {
	dec := json.NewDecoder(r.Body)
	var u models.User
	err := dec.Decode(&u)

	if err != nil {
		return models.User{}, err
	}

	return u, nil
}

func validateUserEntity(user models.User, createMode bool) error {
	if strings.Trim(user.FirstName, " ") == "" {
		return errors.New("property firstname must contain a valid string value")
	}

	if strings.Trim(user.LastName, " ") == "" {
		return errors.New("property lastname must contain a valid string value")
	}

	return nil
}
