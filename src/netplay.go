package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// This defines the number of frames to store for the net buffer inputs (digital and analog)
	NETBUF_NUM_FRAMES int32 = 32

	// Replay files store the same payload for both delay-based and rollback netplay:
	// 2 bytes of digital inputs followed by 6 bytes of signed analog axes per controller.
	REPLAY_NUM_INPUTS  = MaxSimul * 2
	REPLAY_INPUT_BYTES = 2 + 6

	syncConfigVersion     uint16 = 1
	replayFormatVersion   uint16 = 2
	replayMagic                  = "IKRPLCFG"
	replaySyncMagic              = "IKRPLSYN"
	replaySyncSearchLimit        = int64(64 << 20)

	netLoadingPollTimeout = time.Millisecond

	netLoadingReadyToken byte = 0xC7
	netLoadingAckToken   byte = 0x7C
)

type LoadingPhase uint8

const (
	LP_Idle LoadingPhase = iota
	LP_HostReadySent
	LP_GuestReadySent
)

type NetState int

const (
	NS_Stop NetState = iota
	NS_Playing
	NS_End
	NS_Stopped
	NS_Error
)

type SyncScope string

const (
	syncStrict SyncScope = "strict"
	syncHost   SyncScope = "host"
)

type SyncSetting struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type SyncHandshake struct {
	SyncVersion        uint16        `json:"sync_version"`
	Strict             []SyncSetting `json:"strict,omitempty"`
	Host               []SyncSetting `json:"host,omitempty"`
	ContentFingerprint string        `json:"content_fingerprint,omitempty"`
}

type ReplayHeader struct {
	FormatVersion      uint16        `json:"format_version"`
	SyncVersion        uint16        `json:"sync_version"`
	Strict             []SyncSetting `json:"strict,omitempty"`
	Host               []SyncSetting `json:"host,omitempty"`
	ContentFingerprint string        `json:"content_fingerprint,omitempty"`
}

type SessionConfigOverride struct {
	Active             bool
	Source             string
	Strict             []SyncSetting
	OriginalHost       []SyncSetting
	AppliedHost        []SyncSetting
	SyncVersion        uint16
	ContentFingerprint string
}

// -----------------------------------------------------------------------------
// Replay payload encoding
// -----------------------------------------------------------------------------

func writeReplayInput(w io.Writer, ibit InputBits, axes [6]int8) error {
	var buf [REPLAY_INPUT_BYTES]byte
	binary.LittleEndian.PutUint16(buf[:2], uint16(ibit))
	for i := 0; i < len(axes); i++ {
		buf[2+i] = byte(axes[i])
	}
	_, err := w.Write(buf[:])
	return err
}

func readReplayInput(r io.Reader, ibit *InputBits, axes *[6]int8) error {
	var buf [REPLAY_INPUT_BYTES]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return err
	}
	*ibit = InputBits(int16(binary.LittleEndian.Uint16(buf[:2])))
	for i := 0; i < len(axes); i++ {
		axes[i] = int8(buf[2+i])
	}
	return nil
}

// -----------------------------------------------------------------------------
// Net input ring buffer
// -----------------------------------------------------------------------------

// NetBuffer holds the inputs that are sent between players
type NetBuffer struct {
	buf              [NETBUF_NUM_FRAMES]InputBits
	axisBuf          [NETBUF_NUM_FRAMES][6]int8
	curT, inpT, senT int32
	InputReader      *InputReader
}

func NewNetBuffer() NetBuffer {
	return NetBuffer{
		InputReader: NewInputReader(),
	}
}

func (nb *NetBuffer) reset(time int32) {
	nb.curT, nb.inpT, nb.senT = time, time, time
	nb.InputReader.Reset()
}

// Convert local player's key inputs into input bits for sending
func (nb *NetBuffer) writeNetBuffer(in int) {
	if nb.inpT-nb.curT < NETBUF_NUM_FRAMES {
		nb.buf[nb.inpT&(NETBUF_NUM_FRAMES-1)].KeysToBits(nb.InputReader.LocalInput(in))
		nb.axisBuf[nb.inpT&(NETBUF_NUM_FRAMES-1)] = nb.InputReader.LocalAnalogInput(in)
		nb.inpT++
	}
}

// Read input bits from the net buffer
func (nb *NetBuffer) readNetBuffer() [14]bool {
	if nb.curT < nb.inpT {
		return nb.buf[nb.curT&(NETBUF_NUM_FRAMES-1)].BitsToKeys()
	}
	return [14]bool{}
}

func (nb *NetBuffer) readNetBufferAnalog() [6]int8 {
	if nb.curT < nb.inpT {
		return nb.axisBuf[nb.curT&(NETBUF_NUM_FRAMES-1)]
	}
	return [6]int8{}
}

// -----------------------------------------------------------------------------
// Netplay connection
// -----------------------------------------------------------------------------

// NetConnection manages the communication between players
type NetConnection struct {
	ln               *net.TCPListener
	conn             *net.TCPConn
	st               NetState
	sendEnd          chan bool
	recvEnd          chan bool
	buf              [MaxSimul * 2]NetBuffer // We skip attached characters here because they never have human inputs
	locIn            int
	remIn            int
	time             int32
	stoppedcnt       int32
	delay            int32
	recording        *os.File
	host             bool
	preMatchTime     int32
	closing          chan struct{}
	closeOnce        sync.Once
	uiInputDebounced bool
	headerWritten    bool
	loadingPhase     LoadingPhase
}

func NewNetConnection() *NetConnection {
	nc := &NetConnection{
		st:      NS_Stop,
		sendEnd: make(chan bool, 1),
		recvEnd: make(chan bool, 1),
		closing: make(chan struct{}),
	}
	nc.sendEnd <- true
	nc.recvEnd <- true

	for i := range nc.buf {
		nc.buf[i] = NewNetBuffer()
	}

	return nc
}

func (nc *NetConnection) isClosing() bool {
	if nc == nil || nc.closing == nil {
		return true
	}
	select {
	case <-nc.closing:
		return true
	default:
		return false
	}
}

