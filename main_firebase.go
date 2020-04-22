package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type ProcessExcecution struct {
	StartedAt time.Time
	EndendAt  time.Time
}

type UserStruct struct {
	FirstName   string                         `bson:"first_name" json:"first_name"`
	LastName    string                         `bson:"last_name" json:"last_name"`
	Process     map[string][]ProcessExcecution `bson:"process" json:"process"`
	LastProcess []string                       `bson:"last_process" json:"last_process"`
	LastRequest time.Time
}

type DataStruct struct {
	ID      string     `json:"id"`
	Process [][]string `json:"process"`
}

type Database struct {
	Collection *firestore.CollectionRef
}

type Response struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (database *Database) handler(w http.ResponseWriter, r *http.Request) {
	var data DataStruct
	var user UserStruct

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid_http_method")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		panic(err)
	}

	doc, err := database.Collection.Doc(data.ID).Get(context.Background())
	if err != nil {
		panic(err)
	} else {
		err = doc.DataTo(&user)
		if err != nil {
			panic(err)
		}
		log.Println(user.FirstName)
		if user.Process == nil {
			user.Process = make(map[string][]ProcessExcecution)
		}
		timing := time.Now()

		for i, dataToProcess := range data.Process {
			now := timing.Add(time.Duration(-(len(data.Process) - 1 - i)) * time.Minute)
			for _, dataProcess := range dataToProcess {

				if user.Process[dataProcess] != nil && contains(user.LastProcess, dataProcess) && now.Unix()-user.LastRequest.Unix() < 120 {
					log.Println("process exist")
					user.Process[dataProcess][len(user.Process[dataProcess])-1].EndendAt = now
				} else {
					var newProcess ProcessExcecution
					newProcess.StartedAt = now
					newProcess.EndendAt = now
					user.Process[dataProcess] = append(user.Process[dataProcess], newProcess)
				}
			}
			user.LastProcess = dataToProcess
			user.LastRequest = now
		}

		_, err = database.Collection.Doc(data.ID).Set(context.Background(), user)
		response, _ := json.Marshal(Response{Status: "OK"})
		if err != nil {
			response, _ = json.Marshal(Response{Status: "Error"})
			panic(err)
		}
		fmt.Fprintf(w, string(response))
	}
}

func (database *Database) userHandler(w http.ResponseWriter, r *http.Request) {
	var user UserStruct

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid_http_method")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		panic(err)
	}

	if user.FirstName != "" && user.LastName != "" {
		ref, _, err := database.Collection.Add(context.Background(), user)
		if err != nil {
			panic(err)
		}
		response, err := json.Marshal(Response{ID: ref.ID, Status: "OK"})
		fmt.Fprintf(w, string(response))
	}
}

func main() {
	var database Database
	opt := option.WithCredentialsFile("./firebase_key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatal(err)
	}

	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	database.Collection = client.Collection(os.Getenv("ENTERPRISE"))
	log.Println("Setting database for : " + os.Getenv("ENTERPRISE"))

	http.HandleFunc("/", database.handler)
	http.HandleFunc("/user", database.userHandler)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	} else {
		log.Println("server up and running on port :" + os.Getenv("ENTERPRISE"))
	}
}
