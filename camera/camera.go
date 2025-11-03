package camera

import (
	"fmt"
	"image"
	"log"
	"math"
	"sync"
	"time"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"

	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
)

type ScreenSize struct {
	Width  int
	Height int
}

var Resolution = map[string]ScreenSize{
	"HD":     {1280, 720},
	"FullHD": {1920, 1080},
}

type StreamStats struct {
	IsStreaming bool
	VideoPaused bool
	AudioPaused bool
	FrameCount  uint64
	CurrentFPS  float64
	Resolution  string
	Duration    time.Duration
	AudioLevel  float64
}

type VideoStream struct {
	track          *mediadevices.VideoTrack
	audioTrack     *mediadevices.AudioTrack
	isStreaming    bool
	videoPaused    bool
	audioPaused    bool
	stopChan       chan struct{}
	pauseVideoChan chan bool
	pauseAudioChan chan bool
	frameCount     uint64
	startTime      time.Time
	fps            float64
	resolution     string
	audioLevel     float64
	mu             sync.RWMutex
}

func StartVideoStream(resolution string, updateFunc func(image.Image)) (*VideoStream, error) {
	size, ok := Resolution[resolution]
	if !ok {
		log.Printf("Unknown resolution %s, defaulting to HD", resolution)
		size = Resolution["HD"]
		resolution = "HD"
	}

	stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.Width = prop.Int(size.Width)
			c.Height = prop.Int(size.Height)
			c.FrameRate = prop.Float(30.0)
		},
		Audio: func(c *mediadevices.MediaTrackConstraints) {
			c.SampleRate = prop.Int(48000)
			c.ChannelCount = prop.Int(1)
		},
	})
	if err != nil {
		log.Printf("Failed to get user media: %v", err)
		return nil, err
	}

	videoTracks := stream.GetVideoTracks()
	if len(videoTracks) == 0 {
		log.Println("No video tracks available")
		return nil, fmt.Errorf("no video tracks available")
	}
	videoTrack := videoTracks[0].(*mediadevices.VideoTrack)

	audioTracks := stream.GetAudioTracks()
	var audioTrack *mediadevices.AudioTrack
	if len(audioTracks) > 0 {
		audioTrack = audioTracks[0].(*mediadevices.AudioTrack)
		log.Println("Audio track acquired")
	} else {
		log.Println("No audio track available")
	}

	reader := videoTrack.NewReader(false)

	vs := &VideoStream{
		track:          videoTrack,
		audioTrack:     audioTrack,
		isStreaming:    true,
		videoPaused:    false,
		audioPaused:    false,
		stopChan:       make(chan struct{}),
		pauseVideoChan: make(chan bool, 1),
		pauseAudioChan: make(chan bool, 1),
		frameCount:     0,
		startTime:      time.Now(),
		fps:            0,
		audioLevel:     -100.0,
		resolution:     resolution,
	}

	go vs.streamLoop(reader, updateFunc)
	if audioTrack != nil {
		go vs.audioLoop()
	}

	log.Printf("Started video stream at %dx%d", size.Width, size.Height)
	return vs, nil
}

func (vs *VideoStream) streamLoop(reader video.Reader, updateFunc func(image.Image)) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	framesSinceLastTick := 0

	for {
		select {
		case <-vs.stopChan:
			log.Println("Stopping video stream loop")
			return
		case paused := <-vs.pauseVideoChan:
			vs.mu.Lock()
			vs.videoPaused = paused
			vs.mu.Unlock()
			if paused {
				log.Println("Video paused")
			} else {
				log.Println("Video resumed")
			}
		case paused := <-vs.pauseAudioChan:
			vs.mu.Lock()
			vs.audioPaused = paused
			vs.mu.Unlock()
			if paused {
				log.Println("Audio paused")
			} else {
				log.Println("Audio resumed")
			}
		case <-ticker.C:
			vs.mu.Lock()
			vs.fps = float64(framesSinceLastTick)
			vs.mu.Unlock()
			framesSinceLastTick = 0
		default:
			frame, release, err := reader.Read()
			if err != nil {
				log.Printf("Error reading frame: %v", err)
				continue
			}

			vs.mu.RLock()
			paused := vs.videoPaused
			vs.mu.RUnlock()

			if !paused {
				updateFunc(frame)
				vs.mu.Lock()
				vs.frameCount++
				vs.mu.Unlock()
				framesSinceLastTick++
			}

			release()
		}
	}
}

func (vs *VideoStream) PauseVideo() {
	select {
	case vs.pauseVideoChan <- true:
	default:
	}
}

func (vs *VideoStream) ResumeVideo() {
	select {
	case vs.pauseVideoChan <- false:
	default:
	}
}

func (vs *VideoStream) PauseAudio() {
	select {
	case vs.pauseAudioChan <- true:
	default:
	}
}

func (vs *VideoStream) ResumeAudio() {
	select {
	case vs.pauseAudioChan <- false:
	default:
	}
}

func (vs *VideoStream) audioLoop() {
	if vs.audioTrack == nil {
		return
	}

	reader := vs.audioTrack.NewReader(false)

	for {
		select {
		case <-vs.stopChan:
			log.Println("Stopping audio loop")
			return
		default:
			chunk, release, err := reader.Read()
			if err != nil {
				log.Printf("Error reading audio: %v", err)
				continue
			}

			vs.mu.RLock()
			muted := vs.audioPaused
			vs.mu.RUnlock()

			if !muted {
				level := calculateAudioLevel(chunk)
				vs.mu.Lock()
				vs.audioLevel = level
				vs.mu.Unlock()
			} else {
				vs.mu.Lock()
				vs.audioLevel = -100.0
				vs.mu.Unlock()
			}

			release()
		}
	}
}

func calculateAudioLevel(chunk wave.Audio) float64 {
	info := chunk.ChunkInfo()
	if info.Len == 0 {
		return -100.0
	}

	var sumSquares float64

	for i := 0; i < info.Len; i++ {
		for ch := 0; ch < info.Channels; ch++ {
			sample := chunk.At(i, ch)
			var normalized float64

			switch s := sample.(type) {
			case wave.Int16Sample:
				normalized = float64(s) / 32768.0
			case wave.Float32Sample:
				normalized = float64(s)
			default:
				normalized = 0.0
			}

			sumSquares += normalized * normalized
		}
	}

	totalSamples := info.Len * info.Channels
	if totalSamples == 0 {
		return -100.0
	}

	rms := math.Sqrt(sumSquares / float64(totalSamples))

	if rms < 0.00001 {
		return -100.0
	}

	db := 20.0 * math.Log10(rms)
	return db
}

func (vs *VideoStream) GetStats() StreamStats {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	size := Resolution[vs.resolution]
	resolutionStr := fmt.Sprintf("%s - %dx%d", vs.resolution, size.Width, size.Height)

	return StreamStats{
		IsStreaming: vs.isStreaming,
		VideoPaused: vs.videoPaused,
		AudioPaused: vs.audioPaused,
		FrameCount:  vs.frameCount,
		CurrentFPS:  vs.fps,
		Resolution:  resolutionStr,
		Duration:    time.Since(vs.startTime),
		AudioLevel:  vs.audioLevel,
	}
}

func (vs *VideoStream) Stop() error {
	if !vs.isStreaming {
		return nil
	}

	log.Println("Stopping video stream")
	vs.mu.Lock()
	vs.isStreaming = false
	vs.mu.Unlock()
	close(vs.stopChan)

	if vs.track != nil {
		vs.track.Close()
	}

	if vs.audioTrack != nil {
		vs.audioTrack.Close()
	}

	return nil
}
