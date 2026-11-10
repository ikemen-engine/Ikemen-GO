package main

import (
	_ "embed" // Support for go:embed resources
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

//go:embed resources/defaultConfig.ini
var defaultConfig []byte

type AIrampProperties struct {
	Start [2]int32 `ini:"start" sync:"host"`
	End   [2]int32 `ini:"end" sync:"host"`
}

// A Controller Setup binds virtual controls to engine buttons or combinations.
type ControllerSetupProperties struct {
	Name    string            `ini:"Name"`
	Buttons map[string]string `ini:"map:^(?i)(up|down|left|right|a|b|c|x|y|z|start|d|w|menu|l|r)+$" lua:""`
}

// Individual bindings come first, so their glyphs take priority over combinations.
func (p *ControllerSetupProperties) buttonNames() []string {
	var combinations []string
	for name := range p.Buttons {
		if mask := setupButtonMask(name); mask != 0 && mask&(mask-1) != 0 {
			combinations = append(combinations, name)
		}
	}
	sort.Strings(combinations)
	return append(append([]string(nil), engineButtonNames[:]...), combinations...)
}

func (p *ControllerSetupProperties) setup() ControllerSetup {
	var setup ControllerSetup
	for _, name := range p.buttonNames() {
		vc := parseVirtualControl(p.Buttons[name])
		if vc == VC_None {
			continue
		}
		mask := setupButtonMask(name)
		for i := range setup {
			if mask&(1<<uint(i)) != 0 {
				setup[i] = append(setup[i], vc)
			}
		}
	}
	return setup
}

func (p *ControllerSetupProperties) read(section *ini.Section) {
	p.Name = section.Key("Name").String()
	p.Buttons = make(map[string]string)
	for _, key := range section.Keys() {
		if setupButtonMask(key.Name()) != 0 {
			p.Buttons[strings.ToLower(key.Name())] = key.String()
		}
	}
}

// A Button Mapping in config.ini: the physical input behind each virtual control.
// [ButtonMapping_KeyboardN] sections are the keyboard controllers.
type ButtonMappingProperties struct {
	Name  string `ini:"Name"`
	DpU   string `ini:"DP_U"`
	DpD   string `ini:"DP_D"`
	DpL   string `ini:"DP_L"`
	DpR   string `ini:"DP_R"`
	A     string `ini:"A"`
	B     string `ini:"B"`
	X     string `ini:"X"`
	Y     string `ini:"Y"`
	Start string `ini:"START"`
	Back  string `ini:"BACK"`
	LB    string `ini:"LB"`
	RB    string `ini:"RB"`
	LT    string `ini:"LT"`
	RT    string `ini:"RT"`
	LsYn  string `ini:"LS_Y-"`
	LsYp  string `ini:"LS_Y+"`
	LsXn  string `ini:"LS_X-"`
	LsXp  string `ini:"LS_X+"`
	RsYn  string `ini:"RS_Y-"`
	RsYp  string `ini:"RS_Y+"`
	RsXn  string `ini:"RS_X-"`
	RsXp  string `ini:"RS_X+"`
	LS    string `ini:"LS"`
	RS    string `ini:"RS"`
}

func (p *ButtonMappingProperties) mapping() ButtonMapping {
	return ButtonMapping{p.DpU, p.DpD, p.DpL, p.DpR, p.A, p.B, p.X, p.Y, p.Start, p.Back,
		p.LB, p.RB, p.LT, p.RT, p.LsYn, p.LsYp, p.LsXn, p.LsXp, p.RsYn, p.RsYp, p.RsXn, p.RsXp, p.LS, p.RS}
}

func (p *ButtonMappingProperties) setMapping(m ButtonMapping) {
	*p = ButtonMappingProperties{p.Name, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9],
		m[10], m[11], m[12], m[13], m[14], m[15], m[16], m[17], m[18], m[19], m[20], m[21], m[22], m[23]}
}