func (nc *NetConnection) Close() {
	if nc == nil {
		return
	}
	// Reset any pending loading rendezvous.
	nc.loadingPhase = LP_Idle
	// Ensure connect/accept goroutines stop retrying/handshaking.
	nc.closeOnce.Do(func() {
		if nc.closing != nil {
			close(nc.closing)
		}
	})
	// Ensure send/recv goroutines can exit even if they have nothing to write.
	if nc.st == NS_Playing {
		nc.st = NS_End
	}
	if nc.ln != nil {
		nc.ln.Close()
		nc.ln = nil
	}
	if nc.conn != nil {
		nc.conn.Close()
	}
	if nc.sendEnd != nil {
		<-nc.sendEnd
		close(nc.sendEnd)
		nc.sendEnd = nil
	}
	if nc.recvEnd != nil {
		<-nc.recvEnd
		close(nc.recvEnd)
		nc.recvEnd = nil
	}
	if nc.recording != nil {
		nc.recording.Close()
		nc.recording = nil
		nc.headerWritten = false
	}
	nc.conn = nil
	nc.uiInputDebounced = false
}

func (nc *NetConnection) end() {
	if nc.st != NS_Error {
		nc.st = NS_End
	}
	nc.Close()
}

func (nc *NetConnection) Stop() {
	if sys.esc {
		nc.end()
	} else {
		if nc.st != NS_End && nc.st != NS_Error {
			nc.st = NS_Stop
		}
		<-nc.sendEnd
		nc.sendEnd <- true
		<-nc.recvEnd
		nc.recvEnd <- true
	}
}

func (nc *NetConnection) GetHostGuestRemap() (host, guest int) {
	host, guest = -1, -1
	for i, c := range sys.aiLevel {
		if c == 0 {
			if host < 0 {
				host = i
			} else if guest < 0 {
				guest = i
			}
		}
	}
	if host < 0 {
		host = 0
	}
	if guest < 0 {
		guest = (host + 1) % len(nc.buf)
	}
	return
}

func (nc *NetConnection) Accept(port int) error {
	ln, err := net.Listen("tcp", net.JoinHostPort("", Itoa(port)))
	if err != nil {
		return err
	}

	tcpLn, ok := ln.(*net.TCPListener)
	if !ok {
		ln.Close()
		return fmt.Errorf("failed to cast net.Listener to *net.TCPListener")
	}

	nc.ln = tcpLn
	nc.host = true
	nc.conn = nil // Make sure this is a new connection
	nc.locIn, nc.remIn = nc.GetHostGuestRemap()

	lnLocal := nc.ln
	SafeGo(func() {
		defer lnLocal.Close()

		tempConn, err := lnLocal.AcceptTCP()
		if err != nil {
			return
		}

		if nc.isClosing() {
			tempConn.Close()
			return
		}

		// Don't allow the handshake to block forever (important when shutting down).
		_ = tempConn.SetDeadline(time.Now().Add(2 * time.Second))

		if sys.cfg.Netplay.RollbackNetcode {
			sys.rollback.session.remoteIp = tempConn.RemoteAddr().(*net.TCPAddr).IP.String()
		}

		//Send handshake
		if _, err := tempConn.Write([]byte("IKEMENGO")); err != nil {
			tempConn.Close()
			return
		}

		// Wait for client acknowledgment
		ack := make([]byte, 8) // Length of our "password"
		_, err = io.ReadFull(tempConn, ack)
		if err != nil || string(ack) != "IKEMENGO" {
			tempConn.Close()
			return
		}

		// Handshake complete; clear deadlines for normal play.
		_ = tempConn.SetDeadline(time.Time{})

		// Handshake complete. Make temp connection permanent
		if nc.isClosing() {
			tempConn.Close()
			return
		}
		nc.conn = tempConn
	})

	return nil
}

