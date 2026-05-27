package transcribe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"go-backend/db"
	"go-backend/mistral"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type Controller struct {
	mistral *mistral.Mistral
	store   *db.Store
}

type transcribeRequest struct {
	Uri string `json:"uri" validate:"required,uri"`
}

func NewController(store *db.Store, mistral *mistral.Mistral) *Controller {
	return &Controller{store: store, mistral: mistral}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /transcribe", c.transcribe)
	mux.HandleFunc("GET /tokenUsage", c.tokenUsage)
}

func (c *Controller) userID(r *http.Request) (uuid.UUID, error) {
	// temp hack
	return uuid.Parse(r.Header.Get("X-User-ID"))
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
	id, err := c.userID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.store.ExecTx(r.Context(), func(q *db.Queries) error {
		_, err := q.AddTokensUsed(r.Context(), db.AddTokensUsedParams{
			ID:         id,
			TokensUsed: int32(response.Usage.TotalTokens),
		})
		return err
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(response.Text))
}

func (c *Controller) tokenUsage(w http.ResponseWriter, r *http.Request) {
	id, err := c.userID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	usr, err := c.store.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	_, _ = w.Write([]byte(strconv.Itoa(int(usr.TokensUsed))))
}
