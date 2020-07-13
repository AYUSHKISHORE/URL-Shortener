package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	aero "github.com/aerospike/aerospike-client-go"
	as "github.com/aerospike/aerospike-client-go"
	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
)

type shoterUrllink struct {
	ID           string `json:"_id,omitempty"`
	LongUrllink  string `json:"longurl,omitempty"`
	ShortUrllink string `json:"shorturl,omitempty"`
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// Creating the shorturl
func CreateUrl(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	client, err := aero.NewClient("Enter your Ip", 3000)
	panicOnError(err)

	//making the unique hashids for detecting the each url uniquely
	hd := hashids.NewData()
	h, _ := hashids.NewWithData(hd)
	now := time.Now()
	k, _ := h.Encode([]int{int(now.Unix())})

	// Making the unique based on the hashids i.e k
	key, err := aero.NewKey("test", "urlsho", k)
	panicOnError(err)
	var aurl shoterUrllink

	// Decoding the value and putting it into the aurl variable
	_ = json.NewDecoder(request.Body).Decode(&aurl)

	slink := "http://localhost:12345/" + k

	// putting the values in bins
	bins := aero.BinMap{
		"ID":       k,
		"Longurl":  aurl.LongUrllink,
		"ShortUrl": slink,
	}

	// Actual Putting is done here
	_ = client.Put(nil, key, bins)

	// we don't want to show response now
	//json.NewEncoder(response).Encode(k)
}

func GetUrl(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := request.URL.Query()
	id := string(params.Get("longurl"))
	var aurl shoterUrllink

	client, err := aero.NewClient("Enter your ip", 3000)
	panicOnError(err)
	spolicy := aero.NewScanPolicy()
	spolicy.ConcurrentNodes = true
	spolicy.Priority = as.LOW
	spolicy.IncludeBinData = true

	recs, _ := client.ScanAll(spolicy, "test", "urlsho")

	for res := range recs.Results() {
		if res.Err != nil {
			// handle error here
			// if you want to exit, cancel the recordset to release the resources
		} else {
			// process record here
			k1 := res.Record.Bins["Longurl"]
			lg := fmt.Sprintf("%v", k1)
			if lg == id {
				aurl.LongUrllink = id
				k2 := res.Record.Bins["Shorturl"]
				sg := fmt.Sprintf("%v", k2)
				aurl.ShortUrllink = sg
				k3 := res.Record.Bins["ID"]
				ig := fmt.Sprintf("%v", k3)
				aurl.ID = ig
			}

		}
	}
	//aurl.ID =
	//aurl.LongUrllink =
	//aurl.ShortUrllink =
	json.NewEncoder(response).Encode(aurl)
}

// For redirecting the shorturl to the original url (i.e) Longurl
func RootEndpoint(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("content-type", "text/html")
	params := mux.Vars(request)
	id := params["id"]
	client, err := aero.NewClient("Enter your Ip", 3000)
	panicOnError(err)
	// Finding the key (here id is the key)
	key, err := aero.NewKey("test", "urlsho", id)
	panicOnError(err)
	// getting the complete record
	record, err := client.Get(nil, key, "Longurl", "ID")
	panicOnError(err)
	// Getting the interface of longurl
	lk := record.Bins["Longurl"]
	//to convert interface into string
	urlk := fmt.Sprintf("%v", lk)
	//fmt.Printf("record: %#v\n", record.Bins["Longurl"])
	fmt.Println(urlk)
	// Woah .. we are redirecting the shorturl to the longurl
	http.Redirect(response, request, urlk, 301)
}

func main() {
	fmt.Println("Starting the application...")
	router := mux.NewRouter()
	router.HandleFunc("/generate", CreateUrl).Methods("POST")
	router.HandleFunc("/getshorturl", GetUrl).Methods("GET")
	router.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	http.ListenAndServe(":12345", router)
}
