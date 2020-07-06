package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Movie struct {
	ID          primitive.ObjectID `bson:"_id, omitempty" json:"id, omitempty"`
	Name        string             `bson:"name, omitempty" json:"name, omitempty"`
	Description string             `bson:"description, omitempty" json:"description, omitempty"`
	CoverImage  string             `bson:"cover_image, omitempty" json:"cover_image, omitempty"`
}

var coll *mongo.Collection

func homeHandler(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "Movies Home Page")
	if err != nil {
		log.Println(err)
	}
}

func returnAllMovies(w http.ResponseWriter, _ *http.Request) {

	fmt.Println("returnAllMovies Endpoint")

	var movies []Movie

	cursor, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		if err != nil {
			log.Println(err)
		}

		return
	}
	for cursor.Next(context.Background()) {
		var movie Movie
		err = cursor.Decode(&movie)

		if err != nil {
			log.Println(err)
		}
		movies = append(movies, movie)
	}

	err = json.NewEncoder(w).Encode(movies)

	if err != nil {
		log.Println(err)
	}
}

func returnSingleMovie(w http.ResponseWriter, r *http.Request) {

	fmt.Println("returnSingleMovie Endpoint")

	w.Header().Set("content-type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		log.Println(err)
	}
	var movie Movie
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&movie)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{ "message": "` + err.Error() + `" }`))

		if err != nil {
			log.Println(err)
		}
		return
	}
	err = json.NewEncoder(w).Encode(movie)
	if err != nil {
		log.Println(err)
	}

}

func addNewMovie(w http.ResponseWriter, r *http.Request) {

	fmt.Println("addNewMovie Endpoint")

	w.Header().Set("content-type", "application/json")

	var movie Movie
	// json.NewDecoder(r.Body).Decode(movie)
	reqBody, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(reqBody, &movie)

	if err != nil {
		log.Fatal(err)
	}
	movie.ID = primitive.NewObjectID()
	result, err := coll.InsertOne(context.Background(), movie)

	if err != nil {
		log.Fatal(err)
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {

	fmt.Println("deleteMovie Endpoint")

	vars := mux.Vars(r)
	key, _ := primitive.ObjectIDFromHex(vars["id"])

	result, err := coll.DeleteOne(context.Background(), bson.M{"_id": key})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{ "message" : "` + err.Error() + `" }`))
		if err != nil {
			log.Println(err)
		}
		return
	}
	_, err = fmt.Fprintln(w, "Movie Deleted")
	if err != nil {
		log.Println(err)
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Println(err)
	}
}

func updateMovie(w http.ResponseWriter, r *http.Request) {

	fmt.Println("updateMovie Endpoint")

	key, _ := primitive.ObjectIDFromHex(mux.Vars(r)["id"])

	var updateMovie Movie
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(reqBody, &updateMovie)
	if err != nil {
		log.Println(err)
	}

	update := bson.D{{
		"$set",
		bson.D{
			{"name", updateMovie.Name},
			{"description", updateMovie.Description},
			{"cover_image", updateMovie.CoverImage}}}}

	result, err := coll.UpdateOne(context.Background(), bson.M{"_id": key}, update)
	if err != nil {
		log.Println(err)
	}
	err = json.NewEncoder(w).Encode(result)
}

func searchByName(w http.ResponseWriter, r *http.Request) {

	fmt.Println("return Movie search results by name endpoint")

	key := mux.Vars(r)["name"]

	var movies []Movie

	cursor, err := coll.Find(context.Background(), bson.M{"name": key})
	if err != nil {
		log.Println(err)
	}
	for cursor.Next(context.Background()) {
		var movie Movie
		err = cursor.Decode(&movie)
		if err != nil {
			log.Println(err)
		}
		movies = append(movies, movie)
	}

	err = json.NewEncoder(w).Encode(movies)

	if err != nil {
		log.Println(err)
	}
}

func handleRequests() {
	router := mux.NewRouter()

	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/movies", returnAllMovies).Methods("GET")
	router.HandleFunc("/movie/{id}", returnSingleMovie).Methods("GET")
	router.HandleFunc("/movie", addNewMovie).Methods("POST")
	router.HandleFunc("/movie/{id}", deleteMovie).Methods("DELETE")
	router.HandleFunc("/movie/{id}", updateMovie).Methods("PATCH")
	router.HandleFunc("/search/movie-name/{name}", searchByName).Methods("GET")
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
