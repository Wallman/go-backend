package transcribe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"go-backend/mistral"
	"go-backend/user"
	"net/http"
	"strconv"
)

type Controller struct {
	mistral *mistral.Mistral
	repo    *user.Repository
}

type transcribeRequest struct {
	Uri string `json:"uri" validate:"required,uri"`
}

var userId = "b9c43ba1-9572-4aac-b945-ccab71c932b5"

func NewController(repo *user.Repository, mistral *mistral.Mistral) *Controller {
	return &Controller{repo: repo, mistral: mistral}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /transcribe", c.transcribe)
	mux.HandleFunc("GET /tokenUsage", c.tokenUsage)
}

func (c *Controller) transcribe(w http.ResponseWriter, r *http.Request) {
	request := transcribeRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := c.mistral.Transcribe(r.Context(), request.Uri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.repo.AddTokensUsed(r.Context(), response.Usage.TotalTokens, userId)
	w.Write([]byte(response.Text))
}

func (c *Controller) tokenUsage(w http.ResponseWriter, r *http.Request) {
	usr, err := c.repo.Get(r.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	_, _ = w.Write([]byte(strconv.Itoa(usr.TokensUsed)))
}
