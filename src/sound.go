package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"sync"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"

	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/midi"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

const (
	audioOutLen          = 2048
	audioFrequency       = 44100
	audioPrecision       = 4
	audioResampleQuality = 1
)

// ------------------------------------------------------------------
// Normalizer

type Normalizer struct {
	streamer beep.Streamer
	mul      float64
	l, r     *NormalizerLR
}

func NewNormalizer(st beep.Streamer) *Normalizer {
	return &Normalizer{streamer: st, mul: 4,
		l: &NormalizerLR{1, 0, 1, 1 / 32.0, 0, 0},
		r: &NormalizerLR{1, 0, 1, 1 / 32.0, 0, 0}}
}

func (n *Normalizer) Stream(samples [][2]float64) (s int, ok bool) {
	// IDK how this happens, it just does after running for a really,
	// really long time and the below streamer.Stream method does not
	// do a nil check. This should at least prevent crashes, but may
	// lead to sound glitches.
	if n.streamer == nil || len(samples) <= 0 {
		return 0, false
	}
	s, ok = n.streamer.Stream(samples)
	wavVolumeMul := 0.5 * (float64(sys.cfg.Sound.WavVolume) * float64(sys.cfg.Sound.MasterVolume) * 0.0001)
	for i := range samples[:s] {
		if sys.cfg.Sound.AudioDucking {
			lmul := n.l.process(n.mul, &samples[i][0])
			rmul := n.r.process(n.mul, &samples[i][1])
			// AudioDucking controls the adaptive gain only. User-configured WAV/master volume must still be applied.
			samples[i][0] *= wavVolumeMul
			samples[i][1] *= wavVolumeMul
			n.mul = math.Min(16.0, math.Min(lmul, rmul))
		} else {
			// Apply the configured WAV/master volume directly through the normalizer path.
			n.l.process(wavVolumeMul, &samples[i][0])
			n.r.process(wavVolumeMul, &samples[i][1])
		}
	}
	return s, ok
}

func (n *Normalizer) Err() error {
	if n.streamer == nil {
		return nil
	}
	return n.streamer.Err()
}

type NormalizerLR struct {
	edge, edgeDelta, gain, average, bias, bias2 float64
}

func (n *NormalizerLR) process(mul float64, sam *float64) float64 {
	n.bias += (*sam - n.bias) / (float64(sys.cfg.Sound.SampleRate)/110.0 + 1)
	n.bias2 += (*sam - n.bias2) / (float64(sys.cfg.Sound.SampleRate)/112640.0 + 1)
	s := (n.bias2 - n.bias) * mul
	if math.Abs(s) > 1 {
		mul *= math.Pow(math.Abs(s), -n.edge)
		n.edgeDelta += 32 * (1 - n.edge) / float64(sys.cfg.Sound.SampleRate+32)
		s = math.Copysign(1.0, s)
	} else {
		tmp := (1 - math.Pow(1-math.Abs(s), 64)) * math.Pow(0.5-math.Abs(s), 3)
		mul += mul * (n.edge*(1/32.0-n.average)/n.gain + tmp*n.gain*(1-n.edge)/32) /
			(float64(sys.cfg.Sound.SampleRate)*2/8.0 + 1)
		n.edgeDelta -= (0.5 - n.average) * n.edge / (float64(sys.cfg.Sound.SampleRate) * 2)
	}
	n.gain += (1.0 - n.gain*(math.Abs(s)+1/32.0)) / (float64(sys.cfg.Sound.SampleRate) * 2)
	n.average += (math.Abs(s) - n.average) / (float64(sys.cfg.Sound.SampleRate) * 2)
	n.edge = float64(Clamp(float32(n.edge+n.edgeDelta), 0, 1))
	*sam = s
	return mul
}

// Safe wrapper for streamer operations
func WithSpeakerLock(f func()) {
	speaker.Lock()
	defer speaker.Unlock()
	f()
}

// ------------------------------------------------------------------
// SwapSeeker - beep.StreamSeeker that can be swapped to in-memory
// copy at runtime.
type SwapSeeker struct {
	mu sync.RWMutex      // read/write mutex
	ss beep.StreamSeeker // current source
}

func newSwapSeeker(ss beep.StreamSeeker) *SwapSeeker {
	return &SwapSeeker{ss: ss}
}

// Swap(next, absStart): next - new seeker
func (sw *SwapSeeker) Swap(next beep.StreamSeeker) {
	speaker.Lock()
	sw.mu.Lock()
	pos := sw.ss.Position()

	if ln := next.Len(); ln <= 0 { // guard against 0
		// empty buffer, don't try to swap this
		LogMessage("Empty BGM RAM swap buffer somehow. Aborting swap at the absolute last possible moment!")
		sw.mu.Unlock()
		speaker.Unlock()
		return
	}

	next.Seek(pos)
	sw.ss = next
	sw.mu.Unlock()
	speaker.Unlock()
}

func (sw *SwapSeeker) Stream(out [][2]float64) (int, bool) {
	sw.mu.RLock()
	ss := sw.ss
	sw.mu.RUnlock()
	return ss.Stream(out)
}

