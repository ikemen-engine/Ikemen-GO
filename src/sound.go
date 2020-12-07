package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/Windblade-GR01/go-openal/openal"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"

	//"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

const (
	audioOutLen    = 2048
	audioFrequency = 48000
)

// ------------------------------------------------------------------
// Audio Source

// AudioSource structure.
// It contains OpenAl's sound destination and buffer
type AudioSource struct {
	Src  openal.Source
	bufs openal.Buffers
}

func NewAudioSource() (s *AudioSource) {
	s = &AudioSource{Src: openal.NewSource(), bufs: openal.NewBuffers(2)}
	for i := range s.bufs {
		s.bufs[i].SetDataInt16(openal.FormatStereo16, sys.nullSndBuf[:],
			audioFrequency)
	}
	s.Src.QueueBuffers(s.bufs)
	if err := openal.Err(); err != nil {
		println(err.Error())
	}
	return
}
func (s *AudioSource) Delete() {
	for s.Src.BuffersQueued() > 0 {
		s.Src.UnqueueBuffer()
	}
	s.bufs.Delete()
	s.Src.Delete()
}

// ------------------------------------------------------------------
// Mixer

type Mixer struct {
	buf        [audioOutLen * 2]float32
	sendBuf    []int16
	out        chan []int16
	normalizer *Normalizer
}

func newMixer() *Mixer {
	return &Mixer{out: make(chan []int16, 1), normalizer: NewNormalizer()}
}
func (m *Mixer) bufClear() {
	for i := range m.buf {
		m.buf[i] = 0
	}
}
func (m *Mixer) write() bool {
	if m.sendBuf == nil {
		m.sendBuf = make([]int16, len(m.buf))
		for i := 0; i <= len(m.sendBuf)-2; i += 2 {
			l, r := m.normalizer.Process(m.buf[i], m.buf[i+1])
			m.sendBuf[i] = int16(32767 * l)
			m.sendBuf[i+1] = int16(32767 * r)
		}
	}
	select {
	case m.out <- m.sendBuf:
	default:
		return false
	}
	m.sendBuf = nil
	m.bufClear()
	return true
}
func (m *Mixer) Mix(wav []byte, fidx float64, bytesPerSample, channels int,
	sampleRate float64, loop bool, volume float32) float64 {
	fidxadd := sampleRate / audioFrequency
	if fidxadd > 0 {
		switch bytesPerSample {
		case 1:
			switch channels {
			case 1:
				for i := 0; i <= len(m.buf)-2; i += 2 {
					iidx := int(fidx)
					if iidx >= len(wav) {
						if !loop {
							break
						}
						iidx, fidx = 0, 0
					}
					sam := volume * (float32(wav[iidx]) - 128) / 128
					m.buf[i] += sam
					m.buf[i+1] += sam
					fidx += fidxadd
				}
				return fidx
			case 2:
				for i := 0; i <= len(m.buf)-2; i += 2 {
					iidx := 2 * int(fidx)
					if iidx > len(wav)-2 {
						if !loop {
							break
						}
						iidx, fidx = 0, 0
					}
					m.buf[i] += volume * (float32(wav[iidx]) - 128) / 128
					m.buf[i+1] += volume * (float32(wav[iidx+1]) - 128) / 128
					fidx += fidxadd
				}
				return fidx
			}
		case 2:
			switch channels {
			case 1:
				for i := 0; i <= len(m.buf)-2; i += 2 {
					iidx := 2 * int(fidx)
					if iidx > len(wav)-2 {
						if !loop {
							break
						}
						iidx, fidx = 0, 0
					}
					sam := volume *
						float32(int(wav[iidx])|int(int8(wav[iidx+1]))<<8) / (1 << 15)
					m.buf[i] += sam
					m.buf[i+1] += sam
					fidx += fidxadd
				}
				return fidx
			case 2:
				for i := 0; i <= len(m.buf)-2; i += 2 {
					iidx := 4 * int(fidx)
					if iidx > len(wav)-4 {
						if !loop {
							break
						}
						iidx, fidx = 0, 0
					}
					m.buf[i] += volume *
						float32(int(wav[iidx])|int(int8(wav[iidx+1]))<<8) / (1 << 15)
					m.buf[i+1] += volume *
						float32(int(wav[iidx+2])|int(int8(wav[iidx+3]))<<8) / (1 << 15)
					fidx += fidxadd
				}
				return fidx
			}
		}
	}
	return float64(len(wav))
}