func (nc *NetConnection) Connect(server string, port int) {
	nc.host = false
	nc.conn = nil // Make sure this is a new connection
	nc.remIn, nc.locIn = nc.GetHostGuestRemap()

	SafeGo(func() {
		d := net.Dialer{Timeout: 1 * time.Second}
		for {
			if nc.isClosing() {
				return
			}
			tempConn, err := d.Dial("tcp", net.JoinHostPort(server, Itoa(port)))
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			tcpConn := tempConn.(*net.TCPConn)
			if nc.isClosing() {
				tcpConn.Close()
				return
			}

			// Don't allow the handshake to block forever (important when shutting down).
			_ = tcpConn.SetDeadline(time.Now().Add(2 * time.Second))

			// Wait for host handshake
			buf := make([]byte, 8)
			_, err = io.ReadFull(tcpConn, buf)
			if err != nil || string(buf) != "IKEMENGO" {
				tcpConn.Close()
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Send acknowledgment
			if _, err := tcpConn.Write([]byte("IKEMENGO")); err != nil {
				tcpConn.Close()
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Handshake complete; clear deadlines for normal play.
			_ = tcpConn.SetDeadline(time.Time{})

			// Handshake complete. Make temp connection permanent
			if nc.isClosing() {
				tcpConn.Close()
				return
			}
			nc.conn = tcpConn
			return
		}
	})
}

func (nc *NetConnection) IsConnected() bool {
	if nc == nil {
		return false
	}
	connected := nc.conn != nil
	// Stop a held button from registering as a fresh press and auto-accepting the first menu.
	if connected && !nc.uiInputDebounced {
		nc.uiInputDebounced = true
		sys.uiResetTokenGuard()
	} else if !connected {
		nc.uiInputDebounced = false
	}
	return connected
}

// Wire primitives
func (nc *NetConnection) readI8() (int8, error) {
	b := [1]byte{}
	if _, err := nc.conn.Read(b[:]); err != nil {
		return 0, err
	}
	return int8(b[0]), nil
}

func (nc *NetConnection) writeI8(i8 int8) error {
	b := [...]byte{byte(i8)}
	if _, err := nc.conn.Write(b[:]); err != nil {
		return err
	}
	return nil
}

func (nc *NetConnection) writeU8(u8 byte) error {
	_, err := nc.conn.Write([]byte{u8})
	return err
}

func (nc *NetConnection) readI16() (int16, error) {
	b := [2]byte{}
	if _, err := nc.conn.Read(b[:]); err != nil {
		return 0, err
	}
	return int16(b[0]) | int16(b[1])<<8, nil
}

func (nc *NetConnection) writeI16(i16 int16) error {
	b := [...]byte{byte(i16), byte(i16 >> 8)}
	if _, err := nc.conn.Write(b[:]); err != nil {
		return err
	}
	return nil
}

func (nc *NetConnection) readI32() (int32, error) {
	b := [4]byte{}
	if _, err := nc.conn.Read(b[:]); err != nil {
		return 0, err
	}
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24, nil
}

func (nc *NetConnection) writeI32(i32 int32) error {
	b := [...]byte{byte(i32), byte(i32 >> 8), byte(i32 >> 16), byte(i32 >> 24)}
	if _, err := nc.conn.Write(b[:]); err != nil {
		return err
	}
	return nil
}

func (nc *NetConnection) writeBytes(data []byte) error {
	if err := nc.writeI32(int32(len(data))); err != nil {
		return err
	}
	_, err := nc.conn.Write(data)
	return err
}

func (nc *NetConnection) readBytes() ([]byte, error) {
	n, err := nc.readI32()
	if err != nil {
		return nil, err
	}
	if n < 0 || n > 1<<20 {
		return nil, Error("Invalid synchronization payload")
	}
	buf := make([]byte, n)
	_, err = io.ReadFull(nc.conn, buf)
	return buf, err
}

func (nc *NetConnection) writeJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return nc.writeBytes(data)
}

func (nc *NetConnection) readJSON(v interface{}) error {
	data, err := nc.readBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Gameplay-facing netplay API
func (nc *NetConnection) readNetInput(i int) [14]bool {
	if i >= 0 && i < len(nc.buf) {
		return nc.buf[sys.inputRemap[i]].readNetBuffer()
	}
	return [14]bool{}
}

func (nc *NetConnection) readNetInputAnalog(i int) [6]int8 {
	if i >= 0 && i < len(nc.buf) {
		return nc.buf[sys.inputRemap[i]].readNetBufferAnalog()
	}
	return [6]int8{}
}

func (nc *NetConnection) AnyButton() bool {
	for _, nb := range nc.buf {
		if nb.buf[nb.curT&(NETBUF_NUM_FRAMES-1)]&IB_anybutton != 0 {
			return true
		}
	}
	return false
}

func (nc *NetConnection) Synchronize() error {
	if !nc.IsConnected() || nc.st == NS_Error {
		return Error("Cannot connect to the other player")
	}
	// Reset any pending loading rendezvous so it can't interfere with gameplay sync.
	nc.loadingPhase = LP_Idle
	nc.Stop()

	header, err := sys.synchronizeNetplayConfig(nc)
	if err != nil {
		return err
	}
	if nc.recording != nil && !nc.headerWritten && header != nil {
		if err := writeReplayHeader(nc.recording, header); err != nil {
			return err
		}
		nc.headerWritten = true
	}

	// Synchronize to host's random seed
	var seed int32
	if nc.host {
		seed = Random()
		if err := nc.writeI32(seed); err != nil {
			return err
		}
	} else {
		var err error
		if seed, err = nc.readI32(); err != nil {
			return err
		}
	}
	Srand(seed)

	// Synchronize to host's pre-match time
	var pmTime int32
	if nc.host {
		pmTime = sys.preMatchTime
		if err := nc.writeI32(pmTime); err != nil {
			return err
		}
	} else {
		var err error
		if pmTime, err = nc.readI32(); err != nil {
			return err
		}
	}
	nc.preMatchTime = pmTime

	// Write seed and pre-match time to replay file
	if nc.recording != nil {
		if off, err := nc.recording.Seek(0, io.SeekCurrent); err == nil {
			log.Printf("Replay sync write: offset=%d seed=%d pmTime=%d gameRunning=%t rollback=%t",
				off, seed, pmTime, sys.gameRunning, sys.cfg.Netplay.RollbackNetcode)
		}
		if err := writeReplaySync(nc.recording, seed, pmTime); err != nil {
			return err
		}
	}

	// Verify connection time synchronization
	if err := nc.writeI32(nc.time); err != nil {
		return err
	}
	if tmp, err := nc.readI32(); err != nil {
		return err
	} else if tmp != nc.time {
		return Error("Synchronization error")
	}
	if sys.rollback.session != nil {
		sys.rollback.session.netTime = nc.time
	}

	// Reset local and remote input buffers for the current time
	nc.buf[nc.locIn].reset(nc.time)
	nc.buf[nc.remIn].reset(nc.time)
	nc.st = NS_Playing

	// Start sending inputs to remote peer in a goroutine
	sendBuf := &nc.buf[nc.locIn]
	<-nc.sendEnd
	SafeGo(func() {
		nb := sendBuf
		defer func() {
			nc.sendEnd <- true
		}()
		for nc.st == NS_Playing {
			// Check if there are unsent frames
			if nb.senT < nb.inpT {
				// Write digital inputs
				if err := nc.writeI16(int16(nb.buf[nb.senT&(NETBUF_NUM_FRAMES-1)])); err != nil {
					nc.st = NS_Error
					return
				} else {
					// Write analog inputs
					for j := 0; j < len(nb.axisBuf[nb.senT&(NETBUF_NUM_FRAMES-1)]); j++ {
						if err = nc.writeI8(nb.axisBuf[nb.senT&(NETBUF_NUM_FRAMES-1)][j]); err != nil {
							nc.st = NS_Error
							return
						}
					}
				}
				nb.senT++
			}
			time.Sleep(time.Millisecond)
		}
		// Write termination signal to indicate no more input frames
		nc.writeI16(-1)
	})

	// Start receiving inputs from remote peer in a goroutine
	recvBuf := &nc.buf[nc.remIn]
	<-nc.recvEnd
	SafeGo(func() {
		nb := recvBuf
		defer func() {
			nc.recvEnd <- true
		}()
		for nc.st == NS_Playing {
			// Check if there is space in the input buffer
			if nb.inpT-nb.curT < NETBUF_NUM_FRAMES {
				if tmp, err := nc.readI16(); err != nil {
					nc.st = NS_Error
					return
				} else {
					// Read digital inputs
					nb.buf[nb.inpT&(NETBUF_NUM_FRAMES-1)] = InputBits(tmp)
					if tmp < 0 {
						// If remote sent termination signal
						nc.st = NS_Stopped
						return
					} else {
						// Read analog inputs
						for j := 0; j < len(nb.axisBuf[nb.inpT&(NETBUF_NUM_FRAMES-1)]); j++ {
							if tmp2, err := nc.readI8(); err != nil {
								nc.st = NS_Error
								return
							} else {
								nb.axisBuf[nb.inpT&(NETBUF_NUM_FRAMES-1)][j] = tmp2
							}
						}
						nb.inpT++
						nb.senT = nb.inpT
					}
				}
			}
			time.Sleep(time.Millisecond)
		}

		// There may be padding for the axis buffer so safest to just change this.
		for tmp := int16(0); tmp != -1; {
			var err error
			if tmp, err = nc.readI16(); err != nil {
				break
			}
		}
	})

	// Update delay-netplay state after sync. Skip rollback-entry samples;
	// replay frames are appended later from the authoritative GGPO timeline.
	nc.update(sys.rollback.session == nil)

	// Log status
	log.Printf("Network synchronized: seed=%d pmTime=%d time=%d host=%v", seed, pmTime, nc.time, nc.host)

	return nil
}

func (nc *NetConnection) ResetLoadingPhase() {
	if nc == nil {
		return
	}
	nc.loadingPhase = LP_Idle
}

func (nc *NetConnection) tryReadU8() (byte, bool, error) {
	if nc == nil || nc.conn == nil {
		return 0, false, Error("Cannot connect to the other player")
	}
	b := [1]byte{}
	_ = nc.conn.SetReadDeadline(time.Now().Add(netLoadingPollTimeout))
	n, err := nc.conn.Read(b[:])
	_ = nc.conn.SetReadDeadline(time.Time{})
	if n > 0 {
		return b[0], true, nil
	}
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return 0, false, nil
		}
		return 0, false, err
	}
	return 0, false, nil
}

