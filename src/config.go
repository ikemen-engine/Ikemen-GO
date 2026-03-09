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

type AIrampProperties struct {
	Start [2]int32 `ini:"start" sync:"runtime"`
	End   [2]int32 `ini:"end" sync:"runtime"`
}

type KeysProperties struct {
	Joystick int    `ini:"Joystick"`
	Up       string `ini:"up"`
	Down     string `ini:"down"`
	Left     string `ini:"left"`
	Right    string `ini:"right"`
	A        string `ini:"a"`
	B        string `ini:"b"`
	C        string `ini:"c"`
	X        string `ini:"x"`
	Y        string `ini:"y"`
	Z        string `ini:"z"`
	Start    string `ini:"start"`
	D        string `ini:"d"`
	W        string `ini:"w"`
	Menu     string `ini:"menu"`
	GUID     string `ini:"GUID"`
	RumbleOn bool   `ini:"RumbleOn"`
}

// Motif represents the top-level config structure.
type Config struct {
	Def            string
	IniFile        *ini.File
	DefaultOnlyIni *ini.File
	Common         struct {
		Air     map[string][]string `ini:"map:^(?i)Air[0-9]*$" lua:"Air" sync:"runtime"`
		Cmd     map[string][]string `ini:"map:^(?i)Cmd[0-9]*$" lua:"Cmd" sync:"runtime"`
		Const   map[string][]string `ini:"map:^(?i)Const[0-9]*$" lua:"Const" sync:"runtime"`
		States  map[string][]string `ini:"map:^(?i)States[0-9]*$" lua:"States" sync:"runtime"`
		Fx      map[string][]string `ini:"map:^(?i)Fx[0-9]*$" lua:"Fx" sync:"runtime"`
		Modules map[string][]string `ini:"map:^(?i)Modules[0-9]*$" lua:"Modules" sync:"runtime"`
		Lua     map[string][]string `ini:"map:^(?i)Lua[0-9]*$" lua:"Lua" sync:"runtime"`
	} `ini:"Common"`
	Options struct {
		Difficulty    int     `ini:"Difficulty" sync:"runtime"`
		Life          float32 `ini:"Life" sync:"runtime"`
		Time          int32   `ini:"Time" sync:"runtime"`
		GameSpeed     int     `ini:"GameSpeed" sync:"runtime"`
		GameSpeedStep int     `ini:"GameSpeedStep" sync:"runtime"`
		Match         struct {
			Wins         int32 `ini:"Wins" sync:"runtime"`
			MaxDrawGames int32 `ini:"MaxDrawGames" sync:"runtime"`
		} `ini:"Match"`
		Credits       int  `ini:"Credits" sync:"runtime"`
		QuickContinue bool `ini:"QuickContinue" sync:"runtime"`
		AutoGuard     bool `ini:"AutoGuard" sync:"runtime"`
		GuardBreak    bool `ini:"GuardBreak" sync:"runtime"`
		Dizzy         bool `ini:"Dizzy" sync:"runtime"`
		RedLife       bool `ini:"RedLife" sync:"runtime"`
		Team          struct {
			Duplicates       bool    `ini:"Duplicates" sync:"runtime"`
			LifeShare        bool    `ini:"LifeShare" sync:"runtime"`
			PowerShare       bool    `ini:"PowerShare" sync:"runtime"`
			SingleVsTeamLife float32 `ini:"SingleVsTeamLife" sync:"runtime"`
		} `ini:"Team"`
		Simul struct {
			Min   int `ini:"Min" sync:"runtime"`
			Max   int `ini:"Max" sync:"runtime"`
			Match struct {
				Wins int32 `ini:"Wins" sync:"runtime"`
			} `ini:"Match"`
			LoseOnKO bool `ini:"LoseOnKO" sync:"runtime"`
		} `ini:"Simul"`
		Tag struct {
			Min   int `ini:"Min" sync:"runtime"`
			Max   int `ini:"Max" sync:"runtime"`
			Match struct {
				Wins int32 `ini:"Wins" sync:"runtime"`
			} `ini:"Match"`
			LoseOnKO    bool    `ini:"LoseOnKO" sync:"runtime"`
			TimeScaling float32 `ini:"TimeScaling" sync:"runtime"`
		} `ini:"Tag"`
		Turns struct {
			Min      int `ini:"Min" sync:"runtime"`
			Max      int `ini:"Max" sync:"runtime"`
			Recovery struct {
				Base  float32 `ini:"Base" sync:"runtime"`
				Bonus float32 `ini:"Bonus" sync:"runtime"`
			} `ini:"Recovery"`
		} `ini:"Turns"`
		Ratio struct {
			Recovery struct {
				Base  float32 `ini:"Base" sync:"runtime"`
				Bonus float32 `ini:"Bonus" sync:"runtime"`
			} `ini:"Recovery"`
			Level1 struct {
				Attack float32 `ini:"Attack" sync:"runtime"`
				Life   float32 `ini:"Life" sync:"runtime"`
			} `ini:"Level1"`
			Level2 struct {
				Attack float32 `ini:"Attack" sync:"runtime"`
				Life   float32 `ini:"Life" sync:"runtime"`
			} `ini:"Level2"`
			Level3 struct {
				Attack float32 `ini:"Attack" sync:"runtime"`
				Life   float32 `ini:"Life" sync:"runtime"`
			} `ini:"Level3"`
			Level4 struct {
				Attack float32 `ini:"Attack" sync:"runtime"`
				Life   float32 `ini:"Life" sync:"runtime"`
			} `ini:"Level4"`
		} `ini:"Ratio"`
	} `ini:"Options"`
	Config struct {
		Motif             string   `ini:"Motif" sync:"startup"`
		Players           int      `ini:"Players" sync:"startup"`
		Language          string   `ini:"Language"`
		AfterImageMax     int32    `ini:"AfterImageMax" sync:"runtime"`
		ExplodMax         int      `ini:"ExplodMax" sync:"runtime"`
		HelperMax         int32    `ini:"HelperMax" sync:"runtime"`
		ProjectileMax     int      `ini:"ProjectileMax" sync:"runtime"`
		PaletteMax        int      `ini:"PaletteMax" sync:"runtime"`
		TextMax           int      `ini:"TextMax" sync:"runtime"`
		TickInterpolation bool     `ini:"TickInterpolation"`
		ZoomActive        bool     `ini:"ZoomActive" sync:"runtime"`
		EscOpensMenu      bool     `ini:"EscOpensMenu" sync:"runtime"`
		FirstRun          bool     `ini:"FirstRun"`
		WindowTitle       string   `ini:"WindowTitle"`
		WindowIcon        []string `ini:"WindowIcon"`
		System            string   `ini:"System" sync:"startup"`
		ScreenshotFolder  string   `ini:"ScreenshotFolder"`
		TrainingChar      string   `ini:"TrainingChar"`
		TrainingStage     string   `ini:"TrainingStage"`
		GamepadMappings   string   `ini:"GamepadMappings"`
	} `ini:"Config"`
	Debug struct {
		AllowDebugMode      bool    `ini:"AllowDebugMode"`
		AllowDebugKeys      bool    `ini:"AllowDebugKeys"`
		ClipboardRows       int     `ini:"ClipboardRows"`
		ConsoleRows         int     `ini:"ConsoleRows"`
		ClsnDarken          bool    `ini:"ClsnDarken"`
		DumpLuaTables       bool    `ini:"DumpLuaTables"`
		Font                string  `ini:"Font"`
		FontScale           float32 `ini:"FontScale"`
		StartStage          string  `ini:"StartStage"`
		ForceStageZoomout   float32 `ini:"ForceStageZoomout" sync:"runtime"`
		ForceStageZoomin    float32 `ini:"ForceStageZoomin" sync:"runtime"`
		ForceStageAutoZoom  bool    `ini:"ForceStageAutoZoom" sync:"runtime"`
		KeepSpritesOnReload bool    `ini:"KeepSpritesOnReload"`
		MacOSUseCommandKey  bool    `ini:"MacOSUseCommandKey"`
		SpeedTest           int     `ini:"SpeedTest"`
	} `ini:"Debug"`
	Video struct {
		RenderMode              string   `ini:"RenderMode"`
		GameWidth               int32    `ini:"GameWidth" sync:"startup"`
		GameHeight              int32    `ini:"GameHeight" sync:"startup"`
		WindowWidth             int      `ini:"WindowWidth"`
		WindowHeight            int      `ini:"WindowHeight"`
		Framerate               int      `ini:"Framerate"`
		VSync                   int      `ini:"VSync"`
		Fullscreen              bool     `ini:"Fullscreen"`
		Borderless              bool     `ini:"Borderless"`
		RGBSpriteBilinearFilter bool     `ini:"RGBSpriteBilinearFilter"`
		MSAA                    int32    `ini:"MSAA"`
		WindowCentered          bool     `ini:"WindowCentered"`
		ExternalShaders         []string `ini:"ExternalShaders"`
		WindowScaleMode         bool     `ini:"WindowScaleMode"`
		FightAspectWidth        int32    `ini:"FightAspectWidth" sync:"startup"`
		FightAspectHeight       int32    `ini:"FightAspectHeight" sync:"startup"`
		KeepAspect              bool     `ini:"KeepAspect"`
		RendererDebugMode       bool     `ini:"RendererDebugMode"`
		EnableModel             bool     `ini:"EnableModel"`
		EnableModelShadow       bool     `ini:"EnableModelShadow"`
	} `ini:"Video"`
	Sound struct {
		SampleRate        int32   `ini:"SampleRate"`
		SoundFont         string  `ini:"SoundFont"`
		StereoEffects     bool    `ini:"StereoEffects"`
		PanningRange      float32 `ini:"PanningRange"`
		WavChannels       int32   `ini:"WavChannels"`
		MasterVolume      int     `ini:"MasterVolume"`
		PauseMasterVolume int     `ini:"PauseMasterVolume"`
		WavVolume         int     `ini:"WavVolume"`
		BGMVolume         int     `ini:"BGMVolume"`
		BGMRAMBuffer      bool    `ini:"BGMRAMBuffer"`
		MaxBGMVolume      int     `ini:"MaxBGMVolume"`
		AudioDucking      bool    `ini:"AudioDucking"`
	} `ini:"Sound"`
	Arcade struct {
		AI struct {
			RandomColor   bool `ini:"RandomColor" sync:"runtime"`
			SurvivalColor bool `ini:"SurvivalColor" sync:"runtime"`
			Ramping       bool `ini:"Ramping" sync:"runtime"`
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
		ListenPort      string             `ini:"ListenPort"`
		RollbackNetcode bool               `ini:"RollbackNetcode" sync:"startup"`
		IP              map[string]string  `ini:"IP"`
		Rollback        RollbackProperties `ini:"Rollback"`
		JsonSync        []string           `ini:"JsonSync"`
	} `ini:"Netplay"`
	Input struct {
		ButtonAssist               bool    `ini:"ButtonAssist" sync:"runtime"`
		SOCDResolution             int     `ini:"SOCDResolution" sync:"runtime"`
		ControllerStickSensitivity float32 `ini:"ControllerStickSensitivity"`
		XinputTriggerSensitivity   float32 `ini:"XinputTriggerSensitivity"`
		UiRepeatDelay              int32   `ini:"UiRepeatDelay" sync:"runtime"`
		UiRepeatRate               int32   `ini:"UiRepeatRate" sync:"runtime"`
	} `ini:"Input"`
	Keys     map[string]*KeysProperties `ini:"map:^(?i)Keys_P[0-9]+$" lua:"Keys"`
	Joystick map[string]*KeysProperties `ini:"map:^(?i)Joystick_P[0-9]+$" lua:"Joystick"`
}

// Loads and parses the INI file into a Config struct.
func loadConfig(def string) (*Config, error) {
	// Define load options if needed
	// https://github.com/go-ini/ini/blob/main/ini.go
	baseOptions := ini.LoadOptions{
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

	// Preserve duplicates in user config so we can apply "first instance wins".
	userOptions := baseOptions
	userOptions.AllowShadows = true

	// Choose default config source: prefer physical file, else embedded bytes.
	var defaultSrc interface{}
	if fp := FileExist("resources/defaultConfig.ini"); len(fp) != 0 {
		defaultSrc = fp
	} else {
		defaultSrc = defaultConfig
	}
	// Load the INI file
	var iniFile *ini.File
	var defaultOnlyIni *ini.File
	var userIniFile *ini.File

	var err error
	// Load defaults-only.
	defaultOnlyIni, err = ini.LoadSources(baseOptions, defaultSrc)
	if err != nil {
		return nil, fmt.Errorf("failed to read defaults-only data: %v", err)
	}
	// Start merged INI as defaults, then overlay user (first-wins for duplicates).
	iniFile, err = ini.LoadSources(baseOptions, defaultSrc)
	if err != nil {
		return nil, fmt.Errorf("failed to read base data: %v", err)
	}
	if fp := FileExist(def); len(fp) != 0 {
		userIniFile, err = ini.LoadSources(userOptions, def)
		if err != nil {
			return nil, fmt.Errorf("failed to read user data: %v", err)
		}
		overlayUserFirstWins(iniFile, userIniFile)
	}
	var c Config
	c.Def = def
	c.initStruct()

	assignFrom := func(src *ini.File) {
		if src == nil {
			return
		}
		for _, section := range src.Sections() {
			sectionName := section.Name()
			if sectionName == ini.DEFAULT_SECTION {
				continue
			}
			for _, key := range section.Keys() {
				keyName := key.Name()
				value, dup := iniFirstValue(key)
				if dup > 0 {
					fmt.Printf("Warning: Duplicate key [%s] %s (%d duplicate(s) ignored)\n", sectionName, keyName, dup)
				}
				fullKey := strings.ReplaceAll(sectionName, " ", "_") + "." + strings.ReplaceAll(keyName, " ", "_")
				keyParts := parseQueryPath(fullKey)
				if err := assignField(&c, keyParts, value, def); err != nil {
					fmt.Printf("Warning: Failed to assign key [%s]: %v\n", fullKey, err)
				}
			}
		}
	}

	// Apply precedence: struct zero/default tags < defaultConfig.ini < user config
	assignFrom(defaultOnlyIni)
	assignFrom(userIniFile)

	c.IniFile = iniFile
	c.DefaultOnlyIni = defaultOnlyIni
	c.normalize()
	c.sysSet()
	c.Save(def)
	return &c, nil
}

// Initialize struct
func (c *Config) initStruct() {
	initMaps(reflect.ValueOf(c).Elem())
	//applyDefaultsToValue(reflect.ValueOf(c).Elem())
}

// Normalize values
func (c *Config) normalize() {
	c.SetValueUpdate("Config.Players", int(Clamp(int32(c.Config.Players), 1, int32(MaxSimul)*2)))
	c.SetValueUpdate("Options.Difficulty", int(Clamp(int32(c.Options.Difficulty), 1, 8)))
	c.SetValueUpdate("Options.GameSpeed", int(Clamp(int32(c.Options.GameSpeed), -9, 9)))
	c.SetValueUpdate("Options.GameSpeedStep", int(Clamp(int32(c.Options.GameSpeedStep), 1, 60)))
	c.SetValueUpdate("Options.Simul.Max", int(Clamp(int32(c.Options.Simul.Max), int32(c.Options.Simul.Min), int32(MaxSimul))))
	c.SetValueUpdate("Options.Simul.Min", int(Clamp(int32(c.Options.Simul.Min), 2, int32(MaxSimul))))
	c.SetValueUpdate("Options.Tag.Max", int(Clamp(int32(c.Options.Tag.Max), int32(c.Options.Tag.Min), int32(MaxSimul))))
	c.SetValueUpdate("Options.Tag.Min", int(Clamp(int32(c.Options.Tag.Min), 2, int32(MaxSimul))))
	c.SetValueUpdate("Video.Framerate", int(Clamp(int32(c.Video.Framerate), 1, 840)))

	// Options that determine allocation sizes should not be negative
	// Update: AfterImageMax no longer does, but it's good to keep it in mind
	//if c.Config.AfterImageMax < 0 {
	///	c.SetValueUpdate("Config.AfterImageMax", 0)
	//}

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

	c.SetValueUpdate("Input.SOCDResolution", int(Clamp(int32(c.Input.SOCDResolution), 0, 4)))
	c.SetValueUpdate("Input.UiRepeatDelay", int(Max(int32(c.Input.UiRepeatDelay), 0)))
	c.SetValueUpdate("Input.UiRepeatRate", int(Max(int32(c.Input.UiRepeatRate), 1)))
	c.SetValueUpdate("Sound.MaxBGMVolume", int(Clamp(int32(c.Sound.MaxBGMVolume), 100, 250)))
	c.SetValueUpdate("Sound.PanningRange", ClampF(c.Sound.PanningRange, 0, 100))
	c.SetValueUpdate("Sound.PauseMasterVolume", int(Clamp(int32(c.Sound.PauseMasterVolume), 0, 100)))
	c.SetValueUpdate("Sound.WavChannels", Clamp(c.Sound.WavChannels, 1, 256))

	switch c.Video.MSAA {
	case 0, 2, 4, 6, 8, 16, 32:
	default:
		c.SetValueUpdate("Video.MSAA", 0)
	}
}

// Sets system values
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
	for i := 1; i <= c.Config.Players; i++ {
		if kc, ok := c.Keys[fmt.Sprintf("keys_p%d", i)]; ok {
			newKeyConfig := KeyConfig{
				Joy:      kc.Joystick,
				GUID:     kc.GUID,
				dU:       stoki(kc.Up),
				dD:       stoki(kc.Down),
				dL:       stoki(kc.Left),
				dR:       stoki(kc.Right),
				bA:       stoki(kc.A),
				bB:       stoki(kc.B),
				bC:       stoki(kc.C),
				bX:       stoki(kc.X),
				bY:       stoki(kc.Y),
				bZ:       stoki(kc.Z),
				bS:       stoki(kc.Start),
				bD:       stoki(kc.D),
				bW:       stoki(kc.W),
				bM:       stoki(kc.Menu),
				rumbleOn: kc.RumbleOn,
			}
			sys.keyConfig = append(sys.keyConfig, newKeyConfig)
		} else {
			sys.keyConfig = append(sys.keyConfig, KeyConfig{Joy: -1})
		}
		if _, ok := sys.cmdFlags["-nojoy"]; !ok {
			if kc, ok := c.Joystick[fmt.Sprintf("joystick_p%d", i)]; ok {
				newKeyConfig := KeyConfig{
					Joy:      kc.Joystick,
					GUID:     kc.GUID,
					dU:       StringToButtonLUT[kc.Up],
					dD:       StringToButtonLUT[kc.Down],
					dL:       StringToButtonLUT[kc.Left],
					dR:       StringToButtonLUT[kc.Right],
					bA:       StringToButtonLUT[kc.A],
					bB:       StringToButtonLUT[kc.B],
					bC:       StringToButtonLUT[kc.C],
					bX:       StringToButtonLUT[kc.X],
					bY:       StringToButtonLUT[kc.Y],
					bZ:       StringToButtonLUT[kc.Z],
					bS:       StringToButtonLUT[kc.Start],
					bD:       StringToButtonLUT[kc.D],
					bW:       StringToButtonLUT[kc.W],
					bM:       StringToButtonLUT[kc.Menu],
					rumbleOn: kc.RumbleOn,
				}
				sys.joystickConfig = append(sys.joystickConfig, newKeyConfig)
			} else {
				sys.joystickConfig = append(sys.joystickConfig, KeyConfig{Joy: i - 1})
			}
		}
	}
}

// GetValue retrieves the value based on the query string.
func (c *Config) GetValue(query string) (interface{}, error) {
	return GetValue(c, query)
}

// SetValueUpdate sets the value based on the query string and updates the IniFile.
func (c *Config) SetValueUpdate(query string, value interface{}) error {
	return SetValueUpdate(c, c.IniFile, query, value)
}

// Save writes the current IniFile to disk, preserving comments and syntax.
func (c *Config) Save(file string) error {
	return SaveINI(c.IniFile, file)
}
