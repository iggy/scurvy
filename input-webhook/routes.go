package main

import "net/http"

// Route for gorilla mux
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes for handling JSON
type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"POST",
		"/",
		defaultHandler,
	},
	Route{
		"couchpotato",
		"POST",
		"/couchpotato",
		couchpotatoHandler,
	},
	Route{
		"sickbeard",
		"POST",
		"/jsonrpc",
		sickbeardHandler,
	},
	Route{
		"sabnzbd",
		"POST",
		"/sabnzbd",
		sabnzbdHandler,
	},
}