func (nc *NetConnection) finishLoadingBarrier() {
	nc.loadingPhase = LP_Idle
	nc.time = 0
	nc.stoppedcnt = 0
}

// Handshake for both peers finished loading
func (nc *NetConnection) LoadingReady() (bool, error) {
	if sys.esc || sys.gameEnd {
		return false, nil
	}
	if nc == nil || !nc.IsConnected() || nc.st == NS_Error {
		return false, Error("Cannot connect to the other player")
	}
	nc.Stop()

	// Host
	if nc.host {
		switch nc.loadingPhase {
		case LP_Idle:
			if err := nc.writeU8(netLoadingReadyToken); err != nil {
				nc.loadingPhase = LP_Idle
				return false, err
			}
			nc.loadingPhase = LP_HostReadySent
			return false, nil
		case LP_HostReadySent:
			v, ok, err := nc.tryReadU8()
			if err != nil {
				nc.loadingPhase = LP_Idle
				return false, err
			}
			if !ok {
				return false, nil
			}
			if v != netLoadingReadyToken {
				nc.loadingPhase = LP_Idle
				return false, Error("Synchronization error")
			}
			if err := nc.writeU8(netLoadingAckToken); err != nil {
				nc.loadingPhase = LP_Idle
				return false, err
			}
			nc.finishLoadingBarrier()
			return true, nil
		}
		nc.loadingPhase = LP_Idle
		return false, Error("Synchronization error")
	}

	// Guest waits for host readiness, replies with its own readiness, then waits for the host ACK
	switch nc.loadingPhase {
	case LP_Idle:
		v, ok, err := nc.tryReadU8()
		if err != nil {
			nc.loadingPhase = LP_Idle
			return false, err
		}
		if !ok {
			return false, nil
		}
		if v != netLoadingReadyToken {
			nc.loadingPhase = LP_Idle
			return false, Error("Synchronization error")
		}
		if err := nc.writeU8(netLoadingReadyToken); err != nil {
			nc.loadingPhase = LP_Idle
			return false, err
		}
		nc.loadingPhase = LP_GuestReadySent
		return false, nil
	case LP_GuestReadySent:
		v, ok, err := nc.tryReadU8()
		if err != nil {
			nc.loadingPhase = LP_Idle
			return false, err
		}
		if !ok {
			return false, nil
		}
		if v != netLoadingAckToken {
			nc.loadingPhase = LP_Idle
			return false, Error("Synchronization error")
		}
		nc.finishLoadingBarrier()
		return true, nil
	}
	nc.loadingPhase = LP_Idle
	return false, Error("Synchronization error")
}

func (nc *NetConnection) Update() bool {
	return nc.update(true)
}

