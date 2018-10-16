package main

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"io/ioutil"
	"log"

	"net/http"

	"github.com/iggy/scurvy/common"
	"github.com/iggy/scurvy/msgs"
)

// handle notifications from CouchPotato
// This doesn't really do much right now, but will eventually allow messages specific to downloaded
// movies
func couchpotatoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	common.CheckErr(err)

	cerr := r.Body.Close()
	common.CheckErr(cerr)

	log.Println(body)
}

// handle notifications from SABNZBD
func sabnzbdHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	common.CheckErr(err)
	cerr := r.Body.Close()
	common.CheckErr(cerr)

	log.Printf("body: %s\n", body)

	jreq := SABJSONRequest{}
	if jerr := json.Unmarshal(body, &jreq); jerr != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		eerr := json.NewEncoder(w).Encode(jerr)
		common.CheckErr(eerr)
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

}

// handle notifications from SickBeard
// This is slightly complicated by the fact that sickbeard doesn't have a generic notification
// function, so we have to pretend to be XBMC's JSONRPC interface
func sickbeardHandler(w http.ResponseWriter, r *http.Request) {
	body, rerr := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	common.CheckErr(rerr)
	cerr := r.Body.Close()
	common.CheckErr(cerr)

	log.Printf("SICK: %q\n", bytes.NewBuffer(body).String())

	// parse the json payload and figure out what they want to know
	var jreq = JSONRPCRequest{}
	if jerr := json.Unmarshal(body, &jreq); jerr != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		eerr := json.NewEncoder(w).Encode(jerr)
		common.CheckErr(eerr)
	}
	log.Printf("SICK: method = \"%s\" (%T)\n", jreq.Method, jreq.JSONRPC)

	// answer them
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	switch jreq.Method {
	case "JSONRPC.Version":
		// just something we have to emulate to get sickbeard to talk to us
		log.Println("SICK: JSONRPC.Version called")
		jret := &JSONRPCVersionResponse{ID: jreq.ID, JSONRPC: jreq.JSONRPC}
		jret.Result.Version.Major = 8
		jret.Result.Version.Minor = 0
		jret.Result.Version.Patch = 0

		jstr, err := json.Marshal(jret)
		common.CheckErr(err)
		w.Write(jstr)
	case "GUI.ShowNotification":
		// This case is actually where something has actually downloaded
		log.Println("SICK: GUI.ShowNotification JSONRPC request method")

		// re-parse the JSON to get something more specific
		var jreqp = JSONRPCRequestParamsGUISN{}
		if jgsnerr := json.Unmarshal(body, &jreqp); jgsnerr != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			eerr := json.NewEncoder(w).Encode(jgsnerr)
			common.CheckErr(eerr)
		}

		// reply to the request
		var jret = JSONRPCGenericResponse{}
		jret.ID = jreq.ID
		jret.JSONRPC = jreq.JSONRPC
		jret.Result = "OK"
		jstr, err := json.Marshal(jret)
		common.CheckErr(err)
		w.Write(jstr)

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
		common.CheckErr(err)
		w.Write(jstr)
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	common.CheckErr(err)
	cerr := r.Body.Close()
	common.CheckErr(cerr)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Running")
	log.Printf("DEF: %q (%q)\n", bytes.NewBuffer(body).String(), html.EscapeString(r.URL.Path))
}