// Copyright 2013 Sahid Orentino Ferdjaoui. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package tiny

import (
	"fmt"
	"log"
	"time"
	"regexp"
	"io/ioutil"
	"net/http"
	"strconv"

	"appengine"
	"appengine/urlfetch"
	"appengine/datastore"
)


const USE_MEMCACHE = false
const ENTITY_NAME = "Url"
const MSG_INTERNAL_ERROR = "Internal Error, please wait a moment..."
const MSG_ERROR_400 = "The request could not be understood " +
	"by the server due to malformed syntax"

type Url struct {
        Path string
	Created time.Time
}

/** Checks if the url is valid. */
func checkUrl(url string, r *http.Request) (err error) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	_, err = client.Get(url)
	
	return
}

/** Returns an datastore key from a request path. */
func idFromPath(path string) (id int64, err error) {
	rex := regexp.MustCompile("[a-zA-Z0-9]+")
	sid := rex.FindString(path)
	
	id, err = strconv.ParseInt(sid, 36, 0)
	return
}

/** Finds an Url by a datastore id. */
func find(id int64, r *http.Request) (ref Url, err error) {
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, ENTITY_NAME, "", id, nil)
	
	err = datastore.Get(ctx, key, &ref);
	return
}

/** Creates a new Url. */
func create(url string, r *http.Request) (sid string, err error) {
	ctx := appengine.NewContext(r)
	ref := Url{
		Path: url,
		Created: time.Now(),
	}
	key, err := datastore.Put(
		ctx, datastore.NewIncompleteKey(ctx, ENTITY_NAME, nil), &ref)
	sid = strconv.FormatInt(key.IntID(), 36)
	return
}

func usage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Hello :')")
}


func redirect(w http.ResponseWriter, r *http.Request, u string) {
	http.Redirect(w, r, u, http.StatusMovedPermanently)
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		res, err := ioutil.ReadAll(r.Body)
		if (err != nil) {
			http.Error(w, MSG_ERROR_400, http.StatusBadRequest)
		}
		url := string(res[:])
		err = checkUrl(url, r)
		if err != nil {
			http.Error(
				w, 
				"Not a valid url", 
				http.StatusInternalServerError)
			return
		}
		
		log.Printf("Generating a new tiny url for: %s.", url)
		sid, err := create(url, r)
		
		// TODO(sahid): checks the content type.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "http://%s/%s\n", r.Host, sid)
		return
	case "GET":
		id, err := idFromPath(r.URL.Path)
		if err != nil {
			usage(w, r)
			return
		}
		ref, err := find(id, r)
		if err != nil {
			log.Printf("Can't find %d on datastore: %s", id, err)
			http.NotFound(w, r)
			return
		}
		redirect(w, r, ref.Path)
	default:
		http.Error(w, "What are you doing my friend?", 500)
	}
}

func init() {
	http.HandleFunc("/", handle)
}
