package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Scalingo/go-utils/logger"
	"github.com/curzolapierre/hook-manager/models"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type ExcuseController struct {
	Codexcuse  models.Codexcuse
	RedisStore *models.RedisStoreCodexcuses
}

func NewExcuseController() ExcuseController {
	return ExcuseController{
		RedisStore: &models.RedisStoreCodexcuses{},
	}
}

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
func (c ExcuseController) GetExcuses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	log.WithField("function", "GetExcuses").Infoln("received on", r.URL.Path)
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
	meta, err := c.RedisStore.GetAll(ctx, vars["source"], int(page), &excuses)
	if err != nil {
		log.Error(errors.Wrap(err, "fail to get excuse"))
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
func (c ExcuseController) GetExcuse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	log.WithField("function", "GetExcuse").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	excuse, err := c.RedisStore.Get(ctx, vars["source"], vars["id"])
	if err != nil {
		log.Error(errors.Wrap(err, "fail to get excuse"))
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
func (c ExcuseController) getUserExcuses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	log.WithField("function", "getUserExcuses").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	excuses, err := c.RedisStore.GetByUser(ctx, vars["source"], r.URL.Query().Get("user"))
	if err != nil {
		log.Error(errors.Wrap(err, "fail to get excuses by user"))
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
func (c ExcuseController) getRandomExcuse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	log.WithField("function", "getRandomExcuse").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	excuse, err := c.RedisStore.GetRandom(ctx, vars["source"])
	if err != nil {
		log.Error(errors.Wrap(err, "fail to get random excuse"))
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
func (c ExcuseController) AddExcuse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	log.WithField("function", "AddExcuse").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	var excuse models.Codexcuse
	_ = json.NewDecoder(r.Body).Decode(&excuse)

	var retErrors []string
	if excuse.Author == nil || excuse.Author.UserName == "" {
		errorStr := "missing Author field"
		log.Debugln("fail to save excuse", errorStr)
		retErrors = append(retErrors, errorStr)
	}
	if excuse.Reporter == nil || excuse.Reporter.UserName == "" || excuse.Reporter.ID == "" {
		errorStr := "missing reporter field"
		log.Debugln("fail to save excuse", errorStr)
		retErrors = append(retErrors, errorStr)
	}
	if excuse.Content == "" {
		errorStr := "missing content field"
		log.Debugln("fail to save excuse", errorStr)
		retErrors = append(retErrors, errorStr)
	}
	if excuse.Title == "" {
		errorStr := "missing title field"
		log.Debugln("fail to save excuse", errorStr)
		retErrors = append(retErrors, errorStr)
	}
	if retErrors != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		errArray := make([]string, 0, len(retErrors))
		for _, attrErrs := range retErrors {
			errArray = append(errArray, fmt.Sprintf("\t→ %s", attrErrs))
		}
		resp := response{
			Message: fmt.Sprintf("invalid arguments:\n%s", strings.Join(errArray, "\n")),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	err := c.RedisStore.Add(ctx, vars["source"], excuse)
	if err != nil {
		log.Error(errors.Wrap(err, "fail to save excuse"))
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
func (c ExcuseController) DeleteExcuse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	log.WithField("function", "DeleteExcuse").Infoln("received on", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	err := c.RedisStore.Delete(ctx, vars["source"], vars["id"])
	if err != nil {
		log.Error(errors.Wrap(err, "fail to delete excuse: "+vars["id"]))
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
