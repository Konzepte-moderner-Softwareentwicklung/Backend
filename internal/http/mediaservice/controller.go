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

func (mc *MediaController) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func (mc *MediaController) UploadPicture(w http.ResponseWriter, r *http.Request) {
	img, err := io.ReadAll(r.Body)

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

	json.NewEncoder(w).Encode(response)
}

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
	w.Write(img)
}

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

	json.NewEncoder(w).Encode(links)
}

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