// Advance delay-netplay by one frame. Skip recording the final sync frame before rollback,
// since it belongs to neither the pre-match replay nor the rollback match replay.
func (nc *NetConnection) update(recordReplay bool) bool {
	if nc.st != NS_Stopped {
		nc.stoppedcnt = 0
	}

	if !sys.gameEnd {
		switch nc.st {
		case NS_Stopped:
			nc.stoppedcnt++
			if nc.stoppedcnt > 60 {
				nc.st = NS_End
				break
			}
			fallthrough
		case NS_Playing:
			for {
				// Determine the earliest frame that has been processed by both local and remote buffers
				foo := Min(nc.buf[nc.locIn].senT, nc.buf[nc.remIn].senT)

				// Calculate network delay difference between local and remote input buffers
				tmp := nc.buf[nc.remIn].inpT + nc.delay>>3 - nc.buf[nc.locIn].inpT

				// Adjust local buffer to synchronize with remote
				if tmp >= 0 {
					// Local buffer is behind. Advance it
					nc.buf[nc.locIn].writeNetBuffer(0)
					if nc.delay > 0 {
						nc.delay--
					}
				} else if tmp < -1 {
					// Local buffer is ahead. Increase delay to catch up
					nc.delay += 4
				}

				// Break loop if we have reached the frame that both buffers have sent
				if nc.time >= foo {
					if sys.esc || !sys.await(sys.gameRenderSpeed()) || nc.st != NS_Playing {
						break
					}
					continue
				}

				// Update current frame time for local and remote buffers
				nc.buf[nc.locIn].curT = nc.time
				nc.buf[nc.remIn].curT = nc.time

				// Write inputs to replay file
				if recordReplay && nc.recording != nil {
					for i := range nc.buf {
						ringIdx := nc.time & (NETBUF_NUM_FRAMES - 1)
						if err := writeReplayInput(nc.recording, nc.buf[i].buf[ringIdx], nc.buf[i].axisBuf[ringIdx]); err != nil {
							log.Printf("Error while writing replay input for controller %d: %v", i, err)
							nc.recording = nil
							break
						}
					}
				}

				nc.time++

				// Ensure local buffer writes any remaining frames
				if nc.time >= foo {
					nc.buf[nc.locIn].writeNetBuffer(0)
				}

				break
			}
		case NS_End, NS_Error:
			sys.esc = true
		}
	}

	if sys.esc {
		nc.end()
	}

	return !sys.gameEnd
}

// -----------------------------------------------------------------------------
// Replay file handling
// -----------------------------------------------------------------------------

func writeReplayHeader(w io.Writer, header *ReplayHeader) error {
	if header == nil {
		return nil
	}
	body, err := json.Marshal(header)
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(replayMagic)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(replayFormatVersion)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(body))); err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

func readReplayHeader(f *os.File) (*ReplayHeader, error) {
	if f == nil {
		return nil, nil
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	magic := make([]byte, len(replayMagic))
	if _, err := io.ReadFull(f, magic); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("missing replay header")
		}
		return nil, err
	}
	if string(magic) != replayMagic {
		return nil, fmt.Errorf("invalid replay header magic")
	}

	var version uint16
	if err := binary.Read(f, binary.LittleEndian, &version); err != nil {
		return nil, err
	}
	if version != replayFormatVersion {
		return nil, fmt.Errorf("unsupported replay format version: %d", version)
	}
	var size uint32
	if err := binary.Read(f, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	body := make([]byte, size)
	if _, err := io.ReadFull(f, body); err != nil {
		return nil, err
	}

	var header ReplayHeader
	if err := json.Unmarshal(body, &header); err != nil {
		return nil, err
	}
	header.FormatVersion = version
	return &header, nil
}

func writeReplaySync(w io.Writer, seed, pmTime int32) error {
	if _, err := w.Write([]byte(replaySyncMagic)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, seed); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, pmTime)
}

type ReplayFile struct {
	file               *os.File
	ibit               [REPLAY_NUM_INPUTS]InputBits
	iaxes              [REPLAY_NUM_INPUTS][6]int8
	preMatchTime       int32
	strictSettings     []SyncSetting
	hostSettings       []SyncSetting
	contentFingerprint string
}

func OpenReplayFile(filename string) *ReplayFile {
	rf, err := os.Open(filename)
	if err != nil {
		log.Printf("Failed to open replay file %s: %v", filename, err)
		return nil
	}
	header, err := readReplayHeader(rf)
	if err != nil {
		log.Printf("Failed to read replay header %s: %v", filename, err)
		rf.Close()
		return nil
	}

	out := &ReplayFile{file: rf}
	if header.SyncVersion != syncConfigVersion {
		log.Printf("Replay sync config version mismatch: replay=%d engine=%d", header.SyncVersion, syncConfigVersion)
		rf.Close()
		return nil
	}
	out.strictSettings = cloneSyncSettings(header.Strict)
	out.hostSettings = cloneSyncSettings(header.Host)
	out.contentFingerprint = header.ContentFingerprint

	log.Printf("Replay file opened: %s", filename)
	return out
}

func (rf *ReplayFile) Close() {
	if rf.file != nil {
		rf.file.Close()
		rf.file = nil
	}
}

// Read input buttons from replay input
func (rf *ReplayFile) readReplayInput(i int) [14]bool {
	if i >= 0 && i < len(rf.ibit) {
		return rf.ibit[sys.inputRemap[i]].BitsToKeys()
	}
	return [14]bool{}
}

func (rf *ReplayFile) readReplayInputAnalog(i int) [6]int8 {
	if i >= 0 && i < len(rf.ibit) {
		remap := sys.inputRemap[i] // we'll be using this a lot

		// New replay file, read in the axes too
		if remap >= 0 && remap < len(rf.iaxes) {
			return rf.iaxes[remap]
		}
	}
	return [6]int8{}
}

func (rf *ReplayFile) AnyButton() bool {
	for _, b := range rf.ibit {
		if b&IB_anybutton != 0 {
			return true
		}
	}
	return false
}

func seekReplaySyncMarker(f *os.File) (int64, error) {
	start, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	data, err := io.ReadAll(io.LimitReader(f, replaySyncSearchLimit))
	if err != nil {
		_, _ = f.Seek(start, io.SeekStart)
		return 0, err
	}
	idx := bytes.Index(data, []byte(replaySyncMagic))
	if idx < 0 {
		_, _ = f.Seek(start, io.SeekStart)
		return 0, fmt.Errorf("replay sync marker not found from offset %d", start)
	}
	skipped := int64(idx)
	_, err = f.Seek(start+skipped+int64(len(replaySyncMagic)), io.SeekStart)
	return skipped, err
}

