package main

import (
	"github.com/jfreymuth/go-vorbis/ogg/vorbis"
	"github.com/timshannon/go-openal/openal"
	"io"
	"os"
	"time"
)

const (
	audioOutLen    = 2048
	audioFrequency = 48000
)

var bgm = &Vorbis{openReq: make(chan string, 1)}
var audioContext *openal.Context

func audioOpen() {
	if audioContext == nil {
		device := openal.OpenDevice("")
		if device == nil {
			chk(openal.Err())
		}
		audioContext = device.CreateContext()
		chk(device.Err())
		audioContext.Activate()
		go soundWrite()
	}
}
func soundWrite() {
	src := openal.NewSource()
	bufs := openal.NewBuffers(2)
	for i := range bufs {
		bufs[i].SetDataInt16(openal.FormatStereo16,
			make([]int16, audioOutLen*2), audioFrequency)
	}
	src.QueueBuffers(bufs)
	chk(openal.Err())
	src.Play()
	var out []int16
	for !gameEnd {
		if src.BuffersProcessed() > 0 {
			switch {
			case out != nil:
				buf := src.UnqueueBuffer()
				buf.SetDataInt16(openal.FormatStereo16, out, audioFrequency)
				out = nil
				src.QueueBuffer(buf)
				chk(openal.Err())
				continue
			default:
				time.Sleep(time.Millisecond)
			}
		} else {
			if src.State() != openal.Playing {
				src.Play()
			}
			time.Sleep(10 * time.Millisecond)
		}
		if out == nil {
			out = bgm.read()
		}
	}
	bufs.Delete()
	src.Delete()
	openal.NullContext.Activate()
	device := audioContext.GetDevice()
	audioContext.Destroy()
	audioContext = nil
	device.CloseDevice()
}

type Vorbis struct {
	dec     *vorbis.Vorbis
	fh      *os.File
	buf     []int16
	openReq chan string
}

func (v *Vorbis) Open(file string) {
	v.openReq <- file
}
func (v *Vorbis) openFile(file string) bool {
	v.clear()
	var err error
	if v.fh, err = os.Open(file); err != nil {
		return false
	}
	return v.restart()
}
func (v *Vorbis) restart() bool {
	if v.fh == nil {
		return false
	}
	_, err := v.fh.Seek(0, 0)
	chk(err)
	if v.dec, err = vorbis.Open(v.fh); err != nil {
		v.clear()
		return false
	}
	v.buf = nil
	return true
}
func (v *Vorbis) clear() {
	if v.dec != nil {
		v.dec = nil
	}
	if v.fh != nil {
		chk(v.fh.Close())
		v.fh = nil
	}
}
func (v *Vorbis) samToAudioOut(buf [][]float32) (out []int16) {
	var o1i int
	if len(buf) == 1 {
		o1i = 0
	} else {
		o1i = 1
	}
	sr := audioFrequency / float64(v.dec.SampleRate())
	out = make([]int16, 2*(int(float64(len(buf[0])-1)*sr)+1))
	oldouti := -2
	for i := range buf[0] {
		outi := 2 * int(float64(i)*sr)
		for j := oldouti + 2; j <= outi; j += 2 {
			out[j], out[j+1] = int16(32767*buf[0][i]), int16(32767*buf[o1i][i])
		}
		oldouti = outi
	}
	return
}
func (v *Vorbis) read() (out []int16) {
	select {
	case file := <-v.openReq:
		v.openFile(file)
	default:
	}
	for v.dec != nil {
		if len(v.buf) >= audioOutLen*2 {
			out = v.buf[:audioOutLen*2]
			v.buf = v.buf[audioOutLen*2:]
			return
		}
		localdec := v.dec
		for ; len(v.buf) < audioOutLen*2 && localdec != nil; localdec = v.dec {
			sam, err := localdec.DecodePacket()
			if err == io.EOF {
				v.restart()
				continue
			} else {
				chk(err)
			}
			v.buf = append(v.buf, v.samToAudioOut(sam)...)
		}
	}
	return
}
