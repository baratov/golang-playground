package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/baratov/golang-playground/store"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var s *store.Store

func Serve(port string, restore bool) {
	if restore {
		s = store.New(
			store.WithRestoreFromFile("./store.gob"),
		)
	} else {
		s = store.New()
	}

	r := mux.NewRouter()
	r.Use(recoverMiddleware)
	r.Use(basicAuthMiddleware)
	r.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	r.HandleFunc("/api/v1/keys", GetKeysHandler).Methods("GET")
	r.HandleFunc("/api/v1/keys/{key}", GetHandler).Methods("GET")
	r.HandleFunc("/api/v1/keys/{key}", SetHandler).Methods("POST")
	r.HandleFunc("/api/v1/keys/{key}", UpdateHandler).Methods("PUT")
	r.HandleFunc("/api/v1/keys/{key}", DeleteHandler).Methods("DELETE")

	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second,
		ReadTimeout:  time.Second,
		IdleTimeout:  time.Second * 15,
		Handler:      r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	//graceful shutdown by SIGINT
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	srv.Shutdown(ctx)
	s.Stop()
}

func GetKeysHandler(w http.ResponseWriter, _ *http.Request) {
	withWriter(w).
		Data(s.Keys()).
		WriteResponse()
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	key := parseKey(r)
	val, err := s.Get(key)

	withWriter(w).
		Data(val).
		Error(err).
		WriteResponse()
}

func SetHandler(w http.ResponseWriter, r *http.Request) {
	key := parseKey(r)
	payload := parseBody(r)
	s.Set(key, payload.value, payload.ttl)

	withWriter(w).
		Data(nil).
		WriteResponse()
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	key := parseKey(r)
	payload := parseBody(r)
	err := s.Update(key, payload.value, payload.ttl)

	withWriter(w).
		Data(nil).
		Error(err).
		WriteResponse()
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	key := parseKey(r)
	s.Delete(key)

	withWriter(w).
		Data(nil).
		WriteResponse()
}

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	withWriter(w).
		Field("alive", true).
		WriteResponse()
}

func basicAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// to comply with http should also send WWW-Authenticate header
		writeAuthError := func() {
			http.Error(w, "", http.StatusUnauthorized)
		}

		headerSplit := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(headerSplit) != 2 || headerSplit[0] != "Basic" {
			writeAuthError()
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(headerSplit[1])
		if err != nil {
			writeAuthError()
			return
		}

		namePassPair := strings.SplitN(string(decoded), ":", 2)
		if len(namePassPair) != 2 || !isAuthorized(namePassPair[0], namePassPair[1]) {
			writeAuthError()
			return
		}
		h.ServeHTTP(w, r)
	})
}

func isAuthorized(username, password string) bool {
	return username == "username" && password == "password"
}

func recoverMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, (r.(error)).Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

type payload struct {
	value interface{}
	ttl   time.Duration
}

func parseKey(r *http.Request) string {
	return mux.Vars(r)["key"]
}

func parseBody(r *http.Request) payload {
	decoder := json.NewDecoder(r.Body)
	var p payload
	err := decoder.Decode(&p)
	r.Body.Close()

	if err != nil {
		panic(err) // should probably return error to handle as BadRequest
	}
	return p
}