func (sw *SwapSeeker) Seek(p int) error {
	sw.mu.RLock()
	ss := sw.ss
	err := ss.Seek(p)
	sw.mu.RUnlock()
	return err
}

func (sw *SwapSeeker) Position() int {
	sw.mu.RLock()
	pos := sw.ss.Position()
	sw.mu.RUnlock()
	return pos
}

func (sw *SwapSeeker) Len() int {
	sw.mu.RLock()
	l := sw.ss.Len()
	sw.mu.RUnlock()
	return l
}

func (sw *SwapSeeker) Err() error {
	sw.mu.RLock()
	e := sw.ss.Err()
	sw.mu.RUnlock()
	return e
}

// ------------------------------------------------------------------
// BufferSeeker ‒ wraps *beep.Buffer and gives an in-memory
// StreamSeeker
type BufferSeeker struct {
	buf *beep.Buffer  // the decoded audio buffer
	pos int           // absolute position in samples
	str beep.Streamer // current streamer
}

func newBufferSeeker(buf *beep.Buffer) *BufferSeeker {
	return &BufferSeeker{
		buf: buf,
		str: buf.Streamer(0, buf.Len()),
	}
}

func (b *BufferSeeker) Stream(out [][2]float64) (n int, ok bool) {
	// Pull samples from the current streamer
	n, ok = b.str.Stream(out)
	b.pos += n
	return n, ok
}

func (b *BufferSeeker) Seek(p int) error {
	// Clamp p so we don't panic on bogus values
	if p < 0 {
		p = 0
	} else if p > b.buf.Len() {
		p = b.buf.Len()
	}
	b.pos = p
	// Reset streamer so the next Stream() call starts at p
	b.str = b.buf.Streamer(p, b.buf.Len())
	return nil
}

func (b *BufferSeeker) Position() int { return b.pos }

func (b *BufferSeeker) Len() int { return b.buf.Len() }

func (b *BufferSeeker) Err() error { return nil }

// ------------------------------------------------------------------
// Loop Streamer

// Based on Loop() from Beep package. It adds support for loop points.
type StreamLooper struct {
	s         beep.StreamSeeker
	loopcount int
	loopstart int
	loopend   int
	err       error
}

func newStreamLooper(s beep.StreamSeeker, loopcount, loopstart, loopend int) beep.StreamSeeker {
	sl := &StreamLooper{
		s:         s,
		loopcount: loopcount,
		loopstart: loopstart,
		loopend:   loopend,
	}
	if sl.loopstart < 0 || sl.loopstart >= s.Len() {
		sl.loopstart = 0
	}
	if sl.loopend <= sl.loopstart || sl.loopend >= s.Len() {
		sl.loopend = s.Len()
	}
	return sl
}

// Adapted from beep.Loop2 (for dynamic modification)
func (l *StreamLooper) Stream(samples [][2]float64) (n int, ok bool) {
	if l.err != nil {
		return 0, false
	}
	for len(samples) > 0 {
		toStream := len(samples)
		if l.loopcount != 0 {
			samplesUntilEnd := l.loopend - l.s.Position()
			if samplesUntilEnd <= 0 {
				// End of loop, reset the position and decrease the loop count.
				if l.loopcount > 0 {
					l.loopcount--
				}
				if err := l.s.Seek(l.loopstart); err != nil {
					l.err = err
					return n, true
				}
				continue
			}
			// Stream only up to the end of the loop.
			toStream = Min(samplesUntilEnd, toStream)
		}

		sn, sok := l.s.Stream(samples[:toStream])
		n += sn
		if sn < toStream || !sok {
			l.err = l.s.Err()
			return n, n > 0
		}
		samples = samples[sn:]
	}
	return n, true
}

func (b *StreamLooper) Err() error {
	return b.s.Err()
}

func (b *StreamLooper) Len() int {
	return b.s.Len()
}

func (b *StreamLooper) Position() int {
	return b.s.Position()
}

func (b *StreamLooper) Seek(p int) error {
	return b.s.Seek(p)
}

// ------------------------------------------------------------------
// Bgm

type Bgm struct {
	filename   string
	bgmVolume  int
	volRestore int
	loop       int
	streamer   beep.StreamSeeker
	ctrl       *beep.Ctrl
	volctrl    *effects.Volume
	format     string
	freqmul    float32
	sampleRate beep.SampleRate
	startPos   int
	mu         sync.Mutex
	cancel     context.CancelFunc
}

func newBgm() *Bgm {
	return &Bgm{}
}

func (bgm *Bgm) Stop() {
	if bgm.ctrl != nil {
		WithSpeakerLock(func() {
			bgm.ctrl.Streamer = nil
		})
	}
	bgm.filename = ""
}