// Motif represents the top-level config structure.
type Config struct {
	Def            string
	IniFile        *ini.File
	DefaultOnlyIni *ini.File
	Common         struct {
		Air     map[string][]string `ini:"map:^(?i)Air[0-9]*$" lua:"Air" sync:"host"`
		Cmd     map[string][]string `ini:"map:^(?i)Cmd[0-9]*$" lua:"Cmd" sync:"host"`
		Const   map[string][]string `ini:"map:^(?i)Const[0-9]*$" lua:"Const" sync:"host"`
		States  map[string][]string `ini:"map:^(?i)States[0-9]*$" lua:"States" sync:"host"`
		Fx      map[string][]string `ini:"map:^(?i)Fx[0-9]*$" lua:"Fx" sync:"host"`
		Shaders map[string][]string `ini:"map:^(?i)Shaders[0-9]*$" lua:"Shaders" sync:"host"`
		Modules map[string][]string `ini:"map:^(?i)Modules[0-9]*$" lua:"Modules" sync:"host"`
		Lua     map[string][]string `ini:"map:^(?i)Lua[0-9]*$" lua:"Lua" sync:"host"`
	} `ini:"Common"`
	Options struct {
		Difficulty    int     `ini:"Difficulty" sync:"host"`
		Life          float32 `ini:"Life" sync:"host"`
		Time          int32   `ini:"Time" sync:"host"`
		GameSpeed     int     `ini:"GameSpeed" sync:"host"`
		GameSpeedStep int     `ini:"GameSpeedStep" sync:"host"`
		Match         struct {
			Wins         int32 `ini:"Wins" sync:"host"`
			MaxDrawGames int32 `ini:"MaxDrawGames" sync:"host"`
		} `ini:"Match"`
		Credits       int  `ini:"Credits" sync:"host"`
		QuickContinue bool `ini:"QuickContinue" sync:"host"`
		AutoGuard     bool `ini:"AutoGuard" sync:"host"`
		GuardBreak    bool `ini:"GuardBreak" sync:"host"`
		Dizzy         bool `ini:"Dizzy" sync:"host"`
		RedLife       bool `ini:"RedLife" sync:"host"`
		Team          struct {
			Duplicates       bool    `ini:"Duplicates" sync:"host"`
			LifeShare        bool    `ini:"LifeShare" sync:"host"`
			PowerShare       bool    `ini:"PowerShare" sync:"host"`
			SingleVsTeamLife float32 `ini:"SingleVsTeamLife" sync:"host"`
		} `ini:"Team"`
		Simul struct {
			Min   int `ini:"Min" sync:"host"`
			Max   int `ini:"Max" sync:"host"`
			Match struct {
				Wins int32 `ini:"Wins" sync:"host"`
			} `ini:"Match"`
			LoseOnKO bool `ini:"LoseOnKO" sync:"host"`
		} `ini:"Simul"`
		Tag struct {
			Min   int `ini:"Min" sync:"host"`
			Max   int `ini:"Max" sync:"host"`
			Match struct {
				Wins int32 `ini:"Wins" sync:"host"`
			} `ini:"Match"`
			LoseOnKO    bool    `ini:"LoseOnKO" sync:"host"`
			TimeScaling float32 `ini:"TimeScaling" sync:"host"`
		} `ini:"Tag"`
		Turns struct {
			Min      int `ini:"Min" sync:"host"`
			Max      int `ini:"Max" sync:"host"`
			Recovery struct {
				Base  float32 `ini:"Base" sync:"host"`
				Bonus float32 `ini:"Bonus" sync:"host"`
			} `ini:"Recovery"`
		} `ini:"Turns"`
	} `ini:"Options"`
	Config struct {
		Motif             string   `ini:"Motif" sync:"strict"`
		Players           int      `ini:"Players" sync:"host"`
		Language          string   `ini:"Language"`
		AfterImageMax     int32    `ini:"AfterImageMax" sync:"host"`
		ExplodMax         int      `ini:"ExplodMax" sync:"host"`
		HelperMax         int32    `ini:"HelperMax" sync:"host"`
		ProjectileMax     int      `ini:"ProjectileMax" sync:"host"`
		PaletteMax        int      `ini:"PaletteMax" sync:"host"`
		TextMax           int      `ini:"TextMax" sync:"host"`
		TickInterpolation bool     `ini:"TickInterpolation"`
		ZoomActive        bool     `ini:"ZoomActive" sync:"host"`
		EscOpensMenu      bool     `ini:"EscOpensMenu" sync:"host"`
		VsScreenLoading   bool     `ini:"VsScreenLoading" sync:"host"`
		TurnsLoading      bool     `ini:"TurnsLoading" sync:"host"`
		BootLoadingMode   int32    `ini:"BootLoadingMode" sync:"host"`
		FirstRun          bool     `ini:"FirstRun"`
		WindowTitle       string   `ini:"WindowTitle"`
		WindowIcon        []string `ini:"WindowIcon"`
		System            string   `ini:"System" sync:"strict"`
		ScreenshotFolder  string   `ini:"ScreenshotFolder"`
		TrainingChar      string   `ini:"TrainingChar"`
		TrainingStage     string   `ini:"TrainingStage"`
		GamepadMappings   string   `ini:"GamepadMappings"`
		LegacyTime        bool     `ini:"LegacyTime" sync:"host"`
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
		ForceStageZoomout   float32 `ini:"ForceStageZoomout" sync:"host"`
		ForceStageZoomin    float32 `ini:"ForceStageZoomin" sync:"host"`
		ForceStageAutoZoom  bool    `ini:"ForceStageAutoZoom" sync:"host"`
		KeepSpritesOnReload bool    `ini:"KeepSpritesOnReload"`
		RetryLoadOnError    bool    `ini:"RetryLoadOnError"`
		MacOSUseCommandKey  bool    `ini:"MacOSUseCommandKey"`
		SpeedTest           int     `ini:"SpeedTest"`
	} `ini:"Debug"`
	Video struct {
		RenderMode                string   `ini:"RenderMode"`
		GameWidth                 int32    `ini:"GameWidth" sync:"strict"`
		GameHeight                int32    `ini:"GameHeight" sync:"strict"`
		WindowWidth               int      `ini:"WindowWidth"`
		WindowHeight              int      `ini:"WindowHeight"`
		Framerate                 int      `ini:"Framerate"` // TODO: test if we need strict/host here
		VSync                     int      `ini:"VSync"`
		Fullscreen                bool     `ini:"Fullscreen"`
		Borderless                bool     `ini:"Borderless"`
		RGBSpriteBilinearFilter   bool     `ini:"RGBSpriteBilinearFilter"`
		MSAA                      int32    `ini:"MSAA"`
		WindowCentered            bool     `ini:"WindowCentered"`
		ExternalShaders           []string `ini:"ExternalShaders"`
		WindowScaleMode           bool     `ini:"WindowScaleMode"`
		FightAspectWidth          int32    `ini:"FightAspectWidth" sync:"strict"`
		FightAspectHeight         int32    `ini:"FightAspectHeight" sync:"strict"`
		KeepAspect                bool     `ini:"KeepAspect"`
		RendererDebugMode         bool     `ini:"RendererDebugMode"`
		EnableModel               bool     `ini:"EnableModel"`
		EnableModelShadow         bool     `ini:"EnableModelShadow"`
		LifeBarTextureAtlasSize   int32    `ini:"LifeBarTextureAtlasSize"`
		StageTextureAtlasSize     int32    `ini:"StageTextureAtlasSize"`
		CharacterTextureAtlasSize int32    `ini:"CharacterTextureAtlasSize"`
		FontTextureAtlasSize      int32    `ini:"FontTextureAtlasSize"`
		TextureAtlasSize          int32    `ini:"TextureAtlasSize"`
		ImageSuballocThresholdKB  int      `ini:"ImageSuballocThresholdKB"` // ≤ this (KB) suballocated; 0 to disable
		ImageSuballocBlockSizeMB  int      `ini:"ImageSuballocBlockSizeMB"` // block size (MB) for suballocation pool
		VRAMFallbackEnabled       bool     `ini:"VRAMFallbackEnabled"`      // Vulkan only: master switch for VRAM→HOST_VISIBLE fallback
		VRAMSwapThresholdPercent  int      `ini:"VRAMSwapThresholdPercent"` // Vulkan only: heap % that triggers swappable texture eviction
		VRAMSwapCooldownFrames    int      `ini:"VRAMSwapCooldownFrames"`   // Vulkan only: frames between swap attempts
	} `ini:"Video"`
	Sound struct {
		SampleRate           int32   `ini:"SampleRate"`
		SoundFont            string  `ini:"SoundFont"`
		StereoEffects        bool    `ini:"StereoEffects"`
		PanningRange         float32 `ini:"PanningRange"`
		WavChannels          int32   `ini:"WavChannels"`
		MasterVolume         int     `ini:"MasterVolume"`
		PauseMasterVolume    int     `ini:"PauseMasterVolume"`
		WavVolume            int     `ini:"WavVolume"`
		BGMVolume            int     `ini:"BGMVolume"`
		BGMRAMBuffer         bool    `ini:"BGMRAMBuffer"`
		MaxBGMVolume         int     `ini:"MaxBGMVolume"`
		AudioDucking         bool    `ini:"AudioDucking"`
		AudioResampleQuality int     `ini:"AudioResampleQuality"`
		FocusLossVolume      int     `ini:"FocusLossVolume"`
	} `ini:"Sound"`
	Arcade struct {
		AI struct {
			RandomColor   bool `ini:"RandomColor" sync:"host"`
			SurvivalColor bool `ini:"SurvivalColor" sync:"host"`
			Ramping       bool `ini:"Ramping" sync:"host"`
		} `ini:"AI"`
		//items map[string]AIrampProperties `ini:"items"`
		Arcade struct {
			AIramp AIrampProperties `ini:"AIramp"`
		} `ini:"arcade"`
		Team struct {
			AIramp AIrampProperties `ini:"AIramp"`
		} `ini:"team"`
		Survival struct {
			AIramp AIrampProperties `ini:"AIramp"`
		} `ini:"survival"`
	} `ini:"Arcade"`
	Netplay struct {
		ListenPort      int                `ini:"ListenPort"`
		SyncTimeout     int                `ini:"SyncTimeout" sync:"host"`
		LoadingTimeout  int                `ini:"LoadingTimeout" sync:"host"`
		RollbackNetcode bool               `ini:"RollbackNetcode" sync:"strict"`
		IP              map[string]string  `ini:"IP" insensitivekeys:"false"`
		Rollback        RollbackProperties `ini:"Rollback"`
	} `ini:"Netplay"`
	Input struct {
		// Device preferences, by keyboard controller ID or controller GUID.
		ControllerSetup map[string]string `ini:"ControllerSetup"`

		ControllerRumble           bool    `ini:"ControllerRumble"`
		ButtonAssist               bool    `ini:"ButtonAssist" sync:"host"`
		SOCDResolution             int     `ini:"SOCDResolution" sync:"host"`
		ControllerStickSensitivity float32 `ini:"ControllerStickSensitivity" sync:"host"`
		XinputTriggerSensitivity   float32 `ini:"XinputTriggerSensitivity" sync:"host"`
		UiRepeatDelay              int32   `ini:"UiRepeatDelay" sync:"host"`
		UiRepeatRate               int32   `ini:"UiRepeatRate" sync:"host"`
		PauseExitDelay             int32   `ini:"PauseExitDelay" sync:"host"`
	} `ini:"Input"`
	ButtonMapping   map[string]*ButtonMappingProperties   `ini:"map:^(?i)ButtonMapping_[a-z0-9_]+$" lua:"ButtonMapping"`
	ControllerSetup map[string]*ControllerSetupProperties `ini:"map:^(?i)ControllerSetup_[a-z0-9_]+$" lua:"ControllerSetup"`
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
		PreserveSurroundedQuote:    false,
		//KeyValueDelimiters: "=:",
		//KeyValueDelimiterOnWrite: "=",
		//ChildSectionDelimiter: ".",
		//AllowNonUniqueSections: true,
		//AllowDuplicateShadowValues: true,
	}

	// Preserve duplicates in user config so we can apply "first instance wins".
	userOptions := baseOptions
	userOptions.AllowShadows = true

	var err error

	defaultText := string(defaultConfig)
	if fp := FileExist("resources/defaultConfig.ini"); len(fp) != 0 {
		defaultText, err = LoadText(fp)
		if err != nil {
			return nil, fmt.Errorf("failed to load default config %s: %w", fp, err)
		}
	}
	// Load the INI file
	var iniFile *ini.File
	var defaultOnlyIni *ini.File
	var userIniFile *ini.File

	// Load defaults-only.
	defaultOnlyIni, err = LoadINIText(defaultText, baseOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to read defaults-only data: %v", err)
	}
	// Start merged INI as defaults, then overlay user (first-wins for duplicates).
	iniFile, err = LoadINIText(defaultText, baseOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to read base data: %v", err)
	}
	if fp := FileExist(def); len(fp) != 0 {
		userIniFile, _, err = LoadINIFile(fp, userOptions)
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
	c.dropDeletedSetups(userIniFile)
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
	c.SetValueUpdate("Netplay.ListenPort", int(Clamp(int32(c.Netplay.ListenPort), 1, 65535)))

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
	c.SetValueUpdate("Sound.AudioResampleQuality", Clamp(c.Sound.AudioResampleQuality, 1, 16))
	c.SetValueUpdate("Sound.MaxBGMVolume", int(Clamp(int32(c.Sound.MaxBGMVolume), 100, 250)))
	c.SetValueUpdate("Sound.PanningRange", Clamp(c.Sound.PanningRange, 0, 100))
	c.SetValueUpdate("Sound.PauseMasterVolume", int(Clamp(int32(c.Sound.PauseMasterVolume), 0, 100)))
	c.SetValueUpdate("Sound.FocusLossVolume", int(Clamp(int32(c.Sound.FocusLossVolume), 0, 100)))
	c.SetValueUpdate("Sound.WavChannels", Clamp(c.Sound.WavChannels, 1, 256))

	renderer, _ := selectRenderer(c.Video.RenderMode)
	renderMode := renderer.GetName()
	if strings.HasPrefix(renderMode, "Vulkan 1.3") {
		renderMode = "Vulkan 1.3"
	}
	if c.Video.RenderMode != renderMode {
		c.SetValueUpdate("Video.RenderMode", renderMode)
	}

	switch c.Video.MSAA {
	case 0, 2, 4, 6, 8, 16, 32:
	default:
		c.SetValueUpdate("Video.MSAA", 0)
	}

	if c.Video.ImageSuballocThresholdKB < 0 {
		c.SetValueUpdate("Video.ImageSuballocThresholdKB", 256)
	}
	if c.Video.ImageSuballocBlockSizeMB < 1 {
		c.SetValueUpdate("Video.ImageSuballocBlockSizeMB", 64)
	}
}

