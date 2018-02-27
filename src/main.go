package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"syscall"
	"regexp"
)

func init() {
	runtime.LockOSThread()
}
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
func createLog(p string) *os.File {
	//fmt.Println("Creating log")
	f, err := os.Create(p)
	if err != nil {
		panic(err)
	}
	return f
}
func closeLog(f *os.File) {
	//fmt.Println("Closing log")
	f.Close()
}
func main() {
	if len(os.Args[1:]) > 0 {
		sys.cmdFlags = make(map[string]string)
		key := ""
		player := 1
		for _, a := range os.Args[1:] {
			match, _ := regexp.MatchString("^-", a)
			if match {
				help, _ := regexp.MatchString("^-[h%?]", a)
				if help {
					modkernel32 := syscall.NewLazyDLL("kernel32.dll")
					procAllocConsole := modkernel32.NewProc("AllocConsole")
					syscall.Syscall(procAllocConsole.Addr(), 0, 0, 0, 0)
					hout, err1 := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
					hin, err2 := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
					if err1 != nil || err2 != nil { // nowhere to print the message
						os.Exit(2)
					}
					os.Stdout = os.NewFile(uintptr(hout), "/dev/stdout")
					os.Stdin = os.NewFile(uintptr(hin), "/dev/stdin")
					fmt.Println("I.K.E.M.E.N\nOptions (case sensitive):")
					fmt.Println(" -h -?               Help")
					fmt.Println(" -log <logfile>      Records match data to <logfile>")
					fmt.Println(" -r <sysfile>        Loads motif <sysfile>. eg. -r motifdir or -r motifdir/system.def")
					fmt.Println("\nQuick VS Options:")
					fmt.Println(" -p<n> <playername>  Loads player n, eg. -p3 kfm")
					fmt.Println(" -p<n>.ai <level>    Set player n's AI to <level>, eg. -p1.ai 8")
					fmt.Println(" -p<n>.color <col>   Set player n's color to <col>")
					fmt.Println(" -p<n>.life <life>   Sets player n's life to <life>")
					fmt.Println(" -p<n>.power <power> Sets player n's power to <power>")
					fmt.Println(" -rounds <num>       Plays for <num> rounds, and then quits")
					fmt.Println(" -s <stagename>      Loads stage <stagename>")
					fmt.Println("\nPress ENTER to exit.")
					var s string
					fmt.Scanln(&s)
					os.Exit(0)
				}
				sys.cmdFlags[a] = ""
				key = a
			} else if key == "" {
				sys.cmdFlags[fmt.Sprintf("-p%v", player)] = a
				player += 1
			} else {
				sys.cmdFlags[key] = a
				key = ""
			}
		}
	}
	chk(glfw.Init())
	defer glfw.Terminate()
	defcfg := []byte(strings.Join(strings.Split(`{
  "HelperMax":56,
  "PlayerProjectileMax":50,
  "ExplodMax":256,
  "AfterImageMax":8,
  "Attack.LifeToPowerMul":0.7,
  "GetHit.LifeToPowerMul":0.6,
  "Width":640,
  "Height":480,
  "Super.TargetDefenceMul":1.5,
  "LifebarFontScale":1,
  "System":"script/main.lua",
  "KeyConfig":[{
      "Joystick":-1,
      "Buttons":["UP","DOWN","LEFT","RIGHT","z","x","c","a","s","d","RETURN","q","w"]
    },{
      "Joystick":-1,
      "Buttons":["t","g","f","h","j","k","l","u","i","o","RSHIFT","LEFTBRACKET","RIGHTBRACKET"]
    }],
  "_comment":{
    "_comment":"ジョイスティック (0番) の場合の KeyConfig",
    "KeyConfig":[{
        "Joystick":0,
        "Buttons":["-7","-8","-5","-6","0","1","4","2","3","5","7","6","8"]
      },{
        "Joystick":1,
        "Buttons":["-7","-8","-5","-6","0","1","4","2","3","5","7","6","8"]
      }]
  },
  "Motif":"data/system.def",
  "CommonAir":"data/common.air",
  "CommonCmd":"data/common.cmd",
  "SimulMode":true,
  "LifeMul":100,
  "Team1VS2Life":120,
  "TurnsRecoveryRate":300,
  "ZoomActive":true,
  "ZoomMin":0.75,
  "ZoomMax":1.1,
  "ZoomSpeed":1,
  "RoundTime":99,
  "NumTurns":4,
  "NumSimul":4,
  "NumTag":4,
  "Difficulty":8,
  "Credits":10,
  "ListenPort":7500,
  "ContSelection":true,
  "AIRandomColor":true,
  "AIRamping":true,
  "AutoGuard":false,
  "TeamPowerShare":false,
  "TeamLifeShare":false,
  "Fullscreen":false,
  "AllowDebugKeys":true,
  "IP":{
  }
}
`, "\n"), "\r\n"))
	tmp := struct {
		HelperMax              int32
		PlayerProjectileMax    int
		ExplodMax              int
		AfterImageMax          int32
		Attack_LifeToPowerMul  float32 `json:"Attack.LifeToPowerMul"`
		GetHit_LifeToPowerMul  float32 `json:"GetHit.LifeToPowerMul"`
		Width                  int32
		Height                 int32
		Super_TargetDefenceMul float32 `json:"Super.TargetDefenceMul"`
		LifebarFontScale       float32
		System                 string
		KeyConfig              []struct {
			Joystick int
			Buttons  []interface{}
		}
		NumTag         int
		TeamLifeShare  bool
		AIRandomColor  bool
		Fullscreen     bool
		AllowDebugKeys bool
		CommonAir      string
		CommonCmd      string
	}{}
	chk(json.Unmarshal(defcfg, &tmp))
	const configFile = "data/config.json"
	if bytes, err := ioutil.ReadFile(configFile); err != nil {
		f, err := os.Create(configFile)
		chk(err)
		f.Write(defcfg)
		chk(f.Close())
	} else {
		if len(bytes) >= 3 &&
			bytes[0] == 0xef && bytes[1] == 0xbb && bytes[2] == 0xbf {
			bytes = bytes[3:]
		}
		chk(json.Unmarshal(bytes, &tmp))
	}
	sys.helperMax = tmp.HelperMax
	sys.playerProjectileMax = tmp.PlayerProjectileMax
	sys.explodMax = tmp.ExplodMax
	sys.afterImageMax = tmp.AfterImageMax
	sys.attack_LifeToPowerMul = tmp.Attack_LifeToPowerMul
	sys.getHit_LifeToPowerMul = tmp.GetHit_LifeToPowerMul
	sys.super_TargetDefenceMul = tmp.Super_TargetDefenceMul
	sys.lifebarFontScale = tmp.LifebarFontScale
	stoki := func(key string) int {
		return int(StringToKey(key))
	}
	for a := 0; a < tmp.NumTag; a++ {
		for _, kc := range tmp.KeyConfig {
			b := kc.Buttons
			if kc.Joystick >= 0 {
				sys.keyConfig = append(sys.keyConfig, KeyConfig{kc.Joystick,
					int(b[0].(float64)), int(b[1].(float64)),
					int(b[2].(float64)), int(b[3].(float64)),
					int(b[4].(float64)), int(b[5].(float64)), int(b[6].(float64)),
					int(b[7].(float64)), int(b[8].(float64)), int(b[9].(float64)),
					int(b[10].(float64)), int(b[11].(float64)), int(b[12].(float64))})
			} else {
				sys.keyConfig = append(sys.keyConfig, KeyConfig{kc.Joystick,
					stoki(b[0].(string)), stoki(b[1].(string)),
					stoki(b[2].(string)), stoki(b[3].(string)),
					stoki(b[4].(string)), stoki(b[5].(string)), stoki(b[6].(string)),
					stoki(b[7].(string)), stoki(b[8].(string)), stoki(b[9].(string)),
					stoki(b[10].(string)), stoki(b[11].(string)), stoki(b[12].(string))})
			}
		}
	}
	sys.teamLifeShare = tmp.TeamLifeShare
	sys.fullscreen = tmp.Fullscreen
	sys.aiRandomColor = tmp.AIRandomColor
	sys.allowDebugKeys = tmp.AllowDebugKeys
	air, err := ioutil.ReadFile(tmp.CommonAir)
	if err != nil {
		fmt.Print(err)
	}
	sys.commonAir = string("\n") + string(air)
	cmd, err := ioutil.ReadFile(tmp.CommonCmd)
	if err != nil {
		fmt.Print(err)
	}
	sys.commonCmd = string("\n") + string(cmd)
	os.Mkdir("debug", os.ModeSticky|0755)
	log := createLog("debug/log.txt")
	defer closeLog(log)
	l := sys.init(tmp.Width, tmp.Height)
	if err := l.DoFile(tmp.System); err != nil {
		fmt.Fprintln(log, err)
		switch err.(type) {
		case *lua.ApiError:
			errstr := strings.Split(err.Error(), "\n")[0]
			if len(errstr) < 10 || errstr[len(errstr)-10:] != "<game end>" {
				panic(err)
			}
		default:
			panic(err)
		}
	}
	if !sys.gameEnd {
		sys.gameEnd = true
	}
	<-sys.audioClose
}
