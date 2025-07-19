package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os/exec"
)

type VideoAspectRatio struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

func getVideoAspectRatio(file_path string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", file_path)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	var aspectRatio VideoAspectRatio

	if err := json.Unmarshal(buffer.Bytes(), &aspectRatio); err != nil {
		return "", err
	}

	ratio := float64(aspectRatio.Streams[0].Width) / float64(aspectRatio.Streams[0].Height)

	if math.Abs(ratio-16.0/9.0) < 0.1 {
		return "16:9", nil
	}

	if math.Abs(ratio-9.0/16.0) < 0.1 {
		return "9:16", nil
	}

	return "other", nil

}
