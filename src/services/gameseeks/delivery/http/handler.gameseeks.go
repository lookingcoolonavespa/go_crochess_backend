package delivery_http_gameseeks

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
)

type GameseeksHandler struct {
	repo domain.GameseeksRepo
}

func NewGameseeksHandler(r *httprouter.Router, ctx domain.GameseeksRepo) *httprouter.Router {
	handler := &GameseeksHandler{ctx}

	r.GET("/gameseeks", handler.HandlerGetGameseeksList)
	r.POST("/gameseeks", handler.HandleGameseekInsert)

	return r
}

func (g *GameseeksHandler) HandlerGetGameseeksList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	list, err := g.repo.List()
	if err != nil {
		log.Printf("%s : %v", "GameseeksHandler/HandlerGetGameseeksList/List/ShouldFindList", err)
		http.Error(w, fmt.Sprintf("There was an error retreiving game seeks. %v", err), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(list)
	if err != nil {
		log.Printf("%s : %v", "GameseeksHandler/HandlerGetGameseeksList/List/ShouldEncodeIntoJson", err)
		http.Error(w, fmt.Sprintf("There was an error retreiving game seeks. %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}

}

func (g *GameseeksHandler) HandleGameseekInsert(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var param domain.Gameseek
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	err := g.repo.Insert(&param)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save game seek: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