// Sets system values
func (c *Config) sysSet() {
	if _, ok := sys.cmdFlags["-width"]; ok {
		var w, _ = strconv.ParseInt(sys.cmdFlags["-width"], 10, 32)
		sys.gameWidth = float32(int32(w))
	} else {
		sys.gameWidth = float32(c.Video.GameWidth)
	}
	if _, ok := sys.cmdFlags["-height"]; ok {
		var h, _ = strconv.ParseInt(sys.cmdFlags["-height"], 10, 32)
		sys.gameHeight = float32(int32(h))
	} else {
		sys.gameHeight = float32(c.Video.GameHeight)
	}
	sys.msaa = c.Video.MSAA
	sys.loadInputConfig(c)
}

// Stores a Controller Setup and updates its INI section. Disk writes still use Config.Save.
func (c *Config) saveControllerSetup(id string, p *ControllerSetupProperties) error {
	id = strings.ToLower(id)
	key := "controllersetup_" + id
	section := iniSection(c.IniFile, key)
	if section == nil {
		var err error
		if section, err = c.IniFile.NewSection("ControllerSetup_" + id); err != nil {
			return err
		}
	}
	section.Key("Name").SetValue(p.Name)
	for _, field := range section.Keys() {
		if setupButtonMask(field.Name()) != 0 {
			if _, exists := p.Buttons[strings.ToLower(field.Name())]; !exists {
				section.DeleteKey(field.Name())
			}
		}
	}
	for _, name := range p.buttonNames() {
		section.Key(name).SetValue(parseVirtualControl(p.Buttons[name]).String())
	}
	setup := *p
	c.ControllerSetup[key] = &setup
	return nil
}

