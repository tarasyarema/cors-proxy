package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

type e struct {
	string `json:"error"`
}

func logReq(r *http.Request, t time.Time) {
	log.Infof("%s %s %s - %s", r.Method, r.URL.EscapedPath(), r.RemoteAddr, time.Since(t))
}

func twilioCallback(w http.ResponseWriter, r *http.Request) {
	defer logReq(r, time.Now())

	defer r.Body.Close()

	if b, err := ioutil.ReadAll(r.Body); err != nil {
		log.Infoln("Error parsing body:", err.Error())
	} else {
		log.Infoln("Body:", string(b))
	}

	log.Infoln("Query:", r.Form.Encode())

	w.WriteHeader(http.StatusNoContent)
}
func corsProxy(w http.ResponseWriter, r *http.Request) {
	defer logReq(r, time.Now())

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	reqURL := q.Get("url")

	if reqURL == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	q.Del("url")
	reqURL = fmt.Sprintf("%s?%s", reqURL, q.Encode())

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	newReq, err := http.NewRequest(r.Method, reqURL, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		b, _ := json.Marshal(e{err.Error()})
		w.Write(b)

		return
	}

	for h := range r.Header {
		for _, v := range r.Header[h] {
			newReq.Header.Add(h, v)
		}
	}

	resp, err := client.Do(newReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		b, _ := json.Marshal(e{err.Error()})
		w.Write(b)

		return
	}

	w.WriteHeader(resp.StatusCode)

	n, _ := io.Copy(w, resp.Body)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", n))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", corsProxy)
	mux.HandleFunc("/twilio/", twilioCallback)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "PATCH"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Infof("server running at %s", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler); err != nil {
		log.Fatal(err)
	}
}