// ------------------------------------------------------------------
// Normalizer

type Normalizer struct {
	mul  float64
	l, r *NormalizerLR
}

func NewNormalizer() *Normalizer {
	return &Normalizer{mul: 4, l: &NormalizerLR{1, 0, 1, 1 / 32.0, 0, 0},
		r: &NormalizerLR{1, 0, 1, 1 / 32.0, 0, 0}}
}
func (n *Normalizer) Process(l, r float32) (float32, float32) {
	lmul := n.l.process(n.mul, &l)
	rmul := n.r.process(n.mul, &r)
	if sys.audioDucking {
		if lmul < rmul {
			n.mul = lmul
		} else {
			n.mul = rmul
		}
		if n.mul > 16 {
			n.mul = 16
		}
	} else {
		n.mul = 0.5 * (float64(sys.wavVolume) * float64(sys.masterVolume) * 0.0001)
	}
	return l, r
}

type NormalizerLR struct {
	heri, herihenka, fue, heikin, katayori, katayori2 float64
}

func (n *NormalizerLR) process(bai float64, sam *float32) float64 {
	n.katayori = (n.katayori*audioFrequency/110 + float64(*sam)) /
		(audioFrequency/110.0 + 1)
	n.katayori2 = (n.katayori2*audioFrequency/112640 + float64(*sam)) /
		(audioFrequency/112640.0 + 1)
	s := (n.katayori2 - n.katayori) * bai
	if math.Abs(s) > 1 {
		bai *= math.Pow(1/math.Abs(s), n.heri)
		n.herihenka += 32 * (1 - n.heri) / float64(audioFrequency+32)
		if s < 0 {
			s = -1
		} else {
			s = 1
		}
	} else {
		tmp := (1 - math.Pow(1-math.Abs(s), 64)) * math.Pow(0.5-math.Abs(s), 3)
		bai += bai * (n.heri*(1/32.0-n.heikin)/n.fue + tmp*n.fue*(1-n.heri)/32) /
			(audioFrequency*2/8.0 + 1)
		n.herihenka -= (0.5 - n.heikin) * n.heri / (audioFrequency * 2)
	}
	n.fue += (32*n.fue*(1/n.fue-math.Abs(s)) - n.fue) /
		(32 * audioFrequency * 2)
	n.heikin += (math.Abs(s) - n.heikin) / (audioFrequency * 2)
	n.heri += n.herihenka
	if n.heri < 0 {
		n.heri = 0
	} else if n.heri > 0 {
		n.heri = 1
	}
	*sam = float32(s)
	return bai
}

// ------------------------------------------------------------------
// Bgm

type Bgm struct {
	filename            string
	bgmVolume           int
	bgmLoopStart        int
	bgmLoopEnd          int
	defaultFilename     string
	defaultBgmVolume    int
	defaultbgmLoopStart int
	defaultbgmLoopEnd   int
	loop                int
	// TODO: Use this.
	//sampleRate          beep.SampleRate
	streamer  beep.StreamSeekCloser
	ctrl      *beep.Ctrl
	resampler *beep.Resampler
	volume    *effects.Volume
	format    string
}

func newBgm() *Bgm {
	return &Bgm{}
}