// Adds a setup with a unique ID and display name. A name already in use is
// renamed with copyName, where %s is the original name, then numbered if needed.
func (c *Config) addControllerSetup(p *ControllerSetupProperties, copyName string) (string, error) {
	id := ""
	for n := 1; ; n++ {
		id = fmt.Sprintf("custom%d", n)
		if _, exists := c.ControllerSetup["controllersetup_"+id]; !exists {
			break
		}
	}
	// %s in the name becomes the first free number.
	if strings.Contains(p.Name, "%s") {
		pattern := p.Name
		for n := 1; ; n++ {
			if p.Name = strings.ReplaceAll(pattern, "%s", strconv.Itoa(n)); !c.controllerSetupNameUsed(p.Name) {
				break
			}
		}
	}
	if c.controllerSetupNameUsed(p.Name) {
		base := strings.TrimSpace(strings.ReplaceAll(copyName, "%s", p.Name))
		if base == "" {
			base = p.Name
		}
		p.Name = base
		for n := 2; c.controllerSetupNameUsed(p.Name); n++ {
			p.Name = fmt.Sprintf("%s %d", base, n)
		}
	}
	return id, c.saveControllerSetup(id, p)
}

func (c *Config) controllerSetupNameUsed(name string) bool {
	for _, setup := range c.ControllerSetup {
		if strings.EqualFold(setup.Name, name) {
			return true
		}
	}
	return false
}

