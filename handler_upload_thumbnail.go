package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, headers, err := r.FormFile("thumbnail")

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse from file", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(headers.Header.Get("content-type"))

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error parsing media type", err)
		return
	}

	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "no media type found", err)
		return
	}

	videoData, err := cfg.db.GetVideo(videoID)

	if err != nil || videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get video", nil)
		return
	}

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "invalid media type", nil)
		return
	}

	key := make([]byte, 32)
	rand.Read(key)

	keyString := base64.RawURLEncoding.EncodeToString(key)

	fileExtention := strings.Split(mediaType, "/") // Does not work for every MIME type.
	filePath := filepath.Join(cfg.assetsRoot, keyString+"."+fileExtention[1])
	tnFile, err := os.Create(filePath)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to create file", err)
		return
	}

	_, err = io.Copy(tnFile, file)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create file", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/%s", cfg.port, filePath)
	videoData.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(videoData)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to update thumbnail", err)
		return
	}

	payload := database.Video{
		ID:           videoData.ID,
		CreatedAt:    videoData.CreatedAt,
		UpdatedAt:    videoData.UpdatedAt,
		ThumbnailURL: videoData.ThumbnailURL,
		VideoURL:     videoData.VideoURL,
	}

	respondWithJSON(w, http.StatusOK, payload)
}
