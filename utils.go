package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	buff := &bytes.Buffer{}
	cmd.Stdout = buff
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	output := buff.Bytes()

	var ffprobeOutput struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	err = json.Unmarshal(output, &ffprobeOutput)
	if err != nil {
		return "", err
	}

	if len(ffprobeOutput.Streams) == 0 {
		return "other", nil
	}

	width := ffprobeOutput.Streams[0].Width
	height := ffprobeOutput.Streams[0].Height

	if width*9 == height*16 {
		return "16:9", nil
	} else if width*16 == height*9 {
		return "9:16", nil
	}

	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFilePath := filePath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return outputFilePath, nil
}
