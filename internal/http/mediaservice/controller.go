package mediaservice

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice/service"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/gorilla/mux"
)

const (
	UserIdHeader = "UserId"
)

type MediaController struct {
	mediaservice *service.MediaService
	*server.Server
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
	mc.WithHandlerFunc("/", mc.handleIndex, http.MethodGet)
	mc.WithHandlerFunc("/", mc.UploadPicture, http.MethodPost)
	mc.WithHandlerFunc("/{id}", mc.DownloadPicture, http.MethodGet)
}

func (mc *MediaController) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func (mc *MediaController) UploadPicture(w http.ResponseWriter, r *http.Request) {
	img, err := io.ReadAll(r.Body)

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		http.Error(w, "content type cannot be empty", http.StatusBadRequest)
		return
	}

	// get user id from context
	user := r.Header.Get(UserIdHeader)

	// upload image
	name, err := mc.mediaservice.UploadPicture(context.Background(), user, contentType, img)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	json.NewEncoder(w).Encode(response)
}

func (mc *MediaController) DownloadPicture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	if name == "" {
		http.Error(w, "name cannot be empty", http.StatusBadRequest)
		return
	}

	img, err := mc.mediaservice.GetPicture(context.Background(), name)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(img)
}
