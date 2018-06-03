package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return origin == "http://localhost:3001"
		},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})
	router, err := rest.MakeRouter(
		rest.Get("/graph", GetGraphMis),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

type Graph struct {
	Nodes []struct {
		ID    int    `json:"id"`
		Label string `json:"label"`
		Group int    `json:"group"`
	} `json:"nodes"`
	Edges []struct {
		From int `json:"from"`
		To   int `json:"to"`
	} `json:"edges"`
}

var lock = sync.RWMutex{}

func GetGraphMis(w rest.ResponseWriter, r *rest.Request) {
	lock.RLock()
	bytes, err := ioutil.ReadFile("./graph_mis.json")
	if err != nil {
		log.Fatal(err)
	}

	var graph Graph
	if err := json.Unmarshal(bytes, &graph); err != nil {
		log.Fatal(err)
	}
	lock.RUnlock()

	w.WriteJson(graph)
}
