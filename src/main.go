package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	lua "github.com/yuin/gopher-lua"
)

var Version = "development"
var BuildTime = "" // Set automatically by GitHub Actions

func init() {
	if runtime.GOOS != "android" {
		runtime.LockOSThread()
	}
}

// Checks if error is not null, if there is an error it displays a error dialogue box and crashes the program.
func chk(err error) {
	if err != nil {
		ShowErrorDialog(err.Error())
		panic(err)
	}
}

// Extended version of 'chk()'
func chkEX(err error, txt string, crash bool) bool {
	if err != nil {
		ShowErrorDialog(txt + err.Error())
		if crash {
			panic(Error(txt + err.Error()))
		}
		return true
	}
	return false
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
	realMain()
}

func realMain() {
	if runtime.GOOS == "android" {
		Logcat("Inside realMain...")
		runtime.LockOSThread()
		sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_ES)
		sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
		sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 2)
		sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
		sdl.GLSetAttribute(sdl.GL_ALPHA_SIZE, 0)
		sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 24)
		// sdl.SetHint("SDL_VIDEO_EXTERNAL_CONTEXT", "0")
		// sdl.SetHint("SDL_HIDAPI_IGNORE_DEVICES", "1")
		// sdl.SetHint("SDL_JOYSTICK_ALLOW_BACKGROUND_EVENTS", "1")
		// sdl.SetHint(sdl.HINT_ORIENTATIONS, "LandscapeLeft LandscapeRight")
		// sdl.SetHint("SDL_ANDROID_TRAP_BACK_BUTTON", "1")
		// sdl.SetHint("SDL_JOYSTICK_HIDAPI", "0")
		// sdl.SetHint("SDL_ANDROID_SEPARATE_MOUSE_AND_TOUCH", "1")

		if sys.baseDir == "" {
			panic("FATAL: Android baseDir not set")
		}

		Logcat("sys.baseDir is: " + sys.baseDir)

		// Check if the directory even exists to Go
		if info, err := os.Stat(sys.baseDir); err != nil {
			Logcat(fmt.Sprintf("LOG: STAT ERROR: %v\n", err))
		} else {
			Logcat(fmt.Sprintf("LOG: STAT OK: %s is a dir: %v\n", sys.baseDir, info.IsDir()))
		}

		// FIX 1: Explicitly initialize os.Args before processCommandLine
		if os.Args == nil || len(os.Args) == 0 {
			os.Args = []string{"ikemen-go"}
		}

		if err := os.Chdir(sys.baseDir); err != nil {
			Logcat(fmt.Sprintf("LOG: CHDIR FAILED: %v\n", err))
			// Don't panic yet, let's see if we can continue
		} else {
			Logcat("LOG: CHDIR SUCCESSFUL")
		}

		// Init SDL NOW
		if err := sdl.Init(sdl.INIT_AUDIO | sdl.INIT_VIDEO | sdl.INIT_EVENTS | sdl.INIT_TIMER); err != nil {
			Logcat("LOG: SDL Init Failed: " + err.Error())
			return
		}
		Logcat("LOG: SDL Init SUCCESS")
	} else {
		sys.baseDir = "./"
	}

	// 1. Handle Permissions and Directory Creation
	permission := os.FileMode(0755)
	if runtime.GOOS != "android" {
		permission |= os.ModeSticky
	}

	// Create directories for ALL platforms
	os.MkdirAll(filepath.Join(sys.baseDir, "save/replays"), permission)
	os.MkdirAll(filepath.Join(sys.baseDir, "save/logs"), permission)

	processCommandLine()

	// Ensure cmdFlags exists even when there are no CLI args,
	// since we assign defaults below.
	if sys.cmdFlags == nil {
		sys.cmdFlags = make(map[string]string)
	}

	// Stats file path
	if _, ok := sys.cmdFlags["-stats"]; !ok {
		sys.cmdFlags["-stats"] = filepath.Join(sys.baseDir, "save/stats.json")
	}

	// Try reading stats
	if _, err := os.ReadFile(sys.cmdFlags["-stats"]); err != nil {
		// If there was an error reading, write an empty json file
		f, err := os.Create(sys.cmdFlags["-stats"])
		chk(err)
		f.Write([]byte("{}"))
		chk(f.Close())
	}

	if runtime.GOOS == "android" {
		sdl.InitSubSystem(sdl.INIT_JOYSTICK)
		sdl.InitSubSystem(sdl.INIT_GAMECONTROLLER)
		Logcat("LOG: Subsystems initialized!")
	}

	// Init the SDL LUT's
	initLUTs()

	// Config file path
	configPath := "save/config.ini"
	if val, ok := sys.cmdFlags["-config"]; ok {
		configPath = val
	}

	// Logcat("LOG: Loading config from: " + configPath)
	cfg, err := loadConfig(configPath)
	if err != nil {
		Logcat("LOG: loadConfig failed: " + err.Error())
		// For Android, let's see exactly what failed
		panic(err)
	}
	// Force to OpenGL ES 3.2 for Android
	if runtime.GOOS == "android" {
		cfg.Video.RenderMode = "OpenGL ES 3.2"
	}
	sys.cfg = *cfg
	// Logcat("LOG: Config Loaded. System Script: " + sys.cfg.Config.System)

	if sys.cfg.Debug.DumpLuaTables {
		os.MkdirAll(filepath.Join(sys.baseDir, "debug"), permission)
	}

	// Check Lua file path
	ftemp, err := os.Open(sys.cfg.Config.System)
	if err != nil {
		Logcat("LOG: LUA OPEN FAILED: " + err.Error())
		panic(err)
	}
	ftemp.Close()

	// Initialize game and create window
	// This is where the window is born!
	sys.luaLState = sys.init(sys.gameWidth, sys.gameHeight)
	defer sys.shutdown()

	// Begin processing game using its lua scripts
	if err := sys.luaLState.DoFile(sys.cfg.Config.System); err != nil {
		// Display error logs.
		errorLog := createLog("save/logs/Ikemen.log")
		defer closeLog(errorLog)

		// Write version and build time at the top
		fmt.Fprintf(errorLog, "Version: %s\nBuild Time: %s\n\nError log:\n", Version, BuildTime)

		// Write the rest of the log
		fmt.Fprintln(errorLog, err)

		switch err.(type) {
		case *lua.ApiError:
			errstr := strings.Split(err.Error(), "\n")[0]
			if len(errstr) < 10 || errstr[len(errstr)-10:] != "<game end>" {
				ShowErrorDialog(fmt.Sprintf("%s\n\nError saved to Ikemen.log", err))
				panic(err)
			}
		default:
			ShowErrorDialog(fmt.Sprintf("%s\n\nError saved to Ikemen.log", err))
			panic(err)
		}
	}
}

