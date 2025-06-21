package angebotservice

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice"
	"github.com/nats-io/nats.go"

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
	*nats.Conn
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func New(svc service.Service, secret []byte) *OfferController {
	NATS_URL := os.Getenv("NATS_URL")
	conn, err := nats.Connect(NATS_URL)
	if err != nil {
		panic(err)
	}

	svr := &OfferController{
		Server:         server.NewServer(),
		service:        svc,
		AuthMiddleware: auth.NewAuthMiddleware(secret),
		Conn:           conn,
	}
	svr.setupRoutes()
	return svr
}

func (c *OfferController) setupRoutes() {
	c.WithHandlerFunc("/filter", c.handleGetOfferByFilter, http.MethodPost)
	c.WithHandlerFunc("/", c.EnsureJWT(c.handleCreateOffer), http.MethodPost)
	c.WithHandlerFunc("/{id}", c.handleGetOffer, http.MethodGet)
	c.WithHandlerFunc("/{id}/occupy", c.EnsureJWT(c.OccupyOffer), http.MethodPost)

	c.WithHandlerFunc("/{id}/rating", c.EnsureJWT(c.handlePostRating), http.MethodPost)
}

// OccupyOffer godoc
// @Summary      Occupy an offer
// @Description  Marks an offer as occupied by the user, specifying the desired space parameters.
// @Tags         offers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        id path string true "Offer ID (UUID)"
// @Param        body body  repoangebot.Space true "Space details for the occupation"
// @Success      200
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /offers/{id}/occupy [post]
func (c *OfferController) OccupyOffer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(r.Header.Get(UserIdHeader))
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	space := repoangebot.Space{}
	if err := json.NewDecoder(r.Body).Decode(&space); err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.service.OccupieOffer(id, userId, space); err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handlePostRating godoc
// @Summary      Post a rating
// @Description  Submits a rating for a specific offer.
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        id path string true "Offer ID (UUID)"
// @Param        body body ratingservice.Rating true "Rating object"
// @Success      200
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /offers/{id}/rating [post]
func (c *OfferController) handlePostRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := uuid.Parse(vars["id"])
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(r.Header.Get(UserIdHeader))
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
	}
	var rating ratingservice.Rating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer r.Body.Close()
	if err = c.Publish("rating."+userId.String(), body); err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleCreateOffer godoc
// @Summary      Create a new offer
// @Description  Creates a new offer by the authenticated user, generates image URLs.
// @Tags         offers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        body body repoangebot.Offer true "Offer data"
// @Success      200  {object}  CreateOfferResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /offers [post]
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

// handleGetOffer godoc
// @Summary      Get offer details
// @Description  Retrieves detailed information about a specific offer by ID.
// @Tags         offers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        id path string true "Offer ID (UUID)"
// @Success      200  {object}  repoangebot.Offer
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /offers/{id} [get]
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

// handleGetOfferByFilter godoc
// @Summary      Get offers by filter
// @Description  Retrieves a list of offers filtered by the specified criteria.
// @Tags         offers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        body body repoangebot.Filter true "Filter criteria"
// @Success      200  {array}   []repoangebot.Offer
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /offers/filter [post]
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
