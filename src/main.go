package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

var Version = "development"
var BuildTime = "" // Set automatically by GitHub Actions

func init() {
	runtime.LockOSThread()
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

	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
	} else {
		// Change the context for Darwin if we're in an app bundle
		if isRunningInsideAppBundle(exePath) {
			os.Chdir(path.Dir(exePath))
			os.Chdir("../../../")
		}
	}

	// Make save directories, if they don't exist
	os.Mkdir("save", os.ModeSticky|0755)
	os.Mkdir("save/replays", os.ModeSticky|0755)
	os.Mkdir("save/logs", os.ModeSticky|0755)

	processCommandLine()

	// Try reading stats
	if _, err := os.ReadFile("save/stats.json"); err != nil {
		// If there was an error reading, write an empty json file
		f, err := os.Create("save/stats.json")
		chk(err)
		f.Write([]byte("{}"))
		chk(f.Close())
	}

	// Config file path
	cfgPath := "save/config.ini"
	// If a different config file is defined in the command line parameters, use it instead
	if _, ok := sys.cmdFlags["-config"]; ok {
		cfgPath = sys.cmdFlags["-config"]
	}

	if cfg, err := loadConfig(cfgPath); err != nil {
		chk(err)
	} else {
		sys.cfg = *cfg
	}

	//os.Mkdir("debug", os.ModeSticky|0755)

	// Check if the main lua file exists.
	if ftemp, err1 := os.Open(sys.cfg.Config.System); err1 != nil {
		ftemp.Close()
		var err2 = Error(
			"Main lua file \"" + sys.cfg.Config.System + "\" error." +
				"\n" + err1.Error(),
		)
		ShowErrorDialog(err2.Error())
		panic(err2)
	} else {
		ftemp.Close()
	}

	// Initialize game and create window
	sys.luaLState = sys.init(sys.gameWidth, sys.gameHeight)
	defer sys.shutdown()

	// Begin processing game using its lua scripts
	if err := sys.luaLState.DoFile(sys.cfg.Config.System); err != nil {
		// Display error logs.
		errorLog := createLog("Ikemen.log")
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
			"-windowed":       true,
			"-togglelifebars": true,
			"-maxpowermode":   true,
			"-debug":          true,
			"-nojoy":          true,
			"-nomusic":        true,
			"-nosound":        true,
			"-speedtest":      true,
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
-speedtest              Speed test (match speed x100)`
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
	}
}
