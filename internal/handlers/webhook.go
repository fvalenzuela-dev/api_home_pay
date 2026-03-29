package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	svix "github.com/svix/svix-webhooks/go"
)

type WebhookHandler struct {
	users         repository.UserRepository
	webhookSecret string
}

func NewWebhookHandler(users repository.UserRepository, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{users: users, webhookSecret: webhookSecret}
}

type clerkEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type clerkUserData struct {
	ID             string `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	EmailAddresses []struct {
		EmailAddress string `json:"email_address"`
	} `json:"email_addresses"`
}

// Handle godoc
// @Summary     Webhook de Clerk
// @Description Recibe eventos de Clerk (user.created, user.updated, user.deleted). Verifica la firma con svix.
// @Tags        webhook
// @Accept      json
// @Produce     json
// @Success     200
// @Failure     400  {object}  map[string]string
// @Failure     401  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /webhooks/clerk [post]
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "error reading body")
		return
	}

	wh, err := svix.NewWebhook(h.webhookSecret)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}

	if err := wh.Verify(body, r.Header); err != nil {
		writeError(w, http.StatusUnauthorized, "firma inválida")
		return
	}

	var event clerkEvent
	if err := json.Unmarshal(body, &event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	switch event.Type {
	case "user.created", "user.updated":
		var data clerkUserData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			writeError(w, http.StatusBadRequest, "invalid user data")
			return
		}
		email := ""
		if len(data.EmailAddresses) > 0 {
			email = data.EmailAddresses[0].EmailAddress
		}
		user := &models.User{
			AuthUserID: data.ID,
			Email:      email,
			FullName:   data.FirstName + " " + data.LastName,
		}
		if err := h.users.Upsert(r.Context(), user); err != nil {
			writeInternalError(w, r, err)
			return
		}

	case "user.deleted":
		var data struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			writeError(w, http.StatusBadRequest, "invalid user data")
			return
		}
		if err := h.users.SoftDelete(r.Context(), data.ID); err != nil {
			writeInternalError(w, r, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
