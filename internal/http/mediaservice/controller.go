package mediaservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice/service"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	UserIdHeader = "UserId"
)

type MediaController struct {
	mediaservice *service.MediaService
	*server.Server
}

type ErrorResponse struct {
	Message string `json:"message"`
}
type UploadResponse struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
}

func New(svc *service.MediaService) *MediaController {
	svr := &MediaController{
		mediaservice: svc,
		Server:       server.NewServer(),
	}
	svr.setupRoutes()
	return svr
}

func (mc *MediaController) setupRoutes() {
	// for example for angebot bilder
	mc.WithHandlerFunc("/multi/{id}", mc.GetCompoundLinks, http.MethodGet)
	mc.WithHandlerFunc("/multi/{id}", mc.UploadToCompoundLinks, http.MethodPost)

	// fuer einzelne bilder wie profilbilder oder banner
	mc.WithHandlerFunc("/image", mc.handleIndex, http.MethodGet)
	mc.WithHandlerFunc("/image", mc.UploadPicture, http.MethodPost)
	mc.WithHandlerFunc("/image/{id}", mc.DownloadPicture, http.MethodGet)
}

// handleIndex godoc
// @Summary      Health check endpoint
// @Description  Simple endpoint to check if the media service is running.
// @Tags         media
// @Produce      plain
// @Success      200  {string}  string  "Hello World"
// @Failure      500  {object}  ErrorResponse
// @Router       /media [get]
func (mc *MediaController) handleIndex(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("Hello World")); err != nil {
		mc.GetLogger().Err(err)
	}
}

// UploadPicture godoc
// @Summary      Upload a picture
// @Description  Uploads an image for the authenticated user.
// @Tags         media
// @Accept       octet-stream
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        file body []byte true "Image file bytes"
//
//	@Success      200  {object}  UploadResponse
//
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /media/upload [post]
func (mc *MediaController) UploadPicture(w http.ResponseWriter, r *http.Request) {
	img, err := io.ReadAll(r.Body)
	if err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		mc.Error(w, "content type cannot be empty", http.StatusBadRequest)
		return
	}

	// get user id from context
	user := r.Header.Get(UserIdHeader)

	// upload image
	name, err := mc.mediaservice.UploadPicture(context.Background(), user, contentType, img)

	if err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	response := struct {
		Name    string `json:"name"`
		Success bool   `json:"success"`
	}{
		Name:    name,
		Success: true,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DownloadPicture godoc
// @Summary      Download a picture
// @Description  Downloads an image by its name.
// @Tags         media
// @Produce      image/jpeg
// @Param        id path string true "Image name"
// @Success      200  {file}  []byte
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /media/{id} [get]
func (mc *MediaController) DownloadPicture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	if name == "" {
		mc.Error(w, "name cannot be empty", http.StatusBadRequest)
		return
	}

	img, err := mc.mediaservice.GetPicture(context.Background(), name)

	if err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	if _, err := w.Write(img); err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetCompoundLinks godoc
// @Summary      Get compound image links
// @Description  Returns a list of image URLs associated with a given user ID.
// @Tags         media
// @Produce      json
// @Param        id path string true "User UUID"
// @Success      200  {array}  string "List of image URLs"
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /media/links/{id} [get]
func (mc *MediaController) GetCompoundLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	uid, err := uuid.Parse(id)
	if id == "" || err != nil {
		mc.Error(w, "id cannot be empty", http.StatusBadRequest)
		return
	}

	links, err := mc.mediaservice.GetMultiPicture(context.Background(), uid)
	for i := range links {
		links[i] = fmt.Sprintf("/media/image/%s", links[i])
	}
	if err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(links); err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UploadToCompoundLinks godoc
// @Summary      Upload image to compound links
// @Description  Uploads an image to the compound links associated with the given ID.
// @Tags         media
// @Accept       octet-stream
// @Param        id path string true "Compound Link ID"
// @Param        Authorization header string true "User JWT token"
// @Produce      json
// @Success      200
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /media/links/{id} [post]
func (mc *MediaController) UploadToCompoundLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		mc.Error(w, "content type cannot be empty", http.StatusBadRequest)
		return
	}

	// get user id from context
	user := r.Header.Get(UserIdHeader)

	if user == "" {
		mc.Error(w, "user id cannot be empty", http.StatusBadRequest)
		return
	}

	img, err := io.ReadAll(r.Body)
	if err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// upload image
	err = mc.mediaservice.UploadPictureToMulti(context.Background(), user, id, contentType, img)
	if err != nil {
		mc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
