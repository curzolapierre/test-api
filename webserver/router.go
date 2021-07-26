package webserver

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/Scalingo/go-utils/logger"
	"github.com/curzolapierre/hook-manager/config"
	"github.com/curzolapierre/hook-manager/controllers"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func NewRouter(ctx context.Context, config config.Config) *mux.Router {
	log := logger.Get(ctx)
	username := config.BasicAuthApiUser
	password := config.BasicAuthApiPass

	v1Path := "/api"
	healthPath := "/health"

	topRouter := mux.NewRouter().StrictSlash(true)
	healthRouter := mux.NewRouter().PathPrefix(healthPath).Subrouter().StrictSlash(true)
	v1Router := mux.NewRouter().PathPrefix(v1Path).Subrouter().StrictSlash(true)

	healthRouter.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Health check called")
		log.Info("IP of user using x-forwarded-for:", w.Header().Get("x-forwarded-for"))
		log.Info("IP of user using x-real-ip:", w.Header().Get("x-real-ip"))
		endAPICall(w, 200, heath{
			Service:     "api",
			Environment: config.GoEnv,
			Status:      "healthy",
		})
	})

	addRoutes(v1Router)

	topRouter.PathPrefix(healthPath).Handler(negroni.New(
		/* Health-check routes are unprotected */
		negroni.Wrap(healthRouter),
	))

	topRouter.PathPrefix(v1Path).Handler(negroni.New(
		negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			if BasicAuth(w, r, username, password, "Provide user name and password") {
				/* Call the next handler if Basic-Auth succeeded */
				next(w, r)
			}
		}),
		negroni.Wrap(v1Router),
	))

	return topRouter
}

func addRoutes(router *mux.Router) {
	ctrl := controllers.NewExcuseController()

	router.HandleFunc("/codexcuses/{source}", ctrl.GetExcuses).Methods("GET")
	router.HandleFunc("/codexcuses/{source}/{id}", ctrl.GetExcuse).Methods("GET")
	router.HandleFunc("/codexcuses/{source}", ctrl.AddExcuse).Methods("POST")
	router.HandleFunc("/codexcuses/{source}/{id}", ctrl.DeleteExcuse).Methods("DELETE")
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
