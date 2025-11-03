package gui

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"github.com/javanhut/zero/camera"
	"github.com/javanhut/zero/sessionmanager"
	"github.com/javanhut/zero/signaling"
	"github.com/javanhut/zero/webrtc"
	pwebrtc "github.com/pion/webrtc/v4"
)

func calculateAudioColor(dbLevel float64) color.Color {
	normalized := (dbLevel + 60) / 60
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	var r, g, b uint8

	if normalized < 0.33 {
		t := normalized / 0.33
		r = 0
		g = uint8(100 + (155 * t))
		b = 0
	} else if normalized < 0.66 {
		t := (normalized - 0.33) / 0.33
		r = uint8(255 * t)
		g = 255
		b = 0
	} else {
		t := (normalized - 0.66) / 0.34
		r = 255
		g = uint8(255 * (1 - t))
		b = 0
	}

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func calculateAudioSize(dbLevel float64) float32 {
	normalized := (dbLevel + 60) / 60
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	minSize := float32(10)
	maxSize := float32(30)

	return minSize + float32(normalized)*(maxSize-minSize)
}

func showStatsDialog(a fyne.App, vs *camera.VideoStream) {
	if vs == nil {
		return
	}

	statsWindow := a.NewWindow("Stream Statistics")
	statsWindow.Resize(fyne.NewSize(400, 400))

	statusLabel := widget.NewLabel("")
	videoStatusLabel := widget.NewLabel("")
	audioStatusLabel := widget.NewLabel("")
	resolutionLabel := widget.NewLabel("")
	fpsLabel := widget.NewLabel("")
	framesLabel := widget.NewLabel("")
	durationLabel := widget.NewLabel("")
	audioLevelLabel := widget.NewLabel("")

	updateStats := func() {
		stats := vs.GetStats()

		fyne.Do(func() {
			statusText := "Stopped"
			if stats.IsStreaming {
				statusText = "Active"
			}
			statusLabel.SetText(fmt.Sprintf("Status: %s", statusText))

			videoStatusText := "Active"
			if stats.VideoPaused {
				videoStatusText = "Paused"
			}
			videoStatusLabel.SetText(fmt.Sprintf("Video: %s", videoStatusText))

			audioStatusText := "Active"
			if stats.AudioPaused {
				audioStatusText = "Muted"
			}
			audioStatusLabel.SetText(fmt.Sprintf("Audio: %s", audioStatusText))

			resolutionLabel.SetText(fmt.Sprintf("Resolution: %s", stats.Resolution))
			fpsLabel.SetText(fmt.Sprintf("Frame Rate: %.1f FPS", stats.CurrentFPS))
			framesLabel.SetText(fmt.Sprintf("Frames Processed: %d", stats.FrameCount))
			durationLabel.SetText(fmt.Sprintf("Duration: %s", stats.Duration.Round(time.Second)))
			audioLevelLabel.SetText(fmt.Sprintf("Audio Level: %.1f dB", stats.AudioLevel))
		})
	}

	updateStats()

	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			updateStats()
		}
	}()

	content := container.NewVBox(
		widget.NewLabel("Stream Status"),
		widget.NewSeparator(),
		statusLabel,
		videoStatusLabel,
		audioStatusLabel,
		widget.NewSeparator(),
		widget.NewLabel("Stream Details"),
		widget.NewSeparator(),
		resolutionLabel,
		fpsLabel,
		framesLabel,
		durationLabel,
		audioLevelLabel,
		widget.NewSeparator(),
		widget.NewButton("Close", func() {
			ticker.Stop()
			statsWindow.Close()
		}),
	)

	statsWindow.SetOnClosed(func() {
		ticker.Stop()
	})

	statsWindow.SetContent(container.NewCenter(content))
	statsWindow.Show()
}