func (bgm *Bgm) Open(filename string, isDefaultBGM bool, loop, bgmVolume, bgmLoopStart, bgmLoopEnd int) {
	if filepath.Base(bgm.filename) == filepath.Base(filename) {
		return
	}
	bgm.filename = filename
	bgm.loop = loop
	bgm.bgmVolume = bgmVolume
	bgm.bgmLoopStart = bgmLoopStart
	bgm.bgmLoopEnd = bgmLoopEnd
	if isDefaultBGM {
		bgm.defaultFilename = filename
		bgm.defaultBgmVolume = bgmVolume
		bgm.defaultbgmLoopStart = bgmLoopStart
		bgm.defaultbgmLoopEnd = bgmLoopEnd
	}
	speaker.Clear()

	if HasExtension(bgm.filename, "^\\.[Oo][Gg][Gg]$") {
		bgm.ReadVorbis(loop, bgmVolume)
	} else if HasExtension(bgm.filename, "^\\.[Mm][Pp]3$") {
		bgm.ReadMp3(loop, bgmVolume)
		//} else if HasExtension(bgm.filename, "^\\.[Ff][Ll][Aa][Cc]$") {
		//	bgm.ConvertFLAC(loop, bgmVolume)
	} else if HasExtension(bgm.filename, "^\\.[Ww][Aa][Vv]$") {
		bgm.ReadWav(loop, bgmVolume)
	}
}

func (bgm *Bgm) ReadMp3(loop int, bgmVolume int) {
	f, _ := os.Open(bgm.filename)
	s, format, err := mp3.Decode(f)
	bgm.streamer = s
	bgm.format = "mp3"
	if err != nil {
		return
	}
	bgm.ReadFormat(format, loop, bgmVolume)
}

/*
// TODO: Now that we are using modules this should work again if we configure it correctly.
func (bgm *Bgm) ReadFLAC(loop int, bgmVolume int) {
	f, _ := os.Open(bgm.filename)
	s, format, err := flac.Decode(f)
	bgm.streamer = s
	bgm.format = "flac"

	if err != nil {
		return
	}
	bgm.ReadFormat(format, loop, bgmVolume)
}

// SCREW THE FLAC.SEEK FUNCTION, IT DOES NOT WORK SO WE ARE GOING TO CONVERT THIS TO WAV
// Update: Now the flac dependecy broke. (-_-)
func (bgm *Bgm) ConvertFLAC(loop int, bgmVolume int) {
	// We open the flac
	f1, _ := os.Open(bgm.filename)
	// And create a temp one
	f2, _ := os.Create("save/tempaudio.wav")

	// Open decode and convert
	s, format, err := flac.Decode(f1)
	wav.Encode(f2, s, format)

	bgm.filename = "save/tempaudio.wav"
	//bgm.tempfile = f2
	bgm.format = "flac"

	s.Close()

	if err != nil {
		return
	}

	sys.FLAC_FrameWait = 120
}
*/

func (bgm *Bgm) PlayMemAudio(loop int, bgmVolume int) {
	f, _ := os.Open(bgm.filename)
	s, format, err := wav.Decode(f)
	bgm.streamer = s
	if err != nil {
		return
	}
	bgm.ReadFormat(format, loop, bgmVolume)
}

func (bgm *Bgm) ReadVorbis(loop int, bgmVolume int) {
	f, _ := os.Open(bgm.filename)
	s, format, err := vorbis.Decode(f)
	bgm.streamer = s
	bgm.format = "ogg"
	if err != nil {
		return
	}
	bgm.ReadFormat(format, loop, bgmVolume)
}

func (bgm *Bgm) ReadWav(loop int, bgmVolume int) {
	f, _ := os.Open(bgm.filename)
	s, format, err := wav.Decode(f)
	bgm.streamer = s
	bgm.format = "wav"
	if err != nil {
		return
	}
	bgm.ReadFormat(format, loop, bgmVolume)
}

func (bgm *Bgm) ReadFormat(format beep.Format, loop int, bgmVolume int) {
	loopCount := int(1)
	if loop > 0 {
		loopCount = -1
	}
	streamer := beep.Loop(loopCount, bgm.streamer)
	volume := -5 + float64(sys.bgmVolume)*0.06*(float64(sys.masterVolume)/100)*(float64(bgmVolume)/100)
	bgm.volume = &effects.Volume{Streamer: streamer, Base: 2, Volume: volume, Silent: volume <= -5}
	bgm.resampler = beep.Resample(int(3), format.SampleRate, beep.SampleRate(Mp3SampleRate), bgm.volume)
	bgm.ctrl = &beep.Ctrl{Streamer: bgm.resampler}
	speaker.Play(bgm.ctrl)
}

