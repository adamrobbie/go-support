package video

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/adamrobbie/go-support/pkg/screenshot"
)

// Quality represents the quality of the video stream
type Quality int

const (
	// Low quality (faster, smaller file size)
	Low Quality = iota
	// Medium quality (balanced)
	Medium
	// High quality (slower, larger file size)
	High
)

// VideoStream represents a video stream
type VideoStream struct {
	quality        Quality
	fps            int
	isStreaming    bool
	isRecording    bool
	ctx            context.Context
	cancel         context.CancelFunc
	mutex          sync.Mutex
	frames         [][]byte
	onFrameCapture func([]byte) error
	verbose        bool
}

// NewVideoStream creates a new video stream
func NewVideoStream(quality Quality, fps int, verbose bool) *VideoStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &VideoStream{
		quality:     quality,
		fps:         fps,
		isStreaming: false,
		isRecording: false,
		ctx:         ctx,
		cancel:      cancel,
		frames:      make([][]byte, 0),
		verbose:     verbose,
	}
}

// SetOnFrameCapture sets the callback function to be called when a frame is captured
func (v *VideoStream) SetOnFrameCapture(callback func([]byte) error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.onFrameCapture = callback
}

// StartStreaming starts streaming video frames
func (v *VideoStream) StartStreaming() error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.isStreaming {
		return fmt.Errorf("streaming is already in progress")
	}

	v.isStreaming = true
	go v.streamLoop()

	if v.verbose {
		log.Printf("Started video streaming at %d FPS with quality %d", v.fps, v.quality)
	}

	return nil
}

// StopStreaming stops streaming video frames
func (v *VideoStream) StopStreaming() {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if !v.isStreaming {
		return
	}

	v.isStreaming = false
	v.cancel()

	// Create a new context for future streaming
	v.ctx, v.cancel = context.WithCancel(context.Background())

	if v.verbose {
		log.Println("Stopped video streaming")
	}
}

// StartRecording starts recording video frames
func (v *VideoStream) StartRecording() error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.isRecording {
		return fmt.Errorf("recording is already in progress")
	}

	// Clear previous frames
	v.frames = make([][]byte, 0)
	v.isRecording = true

	// Start streaming if not already streaming
	if !v.isStreaming {
		go v.streamLoop()
		v.isStreaming = true
	}

	if v.verbose {
		log.Println("Started video recording")
	}

	return nil
}

// StopRecording stops recording video frames and returns the recorded frames
func (v *VideoStream) StopRecording() ([][]byte, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if !v.isRecording {
		return nil, fmt.Errorf("no recording in progress")
	}

	v.isRecording = false
	frames := v.frames

	// If we're not streaming for any other reason, stop the stream loop
	if !v.isStreaming {
		v.cancel()
		// Create a new context for future streaming
		v.ctx, v.cancel = context.WithCancel(context.Background())
	}

	if v.verbose {
		log.Printf("Stopped video recording, captured %d frames", len(frames))
	}

	return frames, nil
}

// IsStreaming returns true if streaming is in progress
func (v *VideoStream) IsStreaming() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.isStreaming
}

// IsRecording returns true if recording is in progress
func (v *VideoStream) IsRecording() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.isRecording
}

// streamLoop captures frames at the specified FPS
func (v *VideoStream) streamLoop() {
	interval := time.Second / time.Duration(v.fps)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-v.ctx.Done():
			return
		case <-ticker.C:
			frame, err := v.captureFrame()
			if err != nil {
				if v.verbose {
					log.Printf("Error capturing frame: %v", err)
				}
				continue
			}

			v.mutex.Lock()
			// If recording, store the frame
			if v.isRecording {
				v.frames = append(v.frames, frame)
			}

			// If there's a callback, call it
			if v.onFrameCapture != nil {
				callback := v.onFrameCapture
				v.mutex.Unlock()

				// Call the callback outside the lock
				err := callback(frame)
				if err != nil && v.verbose {
					log.Printf("Error in frame capture callback: %v", err)
				}
			} else {
				v.mutex.Unlock()
			}
		}
	}
}

// captureFrame captures a single frame
func (v *VideoStream) captureFrame() ([]byte, error) {
	// Convert quality to screenshot quality
	var ssQuality screenshot.Quality
	switch v.quality {
	case Low:
		ssQuality = screenshot.Low
	case Medium:
		ssQuality = screenshot.Medium
	case High:
		ssQuality = screenshot.High
	default:
		ssQuality = screenshot.Medium
	}

	// Capture screenshot
	ss, err := screenshot.Capture(ssQuality)
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return ss.Data, nil
}

// SaveRecordingAsImages saves the recorded frames as individual images
func (v *VideoStream) SaveRecordingAsImages(directory string, prefix string) error {
	v.mutex.Lock()
	frames := v.frames
	v.mutex.Unlock()

	if len(frames) == 0 {
		return fmt.Errorf("no frames to save")
	}

	for i, frame := range frames {
		filename := fmt.Sprintf("%s/%s_%04d.jpg", directory, prefix, i)

		// Save the frame directly to file
		err := os.WriteFile(filename, frame, 0644)
		if err != nil {
			return fmt.Errorf("failed to save frame %d: %w", i, err)
		}
	}

	if v.verbose {
		log.Printf("Saved %d frames to %s with prefix %s", len(frames), directory, prefix)
	}

	return nil
}

// GetFrameCount returns the number of recorded frames
func (v *VideoStream) GetFrameCount() int {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return len(v.frames)
}

// GetFrame returns a specific frame
func (v *VideoStream) GetFrame(index int) ([]byte, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if index < 0 || index >= len(v.frames) {
		return nil, fmt.Errorf("frame index out of range")
	}

	return v.frames[index], nil
}
