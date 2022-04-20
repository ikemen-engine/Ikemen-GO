package main

import (
	"encoding/binary"
	"math"
	"os"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

const (
	audioOutLen    = 2048
	audioFrequency = 48000
	audioPrecision = 4
	audioResampleQuality = 3
)

// ------------------------------------------------------------------
// Normalizer

type Normalizer struct {
	streamer beep.Streamer
	mul  float64
	l, r *NormalizerLR
}

func NewNormalizer(st beep.Streamer) *Normalizer {
	return &Normalizer{streamer: st, mul: 4,
		l: &NormalizerLR{1, 0, 1, 1 / 32.0, 0, 0},
		r: &NormalizerLR{1, 0, 1, 1 / 32.0, 0, 0}}
}

func (n *Normalizer) Stream(samples [][2]float64) (s int, ok bool) {
	s, ok = n.streamer.Stream(samples)
	for i:= range samples[:s] {
		lmul := n.l.process(n.mul, &samples[i][0])
		rmul := n.r.process(n.mul, &samples[i][1])
		if sys.audioDucking {
			n.mul = math.Min(16.0, math.Min(lmul, rmul))
		} else {
			n.mul = 0.5 * (float64(sys.wavVolume) * float64(sys.masterVolume) * 0.0001)
		}
	}
	return s, ok
}

func (n *Normalizer) Err() error {
        return n.streamer.Err()
}

type NormalizerLR struct {
	heri, herihenka, fue, heikin, katayori, katayori2 float64
}

func (n *NormalizerLR) process(bai float64, sam *float64) float64 {
	n.katayori += (*sam - n.katayori) / (audioFrequency/110.0 + 1)
	n.katayori2 += (*sam - n.katayori2) / (audioFrequency/112640.0 + 1)
	s := (n.katayori2 - n.katayori) * bai
	if math.Abs(s) > 1 {
		bai *= math.Pow(math.Abs(s), -n.heri)
		n.herihenka += 32 * (1 - n.heri) / float64(audioFrequency+32)
		s = math.Copysign(1.0, s)
	} else {
		tmp := (1 - math.Pow(1-math.Abs(s), 64)) * math.Pow(0.5-math.Abs(s), 3)
		bai += bai * (n.heri*(1/32.0-n.heikin)/n.fue + tmp*n.fue*(1-n.heri)/32) /
			(audioFrequency*2/8.0 + 1)
		n.herihenka -= (0.5 - n.heikin) * n.heri / (audioFrequency * 2)
	}
	n.fue += (1.0 - n.fue*(math.Abs(s)+1/32.0)) / (audioFrequency * 2)
	n.heikin += (math.Abs(s) - n.heikin) / (audioFrequency * 2)
	n.heri += n.herihenka
	if n.heri < 0 {
		n.heri = 0
	} else if n.heri > 0 {
		n.heri = 1
	}
	*sam = s
	return bai
}

// ------------------------------------------------------------------
// Bgm

type Bgm struct {
	filename     string
	bgmVolume    int
	bgmLoopStart int
	bgmLoopEnd   int
	loop         int
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

func (bgm *Bgm) Open(filename string, loop, bgmVolume, bgmLoopStart, bgmLoopEnd int) {
	bgm.filename = filename
	bgm.loop = loop
	bgm.bgmVolume = bgmVolume
	bgm.bgmLoopStart = bgmLoopStart
	bgm.bgmLoopEnd = bgmLoopEnd
	// Starve the current music streamer
	if bgm.ctrl != nil {
		speaker.Lock()
		bgm.ctrl.Streamer = nil
		speaker.Unlock()
	}

	// TODO: Throw a degbug warning if this triggers
	if bgmVolume > sys.maxBgmVolume {
		bgmVolume = sys.maxBgmVolume
	}

	if HasExtension(bgm.filename, ".ogg") {
		bgm.ReadVorbis(loop, bgmVolume)
	} else if HasExtension(bgm.filename, ".mp3") {
		bgm.ReadMp3(loop, bgmVolume)
		//} else if HasExtension(bgm.filename, ".flac") {
		//	bgm.ConvertFLAC(loop, bgmVolume)
	} else if HasExtension(bgm.filename, ".wav") {
		bgm.ReadWav(loop, bgmVolume)
	}

	speaker.Play(bgm.ctrl)
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
	bgm.resampler = beep.Resample(audioResampleQuality, format.SampleRate, audioFrequency, bgm.volume)
	bgm.ctrl = &beep.Ctrl{Streamer: bgm.resampler}
}

func (bgm *Bgm) Pause() {
	if bgm.ctrl != nil {
		speaker.Lock()
		bgm.ctrl.Paused = true
		speaker.Unlock()
	}
}

func (bgm *Bgm) UpdateVolume() {
	if bgm.volume == nil {
		return
	}
	speaker.Lock()
	bgm.volume.Volume = -5 + float64(sys.bgmVolume)*0.06*(float64(sys.masterVolume)/100)*(float64(bgm.bgmVolume)/100)
	speaker.Unlock()
}

// ------------------------------------------------------------------
// Wave

type Wave struct {
	Buffer *beep.Buffer
}

func ReadWave(f *os.File, ofs int64) (*Wave, error) {
	s, fmt, err := wav.Decode(f)
	if err != nil {
		return nil, err
	}
	w := newWave(fmt.SampleRate)
	w.Buffer.Append(s)
	return w, nil
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
	return LoadSndFiltered(filename, func(gn [2]int32) bool { return gn[0] >= 0 && gn[1] >= 0 }, 0)
}

// Parse a .snd file and return an Snd structure with its contents
// The "keepItem" function allows to filter out unwanted waves.
// If max > 0, the function returns immediately when a matching entry is found. It also gives up after "max" non-matching entries.
func LoadSndFiltered(filename string, keepItem func([2]int32) bool, max uint32) (*Snd, error) {
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
		if keepItem(num) {
			_, ok := s.table[num]
			if !ok {
				tmp, err := ReadWave(f, int64(subHeaderOffset))
				if err != nil {
					sys.errLog.Printf("%v sound can't be read: %v,%v\n", filename, num[0], num[1])
					if max > 0 {
						return nil, err
					}
				} else {
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
func (s *Snd) Get(gn [2]int32) *Wave {
	return s.table[gn]
}
func (s *Snd) play(gn [2]int32, volumescale int32, pan float32) bool {
	c := sys.sounds.reserveChannel()
	if c == nil {
		return false
	}
	w := s.Get(gn)
	c.Play(w, false, 1.0)
	c.SetVolume(float32(volumescale * 64 / 25))
	c.SetPan(pan, 0, nil)
	return w != nil
}
func (s *Snd) stop(gn [2]int32) {
	sys.sounds.stop(s.Get(gn))
}

func newWave(sampleRate beep.SampleRate) *Wave {
	return &Wave{beep.NewBuffer(beep.Format{SampleRate: sampleRate, NumChannels: 2, Precision: audioPrecision})}
}
func loadFromSnd(filename string, g, s int32, max uint32) (*Wave, error) {
	// Load the snd file
	snd, err := LoadSndFiltered(filename, func(gn [2]int32) bool { return gn[0] == g && gn[1] == s }, max)
	if err != nil {
		return nil, err
	}
	tmp, ok := snd.table[[2]int32{g, s}]
	if !ok {
		return newWave(11025), nil
	}
	return tmp, nil
}
func (w *Wave) play() bool {
	c := sys.sounds.reserveChannel()
	if c == nil {
		return false
	}
	c.Play(w, false, 1.0)
	return w != nil
}
func (w *Wave) getDuration() float32 {
	return float32(w.Buffer.Format().SampleRate.D(w.Buffer.Len()))
}

// ------------------------------------------------------------------
// SoundEffect (handles volume and panning)

type SoundEffect struct {
	streamer beep.Streamer
	volume float32
	ls, p float32
	x *float32
}

func (s *SoundEffect) Stream(samples [][2]float64) (n int, ok bool) {
	// TODO: Test mugen panning in relation to PanningWidth and zoom settings
	lv, rv := s.volume, s.volume
	if sys.stereoEffects && (s.x != nil || s.p != 0) {
		var r float32
		if s.x != nil { // pan
			r = ((sys.xmax - s.ls**s.x) - s.p) / (sys.xmax - sys.xmin)
		} else { // abspan
			r = ((sys.xmax-sys.xmin)/2 - s.p) / (sys.xmax - sys.xmin)
		}
		sc := sys.panningRange / 100
		of := (100 - sys.panningRange) / 200
		lv = s.volume * 2 * (r*sc + of)
		rv = s.volume * 2 * ((1-r)*sc + of)
		if lv > 512 {
			lv = 512
		} else if lv < 0 {
			lv = 0
		}
		if rv > 512 {
			rv = 512
		} else if rv < 0 {
			rv = 0
		}
	}

	n, ok = s.streamer.Stream(samples)
	for i:= range samples[:n] {
		samples[i][0] *= float64(lv / 256)
		samples[i][1] *= float64(rv / 256)
	}
	return n, ok
}

func (s *SoundEffect) Err() error {
	return s.streamer.Err()
}

// ------------------------------------------------------------------
// Sound (sound channel)

type Sound struct {
	streamer  beep.StreamSeeker
	sfx     *SoundEffect
	ctrl    *beep.Ctrl
	sound   *Wave
}

func (s *Sound) mix() {
}
func (s *Sound) Play(w *Wave, loop bool, freqmul float32) {
	if w == nil {
		return
	}
	s.sound = w
	s.streamer = s.sound.Buffer.Streamer(0, s.sound.Buffer.Len())
	loopCount := int(1)
	if loop {
		loopCount = -1
	}
	looper := beep.Loop(loopCount, s.streamer)
	s.sfx = &SoundEffect{streamer: looper, volume: 256}
	srcRate := s.sound.Buffer.Format().SampleRate
	dstRate := beep.SampleRate(audioFrequency / freqmul)
	resampler := beep.Resample(audioResampleQuality, srcRate, dstRate, s.sfx)
	s.ctrl = &beep.Ctrl{Streamer: resampler}
	speaker.Play(s.ctrl)
}
func (s *Sound) IsPlaying() bool {
	return s.sound != nil
}
func (s *Sound) Stop() {
	if s.ctrl != nil {
		speaker.Lock()
		s.ctrl.Streamer = nil
		speaker.Unlock()
	}
	s.sound = nil
}
func (s *Sound) SetVolume(vol float32) {
	if s.ctrl != nil {
		s.sfx.volume = float32(math.Max(0, math.Min(float64(vol), 512)))
	}
}
func (s *Sound) SetPan(p, ls float32, x *float32) {
	if s.ctrl != nil {
		s.sfx.ls = ls
		s.sfx.x = x
		s.sfx.p = p * ls
	}
}

// ------------------------------------------------------------------
// Sounds (collection of prioritised sound channels)

type Sounds struct {
	channels []Sound
}

func newSounds(size int32) *Sounds {
	s := &Sounds{}
	s.setSize(size)
	return s
}
func (s *Sounds) setSize(size int32)  {
	if size > s.numChannels() {
		c := make([]Sound, size - s.numChannels())
		s.channels = append(s.channels, c...)
	} else if size < s.numChannels() {
		s.channels = s.channels[:size]
	}
}
func (s *Sounds) newChannel(ch int32, lowpriority bool) *Sound {
        ch = Min(255, ch)
        if ch >= 0 {
                if lowpriority {
                        if s.numChannels() > ch && s.channels[ch].IsPlaying() {
                                return nil
                        }
                }
                if s.numChannels() < ch+1 {
			s.setSize(ch+1)
                }
		s.channels[ch].Stop()
                return &s.channels[ch]
        }
        if s.numChannels() < 256 {
		s.setSize(256)
        }
        for i := 255; i >= 0; i-- {
                if !s.channels[i].IsPlaying() {
                        return &s.channels[i]
                }
        }
        return nil
}
func (s *Sounds) numChannels() int32 {
	return int32(len(s.channels))
}
func (s *Sounds) reserveChannel() *Sound {
	for i := range s.channels {
		if !s.channels[i].IsPlaying() {
			return &s.channels[i]
		}
	}
	return nil
}
func (s *Sounds) getChannel(ch int32) *Sound {
	if ch >= 0 && ch < s.numChannels() {
		return &s.channels[ch]
	}
	return nil
}
func (s *Sounds) IsPlaying(w *Wave) bool {
	for _, v := range s.channels {
		if v.sound != nil && v.sound == w {
			return true
		}
	}
	return false
}
func (s *Sounds) stop(w *Wave) {
	for k, v := range s.channels {
		if v.sound != nil && v.sound == w {
			s.channels[k].Stop()
		}
	}
}
func (s *Sounds) tickSounds() {
	for i := range s.channels {
		if s.channels[i].IsPlaying() {
			if s.channels[i].streamer.Position() >= s.channels[i].sound.Buffer.Len() {
				s.channels[i].sound = nil
			}
		}
	}
}
