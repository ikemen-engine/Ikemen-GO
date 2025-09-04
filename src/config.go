// config.go handles runtime configuration for Ikemen GO.
// Configuration is loaded from embedded defaults, user INI files, and command-line flags.
// Parsing and reflection helpers live in iniutils.go.
package main

import (
	_ "embed" // Support for go:embed resources
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

//go:embed resources/defaultConfig.ini
var defaultConfig []byte

// AIrampProperties describes the AI difficulty ramp.
// Start and End store {match, difficulty} pairs.
type AIrampProperties struct {
	Start [2]int32 `ini:"start"` // {match, diff}
	End   [2]int32 `ini:"end"`   // {match, diff}
}

// KeysProperties lists per-player input bindings.
// Fields hold raw key names; command-line flags such as -nojoy can disable joystick reading.
type KeysProperties struct {
	Joystick int    `ini:"Joystick"`
	Up       string `ini:"Up"`
	Down     string `ini:"Down"`
	Left     string `ini:"Left"`
	Right    string `ini:"Right"`
	A        string `ini:"A"`
	B        string `ini:"B"`
	C        string `ini:"C"`
	X        string `ini:"X"`
	Y        string `ini:"Y"`
	Z        string `ini:"Z"`
	Start    string `ini:"Start"`
	D        string `ini:"D"`
	W        string `ini:"W"`
	Menu     string `ini:"Menu"`
	GUID     string `ini:"GUID"`
}

// Motif represents the top-level config structure.
type Config struct {
	Def     string
	IniFile *ini.File
	Common  struct {
		Air     map[string][]string `ini:"map:^(?i)Air[0-9]*$" lua:"Air"`
		Cmd     map[string][]string `ini:"map:^(?i)Cmd[0-9]*$" lua:"Cmd"`
		Const   map[string][]string `ini:"map:^(?i)Const[0-9]*$" lua:"Const"`
		States  map[string][]string `ini:"map:^(?i)States[0-9]*$" lua:"States"`
		Fx      map[string][]string `ini:"map:^(?i)Fx[0-9]*$" lua:"Fx"`
		Modules map[string][]string `ini:"map:^(?i)Modules[0-9]*$" lua:"Modules"`
		Lua     map[string][]string `ini:"map:^(?i)Lua[0-9]*$" lua:"Lua"`
	} `ini:"Common"`
	Options struct {
		Difficulty int     `ini:"Difficulty"` // 1-8 scale; -ailevel flag overrides
		Life       float32 `ini:"Life"`       // percent of default life
		Time       int32   `ini:"Time"`       // seconds; -1 = infinite
		GameSpeed  float32 `ini:"GameSpeed"`  // speed steps [-9..9]; -speed flag overrides
		Match      struct {
			Wins         int32 `ini:"Wins"`
			MaxDrawGames int32 `ini:"MaxDrawGames"`
		} `ini:"Match"`
		Credits       int  `ini:"Credits"`
		QuickContinue bool `ini:"QuickContinue"`
		AutoGuard     bool `ini:"AutoGuard"`
		GuardBreak    bool `ini:"GuardBreak"`
		Dizzy         bool `ini:"Dizzy"`
		RedLife       bool `ini:"RedLife"`
		Team          struct {
			Duplicates       bool    `ini:"Duplicates"`
			LifeShare        bool    `ini:"LifeShare"`
			PowerShare       bool    `ini:"PowerShare"`
			SingleVsTeamLife float32 `ini:"SingleVsTeamLife"` // percent
		} `ini:"Team"`
		Simul struct {
			Min   int `ini:"Min"` // min team size
			Max   int `ini:"Max"` // max team size
			Match struct {
				Wins int32 `ini:"Wins"`
			} `ini:"Match"`
			LoseOnKO bool `ini:"LoseOnKO"`
		} `ini:"Simul"`
		Tag struct {
			Min   int `ini:"Min"`
			Max   int `ini:"Max"`
			Match struct {
				Wins int32 `ini:"Wins"`
			} `ini:"Match"`
			LoseOnKO    bool    `ini:"LoseOnKO"`
			TimeScaling float32 `ini:"TimeScaling"`
		} `ini:"Tag"`
		Turns struct {
			Min      int `ini:"Min"`
			Max      int `ini:"Max"`
			Recovery struct {
				Base  float32 `ini:"Base"`
				Bonus float32 `ini:"Bonus"`
			} `ini:"Recovery"`
		} `ini:"Turns"`
		Ratio struct {
			Recovery struct {
				Base  float32 `ini:"Base"`
				Bonus float32 `ini:"Bonus"`
			} `ini:"Recovery"`
			Level1 struct {
				Attack float32 `ini:"Attack"`
				Life   float32 `ini:"Life"`
			} `ini:"Level1"`
			Level2 struct {
				Attack float32 `ini:"Attack"`
				Life   float32 `ini:"Life"`
			} `ini:"Level2"`
			Level3 struct {
				Attack float32 `ini:"Attack"`
				Life   float32 `ini:"Life"`
			} `ini:"Level3"`
			Level4 struct {
				Attack float32 `ini:"Attack"`
				Life   float32 `ini:"Life"`
			} `ini:"Level4"`
		} `ini:"Ratio"`
	} `ini:"Options"`
	Config struct {
		Motif               string   `ini:"Motif"`
		Players             int      `ini:"Players"`
		Framerate           int      `ini:"Framerate"` // frames per second
		Language            string   `ini:"Language"`
		AfterImageMax       int32    `ini:"AfterImageMax"`
		ExplodMax           int      `ini:"ExplodMax"`
		HelperMax           int32    `ini:"HelperMax"`
		PlayerProjectileMax int      `ini:"PlayerProjectileMax"`
		ZoomActive          bool     `ini:"ZoomActive"`
		EscOpensMenu        bool     `ini:"EscOpensMenu"`
		FirstRun            bool     `ini:"FirstRun"`
		WindowTitle         string   `ini:"WindowTitle"`
		WindowIcon          []string `ini:"WindowIcon"`
		System              string   `ini:"System"`
		ScreenshotFolder    string   `ini:"ScreenshotFolder"`
		TrainingChar        string   `ini:"TrainingChar"`
		GamepadMappings     string   `ini:"GamepadMappings"`
	} `ini:"Config"`
	Debug struct {
		AllowDebugMode    bool    `ini:"AllowDebugMode"`
		AllowDebugKeys    bool    `ini:"AllowDebugKeys"`
		ClipboardRows     int     `ini:"ClipboardRows"`
		ConsoleRows       int     `ini:"ConsoleRows"`
		ClsnDarken        bool    `ini:"ClsnDarken"`
		Font              string  `ini:"Font"`
		FontScale         float32 `ini:"FontScale"`
		StartStage        string  `ini:"StartStage"`
		ForceStageZoomout float32 `ini:"ForceStageZoomout"`
		ForceStageZoomin  float32 `ini:"ForceStageZoomin"`
	} `ini:"Debug"`
	Video struct {
		RenderMode              string   `ini:"RenderMode"`
		GameWidth               int32    `ini:"GameWidth"`    // pixels; overridden by -width flag
		GameHeight              int32    `ini:"GameHeight"`   // pixels; overridden by -height flag
		WindowWidth             int      `ini:"WindowWidth"`  // pixels
		WindowHeight            int      `ini:"WindowHeight"` // pixels
		VSync                   int      `ini:"VSync"`
		Fullscreen              bool     `ini:"Fullscreen"`
		Borderless              bool     `ini:"Borderless"`
		RGBSpriteBilinearFilter bool     `ini:"RGBSpriteBilinearFilter"`
		MSAA                    int32    `ini:"MSAA"`
		WindowCentered          bool     `ini:"WindowCentered"`
		ExternalShaders         []string `ini:"ExternalShaders"`
		WindowScaleMode         bool     `ini:"WindowScaleMode"`
		KeepAspect              bool     `ini:"KeepAspect"`
		EnableModel             bool     `ini:"EnableModel"`
		EnableModelShadow       bool     `ini:"EnableModelShadow"`
	} `ini:"Video"`
	Sound struct {
		SampleRate        int32   `ini:"SampleRate"` // Hz
		StereoEffects     bool    `ini:"StereoEffects"`
		PanningRange      float32 `ini:"PanningRange"` // percent
		WavChannels       int32   `ini:"WavChannels"`
		MasterVolume      int     `ini:"MasterVolume"`      // percent; -setvolume flag overrides
		PauseMasterVolume int     `ini:"PauseMasterVolume"` // percent
		WavVolume         int     `ini:"WavVolume"`         // percent
		BGMVolume         int     `ini:"BGMVolume"`         // percent
		MaxBGMVolume      int     `ini:"MaxBGMVolume"`      // percent
		AudioDucking      bool    `ini:"AudioDucking"`
	} `ini:"Sound"`
	Arcade struct {
		AI struct {
			RandomColor   bool `ini:"RandomColor"`
			SurvivalColor bool `ini:"SurvivalColor"`
			Ramping       bool `ini:"Ramping"`
		} `ini:"AI"`
		//items map[string]AIrampProperties `ini:"items"`
		Arcade struct {
			AIramp AIrampProperties `ini:"AIramp"`
		} `ini:"arcade"`
		Team struct {
			AIramp AIrampProperties `ini:"AIramp"`
		} `ini:"team"`
		Ratio struct {
			AIramp AIrampProperties `ini:"AIramp"`
		} `ini:"ratio"`
		Survival struct {
			AIramp AIrampProperties `ini:"AIramp"`
		} `ini:"survival"`
	} `ini:"Arcade"`
	Netplay struct {
		ListenPort string            `ini:"ListenPort"`
		IP         map[string]string `ini:"IP"`
	} `ini:"Netplay"`
	Input struct {
		ButtonAssist               bool    `ini:"ButtonAssist"`
		SOCDResolution             int     `ini:"SOCDResolution"`
		ControllerStickSensitivity float32 `ini:"ControllerStickSensitivity"`
		XinputTriggerSensitivity   float32 `ini:"XinputTriggerSensitivity"`
	} `ini:"Input"`
	Keys     map[string]*KeysProperties `ini:"map:^(?i)Keys_P[0-9]+$" lua:"Keys"`
	Joystick map[string]*KeysProperties `ini:"map:^(?i)Joystick_P[0-9]+$" lua:"Joystick"`
}

// loadConfig reads configuration from the provided INI file path.
// Assumes `def` points to a valid config file; if missing, embedded defaults are used.
// Command-line flags may override values after loading. Environment variables are ignored.
// [MUGEN-Compat] Section and key names follow M.U.G.E.N's configuration format.
func loadConfig(def string) (*Config, error) {
	// Define load options if needed
	// https://github.com/go-ini/ini/blob/main/ini.go
	options := ini.LoadOptions{
		Insensitive: false,
		//InsensitiveSections: true,
		//InsensitiveKeys: true,
		IgnoreInlineComment:     false,
		SkipUnrecognizableLines: true,
		//AllowBooleanKeys: true,
		AllowShadows: false,
		//AllowNestedValues: true,
		UnparseableSections:        []string{},
		AllowPythonMultilineValues: false,
		//KeyValueDelimiters: "=:",
		//KeyValueDelimiterOnWrite: "=",
		//ChildSectionDelimiter: ".",
		//AllowNonUniqueSections: true,
		//AllowDuplicateShadowValues: true,
	}

	// Load the INI file
	var iniFile *ini.File
	var err error
	if fp := FileExist(def); len(fp) == 0 {
		iniFile, err = ini.LoadSources(options, defaultConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to read data: %v", err)
		}
	} else {
		iniFile, err = ini.LoadSources(options, defaultConfig, def)
		if err != nil {
			return nil, fmt.Errorf("failed to read data: %v", err)
		}
	}
	var c Config
	c.Def = def
	c.initStruct()

	// Iterate through all sections
	for _, section := range iniFile.Sections() {
		sectionName := section.Name()

		// Skip the default section
		if sectionName == ini.DEFAULT_SECTION {
			continue
		}

		// Always include the section name as the first part of the key
		for _, key := range section.Keys() {
			keyName := key.Name()
			values := key.ValueWithShadows() // Retrieve all shadowed values

			for _, value := range values {
				// Replace spaces with underscores in section and key names before parsing.
				fullKey := strings.ReplaceAll(sectionName, " ", "_") + "." + strings.ReplaceAll(keyName, " ", "_")

				keyParts := parseQueryPath(fullKey)
				if err := assignField(&c, keyParts, value); err != nil {
					fmt.Printf("Warning: Failed to assign key [%s]: %v\n", fullKey, err)
				}
			}
		}
	}

	c.IniFile = iniFile
	c.normalize()
	c.sysSet()
	c.Save(def)
	return &c, nil
}

// initStruct prepares nested maps before assignment.
// Must be called prior to parsing to avoid nil map writes.
// [MUGEN-Compat] mirrors M.U.G.E.N's default structure.
func (c *Config) initStruct() {
	initMaps(reflect.ValueOf(c).Elem())
	//applyDefaultsToValue(reflect.ValueOf(c).Elem())
}

// normalize clamps and sanitizes loaded values.
// Assumes fields are already populated from INI; no environment overrides.
// [MUGEN-Compat] keeps values within engine-supported ranges.
func (c *Config) normalize() {
	c.SetValueUpdate("Options.GameSpeed", ClampF(c.Options.GameSpeed, -9, 9))
	c.SetValueUpdate("Options.Simul.Min", int(Clamp(int32(c.Options.Simul.Min), 2, int32(MaxSimul))))
	c.SetValueUpdate("Options.Simul.Max", int(Clamp(int32(c.Options.Simul.Max), int32(c.Options.Simul.Min), int32(MaxSimul))))
	c.SetValueUpdate("Options.Tag.Min", int(Clamp(int32(c.Options.Tag.Min), 2, int32(MaxSimul))))
	c.SetValueUpdate("Options.Tag.Max", int(Clamp(int32(c.Options.Tag.Max), int32(c.Options.Tag.Min), int32(MaxSimul))))
	c.SetValueUpdate("Config.Players", int(Clamp(int32(c.Config.Players), 1, int32(MaxSimul)*2)))
	c.SetValueUpdate("Config.Framerate", int(Clamp(int32(c.Config.Framerate), 1, 840)))
	path := strings.TrimSpace(c.Config.ScreenshotFolder)
	if path != "" {
		path = strings.ReplaceAll(path, "\\", "/")
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		if path != c.Config.ScreenshotFolder {
			c.SetValueUpdate("Config.ScreenshotFolder", path)
		}
	}
	switch c.Sound.SampleRate {
	case 22050, 44100, 48000:
	default:
		c.SetValueUpdate("Sound.SampleRate", 44100)
	}
	c.SetValueUpdate("Sound.PanningRange", ClampF(c.Sound.PanningRange, 0, 100))
	c.SetValueUpdate("Sound.WavChannels", Clamp(c.Sound.WavChannels, 1, 256))
	c.SetValueUpdate("Sound.PauseMasterVolume", int(Clamp(int32(c.Sound.PauseMasterVolume), 0, 100)))
	c.SetValueUpdate("Sound.MaxBGMVolume", int(Clamp(int32(c.Sound.MaxBGMVolume), 100, 250)))
	c.SetValueUpdate("Input.SOCDResolution", int(Clamp(int32(c.Input.SOCDResolution), 0, 4)))
	switch c.Video.MSAA {
	case 0, 2, 4, 6, 8, 16, 32:
	default:
		c.SetValueUpdate("Video.MSAA", 0)
	}
}

// sysSet applies configuration to the runtime system.
// Command-line flags like -width/-height take precedence over INI values.
// Environment variables are not checked.
// [MUGEN-Compat] keeps behavior aligned with original key mapping rules.
func (c *Config) sysSet() {
	if _, ok := sys.cmdFlags["-width"]; ok {
		var w, _ = strconv.ParseInt(sys.cmdFlags["-width"], 10, 32)
		sys.gameWidth = int32(w)
	} else {
		sys.gameWidth = c.Video.GameWidth
	}
	if _, ok := sys.cmdFlags["-height"]; ok {
		var h, _ = strconv.ParseInt(sys.cmdFlags["-height"], 10, 32)
		sys.gameHeight = int32(h)
	} else {
		sys.gameHeight = c.Video.GameHeight
	}
	sys.msaa = c.Video.MSAA
	stoki := func(key string) int {
		return int(StringToKey(key))
	}
	Atoi := func(key string) int {
		if i, err := strconv.Atoi(key); err == nil {
			return i
		}
		return 999
	}
	for i := 1; i <= c.Config.Players; i++ {
		if kc, ok := c.Keys[fmt.Sprintf("keys_p%d", i)]; ok {
			newKeyConfig := KeyConfig{
				Joy:  kc.Joystick,
				GUID: kc.GUID,
				dU:   stoki(kc.Up),
				dD:   stoki(kc.Down),
				dL:   stoki(kc.Left),
				dR:   stoki(kc.Right),
				kA:   stoki(kc.A),
				kB:   stoki(kc.B),
				kC:   stoki(kc.C),
				kX:   stoki(kc.X),
				kY:   stoki(kc.Y),
				kZ:   stoki(kc.Z),
				kS:   stoki(kc.Start),
				kD:   stoki(kc.D),
				kW:   stoki(kc.W),
				kM:   stoki(kc.Menu),
			}
			sys.keyConfig = append(sys.keyConfig, newKeyConfig)
		} else {
			sys.keyConfig = append(sys.keyConfig, KeyConfig{Joy: -1})
		}
		if _, ok := sys.cmdFlags["-nojoy"]; !ok {
			if kc, ok := c.Joystick[fmt.Sprintf("joystick_p%d", i)]; ok {
				newKeyConfig := KeyConfig{
					Joy:  kc.Joystick,
					GUID: kc.GUID,
					dU:   Atoi(kc.Up),
					dD:   Atoi(kc.Down),
					dL:   Atoi(kc.Left),
					dR:   Atoi(kc.Right),
					kA:   Atoi(kc.A),
					kB:   Atoi(kc.B),
					kC:   Atoi(kc.C),
					kX:   Atoi(kc.X),
					kY:   Atoi(kc.Y),
					kZ:   Atoi(kc.Z),
					kS:   Atoi(kc.Start),
					kD:   Atoi(kc.D),
					kW:   Atoi(kc.W),
					kM:   Atoi(kc.Menu),
				}
				sys.joystickConfig = append(sys.joystickConfig, newKeyConfig)
			} else {
				sys.joystickConfig = append(sys.joystickConfig, KeyConfig{Joy: i - 1})
			}
		}
	}
}

// GetValue returns the value referenced by a dotted query path.
// Query syntax matches assignField/SetValueUpdate. No env var expansion.
// [MUGEN-Compat] case-insensitive tags mimic M.U.G.E.N lookups.
func (c *Config) GetValue(query string) (interface{}, error) {
	return GetValue(c, query)
}

// SetValueUpdate updates a field using a dotted query path and writes back to IniFile.
// Accepts basic types; invalid paths return an error. Environment variables are ignored.
// [MUGEN-Compat] allows runtime overrides similar to original engine's config editing.
func (c *Config) SetValueUpdate(query string, value interface{}) error {
	return SetValueUpdate(c, c.IniFile, query, value)
}

// Save persists the current configuration to disk.
// Assumes `file` is writeable; environment is not consulted for path expansion.
// [MUGEN-Compat] preserves INI formatting to match upstream expectations.
func (c *Config) Save(file string) error {
	return SaveINI(c.IniFile, file)
}