// Loops through given comand line arguments and processes them for later use by the game
func processCommandLine() {
	// If there are command line arguments
	if len(os.Args[1:]) > 0 {
		sys.cmdFlags = make(map[string]string)
		boolFlags := map[string]bool{
			"-windowed":                true,
			"-togglelifebars":          true,
			"-maxpowermode":            true,
			"-debug":                   true,
			"-nojoy":                   true,
			"-nomusic":                 true,
			"-nosound":                 true,
			"-speedtest":               true,
			"--enable-extended-inputs": true,
			"--input-ipc":              true,
		}
		key := ""
		player := 1
		flagsEncountered := false // New variable to track if any flags have been encountered
		r1, _ := regexp.Compile("^-[h%?]$")
		r2, _ := regexp.Compile("^-")
		// Loop through arguments
		for _, a := range os.Args[1:] {
			// Check if the current argument 'a' is a flag (starts with '-')
			// Check if 'a' is a number (could be negative)
			_, err := strconv.ParseFloat(a, 64)
			isNumber := err == nil

			// If there was a flag 'key' expecting a value, and 'a' is a number or not a flag
			if key != "" && (isNumber || !r2.MatchString(a)) { // If 'a' is a number OR 'a' is not a flag, it's a value for 'key'
				sys.cmdFlags[key] = a // Assign 'a' as the value for 'key'
				key = ""              // Value consumed, clear key.
			} else if r2.MatchString(a) { // 'a' is a flag (starts with '-')
				flagsEncountered = true // A flag has been encountered

				// If getting help about command line options
				if r1.MatchString(a) {
					text := `Options (case sensitive):
-h -?                   Help
-log <logfile>          Records match data to <logfile>
-r <path>               Loads motif <path>. eg. -r motifdir or -r motifdir/system.def
-lifebar <path>         Loads lifebar <path>. eg. -lifebar data/fight.def
-storyboard <path>      Loads storyboard <path>. eg. -storyboard chars/kfm/intro.def
-windowed               Starts in windowed mode (disables fullscreen)
-width <num>            Sets game width
-height <num>           Sets game height
-setvolume <num>        Sets master volume to <num> (0-100)
	
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
-speed <speed>          Changes game speed setting to <speed> (-9 to 9)
-stresstest <frameskip> Stability test (AI matches at speed increased by <frameskip>)
-speedtest              Speed test (match speed x100)
--enable-extended-inputs Allow up to 12 human input slots (P1-P12); P3-P12 can be toggled at runtime
--input-ipc              Read inputs from Unix socket (requires --enable-extended-inputs)
--input-socket-path <path> Socket path for IPC (default: /tmp/ikemen-input.sock)`
					//ShowInfoDialog(text, "I.K.E.M.E.N Command line options")
					fmt.Printf("I.K.E.M.E.N Command line options\n\n" + text + "\nPress ENTER to exit")
					var s string
					fmt.Scanln(&s)
					os.Exit(0)
				}
				// If 'a' is a boolean flag, set its value to "true".
				if _, isBool := boolFlags[a]; isBool {
					sys.cmdFlags[a] = "true"
					// key remains "" because boolean flags don't consume the next argument.
				} else {
					// 'a' is a value-expecting flag. Set its value to blank and store its name in 'key'.
					sys.cmdFlags[a] = ""
					key = a // Now 'key' expects a value in the next iteration.
				}
			} else if !flagsEncountered && player <= 2 { // Only assign to -p1 or -p2 if no flag has been encountered yet and player count is within limit
				// This block handles initial positional player names like "kfm kfm"
				sys.cmdFlags[fmt.Sprintf("-p%v", player)] = a
				player += 1
			}
			// If key is empty and player > 2, and it's not a flag, then it's an unhandled positional argument.
			// We just ignore it in this case to prevent the 8/8.def error.
		}
		// After the loop, if a key is still waiting for a value, set it to "true".
		if key != "" {
			sys.cmdFlags[key] = "true"
		}
		if sys.cmdFlags["--enable-extended-inputs"] == "true" {
			sys.extendedInputsEnabled = true
		}
		if sys.cmdFlags["--input-ipc"] == "true" {
			sys.externalInputMode = true
		}
		if path, ok := sys.cmdFlags["--input-socket-path"]; ok && path != "" && path != "true" {
			sys.inputSocketPath = path
		} else if sys.externalInputMode {
			sys.inputSocketPath = "/tmp/ikemen-input.sock"
		}
	}
}
