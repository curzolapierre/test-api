package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/curzolapierre/hook-manager/controllers"
	"github.com/curzolapierre/hook-manager/environment"
	redisCtr "github.com/curzolapierre/hook-manager/redis"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"gopkg.in/errgo.v1"
)

func logLevel() logrus.Level {
	switch os.Getenv("LOGGER_LEVEL") {
	case "panic":
		return logrus.PanicLevel
	case "fatal":
		return logrus.FatalLevel
	case "warn":
		return logrus.WarnLevel
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}

func initLogger() logrus.FieldLogger {
	logger := logrus.New()
	logger.SetLevel(logLevel())
	logger.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	}

	var fieldLogger logrus.FieldLogger = logger

	return fieldLogger
}

func main() {
	log := initLogger()
	environment.Lookup()

	initRedis, err := redisCtr.Client()
	if err != nil {
		log.Error(errgo.Notef(err, "fail to init redis client"))
		return
	}
	ctr := &controllers.RequestContext{}
	ctr.Log = log
	ctr.InitStore(initRedis)

	// Define routers
	if r := routers(log, ctr); r != nil {
		log.Fatal("Server exited:", http.ListenAndServe(fmt.Sprintf("%s:%s", environment.ENV["HTTP_HOST"], environment.ENV["HTTP_PORT"]), r))
	}

	// Init general handler: containing health route

	// Init Github handler: containing all use listening route of github

	// Inint Docker handler: containing all use listening route of docker hub for each application

}

func endAPICall(w http.ResponseWriter, httpStatus int, anyStruct interface{}) {

	result, err := json.MarshalIndent(anyStruct, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(result)
}

func BasicAuth(w http.ResponseWriter, r *http.Request, username, password, realm string) bool {
	user, pass, ok := r.BasicAuth()

	if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
		return false
	}

	return true
}

type heath struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
}

func routers(log logrus.FieldLogger, ctr *controllers.RequestContext) *mux.Router {
	username := environment.ENV["BASIC_AUTH_API_USER"]
	password := environment.ENV["BASIC_AUTH_API_PASS"]

	v1Path := "/api"
	healthPath := "/health"

	topRouter := mux.NewRouter().StrictSlash(true)
	healthRouter := mux.NewRouter().PathPrefix(healthPath).Subrouter().StrictSlash(true)
	v1Router := mux.NewRouter().PathPrefix(v1Path).Subrouter().StrictSlash(true)

	healthRouter.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Health check called")
		endAPICall(w, 200, heath{
			Service:     "api",
			Environment: environment.ENV["GO_ENV"],
			Status:      "healthy",
		})
	})

	v1Router.HandleFunc("/codexcuses/{source}", ctr.GetExcuses).Methods("GET")
	v1Router.HandleFunc("/codexcuses/{source}/{id}", ctr.GetExcuse).Methods("GET")
	v1Router.HandleFunc("/codexcuses/{source}", ctr.AddExcuse).Methods("POST")
	v1Router.HandleFunc("/codexcuses/{source}/{id}", ctr.DeleteExcuse).Methods("DELETE")

	topRouter.PathPrefix(healthPath).Handler(negroni.New(
		/* Health-check routes are unprotected */
		negroni.Wrap(healthRouter),
	))

	topRouter.PathPrefix(v1Path).Handler(negroni.New(
		negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			if BasicAuth(w, r, username, password, "Provide user name and password") {
				/* Call the next handler iff Basic-Auth succeeded */
				next(w, r)
			}
		}),
		negroni.Wrap(v1Router),
	))

	return topRouter
}
