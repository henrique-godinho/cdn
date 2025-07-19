package main

import (
	"os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {
	processingFile := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", processingFile)

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return processingFile, nil
}
