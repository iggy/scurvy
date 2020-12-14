package main

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"io/ioutil"
	"log"

	"net/http"

	"github.com/iggy/scurvy/pkg/msgs"
)

// handle notifications from CouchPotato
// This doesn't really do much right now, but will eventually allow messages specific to downloaded
// movies
func couchpotatoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Println("failed to read request body")
	}

	err = r.Body.Close()
	if err != nil {
		log.Println("failed to close request body")
	}

	log.Println(body)
}

// handle notifications from SABNZBD
func sabnzbdHandler(w http.ResponseWriter, r *http.Request) {
	// Some examples:
	// body: {"message": "Too little diskspace forcing PAUSE", "version": "1.0", "type": "info", "title": "SABnzbd: Warning"}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Printf("failed to read request body: %#v\n", err)
	}

	err = r.Body.Close()
	if err != nil {
		log.Println("failed to close request body: %#v\n", err)
	}

	log.Printf("body: %s\n", body)

	jreq := SABJSONRequest{}
	if jerr := json.Unmarshal(body, &jreq); jerr != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		err = json.NewEncoder(w).Encode(jerr)
		if err != nil {
			log.Printf("failed to encode json error: %#v\n", err)
		}
	}
	log.Printf("SAB: message: %s\n\ttitle: %s\n\ttype: %s\n\tversion: %s\n",
		jreq.Message, jreq.Title, jreq.Type, jreq.Version)

	if jreq.Title == "SABnzbd: Job finished" {
		nd := msgs.NewDownload{Name: jreq.Message, Path: "/scurvy"} // TODO find actual path
		msgs.SendNatsMsg("scurvy.notify.newdownload", nd)
	}
	if jreq.Title == "SABnzbd: Job failed" {
		fd := msgs.FailedDownload{Name: jreq.Message, Path: "/scurvy"} // TODO find actual path}
		msgs.SendNatsMsg("scurvy.notify.faileddownload", fd)
	}
	if jreq.Title == "SABnzbd: Warning" {
		df := msgs.DiskFull{Message: jreq.Title}
		msgs.SendNatsMsg("scurvy.notify.diskfull", df)
	}

}

// handle notifications from SickBeard
// This is slightly complicated by the fact that sickbeard doesn't have a generic notification
// function, so we have to pretend to be XBMC's JSONRPC interface
func sickbeardHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Printf("failed to read request body: %#v\n", err)
	}

	err = r.Body.Close()
	if err != nil {
		log.Println("failed to close request body: %#v\n", err)
	}

	log.Printf("SICK: %q\n", bytes.NewBuffer(body).String())

	// parse the json payload and figure out what they want to know
	var jreq = JSONRPCRequest{}
	if jerr := json.Unmarshal(body, &jreq); jerr != nil {
		log.Printf("Failed to unmarshall json body: \n%v\n%v", jerr, body)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		err := json.NewEncoder(w).Encode(jerr)
		if err != nil {
			log.Printf("failed to encode json error: %#v\n", err)
		}
	}
	log.Printf("SICK: method = \"%s\" (%T)\n", jreq.Method, jreq.JSONRPC)

	// answer them
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	switch jreq.Method {
	case "JSONRPC.Version":
		// just something we have to emulate to get sickbeard to talk to us
		log.Println("SICK: JSONRPC.Version called")
		jret := &JSONRPCVersionResponse{ID: jreq.ID, JSONRPC: jreq.JSONRPC}
		jret.Result.Version.Major = 8
		jret.Result.Version.Minor = 0
		jret.Result.Version.Patch = 0

		jstr, err := json.Marshal(jret)
		if err != nil {
			log.Printf("failed to marshal json: %#v\n", err)
		}
		c, err := w.Write(jstr)
		if err != nil {
			log.Println("Failed to write json response", jstr, err, c)
		}
	case "GUI.ShowNotification":
		// This case is actually where something has actually downloaded
		log.Println("SICK: GUI.ShowNotification JSONRPC request method")

		// re-parse the JSON to get something more specific
		var jreqp = JSONRPCRequestParamsGUISN{}
		if jgsnerr := json.Unmarshal(body, &jreqp); jgsnerr != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			err := json.NewEncoder(w).Encode(jgsnerr)
			if err != nil {
				log.Printf("GUI.ShowNotification: failed to encode json error: %#v\n", err)
			}
		}

		// reply to the request
		var jret = JSONRPCGenericResponse{}
		jret.ID = jreq.ID
		jret.JSONRPC = jreq.JSONRPC
		jret.Result = "OK"
		jstr, err := json.Marshal(jret)
		if err != nil {
			log.Printf("failed to marshal json: %#v\n", err)
		}
		c, err := w.Write(jstr)
		if err != nil {
			log.Println("Failed to write json response", jstr, err, c)
		}

		// use the actual data we got
		log.Println(jreqp)
		nd := msgs.NewDownload{Name: jreqp.Params.Message, Path: "/scurvy"} // TODO find actual path
		msgs.SendNatsMsg("scurvy.notify.newdownload", nd)
	default:
		log.Println("SICK: Unknown JSONRPC request method")
		var jret = JSONRPCGenericResponse{}
		jret.ID = jreq.ID
		jret.JSONRPC = jreq.JSONRPC
		jret.Result = "Error"
		jstr, err := json.Marshal(jret)
		if err != nil {
			log.Printf("failed to marshal json: %#v\n", err)
		}
		c, err := w.Write(jstr)
		if err != nil {
			log.Println("Failed to write json response", jstr, err, c)
		}
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Printf("failed to read body: %#v\n", err)
	}
	err = r.Body.Close()
	if err != nil {
		log.Printf("failed to close body: %#v\n", err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	c, err := io.WriteString(w, "Running")
	if err != nil {
		log.Println("Failed to write defaultHandler response", err, c)
	}
	log.Printf("DEF: %q (%q)\n", bytes.NewBuffer(body).String(), html.EscapeString(r.URL.Path))
}
