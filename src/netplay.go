package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
)

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
	nc.conn = nil
	nc.uiInputDebounced = false
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

func (nc *NetConnection) Accept(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
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

func (nc *NetConnection) Connect(server, port string) {
	nc.host = false
	nc.conn = nil // Make sure this is a new connection
	nc.remIn, nc.locIn = nc.GetHostGuestRemap()

	SafeGo(func() {
		d := net.Dialer{Timeout: 1 * time.Second}
		for {
			if nc.isClosing() {
				return
			}
			tempConn, err := d.Dial("tcp", server+":"+port)
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

func (nc *NetConnection) end() {
	if nc.st != NS_Error {
		nc.st = NS_End
	}
	nc.Close()
}

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

func (nc *NetConnection) Synchronize() error {
	if !nc.IsConnected() || nc.st == NS_Error {
		return Error("Cannot connect to the other player")
	}
	nc.Stop()

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
		binary.Write(nc.recording, binary.LittleEndian, &seed)
		binary.Write(nc.recording, binary.LittleEndian, &pmTime)
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
	<-nc.sendEnd
	go func(nb *NetBuffer) {
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
	}(&nc.buf[nc.locIn])

	// Start receiving inputs from remote peer in a goroutine
	<-nc.recvEnd
	go func(nb *NetBuffer) {
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
	}(&nc.buf[nc.remIn])

	// Update game state after synchronization
	nc.Update()

	// Log status
	log.Printf("Network synchronized: seed=%d pmTime=%d time=%d host=%v", seed, pmTime, nc.time, nc.host)

	return nil
}

func (nc *NetConnection) Update() bool {
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
				if nc.recording != nil {
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

type ReplayFile struct {
	file         *os.File
	ibit         [REPLAY_NUM_INPUTS]InputBits
	iaxes        [REPLAY_NUM_INPUTS][6]int8
	preMatchTime int32
}

func OpenReplayFile(filename string) *ReplayFile {
	rf, err := os.Open(filename)
	if err != nil {
		log.Printf("Failed to open replay file %s: %v", filename, err)
		return nil
	}
	log.Printf("Replay file opened: %s", filename)
	return &ReplayFile{file: rf}
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

// Read system variables from replay file
func (rf *ReplayFile) Synchronize() {
	if rf.file != nil {
		// Read random seed
		var seed int32
		if err := binary.Read(rf.file, binary.LittleEndian, &seed); err == nil {
			Srand(seed)
		}

		// Read pre-match time
		var pmTime int32
		if err := binary.Read(rf.file, binary.LittleEndian, &pmTime); err == nil {
			rf.preMatchTime = pmTime
			// Advance first frame
			rf.Update()
		}

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
