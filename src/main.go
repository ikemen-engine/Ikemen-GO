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
					text := `Options (case sensitive):
-h -?                   Help
-log <logfile>          Records match data to <logfile>
-r <path>               Loads motif <path>. eg. -r motifdir or -r motifdir/system.def
-lifebar <path>         Loads lifebar <path>. eg. -lifebar data/fight.def
-storyboard <path>      Loads storyboard <path>. eg. -storyboard chars/kfm/intro.def

Quick VS Options:
-p<n> <playername>      Loads player n, eg. -p3 kfm
-p<n>.ai <level>        Sets player n's AI to <level>, eg. -p1.ai 8
-p<n>.color <col>       Sets player n's color to <col>
-p<n>.power <power>     Sets player n's power to <power>
-p<n>.life <life>       Sets player n's life to <life>
-tmode1 <tmode>         Sets p1 team mode to <tmode>
-tmode2 <tmode>         Sets p2 team mode to <tmode>
-time <num>             Round time (-1 to disable)
-rounds <num>           Plays for <num> rounds, and then quits
-s <stagename>          Loads stage <stagename>

Debug Options:
-nojoy                  Disables joysticks
-nomusic                Disables music
-nosound                Disables all sound effects and music
-togglelifebars         Disables display of the Life and Power bars
-maxpowermode           Enables auto-refill of Power bars
-ailevel <level>        Changes game difficulty setting to <level> (1-8)
-speed <speed>          Changes game speed setting to <speed> (10%%-200%%)
-stresstest <frameskip> Stability test (AI matches at speed increased by <frameskip>)
-speedtest              Speed test (match speed x100)`
					//dialog.Message(text).Title("I.K.E.M.E.N Command line options").Info()
					fmt.Printf("I.K.E.M.E.N Command line options\n\n" + text + "\nPress ENTER to exit")
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
	"AIRamping": true,
	"AIRandomColor": true,
	"AudioDucking": false,
	"AutoGuard": false,
	"Borderless": false,
	"ComboExtraFrameWindow": 1,
	"CommonAir": "data/common.air",
	"CommonCmd": "data/common.cmd",
	"CommonConst": "data/common.const",
	"CommonRules": "data/rules.zss",
	"CommonScore": "data/score.zss",
	"CommonTag": "data/tag.zss",
	"ControllerStickSensitivity": 0.4,
	"Credits": 10,
	"DebugKeys": true,
	"DebugMode": false,
	"Difficulty": 8,
	"ExternalShaders": [],
	"Fullscreen": false,
	"GameWidth": 640,
	"GameHeight": 480,
	"GameSpeed": 100,
	"GaugeGuard": false,
	"GaugeStun": false,
	"IP": {},
	"LifebarFontScale": 1,
	"LifeMul": 100,
	"ListenPort": "7500",
	"LocalcoordScalingType": 1,
	"MaxAfterImage": 128,
	"MaxDrawGames": -2,
	"MaxExplod": 512,
	"MaxHelper": 56,
	"MaxPlayerProjectile": 256,
	"Motif": "data/system.def",
	"MSAA": false,
	"NumSimul": [
		2,
		4
	],
	"NumTag": [
		2,
		4
	],
	"NumTurns": [
		2,
		4
	],
	"PostProcessingShader": 0,
	"PreloadingSmall": true,
	"PreloadingBig": true,
	"PreloadingVersus": true,
	"PreloadingStage": true,
	"QuickContinue": false,
	"RatioAttack": [
		0.82,
		1,
		1.17,
		1.3
	],
	"RatioLife": [
		0.8,
		1,
		1.17,
		1.4
	],
	"RoundsNumSingle": 2,
	"RoundsNumTeam": 2,
	"RoundTime": 99,
	"SimulLoseKO": true,
	"SingleVsTeamLife": 100,
	"System": "external/script/main.lua",
	"TagLoseKO": false,
	"TeamLifeAdjustment": false,
	"TeamPowerShare": true,
	"TurnsRecoveryBase": 0,
	"TurnsRecoveryBonus": 20,
	"VolumeBgm": 80,
	"VolumeMaster": 80,
	"VolumeSfx": 80,
	"VRetrace": 1, 
	"WindowMainIconLocation": [
		"external/icons/IkemenCylia.png"
	],
	"WindowTitle": "Ikemen GO",
	"XinputTriggerSensitivity": 0,
	"ZoomActive": true,
	"ZoomMax": 1,
	"ZoomMin": 1,
	"ZoomSpeed": 1,
	"KeyConfig": [
		{
			"Joystick": -1,
			"Buttons": [
				"UP",
				"DOWN",
				"LEFT",
				"RIGHT",
				"z",
				"x",
				"c",
				"a",
				"s",
				"d",
				"RETURN",
				"q",
				"w"
			]
		},
		{
			"Joystick": -1,
			"Buttons": [
				"t",
				"g",
				"f",
				"h",
				"j",
				"k",
				"l",
				"u",
				"i",
				"o",
				"RSHIFT",
				"LEFTBRACKET",
				"RIGHTBRACKET"
			]
		}
	],
	"JoystickConfig": [
		{
			"Joystick": 0,
			"Buttons": [
				"-3",
				"-4",
				"-1",
				"-2",
				"0",
				"1",
				"4",
				"2",
				"3",
				"5",
				"7",
				"-10",
				"-12"
			]
		},
		{
			"Joystick": 1,
			"Buttons": [
				"-3",
				"-4",
				"-1",
				"-2",
				"0",
				"1",
				"4",
				"2",
				"3",
				"5",
				"7",
				"-10",
				"-12"
			]
		}
	]
}
`, "\n"), "\r\n"))
	tmp := struct {
		AIRamping                  bool
		AIRandomColor              bool
		AudioDucking               bool
		AutoGuard                  bool
		Borderless                 bool
		ComboExtraFrameWindow      int32
		CommonAir                  string
		CommonCmd                  string
		CommonConst                string
		CommonRules                string
		CommonScore                string
		CommonTag                  string
		ControllerStickSensitivity float32
		Credits                    int
		DebugKeys                  bool
		DebugMode                  bool
		Difficulty                 int
		ExternalShaders            []string
		Fullscreen                 bool
		GameWidth                  int32
		GameHeight                 int32
		GameSpeed                  float32
		GaugeGuard                 bool
		GaugeStun                  bool
		IP                         map[string]string
		LifebarFontScale           float32
		LifeMul                    float32
		ListenPort                 string
		LocalcoordScalingType      int32
		MaxAfterImage              int32
		MaxDrawGames               int32
		MaxExplod                  int
		MaxHelper                  int32
		MaxPlayerProjectile        int
		Motif                      string
		MSAA                       bool
		NumSimul                   [2]int
		NumTag                     [2]int
		NumTurns                   [2]int
		PostProcessingShader       int32
		PreloadingSmall            bool
		PreloadingBig              bool
		PreloadingVersus           bool
		PreloadingStage            bool
		QuickContinue              bool
		RatioAttack                [4]float32
		RatioLife                  [4]float32
		RoundsNumSingle            int32
		RoundsNumTeam              int32
		RoundTime                  int32
		SimulLoseKO                bool
		SingleVsTeamLife           float32
		System                     string
		TagLoseKO                  bool
		TeamLifeAdjustment         bool
		TeamPowerShare             bool
		TurnsRecoveryBase          float32
		TurnsRecoveryBonus         float32
		VolumeBgm                  int
		VolumeMaster               int
		VolumeSfx                  int
		VRetrace                   int
		WindowMainIconLocation     []string
		WindowTitle                string
		XinputTriggerSensitivity   float32
		ZoomActive                 bool
		ZoomMax                    float32
		ZoomMin                    float32
		ZoomSpeed                  float32
		KeyConfig                  []struct {
			Joystick int
			Buttons  []interface{}
		}
		JoystickConfig             []struct {
			Joystick int
			Buttons  []interface{}
		}
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
	chk(ioutil.WriteFile("save/config.json", cfg, 0644))
	sys.audioDucking = tmp.AudioDucking
	sys.comboExtraFrameWindow = tmp.ComboExtraFrameWindow
	air, err := ioutil.ReadFile(tmp.CommonAir); if err == nil {
		sys.commonAir = "\n" + string(air)
	}
	cmd, err := ioutil.ReadFile(tmp.CommonCmd); if err == nil {
		sys.commonCmd = "\n" + string(cmd)
	}
	sys.commonConst = tmp.CommonConst
	sys.commonRules = tmp.CommonRules
	sys.commonScore = tmp.CommonScore
	sys.commonTag = tmp.CommonTag
	sys.controllerStickSensitivity = tmp.ControllerStickSensitivity
	sys.allowDebugKeys = tmp.DebugKeys
	sys.debugDraw = tmp.DebugMode
	sys.externalShaderList = tmp.ExternalShaders
	sys.fullscreen = tmp.Fullscreen
	sys.borderless = tmp.Borderless
	sys.lifebarFontScale = tmp.LifebarFontScale
	sys.listenPort = tmp.ListenPort
	sys.LocalcoordScalingType = tmp.LocalcoordScalingType
	sys.helperMax = tmp.MaxHelper
	sys.playerProjectileMax = tmp.MaxPlayerProjectile
	sys.explodMax = tmp.MaxExplod
	sys.afterImageMax = tmp.MaxAfterImage
	sys.MultisampleAntialiasing = tmp.MSAA
	sys.PostProcessingShader = tmp.PostProcessingShader
	sys.preloading.small = tmp.PreloadingSmall
	sys.preloading.big = tmp.PreloadingBig
	sys.preloading.versus = tmp.PreloadingVersus
	sys.preloading.stage = tmp.PreloadingStage
	sys.lifeAdjustment = tmp.TeamLifeAdjustment
	sys.bgmVolume = tmp.VolumeBgm
	sys.masterVolume = tmp.VolumeMaster
	sys.wavVolume = tmp.VolumeSfx
	sys.windowMainIconLocation = tmp.WindowMainIconLocation
	sys.windowTitle = tmp.WindowTitle
	sys.vRetrace = tmp.VRetrace
	sys.xinputTriggerSensitivity = tmp.XinputTriggerSensitivity
	sys.cam.ZoomEnable = tmp.ZoomActive
	sys.cam.ZoomMax = tmp.ZoomMax
	sys.cam.ZoomMin = tmp.ZoomMin
	sys.cam.ZoomSpeed = 12 - tmp.ZoomSpeed
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
	for a := 0; a < Max(tmp.NumSimul[1], tmp.NumTag[1]); a++ {
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
		if _, ok := sys.cmdFlags["-nojoy"]; !ok {
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
	}
	//os.Mkdir("debug", os.ModeSticky|0755)
	log := createLog("Ikemen.log")
	defer closeLog(log)
	l := sys.init(tmp.GameWidth, tmp.GameHeight)
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
