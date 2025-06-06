package angebotservice

import (
	"encoding/json"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service"
	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice/msclient"
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
}

func New(svc service.Service) *OfferController {
	svr := &OfferController{
		Server:  server.NewServer(),
		service: svc,
	}
	svr.setupRoutes()
	return svr
}

func (c *OfferController) setupRoutes() {
	c.WithHandlerFunc("/", c.handleCreateOffer, http.MethodPost)
	c.WithHandlerFunc("/{id}", c.handleGetOffer, http.MethodGet)
	c.WithHandlerFunc("/", c.handleGetOfferByFilter, http.MethodGet)
}

func (c *OfferController) handleCreateOffer(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var offer repoangebot.Offer
	json.NewDecoder(r.Body).Decode(&offer)
	offer.Creator = uid
	imageURL := c.CreateMultiImageUrl()
	offerId, err := c.service.CreateOffer(&offer, imageURL)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(struct {
		ID       string `json:"id"`
		ImageURL string `json:"image_url"`
	}{
		ID:       offerId.String(),
		ImageURL: imageURL,
	})
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
	json.NewEncoder(w).Encode(offer)
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
	json.NewEncoder(w).Encode(offers)
}
