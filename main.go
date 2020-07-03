package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

type Movie struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	CoverImage  string             `bson:"cover_image" json:"cover_image"`
}

var coll *mongo.Collection

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Movies Home Page")
}

func handleRequests() {
	router := mux.NewRouter()

	router.HandleFunc("/", homeHandler)

	log.Fatal(http.ListenAndServe(":3000", router))
}

func main() {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	coll = client.Database("moviesdb").Collection("movies")
	defer client.Disconnect(context.Background())
	fmt.Println("Application started Successfully")

	handleRequests()

}