func (rf *ReplayFile) readReplaySync() (seed, pmTime int32, err error) {
	if rf == nil || rf.file == nil {
		return 0, 0, io.ErrClosedPipe
	}

	start, _ := rf.file.Seek(0, io.SeekCurrent)
	skipped, err := seekReplaySyncMarker(rf.file)
	if err != nil {
		return 0, 0, err
	}
	if skipped > 0 {
		log.Printf("Replay sync read: skipped %d byte(s) to recover sync marker from offset %d", skipped, start)
	}

	if err := binary.Read(rf.file, binary.LittleEndian, &seed); err != nil {
		return 0, 0, err
	}
	if err := binary.Read(rf.file, binary.LittleEndian, &pmTime); err != nil {
		return 0, 0, err
	}
	if off, err := rf.file.Seek(0, io.SeekCurrent); err == nil {
		log.Printf("Replay sync read: offset=%d seed=%d pmTime=%d", off, seed, pmTime)
	}
	return seed, pmTime, nil
}

func (rf *ReplayFile) atReplaySyncMarker() bool {
	if rf == nil || rf.file == nil {
		return false
	}
	pos, err := rf.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	var magic [len(replaySyncMagic)]byte
	_, err = io.ReadFull(rf.file, magic[:])
	_, _ = rf.file.Seek(pos, io.SeekStart)
	return err == nil && string(magic[:]) == replaySyncMagic
}

// Read system variables from replay file
func (rf *ReplayFile) Synchronize() {
	if rf.file != nil {
		seed, pmTime, err := rf.readReplaySync()
		if err != nil {
			log.Printf("Replay synchronization failed: %v", err)
			sys.esc = true
			rf.Close()
			return
		}
		Srand(seed)
		rf.preMatchTime = pmTime
		sys.preMatchTime = pmTime
		rf.Update()
		// Log status
		log.Printf("Replay synchronized: seed=%d pmTime=%d", seed, pmTime)
	}
}

// Read a chunk of inputs from the replay file
func (rf *ReplayFile) Update() bool {
	if rf.file == nil {
		sys.esc = true
	} else {
		if sys.oldNextAddTime > 0 {
			if rf.atReplaySyncMarker() {
				rf.ibit = [REPLAY_NUM_INPUTS]InputBits{}
				rf.iaxes = [REPLAY_NUM_INPUTS][6]int8{}
				return !sys.gameEnd
			}
			rf.ibit = [REPLAY_NUM_INPUTS]InputBits{}
			rf.iaxes = [REPLAY_NUM_INPUTS][6]int8{}

			for i := 0; i < len(rf.ibit); i++ {
				if err := readReplayInput(rf.file, &rf.ibit[i], &rf.iaxes[i]); err != nil {
					if err == io.EOF || err == io.ErrUnexpectedEOF {
						log.Printf("Closing replay file")
					} else {
						log.Printf("Error while reading replay input for controller %d: %v", i, err)
					}
					sys.esc = true
					break
				}
			}
		}

		if sys.esc {
			log.Printf("Closing replay file")
			rf.Close()
		}
	}
	return !sys.gameEnd
}

// -----------------------------------------------------------------------------
// Deterministic config helpers
// -----------------------------------------------------------------------------

func cloneSyncSettings(in []SyncSetting) []SyncSetting {
	out := make([]SyncSetting, len(in))
	copy(out, in)
	return out
}

func formatSyncValue(v reflect.Value) (string, error) {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return "1", nil
		}
		return "0", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Slice, reflect.Array:
		parts := make([]string, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			s, err := formatSyncValue(v.Index(i))
			if err != nil {
				return "", err
			}
			parts = append(parts, s)
		}
		return strings.Join(parts, ", "), nil
	default:
		return "", fmt.Errorf("unsupported sync value kind: %s", v.Kind())
	}
}

