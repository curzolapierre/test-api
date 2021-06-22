package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/curzolapierre/hook-manager/models"
	"github.com/gorilla/mux"
	"gopkg.in/errgo.v1"
)

// A global variable that is incremented everytime a excuse is added.
// Used for providing a unique ID to each excuse
var count int

var codexcuses []models.Codexcuse

// response used to answer
type response struct {
	Message string `json:"message"`
}

type excuseResp struct {
	Excuses *[]models.Codexcuse `json:"excuses"`
	Meta    models.Meta         `json:"meta"`
}

// GetExcuses return a page of excuses
func (c RequestContext) GetExcuses(w http.ResponseWriter, r *http.Request) {
	c.Log.WithField("function", "GetExcuses").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	if r.URL.Query().Get("random") != "" {
		c.getRandomExcuse(w, r)
		return
	}
	if r.URL.Query().Get("user") != "" {
		c.getUserExcuses(w, r)
		return
	}

	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		w.WriteHeader(400)
		resp := response{
			Message: "Page must be an integer greater than 0.",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	excuses := []models.Codexcuse{}
	meta, err := c.RedisStore.GetAll(vars["source"], int(page), &excuses)
	if err != nil {
		c.Log.Error(errgo.Notef(err, "fail to get excuse"))
		resp := response{
			Message: "Internal error",
		}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(excuseResp{
		Excuses: &excuses,
		Meta:    meta,
	})
}

// GetExcuse gives an excuse with some ID
func (c RequestContext) GetExcuse(w http.ResponseWriter, r *http.Request) {
	c.Log.WithField("function", "GetExcuse").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	excuse, err := c.RedisStore.Get(vars["source"], vars["id"])
	if err != nil {
		c.Log.Error(errgo.Notef(err, "fail to get excuse"))
		resp := response{
			Message: "Internal error",
		}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(resp)
		return
	}
	if excuse == nil {
		resp := response{
			Message: "ID not found",
		}
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(resp)
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(excuse)
}

// getUserExcuses gives an excuse with some ID
func (c RequestContext) getUserExcuses(w http.ResponseWriter, r *http.Request) {
	c.Log.WithField("function", "getUserExcuses").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	excuses, err := c.RedisStore.GetByUser(vars["source"], r.URL.Query().Get("user"))
	if err != nil {
		c.Log.Error(errgo.Notef(err, "fail to get excuses by user"))
		resp := response{
			Message: "Internal error",
		}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(resp)
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(excuseResp{
		Excuses: excuses,
	})
}

// getRandomExcuse gives an excuse with some ID
func (c RequestContext) getRandomExcuse(w http.ResponseWriter, r *http.Request) {
	c.Log.WithField("function", "getRandomExcuse").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	excuse, err := c.RedisStore.GetRandom(vars["source"])
	if err != nil {
		c.Log.Error(errgo.Notef(err, "fail to get random excuse"))
		resp := response{
			Message: "Internal error",
		}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(resp)
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(excuse)
}

// AddExcuse adds a new Excuse
func (c RequestContext) AddExcuse(w http.ResponseWriter, r *http.Request) {
	c.Log.WithField("function", "AddExcuse").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	var excuse models.Codexcuse
	_ = json.NewDecoder(r.Body).Decode(&excuse)

	var errors []string
	if excuse.Author == nil || excuse.Author.UserName == "" {
		errorStr := "missing Author field"
		c.Log.Debugln("fail to save excuse", errorStr)
		errors = append(errors, errorStr)
	}
	if excuse.Reporter == nil || excuse.Reporter.UserName == "" || excuse.Reporter.ID == "" {
		errorStr := "missing reporter field"
		c.Log.Debugln("fail to save excuse", errorStr)
		errors = append(errors, errorStr)
	}
	if excuse.Content == "" {
		errorStr := "missing content field"
		c.Log.Debugln("fail to save excuse", errorStr)
		errors = append(errors, errorStr)
	}
	if excuse.Title == "" {
		errorStr := "missing title field"
		c.Log.Debugln("fail to save excuse", errorStr)
		errors = append(errors, errorStr)
	}
	if errors != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		errArray := make([]string, 0, len(errors))
		for _, attrErrs := range errors {
			errArray = append(errArray, fmt.Sprintf("\t→ %s", attrErrs))
		}
		resp := response{
			Message: fmt.Sprintf("invalid arguments:\n%s", strings.Join(errArray, "\n")),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	err := c.RedisStore.Add(vars["source"], excuse)
	if err != nil {
		c.Log.Error(errgo.Notef(err, "fail to save excuse"))
		resp := response{
			Message: "Internal error",
		}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(resp)
		return
	}
	w.WriteHeader(200)
	resp := response{
		Message: "ok",
	}
	json.NewEncoder(w).Encode(resp)
}

// DeleteExcuse deletes the excuse with some ID
func (c RequestContext) DeleteExcuse(w http.ResponseWriter, r *http.Request) {
	c.Log.WithField("function", "DeleteExcuse").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	err := c.RedisStore.Delete(vars["source"], vars["id"])
	if err != nil {
		c.Log.Error(errgo.Notef(err, "fail to delete excuse: "+vars["id"]))
		resp := response{
			Message: "Internal error",
		}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(resp)
		return
	}
	w.WriteHeader(200)
	resp := response{
		Message: "ok",
	}
	json.NewEncoder(w).Encode(resp)
}
