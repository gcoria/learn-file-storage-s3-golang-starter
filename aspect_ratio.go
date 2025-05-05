package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

type ffprobeOutput struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

// getVideoAspectRatio determines the aspect ratio of a video file
// Returns "16:9", "9:16", or "other" depending on the aspect ratio
func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var output ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return "", fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	if len(output.Streams) == 0 {
		return "", fmt.Errorf("no streams found in video file")
	}

	width := output.Streams[0].Width
	height := output.Streams[0].Height

	if width == 0 || height == 0 {
		return "", fmt.Errorf("invalid dimensions in video file: %dx%d", width, height)
	}

	ratio := float64(width) / float64(height)

	// Check if ratio is approximately 16:9 (landscape)
	if math.Abs(ratio-16.0/9.0) < 0.1 {
		return "16:9", nil
	}

	// Check if ratio is approximately 9:16 (portrait)
	if math.Abs(ratio-9.0/16.0) < 0.1 {
		return "9:16", nil
	}

	// Otherwise, it's some other aspect ratio
	return "other", nil
}

func processVideoForFastStart(filepath string) (string, error) {
	outputPath := filepath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-f", "mp4", "-movflags", "faststart", outputPath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to process video for fast start: %w", err)
	}

	return outputPath, nil
}

// getVideoPrefixFromAspectRatio returns the prefix to use for the S3 key
// based on the aspect ratio
func getVideoPrefixFromAspectRatio(aspectRatio string) string {
	switch aspectRatio {
	case "16:9":
		return "landscape/"
	case "9:16":
		return "portrait/"
	default:
		return "other/"
	}
}