func (bgm *Bgm) Open(filename string, loop, bgmVolume, bgmLoopStart, bgmLoopEnd, startPosition int, freqmul float32, loopcount int) {
	// Right away, cancel any running goroutines.
	bgm.mu.Lock()
	if bgm.cancel != nil {
		bgm.cancel()
	}
	var ctx context.Context
	ctx, bgm.cancel = context.WithCancel(context.Background())
	bgm.mu.Unlock()

	bgm.filename = filename
	bgm.loop = loop
	bgm.bgmVolume = bgmVolume
	bgm.freqmul = freqmul
	// Starve the current music streamer
	if bgm.ctrl != nil {
		WithSpeakerLock(func() {
			bgm.ctrl.Streamer = nil
		})
	}
	// Special value "" is used to stop music
	if filename == "" {
		return
	}

	f, err := OpenFile(bgm.filename)
	if err != nil {
		// sys.bgm = *newBgm() // removing this gets pause step playsnd to work correctly 100% of the time
		LogMessage("Failed to open BGM: %v", err)
		return
	}
	var format beep.Format
	if HasExtension(bgm.filename, ".ogg") {
		bgm.streamer, format, err = vorbis.Decode(f)
		bgm.format = "ogg"
	} else if HasExtension(bgm.filename, ".mp3") {
		bgm.streamer, format, err = mp3.Decode(f)
		bgm.format = "mp3"
	} else if HasExtension(bgm.filename, ".wav") {
		bgm.streamer, format, err = wav.Decode(f)
		bgm.format = "wav"
	} else if HasExtension(bgm.filename, ".flac") {
		bgm.streamer, format, err = flac.Decode(f)
		bgm.format = "flac"
	} else if HasExtension(bgm.filename, ".mid") || HasExtension(bgm.filename, ".midi") {
		if sf, sferr := loadSoundFont(sys.cfg.Sound.SoundFont); sferr != nil {
			err = sferr
		} else {
			bgm.streamer, format, err = midi.Decode(f, sf, beep.SampleRate(int(sys.cfg.Sound.SampleRate)))
			bgm.format = "midi"
		}
	} else if HasExtension(bgm.filename, ".xm") || HasExtension(bgm.filename, ".mod") || HasExtension(bgm.filename, ".it") || HasExtension(bgm.filename, ".s3m") {
		bgm.streamer, format, err = xmpDecode(f)
		bgm.format = "xmp"
	} else {
		err = Error(fmt.Sprintf("unsupported file extension: %v", bgm.filename))
	}
	if err != nil {
		f.Close()
		LogMessage("Failed to load bgm: %v\n%v", bgm.filename, err)
		return
	}
	lc := 0
	if loop != 0 {
		if loopcount >= 0 {
			lc = Max(0, loopcount-1)
		} else {
			lc = -1
		}

		// Clamp loop end
		if bgmLoopEnd <= 0 || bgmLoopEnd > bgm.streamer.Len() {
			bgmLoopEnd = bgm.streamer.Len()
		}
	}
	// Don't do anything if we have the nomusic command line flag
	if _, ok := sys.cmdFlags["-nomusic"]; ok {
		return
	}
	// Don't do anything if we have the nosound command line flag
	if _, ok := sys.cmdFlags["-nosound"]; ok {
		return
	}
	bgm.startPos = startPosition
	sw := newSwapSeeker(bgm.streamer)
	bgm.streamer = sw // fix pos not updating after swapping
	streamer := newStreamLooper(sw, lc, bgmLoopStart, bgmLoopEnd)
	// we're going to continue to use our own modified streamLooper because beep doesn't allow
	// negative values for loopcount (no forever case)
	bgm.volctrl = &effects.Volume{Streamer: streamer, Base: 2, Volume: 0, Silent: true}
	bgm.sampleRate = format.SampleRate
	dstFreq := beep.SampleRate(float32(sys.cfg.Sound.SampleRate) / bgm.freqmul)
	resampler := beep.Resample(audioResampleQuality, bgm.sampleRate, dstFreq, bgm.volctrl)
	bgm.ctrl = &beep.Ctrl{Streamer: resampler}
	bgm.volRestore = 0 // need this to prevent paused BGM volume from overwriting the new BGM volume
	bgm.UpdateVolume()
	bgm.streamer.Seek(startPosition)
	speaker.Play(bgm.ctrl)

	// Handle the RAM swap in the background (only for looped BGM and only if the user enabled it)
	if lc != 0 && sys.cfg.Sound.BGMRAMBuffer && bgm.format != "xmp" {
		// go func(ctx context.Context) {
		// Changing this to SafeGo should be safe because ctx has already been captured above
		SafeGo(func() {
			// Call the cancel function when the goroutine exits
			// to ensure cleanup.
			defer func() {
				bgm.mu.Lock()
				bgm.cancel()
				bgm.mu.Unlock()
			}()

			// We're gonna be using this a lot to check cancellation.
			isCancelled := func() bool {
				select {
				case <-ctx.Done():
					LogMessage("New BGM queued up - skipping RAM swap for this BGM")
					return true
				default:
					// Continue
					return false
				}
			}

			// New BGM was queued up before opening file, return NOW
			if isCancelled() {
				return
			}

			lf, err := OpenFile(bgm.filename)
			if err != nil {
				LogMessage(err.Error())
				return
			}

			// New BGM was queued up after opening file, close & return NOW
			if isCancelled() {
				lf.Close()
				return
			}

			// We gotta re-decode this crap again (but it won't matter since we're on a different thread)
			var dec beep.StreamSeeker
			switch bgm.format {
			case "ogg":
				dec, _, err = vorbis.Decode(lf)
			case "mp3":
				dec, _, err = mp3.Decode(lf)
			case "wav":
				dec, _, err = wav.Decode(lf)
			case "flac":
				dec, _, err = flac.Decode(lf)
			case "midi":
				sf, e := loadSoundFont(sys.cfg.Sound.SoundFont)
				if e != nil {
					LogMessage(e.Error())
					return
				}
				dec, _, err = midi.Decode(lf, sf, bgm.sampleRate)
			}
			if err != nil {
				LogMessage(err.Error())
				return
			}

			// build RAM buffer with loop span
			dec.Seek(0)
			buf := beep.NewBuffer(format)
			buf.Append(beep.Take(dec.Len(), dec))

			// close if this decoder supports it (all but MIDI)
			if c, ok := dec.(io.Closer); ok {
				c.Close()
			}
			lf.Close() // doing this out of paranoia, harmless to call

			// Now create the new bufferSeeker
			memSeeker := newBufferSeeker(buf)

			// just swap it now so we're not keeping threads open
			if memSeeker.Len() > 0 {
				// There can sometimes be a sample mismatch with re-decoding MIDI so this takes care of that
				if sl, ok := streamer.(*StreamLooper); ok {
					sl.loopend = Min(memSeeker.Len(), sl.loopend)
				}

				for {
					select {
					case <-ctx.Done():
						// Cancelled by a new call to bgm.Open()
						return
					default:
						sw.Swap(memSeeker)
						return
					}
				}
			}
		})
	}
}

