package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"
	lua "github.com/yuin/gopher-lua"
	"github.com/sqweek/dialog"
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
	os.Mkdir("save", os.ModeSticky|0755)
	os.Mkdir("save/replays", os.ModeSticky|0755)
	if len(os.Args[1:]) > 0 {
		sys.cmdFlags = make(map[string]string)
		key := ""
		player := 1
		for _, a := range os.Args[1:] {
			match, _ := regexp.MatchString("^-", a)
			if match {
				help, _ := regexp.MatchString("^-[h%?]", a)
				if help {
					fmt.Println("I.K.E.M.E.N\nOptions (case sensitive):")
					fmt.Println(" -h -?                      Help")
					fmt.Println(" -log <logfile>             Records match data to <logfile>")
					fmt.Println(" -r <sysfile>               Loads motif <sysfile>. eg. -r motifdir or -r motifdir/system.def")
					fmt.Println("\nQuick VS Options:")
					fmt.Println(" -p<n> <playername>         Loads player n, eg. -p3 kfm")
					fmt.Println(" -p<n>.ai <level>           Set player n's AI to <level>, eg. -p1.ai 8")
					fmt.Println(" -p<n>.color <col>          Set player n's color to <col>")
					fmt.Println(" -p<n>.power <power>        Sets player n's power to <power>")
					fmt.Println(" -p<n>.life <life>          Sets player n's life to <life>")
					fmt.Println(" -p<n>.lifeMax <life>       Sets player n's max life to <life>")
					fmt.Println(" -p<n>.lifeRatio <ratio>    Sets player n's life ratio to <ratio>")
					fmt.Println(" -p<n>.attackRatio <ratio>  Sets player n's attack ratio to <ratio>")
					fmt.Println(" -rounds <num>              Plays for <num> rounds, and then quits")
					fmt.Println(" -s <stagename>             Loads stage <stagename>")
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
	if _, err := ioutil.ReadFile("save/stats.json"); err != nil {
		f, err := os.Create("save/stats.json")
		chk(err)
		f.Write([]byte("{}"))
		chk(f.Close())
	}
	defcfg := []byte(strings.Join(strings.Split(
`{
	"WindowTitle": "Ikemen GO",
	"HelperMax": 56,
	"PlayerProjectileMax": 256,
	"ExplodMax": 512,
	"AfterImageMax": 128,
	"MasterVolume": 80,
	"WavVolume": 80,
	"BgmVolume": 80,
	"Attack.LifeToPowerMul": 0.7,
	"GetHit.LifeToPowerMul": 0.6,
	"Width": 640,
	"Height": 480,
	"Super.TargetDefenceMul": 1.5,
	"LifebarFontScale": 1,
	"System": "external/script/main.lua",
	"KeyConfig": [
		{
			"Joystick": -1,
			"Buttons": ["UP", "DOWN", "LEFT", "RIGHT", "z", "x", "c", "a", "s", "d", "RETURN", "q", "w"]
		},
		{
			"Joystick": -1,
			"Buttons": ["t", "g", "f", "h", "j", "k", "l", "u", "i", "o", "RSHIFT", "LEFTBRACKET", "RIGHTBRACKET"]
		}
	],
	"JoystickConfig": [
		{
			"Joystick": 0,
			"Buttons": ["-3", "-4", "-1", "-2", "0", "1", "4", "2", "3", "5", "7", "-10", "-12"]
		},
		{
			"Joystick": 1,
			"Buttons": ["-3", "-4", "-1", "-2", "0", "1", "4", "2", "3", "5", "7", "-10", "-12"]
		}
	],
	"ControllerStickSensitivity": 0.4,
	"XinputTriggerSensitivity": 0,
	"Motif": "data/system.def",
	"CommonAir": "data/common.air",
	"CommonCmd": "data/common.cmd",
	"SimulMode": true,
	"LifeMul": 100,
	"Team1VS2Life": 100,
	"TurnsRecoveryBase": 0,
	"TurnsRecoveryBonus": 20,
	"ZoomActive": false,
	"ZoomMin": 0.75,
	"ZoomMax": 1.1,
	"ZoomSpeed": 1,
	"RoundTime": 99,
	"RoundsNumSingle": -1,
	"RoundsNumTeam": -1,
	"MaxDrawGames": -2,
	"NumTurns": 4,
	"NumSimul": 4,
	"NumTag": 4,
	"Difficulty": 8,
	"Credits": 10,
	"ListenPort": 7500,
	"IP": {
		
	},
	"QuickContinue": false,
	"AIRandomColor": true,
	"AIRamping": true,
	"AutoGuard": false,
	"TeamPowerShare": false,
	"TeamLifeShare": false,
	"Fullscreen": false,
	"AudioDucking": false,
	"QuickLaunch": 0,
	"AllowDebugKeys": true,
	"ComboExtraFrameWindow": 1,
	"ExternalShaders": [],
	"LocalcoordScalingType": 1,
	"MSAA": false,
	"LifeRatio":[0.80, 1.0, 1.17, 1.40],
	"AttackRatio":[0.82, 1.0, 1.17, 1.30],
	"WindowMainIconLocation": [
		"external/icons/IkemenCylia.png"
	]
}
`, "\n"), "\r\n"))
	tmp := struct {
		WindowTitle            string
		HelperMax              int32
		PlayerProjectileMax    int
		ExplodMax              int
		AfterImageMax          int32
		MasterVolume           int
		WavVolume              int
		BgmVolume              int
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
		JoystickConfig         []struct {
			Joystick int
			Buttons  []interface{}
		}
		ControllerStickSensitivity float32
		XinputTriggerSensitivity   float32
		Motif                      string
		CommonAir                  string
		CommonCmd                  string
		SimulMode                  bool
		LifeMul                    float32
		Team1VS2Life               float32
		TurnsRecoveryBase          float32
		TurnsRecoveryBonus         float32
		ZoomActive                 bool
		ZoomMin                    float32
		ZoomMax                    float32
		ZoomSpeed                  float32
		RoundTime                  int32
		RoundsNumSingle            int32
		RoundsNumTeam              int32
		MaxDrawGames               int32
		NumTurns                   int
		NumSimul                   int
		NumTag                     int
		Difficulty                 int
		Credits                    int
		ListenPort                 int
		IP                         map[string]string
		QuickContinue              bool
		AIRandomColor              bool
		AIRamping                  bool
		AutoGuard                  bool
		TeamPowerShare             bool
		TeamLifeShare              bool
		Fullscreen                 bool
		PostProcessingShader       int32
		AudioDucking               bool
		QuickLaunch                int
		AllowDebugKeys             bool
		ComboExtraFrameWindow      int32
		ExternalShaders            []string
		LocalcoordScalingType      int32
		MSAA                       bool
		LifeRatio                  [4]float32
		AttackRatio                [4]float32
		WindowMainIconLocation     []string
	}{}
	chk(json.Unmarshal(defcfg, &tmp))
	if bytes, err := ioutil.ReadFile("save/config.json"); err == nil {
		if len(bytes) >= 3 &&
			bytes[0] == 0xef && bytes[1] == 0xbb && bytes[2] == 0xbf {
			bytes = bytes[3:]
		}
		chk(json.Unmarshal(bytes, &tmp))
	}
	cfg, err := json.MarshalIndent(tmp, "", "	")
	chk(err)
	chk(ioutil.WriteFile("save/config.json", cfg, 0644))
	sys.controllerStickSensitivity = tmp.ControllerStickSensitivity
	sys.xinputTriggerSensitivity = tmp.XinputTriggerSensitivity
	sys.windowTitle = tmp.WindowTitle
	sys.helperMax = tmp.HelperMax
	sys.playerProjectileMax = tmp.PlayerProjectileMax
	sys.explodMax = tmp.ExplodMax
	sys.afterImageMax = tmp.AfterImageMax
	sys.attack_LifeToPowerMul = tmp.Attack_LifeToPowerMul
	sys.getHit_LifeToPowerMul = tmp.GetHit_LifeToPowerMul
	sys.super_TargetDefenceMul = tmp.Super_TargetDefenceMul
	sys.comboExtraFrameWindow = tmp.ComboExtraFrameWindow
	sys.lifebarFontScale = tmp.LifebarFontScale
	sys.quickLaunch = tmp.QuickLaunch
	sys.windowMainIconLocation = tmp.WindowMainIconLocation
	sys.externalShaderList = tmp.ExternalShaders
	// For debug testing letting this here commented because it could be useful in the future.
	// log.Printf("Unmarshaled: %v", tmp.WindowMainIconLocation)
	sys.masterVolume = tmp.MasterVolume
	sys.wavVolume = tmp.WavVolume
	sys.bgmVolume = tmp.BgmVolume
	sys.AudioDucking = tmp.AudioDucking
	stoki := func(key string) int {
		return int(StringToKey(key))
	}
	Atoi := func(key string) int {
		var i int
		i, _ = strconv.Atoi(key)
		return i
	}
	Max := func(x, y int) int {
		if x < y {
			return y
		}
		return x
	}
	for a := 0; a < Max(tmp.NumSimul, tmp.NumTag); a++ {
		for _, kc := range tmp.KeyConfig {
			b := kc.Buttons
			if kc.Joystick < 0 {
				sys.keyConfig = append(sys.keyConfig, KeyConfig{kc.Joystick,
					stoki(b[0].(string)), stoki(b[1].(string)),
					stoki(b[2].(string)), stoki(b[3].(string)),
					stoki(b[4].(string)), stoki(b[5].(string)), stoki(b[6].(string)),
					stoki(b[7].(string)), stoki(b[8].(string)), stoki(b[9].(string)),
					stoki(b[10].(string)), stoki(b[11].(string)), stoki(b[12].(string))})
			}
		}
		for _, jc := range tmp.JoystickConfig {
			b := jc.Buttons
			if jc.Joystick >= 0 {
				sys.joystickConfig = append(sys.joystickConfig, KeyConfig{jc.Joystick,
					Atoi(b[0].(string)), Atoi(b[1].(string)),
					Atoi(b[2].(string)), Atoi(b[3].(string)),
					Atoi(b[4].(string)), Atoi(b[5].(string)), Atoi(b[6].(string)),
					Atoi(b[7].(string)), Atoi(b[8].(string)), Atoi(b[9].(string)),
					Atoi(b[10].(string)), Atoi(b[11].(string)), Atoi(b[12].(string))})
			}
		}
	}
	
	sys.teamLifeShare = tmp.TeamLifeShare
	sys.fullscreen = tmp.Fullscreen
	sys.PostProcessingShader = tmp.PostProcessingShader
	sys.MultisampleAntialiasing = tmp.MSAA
	sys.LocalcoordScalingType = tmp.LocalcoordScalingType
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
	//os.Mkdir("debug", os.ModeSticky|0755)
	log := createLog("Ikemen.log")
	defer closeLog(log)
	l := sys.init(tmp.Width, tmp.Height)
	if err := l.DoFile(tmp.System); err != nil {
		fmt.Fprintln(log, err)
		switch err.(type) {
		case *lua.ApiError:
			errstr := strings.Split(err.Error(), "\n")[0]
			if len(errstr) < 10 || errstr[len(errstr)-10:] != "<game end>" {
				dialog.Message("%s\n\nError saved to Ikemen.log", err).Title("I.K.E.M.E.N Error").Error()
				panic(err)
			}
		default:
			dialog.Message("%s\n\nError saved to Ikemen.log", err).Title("I.K.E.M.E.N Error").Error()
			panic(err)
		}
	}
	if !sys.gameEnd {
		sys.gameEnd = true
	}
	<-sys.audioClose
}
