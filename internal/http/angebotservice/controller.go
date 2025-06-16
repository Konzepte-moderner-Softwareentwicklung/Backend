package angebotservice

import (
	"encoding/json"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service"
	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice/msclient"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	UserIdHeader = "UserId"
)

type OfferController struct {
	*server.Server
	*msclient.Client
	service service.Service
	*auth.AuthMiddleware
}

func New(svc service.Service, secret []byte) *OfferController {
	svr := &OfferController{
		Server:         server.NewServer(),
		service:        svc,
		AuthMiddleware: auth.NewAuthMiddleware(secret),
	}
	svr.setupRoutes()
	return svr
}

func (c *OfferController) setupRoutes() {
	c.WithHandlerFunc("/filter", c.handleGetOfferByFilter, http.MethodPost)
	c.WithHandlerFunc("/", c.EnsureJWT(c.handleCreateOffer), http.MethodPost)
	c.WithHandlerFunc("/{id}", c.handleGetOffer, http.MethodGet)
}

func (c *OfferController) handleCreateOffer(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var offer repoangebot.Offer
	err = json.NewDecoder(r.Body).Decode(&offer)
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	offer.Creator = uid
	imageURL := c.CreateMultiImageUrl()
	offerId, err := c.service.CreateOffer(&offer, imageURL)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.NewEncoder(w).Encode(struct {
		ID       string `json:"id"`
		ImageURL string `json:"image_url"`
	}{
		ID:       offerId.String(),
		ImageURL: imageURL,
	}); err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *OfferController) handleGetOffer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	offer, err := c.service.GetOffer(uid)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(offer); err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *OfferController) handleGetOfferByFilter(w http.ResponseWriter, r *http.Request) {
	var filter repoangebot.Filter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	offers, err := c.service.GetOffersByFilter(filter)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(offers); err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