func (c *Config) removeControllerSetup(id string) {
	key := "controllersetup_" + strings.ToLower(id)
	delete(c.ControllerSetup, key)
	for device, setup := range c.Input.ControllerSetup {
		if setup == id {
			c.SetValueUpdate("Input.ControllerSetup."+device, nil)
		}
	}
	if sec := iniSection(c.IniFile, key); sec != nil {
		c.IniFile.DeleteSection(sec.Name())
	}
}

// Once setups have been saved, their section list is authoritative. Otherwise
// deleted default setups would come back on every launch.
func (c *Config) dropDeletedSetups(user *ini.File) {
	if user == nil {
		return
	}
	kept := make(map[string]bool)
	for _, sec := range user.Sections() {
		key := iniSectionID(sec)
		if _, ok := c.ControllerSetup[key]; ok {
			kept[key] = true
		}
	}
	if len(kept) == 0 {
		return
	}
	for key := range c.ControllerSetup {
		if !kept[key] {
			c.removeControllerSetup(strings.TrimPrefix(key, "controllersetup_"))
		}
	}
}

// Shipped values of a config.ini section, if defaultConfig.ini has it.
func (c *Config) shippedSection(key string, v interface{}) bool {
	if sec := iniSection(c.DefaultOnlyIni, key); sec != nil {
		if setup, ok := v.(*ControllerSetupProperties); ok {
			setup.read(sec)
			return true
		}
		return sec.MapTo(v) == nil
	}
	return false
}