func loadSoundFont(filename string) (*midi.SoundFont, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	soundfont, err := midi.NewSoundFont(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return soundfont, nil
}

func (bgm *Bgm) SetPaused(pause bool) {
	if bgm.ctrl == nil || bgm.ctrl.Paused == pause {
		return
	}
	WithSpeakerLock(func() {
		bgm.ctrl.Paused = pause
	})
}

func (bgm *Bgm) UpdateVolume() {
	if bgm.volctrl == nil {
		return
	}
	// TODO: Throw a debug warning if this triggers
	if bgm.bgmVolume > sys.cfg.Sound.MaxBGMVolume {
		LogMessage("WARNING: BGM volume set beyond expected range (value: %v). Clamped to MaxBgmVolume", bgm.bgmVolume)
		bgm.bgmVolume = sys.cfg.Sound.MaxBGMVolume
	}

	// NOTE: This is what we're going to do, no matter the complaints, because BGMVolume is handled differently
	// than WAV volume anyway.  We've had problems changing this in the past so it's best to keep it as-is.
	volume := -5 + float64(sys.cfg.Sound.BGMVolume)*0.06*(float64(sys.cfg.Sound.MasterVolume)/100)*(float64(bgm.bgmVolume)/100)

	// clamp to 1
	if volume >= 1 {
		volume = 1
	}
	silent := volume <= -5
	WithSpeakerLock(func() {
		bgm.volctrl.Volume = volume
		bgm.volctrl.Silent = silent
	})
}

func (bgm *Bgm) SetFreqMul(freqmul float32) {
	if bgm.freqmul != freqmul {
		if bgm.ctrl != nil {
			// Special case: freqmul == 0 pauses
			if freqmul == 0 {
				// Taken care of in system.go in tickSound
				bgm.freqmul = freqmul
				return
			}
			srcRate := bgm.sampleRate
			dstRate := beep.SampleRate(float32(sys.cfg.Sound.SampleRate) / freqmul)
			if resampler, ok := bgm.ctrl.Streamer.(*beep.Resampler); ok {
				WithSpeakerLock(func() {
					resampler.SetRatio(float64(srcRate) / float64(dstRate))
					bgm.freqmul = freqmul
				})
			}
		}
	}
}

// OpenFromStreamer wires an arbitrary Beep streamer (e.g. Reisen-backed audio)
// into the existing BGM path so the video BGM replaces/uses the same channel.
func (bgm *Bgm) OpenFromStreamer(stream beep.Streamer, srcSampleRate beep.SampleRate, bgmVolume int) {
	// Right away, cancel any running goroutines.
	bgm.mu.Lock()
	if bgm.cancel != nil {
		bgm.cancel()
	}
	var ctx context.Context
	ctx, bgm.cancel = context.WithCancel(context.Background())
	_ = ctx // reserved for future use (mirrors Open)
	bgm.mu.Unlock()

	bgm.filename = "<video-stream>"
	bgm.loop = 0
	bgm.bgmVolume = bgmVolume
	bgm.freqmul = 1

	// Starve the current music streamer
	if bgm.ctrl != nil {
		WithSpeakerLock(func() {
			bgm.ctrl.Streamer = nil
		})
	}
	// Honor CLI flags just like normal Open()
	if _, ok := sys.cmdFlags["-nomusic"]; ok {
		return
	}
	if _, ok := sys.cmdFlags["-nosound"]; ok {
		return
	}

	// Build the standard BGM chain: Volume -> Resample -> Ctrl -> Mixer
	bgm.sampleRate = srcSampleRate
	bgm.volctrl = &effects.Volume{Streamer: stream, Base: 2, Volume: 0, Silent: true}
	dstFreq := beep.SampleRate(float32(sys.cfg.Sound.SampleRate) / bgm.freqmul)
	resampler := beep.Resample(audioResampleQuality, bgm.sampleRate, dstFreq, bgm.volctrl)
	bgm.ctrl = &beep.Ctrl{Streamer: resampler}
	bgm.volRestore = 0
	bgm.UpdateVolume()
	speaker.Play(bgm.ctrl)
}

func (bgm *Bgm) SetLoopPoints(bgmLoopStart int, bgmLoopEnd int) {
	if sl, ok := bgm.volctrl.Streamer.(*StreamLooper); ok {
		if sl.loopstart != bgmLoopStart && sl.loopend != bgmLoopEnd {
			// Set both at once, why not
			WithSpeakerLock(func() {
				sl.loopstart = bgmLoopStart
				sl.loopend = bgmLoopEnd
			})
		} else {
			// Set one at a time
			if sl.loopstart != bgmLoopStart {
				WithSpeakerLock(func() {
					sl.loopstart = bgmLoopStart
				})
			} else if sl.loopend != bgmLoopEnd {
				WithSpeakerLock(func() {
					sl.loopend = bgmLoopEnd
				})
			}
		}
	}
}

func (bgm *Bgm) Seek(positionSample int) {
	// For stream-only sources (e.g., video audio) we don't support seeking; ignore safely.
	if bgm.streamer == nil {
		return
	}
	// Reset to 0 if out of range
	WithSpeakerLock(func() {
		if positionSample < 0 || positionSample > bgm.streamer.Len() {
			positionSample = 0
		}
		_ = bgm.streamer.Seek(positionSample)
	})
}

// ------------------------------------------------------------------
// Sound

type Sound struct {
	wavData []byte
	format  beep.Format
	length  int
}

func readSound(f io.ReadSeekCloser, size uint32) (*Sound, error) {
	if size < 128 {
		return nil, fmt.Errorf("wav size is too small")
	}
	wavData := make([]byte, size)
	if _, err := f.Read(wavData); err != nil {
		return nil, err
	}
	// Decode the sound at least once, so that we know the format is OK
	s, wavfmt, err := wav.Decode(bytes.NewReader(wavData))
	if err != nil {
		return nil, err
	}
	// Check if the file can be fully played
	// Run a decode test and catch any panics.
	var recovered interface{}
	func() {
		defer func() {
			// Catch any panic that occurs inside this anonymous function.
			if r := recover(); r != nil {
				recovered = r
			}
		}()

		// Try streaming until the end of the file.
		var samples [512][2]float64
		for {
			n, ok := s.Stream(samples[:])
			if n == 0 || !ok {
				// When the end is reached normally.
				if s.Err() == nil && s.Position() >= s.Len() {
					break
				}
				// Other errors.
				if s.Err() != nil {
					// Errors not caught by recover() are stored here.
					recovered = s.Err()
				}
				break
			}
		}
	}()
	// If a panic was caught.
	if recovered != nil {
		return nil, nil // If sound wasn't able to be fully played, we disable it to avoid engine freezing
	}
	return &Sound{wavData, wavfmt, s.Len()}, nil
}

func (s *Sound) GetStreamer() beep.StreamSeeker {
	streamer, _, _ := wav.Decode(bytes.NewReader(s.wavData))
	return streamer
}

// ------------------------------------------------------------------
// Snd

type Snd struct {
	table     map[[2]int32]*Sound
	ver, ver2 uint16
	filename  string
}

func newSnd() *Snd {
	return &Snd{table: make(map[[2]int32]*Sound)}
}

// Try to reuse a sound file if it's already loaded somewhere else
// This doesn't impact loading times as much as SFF, but it saves memory without any work
func findActiveSnd(filename string) *Snd {
	// Check characters
	for i := range sys.cgi {
		if sys.cgi[i].snd != nil && sys.cgi[i].snd.filename == filename {
			return sys.cgi[i].snd
		}
	}

	// Check common FX
	for _, ffx := range sys.ffx {
		if ffx != nil && ffx.snd != nil && ffx.snd.filename == filename {
			return ffx.snd
		}
	}

	return nil
}

func LoadSnd(filename string) (*Snd, error) {
	if s := findActiveSnd(filename); s != nil {
		return s, nil
	}

	s, err := LoadSndFiltered(filename, func(gn [2]int32) bool { return gn[0] >= 0 && gn[1] >= 0 }, 0)
	if err != nil {
		return nil, Error(fmt.Sprintf("LoadSnd failed: %v\n%v", filename, err))
	}

	s.filename = filename
	return s, nil
}

// Parse a .snd file and return an Snd structure with its contents
// The "keepItem" function allows to filter out unwanted waves.
// If max > 0, the function returns immediately when a matching entry is found. It also gives up after "max" non-matching entries.
func LoadSndFiltered(filename string, keepItem func([2]int32) bool, max uint32) (*Snd, error) {
	s := newSnd()
	f, err := OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer func() { chk(f.Close()) }()
	buf := make([]byte, 12)
	var n int
	if n, err = f.Read(buf); err != nil {
		return nil, err
	}
	if string(buf[:n]) != "ElecbyteSnd\x00" {
		return nil, Error("Unrecognized SND file, invalid header")
	}
	read := func(x interface{}) error {
		return binary.Read(f, binary.LittleEndian, x)
	}
	if err := read(&s.ver); err != nil {
		return nil, err
	}
	if err := read(&s.ver2); err != nil {
		return nil, err
	}
	var numberOfSounds uint32
	if err := read(&numberOfSounds); err != nil {
		return nil, err
	}
	var subHeaderOffset uint32
	if err := read(&subHeaderOffset); err != nil {
		return nil, err
	}
	loops := numberOfSounds
	if max > 0 && max < numberOfSounds {
		loops = max
	}
	for i := uint32(0); i < loops; i++ {
		f.Seek(int64(subHeaderOffset), 0)
		var nextSubHeaderOffset uint32
		if err := read(&nextSubHeaderOffset); err != nil {
			return nil, err
		}
		var subFileLength uint32
		if err := read(&subFileLength); err != nil {
			return nil, err
		}
		var num [2]int32
		if err := read(&num); err != nil {
			return nil, err
		}
		if keepItem(num) {
			_, exists := s.table[num]
			if exists {
				LogMessage("WARNING: Duplicate sound key in %v: %v,%v (index %v ignored)", filename, num[0], num[1], i)
			} else {
				tmp, err := readSound(f, subFileLength)
				if err != nil {
					LogMessage("Sound %v,%v in %v can't be read: %v", num[0], num[1], filename, err)
					if max > 0 {
						return nil, err
					}
				} else {
					// Sound is corrupted and can't be played, so we export a warning message to the console
					if tmp == nil {
						sys.appendToConsole(fmt.Sprintf("WARNING: %v sound %v,%v is corrupted and can't be played, so it was disabled", filename, num[0], num[1]))
					}
					s.table[num] = tmp
					if max > 0 {
						break
					}
				}
			}
		}
		subHeaderOffset = nextSubHeaderOffset
	}
	return s, nil
}

func (s *Snd) Get(gn [2]int32) *Sound {
	return s.table[gn]
}

func (s *Snd) play(gn [2]int32, volumescale int32, pan float32, loopstart, loopend, startposition int) bool {
	sound := s.Get(gn)
	return sys.soundChannels.Play(sound, gn[0], gn[1], volumescale, pan, loopstart, loopend, startposition)
}

func (s *Snd) stop(gn [2]int32) {
	sound := s.Get(gn)
	sys.soundChannels.Stop(sound)
}

func loadFromSnd(filename string, g, s int32, max uint32) (*Sound, error) {
	// Load the snd file
	snd, err := LoadSndFiltered(filename, func(gn [2]int32) bool { return gn[0] == g && gn[1] == s }, max)
	if err != nil {
		return nil, err
	}
	tmp, ok := snd.table[[2]int32{g, s}]
	if !ok {
		return nil, nil
	}
	return tmp, nil
}

// ------------------------------------------------------------------
// SoundEffect (handles volume and panning)

type SoundEffect struct {
	streamer beep.Streamer
	volume   float32
	localscl float32
	pan      float32
	x        *float32
	priority int32
	loop     int32
	freqmul  float32
	startPos int
}

func (s *SoundEffect) Stream(samples [][2]float64) (n int, ok bool) {
	// TODO: Test mugen panning in relation to PanningWidth and zoom settings
	lv, rv := s.volume, s.volume
	if sys.cfg.Sound.StereoEffects && (s.x != nil || s.pan != 0) {
		var r float32
		if s.x != nil { // pan
			r = ((sys.xmax - s.localscl**s.x) - s.pan) / (sys.xmax - sys.xmin)
		} else { // abspan
			r = ((sys.xmax-sys.xmin)/2 - s.pan) / (sys.xmax - sys.xmin)
		}
		sc := sys.cfg.Sound.PanningRange / 100
		of := (100 - sys.cfg.Sound.PanningRange) / 200
		lv = Clamp(s.volume*2*(r*sc+of), 0, 512)
		rv = Clamp(s.volume*2*((1-r)*sc+of), 0, 512)
	}

	n, ok = s.streamer.Stream(samples)
	for i := range samples[:n] {
		samples[i][0] *= float64(lv / 256)
		samples[i][1] *= float64(rv / 256)
	}
	return n, ok
}

func (s *SoundEffect) Err() error {
	return s.streamer.Err()
}

// ------------------------------------------------------------------
// SoundChannel

type SoundChannel struct {
	streamer          beep.StreamSeeker
	sfx               *SoundEffect
	ctrl              *beep.Ctrl
	sound             *Sound
	playerID          int32
	channelNo         int32 // Logical channel assigned by char code
	stopOnGetHit      bool
	stopOnChangeState bool
	group             int32
	number            int32
	timeStamp         int32
	volResume         float32 // For pausing/unpausing
}

// The old Stop() plus more
func (s *SoundChannel) Reset() {
	if s.ctrl != nil {
		WithSpeakerLock(func() {
			s.ctrl.Streamer = nil
		})
	}

	s.streamer = nil
	s.sfx = nil
	s.ctrl = nil
	s.sound = nil

	s.playerID = -1
	s.channelNo = -1
	s.stopOnGetHit = false
	s.stopOnChangeState = false
	s.group = -1
	s.number = -1
	s.timeStamp = -1
}

func (s *SoundChannel) Play(sound *Sound, group, number, loop int32, freqmul float32, loopStart, loopEnd, startPosition int) {
	if sound == nil {
		return
	}

	s.sound = sound
	s.group = group
	s.number = number
	s.timeStamp = sys.gameTime()
	s.streamer = s.sound.GetStreamer()

	loopCount := int(0)
	if loop < 0 {
		loopCount = -1
	} else {
		loopCount = Max(0, int(loop-1))
	}

	// going to continue using our streamLooper which is now modified from beep.Loop2
	looper := newStreamLooper(s.streamer, loopCount, loopStart, loopEnd)
	s.sfx = &SoundEffect{streamer: looper, volume: 256, priority: 0, loop: int32(loopCount), freqmul: freqmul, startPos: startPosition}
	srcRate := s.sound.format.SampleRate
	dstRate := beep.SampleRate(float32(sys.cfg.Sound.SampleRate) / s.sfx.freqmul)
	resampler := beep.Resample(audioResampleQuality, srcRate, dstRate, s.sfx)
	s.ctrl = &beep.Ctrl{Streamer: resampler}
	s.streamer.Seek(startPosition)

	WithSpeakerLock(func() {
		sys.soundMixer.Add(s.ctrl)
	})
}

func (s *SoundChannel) IsPlaying() bool {
	return s.sound != nil
}

func (s *SoundChannel) SetPaused(pause bool) {
	if s.ctrl == nil || s.ctrl.Paused == pause {
		return
	}
	WithSpeakerLock(func() {
		s.ctrl.Paused = pause
	})
}

func (s *SoundChannel) SetVolume(vol float32) {
	if s.ctrl != nil {
		s.sfx.volume = Clamp(vol, 0, 512)
	}
}

func (s *SoundChannel) SetPan(p, ls float32, x *float32) {
	if s.ctrl != nil {
		s.sfx.localscl = ls
		s.sfx.x = x
		s.sfx.pan = p * ls
	}
}

func (s *SoundChannel) SetPriority(priority int32) {
	if s.ctrl != nil {
		s.sfx.priority = priority
	}
}

func (s *SoundChannel) SetFreqMul(freqmul float32) {
	if s.ctrl != nil {
		if s.sound != nil {
			// Special case: freqmul == 0 pauses
			if freqmul == 0 {
				s.sfx.freqmul = freqmul
				s.SetPaused(true)
				return
			}
			srcRate := s.sound.format.SampleRate
			dstRate := beep.SampleRate(float32(sys.cfg.Sound.SampleRate) / freqmul)
			if resampler, ok := s.ctrl.Streamer.(*beep.Resampler); ok {
				WithSpeakerLock(func() {
					resampler.SetRatio(float64(srcRate) / float64(dstRate))
					s.sfx.freqmul = freqmul
				})
			}
		}
	}
}

func (s *SoundChannel) SetLoopPoints(loopstart, loopend int) {
	// Set both at once, why not
	if sl, ok := s.sfx.streamer.(*StreamLooper); ok {
		if sl.loopstart != loopstart && sl.loopend != loopend {
			WithSpeakerLock(func() {
				sl.loopstart = loopstart
				sl.loopend = loopend
			})
			// Set one at a time
		} else {
			if sl.loopstart != loopstart {
				WithSpeakerLock(func() {
					sl.loopstart = loopstart
				})
			} else if sl.loopend != loopend {
				WithSpeakerLock(func() {
					sl.loopend = loopend
				})
			}
		}
	}
}

// ------------------------------------------------------------------
// SoundChannels (collection of prioritised sound channels)

type SoundChannels []SoundChannel

/*
func newSoundChannels(size int32) *SoundChannels {
	s := &SoundChannels{}
	s.SetSize(size)
	return s
}
*/

func (s *SoundChannels) SetSize(size int32) {
	currentSize := int32(len(*s))

	switch {
	case size > currentSize:
		// Add new channels
		newSlotsCount := size - currentSize
		newSlots := make([]SoundChannel, newSlotsCount)

		// Initialize the new slots
		for i := range newSlots {
			newSlots[i].Reset()
		}

		*s = append(*s, newSlots...)

	case size < currentSize:
		// Remove channels safely
		for i := currentSize - 1; i >= size; i-- {
			(*s)[i].Reset()
		}

		*s = (*s)[:size]
	}
}

// Returns a channel with the requested logical ID
func (s *SoundChannels) Request(pid int32, chNo int32, lowpriority bool, priority int32) *SoundChannel {
	// Ensure capacity
	if len(*s) < int(sys.cfg.Sound.WavChannels) {
		s.SetSize(sys.cfg.Sound.WavChannels)
	}

	// Normalize channel number
	// Since channels are just ID's there's no practical reason to have this limit, but Mugen does it
	if chNo < 0 || chNo >= sys.cfg.Sound.WavChannels {
		chNo = -1
	}

	// Specific channel request
	// Try to use the same index if it's active
	if chNo >= 0 {
		for i := range *s {
			ch := &(*s)[i]
			if ch.playerID == pid && ch.channelNo == chNo {
				if ch.IsPlaying() {
					if lowpriority || ch.sfx != nil && priority < ch.sfx.priority {
						return nil
					}
					ch.Reset()
				}
				ch.playerID = pid
				ch.channelNo = chNo
				return ch
			}
		}
	}

	// If same channel was not found or if channel was not specified
	// Look for any free index
	for i := range *s {
		ch := &(*s)[i]
		if !ch.IsPlaying() {
			ch.playerID = pid
			ch.channelNo = chNo
			return ch
		}
	}

	// If "lowpriority" we don't even need to run the replacement code
	if lowpriority {
		return nil
	}

	// All channels full
	// Replace the oldest sound. Negative channels are more likely to be replaced
	var oldestNegativeIdx int = -1
	var minTimeNegative int32 = math.MaxInt32
	var oldestPositiveIdx int = -1
	var minTimePositive int32 = math.MaxInt32

	for i := range *s {
		ch := &(*s)[i]

		// We still take priority into account here
		if ch.IsPlaying() && ch.sfx != nil && priority < ch.sfx.priority {
			continue
		}

		if ch.channelNo < 0 {
			if ch.timeStamp < minTimeNegative {
				minTimeNegative = ch.timeStamp
				oldestNegativeIdx = i
			}
		} else {
			if ch.timeStamp < minTimePositive {
				minTimePositive = ch.timeStamp
				oldestPositiveIdx = i
			}
		}
	}

	// Kick out the oldest negative channel first
	if oldestNegativeIdx != -1 {
		ch := &(*s)[oldestNegativeIdx]
		ch.Reset()
		ch.playerID = pid
		ch.channelNo = chNo
		return ch
	}

	// If no negative channels can be evicted, try the oldest positive one
	if oldestPositiveIdx != -1 {
		ch := &(*s)[oldestPositiveIdx]
		ch.Reset()
		ch.playerID = pid
		ch.channelNo = chNo
		return ch
	}

	return nil
}

/*
// This is just a simplified Request(), so we will call that one instead
func (s *SoundChannels) reserveChannel() *SoundChannel {
	for i := range s.channels {
		if !s.channels[i].IsPlaying() {
			return &s.channels[i]
		}
	}
	return nil
}
*/

func (s *SoundChannels) Get(pid int32, ch int32) *SoundChannel {
	// Dereference once
	channels := *s

	// Only consider channels with a specific ID
	if ch >= 0 && ch < int32(len(channels)) {
		for i := range channels {
			// Match logical channel number
			if channels[i].channelNo == ch {
				// Match player ID. System sounds use a negative ID
				if (pid < 0 && channels[i].playerID < 0) || (pid >= 0 && channels[i].playerID == pid) {
					if channels[i].IsPlaying() {
						return &channels[i]
					}
				}
			}
		}
		//return &channels[ch]
	}
	return nil
}

func (s *SoundChannels) Play(sound *Sound, group, number, volumescale int32, pan float32, loopStart, loopEnd, startPosition int) bool {
	if sound == nil {
		return false
	}
	c := s.Request(-1, -1, false, 0)
	if c == nil {
		return false
	}
	c.Play(sound, group, number, 0, 1.0, loopStart, loopEnd, startPosition)
	c.SetVolume(float32(volumescale * 64 / 25))
	c.SetPan(pan, 0, nil)
	return true
}

// Checks if a specific sound is playing in any channel
func (s *SoundChannels) IsPlaying(sound *Sound) bool {
	channels := *s
	for i := range channels {
		v := &channels[i]
		if v.sound != nil && v.sound == sound {
			if v.IsPlaying() {
				return true
			}
		}
	}
	return false
}

func (s *SoundChannels) Stop(sound *Sound) {
	channels := *s
	for i := range channels {
		v := &channels[i]
		if v.sound != nil && v.sound == sound {
			v.Reset()
		}
	}
}

func (s *SoundChannels) StopAll() {
	channels := *s
	for i := range channels {
		if channels[i].sound != nil {
			channels[i].Reset()
		}
	}
}

func (s *SoundChannels) Tick() {
	channels := *s
	for i := range channels {
		v := &channels[i]
		if v.IsPlaying() {
			if v.streamer.Position() >= v.sound.length && v.sfx.loop != -1 { // End of sound
				v.Reset()
			}
		}
	}
}
