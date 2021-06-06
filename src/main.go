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
	"github.com/sqweek/dialog"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	runtime.LockOSThread()
}

// Checks if error is not null, if there is an error it displays a error dialogue box and crashes the program.
func chk(err error) {
	if err != nil {
		dialog.Message(err.Error()).Title("I.K.E.M.E.N Error").Error()
		panic(err)
	}
}

// Extended version of 'chk()'
func chkEX(err error, txt string) {
	if err != nil {
		dialog.Message(txt + err.Error()).Title("I.K.E.M.E.N Error").Error()
		panic(Error(txt + err.Error()))
	}
}

func createLog(p string) *os.File {
	f, err := os.Create(p)
	if err != nil {
		panic(err)
	}
	return f
}
func closeLog(f *os.File) {
	f.Close()
}

func main() {
	// Make save directories, if they don't exist
	os.Mkdir("save", os.ModeSticky|0755)
	os.Mkdir("save/replays", os.ModeSticky|0755)

	processCommandLine()

	// Initialize OpenGL
	chk(glfw.Init())
	defer glfw.Terminate()

	// Try reading stats
	if _, err := ioutil.ReadFile("save/stats.json"); err != nil {
		// If there was an error reading, write an empty json file
		f, err := os.Create("save/stats.json")
		chk(err)
		f.Write([]byte("{}"))
		chk(f.Close())
	}

	// Setup config values, and get a reference to the config object for the main script and window size
	tmp := setupConfig()

	//os.Mkdir("debug", os.ModeSticky|0755)

	// Check if the main lua file exists.
	if !fileExists(tmp.System) {
		var err = Error("Main lua file '" + tmp.System + "' can not be found.")
		dialog.Message(err.Error()).Title("I.K.E.M.E.N Error").Error()
		panic(err)
	}

	log := createLog("Ikemen.log")
	defer closeLog(log)
	// Initialize game and create window
	sys.luaLState = sys.init(tmp.GameWidth, tmp.GameHeight)

	// Begin processing game using its lua scripts
	if err := sys.luaLState.DoFile(tmp.System); err != nil {
		// Display error logs.
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

	// Shutdown
	if !sys.gameEnd {
		sys.gameEnd = true
	}
	<-sys.audioClose
}

// Loops through given comand line arguments and processes them for later use by the game
func processCommandLine() {
	// If there are command line arguments
	if len(os.Args[1:]) > 0 {
		sys.cmdFlags = make(map[string]string)
		key := ""
		player := 1
		r1, _ := regexp.Compile("^-[h%?]")
		r2, _ := regexp.Compile("^-")
		// Loop through arguments
		for _, a := range os.Args[1:] {
			// If getting help about command line options
			if r1.MatchString(a) {
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
				// If a control argument starting with - (eg. -p3, -s, -rounds)
			} else if r2.MatchString(a) {
				// Set a blank value for the key to start with
				sys.cmdFlags[a] = ""
				// Prepare the key for the next argument
				key = a
				// If an argument with no key
			} else if key == "" {
				// Set p1/p2's name
				sys.cmdFlags[fmt.Sprintf("-p%v", player)] = a
				player += 1
				// If a key is prepared for this argument
			} else {
				// Set the argument for this key
				sys.cmdFlags[key] = a
				key = ""
			}
		}
	}
}

type configSettings struct {
	AIRamping                  bool
	AIRandomColor              bool
	AudioDucking               bool
	AudioSampleRate            int32
	AutoGuard                  bool
	BarGuard                   bool
	BarRedLife                 bool
	BarStun                    bool
	Borderless                 bool
	ComboExtraFrameWindow      int32
	CommonAir                  string
	CommonCmd                  string
	CommonConst                string
	CommonLua                  []string
	CommonStates               []string
	ControllerStickSensitivity float32
	Credits                    int
	DebugClipboardRows         int
	DebugConsoleRows           int
	DebugFont                  string
	DebugFontScale             float32
	DebugKeys                  bool
	DebugMode                  bool
	Difficulty                 int
	EscOpensMenu               bool
	ExternalShaders            []string
	FirstRun                   bool
	FontShaderVer              string
	ForceStageZoomin           float32
	ForceStageZoomout          float32
	Fullscreen                 bool
	GameWidth                  int32
	GameHeight                 int32
	GameSpeed                  float32
	Framerate                  int32
	IP                         map[string]string
	LifebarFontScale           float32
	LifeMul                    float32
	ListenPort                 string
	LoseSimul                  bool
	LoseTag                    bool
	MaxAfterImage              int32
	MaxDrawGames               int32
	MaxExplod                  int
	MaxHelper                  int32
	MaxPlayerProjectile        int
	Modules                    []string
	Motif                      string
	MSAA                       bool
	NumSimul                   [2]int
	NumTag                     [2]int
	NumTurns                   [2]int
	Players                    int
	PngSpriteFilter            bool
	PostProcessingShader       int32
	QuickContinue              bool
	RatioAttack                [4]float32
	RatioLife                  [4]float32
	RatioRecoveryBase          float32
	RatioRecoveryBonus         float32
	RoundsNumSimul             int32
	RoundsNumSingle            int32
	RoundsNumTag               int32
	RoundTime                  int32
	ScreenshotFolder           string
	StartStage                 string
	System                     string
	Team1VS2Life               float32
	TeamDuplicates             bool
	TeamLifeShare              bool
	TeamPowerShare             bool
	TrainingChar               string
	TurnsRecoveryBase          float32
	TurnsRecoveryBonus         float32
	VolumeBgm                  int
	VolumeMaster               int
	VolumeSfx                  int
	VolumeWarning              bool
	MaxBgmVolume               int
	VRetrace                   int
	WindowIcon                 []string
	WindowTitle                string
	XinputTriggerSensitivity   float32
	ZoomActive                 bool
	ZoomDelay                  bool
	ZoomSpeed                  float32
	KeyConfig                  []struct {
		Joystick int
		Buttons  []interface{}
	}
	JoystickConfig []struct {
		Joystick int
		Buttons  []interface{}
	}
}

// Sets default config settings, then attemps to load existing config from disk
func setupConfig() configSettings {
	// Default Config
	defcfg := []byte(strings.Join(strings.Split(
		`{
	"AIRamping": true,
	"AIRandomColor": false,
	"AudioDucking": false,
	"AudioSampleRate": 44100,
	"AutoGuard": false,
	"BarGuard": false,
	"BarRedLife": true,
	"BarStun": false,
	"Borderless": false,
	"ComboExtraFrameWindow": 1,
	"CommonAir": "data/common.air",
	"CommonCmd": "data/common.cmd",
	"CommonConst": "data/common.const",
	"CommonLua": [
		"loop()"
    ],
	"CommonStates": [
		"data/action.zss",
		"data/dizzy.zss",
		"data/guardbreak.zss",
		"data/rank.zss",
		"data/score.zss",
		"data/tag.zss"
    ],
	"ControllerStickSensitivity": 0.4,
	"Credits": 10,
	"DebugClipboardRows": 2,
	"DebugConsoleRows": 15,
	"DebugFont": "font/default-3x5-bold.def",
	"DebugFontScale": 1,
	"DebugKeys": true,
	"DebugMode": true,
	"Difficulty": 8,
	"EscOpensMenu": true,
	"ExternalShaders": [],
	"FirstRun": true,
	"FontShaderVer": "150 core",
	"ForceStageZoomin": 0,
	"ForceStageZoomout": 0,
	"Framerate": 60,
	"Fullscreen": false,
	"GameWidth": 640,
	"GameHeight": 480,
	"GameSpeed": 100,
	"IP": {},
	"LifebarFontScale": 1,
	"LifeMul": 100,
	"ListenPort": "7500",
	"LoseSimul": true,
	"LoseTag": false,
	"MaxAfterImage": 128,
	"MaxDrawGames": -2,
	"MaxExplod": 512,
	"MaxHelper": 56,
	"MaxPlayerProjectile": 256,
	"Modules": [],
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
	"Players": 4,
	"PngSpriteFilter": true,
	"PostProcessingShader": 0,
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
	"RatioRecoveryBase": 0,
	"RatioRecoveryBonus": 20,
	"RoundsNumSimul": 2,
	"RoundsNumSingle": 2,
	"RoundsNumTag": 2,
	"RoundTime": 99,
	"ScreenshotFolder": "",
	"StartStage" : "stages/stage0-720.def",
	"System": "external/script/main.lua",
	"Team1VS2Life": 150,
	"TeamDuplicates": true,
	"TeamLifeShare": false,
	"TeamPowerShare": true,
	"TrainingChar": "chars/training/training.def",
	"TurnsRecoveryBase": 0,
	"TurnsRecoveryBonus": 20,
	"VolumeBgm": 80,
	"VolumeMaster": 80,
	"VolumeSfx": 80,
	"VolumeWarning": true,
	"MaxBgmVolume": 0,
	"VRetrace": 1, 
	"WindowIcon": [
		"external/icons/IkemenCylia_256.png",
		"external/icons/IkemenCylia_96.png",
		"external/icons/IkemenCylia_48.png"
	],
	"WindowTitle": "Ikemen GO",
	"XinputTriggerSensitivity": 0,
	"ZoomActive": true,
	"ZoomDelay": false,
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
				"w",
				"Not used"
			]
		},
		{
			"Joystick": -1,
			"Buttons": [
				"i",
				"k",
				"j",
				"l",
				"f",
				"g",
				"h",
				"r",
				"t",
				"y",
				"RSHIFT",
				"LEFTBRACKET",
				"RIGHTBRACKET",
				"Not used"
			]
		},
		{
			"Joystick": -1,
			"Buttons": [
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used"
			]
		},
		{
			"Joystick": -1,
			"Buttons": [
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used",
				"Not used"
			]
		}
	],
	"JoystickConfig": [
		{
			"Joystick": 0,
			"Buttons": [
				"10",
				"12",
				"13",
				"11",
				"0",
				"1",
				"4",
				"2",
				"3",
				"5",
				"7",
				"-10",
				"-12",
				"6"
			]
		},
		{
			"Joystick": 1,
			"Buttons": [
				"10",
				"12",
				"13",
				"11",
				"0",
				"1",
				"4",
				"2",
				"3",
				"5",
				"7",
				"-10",
				"-12",
				"6"
			]
		},
		{
			"Joystick": 2,
			"Buttons": [
				"10",
				"12",
				"13",
				"11",
				"0",
				"1",
				"4",
				"2",
				"3",
				"5",
				"7",
				"-10",
				"-12",
				"6"
			]
		},
		{
			"Joystick": 3,
			"Buttons": [
				"10",
				"12",
				"13",
				"11",
				"0",
				"1",
				"4",
				"2",
				"3",
				"5",
				"7",
				"-10",
				"-12",
				"6"
			]
		}
	]
}
`, "\n"), "\r\n"))

	// Unmarshal default config string into a struct
	tmp := configSettings{}
	chk(json.Unmarshal(defcfg, &tmp))

	// Config file path
	cfgPath := "save/config.json"
	// If a different config file is defined in the command line parameters, use it instead
	if _, ok := sys.cmdFlags["-config"]; ok {
		cfgPath = sys.cmdFlags["-config"]
	}
	// Load the config file, overwriting the defaults
	if bytes, err := ioutil.ReadFile(cfgPath); err == nil {
		if len(bytes) >= 3 &&
			bytes[0] == 0xef && bytes[1] == 0xbb && bytes[2] == 0xbf {
			bytes = bytes[3:]
		}
		chkEX(json.Unmarshal(bytes, &tmp), "Error while loading the config file.\n")
	}

	cfg, _ := json.MarshalIndent(tmp, "", "	")
	chk(ioutil.WriteFile(cfgPath, cfg, 0644))

	// Set each config property to the system object
	sys.afterImageMax = tmp.MaxAfterImage
	sys.allowDebugKeys = tmp.DebugKeys
	sys.allowDebugMode = tmp.DebugMode
	sys.audioDucking = tmp.AudioDucking
	switch tmp.AudioSampleRate {
	case 44100, 48000:
		Mp3SampleRate = int(tmp.AudioSampleRate)
	default:
		Mp3SampleRate = 44100
	}
	sys.bgmVolume = tmp.VolumeBgm
	//sys.volumeWarning = tmp.VolumeWarning
	if tmp.MaxBgmVolume <= 400 {
		sys.maxBgmVolume = tmp.MaxBgmVolume
	}
	sys.borderless = tmp.Borderless
	sys.cam.ZoomDelayEnable = tmp.ZoomDelay
	sys.cam.ZoomEnable = tmp.ZoomActive
	sys.cam.ZoomMax = tmp.ForceStageZoomin
	sys.cam.ZoomMin = tmp.ForceStageZoomout
	sys.cam.ZoomSpeed = 12 - tmp.ZoomSpeed
	sys.comboExtraFrameWindow = tmp.ComboExtraFrameWindow
	if air, err := ioutil.ReadFile(tmp.CommonAir); err == nil {
		sys.commonAir = "\n" + string(air)
	}
	if cmd, err := ioutil.ReadFile(tmp.CommonCmd); err == nil {
		sys.commonCmd = "\n" + string(cmd)
	}
	sys.commonConst = tmp.CommonConst
	sys.commonLua = tmp.CommonLua
	sys.commonStates = tmp.CommonStates
	sys.clipboardRows = tmp.DebugClipboardRows
	sys.consoleRows = tmp.DebugConsoleRows
	sys.controllerStickSensitivity = tmp.ControllerStickSensitivity
	sys.explodMax = tmp.MaxExplod
	sys.externalShaderList = tmp.ExternalShaders
	if len(tmp.FontShaderVer) > 1 {
		sys.fontShaderVer = tmp.FontShaderVer
	}
	sys.fullscreen = tmp.Fullscreen

	if tmp.Framerate > 0 || tmp.Framerate <= 840 {
		FPS = int(tmp.Framerate)
	}
	sys.gameSpeed = tmp.GameSpeed / 100
	sys.helperMax = tmp.MaxHelper
	sys.lifebarFontScale = tmp.LifebarFontScale
	sys.lifeMul = tmp.LifeMul / 100
	sys.lifeShare = [...]bool{tmp.TeamLifeShare, tmp.TeamLifeShare}
	sys.listenPort = tmp.ListenPort
	sys.loseSimul = tmp.LoseSimul
	sys.loseTag = tmp.LoseTag
	sys.masterVolume = tmp.VolumeMaster
	sys.multisampleAntialiasing = tmp.MSAA
	sys.playerProjectileMax = tmp.MaxPlayerProjectile
	sys.postProcessingShader = tmp.PostProcessingShader
	sys.pngFilter = tmp.PngSpriteFilter
	sys.powerShare = [...]bool{tmp.TeamPowerShare, tmp.TeamPowerShare}
	tmp.ScreenshotFolder = strings.TrimSpace(tmp.ScreenshotFolder)
	if tmp.ScreenshotFolder != "" {
		tmp.ScreenshotFolder = strings.Replace(tmp.ScreenshotFolder, "\\", "/", -1)
		tmp.ScreenshotFolder = strings.TrimRight(tmp.ScreenshotFolder, "/")
		sys.screenshotFolder = tmp.ScreenshotFolder + "/"
	} else {
		sys.screenshotFolder = tmp.ScreenshotFolder
	}
	sys.team1VS2Life = tmp.Team1VS2Life / 100
	sys.vRetrace = tmp.VRetrace
	sys.wavVolume = tmp.VolumeSfx
	sys.windowMainIconLocation = tmp.WindowIcon
	sys.windowTitle = tmp.WindowTitle
	sys.xinputTriggerSensitivity = tmp.XinputTriggerSensitivity
	stoki := func(key string) int {
		return int(StringToKey(key))
	}
	Atoi := func(key string) int {
		if i, err := strconv.Atoi(key); err == nil {
			return i
		}
		return 999
	}
	for _, kc := range tmp.KeyConfig {
		b := kc.Buttons
		sys.keyConfig = append(sys.keyConfig, KeyConfig{kc.Joystick,
			stoki(b[0].(string)), stoki(b[1].(string)), stoki(b[2].(string)),
			stoki(b[3].(string)), stoki(b[4].(string)), stoki(b[5].(string)),
			stoki(b[6].(string)), stoki(b[7].(string)), stoki(b[8].(string)),
			stoki(b[9].(string)), stoki(b[10].(string)), stoki(b[11].(string)),
			stoki(b[12].(string)), stoki(b[13].(string))})
	}
	if _, ok := sys.cmdFlags["-nojoy"]; !ok {
		for _, jc := range tmp.JoystickConfig {
			b := jc.Buttons
			sys.joystickConfig = append(sys.joystickConfig, KeyConfig{jc.Joystick,
				Atoi(b[0].(string)), Atoi(b[1].(string)), Atoi(b[2].(string)),
				Atoi(b[3].(string)), Atoi(b[4].(string)), Atoi(b[5].(string)),
				Atoi(b[6].(string)), Atoi(b[7].(string)), Atoi(b[8].(string)),
				Atoi(b[9].(string)), Atoi(b[10].(string)), Atoi(b[11].(string)),
				Atoi(b[12].(string)), Atoi(b[13].(string))})
		}
	}

	return tmp
}

// fileExists checks if a file exists and is not a directory before we use it
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
