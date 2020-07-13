package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type shoterUrllink struct {
	ID           string `json:"_id,omitempty" bson:"_id,omitempty"`
	LongUrllink  string `json:"longurl,omitempty" bson:"longurl,omitempty"`
	ShortUrllink string `json:"shorturl,omitempty" bson:"shorturl,omitempty"`
}

func CreateUrl(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var aurl shoterUrllink
	_ = json.NewDecoder(request.Body).Decode(&aurl)
	collection := client.Database("urlshortenerproject").Collection("short")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	hd := hashids.NewData()
	h, _ := hashids.NewWithData(hd)
	now := time.Now()
	k, _ := h.Encode([]int{int(now.Unix())})
	aurl.ID = k
	aurl.ShortUrllink = "http://localhost:12345/" + k
	result, _ := collection.InsertOne(ctx, aurl)
	json.NewEncoder(response).Encode(result)
}

func GetUrl(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	//params := mux.Vars(request)
	//id := string(params["id"])
	params := request.URL.Query()
	id := string(params.Get("longurl"))
	var aurl shoterUrllink
	collection := client.Database("urlshortenerproject").Collection("short")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, shoterUrllink{LongUrllink: id}).Decode(&aurl)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(aurl)
}
func GetAllUrl(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var vurl []shoterUrllink
	collection := client.Database("urlshortenerproject").Collection("short")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var aurl shoterUrllink
		cursor.Decode(&aurl)
		vurl = append(vurl, aurl)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(vurl)
}

func RootEndpoint(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("content-type", "text/html")
	params := mux.Vars(request)
	id := string(params["id"])
	var aurl shoterUrllink
	collection := client.Database("urlshortenerproject").Collection("short")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, shoterUrllink{ID: id}).Decode(&aurl)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	http.Redirect(response, request, aurl.LongUrllink, 301)
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/generate", CreateUrl).Methods("POST")
	router.HandleFunc("/allurls", GetAllUrl).Methods("GET")
	router.HandleFunc("/getshorturl", GetUrl).Methods("GET")
	router.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	http.ListenAndServe(":12345", router)
}