func Gui() {
	sessions := sessionmanager.New()
	a := app.New()
	w := a.NewWindow("Session Login")
	videoWindow := a.NewWindow("Video Window")
	w.Resize(fyne.NewSize(720, 477))
	videoWindow.Resize(fyne.NewSize(1280, 720))

	var currentSessionID string
	var currentUsername string
	var currentPeerID string
	var videoStream *camera.VideoStream
	var signalingClient *signaling.Client
	var webrtcManager *webrtc.Manager
	signalingServerURL := "ws://localhost:8080/ws"

	videoCanvas := canvas.NewImageFromImage(nil)
	videoCanvas.FillMode = canvas.ImageFillOriginal
	videoCanvas.ScaleMode = canvas.ImageScaleSmooth
	videoCanvas.SetMinSize(fyne.NewSize(1280, 720))

	videoLabel := widget.NewLabel("Video stream will appear here...")

	pauseBackground := canvas.NewRectangle(color.Black)
	pauseBackground.SetMinSize(fyne.NewSize(1280, 720))

	pauseUsernameLabel := widget.NewLabel("")
	pauseUsernameLabel.TextStyle = fyne.TextStyle{Bold: true}
	pauseUsernameLabel.Alignment = fyne.TextAlignCenter

	pauseTextLabel := widget.NewLabel("Video Paused")
	pauseTextLabel.Alignment = fyne.TextAlignCenter

	pauseContainer := container.NewCenter(
		container.NewVBox(
			pauseUsernameLabel,
			pauseTextLabel,
		),
	)

	pauseOverlay := container.NewStack(pauseBackground, pauseContainer)
	pauseOverlay.Hide()

	audioCircle := canvas.NewCircle(color.RGBA{R: 0, G: 255, B: 0, A: 255})
	audioCircle.Resize(fyne.NewSize(30, 30))

	audioMeterLabel := widget.NewLabel("Mic")
	audioMeterLabel.Alignment = fyne.TextAlignCenter

	audioMeterContainer := container.NewVBox(
		audioCircle,
		audioMeterLabel,
	)

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if videoStream != nil {
				stats := videoStream.GetStats()
				level := stats.AudioLevel

				fyne.Do(func() {
					circleColor := calculateAudioColor(level)
					circleSize := calculateAudioSize(level)

					audioCircle.FillColor = circleColor
					audioCircle.Resize(fyne.NewSize(circleSize, circleSize))
					audioCircle.Refresh()
				})
			}
		}
	}()

	cameraEnabled := true
	audioEnabled := true

	var cameraBtn *widget.Button
	var audioBtn *widget.Button
	var statsBtn *widget.Button

	cameraBtn = widget.NewButton("Camera On", func() {
		if videoStream == nil {
			return
		}
		if cameraEnabled {
			videoStream.PauseVideo()
			cameraBtn.SetText("Camera Off")
			cameraEnabled = false
			pauseUsernameLabel.SetText(currentUsername)
			pauseOverlay.Show()
		} else {
			videoStream.ResumeVideo()
			cameraBtn.SetText("Camera On")
			cameraEnabled = true
			pauseOverlay.Hide()
		}
	})

	audioBtn = widget.NewButton("Audio On", func() {
		if videoStream == nil {
			return
		}
		if audioEnabled {
			videoStream.PauseAudio()
			audioBtn.SetText("Audio Off")
			audioEnabled = false
		} else {
			videoStream.ResumeAudio()
			audioBtn.SetText("Audio On")
			audioEnabled = true
		}
	})

	statsBtn = widget.NewButton("Stats", func() {
		showStatsDialog(a, videoStream)
	})

	cameraBtn.Disable()
	audioBtn.Disable()
	statsBtn.Disable()

	buttonContainer := container.NewHBox(cameraBtn, audioBtn, statsBtn, audioMeterContainer)
	controlPanel := container.NewVBox(
		container.NewCenter(buttonContainer),
	)

	videoContainer := container.NewStack(
		videoCanvas,
		pauseOverlay,
		videoLabel,
		container.NewBorder(nil, controlPanel, nil, nil),
	)
	videoWindow.SetContent(videoContainer)

	updateVideo := func(frame image.Image) {
		fyne.Do(func() {
			videoCanvas.Image = frame
			videoCanvas.Refresh()
		})
	}

	videoWindow.SetCloseIntercept(func() {
		if webrtcManager != nil {
			webrtcManager.Close()
			webrtcManager = nil
		}
		if signalingClient != nil {
			signalingClient.Disconnect()
			signalingClient = nil
		}
		if videoStream != nil {
			videoStream.Stop()
			videoStream = nil
		}
		if currentSessionID != "" && currentPeerID != "" {
			sessions.RemovePeerFromSession(currentSessionID, currentPeerID)
		}
		cameraBtn.Disable()
		audioBtn.Disable()
		statsBtn.Disable()
		cameraEnabled = true
		audioEnabled = true
		cameraBtn.SetText("Camera On")
		audioBtn.SetText("Audio On")
		pauseOverlay.Hide()
		videoWindow.Hide()
	})

	startSession := widget.NewLabel("Zero App")
	entry := widget.NewEntry()
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Session ID", Widget: entry},
		},
	}
	w.SetContent(
		container.NewVBox(
			startSession,
			form,
			container.NewHBox(
				widget.NewButton("Start New Session", func() {
					log.Println("Creating new session....")
					sessionID, username := sessions.CreateNewSession()
					currentSessionID = sessionID
					currentUsername = username
					currentPeerID = uuid.New().String()
					entry.SetText(currentSessionID)

					videoLabel.SetText("Starting camera...")
					videoWindow.Show()

					stream, err := camera.StartVideoStream("HD", updateVideo)
					if err != nil {
						log.Printf("Failed to start camera: %v", err)
						videoLabel.SetText(fmt.Sprintf("Camera error: %v", err))
						return
					}

					videoStream = stream
					videoLabel.SetText("Connecting to signaling server...")

					signalingClient = signaling.NewClient(signalingServerURL, currentSessionID, currentPeerID, currentUsername)
					if err := signalingClient.Connect(); err != nil {
						log.Printf("Failed to connect to signaling server: %v", err)
						videoLabel.SetText(fmt.Sprintf("Signaling error: %v", err))
						return
					}

					webrtcConfig := webrtc.DefaultConfig()
					webrtcManager = webrtc.NewManager(webrtc.ManagerConfig{
						WebRTCConfig:    webrtcConfig,
						SignalingClient: signalingClient,
						OnRemoteTrack: func(peerID string, track *pwebrtc.TrackRemote, receiver *pwebrtc.RTPReceiver) {
							log.Printf("Received remote track from peer %s: %s", peerID, track.Kind().String())
						},
						OnPeerDisconnect: func(peerID string) {
							log.Printf("Peer disconnected: %s", peerID)
						},
					})

					videoTrack, audioTrack, err := videoStream.CreateWebRTCTracks()
					if err != nil {
						log.Printf("Failed to create WebRTC tracks: %v", err)
					} else {
						if videoTrack != nil {
							webrtcManager.AddLocalTrack(videoTrack)
						}
						if audioTrack != nil {
							webrtcManager.AddLocalTrack(audioTrack)
						}
					}

					videoLabel.SetText("")
					cameraBtn.Enable()
					audioBtn.Enable()
					statsBtn.Enable()
					cameraEnabled = true
					audioEnabled = true
					cameraBtn.SetText("Camera On")
					audioBtn.SetText("Audio On")
					log.Printf("Session created and WebRTC initialized: %s", currentSessionID)
				}),
				widget.NewButton("Connect", func() {
					log.Println("Attempting to connect to session....")
					sessionIDInput := entry.Text
					if sessionIDInput == "" {
						log.Println("Please enter a session ID")
						return
					}

					currentSessionID = sessionIDInput
					currentPeerID = uuid.New().String()

					peerID, username, err := sessions.JoinSession(currentSessionID)
					if err != nil {
						log.Printf("Failed to join session: %v", err)
						videoLabel.SetText(fmt.Sprintf("Join error: %v", err))
						return
					}
					currentPeerID = peerID
					currentUsername = username

					videoLabel.SetText("Starting camera...")
					videoWindow.Show()

					stream, err := camera.StartVideoStream("HD", updateVideo)
					if err != nil {
						log.Printf("Failed to start camera: %v", err)
						videoLabel.SetText(fmt.Sprintf("Camera error: %v", err))
						return
					}

					videoStream = stream
					videoLabel.SetText("Connecting to signaling server...")

					signalingClient = signaling.NewClient(signalingServerURL, currentSessionID, currentPeerID, currentUsername)
					if err := signalingClient.Connect(); err != nil {
						log.Printf("Failed to connect to signaling server: %v", err)
						videoLabel.SetText(fmt.Sprintf("Signaling error: %v", err))
						return
					}

					webrtcConfig := webrtc.DefaultConfig()
					webrtcManager = webrtc.NewManager(webrtc.ManagerConfig{
						WebRTCConfig:    webrtcConfig,
						SignalingClient: signalingClient,
						OnRemoteTrack: func(peerID string, track *pwebrtc.TrackRemote, receiver *pwebrtc.RTPReceiver) {
							log.Printf("Received remote track from peer %s: %s", peerID, track.Kind().String())
						},
						OnPeerDisconnect: func(peerID string) {
							log.Printf("Peer disconnected: %s", peerID)
						},
					})

					videoTrack, audioTrack, err := videoStream.CreateWebRTCTracks()
					if err != nil {
						log.Printf("Failed to create WebRTC tracks: %v", err)
					} else {
						if videoTrack != nil {
							webrtcManager.AddLocalTrack(videoTrack)
						}
						if audioTrack != nil {
							webrtcManager.AddLocalTrack(audioTrack)
						}
					}

					videoLabel.SetText("")
					cameraBtn.Enable()
					audioBtn.Enable()
					statsBtn.Enable()
					cameraEnabled = true
					audioEnabled = true
					cameraBtn.SetText("Camera On")
					audioBtn.SetText("Audio On")
					log.Printf("Connected to session: %s", currentSessionID)
				}),
			),
		),
	)
	w.ShowAndRun()
}
