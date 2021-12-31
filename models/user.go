package models

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	Id        string `json:"id" bson:"_id,omitempty"`
	FirstName string
	LastName  string
}

var (
	users  []*User
	nextId = 1
)

func GetUsers(collection *mongo.Collection, ctx context.Context) []*User {
	cur, currErr := collection.Find(ctx, bson.D{})

	if currErr != nil {
		panic(currErr)
	}
	defer cur.Close(ctx)

	var usersCollection []*User
	if err := cur.All(ctx, &usersCollection); err != nil {
		panic(err)
	}
	fmt.Println(usersCollection)
	return usersCollection
}

func AddUser(u User, collection *mongo.Collection, ctx context.Context) (User, error) {
	nextId++

	res, insertErr := collection.InsertOne(ctx, u)
	if insertErr != nil {
		log.Fatal(insertErr)
	}
	fmt.Println(res)

	return u, nil
}

func GetUserById(id string, collection *mongo.Collection, ctx context.Context) (User, error) {
	user := User{}
	fmt.Println(id)
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Invalid id")
	}

	err = collection.FindOne(ctx, bson.D{{"_id", objectId}}).Decode(&user)

	if err != nil {
		panic(err)
	}

	return user, nil
}

func UpdateUser(u User) (User, error) {
	for i, candidate := range users {
		if candidate.Id == u.Id {
			users[i] = &u
			return u, nil
		}
	}

	return User{}, fmt.Errorf("User with ID '%v' not found", u.Id)
}

func RemoveUser(id string) error {
	for i, u := range users {
		if u.Id == id {
			users = append(users[:i], users[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("User with ID '%v' not found", id)
}
