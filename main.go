package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
	ldap "gopkg.in/ldap.v2"
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
		rest.Get("/users", GetAllLdapUsers),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

type Graph struct {
	Nodes []*Node
	Links []*Link
}

type Node struct {
	id    string
	group int
}

type Link struct {
	Source string
	Target string
}

var lock = sync.RWMutex{}

// GetAllLdapUsers : get users in ldap
func GetAllLdapUsers(w rest.ResponseWriter, r *rest.Request) {
	lock.RLock()
	userEntries := LdapUsersSearch()
	//group Entries

	// convert Entries to graph model json
	//graphJSON := jsonConverter(userEntries)
	lock.RUnlock()
	w.WriteJson(&userEntries)
}

// LdapUsersSearch : return Entries of LDAP users
func LdapUsersSearch() []*ldap.Entry {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", "localhost", 389))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	err = l.Bind("cn=admin,dc=example,dc=org", "admin")
	if err != nil {
		log.Fatal(err)
	}

	searchRequest := ldap.NewSearchRequest(
		"dc=example,dc=org", // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=inetOrgPerson))",    // The filter to apply
		[]string{"dn", "cn", "displayName"}, // A list attributes to retrieve
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	return sr.Entries
}