func (bgm *Bgm) Pause() {
	speaker.Lock()
	bgm.ctrl.Paused = true
	speaker.Unlock()
}

// ------------------------------------------------------------------
// Wave

type Wave struct {
	SamplesPerSec  uint32
	Channels       uint16
	BytesPerSample uint16
	Wav            []byte
}

func ReadWave(f *os.File, ofs int64) (*Wave, error) {
	buf := make([]byte, 4)
	n, err := f.Read(buf)
	if err != nil {
		return nil, err
	}
	if string(buf[:n]) != "RIFF" {
		return nil, Error("Not RIFF")
	}
	read := func(x interface{}) error {
		return binary.Read(f, binary.LittleEndian, x)
	}
	var riffSize uint32
	if err := read(&riffSize); err != nil {
		return nil, err
	}
	riffSize += 8
	if n, err = f.Read(buf); err != nil {
		return nil, err
	}
	if string(buf[:n]) != "WAVE" {
		return &Wave{SamplesPerSec: 11025, Channels: 1, BytesPerSample: 1}, nil
	}
	fmtSize, dataSize := uint32(0), uint32(0)
	w := Wave{}
	riffend := ofs + 16 + int64(riffSize)
	ofs += 28
	for (fmtSize == 0 || dataSize == 0) && ofs < riffend {
		if n, err = f.Read(buf); err != nil {
			return nil, err
		}
		var size uint32
		if err := read(&size); err != nil {
			return nil, err
		}
		switch string(buf[:n]) {
		case "fmt ":
			fmtSize = size
			var fmtID uint16
			if err := read(&fmtID); err != nil {
				return nil, err
			}
			if fmtID != 1 {
				return nil, Error("Not a linear PCM")
			}
			if err := read(&w.Channels); err != nil {
				return nil, err
			}
			if w.Channels < 1 || w.Channels > 2 {
				return nil, Error(fmt.Sprintf("Invalid number of channels: %d", w.Channels))
			}
			if err := read(&w.SamplesPerSec); err != nil {
				return nil, err
			}
			if w.SamplesPerSec < 1 || w.SamplesPerSec >= 0xfffff {
				return nil, Error(fmt.Sprintf("Invalid frequency: %d", w.SamplesPerSec))
			}
			var musi uint32
			if err := read(&musi); err != nil {
				return nil, err
			}
			var mushi uint16
			if err := read(&mushi); err != nil {
				return nil, err
			}
			if err := read(&w.BytesPerSample); err != nil {
				return nil, err
			}
			if w.BytesPerSample != 8 && w.BytesPerSample != 16 {
				return nil, Error(fmt.Sprintf("Invalid bit number: %d", w.BytesPerSample))
			}
			w.BytesPerSample >>= 3
		case "data":
			dataSize = size
			w.Wav = make([]byte, dataSize)
			if err := binary.Read(f, binary.LittleEndian, w.Wav); err != nil {
				return nil, err
			}
		}
		ofs += int64(size) + 8
		f.Seek(ofs, 0)
	}
	if fmtSize == 0 {
		if dataSize > 0 {
			return nil, Error("fmt is missing")
		}
		return nil, nil
	}
	return &w, nil
}

// ------------------------------------------------------------------
// Snd

type Snd struct {
	table     map[[2]int32]*Wave
	ver, ver2 uint16
}

func newSnd() *Snd {
	return &Snd{table: make(map[[2]int32]*Wave)}
}

