// This replaces the "final mix" from a beep speaker
// and wraps it in a similar interface for use with SDL
package main

import (
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/veandco/go-sdl2/sdl"
)

type AudioSink interface {
	Init(sr beep.SampleRate, bufferSize int) error
	Play(s beep.Streamer)
	Lock()
	Unlock()
	Close()
	FillAudio()
}

var speaker AudioSink

type SDLSpeaker struct {
	dev        sdl.AudioDeviceID
	mixer      *beep.Mixer
	mu         sync.Mutex
	sampleRate beep.SampleRate
	bufferSize int
	buf        [][2]float64
}

func floatToS16(v float64) int16 {
	if v > 1 {
		v = 1
	} else if v < -1 {
		v = -1
	}
	return int16(v * 32767)
}

func (s *SDLSpeaker) Init(sampleRate beep.SampleRate, bufferSize int) error {
	s.sampleRate = sampleRate
	s.mixer = &beep.Mixer{}
	s.buf = make([][2]float64, bufferSize)
	s.bufferSize = bufferSize

	spec := sdl.AudioSpec{
		Freq:     int32(s.sampleRate),
		Format:   sdl.AUDIO_S16SYS,
		Channels: 2,
		Samples:  uint16(bufferSize),
	}

	dev, err := sdl.OpenAudioDevice("", false, &spec, nil, 0)
	if err != nil {
		return err
	}

	s.dev = dev
	sdl.ClearQueuedAudio(s.dev)
	sdl.PauseAudioDevice(s.dev, false)

	// Start the audio pushing thread
	SafeGo(func() {
		for true {
			s.FillAudio()
			time.Sleep(time.Millisecond * 17) // prevents fans from going crazy on Android handhelds
		}
	})

	return nil
}

func (s *SDLSpeaker) FillAudio() {
	if s.dev == 0 {
		return
	}

	// Only queue if SDL needs data
	if sdl.GetQueuedAudioSize(s.dev) > uint32(s.bufferSize*4) {
		return
	}

	frames := s.bufferSize
	buf := make([]byte, frames*4)

	s.mu.Lock()
	n, _ := s.mixer.Stream(s.buf[:frames])
	s.mu.Unlock()

	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		l := floatToS16(s.buf[i][0])
		r := floatToS16(s.buf[i][1])

		buf[i*4] = byte(l)
		buf[i*4+1] = byte(l >> 8)
		buf[i*4+2] = byte(r)
		buf[i*4+3] = byte(r >> 8)
	}

	sdl.QueueAudio(s.dev, buf)
}

func (s *SDLSpeaker) Play(st beep.Streamer) {
	s.mu.Lock()
	s.mixer.Add(st)
	s.mu.Unlock()
}

func (s *SDLSpeaker) Lock()   { s.mu.Lock() }
func (s *SDLSpeaker) Unlock() { s.mu.Unlock() }

func (s *SDLSpeaker) Close() {
	if s.dev != 0 {
		sdl.CloseAudioDevice(s.dev)
	}
}