func collectSyncSettings(cfg *Config, scope SyncScope) ([]SyncSetting, error) {
	out := make([]SyncSetting, 0, 64)

	var walk func(v reflect.Value, path []string) error
	walk = func(v reflect.Value, path []string) error {
		for v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		}
		if !v.IsValid() || v.Kind() != reflect.Struct {
			return nil
		}

		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			if sf.PkgPath != "" {
				continue
			}

			iniTag := sf.Tag.Get("ini")
			syncTag := sf.Tag.Get("sync")
			fv := v.Field(i)

			if strings.HasPrefix(strings.ToLower(iniTag), "map:") {
				if syncTag != string(scope) {
					continue
				}
				if fv.Kind() != reflect.Map || fv.IsNil() {
					continue
				}
				keys := fv.MapKeys()
				sort.Slice(keys, func(i, j int) bool {
					return strings.ToLower(keys[i].String()) < strings.ToLower(keys[j].String())
				})
				for _, mk := range keys {
					mv := fv.MapIndex(mk)
					s, err := formatSyncValue(mv)
					if err != nil {
						return fmt.Errorf("%s: %w", strings.Join(append(path, mk.String()), "."), err)
					}
					out = append(out, SyncSetting{
						Path:  strings.Join(append(path, mk.String()), "."),
						Value: s,
					})
				}
				continue
			}

			if iniTag == "" {
				if fv.Kind() == reflect.Struct || (fv.Kind() == reflect.Ptr && fv.Type().Elem().Kind() == reflect.Struct) {
					if err := walk(fv, path); err != nil {
						return err
					}
				}
				continue
			}

			nextPath := append(path, iniTag)
			if syncTag == string(scope) {
				s, err := formatSyncValue(fv)
				if err != nil {
					return fmt.Errorf("%s: %w", strings.Join(nextPath, "."), err)
				}
				out = append(out, SyncSetting{
					Path:  strings.Join(nextPath, "."),
					Value: s,
				})
				continue
			}

			if fv.Kind() == reflect.Struct || (fv.Kind() == reflect.Ptr && fv.Type().Elem().Kind() == reflect.Struct) {
				if err := walk(fv, nextPath); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := walk(reflect.ValueOf(cfg).Elem(), nil); err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func settingsToMap(in []SyncSetting) map[string]string {
	out := make(map[string]string, len(in))
	for _, s := range in {
		out[s.Path] = s.Value
	}
	return out
}

func verifySyncSettings(expected, actual []SyncSetting) []string {
	exp := settingsToMap(expected)
	got := settingsToMap(actual)
	out := make([]string, 0, 8)

	for path, ev := range exp {
		gv, ok := got[path]
		if !ok {
			out = append(out, fmt.Sprintf("%s missing (expected=%s)", path, ev))
			continue
		}
		if gv != ev {
			out = append(out, fmt.Sprintf("%s mismatch (expected=%s actual=%s)", path, ev, gv))
		}
	}
	for path, gv := range got {
		if _, ok := exp[path]; !ok {
			out = append(out, fmt.Sprintf("%s unexpected (actual=%s)", path, gv))
		}
	}

	sort.Strings(out)
	return out
}

func logSyncVerification(prefix string, expected, actual []SyncSetting) {
	issues := verifySyncSettings(expected, actual)
	if len(issues) == 0 {
		log.Printf("%s: OK (%d settings)", prefix, len(expected))
		return
	}
	log.Printf("%s: %d issue(s)", prefix, len(issues))
	for _, issue := range issues {
		log.Printf("%s: %s", prefix, issue)
	}
}

func normalizeRatioKey(w, h int64) string {
	if w == 0 || h == 0 {
		return "invalid"
	}
	if w < 0 {
		w = -w
	}
	if h < 0 {
		h = -h
	}
	gcd := func(a, b int64) int64 {
		for b != 0 {
			a, b = b, a%b
		}
		if a == 0 {
			return 1
		}
		return a
	}
	g := gcd(w, h)
	return fmt.Sprintf("%d:%d", w/g, h/g)
}

func effectiveFightAspectKey(m map[string]string) (string, error) {
	parse := func(path string) (int64, error) {
		v, ok := m[path]
		if !ok {
			return 0, fmt.Errorf("missing %s", path)
		}
		return strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	}

	fw, err := parse("Video.FightAspectWidth")
	if err != nil {
		return "", err
	}
	fh, err := parse("Video.FightAspectHeight")
	if err != nil {
		return "", err
	}

	switch {
	case fw == -1 && fh == -1:
		return "stage", nil
	case fw == 0 && fh == 0:
		gw, err := parse("Video.GameWidth")
		if err != nil {
			return "", err
		}
		gh, err := parse("Video.GameHeight")
		if err != nil {
			return "", err
		}
		return "resolution:" + normalizeRatioKey(gw, gh), nil
	default:
		return "custom:" + normalizeRatioKey(fw, fh), nil
	}
}

func validateStrictCompatibility(local, remote []SyncSetting, replayMode bool) error {
	localMap := settingsToMap(local)
	remoteMap := settingsToMap(remote)

	ignoreDirect := map[string]bool{
		"Video.GameWidth":         true,
		"Video.GameHeight":        true,
		"Video.FightAspectWidth":  true,
		"Video.FightAspectHeight": true,
	}
	if replayMode {
		ignoreDirect["Netplay.RollbackNetcode"] = true
	}

	for path, lv := range localMap {
		if ignoreDirect[path] {
			continue
		}
		rv, ok := remoteMap[path]
		if !ok {
			continue
		}
		if lv != rv {
			sys.sessionWarning = fmt.Sprintf(
				sys.motif.WarningInfo.Text.Text["config"],
				fmt.Sprintf("%s (local=%s remote=%s)", path, lv, rv),
			)
			return Error(sys.sessionWarning)
		}
	}

	la, lerr := effectiveFightAspectKey(localMap)
	ra, rerr := effectiveFightAspectKey(remoteMap)
	if lerr == nil && rerr == nil && la != ra {
		sys.sessionWarning = fmt.Sprintf(sys.motif.WarningInfo.Text.Text["config"], fmt.Sprintf("effective fight aspect differs (local=%s remote=%s)", la, ra))
		return Error(sys.sessionWarning)
	}

	return nil
}

func validateHostSignature(local, remote []SyncSetting) error {
	localMap := settingsToMap(local)
	remoteMap := settingsToMap(remote)
	if len(localMap) != len(remoteMap) {
		return Error("Sync host config schema mismatch")
	}
	for path := range localMap {
		if _, ok := remoteMap[path]; !ok {
			return Error("Sync host config schema mismatch")
		}
	}
	return nil
}

func applySyncSettings(cfg *Config, settings []SyncSetting, strict bool) error {
	for _, s := range settings {
		if err := SetValue(cfg, s.Path, s.Value); err != nil {
			if strict {
				return fmt.Errorf("%s: %w", s.Path, err)
			}
			log.Printf("Ignoring sync setting %s during best-effort apply: %v", s.Path, err)
		}
	}
	cfg.normalize()
	sys.onSyncConfigChanged()
	return nil
}

func validateContentFingerprint(local, remote string) error {
	if local != "" && remote != "" && local != remote {
		return Error("Content/build mismatch between peers")
	}
	return nil
}

// -----------------------------------------------------------------------------
// System deterministic/session lifecycle
// -----------------------------------------------------------------------------

func (s *System) onSyncConfigChanged() {
	s.rollbackConfig = s.cfg.Netplay.Rollback
	if s.rollback.session != nil {
		s.rollback.session.config = s.cfg.Netplay.Rollback
	}
}

func (s *System) currentContentFingerprint() string {
	// TODO: replace with real content/build fingerprint handshake.
	return ""
}

func (s *System) beginSessionOverride(source string, strict, host []SyncSetting, contentFingerprint string) error {
	if s.netplayOverride.Active {
		return nil
	}
	originalHost, err := collectSyncSettings(&s.cfg, syncHost)
	if err != nil {
		return err
	}
	if err := applySyncSettings(&s.cfg, host, source == "netplay"); err != nil {
		return err
	}
	appliedHost, err := collectSyncSettings(&s.cfg, syncHost)
	if err != nil {
		return err
	}
	s.netplayOverride = SessionConfigOverride{
		Active:             true,
		Source:             source,
		Strict:             cloneSyncSettings(strict),
		OriginalHost:       originalHost,
		AppliedHost:        cloneSyncSettings(host),
		SyncVersion:        syncConfigVersion,
		ContentFingerprint: contentFingerprint,
	}
	log.Printf("%s sync config override started: strict=%d host=%d fingerprint=%q",
		strings.Title(source), len(strict), len(host), contentFingerprint)
	logSyncVerification(strings.Title(source)+" host override verification", host, appliedHost)
	return nil
}

func (s *System) endSyncSessionOverride() error {
	if !s.netplayOverride.Active {
		return nil
	}
	source := s.netplayOverride.Source
	expected := cloneSyncSettings(s.netplayOverride.OriginalHost)
	if err := applySyncSettings(&s.cfg, s.netplayOverride.OriginalHost, false); err != nil {
		return err
	}
	restoredHost, err := collectSyncSettings(&s.cfg, syncHost)
	if err != nil {
		return err
	}
	logSyncVerification(strings.Title(source)+" host restore verification", expected, restoredHost)
	s.netplayOverride = SessionConfigOverride{}
	return nil
}

func (s *System) currentReplayHeader() *ReplayHeader {
	if !s.netplayOverride.Active {
		return nil
	}
	return &ReplayHeader{
		FormatVersion:      replayFormatVersion,
		SyncVersion:        s.netplayOverride.SyncVersion,
		Strict:             cloneSyncSettings(s.netplayOverride.Strict),
		Host:               cloneSyncSettings(s.netplayOverride.AppliedHost),
		ContentFingerprint: s.netplayOverride.ContentFingerprint,
	}
}

func (s *System) popSessionWarning() string {
	text := s.sessionWarning
	s.sessionWarning = ""
	return text
}

func (s *System) synchronizeNetplayConfig(nc *NetConnection) (*ReplayHeader, error) {
	if s.netplayOverride.Active {
		return s.currentReplayHeader(), nil
	}

	localStrict, err := collectSyncSettings(&s.cfg, syncStrict)
	if err != nil {
		return nil, err
	}
	localHost, err := collectSyncSettings(&s.cfg, syncHost)
	if err != nil {
		return nil, err
	}
	localFingerprint := s.currentContentFingerprint()

	if nc.host {
		hostPayload := SyncHandshake{
			SyncVersion:        syncConfigVersion,
			Strict:             localStrict,
			Host:               localHost,
			ContentFingerprint: localFingerprint,
		}
		log.Printf("Netplay sync config host->peer: sending strict=%d host=%d fingerprint=%q",
			len(localStrict), len(localHost), localFingerprint)
		if err := nc.writeJSON(hostPayload); err != nil {
			return nil, err
		}

		var guestPayload SyncHandshake
		if err := nc.readJSON(&guestPayload); err != nil {
			return nil, err
		}
		log.Printf("Netplay sync config peer->host ack: strict=%d host=%d fingerprint=%q",
			len(guestPayload.Strict), len(guestPayload.Host), guestPayload.ContentFingerprint)
		if guestPayload.SyncVersion != syncConfigVersion {
			return nil, Error("Sync config version mismatch")
		}
		if err := validateHostSignature(localHost, guestPayload.Host); err != nil {
			return nil, err
		}
		if err := validateStrictCompatibility(localStrict, guestPayload.Strict, false); err != nil {
			return nil, err
		}
		if err := validateContentFingerprint(localFingerprint, guestPayload.ContentFingerprint); err != nil {
			return nil, err
		}
		if err := s.beginSessionOverride("netplay", localStrict, localHost, localFingerprint); err != nil {
			return nil, err
		}
		return s.currentReplayHeader(), nil
	}

	var hostPayload SyncHandshake
	if err := nc.readJSON(&hostPayload); err != nil {
		return nil, err
	}
	log.Printf("Netplay sync config host->peer received: strict=%d host=%d fingerprint=%q",
		len(hostPayload.Strict), len(hostPayload.Host), hostPayload.ContentFingerprint)

	guestPayload := SyncHandshake{
		SyncVersion:        syncConfigVersion,
		Strict:             localStrict,
		Host:               localHost,
		ContentFingerprint: localFingerprint,
	}
	sendGuestPayload := func() error {
		log.Printf("Netplay sync config peer->host ack: sending strict=%d host=%d fingerprint=%q",
			len(guestPayload.Strict), len(guestPayload.Host), localFingerprint)
		return nc.writeJSON(guestPayload)
	}

	if hostPayload.SyncVersion != syncConfigVersion {
		if err := sendGuestPayload(); err != nil {
			return nil, err
		}
		return nil, Error("Sync config version mismatch")
	}
	if err := validateHostSignature(localHost, hostPayload.Host); err != nil {
		if werr := sendGuestPayload(); werr != nil {
			return nil, werr
		}
		return nil, err
	}
	if err := validateStrictCompatibility(localStrict, hostPayload.Strict, false); err != nil {
		if werr := sendGuestPayload(); werr != nil {
			return nil, werr
		}
		return nil, err
	}
	if err := validateContentFingerprint(localFingerprint, hostPayload.ContentFingerprint); err != nil {
		if werr := sendGuestPayload(); werr != nil {
			return nil, werr
		}
		return nil, err
	}
	if err := s.beginSessionOverride("netplay", hostPayload.Strict, hostPayload.Host, hostPayload.ContentFingerprint); err != nil {
		return nil, err
	}

	if err := sendGuestPayload(); err != nil {
		return nil, err
	}

	return s.currentReplayHeader(), nil
}

func (s *System) beginReplaySession(rf *ReplayFile) error {
	if rf == nil {
		return nil
	}
	log.Printf("Replay sync header loaded: strict=%d host=%d fingerprint=%q",
		len(rf.strictSettings), len(rf.hostSettings), rf.contentFingerprint)
	if len(rf.hostSettings) == 0 && len(rf.strictSettings) == 0 {
		log.Printf("Replay sync config: no header settings to apply")
		return nil
	}

	localStrict, err := collectSyncSettings(&s.cfg, syncStrict)
	if err != nil {
		return err
	}
	if err := validateStrictCompatibility(localStrict, rf.strictSettings, true); err != nil {
		log.Printf("Replay strict compatibility check failed: %v", err)
		return err
	}
	log.Printf("Replay strict compatibility check: OK")
	return s.beginSessionOverride("replay", rf.strictSettings, rf.hostSettings, rf.contentFingerprint)
}