func LoadSnd(filename string) (*Snd, error) {
	s := newSnd()
	f, err := os.Open(filename)
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
	for i := uint32(0); i < numberOfSounds; i++ {
		f.Seek(int64(subHeaderOffset), 0)
		var nextSubHeaderOffset uint32
		if err := read(&nextSubHeaderOffset); err != nil {
			return nil, err
		}
		var subFileLenght uint32
		if err := read(&subFileLenght); err != nil {
			return nil, err
		}
		var num [2]int32
		if err := read(&num); err != nil {
			return nil, err
		}
		if num[0] >= 0 && num[1] >= 0 {
			_, ok := s.table[num]
			if !ok {
				tmp, err := ReadWave(f, int64(subHeaderOffset))
				if err != nil {
					sys.appendToConsole(fmt.Sprintf("WARNING: %v sound can't be read: %v,%v", filename, num[0], num[1]))
					sys.errLog.Printf("%v sound can't be read: %v,%v\n", filename, num[0], num[1])
				} else {
					s.table[num] = tmp
				}
			}
		}
		subHeaderOffset = nextSubHeaderOffset
	}
	return s, nil
}
func (s *Snd) Get(gn [2]int32) *Wave {
	return s.table[gn]
}
func (s *Snd) play(gn [2]int32, volumescale int32) bool {
	c := sys.sounds.GetChannel()
	if c == nil {
		return false
	}
	c.sound = s.Get(gn)
	c.SetVolume(volumescale * 64 / 25)
	return c.sound != nil
}
func (s *Snd) stop(gn [2]int32) {
	w := s.Get(gn)
	for k, v := range sys.sounds {
		if v.sound != nil && v.sound == w {
			sys.sounds[k] = Sound{volume: 256, freqmul: 1}
			//break
		}
	}
}

func newWave() *Wave {
	return &Wave{SamplesPerSec: 11025, Channels: 1, BytesPerSample: 1}
}
func loadFromSnd(filename string, g, s int32, max uint32) (*Wave, error) {
	w := newWave()
	f, err := os.Open(filename)
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
	var ver uint16
	if err := read(&ver); err != nil {
		return nil, err
	}
	var ver2 uint16
	if err := read(&ver2); err != nil {
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
		var subFileLenght uint32
		if err := read(&subFileLenght); err != nil {
			return nil, err
		}
		var num [2]int32
		if err := read(&num); err != nil {
			return nil, err
		}
		if num[0] == g && num[1] == s {
			tmp, err := ReadWave(f, int64(subHeaderOffset))
			if err != nil {
				return nil, err
			}
			return tmp, nil
		}
		subHeaderOffset = nextSubHeaderOffset
	}
	return w, nil
}
func (w *Wave) play() bool {
	c := sys.sounds.GetChannel()
	if c == nil {
		return false
	}
	c.sound = w
	return c.sound != nil
}

// ------------------------------------------------------------------
// Sound

type Sound struct {
	sound   *Wave
	volume  int16
	loop    bool
	freqmul float32
	fidx    float64
}

func (s *Sound) mix() {
	if s.sound == nil {
		return
	}
	s.fidx = sys.mixer.Mix(s.sound.Wav, s.fidx,
		int(s.sound.BytesPerSample), int(s.sound.Channels),
		float64(s.sound.SamplesPerSec)*float64(s.freqmul), s.loop,
		float32(s.volume)/256)
	if int(s.fidx) >= len(s.sound.Wav)/
		int(s.sound.BytesPerSample*s.sound.Channels) {
		s.sound = nil
		s.fidx = 0
	}
}
func (s *Sound) SetVolume(vol int32) {
	if vol < 0 {
		s.volume = 0
	} else if vol > 512 {
		s.volume = 512
	} else {
		s.volume = int16(vol)
	}
}
func (s *Sound) SetPan(pan float32, offset *float32) {
	// 未実装
}

type Sounds []Sound

func newSounds(size int) (s Sounds) {
	s = make(Sounds, size)
	for i := range s {
		s[i] = Sound{volume: 256, freqmul: 1}
	}
	return
}
func (s Sounds) GetChannel() *Sound {
	for i := range s {
		if s[i].sound == nil {
			return &s[i]
		}
	}
	return nil
}
func (s Sounds) mixSounds() {
	for i := range s {
		s[i].mix()
	}
}