// Stores a Button Mapping (a keyboard controller or a gamepad override).
func (c *Config) saveButtonMapping(id string, name string, m ButtonMapping) error {
	id = strings.ToLower(id)
	key := "buttonmapping_" + id
	section := iniSection(c.IniFile, key)
	if section == nil {
		var err error
		if section, err = c.IniFile.NewSection("ButtonMapping_" + id); err != nil {
			return err
		}
	}
	p := &ButtonMappingProperties{Name: name}
	p.setMapping(m)
	writeIniFields(section, p)
	c.ButtonMapping[key] = p
	return nil
}

// Drops a gamepad override, so the controller maps its controls to themselves again.
func (c *Config) removeButtonMapping(id string) {
	key := "buttonmapping_" + strings.ToLower(id)
	delete(c.ButtonMapping, key)
	if sec := iniSection(c.IniFile, key); sec != nil {
		c.IniFile.DeleteSection(sec.Name())
	}
}

// Saves one device preference straight to disk, so other unsaved Options changes stay pending.
// A value naming a section that is not on disk yet is kept until Options is saved.
func (c *Config) rememberPreference(group, device, value, section string) error {
	device = strings.ToLower(device)
	query := "Input." + group + "." + device
	if device == "" {
		return nil
	}
	var v interface{} = value
	if value == "" {
		v = nil
	}
	if err := c.SetValueUpdate(query, v); err != nil {
		return err
	}
	disk, _, err := LoadINIFile(c.Def, ini.LoadOptions{SkipUnrecognizableLines: true})
	if err != nil {
		return err
	}
	if value == "" {
		disk.Section("Input").DeleteKey(group + "." + device)
	} else if section != "" && iniSection(disk, section) == nil {
		return nil
	} else {
		disk.Section("Input").Key(group + "." + device).SetValue(value)
	}
	return SaveINI(disk, c.Def)
}

// Remembers the Controller Setup a device uses.
func (c *Config) rememberDeviceSetup(device, setup string) error {
	if c.Input.ControllerSetup[strings.ToLower(device)] == setup {
		return nil
	}
	return c.rememberPreference("ControllerSetup", device, setup, "controllersetup_"+setup)
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
