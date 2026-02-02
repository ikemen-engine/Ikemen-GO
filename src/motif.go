package main

import (
	_ "embed" // Support for go:embed resources
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/ini.v1"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

//go:embed resources/defaultMotif.ini
var defaultMotif []byte

// Motif parsing flow:
// 1. A Motif struct is created, all maps are initialized and `default` struct tags
//    are applied via applyDefaultsToValue, so every field has a well-defined base.
// 2. Two INI files are loaded: resources/defaultMotif.ini (defaults) and the
//    user motif .def. defaultOnlyIni holds only the embedded defaults, userIniFile
//    holds only user values, and iniFile is a merged view used for saving/lookup.
// 3. Values from defaultOnlyIni are assigned first, then values from userIniFile
//    overwrite them. INI section/key names are mapped to struct fields using `ini`
//    tags (including maps, pattern maps and flattening). Map fields can also opt
//    into "key-first" access using `keyfirst:"true"`, allowing "<key>.<field>"
//    to be treated as "<field>.<key>". After that, additional passes (custom
//    defaults, inheritance, localcoord fixes, font resolution, PopulateDataPointers,
//    etc.) adjust the final runtime data.

type PalFxProperties struct {
	Time        int32    `ini:"time" default:"-1"`
	Color       float32  `ini:"color" default:"256"`
	Hue         float32  `ini:"hue"`
	Add         [3]int32 `ini:"add"`
	Mul         [3]int32 `ini:"mul" default:"256,256,256"`
	SinAdd      [4]int32 `ini:"sinadd"`
	SinMul      [4]int32 `ini:"sinmul"`
	SinColor    [2]int32 `ini:"sincolor"`
	SinHue      [2]int32 `ini:"sinhue"`
	InvertAll   bool     `ini:"invertall"`
	InvertBlend int32    `ini:"invertblend"`
	PalFxData   *PalFX
}

type InfoProperties struct {
	Name          string   `ini:"name"`
	Author        string   `ini:"author"`
	VersionDate   string   `ini:"versiondate"`
	MugenVersion  string   `ini:"mugenversion"`
	IkemenVersion string   `ini:"ikemenversion"`
	Localcoord    [2]int32 `ini:"localcoord" default:"320,240"`
}

type FontProperties struct {
	Font    string    `ini:"" lua:"font" lookup:"def,font/,,data/"`
	Height  int32     `ini:"height" default:"-1"`
	Type    string    `ini:"type"`
	Size    [2]uint16 `ini:"size"`
	Spacing [2]int32  `ini:"spacing"`
	Offset  [2]int32  `ini:"offset"`
}

type FilesProperties struct {
	Spr     string `ini:"spr" lookup:"def,,data/"`
	Snd     string `ini:"snd" lookup:"def,,data/"`
	Loading struct {
		Storyboard string `ini:"storyboard" lookup:"def,,data/"`
	} `ini:"loading"`
	Logo struct {
		Storyboard string `ini:"storyboard" lookup:"def,,data/"`
	} `ini:"logo"`
	Intro struct {
		Storyboard string `ini:"storyboard" lookup:"def,,data/"`
	} `ini:"intro"`
	Select string                     `ini:"select" default:"select.def" lookup:"def,,data/"`
	Fight  string                     `ini:"fight" default:"fight.def" lookup:"def,,data/"`
	Font   map[string]*FontProperties `ini:"map:^(?i)font[0-9]+$" lua:"font"`
	Glyphs string                     `ini:"glyphs" lookup:"def,,data/"`
	Module string                     `ini:"module" lookup:"def,,data/"`
	Model  string                     `ini:"model" lookup:"def,,data/"`
}

type BgmProperties struct {
	Bgm           []string  `ini:"" lua:"bgm" lookup:"def,,data/,sound/"`
	Loop          []int32   `ini:"loop" default:"1"`
	Volume        []int32   `ini:"volume" default:"100"`
	LoopStart     []int32   `ini:"loopstart"`
	LoopEnd       []int32   `ini:"loopend"`
	StartPosition []int32   `ini:"startposition"`
	FreqMul       []float32 `ini:"freqmul" default:"1"`
	LoopCount     []int32   `ini:"loopcount" default:"-1"`
}

type FadeProperties struct {
	Time       int32    `ini:"time"`
	Col        [3]int32 `ini:"col"`
	Anim       int32    `ini:"anim" default:"-1"`
	Localcoord [2]int32 `ini:"localcoord"`
	AnimData   *Anim
	Snd        [2]int32 `ini:"snd" default:"-1,0"`
	FadeData   *Fade
}

type AnimationProperties struct {
	Anim        int32      `ini:"anim" default:"-1"`
	Spr         [2]int32   `ini:"spr" default:"-1,0"`
	Offset      [2]float32 `ini:"offset"`
	Facing      int32      `ini:"facing" default:"1"`
	Scale       [2]float32 `ini:"scale" default:"1,1"`
	Xshear      float32    `ini:"xshear"`
	Angle       float32    `ini:"angle"`
	XAngle      float32    `ini:"xangle"`
	YAngle      float32    `ini:"yangle"`
	Projection  string     `ini:"projection" default:"orthographic"`
	Focallength float32    `ini:"focallength" default:"2048"`
	Layerno     int16      `ini:"layerno"`
	Window      [4]int32   `ini:"window"`
	Localcoord  [2]int32   `ini:"localcoord"`
	AnimData    *Anim
}

type BgAnimationProperties struct {
	AnimationProperties
	Spacing [2]float32 `ini:"spacing"`
}

type AnimationTextProperties struct {
	AnimationProperties
	Font           [8]int32 `ini:"font" default:"-1,0,0,255,255,255,255,-1"`
	Text           string   `ini:"text"`
	TextSpriteData *TextSprite
}

type AnimationCharPreloadProperties struct {
	Anim        int32      `ini:"anim" default:"-1" preload:"char"`
	Spr         [2]int32   `ini:"spr" default:"-1,0" preload:"char"`
	Offset      [2]float32 `ini:"offset"`
	Facing      int32      `ini:"facing" default:"1"`
	Scale       [2]float32 `ini:"scale" default:"1,1"`
	Xshear      float32    `ini:"xshear"`
	Angle       float32    `ini:"angle"`
	XAngle      float32    `ini:"xangle"`
	YAngle      float32    `ini:"yangle"`
	Projection  string     `ini:"projection" default:"orthographic"`
	Focallength float32    `ini:"focallength" default:"2048"`
	Layerno     int16      `ini:"layerno"`
	Window      [4]int32   `ini:"window"`
	Localcoord  [2]int32   `ini:"localcoord"`
	AnimData    *Anim
	ApplyPal    bool  `ini:"applypal" preload:"pal"` // not used by [Select Info] portrait
	DrawOrder   int32 `ini:"draworder"`              // not used by [Select Info] portrait
}

type AnimationStagePreloadProperties struct {
	Anim        int32      `ini:"anim" default:"-1" preload:"stage"`
	Spr         [2]int32   `ini:"spr" default:"-1,0" preload:"stage"`
	Offset      [2]float32 `ini:"offset"`
	Facing      int32      `ini:"facing" default:"1"`
	Scale       [2]float32 `ini:"scale" default:"1,1"`
	Xshear      float32    `ini:"xshear"`
	Angle       float32    `ini:"angle"`
	XAngle      float32    `ini:"xangle"`
	YAngle      float32    `ini:"yangle"`
	Projection  string     `ini:"projection" default:"orthographic"`
	Focallength float32    `ini:"focallength" default:"2048"`
	Layerno     int16      `ini:"layerno"`
	Window      [4]int32   `ini:"window"`
	Localcoord  [2]int32   `ini:"localcoord"`
	AnimData    *Anim
}

type TextProperties struct {
	Font           [8]int32   `ini:"font" default:"-1,0,0,255,255,255,255,-1"`
	Offset         [2]float32 `ini:"offset"`
	Scale          [2]float32 `ini:"scale" default:"1,1"`
	Xshear         float32    `ini:"xshear"`
	Angle          float32    `ini:"angle"`
	XAngle         float32    `ini:"xangle"`
	YAngle         float32    `ini:"yangle"`
	Projection     string     `ini:"projection" default:"orthographic"`
	Focallength    float32    `ini:"focallength" default:"2048"`
	Text           string     `ini:"text"`
	Layerno        int16      `ini:"layerno"`
	Window         [4]int32   `ini:"window"`
	Localcoord     [2]int32   `ini:"localcoord"`
	TextSpriteData *TextSprite
}

type TextMapProperties struct {
	Font           [8]int32          `ini:"font" default:"-1,0,0,255,255,255,255,-1"`
	Offset         [2]float32        `ini:"offset"`
	Scale          [2]float32        `ini:"scale" default:"1,1"`
	Xshear         float32           `ini:"xshear"`
	Angle          float32           `ini:"angle"`
	XAngle         float32           `ini:"xangle"`
	YAngle         float32           `ini:"yangle"`
	Projection     string            `ini:"projection" default:"orthographic"`
	Focallength    float32           `ini:"focallength" default:"2048"`
	Text           map[string]string `ini:"text" keyfirst:"true"`
	Layerno        int16             `ini:"layerno"`
	Window         [4]int32          `ini:"window"`
	Localcoord     [2]int32          `ini:"localcoord"`
	TextSpriteData *TextSprite
}

type BoxBgProperties struct {
	Visible    bool     `ini:"visible"`
	Col        [3]int32 `ini:"col"`
	Alpha      [2]int32 `ini:"alpha" default:"0,255"`
	Layerno    int16    `ini:"layerno"`
	Localcoord [2]int32 `ini:"localcoord"`
	RectData   *Rect
}

type TweenProperties struct {
	Factor [2]float32 `ini:"factor"`
	Snap   bool       `ini:"snap"`
	Wrap   struct {
		Snap bool `ini:"snap"`
	} `ini:"wrap"`
}

type BoxCursorProperties struct {
	Visible    bool     `ini:"visible"`
	Coords     [4]int32 `ini:"coords"`
	Col        [3]int32 `ini:"col"`
	Layerno    int16    `ini:"layerno"`
	Localcoord [2]int32 `ini:"localcoord"`
	Pulse      [3]int32 `ini:"pulse"`
	//Alpha      [2]int32 `ini:"alpha" default:"0,255"`
	//Palfx      PalFxProperties `ini:"palfx"`
	RectData *Rect
	Tween    TweenProperties `ini:"tween"`
}

type OverlayProperties struct {
	Col        [3]int32 `ini:"col"`
	Alpha      [2]int32 `ini:"alpha" default:"0,255"`
	Layerno    int16    `ini:"layerno"`
	Window     [4]int32 `ini:"window"`
	Localcoord [2]int32 `ini:"localcoord"`
	RectData   *Rect
}

type BgDefProperties struct {
	Sff            *Sff
	BGDef          *BGDef
	Spr            string   `ini:"spr"`
	BgClearColor   [3]int32 `ini:"bgclearcolor" default:"-1,0,0"`
	BgClearAlpha   [2]int32 `ini:"bgclearalpha" default:"255,0"`
	BgClearLayerno int16    `ini:"bgclearlayerno" default:"0"`
	StartLayer     int32    `ini:"startlayer" default:"0"`
	Localcoord     [2]int32 `ini:"localcoord"`
	RectData       *Rect
}

type GlyphProperties struct {
	Spr    [2]int32   `ini:""` // default ini target: e.g. [Glyphs] ^3K = 63,0 -> Glyphs["^3K"].Spr = [63,0]
	Offset [2]float32 `ini:"offset"`
	//Facing     int32      `ini:"facing" default:"1"`
	Scale [2]float32 `ini:"scale" default:"1,1"`
	//Xshear     float32    `ini:"xshear"`
	//Angle      float32    `ini:"angle"`
	Layerno int16 `ini:"layerno"`
	//Window     [4]int32   `ini:"window"`
	Localcoord [2]int32 `ini:"localcoord"`
	AnimData   *Anim
	Size       [2]int32 `lua:"Size"`
}

type MenuProperties struct {
	Pos   [2]float32      `ini:"pos" default:"1000,900"`
	Tween TweenProperties `ini:"tween"`
	Item  struct {
		TextProperties
		Uppercase bool                              `ini:"uppercase"`
		Spacing   [2]float32                        `ini:"spacing"`
		Tween     TweenProperties                   `ini:"tween"`
		Bg        map[string]*BgAnimationProperties `ini:"bg" flatten:"true"`
		Active    struct {
			TextProperties
			Bg map[string]*BgAnimationProperties `ini:"bg" flatten:"true"`
		} `ini:"active"`
		Selected struct { // not used by [Title Info], [Option Info].keymenu, [Replay Info], [Attract Mode]
			TextProperties
			Active TextProperties `ini:"active"`
		} `ini:"selected"`
		Value struct { // not used by [Title Info], [Replay Info], [Attract Mode]
			TextProperties
			Active   TextProperties `ini:"active"`
			Conflict TextProperties `ini:"conflict"`
		} `ini:"value"`
		Info struct { // not used by [Title Info], [Replay Info], [Pause Menu], [Attract Mode]
			TextProperties
			Active TextProperties `ini:"active"`
		} `ini:"info"`
	} `ini:"item"`
	Window struct {
		Margins struct {
			Y [2]float32 `ini:"y"`
		} `ini:"margins"`
		VisibleItems int32 `ini:"visibleitems"`
	} `ini:"window"`
	BoxCursor BoxCursorProperties `ini:"boxcursor"`
	BoxBg     BoxBgProperties     `ini:"boxbg"`
	Arrow     struct {
		Up   AnimationProperties `ini:"up"`
		Down AnimationProperties `ini:"down"`
	} `ini:"arrow"`
	Title struct {
		Uppercase bool `ini:"uppercase"`
	} `ini:"title"`
	Next struct {
		Key []string `ini:"key"`
	} `ini:"next"`
	Previous struct {
		Key []string `ini:"key"`
	} `ini:"previous"`
	Add struct { // only used by [Option Info], [Pause Menu], [Training Pause Menu]
		Key []string `ini:"key"`
	} `ini:"add"`
	Subtract struct { // only used by [Option Info], [Pause Menu], [Training Pause Menu]
		Key []string `ini:"key"`
	} `ini:"subtract"`
	Cancel struct {
		Key []string `ini:"key"`
	} `ini:"cancel"`
	Done struct {
		Key []string `ini:"key"`
	} `ini:"done"`
	Hiscore struct { // not used by [Option Info], [Option Info].keymenu, [Replay Info], [Pause Menu]
		Key []string `ini:"key"`
	} `ini:"hiscore"`
	Unlock    map[string]string `ini:"unlock"`    // not used by [Option Info].keymenu, [Replay Info], [Pause Menu]
	Valuename map[string]string `ini:"valuename"` // not used by [Title Info], [Option Info].keymenu, [Replay Info], [Pause Menu], [Attract Mode]
	Itemname  map[string]string `ini:"itemname"`
}

type InfoBoxProperties struct {
	Title   TextProperties    `ini:"title"`
	Text    TextProperties    `ini:"text"`
	Overlay OverlayProperties `ini:"overlay"`
}

type CellOverrideProperties struct {
	Offset      [2]float32 `ini:"offset"`
	Spacing     [2]float32 `ini:"spacing"`
	Facing      int32      `ini:"facing" default:"1"`
	Skip        bool       `ini:"skip"`
	Scale       [2]float32 `ini:"scale"`
	XShear      float32    `ini:"xshear"`
	Angle       float32    `ini:"angle"`
	XAngle      float32    `ini:"xangle"`
	YAngle      float32    `ini:"yangle"`
	Projection  string     `ini:"projection" default:"orthographic"`
	FocalLength float32    `ini:"focallength"`
}

type TimerProperties struct {
	TextProperties
	Count          int32 `ini:"count"`
	Framespercount int32 `ini:"framespercount" default:"60"`
	Displaytime    int32 `ini:"displaytime"`
}

type PlayerCursorDoneProperties struct {
	AnimationProperties
	Snd [2]int32 `ini:"snd" default:"-1,0"`
}

type PlayerCursorProperties struct {
	StartCell [2]int32                               `ini:"startcell"`
	Active    map[string]*AnimationProperties        `ini:"active"`
	Done      map[string]*PlayerCursorDoneProperties `ini:"done"`
	Move      struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"move"`
	Blink      bool            `ini:"blink"`      // only used by P2
	SwitchTime int32           `ini:"switchtime"` // only used by P2
	Tween      TweenProperties `ini:"tween"`
	Reset      bool            `ini:"reset"`
}

type ItemProperties struct {
	TextMapProperties
	Spacing [2]float32      `ini:"spacing"`
	Tween   TweenProperties `ini:"tween"`
	Active  struct {
		TextMapProperties
		Switchtime int32 `ini:"switchtime"`
	} `ini:"active"`
	Active2   TextMapProperties   `ini:"active2"`
	Cursor    AnimationProperties `ini:"cursor"`    // only used by [Select Info].pX.teammenu.item
	Uppercase bool                `ini:"uppercase"` // only used by [Hiscore Info].item.name
}

type FaceProperties struct {
	AnimationCharPreloadProperties
	Done struct { // not used by [Victory Screen]
		AnimationCharPreloadProperties
		Key []string `ini:"key"` // only used by [VS Screen]
	} `ini:"done"`
	Random   AnimationProperties `ini:"random"` // only used by [Select Info]
	Velocity [2]float32          `ini:"velocity"`
	MaxDist  [2]float32          `ini:"maxdist"`
	Accel    [2]float32          `ini:"accel"`
	Friction [2]float32          `ini:"friction" default:"1,1"`
	Pos      [2]float32          `ini:"pos"`
	Num      int32               `ini:"num"`     // only used by P1-P2
	Spacing  [2]float32          `ini:"spacing"` // only used by P1-P2
	Padding  bool                `ini:"padding"` // only used by P1-P2
}

type PlayerSelectProperties struct {
	Cursor PlayerCursorProperties `ini:"cursor"`
	Random struct {
		Move struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"move"`
	} `ini:"random"`
	Face  FaceProperties `ini:"face"`
	Face2 FaceProperties `ini:"face2"`
	Name  struct {
		TextProperties
		Num     int32      `ini:"num"`     // only used by P1-P2
		Spacing [2]float32 `ini:"spacing"` // only used by P1-P2
		Random  struct {
			Text string `ini:"text"`
		} `ini:"random"`
	} `ini:"name"`
	Swap struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"swap"`
	Select struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"select"`
	TeamMenu struct { // only used by P1-P2
		Pos    [2]float32                        `ini:"pos"`
		Bg     map[string]*BgAnimationProperties `ini:"bg" flatten:"true"`
		Active struct {
			Bg map[string]*BgAnimationProperties `ini:"bg" flatten:"true"`
		} `ini:"active"`
		SelfTitle struct {
			AnimationTextProperties
		} `ini:"selftitle"`
		EnemyTitle struct {
			AnimationTextProperties
		} `ini:"enemytitle"`
		Move struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"move"`
		Value struct {
			Snd   [2]int32            `ini:"snd" default:"-1,0"`
			Icon  AnimationProperties `ini:"icon"`
			Empty struct {
				Icon AnimationProperties `ini:"icon"`
			} `ini:"empty"`
			Spacing [2]float32 `ini:"spacing"`
		} `ini:"value"`
		Done struct {
			Key []string `ini:"key"`
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"done"`
		Next struct {
			Key []string `ini:"key"`
		} `ini:"next"`
		Previous struct {
			Key []string `ini:"key"`
		} `ini:"previous"`
		Add struct {
			Key []string `ini:"key"`
		} `ini:"add"`
		Subtract struct {
			Key []string `ini:"key"`
		} `ini:"subtract"`
		Item   ItemProperties `ini:"item"`
		Ratio1 struct {
			Icon AnimationProperties `ini:"icon"`
		} `ini:"ratio1"`
		Ratio2 struct {
			Icon AnimationProperties `ini:"icon"`
		} `ini:"ratio2"`
		Ratio3 struct {
			Icon AnimationProperties `ini:"icon"`
		} `ini:"ratio3"`
		Ratio4 struct {
			Icon AnimationProperties `ini:"icon"`
		} `ini:"ratio4"`
		Ratio5 struct {
			Icon AnimationProperties `ini:"icon"`
		} `ini:"ratio5"`
		Ratio6 struct {
			Icon AnimationProperties `ini:"icon"`
		} `ini:"ratio6"`
		Ratio7 struct {
			Icon AnimationProperties `ini:"icon"`
		} `ini:"ratio7"`
	} `ini:"teammenu"`
	PalMenu struct {
		Pos  [2]float32 `ini:"pos"`
		Next struct {
			Key []string `ini:"key"`
		} `ini:"next"`
		Previous struct {
			Key []string `ini:"key"`
		} `ini:"previous"`
		Done struct {
			Key []string `ini:"key"`
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"done"`
		Cancel struct {
			Key []string `ini:"key"`
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"cancel"`
		Random struct {
			Key  []string `ini:"key"`
			Text string   `ini:"text"`
		} `ini:"random"`
		Value struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"value"`
		Preview struct {
			AnimationCharPreloadProperties
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"preview"`
		Number TextProperties      `ini:"number"`
		Text   TextProperties      `ini:"text"`
		Bg     AnimationProperties `ini:"bg"`
	} `ini:"palmenu"`
}

type TeamModesProperties struct {
	Single string `ini:"single"`
	Simul  string `ini:"simul"`
	Turns  string `ini:"turns"`
	Tag    string `ini:"tag"`
	Ratio  string `ini:"ratio"`
}

type ValueIconVsProperties struct {
	AnimationProperties
	Spacing [2]float32 `ini:"spacing"` // not used by value.empty.icon, only used by P1-P2
}

type PlayerVsProperties struct {
	FaceProperties
	Face2 FaceProperties `ini:"face2"`
	Key   []string       `ini:"key"`
	Name  struct {
		TextProperties
		Pos     [2]float32 `ini:"pos"`
		Num     int32      `ini:"num"`     // only used by P1-P2
		Spacing [2]float32 `ini:"spacing"` // only used by P1-P2
	} `ini:"name"`
	Icon struct {
		AnimationProperties
		Done AnimationProperties `ini:"done"`
	} `ini:"icon"`
	Value struct {
		Icon  ValueIconVsProperties `ini:"icon"`
		Empty struct {
			Icon ValueIconVsProperties `ini:"icon"`
		} `ini:"empty"`
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"value"`
}

type PlayerLoseProperties struct {
	FaceProperties
	Brightness int32 `ini:"brightness"`
}

type PlayerVictoryProperties struct {
	FaceProperties
	Lose  PlayerLoseProperties `ini:"lose"`
	Face2 struct {
		FaceProperties
		Lose PlayerLoseProperties `ini:"lose"`
	}
	Name     TextProperties `ini:"name"`
	State    []int32        `ini:"state"`
	Teammate struct {
		State []int32 `ini:"state"`
	} `ini:"teammate"`
}

type PlayerResultsProperties struct {
	State []int32  `ini:"state"`
	Win   struct { // not used by [Win Screen], [Time Attack Result Screen]
		State []int32 `ini:"state"`
	} `ini:"win"`
	Teammate struct {
		State []int32  `ini:"state"`
		Win   struct { // not used by [Win Screen], [Time Attack Result Screen]
			State []int32 `ini:"state"`
		} `ini:"win"`
	} `ini:"teammate"`
}

type PlayerDialogueProperties struct {
	Bg   AnimationProperties `ini:"bg"`
	Face struct {
		AnimationProperties
		Active AnimationProperties `ini:"active"`
	} `ini:"face"`
	Name TextProperties `ini:"name"`
	Text struct {
		TextProperties
		TextSpacing [2]float32 `ini:"textspacing"`
		TextDelay   int32      `ini:"textdelay"`
		TextWrap    string     `ini:"textwrap" default:"w"`
	} `ini:"text"`
	Active AnimationProperties `ini:"active"`
}

type TitleInfoProperties struct {
	FadeIn  FadeProperties `ini:"fadein"`
	FadeOut FadeProperties `ini:"fadeout"`
	Title   TextProperties `ini:"title"`
	Menu    MenuProperties `ini:"menu"`
	Cursor  struct {
		Move struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"move"`
		Done struct {
			Snd map[string][2]int32 `ini:"snd" default:"-1,0" keyfirst:"true"`
		} `ini:"done"`
	} `ini:"cursor"`
	Cancel struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
	Loading TextProperties `ini:"loading"`
	Footer  struct {
		Title   TextProperties    `ini:"title"`
		Info    TextProperties    `ini:"info"`
		Version TextProperties    `ini:"version"`
		Overlay OverlayProperties `ini:"overlay"`
	} `ini:"footer"`
	Connecting struct {
		TextMapProperties
		Overlay OverlayProperties `ini:"overlay"`
	} `ini:"connecting"`
	TextInput struct {
		TextMapProperties
		Overlay OverlayProperties `ini:"overlay"`
	} `ini:"textinput"`
}

type SelectInfoProperties struct {
	FadeIn               FadeProperties `ini:"fadein"`
	FadeOut              FadeProperties `ini:"fadeout"`
	Rows                 int32          `ini:"rows" default:"2"`
	Columns              int32          `ini:"columns" default:"5"`
	Wrapping             bool           `ini:"wrapping"`
	Pos                  [2]float32     `ini:"pos"`
	ShowEmptyBoxes       bool           `ini:"showemptyboxes"`
	MoveOverEmptyBoxes   bool           `ini:"moveoveremptyboxes"`
	SearchEmptyBoxesUp   bool           `ini:"searchemptyboxesup"`
	SearchEmptyBoxesDown bool           `ini:"searchemptyboxesdown"`
	Cell                 struct {
		Size    [2]int32   `ini:"size"`
		Spacing [2]float32 `ini:"spacing"`
		Up      struct {
			Key []string `ini:"key"`
		} `ini:"up"`
		Down struct {
			Key []string `ini:"key"`
		} `ini:"down"`
		Left struct {
			Key []string `ini:"key"`
		} `ini:"left"`
		Right struct {
			Key []string `ini:"key"`
		} `ini:"right"`
		Bg     AnimationProperties `ini:"bg"`
		Random struct {
			AnimationProperties
			SwitchTime int32 `ini:"switchtime"`
		} `ini:"random"`
		MapCell map[string]*CellOverrideProperties `ini:"map:^[0-9*]+-[0-9*]+$" lua:""`
	} `ini:"cell"`
	P1      PlayerSelectProperties `ini:"p1"`
	P2      PlayerSelectProperties `ini:"p2"`
	P3      PlayerSelectProperties `ini:"p3"`
	P4      PlayerSelectProperties `ini:"p4"`
	P5      PlayerSelectProperties `ini:"p5"`
	P6      PlayerSelectProperties `ini:"p6"`
	P7      PlayerSelectProperties `ini:"p7"`
	P8      PlayerSelectProperties `ini:"p8"`
	PalMenu struct {
		Random struct {
			SwitchTime int32 `ini:"switchtime"`
			ApplyPal   bool  `ini:"applypal" preload:"pal"`
		} `ini:"random"`
	} `ini:"palmenu"`
	Random struct {
		Move struct {
			Snd struct {
				Cancel bool `ini:"cancel"`
			} `ini:"snd"`
		} `ini:"move"`
	} `ini:"random"`
	Stage struct {
		Pos [2]float32 `ini:"pos"`
		TextProperties
		Random struct {
			Text string `ini:"text"`
		} `ini:"random"`
		Randomselect int32 `ini:"randomselect"`
		Active       struct {
			TextProperties
			Switchtime int32 `ini:"switchtime"`
		} `ini:"active"`
		Active2 TextProperties `ini:"active2"`
		Done    struct {
			TextProperties
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"done"`
		Move struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"move"`
		Portrait struct {
			AnimationStagePreloadProperties
			Bg     AnimationProperties `ini:"bg"`
			Random AnimationProperties `ini:"random"`
		} `ini:"portrait"`
	} `ini:"stage"`
	Done struct {
		Key []string `ini:"key"`
	} `ini:"done"`
	Cancel struct {
		Key []string `ini:"key"`
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
	Portrait AnimationCharPreloadProperties `ini:"portrait"`
	Title    TextMapProperties              `ini:"title"`
	Record   TextMapProperties              `ini:"record"`
	TeamMenu struct {
		Move struct {
			Wrapping bool `ini:"wrapping"`
		} `ini:"move"`
		Itemname map[string]*TeamModesProperties `ini:"itemname"`
	} `ini:"teammenu"`
	Timer         TimerProperties `ini:"timer"`
	PaletteSelect int32           `ini:"paletteselect"`
}

type VsScreenProperties struct {
	FadeIn      FadeProperties `ini:"fadein"`
	FadeOut     FadeProperties `ini:"fadeout"`
	Time        int32          `ini:"time"`
	Match       TextProperties `ini:"match"`
	OrderSelect struct {
		Enabled bool `ini:"enabled"`
	} `ini:"orderselect"`
	P1    PlayerVsProperties `ini:"p1"`
	P2    PlayerVsProperties `ini:"p2"`
	P3    PlayerVsProperties `ini:"p3"`
	P4    PlayerVsProperties `ini:"p4"`
	P5    PlayerVsProperties `ini:"p5"`
	P6    PlayerVsProperties `ini:"p6"`
	P7    PlayerVsProperties `ini:"p7"`
	P8    PlayerVsProperties `ini:"p8"`
	Timer TimerProperties    `ini:"timer"`
	Done  struct {
		Key  []string `ini:"key"`
		Time int32    `ini:"time"`
	} `ini:"done"`
	Skip struct {
		Key []string `ini:"key"`
	} `ini:"skip"`
	Cancel struct {
		Key []string `ini:"key"`
	} `ini:"cancel"`
	Stage struct {
		Pos [2]float32 `ini:"pos"`
		TextProperties
		Portrait struct {
			AnimationStagePreloadProperties
			Bg AnimationProperties `ini:"bg"`
		} `ini:"portrait"`
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"stage"`
}

type DemoModeProperties struct {
	Enabled bool           `ini:"enabled"`
	FadeIn  FadeProperties `ini:"fadein"`
	FadeOut FadeProperties `ini:"fadeout"`
	Title   struct {
		WaitTime int32 `ini:"waittime"`
	} `ini:"title"`
	Fight struct {
		EndTime int32 `ini:"endtime"`
		PlayBgm bool  `ini:"playbgm"`
		StopBgm bool  `ini:"stopbgm"`
		Bars    struct {
			Display bool `ini:"display"`
		} `ini:"bars"`
	} `ini:"fight"`
	Intro struct {
		WaitCycles int32 `ini:"waitcycles"`
	} `ini:"intro"`
	DebugInfo bool `ini:"debuginfo"`
	Select    struct {
		Enabled bool `ini:"enabled"`
	} `ini:"select"`
	VsScreen struct {
		Enabled bool `ini:"enabled"`
	} `ini:"vsscreen"`
	Cancel struct {
		Key []string `ini:"key"`
	} `ini:"cancel"`
}

type ContinueCounterProperties struct {
	SkipTime int32    `ini:"skiptime"`
	Snd      [2]int32 `ini:"snd" default:"-1,0"`
}

type ContinueScreenProperties struct {
	Enabled  bool           `ini:"enabled"`
	FadeIn   FadeProperties `ini:"fadein"`
	FadeOut  FadeProperties `ini:"fadeout"`
	Pos      [2]float32     `ini:"pos"`
	Continue TextProperties `ini:"continue"`
	Yes      struct {
		TextProperties
		Active TextProperties `ini:"active"`
	} `ini:"yes"`
	No struct {
		TextProperties
		Active TextProperties `ini:"active"`
	} `ini:"no"`
	Sounds struct {
		Enabled bool `ini:"enabled"`
	} `ini:"sounds"`
	LegacyMode struct {
		Enabled bool `ini:"enabled" default:"true"`
	} `ini:"legacymode"`
	GameOver struct {
		Enabled bool `ini:"enabled"`
	} `ini:"gameover"`
	Move struct {
		Key []string `ini:"key"`
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"move"`
	Cancel struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
	Done struct {
		Key []string `ini:"key"`
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"done"`
	Skip struct {
		Key []string `ini:"key"`
	} `ini:"skip"`
	Overlay OverlayProperties `ini:"overlay"`
	P1      struct {
		State []int32 `ini:"state"`
		Yes   struct {
			State []int32 `ini:"state"`
		} `ini:"yes"`
		No struct {
			State []int32 `ini:"state"`
		} `ini:"no"`
		Teammate struct {
			State []int32 `ini:"state"`
			Yes   struct {
				State []int32 `ini:"state"`
			} `ini:"yes"`
			No struct {
				State []int32 `ini:"state"`
			} `ini:"no"`
		} `ini:"teammate"`
	} `ini:"p1"`
	P2 struct {
		State []int32 `ini:"state"`
		Yes   struct {
			State []int32 `ini:"state"`
		} `ini:"yes"`
		No struct {
			State []int32 `ini:"state"`
		} `ini:"no"`
		Teammate struct {
			State []int32 `ini:"state"`
			Yes   struct {
				State []int32 `ini:"state"`
			} `ini:"yes"`
			No struct {
				State []int32 `ini:"state"`
			} `ini:"no"`
		} `ini:"teammate"`
	} `ini:"p2"`
	Credits TextProperties `ini:"credits"`
	Counter struct {
		AnimationProperties
		StartTime int32 `ini:"starttime"`
		EndTime   int32 `ini:"endtime"`
		Default   struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"default"`
		SkipStart int32                                 `ini:"skipstart"`
		MapCounts map[string]*ContinueCounterProperties `ini:"map:^[0-9]+$" lua:""`
		End       ContinueCounterProperties             `ini:"end"`
	} `ini:"counter"`
}

type StoryboardProperties struct {
	Enabled    bool   `ini:"enabled"`
	Storyboard string `ini:"storyboard" lookup:"def,,data/"`
}

type VictoryScreenProperties struct {
	Enabled  bool `ini:"enabled"`
	KeepSide struct {
		Enabled bool `ini:"enabled"`
	} `ini:"keepside"`
	Sounds struct {
		Enabled bool `ini:"enabled"`
	} `ini:"sounds"`
	Cpu struct {
		Enabled bool `ini:"enabled"`
	} `ini:"cpu"`
	Vs struct {
		Enabled bool `ini:"enabled"`
	} `ini:"vs"`
	Winner struct {
		TeamKo struct {
			Enabled bool `ini:"enabled"`
		} `ini:"teamko"`
	} `ini:"winner"`
	FadeIn   FadeProperties `ini:"fadein"`
	FadeOut  FadeProperties `ini:"fadeout"`
	Time     int32          `ini:"time"`
	WinName  TextProperties `ini:"winname"`
	WinQuote struct {
		TextProperties
		TextSpacing [2]float32 `ini:"textspacing"`
		TextDelay   int32      `ini:"textdelay"`
		TextWrap    string     `ini:"textwrap" default:"w"`
		DisplayTime int32      `ini:"displaytime"`
	} `ini:"winquote"`
	Overlay OverlayProperties `ini:"overlay"`
	Skip    struct {
		Key []string `ini:"key"`
	} `ini:"skip"`
	Cancel struct {
		Key []string `ini:"key"`
	} `ini:"cancel"`
	P1 PlayerVictoryProperties `ini:"p1"`
	P2 PlayerVictoryProperties `ini:"p2"`
	P3 PlayerVictoryProperties `ini:"p3"`
	P4 PlayerVictoryProperties `ini:"p4"`
	P5 PlayerVictoryProperties `ini:"p5"`
	P6 PlayerVictoryProperties `ini:"p6"`
	P7 PlayerVictoryProperties `ini:"p7"`
	P8 PlayerVictoryProperties `ini:"p8"`
}

type WinScreenProperties struct {
	Enabled bool `ini:"enabled"`
	Sounds  struct {
		Enabled bool `ini:"enabled"`
	} `ini:"sounds"`
	Results map[string]string `ini:"results"`
	FadeIn  FadeProperties    `ini:"fadein"`
	FadeOut FadeProperties    `ini:"fadeout"`
	Pose    struct {
		Time int32 `ini:"time"`
	} `ini:"pose"`
	State struct {
		Time int32 `ini:"time"`
	} `ini:"state"`
	WinText struct {
		TextProperties
		DisplayTime int32 `ini:"displaytime"`
	} `ini:"wintext"`
	Overlay OverlayProperties `ini:"overlay"`
	Cancel  struct {
		Key []string `ini:"key"`
	} `ini:"cancel"`
	P1 PlayerResultsProperties `ini:"p1"`
	P2 PlayerResultsProperties `ini:"p2"`
}

type ResultsScreenProperties struct {
	Enabled bool `ini:"enabled"`
	Sounds  struct {
		Enabled bool `ini:"enabled"`
	} `ini:"sounds"`
	RoundsToWin int32          `ini:"roundstowin"` // used only by [Survival Results Screen]
	FadeIn      FadeProperties `ini:"fadein"`
	FadeOut     FadeProperties `ini:"fadeout"`
	Show        struct {
		Time int32 `ini:"time"`
	} `ini:"show"`
	State struct {
		Time int32 `ini:"time"`
	} `ini:"state"`
	WinsText struct {
		TextProperties
		DisplayTime int32 `ini:"displaytime"`
	} `ini:"winstext"`
	Overlay OverlayProperties `ini:"overlay"`
	Cancel  struct {
		Key []string `ini:"key"`
	} `ini:"cancel"`
	P1 PlayerResultsProperties `ini:"p1"`
	P2 PlayerResultsProperties `ini:"p2"`
}

type OptionInfoProperties struct {
	FadeIn  FadeProperties `ini:"fadein"`
	FadeOut FadeProperties `ini:"fadeout"`
	Title   TextProperties `ini:"title"`
	Menu    MenuProperties `ini:"menu"`
	Cursor  struct {
		Move struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"move"`
		Done struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"done"`
	} `ini:"cursor"`
	Cancel struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
	TextInput struct {
		TextMapProperties
		Overlay OverlayProperties `ini:"overlay"`
	} `ini:"textinput"`
	KeyMenu struct {
		P1 struct {
			MenuOffset [2]float32     `ini:"menuoffset"`
			Playerno   TextProperties `ini:"playerno"`
		} `ini:"p1"`
		P2 struct {
			MenuOffset [2]float32     `ini:"menuoffset"`
			Playerno   TextProperties `ini:"playerno"`
		} `ini:"p2"`
		MenuProperties
	} `ini:"keymenu"`
	Itemname map[string]string `ini:"itemname"` // not used by [Option Info]
}

type ReplayInfoProperties struct {
	FadeIn  FadeProperties `ini:"fadein"`
	FadeOut FadeProperties `ini:"fadeout"`
	Title   TextProperties `ini:"title"`
	Menu    MenuProperties `ini:"menu"`
	Cursor  struct {
		Move struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"move"`
		Done struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"done"`
	} `ini:"cursor"`
	Cancel struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
}

type MovelistGlyphsProperties struct {
	Offset     [2]float32 `ini:"offset"`
	Scale      [2]float32 `ini:"scale" default:"1,1"`
	Layerno    int16      `ini:"layerno"`
	Localcoord [2]int32   `ini:"localcoord"`
	Spacing    [2]float32 `ini:"spacing"`
}

type MenuInfoProperties struct {
	Enabled  bool           `ini:"enabled"`
	FadeIn   FadeProperties `ini:"fadein"`
	FadeOut  FadeProperties `ini:"fadeout"`
	Title    TextProperties `ini:"title"`
	Menu     MenuProperties `ini:"menu"`
	HideBars bool           `ini:"hidebars"`
	Cursor   struct {
		Move struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"move"`
		Done struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"done"`
	} `ini:"cursor"`
	Cancel struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
	Enter struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"enter"`
	Exit struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"exit"`
	Overlay  OverlayProperties `ini:"overlay"`
	Movelist struct {
		Pos   [2]float32 `ini:"pos"`
		Title struct {
			TextProperties
			Uppercase bool `ini:"uppercase"`
		} `ini:"title"`
		Text struct {
			TextProperties
			Spacing [2]float32 `ini:"spacing"`
		} `ini:"text"`
		Glyphs MovelistGlyphsProperties `ini:"glyphs"`
		Window struct {
			Margins struct {
				Y [2]float32 `ini:"y"`
			} `ini:"margins"`
			VisibleItems int32 `ini:"visibleitems"`
			Width        int32 `ini:"width"`
		} `ini:"window"`
		Overlay OverlayProperties `ini:"overlay"`
		Arrow   struct {
			Up   AnimationProperties `ini:"up"`
			Down AnimationProperties `ini:"down"`
		} `ini:"arrow"`
		Itemname map[string]string `ini:"itemname"`
	} `ini:"movelist"`
}

type AttractModeProperties struct {
	Enabled bool           `ini:"enabled"`
	FadeIn  FadeProperties `ini:"fadein"`
	FadeOut FadeProperties `ini:"fadeout"`
	Title   TextProperties `ini:"title"`
	Menu    MenuProperties `ini:"menu"`
	Cursor  struct {
		Move struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"move"`
		Done struct {
			Snd map[string][2]int32 `ini:"snd" default:"-1,0" keyfirst:"true"`
		} `ini:"done"`
		Snd map[string][2]int32 `ini:"snd"`
	} `ini:"cursor"`
	Cancel struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
	Credits struct {
		TextProperties
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"credits"`
	Logo struct {
		Storyboard string `ini:"storyboard" lookup:"def,,data/"`
	} `ini:"logo"`
	Intro struct {
		Storyboard string `ini:"storyboard" lookup:"def,,data/"`
	} `ini:"intro"`
	Start struct {
		Storyboard string `ini:"storyboard" lookup:"def,,data/"`
		Done       struct {
			Snd [2]int32 `ini:"snd" default:"-1,0"`
		} `ini:"done"`
		Time   int32 `ini:"time"`
		Insert struct {
			TextProperties
			Blinktime int32 `ini:"blinktime"`
		} `ini:"insert"`
		Press struct {
			TextProperties
			Blinktime int32    `ini:"blinktime"`
			Key       []string `ini:"key"`
		} `ini:"press"`
		Timer TimerProperties `ini:"timer"`
	} `ini:"start"`
	Options struct {
		KeyCode string `ini:"keycode"`
	} `ini:"options"`
}

type ChallengerInfoProperties struct {
	Enabled bool           `ini:"enabled"`
	FadeIn  FadeProperties `ini:"fadein"`
	FadeOut FadeProperties `ini:"fadeout"`
	Time    int32          `ini:"time"`
	Pause   struct {
		Time int32 `ini:"time"`
	} `ini:"pause"`
	Snd struct {
		Snd  [2]int32 `ini:"" default:"-1,0"`
		Time int32    `ini:"time"`
	} `ini:"snd"`
	Key  []string `ini:"key"`
	Text struct {
		TextProperties
		Displaytime int32 `ini:"displaytime"`
	} `ini:"text"`
	Bg struct {
		AnimationProperties
		Displaytime int32 `ini:"displaytime"`
	} `ini:"bg"`
	Overlay OverlayProperties `ini:"overlay"`
}

type DialogueInfoProperties struct {
	Enabled    bool  `ini:"enabled"`
	StartTime  int32 `ini:"starttime"`
	EndTime    int32 `ini:"endtime"`
	SwitchTime int32 `ini:"switchtime"`
	SkipTime   int32 `ini:"skiptime"`
	Skip       struct {
		Key []string `ini:"key"`
	} `ini:"skip"`
	Cancel struct {
		Key []string `ini:"key"`
	} `ini:"cancel"`
	P1 PlayerDialogueProperties `ini:"p1"`
	P2 PlayerDialogueProperties `ini:"p2"`
}

type HiscoreInfoProperties struct {
	Enabled bool              `ini:"enabled"`
	FadeIn  FadeProperties    `ini:"fadein"`
	FadeOut FadeProperties    `ini:"fadeout"`
	Time    int32             `ini:"time"`
	Pos     [2]float32        `ini:"pos"`
	Ranking map[string]string `ini:"ranking"`
	Title   struct {
		TextMapProperties
		Uppercase bool           `ini:"uppercase"`
		Rank      TextProperties `ini:"rank"`
		Result    TextProperties `ini:"result"`
		Name      TextProperties `ini:"name"`
		Face      TextProperties `ini:"face"`
	} `ini:"title"`
	Item struct {
		Offset  [2]float32     `ini:"offset"`
		Spacing [2]float32     `ini:"spacing"`
		Rank    ItemProperties `ini:"rank"`
		Result  ItemProperties `ini:"result"`
		Name    ItemProperties `ini:"name"`
		Face    struct {
			AnimationCharPreloadProperties
			Num     int32               `ini:"num"`
			Spacing [2]float32          `ini:"spacing"`
			Bg      AnimationProperties `ini:"bg"`
			Unknown AnimationProperties `ini:"unknown"`
		} `ini:"face"`
	} `ini:"item"`
	Timer  TimerProperties `ini:"timer"`
	Window struct {
		Margins struct {
			Y [2]float32 `ini:"y"`
		} `ini:"margins"`
		VisibleItems int32 `ini:"visibleitems"`
		Width        int32 `ini:"width"`
	} `ini:"window"`
	Overlay OverlayProperties `ini:"overlay"`
	Next    struct {
		Key []string `ini:"key"`
	} `ini:"next"`
	Previous struct {
		Key []string `ini:"key"`
	} `ini:"previous"`
	Done struct {
		Key  []string `ini:"key"`
		Snd  [2]int32 `ini:"snd" default:"-1,0"`
		Time int32    `ini:"time"`
	} `ini:"done"`
	Cancel struct {
		Key []string `ini:"key"`
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
	Move struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"move"`
	Glyphs []string `ini:"glyphs"`
}

type WarningInfoProperties struct {
	Title   TextProperties    `ini:"title"`
	Text    TextMapProperties `ini:"text"`
	Overlay OverlayProperties `ini:"overlay"`
	Cancel  struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"cancel"`
	Done struct {
		Snd [2]int32 `ini:"snd" default:"-1,0"`
	} `ini:"done"`
}

type Motif struct {
	IniFile         *ini.File
	UserIniFile     *ini.File
	DefaultOnlyIni  *ini.File
	AnimTable       AnimationTable
	Sff             *Sff
	Snd             *Snd
	Fnt             map[int]*Fnt
	GlyphsSff       *Sff
	Model           *Model
	Music           Music
	Def             string                              `ini:"def"`
	Info            InfoProperties                      `ini:"info"`
	Files           FilesProperties                     `ini:"files"`
	Languages       map[string]string                   `ini:"languages"`
	TitleInfo       TitleInfoProperties                 `ini:"title_info"`
	TitleBgDef      BgDefProperties                     `ini:"titlebgdef"`
	InfoBox         InfoBoxProperties                   `ini:"infobox"`
	SelectInfo      SelectInfoProperties                `ini:"select_info"`
	SelectBgDef     BgDefProperties                     `ini:"selectbgdef"`
	VsScreen        VsScreenProperties                  `ini:"vs_screen"`
	VersusBgDef     BgDefProperties                     `ini:"versusbgdef"`
	DemoMode        DemoModeProperties                  `ini:"demo_mode"`
	ContinueScreen  ContinueScreenProperties            `ini:"continue_screen"`
	ContinueBgDef   BgDefProperties                     `ini:"continuebgdef"`
	GameOverScreen  StoryboardProperties                `ini:"game_over_screen"`
	VictoryScreen   VictoryScreenProperties             `ini:"victory_screen"`
	VictoryBgDef    BgDefProperties                     `ini:"victorybgdef"`
	WinScreen       WinScreenProperties                 `ini:"win_screen"`
	WinBgDef        BgDefProperties                     `ini:"winbgdef"`
	DefaultEnding   StoryboardProperties                `ini:"default_ending"`
	EndCredits      StoryboardProperties                `ini:"end_credits"`
	ResultsScreen   map[string]*ResultsScreenProperties `ini:"map:^(?i).+results.*screen$" lua:"results_screen"`
	ResultsBgDef    map[string]*BgDefProperties         `ini:"map:^(?i).+resultsbgdef$" lua:"resultsbgdef"`
	OptionInfo      OptionInfoProperties                `ini:"option_info"`
	OptionBgDef     BgDefProperties                     `ini:"optionbgdef"`
	ReplayInfo      ReplayInfoProperties                `ini:"replay_info"`
	ReplayBgDef     BgDefProperties                     `ini:"replaybgdef"`
	PauseMenu       map[string]*MenuInfoProperties      `ini:"map:^(?i).*pause.*menu$" lua:"pause_menu"`
	PauseBgDef      map[string]*BgDefProperties         `ini:"map:^(?i).*pausebgdef$" lua:"pausebgdef"`
	AttractMode     AttractModeProperties               `ini:"attract_mode"`
	AttractBgDef    BgDefProperties                     `ini:"attractbgdef"`
	ChallengerInfo  ChallengerInfoProperties            `ini:"challenger_info"`
	ChallengerBgDef BgDefProperties                     `ini:"challengerbgdef"`
	DialogueInfo    DialogueInfoProperties              `ini:"dialogue_info"`
	HiscoreInfo     HiscoreInfoProperties               `ini:"hiscore_info"`
	HiscoreBgDef    BgDefProperties                     `ini:"hiscorebgdef"`
	WarningInfo     WarningInfoProperties               `ini:"warning_info"`
	Glyphs          map[string]*GlyphProperties         `ini:"glyphs" literal:"true" insensitivekeys:"false" sff:"GlyphsSff"`
	fntIndexByKey   map[string]int                      // filepath|height -> index
	ch              MotifChallenger
	co              MotifContinue
	de              MotifDemo
	di              MotifDialogue
	vi              MotifVictory
	wi              MotifWin
	hi              MotifHiscore
	me              MotifMenu
	fadeIn          *Fade
	fadeOut         *Fade
	fadePolicy      FadeStartPolicy
}

// hasUserKey returns true if the given key exists in `section` of the INI.
func hasUserKey(iniFile *ini.File, section, key string) bool {
	sec, err := iniFile.GetSection(section)
	if err != nil {
		return false
	}
	return sec.HasKey(key)
}

// preprocessINIContent removes or modifies specific sections before parsing.
func preprocessINIContent(input string) string {
	// Define a regex to find the [Infobox Text] section
	infoboxRegex := regexp.MustCompile(`(?s)\[Infobox Text\]\n(.*?)(\n\[|$)`)
	// Extract the content of [Infobox Text]
	matches := infoboxRegex.FindStringSubmatch(input)
	if len(matches) < 3 {
		// If the section is not found, return the original input
		return input
	}
	infoboxTextContent := matches[1]
	// Process the extracted text
	processedText := strings.TrimSpace(infoboxTextContent)
	processedText = strings.ReplaceAll(processedText, "\n", `\n`)
	// Resolve first two %s placeholders to Version and BuildTime
	processedText = strings.Replace(processedText, "%s", Version, 1)
	processedText = strings.Replace(processedText, "%s", BuildTime, 1)
	// Create the new text.text line with an added newline at the end
	newTextLine := fmt.Sprintf("\ttext.text = %s\n\n", processedText)
	// Remove the [Infobox Text] section from the input
	output := infoboxRegex.ReplaceAllString(input, "$2")
	// Define a regex to find the [InfoBox] section header
	infoBoxHeaderRegex := regexp.MustCompile(`(?im)(\[InfoBox\]\n)`)
	// Insert the new text.text line right after the [InfoBox] header
	output = infoBoxHeaderRegex.ReplaceAllString(output, "${1}"+newTextLine)
	return output
}

// applyCustomDefaults injects custom defaults.
func applyCustomDefaults(m *Motif, iniFile *ini.File) {
	// Copy the first spacing value into the second one when only a single value is defined in system.def.
	sec := "Select Info"
	// Only act if the user actually set the key in their system.def
	s, err := iniFile.GetSection(sec)
	if err != nil || s == nil {
		return
	}
	k, err := s.GetKey("cell.spacing")
	if err != nil {
		return
	}
	raw := strings.TrimSpace(k.Value())
	// If user specified a single value (no comma/&), duplicate it into the second slot.
	if raw != "" && !strings.Contains(raw, ",") && !strings.Contains(raw, "&") {
		// At this point, the first component has already been parsed into the struct.
		// Just mirror it to the second component.
		m.SelectInfo.Cell.Spacing[1] = m.SelectInfo.Cell.Spacing[0]
	}

	// Inject computed defaults based on system.def localcoord when specific keys in system.def are not defined.
	w := int(m.Info.Localcoord[0])
	h := int(m.Info.Localcoord[1])
	sec = "Title Info"

	// loading.offset = (W - 1 - 10*H/320, H - 8)
	if !hasUserKey(iniFile, sec, "loading.offset") {
		x := w - 1 - (10*h)/320
		y := h - 8
		_ = SetValue(m, "title_info.loading.offset", fmt.Sprintf("%d, %d", x, y))
	}

	// footer.title.offset = (2*W/320, H)
	if !hasUserKey(iniFile, sec, "footer.title.offset") {
		x := (2 * w) / 320
		y := h
		_ = SetValue(m, "title_info.footer.title.offset", fmt.Sprintf("%d, %d", x, y))
	}

	// footer.info.offset = (W/2, H)
	if !hasUserKey(iniFile, sec, "footer.info.offset") {
		x := w / 2
		y := h
		_ = SetValue(m, "title_info.footer.info.offset", fmt.Sprintf("%d, %d", x, y))
	}

	// footer.version.offset = (W - 1 - 2*W/320, H)
	if !hasUserKey(iniFile, sec, "footer.version.offset") {
		x := w - 1 - (2*w)/320
		y := h
		_ = SetValue(m, "title_info.footer.version.offset", fmt.Sprintf("%d, %d", x, y))
	}

	// footer.version.text
	if !hasUserKey(iniFile, sec, "footer.version.text") {
		_ = SetValue(m, "title_info.footer.version.text", fmt.Sprintf("%s - %s", Version, BuildTime))
	}

	// footer.overlay.window = (0, H - 7, W, H)
	if !hasUserKey(iniFile, sec, "footer.overlay.window") {
		x1, y1, x2, y2 := 0, h-7, w, h
		_ = SetValue(m, "title_info.footer.overlay.window", fmt.Sprintf("%d, %d, %d, %d", x1, y1, x2, y2))
	}
}

// reserveUserFontSlots marks user-specified [Files] font indices as taken so
// resolveInlineFonts won't auto-assign inline fonts to those slots.
func reserveUserFontSlots(m *Motif) {
	if m == nil || m.UserIniFile == nil {
		return
	}
	sec, err := m.UserIniFile.GetSection("Files")
	if err != nil || sec == nil {
		return
	}
	re := regexp.MustCompile(`(?i)^font(\d+)$`)
	for _, k := range sec.Keys() {
		name := k.Name()
		match := re.FindStringSubmatch(name)
		if len(match) == 2 {
			idx := int(Atoi(match[1]))
			if _, exists := m.Fnt[idx]; !exists {
				m.Fnt[idx] = nil // placeholder to block ensureFontIndex from using this slot
			}
		}
	}
}

// returns a stable-sorted list of files named "system.def" located anywhere under the given root (e.g. external/mods), including nested subdirs.
func findExternalModSystemDefs(root string) ([]string, error) {
	st, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !st.IsDir() {
		return nil, nil
	}

	var out []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		// Match filename, case-insensitive
		if strings.EqualFold(d.Name(), "system.def") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// loadMotif loads and parses the INI file into a Motif struct.
func loadMotif(def string) (*Motif, error) {
	if def == "" {
		sys.motif.resolvePath()
		def = sys.motif.Def
	}
	// Define load options if needed
	// https://github.com/go-ini/ini/blob/main/ini.go
	baseOptions := ini.LoadOptions{
		Insensitive:             false,
		InsensitiveSections:     true,
		InsensitiveKeys:         false,
		IgnoreInlineComment:     false,
		SkipUnrecognizableLines: true,
		//AllowBooleanKeys: true,
		AllowShadows: false,
		//AllowNestedValues: true,
		UnparseableSections:        []string{"Infobox Text"},
		AllowPythonMultilineValues: false,
		//KeyValueDelimiters: "=:",
		//KeyValueDelimiterOnWrite: "=",
		//ChildSectionDelimiter: ".",
		//AllowNonUniqueSections: true,
		//AllowDuplicateShadowValues: true,
	}

	// Preserve duplicates in user input so we can apply "first instance wins".
	userOptions := baseOptions
	userOptions.AllowShadows = true

	// Load the INI file
	var iniFile *ini.File
	var userIniFile *ini.File
	var defaultOnlyIni *ini.File

	if err := LoadFile(&def, []string{def, "", "data/"}, func(filename string) error {
		def = filename

		// Inline-append any external/mods/**/system.def files before parsing.
		modSystemDefs, err := findExternalModSystemDefs(filepath.FromSlash("external/mods"))
		if err != nil {
			return fmt.Errorf("Failed to discover external mod system.def files: %w", err)
		}

		inputBytes, err := LoadText(filename)
		if err != nil {
			return fmt.Errorf("Failed to load text from %s: %w", filename, err)
		}

		var modsConcat strings.Builder
		for _, p := range modSystemDefs {
			txt, err := LoadText(p)
			if err != nil {
				return fmt.Errorf("Failed to load mod system.def from %s: %w", p, err)
			}
			nt := NormalizeNewlines(txt)
			modsConcat.WriteString(nt)
			if !strings.HasSuffix(nt, "\n") {
				modsConcat.WriteString("\n")
			}
		}

		// Preprocess and load INI sources from memory.
		// Motif file overrides mods
		baseTxt := NormalizeNewlines(string(inputBytes))
		modsTxt := modsConcat.String()
		var combined string
		if modsTxt != "" {
			if baseTxt != "" && !strings.HasSuffix(baseTxt, "\n") {
				baseTxt += "\n"
			}
			combined = baseTxt + modsTxt
		} else {
			combined = baseTxt
		}
		normalizedInput := []byte(preprocessINIContent(combined))
		normalizedDefault := []byte(preprocessINIContent(NormalizeNewlines(string(defaultMotif))))

		// Defaults-only baseline
		defaultOnlyIni, err = ini.LoadSources(baseOptions, normalizedDefault)
		if err != nil {
			return fmt.Errorf("Failed to load defaults-only INI from memory: %w", err)
		}

		// merged starts as defaults, then overlay user (first-wins for duplicates in user)
		iniFile, err = ini.LoadSources(baseOptions, normalizedDefault)
		if err != nil {
			return fmt.Errorf("Failed to load merged baseline INI from memory: %w", err)
		}

		userIniFile, err = ini.LoadSources(userOptions, normalizedInput)
		if err != nil {
			return fmt.Errorf("Failed to load user INI source from memory: %w", err)
		}
		// Seed defaults for all derived [<gamemode> Pause Menu] sections before merging user values.
		seedPauseMenuDefaults := func(user, defaults, merged *ini.File) {
			if defaults == nil || merged == nil {
				return
			}
			baseDefaults, err := defaults.GetSection("Pause Menu")
			if err != nil || baseDefaults == nil {
				return
			}
			isDerivedPauseMenu := func(name string) (string, bool) {
				if name == ini.DEFAULT_SECTION {
					return "", false
				}
				lang, base, _ := splitLangPrefix(name)
				base = strings.TrimSpace(base)
				lower := strings.ToLower(base)
				if !strings.HasSuffix(lower, " pause menu") {
					return "", false
				}
				// [Pause Menu] is the root; every other [* Pause Menu] should be derived from it.
				if lower == "pause menu" {
					return "", false
				}
				if lang != "" {
					return lang + "." + base, true
				}
				return base, true
			}
			seedSection := func(dst *ini.File, secName string) {
				if dst == nil {
					return
				}
				sec, err := dst.GetSection(secName)
				if err != nil || sec == nil {
					sec, err = dst.NewSection(secName)
					if err != nil || sec == nil {
						return
					}
				}
				// Top-up missing keys from [Pause Menu] (do not override existing values).
				for _, k := range baseDefaults.Keys() {
					if sec.HasKey(k.Name()) {
						continue
					}
					_, _ = sec.NewKey(k.Name(), k.Value())
				}
			}

			// Seed derived pause menus found in either defaults or user input.
			seen := map[string]bool{}
			for _, f := range []*ini.File{defaults, user} {
				if f == nil {
					continue
				}
				for _, s := range f.Sections() {
					secName, ok := isDerivedPauseMenu(s.Name())
					if !ok {
						continue
					}
					l := strings.ToLower(secName)
					if seen[l] {
						continue
					}
					seen[l] = true
					seedSection(defaults, secName)
					seedSection(merged, secName)
				}
			}
		}
		seedPauseMenuDefaults(userIniFile, defaultOnlyIni, iniFile)
		overlayUserFirstWins(iniFile, userIniFile)

		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading definition file: %v\n", err)
		return nil, err
	}

	var m Motif
	m.Def = def
	m.initStruct()

	assignFrom := func(src *ini.File) {
		if src == nil {
			return
		}
		type secPair struct {
			sec  *ini.Section
			name string // logical (language-stripped) name
		}
		var baseSecs, langSecs []secPair
		curLang := SelectedLanguage()

		for _, s := range src.Sections() {
			raw := s.Name()
			if raw == ini.DEFAULT_SECTION {
				continue
			}
			lang, base, has := splitLangPrefix(raw)
			logical := raw
			if has {
				logical = base
			}
			// Backgrounds and [Begin Action] blocks are skipped (case-insensitive).
			lb := strings.ToLower(logical)
			// Skip BG element sections like [XResultsBg] (but still allow *BgDef sections).
			if strings.Contains(lb, "resultsbg") {
				if !strings.HasSuffix(lb, "bgdef") {
					goto nextSection
				}
			}
			// Skip raw BG sections which are handled by loadBGDef.
			if strings.HasPrefix(lb, "begin ") {
				goto nextSection
			}
			for _, p := range []string{
				"titlebg", "selectbg", "versusbg", "continuebg",
				"victorybg", "winbg",
				"optionbg", "replaybg", "attractbg",
				"challengerbg", "hiscorebg",
			} {
				if strings.HasPrefix(lb, p) {
					// Allow BgDef sections that should be mapped into BgDefProperties.
					if strings.HasSuffix(lb, "bgdef") {
						break
					}
					goto nextSection
				}
			}
			// "music" is handled separately later.
			if strings.EqualFold(logical, "music") {
				goto nextSection
			}
			// Route by language.
			if has {
				if lang == curLang {
					langSecs = append(langSecs, secPair{s, logical})
				}
			} else {
				baseSecs = append(baseSecs, secPair{s, logical})
			}
		nextSection:
		}

		process := func(pair secPair) {
			section := pair.sec
			sectionName := pair.name
			for _, key := range section.Keys() {
				keyName := key.Name()
				if strings.HasPrefix(keyName, "menu.itemname.") {
					// handled in script.go to preserve order
					continue
				}
				value, dup := iniFirstValue(key)
				if dup > 0 {
					sys.errLog.Printf("Duplicate key [%s] %s (%d duplicate(s) ignored)", sectionName, keyName, dup)
				}

				// Normalize spaces
				secNorm := strings.ReplaceAll(sectionName, " ", "_")
				keyNorm := strings.ReplaceAll(keyName, " ", "_")

				var keyParts []queryPart
				// Literal sections keep dots in keys
				if isLiteralSectionFor(&m, sectionName) {
					keyParts = []queryPart{
						{name: strings.ToLower(secNorm)},
						{name: keyNorm},
					}
				} else {
					fullKey := strings.ToLower(secNorm) + "." + keyNorm
					keyParts = parseQueryPath(fullKey)
				}

				if err := assignField(&m, keyParts, value, def); err != nil {
					fmt.Printf("Warning: Failed to assign key [%s.%s]: %v\n", sectionName, keyName, err)
				}
			}
		}
		for _, sp := range baseSecs {
			process(sp)
		}
		for _, sp := range langSecs {
			process(sp)
		}
	}

	// Apply precedence: struct defaults < defaultMotif.ini < user motif
	assignFrom(defaultOnlyIni)
	assignFrom(userIniFile)
	sys.keepAlive()

	// Localcoord is used during loading (before the final sys.motif assignment)
	sys.motif.Info.Localcoord = m.Info.Localcoord

	m.IniFile = iniFile
	m.UserIniFile = userIniFile
	m.DefaultOnlyIni = defaultOnlyIni
	if userIniFile != nil {
		applyCustomDefaults(&m, userIniFile)
	}

	// Resolve inline fonts early, so TitleInfo.Loading can get a real font index.
	reserveUserFontSlots(&m)
	resolveInlineFonts(m.IniFile, m.Def, m.Fnt, m.fntIndexByKey, m.SetValueUpdate)
	syncFontsMap(&m.Files.Font, m.Fnt, m.fntIndexByKey)
	sys.keepAlive()

	// Build the loading TextSprite and draw it before we proceed to heavier asset loads (SFF, BG, etc).
	m.drawLoading()
	sys.keepAlive()

	// Proceed with the regular heavyweight loads.
	m.loadFiles()
	sys.keepAlive()

	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	lines, i := SplitAndTrim(str, "\n"), 0
	m.AnimTable = ReadAnimationTable(m.Sff, &m.Sff.palList, lines, &i)
	i = 0

	m.overrideParams()
	m.fixLocalcoordOverrides()
	m.applyGlyphDefaultsFromMovelist()
	m.populateDataPointers()
	m.applyPostParsePosAdjustments()

	m.Music = parseMusicSection(pickLangSectionMerged(iniFile, "Music"))
	m.Music.DebugDump("Motif [Music]")

	return &m, nil
}

// Propagates movelist.glyphs defaults
func (m *Motif) applyGlyphDefaultsFromMovelist() {
	if m == nil {
		return
	}
	if len(m.Glyphs) == 0 {
		return
	}

	// Use [Pause Menu] movelist glyph defaults as the global baseline.
	var mg MovelistGlyphsProperties
	if m.PauseMenu != nil {
		if pm := m.PauseMenu["pause_menu"]; pm != nil {
			mg = pm.Movelist.Glyphs
		} else {
			// Best-effort fallback: first available pause menu entry.
			for _, pm := range m.PauseMenu {
				if pm != nil {
					mg = pm.Movelist.Glyphs
					break
				}
			}
		}
	} else {
		return
	}
	for _, g := range m.Glyphs {
		if g == nil {
			continue
		}
		g.Offset = mg.Offset
		g.Scale = mg.Scale
		g.Layerno = mg.Layerno
		g.Localcoord = mg.Localcoord
	}
}

// If an element is customized in userIniFile but its localcoord is missing/empty,
// use Info.Localcoord instead of the defaultMotif value; otherwise keep the default.
func (m *Motif) fixLocalcoordOverrides() {
	if m == nil || m.UserIniFile == nil || m.DefaultOnlyIni == nil || m.IniFile == nil {
		return
	}

	for _, mergedSec := range m.IniFile.Sections() {
		secName := mergedSec.Name()
		if secName == ini.DEFAULT_SECTION {
			continue
		}

		userSec, _ := m.UserIniFile.GetSection(secName)
		defSec, _ := m.DefaultOnlyIni.GetSection(secName)

		var userKeys []*ini.Key
		if userSec != nil {
			userKeys = userSec.Keys()
		}

		for _, key := range mergedSec.Keys() {
			keyName := key.Name()
			lowerKey := strings.ToLower(keyName)

			if !strings.HasSuffix(lowerKey, ".localcoord") {
				continue
			}

			lastDot := strings.LastIndex(keyName, ".")
			if lastDot < 0 {
				continue
			}
			prefix := keyName[:lastDot+1]
			lowerPrefix := strings.ToLower(prefix)

			// 1) If user explicitly sets this .localcoord (non-empty) in their motif, always respect it.
			if userSec != nil {
				if uk, err := userSec.GetKey(keyName); err == nil && uk != nil {
					uv, _ := iniFirstValue(uk)
					if strings.TrimSpace(uv) != "" {
						continue
					}
				}
			}

			// 2) Is this exact .localcoord key present in the embedded defaults?
			inDefaults := false
			if defSec != nil {
				if dk, err := defSec.GetKey(keyName); err == nil && dk != nil {
					inDefaults = true
				}
			}

			// 3) Did the user (or inheritance that wrote back into UserIniFile) touch this element?
			// Check if user customized this element (any other key with same prefix).
			elementTouched := false
			for _, uk := range userKeys {
				n := uk.Name()
				ln := strings.ToLower(n)
				if strings.HasPrefix(ln, lowerPrefix) &&
					ln != lowerKey &&
					!strings.HasSuffix(ln, ".key") &&
					!strings.HasSuffix(ln, ".snd") &&
					strings.TrimSpace(uk.Value()) != "" {
					elementTouched = true
					break
				}
			}

			// 4) If missing from defaults, any same-prefix non-localcoord key in merged INI marks override.
			if !inDefaults && !elementTouched {
				for _, mk := range mergedSec.Keys() {
					n := mk.Name()
					ln := strings.ToLower(n)
					if strings.HasPrefix(ln, lowerPrefix) &&
						ln != lowerKey &&
						strings.TrimSpace(mk.Value()) != "" {
						elementTouched = true
						break
					}
				}
			}
			if !elementTouched {
				// Pure untouched default (or truly isolated entry): keep its default localcoord.
				continue
			}

			// Let PopulateDataPointers re-fill this from motif localcoord:
			// set struct (and ini) value to 0,0 so it is treated as "unspecified".
			secNorm := strings.ReplaceAll(secName, " ", "_")
			keyNorm := strings.ReplaceAll(keyName, " ", "_")
			query := strings.ToLower(secNorm + "." + keyNorm)

			if err := m.SetValueUpdate(query, "0, 0"); err != nil {
				fmt.Printf("Warning: failed to reset localcoord for %s: %v\n", query, err)
			}
		}
	}
}

// applyMenuUseLocalcoord forces menu-related localcoord keys to inherit motif localcoord
// when menu.uselocalcoord is enabled in a section.

// InheritSpec describes how to inherit keys from one prefix to another inside the given INI sections.
// Example: srcSec="Option Info", srcPrefix="menu.", dstSec="Option Info", dstPrefix="keymenu.menu."
type InheritSpec struct {
	SrcSec    string
	SrcPrefix string
	DstSec    string
	DstPrefix string
}

// mergeWithInheritance applies "user overrides default" with intra-file inheritance.
func (m *Motif) mergeWithInheritance(specs []InheritSpec) {
	if m == nil || m.IniFile == nil {
		return
	}
	user := m.UserIniFile
	defs := m.DefaultOnlyIni
	merged := m.IniFile

	isPreviewAnim := func(dstKey string) bool {
		l := strings.ToLower(dstKey)
		return strings.HasSuffix(l, ".palmenu.preview.anim")
	}

	get := func(f *ini.File, sec, key string) (string, bool) {
		if f == nil {
			return "", false
		}
		s, err := f.GetSection(sec)
		if err != nil || s == nil {
			return "", false
		}
		if !s.HasKey(key) {
			return "", false
		}
		k, _ := s.GetKey(key)
		if k == nil {
			return "", false
		}
		v, _ := iniFirstValue(k)
		return v, true
	}

	shouldSkip := func(fullKeyLower string) bool {
		// Avoid inheriting dotted item/valuename lists which are consumed elsewhere.
		return strings.Contains(fullKeyLower, ".itemname.") || strings.Contains(fullKeyLower, ".valuename.")
	}

	// Ensure a section exists in the user ini when we need to mirror
	// a user-originated inherited key into it.
	ensureUserSection := func(name string) *ini.Section {
		if user == nil {
			return nil
		}
		if s, err := user.GetSection(name); err == nil && s != nil {
			return s
		}
		s, err := user.NewSection(name)
		if err != nil {
			fmt.Printf("Warning: failed to create section %s in user ini: %v\n", name, err)
			return nil
		}
		return s
	}

	type valueSource int
	const (
		srcNone valueSource = iota
		srcUserDst
		srcUserSrc
		srcDefDst
		srcDefSrc
	)

	for _, sp := range specs {
		// Build a set of suffixes present under either src/dst in user or default.
		suffixes := map[string]struct{}{}
		collect := func(f *ini.File, sec, prefix string) {
			if f == nil {
				return
			}
			s, err := f.GetSection(sec)
			if err != nil || s == nil {
				return
			}
			lp := strings.ToLower(prefix)
			for _, k := range s.Keys() {
				kn := k.Name()
				lkn := strings.ToLower(kn)
				if strings.HasPrefix(lkn, lp) {
					if shouldSkip(lkn) {
						continue
					}
					// same-length lowercasing keeps indices aligned (ASCII)
					suf := kn[len(prefix):]
					suffixes[suf] = struct{}{}
				}
			}
		}
		collect(user, sp.SrcSec, sp.SrcPrefix)
		collect(user, sp.DstSec, sp.DstPrefix)
		collect(defs, sp.SrcSec, sp.SrcPrefix)
		collect(defs, sp.DstSec, sp.DstPrefix)

		// Resolve and set final values following the precedence.
		for suf := range suffixes {
			dstKey := sp.DstPrefix + suf
			srcKey := sp.SrcPrefix + suf
			lowerFull := strings.ToLower(dstKey)
			if shouldSkip(lowerFull) {
				continue
			}

			// If the user didn't set dst/src, and the merged INI already has dst,
			// skip touching it to avoid clobbering a resolved (e.g. font-index) value.
			if _, uDst := get(user, sp.DstSec, dstKey); !uDst {
				if _, uSrc := get(user, sp.SrcSec, srcKey); !uSrc {
					if _, haveMerged := get(merged, sp.DstSec, dstKey); haveMerged {
						// Nothing to inherit here; keep the existing (possibly resolved) value.
						continue
					}
				}
			}

			// palmenu.preview only inherits when palmenu.preview.anim is defined
			if isPreviewAnim(dstKey) {
				if v, ok := get(user, sp.DstSec, dstKey); ok {
					tv := strings.TrimSpace(v)
					if !IsInt(tv) || Atoi(tv) < 0 {
						continue
					}
				} else {
					continue
				}
			}

			var (
				val string
				src valueSource
			)

			// Priority: user dst > user src > default dst > default src
			if v, ok := get(user, sp.DstSec, dstKey); ok {
				val, src = v, srcUserDst
			} else if v, ok := get(user, sp.SrcSec, srcKey); ok {
				val, src = v, srcUserSrc
			} else if v, ok := get(defs, sp.DstSec, dstKey); ok {
				val, src = v, srcDefDst
			} else if v, ok := get(defs, sp.SrcSec, srcKey); ok {
				val, src = v, srcDefSrc
			} else {
				continue
			}

			// Write into the merged ini & struct.
			secPath := strings.ReplaceAll(sp.DstSec, " ", "_")
			query := strings.ToLower(secPath + "." + dstKey)
			if err := m.SetValueUpdate(query, val); err != nil {
				//fmt.Printf("Warning: inheritance set failed for %s = %q: %v\n", query, val, err)
			}

			// If a value comes from the user INI (directly or via src), copy it into m.UserIniFile
			// so fixLocalcoordOverrides treats the element as user-touched, including inherited keys.
			if user != nil {
				switch src {
				case srcUserDst:
					// Already present in user INI at the destination; nothing to do.
				case srcUserSrc:
					if sec := ensureUserSection(sp.DstSec); sec != nil && !sec.HasKey(dstKey) {
						if _, err := sec.NewKey(dstKey, val); err != nil {
							//fmt.Printf("Warning: failed to write inherited key %s.%s into user ini: %v\n", sp.DstSec, dstKey, err)
						}
					}
				default:
					// Default-sourced inheritance: do NOT mirror into userIniFile.
				}
			}
		}
	}
}

func (m *Motif) overrideParams() {
	// Define inheritance rules (section/prefix based).
	specs := []InheritSpec{
		// [Option Info]
		{SrcSec: "Option Info", SrcPrefix: "menu.", DstSec: "Option Info", DstPrefix: "keymenu."},
		// [Select Info]
		{SrcSec: "Select Info", SrcPrefix: "p1.face.", DstSec: "Select Info", DstPrefix: "p1.face.done."},
		{SrcSec: "Select Info", SrcPrefix: "p1.face2.", DstSec: "Select Info", DstPrefix: "p1.face2.done."},
		{SrcSec: "Select Info", SrcPrefix: "p2.face.", DstSec: "Select Info", DstPrefix: "p2.face.done."},
		{SrcSec: "Select Info", SrcPrefix: "p2.face2.", DstSec: "Select Info", DstPrefix: "p2.face2.done."},
		{SrcSec: "Select Info", SrcPrefix: "p1.face.", DstSec: "Select Info", DstPrefix: "p1.palmenu.preview."},
		{SrcSec: "Select Info", SrcPrefix: "p2.face.", DstSec: "Select Info", DstPrefix: "p2.palmenu.preview."},
		{SrcSec: "Select Info", SrcPrefix: "p1.teammenu.item.", DstSec: "Select Info", DstPrefix: "p1.teammenu.item.active."},
		{SrcSec: "Select Info", SrcPrefix: "p1.teammenu.item.", DstSec: "Select Info", DstPrefix: "p1.teammenu.item.active2."},
		{SrcSec: "Select Info", SrcPrefix: "p2.teammenu.item.", DstSec: "Select Info", DstPrefix: "p2.teammenu.item.active."},
		{SrcSec: "Select Info", SrcPrefix: "p2.teammenu.item.", DstSec: "Select Info", DstPrefix: "p2.teammenu.item.active2."},
		{SrcSec: "Select Info", SrcPrefix: "p1.", DstSec: "Select Info", DstPrefix: "p3."},
		{SrcSec: "Select Info", SrcPrefix: "p1.", DstSec: "Select Info", DstPrefix: "p5."},
		{SrcSec: "Select Info", SrcPrefix: "p1.", DstSec: "Select Info", DstPrefix: "p7."},
		{SrcSec: "Select Info", SrcPrefix: "p2.", DstSec: "Select Info", DstPrefix: "p4."},
		{SrcSec: "Select Info", SrcPrefix: "p2.", DstSec: "Select Info", DstPrefix: "p6."},
		{SrcSec: "Select Info", SrcPrefix: "p2.", DstSec: "Select Info", DstPrefix: "p8."},
		{SrcSec: "Select Info", SrcPrefix: "stage.", DstSec: "Select Info", DstPrefix: "stage.active."},
		{SrcSec: "Select Info", SrcPrefix: "stage.", DstSec: "Select Info", DstPrefix: "stage.active2."},
		{SrcSec: "Select Info", SrcPrefix: "stage.", DstSec: "Select Info", DstPrefix: "stage.done."},
		// [VS Screen]
		{SrcSec: "VS Screen", SrcPrefix: "p1.", DstSec: "VS Screen", DstPrefix: "p1.done."},
		{SrcSec: "VS Screen", SrcPrefix: "p1.face2.", DstSec: "VS Screen", DstPrefix: "p1.face2.done."},
		{SrcSec: "VS Screen", SrcPrefix: "p2.", DstSec: "VS Screen", DstPrefix: "p2.done."},
		{SrcSec: "VS Screen", SrcPrefix: "p2.face2.", DstSec: "VS Screen", DstPrefix: "p2.face2.done."},
		{SrcSec: "VS Screen", SrcPrefix: "p1.", DstSec: "VS Screen", DstPrefix: "p3."},
		{SrcSec: "VS Screen", SrcPrefix: "p1.", DstSec: "VS Screen", DstPrefix: "p5."},
		{SrcSec: "VS Screen", SrcPrefix: "p1.", DstSec: "VS Screen", DstPrefix: "p7."},
		{SrcSec: "VS Screen", SrcPrefix: "p2.", DstSec: "VS Screen", DstPrefix: "p4."},
		{SrcSec: "VS Screen", SrcPrefix: "p2.", DstSec: "VS Screen", DstPrefix: "p6."},
		{SrcSec: "VS Screen", SrcPrefix: "p2.", DstSec: "VS Screen", DstPrefix: "p8."},
		// [Victory Screen]
		{SrcSec: "Victory Screen", SrcPrefix: "p1.", DstSec: "Victory Screen", DstPrefix: "p3."},
		{SrcSec: "Victory Screen", SrcPrefix: "p1.", DstSec: "Victory Screen", DstPrefix: "p5."},
		{SrcSec: "Victory Screen", SrcPrefix: "p1.", DstSec: "Victory Screen", DstPrefix: "p7."},
		{SrcSec: "Victory Screen", SrcPrefix: "p2.", DstSec: "Victory Screen", DstPrefix: "p4."},
		{SrcSec: "Victory Screen", SrcPrefix: "p2.", DstSec: "Victory Screen", DstPrefix: "p6."},
		{SrcSec: "Victory Screen", SrcPrefix: "p2.", DstSec: "Victory Screen", DstPrefix: "p8."},
		// [Dialogue Info]
		{SrcSec: "Dialogue Info", SrcPrefix: "p1.face.", DstSec: "Dialogue Info", DstPrefix: "p1.face.active."},
		{SrcSec: "Dialogue Info", SrcPrefix: "p2.face.", DstSec: "Dialogue Info", DstPrefix: "p2.face.active."},
		// [Hiscore Info]
		{SrcSec: "Hiscore Info", SrcPrefix: "item.rank.", DstSec: "Hiscore Info", DstPrefix: "item.rank.active."},
		{SrcSec: "Hiscore Info", SrcPrefix: "item.rank.", DstSec: "Hiscore Info", DstPrefix: "item.rank.active2."},
		{SrcSec: "Hiscore Info", SrcPrefix: "item.result.", DstSec: "Hiscore Info", DstPrefix: "item.result.active."},
		{SrcSec: "Hiscore Info", SrcPrefix: "item.result.", DstSec: "Hiscore Info", DstPrefix: "item.result.active2."},
		{SrcSec: "Hiscore Info", SrcPrefix: "item.name.", DstSec: "Hiscore Info", DstPrefix: "item.name.active."},
		{SrcSec: "Hiscore Info", SrcPrefix: "item.name.", DstSec: "Hiscore Info", DstPrefix: "item.name.active2."},
	}
	// Apply [Pause Menu] inheritance to every other [* Pause Menu]
	for _, sec := range m.customPauseMenuSections() {
		specs = append(specs, InheritSpec{SrcSec: "Pause Menu", SrcPrefix: "", DstSec: sec, DstPrefix: ""})
	}
	m.mergeWithInheritance(specs)
	// Inheritance may add new font filenames; rerun resolver to map them to deduped font indices.
	//resolveInlineFonts(m.IniFile, m.Def, m.Fnt, m.fntIndexByKey, m.SetValueUpdate)
}

func (m *Motif) customPauseMenuSections() []string {
	if m == nil || m.IniFile == nil {
		return nil
	}
	// Sections ending with " Pause Menu" (case-insensitive) are considered pause menus
	seen := map[string]bool{}
	var out []string
	for _, s := range m.IniFile.Sections() {
		raw := s.Name()
		if raw == ini.DEFAULT_SECTION {
			continue
		}
		_, base, _ := splitLangPrefix(raw)
		base = strings.TrimSpace(base)
		if base == "" {
			continue
		}
		lower := strings.ToLower(base)
		if !strings.HasSuffix(lower, " pause menu") {
			continue
		}
		if lower == "pause menu" {
			continue
		}
		if seen[lower] {
			continue
		}
		seen[lower] = true
		out = append(out, base)
	}
	return out
}

// Initialize struct
func (m *Motif) initStruct() {
	initMaps(reflect.ValueOf(m).Elem())
	applyDefaultsToValue(reflect.ValueOf(m).Elem())
	m.fadeIn = newFade()
	m.fadeOut = newFade()
	m.fadePolicy = FadeContinue
	m.fntIndexByKey = make(map[string]int)
}

func (m *Motif) loadBgDefProperties(bgDef *BgDefProperties, bgname, spr string) {
	if bgDef.Spr == "" || bgDef.Spr == spr || bgDef.Spr == m.Files.Spr {
		bgDef.Sff = m.Sff
	} else {
		LoadFile(&bgDef.Spr, []string{bgDef.Spr, m.Def, "", "data/"}, func(filename string) error {
			if filename != "" {
				var err error
				bgDef.Sff, err = loadSff(filename, false, true, false)
				if err != nil {
					sys.errLog.Printf("Failed to load %v: %v", filename, err)
				}
			}
			if bgDef.Sff == nil {
				bgDef.Sff = m.Sff
			}
			return nil
		})
	}
	if bgname != "" {
		var err error
		bgDef.BGDef, err = loadBGDef(bgDef.Sff, m.Model, m.Def, bgname, bgDef.StartLayer)
		if err != nil {
			sys.errLog.Printf("Failed to load %v (%v): %v\n", bgname, m.Def, err.Error())
		}
	}
	if bgDef.BGDef == nil {
		bgDef.BGDef = newBGDef(m.Def)
	}
	sys.keepAlive()
}

func (m *Motif) loadFiles() {
	LoadFile(&m.Files.Spr, []string{m.Files.Spr}, func(filename string) error {
		if filename != "" {
			var err error
			m.Sff, err = loadSff(filename, false, true, false)
			if err != nil {
				sys.errLog.Printf("Failed to load %v: %v", filename, err)
			}
		}
		if m.Sff == nil {
			m.Sff = newSff()
		}
		return nil
	})
	sys.keepAlive()

	LoadFile(&m.Files.Glyphs, []string{m.Files.Glyphs}, func(filename string) error {
		if filename != "" {
			var err error
			m.GlyphsSff, err = loadSff(filename, false, true, false)
			if err != nil {
				sys.errLog.Printf("Failed to load %v: %v", filename, err)
			}
		}
		if m.GlyphsSff == nil {
			m.GlyphsSff = newSff()
		}
		return nil
	})
	sys.keepAlive()

	LoadFile(&m.Files.Model, []string{m.Files.Model}, func(filename string) error {
		if filename != "" {
			var err error
			m.Model, err = loadglTFModel(filename)
			if err != nil {
				sys.errLog.Printf("Failed to load %v: %v", filename, err)
			}
			sys.mainThreadTask <- func() {
				gfx.SetModelVertexData(1, m.Model.vertexBuffer)
				gfx.SetModelIndexData(1, m.Model.elementBuffer...)
			}
			sys.runMainThreadTask()
			sys.keepAlive()
		}
		return nil
	})

	m.loadBgDefProperties(&m.TitleBgDef, "titlebg", m.Files.Spr)
	m.loadBgDefProperties(&m.SelectBgDef, "selectbg", m.Files.Spr)
	m.loadBgDefProperties(&m.VersusBgDef, "versusbg", m.Files.Spr)
	m.loadBgDefProperties(&m.ContinueBgDef, "continuebg", m.Files.Spr)
	m.loadBgDefProperties(&m.VictoryBgDef, "victorybg", m.Files.Spr)
	m.loadBgDefProperties(&m.WinBgDef, "winbg", m.Files.Spr)

	// Load all ResultsBgDef entries (any mode that defines <mode>ResultsBgDef).
	for secKey, bg := range m.ResultsBgDef {
		if bg == nil {
			continue
		}
		key := strings.ToLower(secKey)
		// "survivalresultsbgdef" -> "survivalresultsbg"
		bgName := strings.TrimSuffix(key, "def")
		m.loadBgDefProperties(bg, bgName, m.Files.Spr)
	}

	m.loadBgDefProperties(&m.OptionBgDef, "optionbg", m.Files.Spr)
	if _, err := m.UserIniFile.GetSection("ReplayBgDef"); err == nil {
		m.loadBgDefProperties(&m.ReplayBgDef, "replaybg", m.Files.Spr)
	} else {
		m.ReplayBgDef = m.TitleBgDef
	}
	// Load all PauseBgDef entries (any mode that defines <mode>PauseBgDef).
	for secKey, bg := range m.PauseBgDef {
		if bg == nil {
			continue
		}
		key := strings.ToLower(secKey)
		bgName := strings.TrimSuffix(key, "def") // e.g. "pausebgdef" -> "pausebg"
		m.loadBgDefProperties(bg, bgName, m.Files.Spr)
	}
	if _, err := m.UserIniFile.GetSection("AttractBgDef"); err == nil {
		m.loadBgDefProperties(&m.AttractBgDef, "attractbg", m.Files.Spr)
	} else {
		m.AttractBgDef = m.TitleBgDef
	}
	m.loadBgDefProperties(&m.ChallengerBgDef, "challengerbg", m.Files.Spr)
	if _, err := m.UserIniFile.GetSection("HiscoreBgDef"); err == nil {
		m.loadBgDefProperties(&m.HiscoreBgDef, "hiscorebg", m.Files.Spr)
	} else {
		m.HiscoreBgDef = m.TitleBgDef
	}

	LoadFile(&m.Files.Snd, []string{m.Files.Snd}, func(filename string) error {
		if filename != "" {
			var err error
			m.Snd, err = LoadSnd(filename)
			if err != nil {
				sys.errLog.Printf("Failed to load %v: %v", filename, err)
			}
		}
		if m.Snd == nil {
			m.Snd = newSnd()
		}
		return nil
	})
	sys.keepAlive()

	for key, fnt := range m.Files.Font {
		LoadFile(&fnt.Font, []string{fnt.Font}, func(filename string) error {
			re := regexp.MustCompile(`\d+`)
			i := int(Atoi(re.FindString(key)))

			if filename != "" {
				var err error
				m.Fnt[i], err = loadFnt(filename, fnt.Height)
				registerFontIndex(m.fntIndexByKey, filename, fnt.Height, i)
				if err != nil {
					sys.errLog.Printf("Failed to load %v: %v", filename, err)
				}
			}
			if m.Fnt[i] == nil {
				m.Fnt[i] = newFnt()
			}
			// Populate extended properties from the loaded font
			if m.Fnt[i] != nil {
				fnt.Type = m.Fnt[i].Type
				fnt.Size = m.Fnt[i].Size
				fnt.Spacing = m.Fnt[i].Spacing
				fnt.Offset = m.Fnt[i].offset
			}
			return nil
		})
		sys.keepAlive()
	}
}

func (m *Motif) populateDataPointers() {
	PopulateDataPointers(m, m.Info.Localcoord)
}

func (m *Motif) resolvePath() {
	v := sys.cfg.Config.Motif
	if x, ok := sys.cmdFlags["-r"]; ok {
		v = x
	} else if x, ok := sys.cmdFlags["-rubric"]; ok {
		v = x
	}

	v = filepath.ToSlash(v)
	lower := strings.ToLower(v)

	if strings.HasPrefix(lower, "data/") {
		if FileExist(v) != "" {
			m.Def = v
			return
		}
	}

	if strings.HasSuffix(lower, ".def") {
		if cand := SearchFile(v, []string{"data/"}); FileExist(cand) != "" {
			m.Def = filepath.ToSlash(cand)
			return
		}
	}

	dir := filepath.ToSlash(filepath.Join("data", v)) + "/"
	if cand := SearchFile("system.def", []string{dir}); FileExist(cand) != "" {
		m.Def = filepath.ToSlash(cand)
		return
	}

	m.Def = v
}

func (m *Motif) applyPostParsePosAdjustments() {
	animSetPos := func(a *Anim, dx, dy float32) {
		a.SetPos(a.offsetInit[0]+dx, a.offsetInit[1]+dy)
	}
	textSetPos := func(ts *TextSprite, dx, dy float32) {
		ts.SetPos(ts.offsetInit[0]+dx, ts.offsetInit[1]+dy)
	}
	offsetAnims := func(dx, dy float32, anims ...*Anim) {
		for _, a := range anims {
			animSetPos(a, dx, dy)
		}
	}
	offsetTexts := func(dx, dy float32, texts ...*TextSprite) {
		for _, ts := range texts {
			textSetPos(ts, dx, dy)
		}
	}
	shiftMenu := func(me *MenuProperties) {
		dx, dy := me.Pos[0], me.Pos[1]
		// Arrows
		offsetAnims(dx, dy, me.Arrow.Up.AnimData, me.Arrow.Down.AnimData)
		// Common item texts
		offsetTexts(dx, dy,
			me.Item.TextSpriteData,
			me.Item.Selected.TextSpriteData,
			me.Item.Selected.Active.TextSpriteData,
			me.Item.Active.TextSpriteData,
			me.Item.Value.TextSpriteData,
			me.Item.Value.Active.TextSpriteData,
			// Extra fields some menus use:
			me.Item.Value.Conflict.TextSpriteData,
			me.Item.Info.TextSpriteData,
			me.Item.Info.Active.TextSpriteData,
		)
		// Backgrounds
		for _, ap := range me.Item.Bg {
			animSetPos(ap.AnimData, dx, dy)
		}
		for _, ap := range me.Item.Active.Bg {
			animSetPos(ap.AnimData, dx, dy)
		}
	}
	shiftMovelist := func(mv *MenuInfoProperties) {
		if mv == nil {
			return
		}
		ml := &mv.Movelist
		offsetAnims(ml.Pos[0], ml.Pos[1], ml.Arrow.Up.AnimData, ml.Arrow.Down.AnimData)
		offsetTexts(ml.Pos[0], ml.Pos[1], ml.Title.TextSpriteData, ml.Text.TextSpriteData)
	}
	adjustSelect := func(ps *PlayerSelectProperties) {
		tm := &ps.TeamMenu
		// TeamMenu backgrounds
		for _, ap := range tm.Bg {
			animSetPos(ap.AnimData, tm.Pos[0], tm.Pos[1])
		}
		for _, ap := range tm.Active.Bg {
			animSetPos(ap.AnimData, tm.Pos[0], tm.Pos[1])
		}
		// Titles & base text
		offsetAnims(tm.Pos[0], tm.Pos[1], tm.SelfTitle.AnimData, tm.EnemyTitle.AnimData)
		offsetTexts(tm.Pos[0], tm.Pos[1], tm.SelfTitle.TextSpriteData, tm.EnemyTitle.TextSpriteData)
		offsetTexts(tm.Pos[0], tm.Pos[1], tm.Item.TextSpriteData, tm.Item.Active.TextSpriteData, tm.Item.Active2.TextSpriteData)

		// Icons at (Pos + Item.Offset)
		offX := tm.Pos[0] + tm.Item.Offset[0]
		offY := tm.Pos[1] + tm.Item.Offset[1]
		offsetAnims(offX, offY,
			tm.Item.Cursor.AnimData,
			tm.Value.Icon.AnimData,
			tm.Value.Empty.Icon.AnimData,
			tm.Ratio1.Icon.AnimData,
			tm.Ratio2.Icon.AnimData,
			tm.Ratio3.Icon.AnimData,
			tm.Ratio4.Icon.AnimData,
			tm.Ratio5.Icon.AnimData,
			tm.Ratio6.Icon.AnimData,
			tm.Ratio7.Icon.AnimData,
		)
		// Palette menu
		pm := &ps.PalMenu
		animSetPos(pm.Bg.AnimData, pm.Pos[0], pm.Pos[1])
		offsetTexts(pm.Pos[0], pm.Pos[1], pm.Number.TextSpriteData, pm.Text.TextSpriteData)

		// Face.Random and Face2.Random
		offsetAnims(ps.Face.Pos[0], ps.Face.Pos[1], ps.Face.Random.AnimData)
		offsetAnims(ps.Face2.Pos[0], ps.Face2.Pos[1], ps.Face2.Random.AnimData)
	}

	// Select Screen: Players
	for _, ps := range []*PlayerSelectProperties{
		&m.SelectInfo.P1, &m.SelectInfo.P2, &m.SelectInfo.P3, &m.SelectInfo.P4,
		&m.SelectInfo.P5, &m.SelectInfo.P6, &m.SelectInfo.P7, &m.SelectInfo.P8,
	} {
		adjustSelect(ps)
	}

	// Select Screen: Stage
	{
		st := &m.SelectInfo.Stage
		offsetAnims(st.Pos[0], st.Pos[1], st.Portrait.Bg.AnimData, st.Portrait.Random.AnimData)
		offsetTexts(st.Pos[0], st.Pos[1], st.TextSpriteData, st.Active.TextSpriteData, st.Active2.TextSpriteData, st.Done.TextSpriteData)
	}

	// Menus
	for _, me := range []*MenuProperties{
		&m.TitleInfo.Menu,
		&m.OptionInfo.Menu,
		&m.ReplayInfo.Menu,
		&m.AttractMode.Menu,
		&m.OptionInfo.KeyMenu.MenuProperties,
	} {
		shiftMenu(me)
	}
	if m.PauseMenu != nil {
		for _, pm := range m.PauseMenu {
			if pm != nil {
				shiftMenu(&pm.Menu)
			}
		}
	}

	// KeyMenu
	{
		km := &m.OptionInfo.KeyMenu
		textSetPos(km.P1.Playerno.TextSpriteData, km.Pos[0]+km.P1.MenuOffset[0], km.Pos[1]+km.P1.MenuOffset[1])
		textSetPos(km.P2.Playerno.TextSpriteData, km.Pos[0]+km.P2.MenuOffset[0], km.Pos[1]+km.P2.MenuOffset[1])
	}

	// Movelists (all pause menus)
	{
		if m.PauseMenu != nil {
			for _, pm := range m.PauseMenu {
				shiftMovelist(pm)
			}
		}
	}

	// VS Screen: player name positions
	{
		type namePos struct {
			ts  *TextSprite
			pos [2]float32
		}
		names := []namePos{
			{m.VsScreen.P1.Name.TextSpriteData, m.VsScreen.P1.Name.Pos},
			{m.VsScreen.P2.Name.TextSpriteData, m.VsScreen.P2.Name.Pos},
			{m.VsScreen.P3.Name.TextSpriteData, m.VsScreen.P3.Name.Pos},
			{m.VsScreen.P4.Name.TextSpriteData, m.VsScreen.P4.Name.Pos},
			{m.VsScreen.P5.Name.TextSpriteData, m.VsScreen.P5.Name.Pos},
			{m.VsScreen.P6.Name.TextSpriteData, m.VsScreen.P6.Name.Pos},
			{m.VsScreen.P7.Name.TextSpriteData, m.VsScreen.P7.Name.Pos},
			{m.VsScreen.P8.Name.TextSpriteData, m.VsScreen.P8.Name.Pos},
		}
		for _, n := range names {
			textSetPos(n.ts, n.pos[0], n.pos[1])
		}
	}

	// VS Screen: stage
	{
		st := &m.VsScreen.Stage
		offsetAnims(st.Pos[0], st.Pos[1], st.Portrait.Bg.AnimData)
		offsetTexts(st.Pos[0], st.Pos[1], st.TextSpriteData)
	}

	// Hiscore
	{
		hi := &m.HiscoreInfo
		offsetTexts(hi.Pos[0], hi.Pos[1], hi.Item.Rank.TextSpriteData, hi.Item.Result.TextSpriteData, hi.Item.Name.TextSpriteData)
	}
}

func (m *Motif) drawLoading() {
	// Ensure the loading font slot is populated before creating the TextSprite.
	fontIdx := m.TitleInfo.Loading.Font[0]
	if fontIdx >= 0 {
		if m.Fnt == nil {
			m.Fnt = make(map[int]*Fnt)
		}
		if m.Fnt[int(fontIdx)] == nil && m.Files.Font != nil {
			key := fmt.Sprintf("font%d", fontIdx)
			if fp, ok := m.Files.Font[key]; ok && fp != nil && fp.Font != "" {
				f, err := loadFnt(fp.Font, fp.Height)
				if err != nil {
					sys.errLog.Printf("Failed to preload %v for loading screen (%s): %v", fp.Font, key, err)
				}
				if f == nil {
					f = newFnt()
				}
				m.Fnt[int(fontIdx)] = f
				registerFontIndex(m.fntIndexByKey, fp.Font, fp.Height, int(fontIdx))
				fp.Type = f.Type
				fp.Size = f.Size
				fp.Spacing = f.Spacing
				fp.Offset = f.offset
			}
		}
	}

	// Build directly from the struct values so we don't need to populate everything.
	v := reflect.ValueOf(&m.TitleInfo.Loading).Elem()
	f := v.FieldByName("TextSpriteData")
	if f.IsValid() && f.CanSet() && f.IsNil() {
		SetTextSprite(m, f, v, reflect.Value{})
	}

	ts := m.TitleInfo.Loading.TextSpriteData
	if ts == nil {
		return
	}
	ts.SetLocalcoord(float32(m.Info.Localcoord[0]), float32(m.Info.Localcoord[1]))
	ts.SetWindow([4]float32{
		0, 0,
		float32(m.Info.Localcoord[0]),
		float32(m.Info.Localcoord[1]),
	})
	ts.Reset()

	sys.mainThreadTask <- func() {
		// End previous frame
		gfx.EndFrame()

		// Start a one-off frame
		gfx.BeginFrame(true)

		FillRect(sys.scrrect, 0x000000, [2]int32{255, 0})

		ts.Draw(ts.layerno)

		// Submit and present
		gfx.EndFrame()
		if strings.HasPrefix(gfx.GetName(), "OpenGL") {
			sys.window.SwapBuffers()
		} else {
			gfx.Await()
		}

		// Prepare a fresh frame for whatever comes next (no clear)
		gfx.BeginFrame(false)
	}
	sys.runMainThreadTask()
}

// GetValue retrieves the value based on the query string.
func (m *Motif) GetValue(query string) (interface{}, error) {
	return GetValue(m, query)
}

// SetValue sets the value based on the query string and updates the IniFile.
func (m *Motif) SetValueUpdate(query string, value interface{}) error {
	return SetValueUpdate(m, m.IniFile, query, value)
}

// Save writes the current IniFile to disk, preserving comments and syntax.
func (m *Motif) Save(file string) error {
	return SaveINI(m.IniFile, file)
}

func (mo *Motif) processStateChange(c *Char, states []int32) bool {
	for _, stateNo := range states {
		// Negative state numbers are treated as hide/disable this character.
		if stateNo < 0 {
			c.setSCF(SCF_disabled)
			return true
		}
		if c.selfStatenoExist(BytecodeInt(stateNo)) == BytecodeBool(true) {
			// If the character was previously hidden/disabled, re-enable it when a valid state transition is requested.
			c.unsetSCF(SCF_disabled)
			c.changeState(int32(stateNo), -1, -1, "")
			return true
		}
	}
	return false
}

// processStateTransitions applies state changes by win/lose outcome.
func (mo *Motif) processStateTransitions(winnerState, winnerTeammateState, loserState, loserTeammateState []int32) {
	isWinnerLeader, isLoserLeader := false, false
	for _, p := range sys.chars {
		if len(p) == 0 {
			continue
		}
		c := p[0]
		if c.win() {
			// Handle P1 state
			if !isWinnerLeader {
				mo.processStateChange(c, winnerState)
				isWinnerLeader = true
			} else {
				mo.processStateChange(c, winnerTeammateState)
			}
		} else {
			// Handle P2 state
			if !isLoserLeader {
				mo.processStateChange(c, loserState)
				isLoserLeader = true
			} else {
				mo.processStateChange(c, loserTeammateState)
			}
		}
	}
}

// applies state changes by team side, not by win/lose outcome.
func (mo *Motif) processStateTransitionsBySide(p1State, p1TeammateState, p2State, p2TeammateState []int32) {
	// Track whether we've assigned a "leader" state for each side.
	leaderDone := [2]bool{false, false}

	// Helper to choose state slices by side.
	getStates := func(side int) (leader []int32, mate []int32) {
		if side == 0 {
			return p1State, p1TeammateState
		}
		return p2State, p2TeammateState
	}

	// 1) Prefer applying leader state to the actual team leader (if available).
	// sys.teamLeader[] is used elsewhere in this file, so it should be stable here too.
	for side := 0; side < 2; side++ {
		leaderPn := sys.teamLeader[side]
		if leaderPn < 0 || leaderPn >= len(sys.chars) || len(sys.chars[leaderPn]) == 0 {
			continue
		}
		c := sys.chars[leaderPn][0]
		// Only accept if the character is actually on that side.
		if int(c.teamside) != side {
			continue
		}
		ls, _ := getStates(side)
		mo.processStateChange(c, ls)
		leaderDone[side] = true
	}

	// 2) Apply remaining members in stable playerNo order.
	// First encountered member becomes leader if we couldn't resolve a leader above.
	for pn, p := range sys.chars {
		if len(p) == 0 {
			continue
		}
		c := p[0]
		side := int(c.teamside)
		if side < 0 || side > 1 {
			continue
		}

		// If we already applied to the explicit leaderPn, skip it here.
		if leaderDone[side] && pn == sys.teamLeader[side] {
			continue
		}

		ls, ms := getStates(side)
		if !leaderDone[side] {
			mo.processStateChange(c, ls)
			leaderDone[side] = true
		} else {
			mo.processStateChange(c, ms)
		}
	}
}

func (m *Motif) reset() {
	if m.IniFile == nil {
		return
	}
	m.fadeIn.reset()
	m.fadeOut.reset()
	m.ch.reset(m)
	m.de.reset(m)
	m.di.reset(m)
	m.di.clear(m)
	m.vi.reset(m)
	m.vi.clear(m)
	m.wi.reset(m)
	m.hi.reset(m)
	m.co.reset(m)
	//sys.storyboard.reset()
	m.me.reset(m)
}

func (m *Motif) step() {
	if m.me.active {
		m.me.step(m)
	} else if sys.escExit() {
		sys.endMatch = true
		return
	} else if sys.esc && sys.paused {
		sys.paused = false
		return
	}
	sys.stepCommandLists()
	if sys.paused && !sys.frameStepFlag {
		return
	}
	if m.ch.active {
		m.ch.step(m)
	}
	if m.de.active {
		m.de.step(m)
	}
	if m.di.active {
		m.di.step(m)
	}
	if m.vi.active {
		m.vi.step(m)
	}
	if m.wi.active {
		m.wi.step(m)
	}
	if m.hi.active {
		m.hi.step(m)
	}
	if m.co.active {
		m.co.step(m)
	}
	if sys.storyboard.active {
		sys.storyboard.step()
	}
}

// drawAspectBars renders black bars when the fight aspect and motif aspect differ.
func (m *Motif) drawAspectBars() {
	fightAspect := sys.getFightAspect()
	motifAspect := sys.getMotifAspect()

	if fightAspect <= 0 || motifAspect <= 0 || fightAspect == motifAspect {
		return
	}

	sw := sys.scrrect[2]
	sh := sys.scrrect[3]

	// Collect up to two bar rectangles (pillarbox or letterbox).
	var rects [][4]int32

	if fightAspect < motifAspect {
		// Fight view is narrower than the motif (e.g. 4:3 fight on 16:9 motif):
		// add vertical bars on the left and right.
		contentWidth := int32(float32(sh) * fightAspect)
		if contentWidth > 0 && contentWidth < sw {
			offsetX := (sw - contentWidth) / 2
			leftBar := [4]int32{0, 0, offsetX, sh}
			rightBarWidth := sw - (offsetX + contentWidth)
			if rightBarWidth < 0 {
				rightBarWidth = 0
			}
			rightBar := [4]int32{offsetX + contentWidth, 0, rightBarWidth, sh}
			rects = append(rects, leftBar, rightBar)
		}
	} else if fightAspect > motifAspect {
		// Fight view is wider than the motif: add horizontal bars top and bottom.
		contentHeight := int32(float32(sw) / fightAspect)
		if contentHeight > 0 && contentHeight < sh {
			offsetY := (sh - contentHeight) / 2
			topBar := [4]int32{0, 0, sw, offsetY}
			bottomBarHeight := sh - (offsetY + contentHeight)
			if bottomBarHeight < 0 {
				bottomBarHeight = 0
			}
			bottomBar := [4]int32{0, offsetY + contentHeight, sw, bottomBarHeight}
			rects = append(rects, topBar, bottomBar)
		}
	}

	for _, r := range rects {
		if r[2] > 0 && r[3] > 0 {
			// 0x000000 = black, fully opaque.
			FillRect(r, 0x000000, [2]int32{255, 0})
		}
	}
}

func (m *Motif) draw(layerno int16) {
	// Draw black bars if fight aspect and motif aspect differ.
	if layerno == 1 && (!sys.middleOfMatch() || m.me.active || m.di.active) && !sys.skipMotifScaling() {
		m.drawAspectBars()
	}
	if m.ch.active {
		m.ch.draw(m, layerno)
	}
	if m.de.active {
		m.de.draw(m, layerno)
	}
	if m.di.active {
		m.di.draw(m, layerno)
	}
	if m.vi.active {
		m.vi.draw(m, layerno)
	}
	if m.wi.active {
		m.wi.draw(m, layerno)
	}
	if m.hi.active {
		m.hi.draw(m, layerno)
	}
	if m.co.active {
		m.co.draw(m, layerno)
	}
	if sys.storyboard.active {
		sys.storyboard.draw(layerno)
	}
	if m.me.active {
		m.me.draw(m, layerno)
	}
	// Screen fading
	if layerno == 3 {
		if m.fadeOut.isActive() {
			m.fadeOut.draw()
		} else if m.fadeIn.isActive() {
			m.fadeIn.draw()
		}
	}
}

func (m *Motif) isDialogueSet() bool {
	if sys.dialogueForce != 0 {
		pn := sys.dialogueForce
		return pn >= 1 && pn <= len(sys.chars) &&
			len(sys.chars[pn-1]) > 0 &&
			len(sys.chars[pn-1][0].dialogue) > 0
	}
	for _, p := range sys.chars {
		if len(p) > 0 && len(p[0].dialogue) > 0 {
			return true
		}
	}
	return false
}

func (m *Motif) act() {
	if (sys.paused && !sys.frameStepFlag) || sys.gsf(GSF_roundfreeze) {
		return
	}
	// Storyboard
	//if sys.storyboard.IniFile != nil && !sys.storyboard.initialized  {
	//	sys.storyboard.init()
	//}
	// Menu / Exit
	if !m.me.initialized {
		m.me.init(m)
	}
	// Demo Mode
	if !m.de.initialized {
		m.de.init(m)
	}
	if sys.postMatchFlg {
		// Victory Screen
		if !m.vi.initialized {
			m.vi.init(m)
		}
		if m.vi.active {
			return
		}
		// Win / Results Screen
		if !m.wi.initialized {
			m.wi.init(m)
		}
		if m.wi.active {
			return
		}
		// Continue Screen
		if !m.co.initialized {
			m.co.init(m)
		}
		if m.co.active {
			return
		}
		sys.postMatchFlg = false
	} else {
		// Challenger
		if !m.ch.initialized {
			m.ch.init(m)
		}
		// TODO: Hiscore

		// Dialogue
		// Normal start: right before "Fight" in round 1, or at match end.
		normalStart := (sys.round == 1 && sys.intro == sys.lifebar.ro.ctrl_time) ||
			(sys.roundStateTicks() == sys.lifebar.ro.fadeOut.time && sys.matchOver())
		// Forced start: ignore normal timing.
		forcedStart := sys.dialogueForce != 0
		if !m.di.initialized && (forcedStart || normalStart) && m.isDialogueSet() {
			m.di.init(m)
		}
	}
}

func formatTimeText(text string, totalSec float64) string {
	h := int(totalSec / 3600)
	m := int(totalSec/60) % 60
	s := int(totalSec) % 60
	x := int((totalSec - float64(int(totalSec))) * 100)
	// Ensure two-digit formatting for minutes, seconds, and fractions
	mStr := fmt.Sprintf("%02d", m)
	sStr := fmt.Sprintf("%02d", s)
	xStr := fmt.Sprintf("%02d", x)
	// Replace placeholders
	result := strings.ReplaceAll(text, "%h", fmt.Sprintf("%d", h))
	result = strings.ReplaceAll(result, "%m", mStr)
	result = strings.ReplaceAll(result, "%s", sStr)
	result = strings.ReplaceAll(result, "%x", xStr)
	return result
}

// replaceFormatSpecifiers converts Lua 5.1/C-style format specifiers
// to Go's fmt.Sprintf equivalents where they differ.
func (mo *Motif) replaceFormatSpecifiers(input string) string {
	// Verbs that exist in Lua's string.format/C but not in Go's fmt: %i, %u, %I
	// Everything else we care about (%d, %o, %x, %X, %e, %E, %f, %g, %G, %c, %s, %q, %p, %%) is already understood by fmt.
	var formatSpecifierMap = map[string]string{
		"i": "d",
		"I": "d",
		"u": "d",
	}
	re := regexp.MustCompile(`%%|%([-+ #0]*)?(\d+)?(\.\d+)?([hlLzjt]*)?([a-zA-Z])`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Keep literal %%
		if match == "%%" {
			return "%%"
		}
		sub := re.FindStringSubmatch(match)
		if len(sub) != 6 {
			return match // should not happen, but be safe
		}
		flags := sub[1]
		width := sub[2]
		precision := sub[3]
		// length := sub[4] // dropped for Go
		verb := sub[5]
		// Map Lua/C-only verbs to Go equivalents
		if mapped, ok := formatSpecifierMap[verb]; ok {
			verb = mapped
		}
		// Rebuild without the C length modifier.
		var b strings.Builder
		b.WriteByte('%')
		if flags != "" {
			b.WriteString(flags)
		}
		if width != "" {
			b.WriteString(width)
		}
		if precision != "" {
			b.WriteString(precision)
		}
		b.WriteString(verb)
		return b.String()
	})
}

// Counts fmt verbs in a format string, ignoring literal %%.
func countFmtVerbs(format string) int {
	re := regexp.MustCompile(`%%|%([-+ #0]*)?(\d+)?(\.\d+)?([hlLzjt]*)?([a-zA-Z])`)
	matches := re.FindAllString(format, -1)
	n := 0
	for _, m := range matches {
		if m != "%%" {
			n++
		}
	}
	return n
}

// fmt.Sprintf helper. Normalizes Lua/C-style verbs, supports any number of args
func (mo *Motif) sprintf(format string, args ...interface{}) string {
	fs := mo.replaceFormatSpecifiers(format)
	// Trim extra args so legacy one-placeholder strings still work when callers pass (wins, losses).
	if n := countFmtVerbs(fs); n >= 0 && len(args) > n {
		args = args[:n]
	}
	return fmt.Sprintf(fs, args...)
}

type MotifMenu struct {
	enabled        bool
	active         bool
	initialized    bool
	counter        int32
	endTimer       int32
	closeRequested bool
	reopenLock     bool
}

func (me *MotifMenu) reset(m *Motif) {
	me.active = false
	me.initialized = false
	me.endTimer = -1
	me.closeRequested = false
	if !m.di.active && !sys.skipMotifScaling() {
		sys.applyFightAspect()
	}
	if err := sys.luaLState.DoString("menuReset()"); err != nil {
		sys.luaLState.RaiseError("Error executing Lua code: %v\n", err.Error())
	}
}

// returns true if the pause/menu open input is still physically held.
func (me *MotifMenu) menuOpenInputHeld(m *Motif) bool {
	// Physical ESC hold
	if sys.keyState != nil && sys.keyState[KeyEscape] {
		return true
	}
	// Also respect configured menu cancel/open bindings
	if m != nil && m.PauseMenu != nil {
		if pm := m.PauseMenu["pause_menu"]; pm != nil && pm.Enabled {
			if sys.button(pm.Menu.Cancel.Key, -1) {
				return true
			}
		}
	}
	return false
}

func (me *MotifMenu) pauseMenuBase(m *Motif) *MenuInfoProperties {
	if m == nil || m.PauseMenu == nil {
		return nil
	}
	if pm := m.PauseMenu["pause_menu"]; pm != nil {
		return pm
	}
	for _, pm := range m.PauseMenu {
		if pm != nil {
			return pm
		}
	}
	return nil
}

func (me *MotifMenu) requestClose(m *Motif) {
	if me.endTimer != -1 {
		return
	}
	me.closeRequested = true
	if pm := me.pauseMenuBase(m); pm != nil {
		startFadeOut(pm.FadeOut.FadeData, m.fadeOut, false, m.fadePolicy)
	}
	me.endTimer = me.counter + m.fadeOut.timeRemaining
}

func (me *MotifMenu) init(m *Motif) {
	pm := me.pauseMenuBase(m)
	if pm == nil || !pm.Enabled || !me.enabled {
		me.initialized = true
		return
	}
	// Don't allow the menu to instantly re-open if the open/cancel key is still held after closing.
	if me.reopenLock {
		if me.menuOpenInputHeld(m) {
			return
		}
		me.reopenLock = false
	}
	if (!sys.esc && !sys.button(pm.Menu.Cancel.Key, -1)) || m.ch.active || sys.postMatchFlg {
		return
	}
	if !sys.skipMotifScaling() {
		sys.setGameSize(sys.scrrect[2], sys.scrrect[3])
	}

	if err := sys.luaLState.DoString("menuInit()"); err != nil {
		sys.luaLState.RaiseError("Error executing Lua code: %v\n", err.Error())
	}

	pm.FadeIn.FadeData.init(m.fadeIn, true)
	me.counter = 0
	me.active = true
	me.initialized = true
	me.closeRequested = false
	me.endTimer = -1
}

func (me *MotifMenu) step(m *Motif) {
	// Close only when requested by Lua (menuRun() returned false) or when ending the match.
	if me.endTimer == -1 && (me.closeRequested || sys.endMatch) {
		me.requestClose(m)
	}

	// Check if the sequence has ended
	if me.endTimer != -1 && me.counter >= me.endTimer {
		if m.fadeOut != nil {
			m.fadeOut.reset()
		}
		me.active = false
		me.closeRequested = false
		me.reopenLock = true
		me.reset(m)
		sys.paused = false
		return
	}

	// Increment counter
	me.counter++
}

func (me *MotifMenu) draw(m *Motif, layerno int16) {
	if layerno == 2 {
		// Once closing has started, stop running the Lua menu loop so it can't keep drawing/flickering.
		if me.endTimer != -1 {
			return
		}
		if ok, err := ExecFunc(sys.luaLState, "menuRun"); err != nil {
			sys.luaLState.RaiseError("Error executing Lua code: %v\n", err.Error())
		} else if !ok {
			// Lua requested to close the pause menu (menuRun returns main.pauseMenu).
			me.requestClose(m)
		}
	}
}

type MotifChallenger struct {
	enabled       bool
	active        bool
	initialized   bool
	counter       int32
	endTimer      int32
	controllerNo  int
	lifebarActive bool
}

func (ch *MotifChallenger) reset(m *Motif) {
	ch.active = false
	ch.initialized = false
	ch.endTimer = -1
	ch.controllerNo = -1
	//if !sys.skipMotifScaling() {
	//	sys.applyFightAspect()
	//}
}

func (ch *MotifChallenger) init(m *Motif) {
	if !m.ChallengerInfo.Enabled || !ch.enabled {
		ch.initialized = true
		return
	}

	controllerNo := sys.buttonController(m.ChallengerInfo.Key)
	if controllerNo == -1 || controllerNo == sys.chars[0][0].controller {
		return
	}
	ch.controllerNo = controllerNo
	//if !sys.skipMotifScaling() {
	//	sys.setGameSize(sys.scrrect[2], sys.scrrect[3])
	//}

	if err := sys.luaLState.DoString("hook.run('game.challenger_init')"); err != nil {
		sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.challenger_init", err.Error())
	}

	if m.AttractMode.Enabled && sys.credits > 0 {
		sys.credits--
	}

	ch.lifebarActive = sys.lifebar.active
	sys.lifebar.active = false

	m.ChallengerBgDef.BGDef.Reset()
	m.ChallengerInfo.Bg.AnimData.Reset()

	m.ChallengerInfo.FadeIn.FadeData.init(m.fadeIn, true)
	ch.counter = 0
	ch.active = true
	ch.initialized = true
}

func (ch *MotifChallenger) step(m *Motif) {
	if ch.counter > 0 {
		if err := sys.luaLState.DoString("hook.run('game.challenger')"); err != nil {
			sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.challenger", err.Error())
		}
	}
	if ch.endTimer == -1 && ch.counter == m.ChallengerInfo.Time {
		startFadeOut(m.ChallengerInfo.FadeOut.FadeData, m.fadeOut, false, m.fadePolicy)
		ch.endTimer = ch.counter + m.fadeOut.timeRemaining
	}
	sys.setGSF(GSF_nobardisplay)
	sys.setGSF(GSF_nomusic)
	sys.setGSF(GSF_timerfreeze)
	if ch.counter == m.ChallengerInfo.Pause.Time {
		sys.pausetime = m.ChallengerInfo.Time + m.ChallengerInfo.FadeOut.Time
	}
	if ch.counter == m.ChallengerInfo.Snd.Time {
		m.Snd.play(m.ChallengerInfo.Snd.Snd, 100, 0, 0, 0, 0)
	}
	if ch.counter >= m.ChallengerInfo.Bg.Displaytime {
		m.ChallengerInfo.Bg.AnimData.Update(false)
	}

	//if ch.endTimer != -1 && ch.counter + 2 >= ch.endTimer {
	//	sys.endMatch = true
	//}

	// Check if the sequence has ended
	if ch.endTimer != -1 && ch.counter >= ch.endTimer {
		if m.fadeOut != nil {
			m.fadeOut.reset()
		}
		ch.active = false
		sys.lifebar.active = ch.lifebarActive
		sys.endMatch = true
		return
	}

	// Increment counter
	ch.counter++
}

func (ch *MotifChallenger) draw(m *Motif, layerno int16) {
	// Background
	if m.ChallengerBgDef.BgClearColor[0] >= 0 {
		m.ChallengerBgDef.RectData.Draw(layerno)
	}
	if layerno <= 1 {
		m.ChallengerBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
	}
	// Overlay
	m.ChallengerInfo.Overlay.RectData.Draw(layerno)
	// Text
	if ch.counter >= m.ChallengerInfo.Text.Displaytime {
		m.ChallengerInfo.Text.TextSpriteData.Draw(layerno)
	}
	// Bg
	if ch.counter >= m.ChallengerInfo.Bg.Displaytime {
		m.ChallengerInfo.Bg.AnimData.Draw(layerno)
	}
	// Top background
	if layerno == 2 {
		m.ChallengerBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
	}
}

type MotifContinue struct {
	enabled     bool
	active      bool
	initialized bool
	counter     int32
	endTimer    int32
	showEndAnim bool
	credits     int32
	yesSide     bool
	selected    bool
	counts      []string
	pn          int
}

func (co *MotifContinue) reset(m *Motif) {
	sys.continueFlg = false
	co.active = false
	co.initialized = false
	co.yesSide = true
	co.selected = false
	co.endTimer = -1
	co.showEndAnim = false
	if !sys.skipMotifScaling() {
		sys.applyFightAspect()
	}
}

func (co *MotifContinue) extractAndSortKeysDescending(m *Motif) []string {
	keys := make([]string, 0, len(m.ContinueScreen.Counter.MapCounts))
	for key := range m.ContinueScreen.Counter.MapCounts {
		keys = append(keys, key)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	return keys
}

func (co *MotifContinue) updateCreditsText(m *Motif) {
	formattedText := fmt.Sprintf(m.replaceFormatSpecifiers(m.ContinueScreen.Credits.Text), sys.credits)
	m.ContinueScreen.Credits.TextSpriteData.text = formattedText
	co.credits = sys.credits
}

func (co *MotifContinue) init(m *Motif) {
	if (!m.ContinueScreen.Enabled || !co.enabled || sys.cfg.Options.QuickContinue) ||
		(sys.winnerTeam() != 0 && sys.winnerTeam() != int32(sys.home)+1) ||
		!sys.sel.gameParams.Continue {
		co.initialized = true
		return
	}
	if !sys.skipMotifScaling() {
		sys.setGameSize(sys.scrrect[2], sys.scrrect[3])
	}
	if err := sys.luaLState.DoString("hook.run('game.continue_init')"); err != nil {
		sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.continue_init", err.Error())
	}

	co.pn = 1 // TODO: Initialize pn appropriately

	// Extract and sort keys in descending order
	co.counts = co.extractAndSortKeysDescending(m)

	m.ContinueBgDef.BGDef.Reset()

	m.ContinueScreen.Continue.TextSpriteData.Reset()
	m.ContinueScreen.Continue.TextSpriteData.AddPos(m.ContinueScreen.Pos[0], m.ContinueScreen.Pos[1])

	m.ContinueScreen.Yes.TextSpriteData.Reset()
	m.ContinueScreen.Yes.TextSpriteData.AddPos(m.ContinueScreen.Pos[0], m.ContinueScreen.Pos[1])

	m.ContinueScreen.Yes.Active.TextSpriteData.Reset()
	m.ContinueScreen.Yes.Active.TextSpriteData.AddPos(m.ContinueScreen.Pos[0], m.ContinueScreen.Pos[1])

	m.ContinueScreen.No.TextSpriteData.Reset()
	m.ContinueScreen.No.TextSpriteData.AddPos(m.ContinueScreen.Pos[0], m.ContinueScreen.Pos[1])

	m.ContinueScreen.No.Active.TextSpriteData.Reset()
	m.ContinueScreen.No.Active.TextSpriteData.AddPos(m.ContinueScreen.Pos[0], m.ContinueScreen.Pos[1])

	m.ContinueScreen.Counter.AnimData.Reset()
	//m.ContinueScreen.Counter.AnimData.Update(false)

	co.updateCreditsText(m)

	// Handle state transitions
	m.processStateTransitions(m.ContinueScreen.P2.State, m.ContinueScreen.P2.Teammate.State, m.ContinueScreen.P1.State, m.ContinueScreen.P1.Teammate.State)

	co.yesSide = true

	if !m.ContinueScreen.Sounds.Enabled {
		sys.clearAllSound()
		sys.noSoundFlg = true
	}

	m.Music.Play("continue", sys.motif.Def)

	m.ContinueScreen.FadeIn.FadeData.init(m.fadeIn, true)
	co.counter = 0
	co.active = true
	co.initialized = true
	co.showEndAnim = false
}

func (co *MotifContinue) processSelection(m *Motif, continueSelected bool) {
	cs := m.ContinueScreen
	if continueSelected {
		m.processStateTransitions(
			cs.P2.Yes.State,
			cs.P2.Teammate.Yes.State,
			cs.P1.Yes.State,
			cs.P1.Teammate.Yes.State,
		)
		sys.continueFlg = true
		if sys.credits != -1 {
			sys.credits--
		}
	} else {
		m.processStateTransitions(
			cs.P2.No.State,
			cs.P2.Teammate.No.State,
			cs.P1.No.State,
			cs.P1.Teammate.No.State,
		)
	}
	startFadeOut(m.ContinueScreen.FadeOut.FadeData, m.fadeOut, false, m.fadePolicy)
	co.endTimer = co.counter + m.fadeOut.timeRemaining
	co.selected = true
}

func (co *MotifContinue) skipCounter(m *Motif) {
	for _, key := range co.counts {
		properties := m.ContinueScreen.Counter.MapCounts[key]
		if co.counter < properties.SkipTime {
			for co.counter < properties.SkipTime {
				co.counter++
				m.ContinueScreen.Counter.AnimData.Update(true)
			}
			break
		}
	}
}

func (co *MotifContinue) playCounterSounds(m *Motif) {
	for _, key := range co.counts {
		properties := m.ContinueScreen.Counter.MapCounts[key]
		if co.counter == properties.SkipTime {
			m.Snd.play(properties.Snd, 100, 0, 0, 0, 0)
			break
		}
	}
}

func (co *MotifContinue) step(m *Motif) {
	if co.counter > 0 {
		if err := sys.luaLState.DoString("hook.run('game.continue')"); err != nil {
			sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.continue", err.Error())
		}
	}
	if co.credits != sys.credits {
		co.updateCreditsText(m)
		if !co.selected {
			co.counter = 0
			m.ContinueScreen.Counter.AnimData.Reset()
		}
	}

	// Keep the counter anim running while showing the integrated end/gameover tail.
	if !co.selected || co.showEndAnim {
		m.ContinueScreen.Counter.AnimData.Update(false)
	}

	// If we're showing the integrated end/gameover tail (gameover.enabled = 0),
	// defer fadeout until the counter animation reaches endtime.
	if co.selected && co.showEndAnim && co.endTimer == -1 &&
		m.ContinueScreen.Counter.EndTime > 0 &&
		co.counter >= m.ContinueScreen.Counter.EndTime {
		startFadeOut(m.ContinueScreen.FadeOut.FadeData, m.fadeOut, false, m.fadePolicy)
		co.endTimer = co.counter + m.fadeOut.timeRemaining
	}

	if !co.selected {
		if m.ContinueScreen.LegacyMode.Enabled {
			if sys.button(m.ContinueScreen.Move.Key, co.pn-1) {
				m.Snd.play(m.ContinueScreen.Move.Snd, 100, 0, 0, 0, 0)
				co.yesSide = !co.yesSide
			} else if sys.button(m.ContinueScreen.Skip.Key, co.pn-1) || sys.button(m.ContinueScreen.Done.Key, co.pn-1) {
				m.Snd.play(m.ContinueScreen.Done.Snd, 100, 0, 0, 0, 0)
				co.processSelection(m, co.yesSide)
			}
		} else {
			if co.counter < m.ContinueScreen.Counter.End.SkipTime {
				if (sys.credits == -1 || sys.credits > 0) && sys.button(m.ContinueScreen.Done.Key, co.pn-1) {
					m.Snd.play(m.ContinueScreen.Done.Snd, 100, 0, 0, 0, 0)
					co.processSelection(m, true)
				} else if sys.button(m.ContinueScreen.Skip.Key, co.pn-1) &&
					co.counter >= m.ContinueScreen.Counter.StartTime+m.ContinueScreen.Counter.SkipStart {
					co.skipCounter(m)
				}
				co.playCounterSounds(m)
			} else if co.counter == m.ContinueScreen.Counter.End.SkipTime {
				m.Snd.play(m.ContinueScreen.Counter.End.Snd, 100, 0, 0, 0, 0)
				m.Music.Play("continue.end", sys.motif.Def)
				// If separate gameover screen is disabled, the end/gameover portion is integrated into the same counter animation.
				// Let it play to endtime, then fade out.
				if !m.ContinueScreen.GameOver.Enabled {
					cs := m.ContinueScreen
					m.processStateTransitions(
						cs.P2.No.State,
						cs.P2.Teammate.No.State,
						cs.P1.No.State,
						cs.P1.Teammate.No.State,
					)
					co.selected = true
					co.showEndAnim = true
					// fadeout will be started when counter reaches Counter.EndTime
				} else {
					co.processSelection(m, false)
				}
			}
		}
	}

	// Check if the sequence has ended
	if co.selected && co.endTimer != -1 && co.counter >= co.endTimer {
		if m.fadeOut != nil {
			m.fadeOut.reset()
		}
		co.active = false
		if !m.ContinueScreen.Sounds.Enabled {
			sys.noSoundFlg = false
		}
		return
	}
	// Increment counter
	co.counter++
}

func (co *MotifContinue) drawLegacyMode(m *Motif, layerno int16) {
	// Continue
	m.ContinueScreen.Continue.TextSpriteData.Draw(layerno)
	// Yes / No
	if co.yesSide {
		m.ContinueScreen.Yes.Active.TextSpriteData.Draw(layerno)
		m.ContinueScreen.No.TextSpriteData.Draw(layerno)
	} else {
		m.ContinueScreen.Yes.TextSpriteData.Draw(layerno)
		m.ContinueScreen.No.Active.TextSpriteData.Draw(layerno)
	}
}

func (co *MotifContinue) draw(m *Motif, layerno int16) {
	// Overlay
	m.ContinueScreen.Overlay.RectData.Draw(layerno)
	// Background
	if m.ContinueBgDef.BgClearColor[0] >= 0 {
		m.ContinueBgDef.RectData.Draw(layerno)
	}
	if layerno <= 1 {
		m.ContinueBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
	}
	// Mugen style
	if m.ContinueScreen.LegacyMode.Enabled {
		co.drawLegacyMode(m, layerno)
	} else if !co.selected || co.showEndAnim {
		// Arcade style Counter
		m.ContinueScreen.Counter.AnimData.Draw(layerno)
	}
	// Credits
	if sys.credits != -1 && co.counter >= m.ContinueScreen.Counter.SkipStart {
		m.ContinueScreen.Credits.TextSpriteData.Draw(layerno)
	}
	// Top background
	if layerno == 2 {
		m.ContinueBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
	}
}

type MotifDemo struct {
	enabled     bool
	active      bool
	initialized bool
	counter     int32
	endTimer    int32
}

func (de *MotifDemo) reset(m *Motif) {
	de.active = false
	de.initialized = false
	de.endTimer = -1
}

func (de *MotifDemo) init(m *Motif) {
	if !m.DemoMode.Enabled || !de.enabled || sys.gameMode != "demo" {
		de.initialized = true
		return
	}

	de.counter = 0

	// Override lifebar fading
	m.DemoMode.FadeIn.FadeData.init(sys.lifebar.ro.fadeIn, true)

	de.active = true
	de.initialized = true
}

func (de *MotifDemo) step(m *Motif) {
	if de.endTimer == -1 {
		cancel := (m.AttractMode.Enabled && sys.credits > 0) || (!m.AttractMode.Enabled && sys.button(m.DemoMode.Cancel.Key, -1))
		if de.counter == m.DemoMode.Fight.EndTime || cancel {
			startFadeOut(m.DemoMode.FadeOut.FadeData, sys.lifebar.ro.fadeOut, cancel, m.fadePolicy)
			de.endTimer = de.counter + sys.lifebar.ro.fadeOut.timeRemaining
		}
	}

	// Check if the sequence has ended
	if de.endTimer != -1 && de.counter >= de.endTimer {
		if sys.lifebar.ro.fadeOut != nil {
			sys.lifebar.ro.fadeOut.reset()
		}
		de.active = false
		sys.endMatch = true
		return
	}

	// Increment counter
	de.counter++
}

func (de *MotifDemo) draw(m *Motif, layerno int16) {
	// nothing to draw, may be expanded in future
}

// MotifDialogue is the top-level container for storing parsed dialogue data.
type MotifDialogue struct {
	enabled           bool
	active            bool
	initialized       bool
	counter           int32
	char              *Char
	faceParams        [2]FaceParams
	parsed            []DialogueParsedLine
	textNum           int
	lineFullyRendered bool
	charDelayCounter  int32
	activeSide        int
	talkingSide       int
	wait              int
	switchCounter     int
	endCounter        int
}

type FaceParams struct {
	grp int
	idx int
	pn  int
}

type DialogueParsedLine struct {
	side     int
	text     string
	tokens   map[int][]DialogueToken
	typedCnt int
}

type DialogueToken struct {
	param       string
	side        int
	redirection string
	pn          int
	value       []interface{}
}

// isValidPlayerNo returns true if pn refers to an existing character entry.
func (di *MotifDialogue) isValidPlayerNo(pn int) bool {
	if pn < 1 || pn > len(sys.chars) {
		return false
	}
	return len(sys.chars[pn-1]) > 0
}

func (di *MotifDialogue) dialogueRedirection(redirect string) int {
	var redirection, val string
	if parts := strings.SplitN(redirect, "(", 2); len(parts) == 2 {
		redirection = strings.ToLower(strings.TrimSpace(parts[0]))
		val = strings.TrimSpace(strings.TrimSuffix(parts[1], ")"))
	} else {
		redirection = strings.ToLower(strings.TrimSpace(redirect))
	}
	switch redirection {
	case "self":
		return di.char.playerNo + 1
	case "playerno":
		pn := int(Atoi(val))
		if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
			return pn
		}
	case "partner":
		if val == "" {
			val = "0"
		}
		partnerNum := Atoi(val)
		if partner := di.char.partner(partnerNum, true); partner != nil {
			return partner.playerNo + 1
		}
	case "enemy":
		if val == "" {
			val = "0"
		}
		enemyNum := Atoi(val)
		if enemy := di.char.enemy(enemyNum); enemy != nil {
			return enemy.playerNo + 1
		}
	case "enemyname":
		for i := int32(0); i < di.char.numEnemy(); i++ {
			if enemy := di.char.enemy(i); enemy != nil {
				if strings.EqualFold(enemy.name, val) {
					return enemy.playerNo + 1
				}
			}
		}
	case "partnername":
		for i := int32(0); i < di.char.numPartner(); i++ {
			if partner := di.char.partner(i, false); partner != nil {
				if strings.EqualFold(partner.name, val) {
					return partner.playerNo + 1
				}
			}
		}
	default:
	}
	return -1
}

func (di *MotifDialogue) parseTag(tag string) []DialogueToken {
	tag = strings.TrimSpace(tag)
	pOnlyRe := regexp.MustCompile(`(?i)^p(\d+)$`)
	if pOnlyRe.MatchString(tag) {
		matches := pOnlyRe.FindStringSubmatch(tag)
		if len(matches) == 2 {
			pnValue := int(Atoi(matches[1]))
			return []DialogueToken{{
				param: "p",
				side:  -1,
				pn:    pnValue,
			}}
		}
	}
	equalIndex := strings.Index(tag, "=")
	if equalIndex == -1 {
		return nil
	}
	paramPart := strings.TrimSpace(tag[:equalIndex])
	valuePart := strings.TrimSpace(tag[equalIndex+1:])
	side := -1
	param := paramPart
	redirection := ""
	pn := -1
	numValues := []interface{}{}
	pPrefixRe := regexp.MustCompile(`(?i)^p(\d+)([a-zA-Z]+)$`)
	if pPrefixRe.MatchString(paramPart) {
		subMatches := pPrefixRe.FindStringSubmatch(paramPart)
		if len(subMatches) == 3 {
			side = int(Atoi(subMatches[1]))
			param = subMatches[2]
		}
	}
	param = strings.ToLower(strings.TrimSpace(param))
	parts := strings.Split(valuePart, ",")
	if len(parts) > 0 {
		head := strings.TrimSpace(parts[0])
		parts[0] = head
		if !IsInt(head) {
			redirection = head
			parts = parts[1:]
		}
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if IsNumeric(p) {
				numValues = append(numValues, float32(Atof(p)))
			} else {
				numValues = append(numValues, p)
			}
		}
		pn = di.dialogueRedirection(redirection)
	}
	return []DialogueToken{{
		param:       param,
		side:        side,
		redirection: redirection,
		pn:          pn,
		value:       numValues,
	}}
}

func (di *MotifDialogue) parseLine(line string) DialogueParsedLine {
	side := 1
	re := regexp.MustCompile(`<([^>]+)>`)
	var finalText strings.Builder
	tokensMap := make(map[int][]DialogueToken)
	offset := 0
	pos := 0
	matches := re.FindAllStringIndex(line, -1)
	for _, match := range matches {
		startIdx := match[0]
		endIdx := match[1]
		if startIdx > pos {
			substr := line[pos:startIdx]
			finalText.WriteString(substr)
			offset += utf8.RuneCountInString(substr)
		}
		tokenContent := line[startIdx+1 : endIdx-1]
		parsedTokens := di.parseTag(tokenContent)
		if len(parsedTokens) == 1 && parsedTokens[0].param == "p" && parsedTokens[0].pn != -1 {
			side = parsedTokens[0].pn
		} else {
			for _, tkn := range parsedTokens {
				tokensMap[offset] = append(tokensMap[offset], tkn)
			}
		}
		pos = endIdx
	}
	if pos < len(line) {
		substr := line[pos:]
		finalText.WriteString(substr)
		offset += utf8.RuneCountInString(substr)
	}
	//fmt.Printf("[Dialogue] parseLine: side=%d text=%q tokens=%v\n", side, strings.TrimSpace(finalText.String()), tokensMap)
	return DialogueParsedLine{
		side:     side,
		text:     strings.TrimSpace(finalText.String()),
		tokens:   tokensMap,
		typedCnt: 0,
	}
}

func (di *MotifDialogue) parseAll(lines []string) []DialogueParsedLine {
	var result []DialogueParsedLine
	for _, line := range lines {
		parsedLine := di.parseLine(line)
		result = append(result, parsedLine)
	}
	return result
}

func (di *MotifDialogue) preprocessNames(lines []string) []string {
	result := make([]string, len(lines))
	nameRe := regexp.MustCompile(`(?i)<(displayname|name)=([^>]+)>`)
	for i, line := range lines {
		newLine := line
		for {
			loc := nameRe.FindStringSubmatchIndex(newLine)
			if loc == nil {
				break
			}
			fullMatch := newLine[loc[0]:loc[1]]
			paramType := strings.ToLower(newLine[loc[2]:loc[3]])
			redirectionValue := newLine[loc[4]:loc[5]]
			resolvedPn := di.dialogueRedirection(redirectionValue)
			replacementText := ""
			if resolvedPn != -1 {
				if paramType == "displayname" {
					replacementText = sys.chars[resolvedPn-1][0].gi().displayname
				} else {
					replacementText = sys.chars[resolvedPn-1][0].name
				}
			}
			newLine = strings.Replace(newLine, fullMatch, replacementText, 1)
		}
		result[i] = newLine
	}
	return result
}

func (di *MotifDialogue) getDialogueLines() ([]string, int, error) {
	pn := sys.dialogueForce
	if pn != 0 && (pn < 1 || pn > MaxSimul*2+MaxAttachedChar) {
		return nil, 0, fmt.Errorf("invalid player number: %v", pn)
	}
	if pn == 0 {
		var validPlayers []int
		for i, p := range sys.chars {
			if len(p) > 0 && len(p[0].dialogue) > 0 {
				validPlayers = append(validPlayers, i+1)
			}
		}
		if len(validPlayers) > 0 {
			idx := int(RandI(0, int32(len(validPlayers)-1)))
			pn = validPlayers[idx]
		}
	}
	lines := []string{}
	if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
		for _, line := range sys.chars[pn-1][0].dialogue {
			lines = append(lines, line)
		}
	}
	return lines, pn, nil
}

// reset re-initializes certain state and animations.
func (di *MotifDialogue) reset(m *Motif) {
	di.active = false
	di.initialized = false
	di.counter = 0
	di.textNum = 0
	di.wait = 0
	di.lineFullyRendered = false
	di.charDelayCounter = 0
	di.switchCounter = 0
	di.endCounter = 0
	di.activeSide = -1
	di.talkingSide = 0
	di.parsed = nil

	m.DialogueInfo.P1.Bg.AnimData.Reset()
	m.DialogueInfo.P2.Bg.AnimData.Reset()
	m.DialogueInfo.P1.Face.AnimData.Reset()
	m.DialogueInfo.P2.Face.AnimData.Reset()
	m.DialogueInfo.P1.Face.Active.AnimData.Reset()
	m.DialogueInfo.P2.Face.Active.AnimData.Reset()
	m.DialogueInfo.P1.Active.AnimData.Reset()
	m.DialogueInfo.P2.Active.AnimData.Reset()

	m.DialogueInfo.P1.Text.TextSpriteData.text = ""
	m.DialogueInfo.P2.Text.TextSpriteData.text = ""
	// Dialogue uses its own typewriter logic, so disable the internal TextSprite typing.
	m.DialogueInfo.P1.Text.TextSpriteData.textDelay = 0
	m.DialogueInfo.P2.Text.TextSpriteData.textDelay = 0

	// Reset cached face parameters so portraits are reapplied when a new dialogue starts.
	for i := range di.faceParams {
		di.faceParams[i] = FaceParams{
			grp: -1,
			idx: -1,
			pn:  -1,
		}
	}

	//if !sys.skipMotifScaling() {
	//	sys.applyFightAspect()
	//}
}

func (di *MotifDialogue) clear(m *Motif) {
	for _, p := range sys.chars {
		if len(p) > 0 {
			p[0].dialogue = nil
		}
	}
	di.initialized = false
	sys.dialogueForce = 0
	sys.dialogueBarsFlg = false
	m.DialogueInfo.P1.Face.AnimData.anim = nil
	m.DialogueInfo.P2.Face.AnimData.anim = nil
	if m.DialogueInfo.P1.Face.Active.AnimData != nil {
		m.DialogueInfo.P1.Face.Active.AnimData.anim = nil
	}
	if m.DialogueInfo.P2.Face.Active.AnimData != nil {
		m.DialogueInfo.P2.Face.Active.AnimData.anim = nil
	}
	if !sys.skipMotifScaling() {
		sys.applyFightAspect()
	}
}

func (di *MotifDialogue) initDefaults(m *Motif) {
	if di.char == nil {
		return
	}

	selfPn := di.dialogueRedirection("self")
	enemyPn := di.dialogueRedirection("enemy")

	setSide := func(side int, pn int) {
		if pn <= 0 || !di.isValidPlayerNo(pn) {
			return
		}

		// Default face (same as pXface); "side" must be 1 or 2.
		var faceCfg *struct {
			AnimationProperties
			Active AnimationProperties `ini:"active"`
		}
		if side == 1 {
			faceCfg = &m.DialogueInfo.P1.Face
		} else if side == 2 {
			faceCfg = &m.DialogueInfo.P2.Face
		} else {
			return
		}

		// For defaults, prefer Anim over Spr. If Anim < 0, fall back to Spr.
		grp, idx := -1, -1
		if faceCfg.Anim >= 0 {
			grp = int(faceCfg.Anim)
			idx = -1 // treat as animation number
		} else if faceCfg.Spr[0] >= 0 {
			grp = int(faceCfg.Spr[0])
			idx = int(faceCfg.Spr[1])
		}

		if grp >= 0 {
			if anim := di.setFace(m, side, pn, grp, idx); anim != nil {
				if side == 1 {
					m.DialogueInfo.P1.Face.AnimData = anim
				} else if side == 2 {
					m.DialogueInfo.P2.Face.AnimData = anim
				}
				di.faceParams[side-1] = FaceParams{grp: grp, idx: idx, pn: pn}
			}
		}

		// Default active face (same as <pXfaceactive>)
		activeCfg := &faceCfg.Active
		grp, idx = -1, -1
		if activeCfg.Anim >= 0 {
			grp = int(activeCfg.Anim)
			idx = -1
		} else if activeCfg.Spr[0] >= 0 {
			grp = int(activeCfg.Spr[0])
			idx = int(activeCfg.Spr[1])
		}

		if grp >= 0 {
			if anim := di.setFaceActive(m, side, pn, grp, idx); anim != nil {
				if side == 1 {
					m.DialogueInfo.P1.Face.Active.AnimData = anim
				} else if side == 2 {
					m.DialogueInfo.P2.Face.Active.AnimData = anim
				}
			}
		}

		// Default name (equivalent to <pXname=...>)
		name := sys.chars[pn-1][0].gi().displayname
		if side == 1 {
			m.DialogueInfo.P1.Name.TextSpriteData.text = name
		} else if side == 2 {
			m.DialogueInfo.P2.Name.TextSpriteData.text = name
		}
	}

	setSide(1, selfPn)
	setSide(2, enemyPn)
}

func (di *MotifDialogue) init(m *Motif) {
	if !m.DialogueInfo.Enabled || !di.enabled {
		di.initialized = true
		return
	}

	di.reset(m)
	if !sys.skipMotifScaling() {
		sys.setGameSize(sys.scrrect[2], sys.scrrect[3])
	}

	lines, pn, _ := di.getDialogueLines()
	di.char = sys.chars[pn-1][0]

	lines = di.preprocessNames(lines)
	di.parsed = di.parseAll(lines)

	di.initDefaults(m)

	/*for i, line := range di.parsed {
		fmt.Printf("\nLine %d, side=%d\nText: %q\nTokens:\n", i+1, line.side, line.text)
		for textPos, tokens := range line.tokens {
			for _, t := range tokens {
				fmt.Printf("  atPos=%d  -> Param=%q Side=%d Redir=%q Pn=%d Value=%v\n",
					textPos, t.param, t.side, t.redirection, t.pn, t.value)
			}
		}
	}*/

	di.active = true
	di.initialized = true
}

// applyTokens checks and applies tokens at the current typed length in the text.
func (di *MotifDialogue) applyTokens(m *Motif, line *DialogueParsedLine) {
	typedLen := int(line.typedCnt)
	runeCount := utf8.RuneCountInString(line.text)
	if typedLen > runeCount {
		typedLen = runeCount
	}

	for i := 0; i <= typedLen; i++ {
		if tokenList, exists := line.tokens[i]; exists && len(tokenList) > 0 {
			for idx := len(tokenList) - 1; idx >= 0; idx-- {
				token := tokenList[idx]
				applied := di.applyToken(m, line, token, i)
				if applied {
					// remove token
					tokenList = SliceDelete(tokenList, idx)
				}
			}
			line.tokens[i] = tokenList
		}
	}
}

func (di *MotifDialogue) resetPortraitIdle(a *Anim) {
	if a == nil || a.anim == nil {
		return
	}
	a.anim.Reset()
}

// buildFaceAnim is a shared helper used by setFace / setFaceActive to build
// a portrait animation from either a character anim number or a raw sprite.
func (di *MotifDialogue) buildFaceAnim(m *Motif, side, pn, grp, idx int, faceCfg *AnimationProperties) *Anim {
	if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
		return nil
	}
	if side != 1 && side != 2 {
		return nil
	}
	c := sys.chars[pn-1][0]
	sc := sys.sel.GetChar(int(c.selectNo))

	a := NewAnim(nil, "")
	var ok bool
	if grp >= 0 && idx == -1 {
		// Treat grp as an animation number from the character.
		if anim := c.gi().animTable.get(int32(grp)); anim != nil {
			a.anim = anim
			ok = true
		}
	} else if grp >= 0 && idx >= 0 {
		// Treat grp/idx as sprite group/index from the character SFF.
		if anim, _ := tryGetPortrait(sc, c, [][2]int32{{int32(grp), int32(idx)}}); anim != nil {
			a.anim = anim
			ok = true
		}
	}

	if !ok || a.anim == nil {
		return nil
	}

	// Use the same palfx as the character.
	a.palfx = c.getPalfx()

	// Localcoord from motif definition (if present).
	lc := faceCfg.Localcoord
	if lc[0] > 0 && lc[1] > 0 {
		a.SetLocalcoord(float32(lc[0]), float32(lc[1]))
	}

	a.layerno = faceCfg.Layerno

	// Clamp to window if configured.
	w := faceCfg.Window
	if w[2] > w[0] && w[3] > w[1] {
		a.SetWindow([4]float32{float32(w[0]), float32(w[1]), float32(w[2]), float32(w[3])})
	}

	// Position.
	a.SetPos(faceCfg.Offset[0], faceCfg.Offset[1])

	// Scale – reuse the same logic as the victory/hiscore screens so portraits
	// respect character localcoord and portraitscale when SelectChar data exists.
	sx, sy := faceCfg.Scale[0], faceCfg.Scale[1]
	if sc != nil && sc.localcoord[0] != 0 {
		sx = faceCfg.Scale[0] * sc.portraitscale * float32(sys.motif.Info.Localcoord[0]) / sc.localcoord[0]
		sy = faceCfg.Scale[1] * sc.portraitscale * float32(sys.motif.Info.Localcoord[0]) / sc.localcoord[0]
	}
	a.SetScale(sx, sy)

	a.facing = float32(faceCfg.Facing)
	a.xshear = faceCfg.Xshear
	a.rot.angle = faceCfg.Angle
	a.rot.xangle = faceCfg.XAngle
	a.rot.yangle = faceCfg.YAngle
	a.projection = int32(Projection_Orthographic) // Default
	v := strings.ToLower(strings.TrimSpace(faceCfg.Projection))
	switch v {
	case "perspective":
		a.projection = int32(Projection_Perspective)
	case "perspective2":
		a.projection = int32(Projection_Perspective2)
	case "orthographic":
		// default, we don't need to do nothing
	default:
		if IsInt(v) {
			a.projection = Atoi(v)
		}
	}
	a.fLength = faceCfg.Focallength
	a.Update(false)

	return a
}

// setFaceActive changes the active face anim for the given side. This is used for the "talking" variant of the portrait.
func (di *MotifDialogue) setFaceActive(m *Motif, side, pn, grp, idx int) *Anim {
	var faceCfg *AnimationProperties
	switch side {
	case 1:
		faceCfg = &m.DialogueInfo.P1.Face.Active
	case 2:
		faceCfg = &m.DialogueInfo.P2.Face.Active
	default:
		return nil
	}
	return di.buildFaceAnim(m, side, pn, grp, idx, faceCfg)
}

// setFace changes the idle/normal face anim for the given side.
func (di *MotifDialogue) setFace(m *Motif, side, pn, grp, idx int) *Anim {
	var faceCfg *AnimationProperties
	switch side {
	case 1:
		faceCfg = &m.DialogueInfo.P1.Face.AnimationProperties
	case 2:
		faceCfg = &m.DialogueInfo.P2.Face.AnimationProperties
	default:
		return nil
	}
	return di.buildFaceAnim(m, side, pn, grp, idx, faceCfg)
}

// applyToken handles the application of a single DialogueToken.
// Returns true if the token should be removed after application.
func (di *MotifDialogue) applyToken(m *Motif, line *DialogueParsedLine, token DialogueToken, index int) bool {
	switch token.param {
	case "clear":
		line.text = ""
	case "wait":
		if len(token.value) > 0 {
			if waitFrames, ok := token.value[0].(float32); ok {
				di.wait = int(waitFrames)
			}
		}
		return true
	case "face":
		if token.side == 1 || token.side == 2 {
			if len(token.value) >= 1 {
				if v1, ok := token.value[0].(float32); ok {
					grp := int(v1)
					idx := -1
					if len(token.value) >= 2 {
						if v2, ok := token.value[1].(float32); ok {
							idx = int(v2)
						}
					}
					cur := &di.faceParams[token.side-1]
					if cur.pn != token.pn || cur.grp != grp || cur.idx != idx {
						if anim := di.setFace(m, token.side, token.pn, grp, idx); anim != nil {
							if token.side == 1 {
								m.DialogueInfo.P1.Face.AnimData = anim
							} else if token.side == 2 {
								m.DialogueInfo.P2.Face.AnimData = anim
							}
							cur.pn = token.pn
							cur.grp = grp
							cur.idx = idx
						}
					}
				}
			}
		}
		return true
	case "faceactive":
		if token.side == 1 || token.side == 2 {
			if len(token.value) >= 1 {
				if v1, ok := token.value[0].(float32); ok {
					grp := int(v1)
					idx := -1
					if len(token.value) >= 2 {
						if v2, ok := token.value[1].(float32); ok {
							idx = int(v2)
						}
					}
					if anim := di.setFaceActive(m, token.side, token.pn, grp, idx); anim != nil {
						if token.side == 1 {
							m.DialogueInfo.P1.Face.Active.AnimData = anim
						} else if token.side == 2 {
							m.DialogueInfo.P2.Face.Active.AnimData = anim
						}
					}
				}
			}
		}
		return true
	case "name":
		//fmt.Printf("[Dialogue] applyToken name: pn=%d side=%d redir=%q value=%v\n", token.pn, token.side, token.redirection, token.value)
		// 1) If we have a valid player number, use that character's display name.
		if token.pn != -1 && di.isValidPlayerNo(token.pn) {
			name := sys.chars[token.pn-1][0].gi().displayname
			if token.side == 1 {
				m.DialogueInfo.P1.Name.TextSpriteData.text = name
			} else if token.side == 2 {
				m.DialogueInfo.P2.Name.TextSpriteData.text = name
			}
			return true
		}

		// 2) Otherwise, try to use an explicit string value from the token.
		var name string
		if len(token.value) > 0 {
			if v, ok := token.value[0].(string); ok {
				name = v
			}
		}

		// 3) For tags like <p1name=name> parseTag puts the text into redirection and leaves value empty, so fall back to redirection if needed.
		if name == "" {
			name = strings.TrimSpace(token.redirection)
		}
		if name == "" {
			return true
		}
		if token.side == 1 {
			m.DialogueInfo.P1.Name.TextSpriteData.text = name
		} else if token.side == 2 {
			m.DialogueInfo.P2.Name.TextSpriteData.text = name
		}
		return true
	case "sound":
		if len(token.value) >= 2 {
			if !di.isValidPlayerNo(token.pn) {
				return true
			}
			f, lw, lp, stopgh, stopcs := false, false, false, false, false
			var g, n, ch, vo, priority, lc int32 = -1, 0, -1, 100, 0, 0
			var loopstart, loopend, startposition int = 0, 0, 0
			var p, fr float32 = 0, 1
			x := &sys.chars[token.pn-1][0].pos[0]
			ls := sys.chars[token.pn-1][0].localscl
			prefix := ""
			if f {
				prefix = "f"
			}
			if v1, ok1 := token.value[0].(float32); ok1 {
				g = int32(v1)
				if v2, ok2 := token.value[1].(float32); ok2 {
					n = int32(v2)
					if len(token.value) >= 3 {
						if v3, ok3 := token.value[2].(float32); ok3 {
							vo = int32(v3)
						}
					}
				}
			}
			if lc == 0 {
				if lp {
					sys.chars[token.pn-1][0].playSound(prefix, lw, -1, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
				} else {
					sys.chars[token.pn-1][0].playSound(prefix, lw, 0, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
				}
				// Otherwise, read the loopcount parameter directly
			} else {
				sys.chars[token.pn-1][0].playSound(prefix, lw, lc, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
			}
		}
		return true
	case "anim":
		if len(token.value) >= 1 {
			if !di.isValidPlayerNo(token.pn) {
				return true
			}
			if v, ok := token.value[0].(float32); ok {
				animNo := int32(v)
				if sys.chars[token.pn-1][0].selfAnimExist(BytecodeInt(animNo)) == BytecodeBool(true) {
					sys.chars[token.pn-1][0].changeAnim(animNo, token.pn-1, -1, "")
				}
			}
		}
		return true
	case "state":
		if len(token.value) >= 1 {
			if !di.isValidPlayerNo(token.pn) {
				return true
			}
			if v, ok := token.value[0].(float32); ok {
				stateNo := int32(v)
				if stateNo == -1 {
					for _, ch := range sys.chars[token.pn-1] {
						ch.setSCF(SCF_disabled)
					}
				} else if sys.chars[token.pn-1][0].selfStatenoExist(BytecodeInt(stateNo)) == BytecodeBool(true) {
					for _, ch := range sys.chars[token.pn-1] {
						if ch.scf(SCF_disabled) {
							ch.unsetSCF(SCF_disabled)
						}
					}
					sys.chars[token.pn-1][0].changeState(int32(stateNo), -1, -1, "")
				}
			}
			return true
		}
		return true
	case "map":
		if len(token.value) >= 2 {
			if !di.isValidPlayerNo(token.pn) {
				return true
			}
			mapName, ok1 := token.value[0].(string)
			mapVal, ok2 := token.value[1].(float32)
			if !ok1 || !ok2 {
				return false
			}
			mapOp := int32(0)
			if len(token.value) >= 3 {
				if op, ok3 := token.value[2].(string); ok3 && op == "add" {
					mapOp = 1
				}
			}
			sys.chars[token.pn-1][0].mapSet(mapName, mapVal, mapOp)
		}
		return true
	default:
		// Unrecognized token parameter.
	}
	return false
}

// step processes dialogue state each frame, handling timing, skipping, cancel, wrapping, etc.
func (di *MotifDialogue) step(m *Motif) {
	// If we have no lines, do nothing
	if len(di.parsed) == 0 {
		di.active = false
		di.clear(m)
		return
	}

	// If user presses "cancel", end the dialogue
	if sys.button(m.DialogueInfo.Cancel.Key, -1) {
		di.active = false
		di.clear(m)
		return
	}

	// Update background animations
	m.DialogueInfo.P1.Bg.AnimData.Update(false)
	m.DialogueInfo.P2.Bg.AnimData.Update(false)

	// Check if we haven't reached StartTime yet
	if di.counter < m.DialogueInfo.StartTime {
		// Before dialogue starts, keep both portraits in a neutral pose.
		di.resetPortraitIdle(m.DialogueInfo.P1.Face.AnimData)
		di.resetPortraitIdle(m.DialogueInfo.P2.Face.AnimData)
		di.counter++
		return
	}

	// Check if we've gone past all lines
	if di.textNum >= len(di.parsed) {
		// If we haven't started EndCounter yet, do so
		if di.endCounter == 0 {
			di.endCounter = int(m.DialogueInfo.EndTime)
		} else {
			di.endCounter--
			if di.endCounter <= 0 {
				// No more dialogue: make sure both portraits end in a neutral pose.
				di.resetPortraitIdle(m.DialogueInfo.P1.Face.AnimData)
				di.resetPortraitIdle(m.DialogueInfo.P2.Face.AnimData)
				// Done
				di.active = false
				di.clear(m)
				return
			}
		}
		di.counter++
		return
	}

	// We have a valid line to render
	currentLine := &di.parsed[di.textNum]
	//fmt.Printf("[Dialogue] step: textNum=%d side=%d typedCnt=%d text=%q\n", di.textNum, currentLine.side, currentLine.typedCnt, currentLine.text)
	di.activeSide = currentLine.side
	prevLineFullyRendered := di.lineFullyRendered

	// Handle "skip" key (only after SkipTime)
	if di.counter >= m.DialogueInfo.SkipTime {
		if sys.button(m.DialogueInfo.Skip.Key, -1) {
			if !di.lineFullyRendered {
				currentLine.typedCnt = utf8.RuneCountInString(currentLine.text)
				di.lineFullyRendered = true
				di.switchCounter = 0
				di.wait = 0
			} else {
				// If line is already fully rendered => move to next line
				di.advanceLine(m)
				return
			}
		}
	}

	// Determine the per-character delay for this side
	var charDelay float32
	if currentLine.side == 1 {
		charDelay = float32(m.DialogueInfo.P1.Text.TextDelay)
	} else if currentLine.side == 2 {
		charDelay = float32(m.DialogueInfo.P2.Text.TextDelay)
	} else {
		charDelay = 1
	}

	// Handle any explicit token-based wait
	if di.wait > 0 {
		di.wait--
	} else if !di.lineFullyRendered {
		// Otherwise, reveal letters one by one
		StepTypewriter(
			currentLine.text,
			&currentLine.typedCnt,
			&di.charDelayCounter,
			&di.lineFullyRendered,
			charDelay,
		)
	}

	// Apply any tokens for newly revealed characters
	di.applyTokens(m, currentLine)

	// Clamp typedLen so it doesn't exceed the line length
	typedLen := currentLine.typedCnt
	runeCount := utf8.RuneCountInString(currentLine.text)
	if typedLen > runeCount {
		typedLen = runeCount
	}

	// If we've just finished the line
	if typedLen >= runeCount && !prevLineFullyRendered {
		di.lineFullyRendered = true // StepTypewriter already set this, but keep explicit.
		di.switchCounter = 0
	}

	// If line is fully rendered, handle auto-switch after SwitchTime
	if di.lineFullyRendered {
		di.switchCounter++
		if di.switchCounter >= int(m.DialogueInfo.SwitchTime) {
			di.advanceLine(m)
			return
		}
	}

	if currentLine.side == 1 {
		m.DialogueInfo.P1.Text.TextSpriteData.wrapText(currentLine.text, typedLen)
		m.DialogueInfo.P1.Text.TextSpriteData.Update()
	} else if currentLine.side == 2 {
		m.DialogueInfo.P2.Text.TextSpriteData.wrapText(currentLine.text, typedLen)
		m.DialogueInfo.P2.Text.TextSpriteData.Update()
	}

	// Decide who is "talking" this frame for portrait animation:
	// - must be the current line's side
	// - only while characters are still being revealed (line not fully rendered)
	// - and no explicit <wait> is active.
	prevTalkingSide := di.talkingSide
	talkingSide := 0
	if di.wait == 0 && !di.lineFullyRendered {
		if currentLine.side == 1 || currentLine.side == 2 {
			talkingSide = currentLine.side
		}
	}
	di.talkingSide = talkingSide

	// Update the active highlight for the current active side.
	if di.activeSide == 1 {
		m.DialogueInfo.P1.Active.AnimData.Update(false)
	} else if di.activeSide == 2 {
		m.DialogueInfo.P2.Active.AnimData.Update(false)
	}

	// If a side just changed talking state, reset its face anims so the next
	// variant (idle or active) starts from the first frame.
	if prevTalkingSide != di.talkingSide {
		// P1 changed state?
		if prevTalkingSide == 1 || di.talkingSide == 1 {
			if m.DialogueInfo.P1.Face.AnimData != nil {
				m.DialogueInfo.P1.Face.AnimData.Reset()
			}
			if m.DialogueInfo.P1.Face.Active.AnimData != nil &&
				m.DialogueInfo.P1.Face.Active.AnimData.anim != nil &&
				m.DialogueInfo.P1.Face.Active.AnimData != m.DialogueInfo.P1.Face.AnimData {
				m.DialogueInfo.P1.Face.Active.AnimData.Reset()
			}
		}
		// P2 changed state?
		if prevTalkingSide == 2 || di.talkingSide == 2 {
			if m.DialogueInfo.P2.Face.AnimData != nil {
				m.DialogueInfo.P2.Face.AnimData.Reset()
			}
			if m.DialogueInfo.P2.Face.Active.AnimData != nil &&
				m.DialogueInfo.P2.Face.Active.AnimData.anim != nil &&
				m.DialogueInfo.P2.Face.Active.AnimData != m.DialogueInfo.P2.Face.AnimData {
				m.DialogueInfo.P2.Face.Active.AnimData.Reset()
			}
		}
	}

	// Step portraits:
	// - If side has an active variant and is talking: step active variant.
	// - Otherwise: always step standard face so idle portraits keep animating.
	// P1
	p1Idle := m.DialogueInfo.P1.Face.AnimData
	var p1Active *Anim
	if m.DialogueInfo.P1.Face.Active.AnimData != nil && m.DialogueInfo.P1.Face.Active.AnimData.anim != nil {
		p1Active = m.DialogueInfo.P1.Face.Active.AnimData
	}
	if di.talkingSide == 1 && p1Active != nil {
		p1Active.Update(false)
	} else if p1Idle != nil {
		p1Idle.Update(false)
	}
	// P2
	p2Idle := m.DialogueInfo.P2.Face.AnimData
	var p2Active *Anim
	if m.DialogueInfo.P2.Face.Active.AnimData != nil && m.DialogueInfo.P2.Face.Active.AnimData.anim != nil {
		p2Active = m.DialogueInfo.P2.Face.Active.AnimData
	}
	if di.talkingSide == 2 && p2Active != nil {
		p2Active.Update(false)
	} else if p2Idle != nil {
		p2Idle.Update(false)
	}

	// Finally increment the global frame counter
	di.counter++
}

// advanceLine moves to the next line, clearing or preserving text depending on side.
func (di *MotifDialogue) advanceLine(m *Motif) {
	// Clear text if next line uses the same side; preserve if different side
	currentSide := -1
	if di.textNum < len(di.parsed) {
		currentSide = di.parsed[di.textNum].side
	}

	di.textNum++
	if di.textNum < len(di.parsed) {
		nextSide := di.parsed[di.textNum].side
		if nextSide == currentSide {
			// Same side => replace text with the new line, so clear now
			if currentSide == 1 {
				m.DialogueInfo.P1.Text.TextSpriteData.text = ""
			} else if currentSide == 2 {
				m.DialogueInfo.P2.Text.TextSpriteData.text = ""
			}
		}
	} else {
		// If we're out of lines, text is presumably done
	}

	// Reset state
	di.lineFullyRendered = false
	di.switchCounter = 0
	di.wait = 0
	di.charDelayCounter = 0
}

// draw renders the dialogue on the screen based on the current state.
func (di *MotifDialogue) draw(m *Motif, layerno int16) {
	// BG
	m.DialogueInfo.P1.Bg.AnimData.Draw(layerno)
	m.DialogueInfo.P2.Bg.AnimData.Draw(layerno)

	// If we haven't reached StartTime yet, or no lines, skip drawing text
	if di.counter < m.DialogueInfo.StartTime || len(di.parsed) == 0 {
		return
	}

	// Faces
	// - If a side is actively talking and has an active variant, draw the active one.
	// - Otherwise draw the normal face.
	// - If active variant is missing, fall back to the normal face for both states.
	// P1
	p1Idle := m.DialogueInfo.P1.Face.AnimData
	var p1Active *Anim
	if m.DialogueInfo.P1.Face.Active.AnimData != nil && m.DialogueInfo.P1.Face.Active.AnimData.anim != nil {
		p1Active = m.DialogueInfo.P1.Face.Active.AnimData
	} else {
		p1Active = p1Idle
	}
	if di.talkingSide == 1 && p1Active != nil {
		p1Active.Draw(layerno)
	} else if p1Idle != nil {
		p1Idle.Draw(layerno)
	}
	// P2
	p2Idle := m.DialogueInfo.P2.Face.AnimData
	var p2Active *Anim
	if m.DialogueInfo.P2.Face.Active.AnimData != nil && m.DialogueInfo.P2.Face.Active.AnimData.anim != nil {
		p2Active = m.DialogueInfo.P2.Face.Active.AnimData
	} else {
		p2Active = p2Idle
	}
	if di.talkingSide == 2 && p2Active != nil {
		p2Active.Draw(layerno)
	} else if p2Idle != nil {
		p2Idle.Draw(layerno)
	}

	// Names
	m.DialogueInfo.P1.Name.TextSpriteData.Draw(layerno)
	m.DialogueInfo.P2.Name.TextSpriteData.Draw(layerno)

	// Text
	m.DialogueInfo.P1.Text.TextSpriteData.Draw(layerno)
	m.DialogueInfo.P2.Text.TextSpriteData.Draw(layerno)

	// Active anim highlight
	if di.activeSide == 1 {
		m.DialogueInfo.P1.Active.AnimData.Draw(layerno)
	} else if di.activeSide == 2 {
		m.DialogueInfo.P2.Active.AnimData.Draw(layerno)
	}
}

type MotifHiscore struct {
	enabled     bool
	active      bool
	initialized bool
	counter     int32
	endTimer    int32
	place       int32
	mode        string
	rows        []rankingRow
	statsRaw    string
	endTime     int32
	noFade      bool
	noBgs       bool
	noOverlay   bool

	// Active-row blink state (Rank / Result / Name)
	rankActiveCount   int32
	resultActiveCount int32
	nameActiveCount   int32
	rankUseActive2    bool
	resultUseActive2  bool
	nameUseActive2    bool

	// Name entry & timer
	input       bool  // true while entering name
	letters     []int // 1-based indices into glyph list
	timerCount  int32 // remaining displayed timer count
	timerFrames int32 // frames left until next decrement
	haveSaved   bool  // true once we've persisted the entered name
}

// rankingRow is a minimal container built from dynamic JSON (no static decode).
type rankingRow struct {
	score int
	time  float64
	win   int
	name  string
	pals  []int32
	chars []string
	bgs   []*Anim
	faces []*Anim

	rankData          *TextSprite
	resultData        *TextSprite
	nameData          *TextSprite
	rankDataActive    *TextSprite
	rankDataActive2   *TextSprite
	resultDataActive  *TextSprite
	resultDataActive2 *TextSprite
	nameDataActive    *TextSprite
	nameDataActive2   *TextSprite
}

func (hi *MotifHiscore) reset(m *Motif) {
	hi.active = false
	hi.initialized = false
	hi.endTimer = -1
	hi.counter = 0
	hi.place = 0
	hi.mode = ""
	hi.endTime = 0
	hi.noFade = false
	hi.noBgs = false
	hi.noOverlay = false
	hi.rankActiveCount, hi.resultActiveCount, hi.nameActiveCount = 0, 0, 0
	hi.rankUseActive2, hi.resultUseActive2, hi.nameUseActive2 = false, false, false
	hi.rows = nil
	hi.input = false
	hi.letters = nil
	hi.timerCount = 0
	hi.timerFrames = 0
	hi.haveSaved = false
}

func (hi *MotifHiscore) makeRowTextSprite(tpl *TextSprite, x, y float32, text string) *TextSprite {
	if tpl == nil {
		return nil
	}
	ts := tpl.Copy()
	if ts == nil {
		return nil
	}
	ts.SetPos(x, y)
	ts.text = text
	return ts
}

func (hi *MotifHiscore) init(m *Motif, mode string, place, endTime int32, noFade, noBgs, noOverlay bool) {
	//if !m.HiscoreInfo.Enabled || !hi.enabled {
	//	hi.initialized = true
	//	return
	//}

	if err := sys.luaLState.DoString("hook.run('game.hiscore_init')"); err != nil {
		sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.hiscore_init", err.Error())
	}

	dataType, _ := m.HiscoreInfo.Ranking[mode]

	hi.reset(m)
	hi.place = place
	hi.mode = mode
	hi.endTime = m.HiscoreInfo.Time
	if endTime > 0 {
		hi.endTime = endTime
	}
	hi.noFade = noFade
	hi.noBgs = noBgs
	hi.noOverlay = noOverlay
	hi.input = (place > 0)
	if hi.input {
		// Start with one letter selected
		hi.letters = []int{1}
	}

	// Timer setup (only matters during input)
	hi.timerCount = m.HiscoreInfo.Timer.Count
	hi.timerFrames = m.HiscoreInfo.Timer.Framespercount

	// Parse and cache rows (read from JSON file)
	hi.rows = parseRankingRows(sys.cmdFlags["-stats"], mode)

	// If a VisibleItems cap is configured, ignore any extra entries in stats.json.
	capVisible := int(m.HiscoreInfo.Window.VisibleItems)
	if capVisible > 0 && capVisible < len(hi.rows) {
		hi.rows = hi.rows[:capVisible]
	}

	// Build portraits/text only for the rows we will actually show.
	visible := len(hi.rows)

	m.HiscoreInfo.Item.Face.AnimData.Reset()
	m.HiscoreInfo.Item.Face.Bg.AnimData.Reset()
	m.HiscoreInfo.Item.Face.Unknown.AnimData.Reset()
	baseX, baseY := m.HiscoreInfo.Pos[0], m.HiscoreInfo.Pos[1]
	itemOffX, itemOffY := m.HiscoreInfo.Item.Offset[0], m.HiscoreInfo.Item.Offset[1]
	for i := 0; i < visible; i++ {
		rowDefs := hi.rows[i].chars
		rowBgs := make([]*Anim, 0, len(rowDefs))
		rowFaces := make([]*Anim, 0, len(rowDefs))
		limit := int(m.HiscoreInfo.Item.Face.Num)
		if limit <= 0 {
			limit = math.MaxInt32
		}
		for j, def := range rowDefs {
			if j >= limit {
				break
			}
			// Compute the on-screen position for this face/background once.
			x := baseX + itemOffX + m.HiscoreInfo.Item.Face.Offset[0] +
				float32(i)*m.HiscoreInfo.Item.Spacing[0] +
				float32(j)*m.HiscoreInfo.Item.Face.Spacing[0]
			y := baseY + itemOffY + m.HiscoreInfo.Item.Face.Offset[1] +
				float32(i)*(m.HiscoreInfo.Item.Spacing[1]+m.HiscoreInfo.Item.Face.Spacing[1])

			// Background anim per face (clone so each slot has its own pos/state)
			if m.HiscoreInfo.Item.Face.Bg.AnimData != nil {
				bg := m.HiscoreInfo.Item.Face.Bg.AnimData.Copy()
				if bg != nil {
					bg.SetPos(x+m.HiscoreInfo.Item.Face.Bg.Offset[0], y+m.HiscoreInfo.Item.Face.Bg.Offset[1])
				}
				rowBgs = append(rowBgs, bg)
			} else {
				rowBgs = append(rowBgs, nil)
			}

			// Face anim (character or unknown)
			if a := hiscorePortraitAnim(def, m, x, y); a != nil {
				rowFaces = append(rowFaces, a)
			} else if m.HiscoreInfo.Item.Face.Unknown.AnimData != nil {
				// Use a cloned "unknown" placeholder per slot so we can position/update independently.
				ua := m.HiscoreInfo.Item.Face.Unknown.AnimData.Copy()
				if ua != nil {
					ua.SetPos(x+m.HiscoreInfo.Item.Face.Unknown.Offset[0], y+m.HiscoreInfo.Item.Face.Unknown.Offset[1])
				}
				rowFaces = append(rowFaces, ua)
			} else {
				rowFaces = append(rowFaces, nil)
			}

			//
		}
		hi.rows[i].bgs = rowBgs
		hi.rows[i].faces = rowFaces
	}

	// Build per-row TextSprites (Rank / Result / Name)
	for i := 0; i < visible; i++ {
		row := &hi.rows[i]
		// Rank
		if m.HiscoreInfo.Item.Rank.TextSpriteData != nil {
			ts := m.HiscoreInfo.Item.Rank.TextSpriteData.Copy()
			if ts != nil {
				rowXBase := baseX + itemOffX + float32(i)*(m.HiscoreInfo.Item.Spacing[0]+m.HiscoreInfo.Item.Rank.Spacing[0])
				stepY := float32(math.Round(float64(
					(float32(ts.fnt.Size[1])+float32(ts.fnt.Spacing[1]))*ts.scaleInit[1] +
						(m.HiscoreInfo.Item.Spacing[1] + m.HiscoreInfo.Item.Rank.Spacing[1]),
				)))
				rowYBase := baseY + itemOffY + stepY*float32(i)
				ts.SetPos(rowXBase+m.HiscoreInfo.Item.Rank.Offset[0], rowYBase+m.HiscoreInfo.Item.Rank.Offset[1])
				rankKey := Itoa(i + 1)
				fmtStr, ok := m.HiscoreInfo.Item.Rank.Text[rankKey]
				if !ok || fmtStr == "" {
					fmtStr = m.HiscoreInfo.Item.Rank.Text["default"]
				}
				fmtStr = m.replaceFormatSpecifiers(fmtStr)
				ts.text = fmt.Sprintf(fmtStr, i+1)
				if m.HiscoreInfo.Item.Rank.Uppercase {
					ts.text = strings.ToUpper(ts.text)
				}
				row.rankData = ts
				// If this is the highlighted row, prepare Active/Active2 from their own templates
				if hi.place > 0 && int(hi.place-1) == i {
					row.rankDataActive = hi.makeRowTextSprite(
						m.HiscoreInfo.Item.Rank.Active.TextSpriteData,
						rowXBase+m.HiscoreInfo.Item.Rank.Active.Offset[0],
						rowYBase+m.HiscoreInfo.Item.Rank.Active.Offset[1],
						ts.text,
					)
					row.rankDataActive2 = hi.makeRowTextSprite(
						m.HiscoreInfo.Item.Rank.Active2.TextSpriteData,
						rowXBase+m.HiscoreInfo.Item.Rank.Active2.Offset[0],
						rowYBase+m.HiscoreInfo.Item.Rank.Active2.Offset[1],
						ts.text,
					)
				}
			}
		}

		// Result
		if m.HiscoreInfo.Item.Result.TextSpriteData != nil {
			ts := m.HiscoreInfo.Item.Result.TextSpriteData.Copy()
			if ts != nil {
				rowXBase := baseX + itemOffX + float32(i)*(m.HiscoreInfo.Item.Spacing[0]+m.HiscoreInfo.Item.Result.Spacing[0])
				stepY := float32(math.Round(float64(
					(float32(ts.fnt.Size[1])+float32(ts.fnt.Spacing[1]))*ts.scaleInit[1] +
						(m.HiscoreInfo.Item.Spacing[1] + m.HiscoreInfo.Item.Result.Spacing[1]),
				)))
				rowYBase := baseY + itemOffY + stepY*float32(i)
				ts.SetPos(rowXBase+m.HiscoreInfo.Item.Result.Offset[0], rowYBase+m.HiscoreInfo.Item.Result.Offset[1])

				fmtStr := m.HiscoreInfo.Item.Result.Text[dataType]
				fmtStr = m.replaceFormatSpecifiers(fmtStr)
				switch dataType {
				case "score":
					// e.g. "%08d" -> zero-padded to width 8
					fmtStr = fmt.Sprintf(fmtStr, row.score)
				case "win":
					// e.g. "Round %d"
					fmtStr = fmt.Sprintf(fmtStr, row.win)
				case "time":
					// e.g. "%m'%s''%x"
					fmtStr = formatTimeText(fmtStr, row.time)
				}
				if m.HiscoreInfo.Item.Result.Uppercase {
					fmtStr = strings.ToUpper(fmtStr)
				}
				ts.text = fmtStr
				row.resultData = ts
				if hi.place > 0 && int(hi.place-1) == i {
					row.resultDataActive = hi.makeRowTextSprite(
						m.HiscoreInfo.Item.Result.Active.TextSpriteData,
						rowXBase+m.HiscoreInfo.Item.Result.Active.Offset[0],
						rowYBase+m.HiscoreInfo.Item.Result.Active.Offset[1],
						ts.text,
					)
					row.resultDataActive2 = hi.makeRowTextSprite(
						m.HiscoreInfo.Item.Result.Active2.TextSpriteData,
						rowXBase+m.HiscoreInfo.Item.Result.Active2.Offset[0],
						rowYBase+m.HiscoreInfo.Item.Result.Active2.Offset[1],
						ts.text,
					)
				}
			}
		}

		// Name
		if m.HiscoreInfo.Item.Name.TextSpriteData != nil {
			ts := m.HiscoreInfo.Item.Name.TextSpriteData.Copy()
			if ts != nil {
				rowXBase := baseX + itemOffX + float32(i)*(m.HiscoreInfo.Item.Spacing[0]+m.HiscoreInfo.Item.Name.Spacing[0])
				stepY := float32(math.Round(float64(
					(float32(ts.fnt.Size[1])+float32(ts.fnt.Spacing[1]))*ts.scaleInit[1] +
						(m.HiscoreInfo.Item.Spacing[1] + m.HiscoreInfo.Item.Name.Spacing[1]),
				)))
				rowYBase := baseY + itemOffY + stepY*float32(i)
				ts.SetPos(rowXBase+m.HiscoreInfo.Item.Name.Offset[0], rowYBase+m.HiscoreInfo.Item.Name.Offset[1])
				// If highlighted & input, start with glyph-based editable text; else use row.name from stats
				if hi.input && int(hi.place-1) == i {
					row.name = "" // TODO: this shouldn't be needed if we're appending new row
					name := buildNameFromLetters(m, hi.letters)
					fs := m.replaceFormatSpecifiers(m.HiscoreInfo.Item.Name.Text["default"])
					ts.text = fmt.Sprintf(fs, name)
				} else {
					fs := m.replaceFormatSpecifiers(m.HiscoreInfo.Item.Name.Text["default"])
					ts.text = fmt.Sprintf(fs, row.name)
				}
				if m.HiscoreInfo.Item.Name.Uppercase {
					ts.text = strings.ToUpper(ts.text)
				}
				row.nameData = ts
				if hi.place > 0 && int(hi.place-1) == i {
					row.nameDataActive = hi.makeRowTextSprite(
						m.HiscoreInfo.Item.Name.Active.TextSpriteData,
						rowXBase+m.HiscoreInfo.Item.Name.Active.Offset[0],
						rowYBase+m.HiscoreInfo.Item.Name.Active.Offset[1],
						ts.text,
					)
					row.nameDataActive2 = hi.makeRowTextSprite(
						m.HiscoreInfo.Item.Name.Active2.TextSpriteData,
						rowXBase+m.HiscoreInfo.Item.Name.Active2.Offset[0],
						rowYBase+m.HiscoreInfo.Item.Name.Active2.Offset[1],
						ts.text,
					)
				}
			}
		}
	}

	// Initialize timer TextSprite (position & first value)
	if m.HiscoreInfo.Timer.TextSpriteData != nil {
		m.HiscoreInfo.Timer.TextSpriteData.Reset()
		m.HiscoreInfo.Timer.TextSpriteData.AddPos(m.HiscoreInfo.Pos[0], m.HiscoreInfo.Pos[1])
		ts := m.HiscoreInfo.Timer.Text
		ts = m.replaceFormatSpecifiers(ts)
		if hi.input && hi.timerCount >= 0 && m.HiscoreInfo.Timer.Count != -1 {
			m.HiscoreInfo.Timer.TextSpriteData.text = fmt.Sprintf(ts, hi.timerCount)
		} else {
			// Leave text as-is if timer disabled
			m.HiscoreInfo.Timer.TextSpriteData.text = fmt.Sprintf(ts, 0)
		}
	}

	m.HiscoreBgDef.BGDef.Reset()
	if !hi.noFade {
		m.HiscoreInfo.FadeIn.FadeData.init(m.fadeIn, true)
	}

	m.HiscoreInfo.Title.TextSpriteData.Reset()
	m.HiscoreInfo.Title.TextSpriteData.AddPos(m.HiscoreInfo.Pos[0], m.HiscoreInfo.Pos[1])
	if v, ok := m.HiscoreInfo.Title.Text[mode]; ok {
		m.HiscoreInfo.Title.TextSpriteData.text = v
	}

	m.HiscoreInfo.Title.Rank.TextSpriteData.Reset()
	m.HiscoreInfo.Title.Rank.TextSpriteData.AddPos(m.HiscoreInfo.Pos[0], m.HiscoreInfo.Pos[1])

	m.HiscoreInfo.Title.Result.TextSpriteData.Reset()
	m.HiscoreInfo.Title.Result.TextSpriteData.AddPos(m.HiscoreInfo.Pos[0], m.HiscoreInfo.Pos[1])

	m.HiscoreInfo.Title.Name.TextSpriteData.Reset()
	m.HiscoreInfo.Title.Name.TextSpriteData.AddPos(m.HiscoreInfo.Pos[0], m.HiscoreInfo.Pos[1])

	m.HiscoreInfo.Title.Face.TextSpriteData.Reset()
	m.HiscoreInfo.Title.Face.TextSpriteData.AddPos(m.HiscoreInfo.Pos[0], m.HiscoreInfo.Pos[1])

	m.Music.Play("hiscore", sys.motif.Def)

	//hi.counter = 0
	hi.active = true
	hi.initialized = true
}

func (hi *MotifHiscore) step(m *Motif) {
	if hi.counter > 0 {
		if err := sys.luaLState.DoString("hook.run('game.hiscore')"); err != nil {
			sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.hiscore", err.Error())
		}
	}
	// Begin fade-out on cancel or when time elapses.
	if hi.endTimer == -1 {
		cancel := sys.esc || sys.button(m.HiscoreInfo.Cancel.Key, -1) ||
			(!hi.input && sys.button(m.HiscoreInfo.Done.Key, -1)) ||
			(!sys.gameRunning && sys.motif.AttractMode.Enabled && sys.credits > 0)
		if cancel || (!hi.input && hi.counter == hi.endTime) {
			if !hi.noFade {
				startFadeOut(m.HiscoreInfo.FadeOut.FadeData, m.fadeOut, cancel, m.fadePolicy)
			}
			hi.endTimer = hi.counter + m.fadeOut.timeRemaining
		}
	}

	hi.handleBlinkers(m)

	// Advance animations for per-face backgrounds and faces
	for i := range hi.rows {
		row := &hi.rows[i]
		for _, bg := range row.bgs {
			if bg != nil {
				bg.Update(false)
			}
		}
		for _, face := range row.faces {
			if face != nil {
				face.Update(false)
			}
		}
	}

	// Name input & timer (only for highlighted row)
	if hi.place > 0 {
		idx := int(hi.place - 1)
		if idx >= 0 && idx < len(hi.rows) {
			row := &hi.rows[idx]
			// Timer tick — only if enabled and input active
			if hi.input && m.HiscoreInfo.Timer.Count != -1 {
				if hi.timerFrames > 0 {
					hi.timerFrames--
				} else {
					if hi.timerCount > 0 {
						hi.timerCount--
					}
					hi.timerFrames = m.HiscoreInfo.Timer.Framespercount
				}
				// Update timer text each step
				if m.HiscoreInfo.Timer.TextSpriteData != nil {
					ts := m.HiscoreInfo.Timer.Text
					ts = m.replaceFormatSpecifiers(ts)
					m.HiscoreInfo.Timer.TextSpriteData.text = fmt.Sprintf(ts, hi.timerCount)
				}
				// When time runs out, auto-finish input
				if hi.timerCount <= 0 {
					m.Snd.play(m.HiscoreInfo.Done.Snd, 100, 0, 0, 0, 0)
					hi.input = false
					// Give a short tail
					hi.counter = hi.endTime - m.HiscoreInfo.Done.Time
					hi.finalizeAndSave()
				}
			}

			// Handle name glyph entry while input is active
			if hi.input {
				controller := -1 // TODO: proper controller
				maxLen := initialsWidth(m.HiscoreInfo.Item.Name.Text["default"])
				glyphCount := len(m.HiscoreInfo.Glyphs)
				// Previous glyph
				if sys.button(m.HiscoreInfo.Previous.Key, controller) {
					m.Snd.play(m.HiscoreInfo.Move.Snd, 100, 0, 0, 0, 0)
					if len(hi.letters) == 0 {
						hi.letters = []int{1}
					} else {
						last := len(hi.letters) - 1
						hi.letters[last]--
						if hi.letters[last] <= 0 {
							hi.letters[last] = glyphCount
						}
					}
					updateRowNameFromLetters(m, row, hi.letters)
					// Next glyph
				} else if sys.button(m.HiscoreInfo.Next.Key, controller) {
					m.Snd.play(m.HiscoreInfo.Move.Snd, 100, 0, 0, 0, 0)
					if len(hi.letters) == 0 {
						hi.letters = []int{1}
					} else {
						last := len(hi.letters) - 1
						hi.letters[last]++
						if hi.letters[last] > glyphCount {
							hi.letters[last] = 1
						}
					}
					updateRowNameFromLetters(m, row, hi.letters)
					// Confirm / Add / Backspace
				} else if sys.button(m.HiscoreInfo.Done.Key, controller) {
					// Current glyph meaning
					curGlyph := currentGlyph(m, hi.letters)
					if curGlyph == "<" {
						// Backspace
						m.Snd.play(m.HiscoreInfo.Cancel.Snd, 100, 0, 0, 0, 0)
						if len(hi.letters) > 1 {
							hi.letters = hi.letters[:len(hi.letters)-1]
						} else {
							hi.letters = []int{1}
						}
						updateRowNameFromLetters(m, row, hi.letters)
					} else if len(hi.letters) < maxLen {
						m.Snd.play(m.HiscoreInfo.Done.Snd, 100, 0, 0, 0, 0)
						lastIdx := hi.letters[len(hi.letters)-1]
						hi.letters = append(hi.letters, lastIdx)
						updateRowNameFromLetters(m, row, hi.letters)
					} else {
						// Finalize
						m.Snd.play(m.HiscoreInfo.Done.Snd, 100, 0, 0, 0, 0)
						hi.input = false
						hi.counter = hi.endTime - m.HiscoreInfo.Done.Time
						hi.finalizeAndSave()
					}
				}
			}
		}
	}

	// Finish after fade-out completes
	if hi.endTimer != -1 && hi.counter >= hi.endTimer {
		if m.fadeOut != nil && !hi.noFade {
			m.fadeOut.reset()
		}
		hi.reset(m)
		//hi.active = false
		return
	}

	hi.counter++
}

func (hi *MotifHiscore) draw(m *Motif, layerno int16) {
	// Background
	if !hi.noBgs {
		if m.HiscoreBgDef.BgClearColor[0] >= 0 {
			m.HiscoreBgDef.RectData.Draw(layerno)
		}
		if layerno <= 1 {
			m.HiscoreBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
		}
	}
	// Overlay
	if !hi.noOverlay {
		m.HiscoreInfo.Overlay.RectData.Draw(layerno)
	}
	// Title and subtitles
	m.HiscoreInfo.Title.TextSpriteData.Draw(layerno)
	m.HiscoreInfo.Title.Rank.TextSpriteData.Draw(layerno)
	m.HiscoreInfo.Title.Result.TextSpriteData.Draw(layerno)
	m.HiscoreInfo.Title.Name.TextSpriteData.Draw(layerno)
	m.HiscoreInfo.Title.Face.TextSpriteData.Draw(layerno)

	for i := 0; i < len(hi.rows); i++ {
		// Portraits bg
		for _, bg := range hi.rows[i].bgs {
			if bg != nil {
				bg.Draw(layerno)
			}
		}
		// Portraits
		for _, a := range hi.rows[i].faces {
			if a == nil {
				continue
			}
			a.Draw(layerno)
		}

		// Text sprites (blink only on highlighted row)
		row := &hi.rows[i]
		if hi.place > 0 && int(hi.place-1) == i {
			// Rank
			if hi.rankUseActive2 {
				row.rankDataActive2.Draw(layerno)
			} else {
				row.rankDataActive.Draw(layerno)
			}
			// Result
			if hi.resultUseActive2 {
				row.resultDataActive2.Draw(layerno)
			} else {
				row.resultDataActive.Draw(layerno)
			}
			// Name
			if hi.nameUseActive2 {
				row.nameDataActive2.Draw(layerno)
			} else {
				row.nameDataActive.Draw(layerno)
			}
		} else {
			row.rankData.Draw(layerno)
			row.resultData.Draw(layerno)
			row.nameData.Draw(layerno)
		}
	}

	// Timer (only when enabled & during input)
	if m.HiscoreInfo.Timer.Count != -1 && hi.input && m.HiscoreInfo.Timer.TextSpriteData != nil {
		m.HiscoreInfo.Timer.TextSpriteData.Draw(layerno)
	}
	// Top background
	if !hi.noBgs && layerno == 2 {
		m.HiscoreBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
	}
}

func (hi *MotifHiscore) finalizeAndSave() {
	if hi.haveSaved || hi.place <= 0 || hi.mode == "" {
		return
	}
	idx := int(hi.place - 1)
	if idx < 0 || idx >= len(hi.rows) {
		return
	}
	name := strings.TrimSpace(hi.rows[idx].name)
	if name == "" {
		// Keep blank if user never entered anything (matches legacy behavior).
		hi.haveSaved = true
		return
	}
	data, err := os.ReadFile(sys.cmdFlags["-stats"])
	if err != nil {
		fmt.Println("hiscore: cannot read save/stats.json for name save:", err)
		hi.haveSaved = true
		return
	}
	path := fmt.Sprintf("modes.%s.ranking.%d.name", hi.mode, idx)
	out, err := sjson.SetBytes(data, path, name)
	if err != nil {
		fmt.Println("hiscore: sjson set failed:", err)
		hi.haveSaved = true
		return
	}
	if err := writeStatsPretty(sys.cmdFlags["-stats"], out); err != nil {
		fmt.Println("hiscore: write save/stats.json failed:", err)
	}
	hi.haveSaved = true
}

func (hi *MotifHiscore) handleBlinkers(m *Motif) {
	// Toggle between Active and Active2 fonts based on switchtime values.
	if hi.place <= 0 {
		return
	}
	if hi.rankActiveCount < m.HiscoreInfo.Item.Rank.Active.Switchtime {
		hi.rankActiveCount++
	} else {
		hi.rankUseActive2 = !hi.rankUseActive2
		hi.rankActiveCount = 0
	}
	if hi.resultActiveCount < m.HiscoreInfo.Item.Result.Active.Switchtime {
		hi.resultActiveCount++
	} else {
		hi.resultUseActive2 = !hi.resultUseActive2
		hi.resultActiveCount = 0
	}
	if hi.nameActiveCount < m.HiscoreInfo.Item.Name.Active.Switchtime {
		hi.nameActiveCount++
	} else {
		hi.nameUseActive2 = !hi.nameUseActive2
		hi.nameActiveCount = 0
	}
}

// hiscorePortraitAnim returns a prepared *Anim for a character "def key" using preloaded
// animations/sprites from sys.sel.charlist. If nothing is found, it returns nil.
func hiscorePortraitAnim(defKey string, m *Motif, x, y float32) *Anim {
	if defKey == "" {
		return nil
	}
	defKey = normalizeDefKey(defKey)
	for i := range sys.sel.charlist {
		sc := &sys.sel.charlist[i]
		if !matchDef(sc.def, defKey) {
			continue
		}
		// Prefer explicit anim number if configured, else sprite tuple.
		var animCopy *Animation
		if m.HiscoreInfo.Item.Face.Anim >= 0 {
			animCopy = sc.anims.get(m.HiscoreInfo.Item.Face.Anim, -1)
		} else if m.HiscoreInfo.Item.Face.Spr[0] >= 0 {
			grp := m.HiscoreInfo.Item.Face.Spr[0]
			idx := m.HiscoreInfo.Item.Face.Spr[1]
			animCopy = sc.anims.get(grp, idx)
		}
		if animCopy == nil {
			// Not preloaded; fall back to unknown handled by caller.
			return nil
		}
		// Wrap *Animation into *Anim
		a := NewAnim(nil, "")
		a.anim = animCopy
		// Optional tuning (localcoord/scale/window/facing) from motif face settings.
		lc := m.HiscoreInfo.Item.Face.Localcoord
		a.SetLocalcoord(float32(lc[0]), float32(lc[1]))
		a.SetPos(x, y)
		sx := m.HiscoreInfo.Item.Face.Scale[0] * sc.portraitscale * float32(sys.motif.Info.Localcoord[0]) / sc.localcoord[0]
		sy := m.HiscoreInfo.Item.Face.Scale[1] * sc.portraitscale * float32(sys.motif.Info.Localcoord[0]) / sc.localcoord[0]
		a.SetScale(sx, sy)
		a.facing = float32(m.HiscoreInfo.Item.Face.Facing)
		w := m.HiscoreInfo.Item.Face.Window
		a.SetWindow([4]float32{float32(w[0]), float32(w[1]), float32(w[2]), float32(w[3])})
		a.layerno = m.HiscoreInfo.Item.Face.Layerno
		return a
	}
	return nil
}

// normalizeDefKey turns "chars/kfm/kfm.def" or "kfm.def" or "kfm" into "kfm".
func normalizeDefKey(s string) string {
	s = strings.ReplaceAll(s, "\\", "/")
	base := filepath.Base(s)
	ext := filepath.Ext(base)
	if ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return strings.ToLower(base)
}

// matchDef checks whether a select entry path matches a stats "def key".
func matchDef(selectDefPath, key string) bool {
	if key == "" {
		return false
	}
	selectDefPath = strings.ReplaceAll(selectDefPath, "\\", "/")
	base := normalizeDefKey(selectDefPath)
	if base == key {
		return true
	}
	// Also allow directory name match (e.g., chars/kfm_zss/char.def -> kfm_zss)
	dir := strings.ToLower(filepath.Base(filepath.Dir(selectDefPath)))
	return dir == key
}

// parseRankingRows reads <path> and converts modes.<mode>.ranking into []rankingRow.
func parseRankingRows(path, mode string) []rankingRow {
	if path == "" || mode == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("hiscore: read stats.json failed:", err)
		return nil
	}
	res := gjson.GetBytes(data, "modes."+mode+".ranking")
	if !res.Exists() || !res.IsArray() {
		return nil
	}
	out := make([]rankingRow, 0, int(res.Get("#").Int()))
	res.ForEach(func(_, v gjson.Result) bool {
		row := rankingRow{
			score: int(v.Get("score").Int()),
			time:  v.Get("time").Float(),
			win:   int(v.Get("win").Int()),
			name:  v.Get("name").Str,
		}
		if ca := v.Get("chars"); ca.Exists() && ca.IsArray() {
			ca.ForEach(func(_, c gjson.Result) bool {
				row.chars = append(row.chars, c.Str)
				return true
			})
		}
		out = append(out, row)
		return true
	})
	return out
}

func initialsWidth(fmtStr string) int {
	// Extract %<N>s (default 3)
	re := regexp.MustCompile(`%([0-9]+)s`)
	m := re.FindStringSubmatch(fmtStr)
	if len(m) >= 2 {
		if n := Atoi(m[1]); n > 0 {
			return int(n)
		}
	}
	return 3
}

// currentGlyph returns current glyph character for the active letter slot.
func currentGlyph(mo *Motif, letters []int) string {
	if len(letters) == 0 {
		return ""
	}
	idx := letters[len(letters)-1] // 1-based
	if idx <= 0 || idx > len(mo.HiscoreInfo.Glyphs) {
		return ""
	}
	return mo.HiscoreInfo.Glyphs[idx-1]
}

// buildNameFromLetters converts indices into the configured glyph table.
func buildNameFromLetters(mo *Motif, letters []int) string {
	var b strings.Builder
	for _, i := range letters {
		if i <= 0 || i > len(mo.HiscoreInfo.Glyphs) {
			continue
		}
		ch := mo.HiscoreInfo.Glyphs[i-1]
		if ch == ">" {
			ch = " "
		}
		b.WriteString(ch)
	}
	return b.String()
}

// updateRowNameFromLetters updates row text sprites (normal + active variants) from letters.
func updateRowNameFromLetters(mo *Motif, row *rankingRow, letters []int) {
	name := buildNameFromLetters(mo, letters)
	if mo.HiscoreInfo.Item.Name.Uppercase {
		name = strings.ToUpper(name)
	}
	row.name = name
	fmtStr := mo.replaceFormatSpecifiers(mo.HiscoreInfo.Item.Name.Text["default"])
	if row.nameData != nil {
		row.nameData.text = fmt.Sprintf(fmtStr, name)
	}
	if row.nameDataActive != nil {
		row.nameDataActive.text = fmt.Sprintf(fmtStr, name)
	}
	if row.nameDataActive2 != nil {
		row.nameDataActive2.text = fmt.Sprintf(fmtStr, name)
	}
}

type MotifVictory struct {
	enabled           bool
	active            bool
	initialized       bool
	counter           int32
	endTimer          int32
	stateDone         bool
	text              string
	lineFullyRendered bool
	charDelayCounter  int32
	typedCnt          int
}

// victoryEntry represents one slot to render on the victory screen
// (either a loaded character or not loaded Turns member).
type victoryEntry struct {
	side     int   // 0 or 1
	memberNo int   // 0-based team order
	cn       int   // index into sys.sel.charlist
	pal      int   // 1-based palette number
	c        *Char // non-nil if this member is currently loaded
}

func (vi *MotifVictory) reset(m *Motif) {
	vi.active = false
	vi.initialized = false
	vi.stateDone = false
	vi.lineFullyRendered = false
	vi.charDelayCounter = 0
	vi.typedCnt = 0
	// Victory screen uses its own typewriter logic, so disable the internal TextSprite typing.
	m.VictoryScreen.WinQuote.TextSpriteData.textDelay = 0
	vi.endTimer = -1
	vi.clear(m)
	if !sys.skipMotifScaling() {
		sys.applyFightAspect()
	}
}

func (vi *MotifVictory) clearProps(props *PlayerVictoryProperties) {
	props.AnimData = NewAnim(nil, "")
	props.Face2.AnimData = NewAnim(nil, "")
	props.Name.TextSpriteData.text = ""
}

func (vi *MotifVictory) clear(m *Motif) {
	vi.clearProps(&m.VictoryScreen.P1)
	vi.clearProps(&m.VictoryScreen.P2)
	vi.clearProps(&m.VictoryScreen.P3)
	vi.clearProps(&m.VictoryScreen.P4)
	vi.clearProps(&m.VictoryScreen.P5)
	vi.clearProps(&m.VictoryScreen.P6)
	vi.clearProps(&m.VictoryScreen.P7)
	vi.clearProps(&m.VictoryScreen.P8)
}

func (vi *MotifVictory) getVictoryQuote(m *Motif, p *Char) string {
	if p == nil || p.playerNo < 0 || p.playerNo >= len(sys.cgi) {
		return m.VictoryScreen.WinQuote.Text
	}
	quoteIndex := int(p.winquote)
	playerQuotes := sys.cgi[p.playerNo].quotes

	//fmt.Printf("[Victory] Winner team=%d playerNo=%d initialWinquote=%d\n", sys.winnerTeam(), p.playerNo, quoteIndex)

	// Check if the quote index is out of range
	if quoteIndex < 0 || quoteIndex >= MaxQuotes {
		// Collect available quote indices
		availableQuotes := []int{}
		for i, quote := range playerQuotes {
			if quote != "" {
				availableQuotes = append(availableQuotes, i)
			}
		}

		// Select a random available quote if any exist
		if len(availableQuotes) > 0 {
			idx := int(RandI(0, int32(len(availableQuotes)-1)))
			quoteIndex = availableQuotes[idx]
		} else {
			quoteIndex = -1
		}
	}

	// Return the selected quote if valid, otherwise fall back to the default
	//fmt.Printf("[Victory] Using quoteIndex=%d (MaxQuotes=%d). Fallback text present=%v\n", quoteIndex, MaxQuotes, m.VictoryScreen.WinQuote.Text != "")
	if quoteIndex != -1 && len(playerQuotes) == MaxQuotes {
		return playerQuotes[quoteIndex]
	}
	return m.VictoryScreen.WinQuote.Text
}

// buildSideOrder reconstructs the list of members to render for a side
//   - Winner: last hitter leader, then other winners (alive first unless allowKO)
//   - Loser : first encountered leader, then other losers
//   - Fill with not loaded Turns members from the original selection
func (vi *MotifVictory) buildSideOrder(side int, allowKO bool, maxNum int) []victoryEntry {
	winnerSide := int(sys.winnerTeam() - 1)
	if maxNum <= 0 {
		//fmt.Printf("[Victory] buildSideOrder side=%d allowKO=%v maxNum=%d winnerSide=%d -> SKIP (num=0)\n", side, allowKO, maxNum, winnerSide)
		return nil
	}
	out := make([]victoryEntry, 0, maxNum)
	usedMember := map[int]bool{}
	//fmt.Printf("[Victory] buildSideOrder side=%d allowKO=%v maxNum=%d winnerSide=%d\n", side, allowKO, maxNum, winnerSide)

	// Helper to push a loaded char
	pushLoaded := func(c *Char) {
		if c == nil || int(c.teamside) != side {
			return
		}
		mn := int(c.memberNo)
		if usedMember[mn] || len(out) >= maxNum {
			return
		}
		out = append(out, victoryEntry{
			side:     side,
			memberNo: mn,
			cn:       int(c.selectNo),
			pal:      int(c.gi().palno),
			c:        c,
		})
		usedMember[mn] = true
		//fmt.Printf("[Victory] -> pushLoaded: side=%d memberNo=%d cn=%d pal=%d alive=%v leader=%v\n", side, mn, int(c.selectNo), int(c.gi().palno), c.alive(), len(out) == 1)
	}

	// 1) Choose leader
	if side == winnerSide {
		leaderPn := sys.lastHitter[side]
		if leaderPn < 0 {
			leaderPn = sys.teamLeader[side]
		}
		if leaderPn >= 0 && leaderPn < MaxPlayerNo && len(sys.chars[leaderPn]) > 0 {
			pushLoaded(sys.chars[leaderPn][0])
		}
	} else {
		// Loser: first encountered from this side
		for i := 0; i < MaxPlayerNo && len(out) < 1; i++ {
			if len(sys.chars[i]) == 0 {
				continue
			}
			if int(sys.chars[i][0].teamside) == side {
				pushLoaded(sys.chars[i][0])
				break
			}
		}
	}

	// 2) Append remaining loaded members from this side
	for i := 0; i < MaxPlayerNo && len(out) < maxNum; i++ {
		if len(sys.chars[i]) == 0 {
			continue
		}
		c := sys.chars[i][0]
		if int(c.teamside) != side {
			continue
		}
		// Skip if already used as leader
		if len(out) > 0 && out[0].c == c {
			continue
		}
		// Winner: prefer alive unless allowKO
		if side == winnerSide {
			if c.alive() || allowKO {
				pushLoaded(c)
			}
		} else {
			// Loser: include regardless of alive status (matches legacy loop)
			pushLoaded(c)
		}
	}

	// 3) Fill with un-loaded Turns team members from original select order
	if len(out) < maxNum {
		sel := sys.sel.selected[side]
		leaderMember := -1
		if len(out) > 0 {
			leaderMember = out[0].memberNo
		}
		for k := 0; k < len(sel) && len(out) < maxNum; k++ {
			if usedMember[k] {
				continue
			}
			if !allowKO && leaderMember != -1 && k <= leaderMember {
				continue
			}
			cn := int(sel[k][0])
			pl := int(sel[k][1])
			out = append(out, victoryEntry{
				side:     side,
				memberNo: k,
				cn:       cn,
				pal:      pl,
				c:        nil, // not loaded this round
			})
			usedMember[k] = true
		}
	}
	if len(out) > maxNum {
		//fmt.Printf("[Victory] Truncating out to %d (had %d)\n", maxNum, len(out))
		out = out[:maxNum]
	}
	return out
}

// applyEntry fills one PlayerVictoryProperties slot from a victoryEntry.
func (vi *MotifVictory) applyEntry(m *Motif, dst *PlayerVictoryProperties, e victoryEntry, slotName string, isLoser bool) {
	// Name
	if e.c != nil {
		dst.Name.TextSpriteData.text = e.c.gi().displayname
	} else {
		sc := sys.sel.GetChar(e.cn)
		if sc != nil {
			name := sc.lifebarname
			if name == "" {
				name = sc.name
			}
			dst.Name.TextSpriteData.text = name
		}
	}
	//fmt.Printf("[Victory] applyEntry slot=%s side=%d memberNo=%d cn=%d pal=%d loaded=%v name=%q\n", slotName, e.side, e.memberNo, e.cn, e.pal, e.c != nil, dst.Name.TextSpriteData.text)
	// Resolve SelectChar (for portraits)
	sc := sys.sel.GetChar(e.cn)

	// Default win portraits
	targetSpr := dst.Spr
	targetAnim := dst.Anim
	targetFace2Spr := dst.Face2.Spr
	targetFace2Anim := dst.Face2.Anim
	faceBrightness := int32(256)
	face2Brightness := int32(256)

	if isLoser {
		if dst.Lose.Spr[0] != -1 {
			targetSpr = dst.Lose.Spr
		}
		if dst.Lose.Anim != -1 {
			targetAnim = dst.Lose.Anim
		}
		if dst.Face2.Lose.Spr[0] != -1 {
			targetFace2Spr = dst.Face2.Lose.Spr
		}
		if dst.Face2.Lose.Anim != -1 {
			targetFace2Anim = dst.Face2.Lose.Anim
		}
		faceBrightness = dst.Lose.Brightness
		face2Brightness = dst.Face2.Lose.Brightness
	}
	memberIdx := float32(e.memberNo)
	// Main face
	mainSpacingX := float32(dst.Spacing[0]) * memberIdx
	mainSpacingY := float32(dst.Spacing[1]) * memberIdx
	mainX := dst.Pos[0] + dst.Offset[0] + mainSpacingX
	mainY := dst.Pos[1] + dst.Offset[1] + mainSpacingY
	dst.AnimData = victoryPortraitAnim(
		m, sc, slotName+".main",
		targetAnim, targetSpr,
		dst.Localcoord, dst.Layerno, dst.Facing,
		dst.Scale, dst.Xshear, dst.Angle, dst.XAngle,
		dst.YAngle, dst.Projection, dst.Focallength,
		dst.Velocity, dst.MaxDist, dst.Accel, dst.Friction,
		dst.Window, mainX, mainY,
		dst.ApplyPal || e.c == nil,
		e.pal, faceBrightness, e.c,
	)
	// Face2
	face2SpacingX := float32(dst.Face2.Spacing[0]) * memberIdx
	face2SpacingY := float32(dst.Face2.Spacing[1]) * memberIdx
	face2X := dst.Pos[0] + dst.Face2.Offset[0] + face2SpacingX
	face2Y := dst.Pos[1] + dst.Face2.Offset[1] + face2SpacingY
	dst.Face2.AnimData = victoryPortraitAnim(
		m, sc, slotName+".face2",
		targetFace2Anim, targetFace2Spr,
		dst.Face2.Localcoord, dst.Face2.Layerno, dst.Face2.Facing,
		dst.Face2.Scale, dst.Face2.Xshear, dst.Face2.Angle, dst.Face2.XAngle,
		dst.Face2.YAngle, dst.Face2.Projection, dst.Face2.Focallength,
		dst.Face2.Velocity, dst.Face2.MaxDist, dst.Face2.Accel, dst.Face2.Friction,
		dst.Face2.Window, face2X, face2Y,
		dst.Face2.ApplyPal || e.c == nil,
		e.pal, face2Brightness, e.c,
	)
	if dst.AnimData == nil && dst.Face2.AnimData == nil {
		//fmt.Printf("[Victory] slot=%s -> WARNING: both main and face2 animations are nil\n", slotName)
	}
}

func (vi *MotifVictory) init(m *Motif) {
	if !m.VictoryScreen.Enabled || !vi.enabled || sys.winnerTeam() < 1 || (sys.winnerTeam() == 2 && !m.VictoryScreen.Cpu.Enabled) ||
		((sys.gameMode == "versus" || sys.gameMode == "netplayversus") && !m.VictoryScreen.Vs.Enabled) ||
		!sys.sel.gameParams.VictoryScreen {
		vi.initialized = true
		return
	}
	if err := sys.luaLState.DoString("hook.run('game.victory_init')"); err != nil {
		sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.victory_init", err.Error())
	}

	// Determine which character is used for victory screen (same one used below as leader/quote),
	// and honor select.def character param: victoryscreen (skip completely when false).
	winnerSide := int(sys.winnerTeam() - 1)
	loserSide := winnerSide ^ 1
	maxW := int(Clamp(m.VictoryScreen.P1.Num, 0, 4))
	maxL := int(Clamp(m.VictoryScreen.P2.Num, 0, 4))
	wEntries := vi.buildSideOrder(winnerSide, m.VictoryScreen.Winner.TeamKo.Enabled, maxW)
	lEntries := vi.buildSideOrder(loserSide, true, maxL) // losers always allow KO display

	if len(wEntries) > 0 {
		cn := wEntries[0].cn
		if cn >= 0 {
			if sc := sys.sel.GetChar(cn); sc != nil && sc.scp != nil && !sc.scp.VictoryScreen {
				vi.initialized = true
				return
			}
		}
	}

	if !sys.skipMotifScaling() {
		sys.setGameSize(sys.scrrect[2], sys.scrrect[3])
	}

	//fmt.Printf("[Victory] init: enabled=%v winnerTeam=%d cpu.enabled=%v p1.num=%d p2.num=%d\n", m.VictoryScreen.Enabled, sys.winnerTeam(), m.VictoryScreen.Cpu.Enabled, m.VictoryScreen.P1.Num, m.VictoryScreen.P2.Num)

	// Apply to motif slots: winners -> P1,P3,P5,P7 ; losers -> P2,P4,P6,P8
	wSlots := []*PlayerVictoryProperties{&m.VictoryScreen.P1, &m.VictoryScreen.P3, &m.VictoryScreen.P5, &m.VictoryScreen.P7}
	lSlots := []*PlayerVictoryProperties{&m.VictoryScreen.P2, &m.VictoryScreen.P4, &m.VictoryScreen.P6, &m.VictoryScreen.P8}
	wNames := []string{"P1", "P3", "P5", "P7"}
	lNames := []string{"P2", "P4", "P6", "P8"}
	if m.VictoryScreen.KeepSide.Enabled && sys.winnerTeam() == 2 {
		wSlots, lSlots = lSlots, wSlots
	}
	for i := 0; i < len(wEntries) && i < len(wSlots); i++ {
		vi.applyEntry(m, wSlots[i], wEntries[i], wNames[i], false)
	}
	for i := 0; i < len(lEntries) && i < len(lSlots); i++ {
		vi.applyEntry(m, lSlots[i], lEntries[i], lNames[i], true)
	}

	var leader *Char
	if len(wEntries) > 0 {
		leader = wEntries[0].c
	}
	vi.text = vi.getVictoryQuote(m, leader)

	m.VictoryScreen.WinName.TextSpriteData.text = ""
	if leader != nil {
		// Displays only the winner’s name regardless of team side or player number
		m.VictoryScreen.WinName.TextSpriteData.text = leader.gi().displayname
	}
	m.VictoryBgDef.BGDef.Reset()

	//fmt.Printf("[Victory] init done. Winners=%d entries, Losers=%d entries. WinQuote=%q\n", len(wEntries), len(lEntries), vi.text)

	if sys.winnerTeam() == 1 {
		m.processStateTransitions(m.VictoryScreen.P1.State, m.VictoryScreen.P1.Teammate.State, m.VictoryScreen.P2.State, m.VictoryScreen.P2.Teammate.State)
	} else if sys.winnerTeam() == 2 {
		m.processStateTransitions(m.VictoryScreen.P2.State, m.VictoryScreen.P2.Teammate.State, m.VictoryScreen.P1.State, m.VictoryScreen.P1.Teammate.State)
	}

	if !m.VictoryScreen.Sounds.Enabled {
		sys.clearAllSound()
		sys.noSoundFlg = true
	}

	m.Music.Play("victory", sys.motif.Def)

	m.VictoryScreen.FadeIn.FadeData.init(m.fadeIn, true)
	vi.counter = 0
	vi.active = true
	vi.initialized = true
}

func (vi *MotifVictory) step(m *Motif) {
	if vi.counter > 0 {
		if err := sys.luaLState.DoString("hook.run('game.victory')"); err != nil {
			sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.victory", err.Error())
		}
	}
	cancelPressed := sys.esc || sys.button(m.VictoryScreen.Cancel.Key, -1)
	skipPressed := sys.button(m.VictoryScreen.Skip.Key, -1)
	prevLineFullyRendered := vi.lineFullyRendered
	//fmt.Printf("[Victory] step: counter=%d time=%d endTimer=%d typedCnt=%d lineFullyRendered=%v cancel=%v skip=%v\n", vi.counter, m.VictoryScreen.Time, vi.endTimer, vi.typedCnt, vi.lineFullyRendered, cancelPressed, skipPressed)

	m.VictoryScreen.P1.AnimData.Update(false)
	m.VictoryScreen.P2.AnimData.Update(false)
	m.VictoryScreen.P3.AnimData.Update(false)
	m.VictoryScreen.P4.AnimData.Update(false)
	m.VictoryScreen.P5.AnimData.Update(false)
	m.VictoryScreen.P6.AnimData.Update(false)
	m.VictoryScreen.P7.AnimData.Update(false)
	m.VictoryScreen.P8.AnimData.Update(false)

	m.VictoryScreen.P1.Face2.AnimData.Update(false)
	m.VictoryScreen.P2.Face2.AnimData.Update(false)
	m.VictoryScreen.P3.Face2.AnimData.Update(false)
	m.VictoryScreen.P4.Face2.AnimData.Update(false)
	m.VictoryScreen.P5.Face2.AnimData.Update(false)
	m.VictoryScreen.P6.Face2.AnimData.Update(false)
	m.VictoryScreen.P7.Face2.AnimData.Update(false)
	m.VictoryScreen.P8.Face2.AnimData.Update(false)

	// First press of Skip: fast-forward the text, but do NOT start fadeout yet.
	if skipPressed && !prevLineFullyRendered {
		totalRunes := utf8.RuneCountInString(vi.text)
		vi.typedCnt = totalRunes
		vi.lineFullyRendered = true
		vi.charDelayCounter = 0
		//fmt.Printf("[Victory] Skip pressed -> fast-forward winquote (totalRunes=%d)\n", totalRunes)
	}

	// While we haven't finished typing the quote, keep revealing characters
	// regardless of the global time limit. Fadeout will only start once the
	// line is fully rendered (see logic below).
	if !vi.lineFullyRendered {
		StepTypewriter(
			vi.text,
			&vi.typedCnt,
			&vi.charDelayCounter,
			&vi.lineFullyRendered,
			float32(m.VictoryScreen.WinQuote.TextDelay),
		)
	}

	// Clamp typedLen so it doesn't exceed the line length
	totalRunes := utf8.RuneCountInString(vi.text)
	typedLen := vi.typedCnt
	if typedLen > totalRunes {
		typedLen = totalRunes
	}

	m.VictoryScreen.WinQuote.TextSpriteData.wrapText(vi.text, typedLen)
	m.VictoryScreen.WinQuote.TextSpriteData.Update()

	// Decide when to start fadeout: Cancel key / Skip key / Time limit
	if vi.endTimer == -1 {
		userInterrupt := cancelPressed || (skipPressed && prevLineFullyRendered)
		timeUp := vi.lineFullyRendered && vi.counter >= m.VictoryScreen.Time

		if userInterrupt || timeUp {
			startFadeOut(m.VictoryScreen.FadeOut.FadeData, m.fadeOut, userInterrupt, m.fadePolicy)
			vi.endTimer = vi.counter + m.fadeOut.timeRemaining
			//fmt.Printf("[Victory] Starting fadeout: counter=%d time=%d endTimer=%d userInterrupt=%v timeUp=%v\n", vi.counter, m.VictoryScreen.Time, vi.endTimer, userInterrupt, timeUp)
		}
	}

	// Check if the sequence has ended
	if vi.endTimer != -1 && vi.counter >= vi.endTimer {
		if m.fadeOut != nil {
			m.fadeOut.reset()
		}
		vi.active = false
		if !m.VictoryScreen.Sounds.Enabled {
			sys.noSoundFlg = false
		}
		return
	}

	// Increment counter
	vi.counter++
}

func (vi *MotifVictory) draw(m *Motif, layerno int16) {
	// Order victory slots by draworder (int). Higher = drawn later (on top).
	type slotRef struct {
		idx   int
		p     *PlayerVictoryProperties
		order int32
	}
	slots := []slotRef{
		{0, &m.VictoryScreen.P1, m.VictoryScreen.P1.DrawOrder},
		{1, &m.VictoryScreen.P2, m.VictoryScreen.P2.DrawOrder},
		{2, &m.VictoryScreen.P3, m.VictoryScreen.P3.DrawOrder},
		{3, &m.VictoryScreen.P4, m.VictoryScreen.P4.DrawOrder},
		{4, &m.VictoryScreen.P5, m.VictoryScreen.P5.DrawOrder},
		{5, &m.VictoryScreen.P6, m.VictoryScreen.P6.DrawOrder},
		{6, &m.VictoryScreen.P7, m.VictoryScreen.P7.DrawOrder},
		{7, &m.VictoryScreen.P8, m.VictoryScreen.P8.DrawOrder},
	}
	sort.SliceStable(slots, func(i, j int) bool {
		if slots[i].order == slots[j].order {
			return slots[i].idx < slots[j].idx
		}
		return slots[i].order < slots[j].order
	})
	// Overlay
	m.VictoryScreen.Overlay.RectData.Draw(layerno)
	// Background
	if m.VictoryBgDef.BgClearColor[0] >= 0 {
		m.VictoryBgDef.RectData.Draw(layerno)
	}
	if layerno <= 1 {
		m.VictoryBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
	}
	// Face2 portraits
	for _, s := range slots {
		s.p.Face2.AnimData.Draw(layerno)
	}
	// Face portraits
	for _, s := range slots {
		s.p.AnimData.Draw(layerno)
	}
	// Name
	m.VictoryScreen.P1.Name.TextSpriteData.Draw(layerno)
	m.VictoryScreen.P2.Name.TextSpriteData.Draw(layerno)
	m.VictoryScreen.P3.Name.TextSpriteData.Draw(layerno)
	m.VictoryScreen.P4.Name.TextSpriteData.Draw(layerno)
	m.VictoryScreen.P5.Name.TextSpriteData.Draw(layerno)
	m.VictoryScreen.P6.Name.TextSpriteData.Draw(layerno)
	m.VictoryScreen.P7.Name.TextSpriteData.Draw(layerno)
	m.VictoryScreen.P8.Name.TextSpriteData.Draw(layerno)
	// Winner Name
	m.VictoryScreen.WinName.TextSpriteData.Draw(layerno)
	// Winquote
	m.VictoryScreen.WinQuote.TextSpriteData.Draw(layerno)
	// Top background
	if layerno == 2 {
		m.VictoryBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
	}
}

// buildSingleFrameFromSFF creates a 1-frame Animation from a raw sprite (grp, idx).
// Used when a motif references .spr (group/index) and the preloaded table lacks it.
func buildSingleFrameFromSFF(sff *Sff, grp, idx int32) *Animation {
	if sff == nil || sff.GetSprite(uint16(grp), uint16(idx)) == nil {
		return nil
	}
	anim := newAnimation(sff, &sff.palList)
	anim.mask = 0
	af := newAnimFrame()
	af.Group, af.Number = grp, idx
	af.Time = 1 // stable single-frame
	anim.frames = append(anim.frames, *af)
	return anim
}

// tryGetPortrait tries a sequence of (group,index) pairs first from preloaded
// SelectChar anims, then by building a single-frame Animation from the owner SFF.
// Returns the first non-nil *Animation and a label describing where it came from.
func tryGetPortrait(sc *SelectChar, ownerC *Char, pairs [][2]int32) (anim *Animation, from string) {
	for _, p := range pairs {
		grp, idx := p[0], p[1]
		if sc != nil {
			if a := sc.anims.get(grp, idx); a != nil {
				return a, fmt.Sprintf("preloaded(%d,%d)", grp, idx)
			}
		}
		if ownerC != nil && ownerC.playerNo >= 0 && ownerC.playerNo < len(sys.cgi) && sys.cgi[ownerC.playerNo].sff != nil {
			if a := buildSingleFrameFromSFF(sys.cgi[ownerC.playerNo].sff, grp, idx); a != nil {
				return a, fmt.Sprintf("sff(%d,%d)", grp, idx)
			}
		}
	}
	return nil, ""
}

// victoryPortraitAnim builds a *Anim for a character select entry and positions it.
// It uses per-character preloaded animations (sys.sel.charlist[cn].anims).
// If the requested anim/spr is missing, it falls back to (9000,1) then (9000,0).
func victoryPortraitAnim(m *Motif, sc *SelectChar, slot string,
	animNo int32, spr [2]int32,
	localcoord [2]int32, layerno int16, facing int32,
	scale [2]float32, xshear float32, angle float32, xangle float32,
	yangle float32, projection string, fLength float32,
	velocity [2]float32, maxDist [2]float32, accel [2]float32, friction [2]float32,
	window [4]int32, x, y float32, applyPal bool, pal int, brightness int32, ownerC *Char) *Anim {

	//fmt.Printf("[Victory] buildPortrait slot=%s scNil=%v animNo=%d spr=(%d,%d) pos=(%.1f,%.1f) scale=(%.3f,%.3f) localcoord=(%d,%d) window=(%d,%d,%d,%d) applyPal=%v pal=%d\n", slot, sc == nil, animNo, spr[0], spr[1], x, y, scale[0], scale[1], localcoord[0], localcoord[1], window[0], window[1], window[2], window[3], applyPal, pal)

	a := NewAnim(nil, "")

	var animCopy *Animation
	if sc != nil && animNo >= 0 {
		// First: explicit animation number
		animCopy = sc.anims.get(animNo, -1)
		if animCopy == nil {
			// if the specific anim is missing, try default big portrait
			if a, _ /*from*/ := tryGetPortrait(sc, ownerC, [][2]int32{{9000, 1} /*, {9000, 0}*/}); a != nil {
				animCopy = a
				//fmt.Printf("[Victory] slot=%s -> fallback from anim %d to %s\n", slot, animNo/*, from*/)
			}
		}
	} else if sc != nil && spr[0] >= 0 {
		// Try requested (grp,idx) first (preloaded or SFF-build), then fall back to 9000,1
		want := [][2]int32{{spr[0], spr[1]}, {9000, 1} /*, {9000, 0}*/}
		if a, _ /*from*/ := tryGetPortrait(sc, ownerC, want); a != nil {
			animCopy = a
		} else {
			// Detailed failure logs for the first requested pair
			if ownerC != nil && ownerC.playerNo >= 0 && ownerC.playerNo < len(sys.cgi) && sys.cgi[ownerC.playerNo].sff != nil {
				if sys.cgi[ownerC.playerNo].sff.GetSprite(uint16(spr[0]), uint16(spr[1])) == nil {
					//fmt.Printf("[Victory] slot=%s -> FAILED to build 1-frame anim: sprite not in SFF (spr=%d,%d)\n", slot, spr[0], spr[1])
				}
			} else {
				//fmt.Printf("[Victory] slot=%s -> owner SFF is nil; cannot build 1-frame anim (spr=%d,%d)\n", slot, spr[0], spr[1])
			}
		}
	}
	if animCopy != nil {
		a.anim = animCopy
	} else {
		//fmt.Printf("[Victory] slot=%s -> animCopy=nil (animNo=%d spr=%d,%d). Check if your portraits are defined as an ANIM or plain SPR.\n", slot, animNo, spr[0], spr[1])
	}
	// Localcoord / window / layer / facing
	//a.SetLocalcoord(float32(localcoord[0]), float32(localcoord[1]))
	if localcoord[0] > 0 && localcoord[1] > 0 {
		a.SetLocalcoord(float32(localcoord[0]), float32(localcoord[1]))
	} else {
		//fmt.Printf("[Victory] slot=%s -> skip SetLocalcoord (0,0); using default engine localcoord\n", slot)
	}
	a.layerno = layerno
	//a.SetWindow([4]float32{float32(window[0]), float32(window[1]), float32(window[2]), float32(window[3])})
	if window[2] > window[0] && window[3] > window[1] {
		a.SetWindow([4]float32{float32(window[0]), float32(window[1]), float32(window[2]), float32(window[3])})
	} else {
		//fmt.Printf("[Victory] slot=%s -> skip SetWindow (no clipping)\n", slot)
	}
	// Position
	a.SetPos(x, y)
	// Scale: include character portraitscale and coord conversion similar to hiscore
	sx, sy := scale[0], scale[1]
	// Only apply SelectChar scaling if sc is present and has valid localcoord.
	if sc != nil && sc.localcoord[0] > 0 {
		base := float32(sys.motif.Info.Localcoord[0]) / sc.localcoord[0]
		sx = scale[0] * sc.portraitscale * base
		sy = scale[1] * sc.portraitscale * base
	}
	a.SetScale(sx, sy)
	a.facing = float32(facing)
	if sx == 0 || sy == 0 {
		//fmt.Printf("[Victory] slot=%s -> WARNING: zero scale sx=%.4f sy=%.4f (check portraitscale/localcoord)\n", slot, sx, sy)
	}
	// Transformations
	a.xshear = xshear
	a.rot.angle = angle
	a.rot.xangle = xangle
	a.rot.yangle = yangle
	a.projection = int32(Projection_Orthographic) // Default
	v := strings.ToLower(strings.TrimSpace(projection))
	switch v {
	case "perspective":
		a.projection = int32(Projection_Perspective)
	case "perspective2":
		a.projection = int32(Projection_Perspective2)
	case "orthographic":
		// default, we don't need to do nothing
	default:
		if IsInt(v) {
			a.projection = Atoi(v)
		}
	}
	a.fLength = fLength
	// Movement
	a.SetVelocity(velocity[0], velocity[1])
	a.SetMaxDist(maxDist[0], maxDist[1])
	a.SetAccel(accel[0], accel[1])
	a.friction = friction
	// Palette for non-loaded (or force-apply if requested)
	isCopied := false
	if applyPal && pal > 0 && a.anim != nil && a.anim.sff != nil {
		if len(a.anim.sff.palList.paletteMap) > 0 {
			a = a.Copy()
			isCopied = true
			a.anim.sff.palList.paletteMap[0] = pal - 1
		}
		//fmt.Printf("[Victory] slot=%s -> applied palette %d\n", slot, pal)
	}

	if brightness > 0 && brightness < 256 && a.anim != nil && a.anim.sff != nil {
		if !isCopied {
			a = a.Copy()
			isCopied = true
		}
		val := brightness
		if val < 0 {
			val = 0
		}
		if val > 256 {
			val = 256
		}
		a.palfx.time = -1
		a.palfx.mul = [3]int32{val, val, val}
	}

	return a
}

type MotifWin struct {
	winEnabled      bool
	loseEnabled     bool
	active          bool
	initialized     bool
	counter         int32
	endTimer        int32
	fadeIn          *Fade
	fadeOut         *Fade
	stateDone       bool
	soundsEnabled   bool
	fadeOutTime     int32
	time            int32
	keyCancel       []string
	p1State         []int32
	p1TeammateState []int32
	p2State         []int32
	p2TeammateState []int32
	stateTime       int32
	//winCount        int32
	//loseCnt         int32

	resultsScreen *ResultsScreenProperties
	resultsBgDef  *BgDefProperties
	resultsKey    string
}

// Assign state data to MotifWin
func (wi *MotifWin) assignStates(p1, p1Teammate, p2, p2Teammate []int32) {
	wi.p1State = p1
	wi.p1TeammateState = p1Teammate
	wi.p2State = p2
	wi.p2TeammateState = p2Teammate
	if !sys.skipMotifScaling() {
		sys.applyFightAspect()
	}
}

func (wi *MotifWin) reset(m *Motif) {
	wi.active = false
	wi.initialized = false
	wi.stateDone = false
	wi.endTimer = -1

	wi.resultsScreen = nil
	wi.resultsBgDef = nil
	wi.resultsKey = ""
}

// Initialize the MotifWin based on the current game mode
func (wi *MotifWin) init(m *Motif) {
	if (wi.winEnabled && sys.winnerTeam() != 0 && sys.winnerTeam() != int32(sys.home)+1) ||
		(wi.loseEnabled && (sys.winnerTeam() == 0 || sys.winnerTeam() == int32(sys.home)+1)) {
		if err := sys.luaLState.DoString("hook.run('game.result_init')"); err != nil {
			sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.result_init", err.Error())
		}
		// Variant selection: [Win Screen] results.<gamemode> = <section name>
		// If not defined, fall back to standard Win Screen.
		variant := strings.TrimSpace(m.WinScreen.Results[sys.gameMode])
		if variant == "" || strings.EqualFold(variant, "win screen") {
			if !wi.initWinScreen(m) {
				wi.initialized = true
				return
			}
		} else if !wi.initResultsVariant(m, variant) {
			// If the configured variant doesn't exist/disabled, fall back to Win Screen.
			if !wi.initWinScreen(m) {
				wi.initialized = true
				return
			}
		}
	} else {
		wi.initialized = true
		return
	}
	if !sys.skipMotifScaling() {
		sys.setGameSize(sys.scrrect[2], sys.scrrect[3])
	}

	if !wi.soundsEnabled {
		sys.clearAllSound()
		sys.noSoundFlg = true
	}

	m.Music.Play("results", sys.motif.Def)

	wi.fadeIn.init(m.fadeIn, true)
	wi.counter = 0
	wi.active = true
	wi.initialized = true
}

func (wi *MotifWin) initResultsVariant(m *Motif, sectionName string) bool {
	// sectionName is something like "X Results Screen".
	// Parsed section keys are lowercased, and spaces become underscores in the map key.
	name := strings.TrimSpace(sectionName)
	if m == nil || name == "" || len(m.ResultsScreen) == 0 {
		return false
	}

	rsKey := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	rs := m.ResultsScreen[rsKey]
	if rs == nil || !rs.Enabled {
		return false
	}

	// BgDef key: "<base>resultsbgdef" where base is section name without " Results Screen".
	base := strings.ToLower(name)
	if strings.HasSuffix(base, " results screen") {
		base = strings.TrimSuffix(base, " results screen")
	}
	base = strings.TrimSpace(base)
	base = strings.ReplaceAll(base, " ", "")
	bgKey := base + "resultsbgdef"
	bg := m.ResultsBgDef[bgKey]

	if bg != nil && bg.BGDef != nil {
		bg.BGDef.Reset()
	}

	wi.resultsScreen = rs
	wi.resultsBgDef = bg
	wi.resultsKey = rsKey

	// Formatting type comes from [Hiscore Info] ranking.<gamemode> = score|time|win
	dataType := strings.ToLower(strings.TrimSpace(m.HiscoreInfo.Ranking[sys.gameMode]))

	// Unified win/lose selection: win anims only if we would get a hiscore entry (name input).
	wouldPlace := rankingWouldPlace(sys.gameMode)

	switch dataType {
	case "win":
		if rs.WinsText.TextSpriteData != nil {
			tal := tallyRun()
			rs.WinsText.TextSpriteData.text = m.sprintf(rs.WinsText.Text, tal.winP1, tal.loseP1)
		}
		if wouldPlace && rs.RoundsToWin > 0 && sys.match >= rs.RoundsToWin {
			wi.assignStates(rs.P1.Win.State, rs.P1.Teammate.Win.State, rs.P2.Win.State, rs.P2.Teammate.Win.State)
		} else {
			wi.assignStates(rs.P1.State, rs.P1.Teammate.State, rs.P2.State, rs.P2.Teammate.State)
		}
	case "time":
		if rs.WinsText.TextSpriteData != nil {
			rs.WinsText.TextSpriteData.text = formatTimeText(rs.WinsText.Text, float64(sys.timeTotal())/60)
		}
		if wouldPlace {
			wi.assignStates(
				rs.P1.Win.State, rs.P1.Teammate.Win.State,
				rs.P2.Win.State, rs.P2.Teammate.Win.State,
			)
		} else {
			wi.assignStates(rs.P1.State, rs.P1.Teammate.State, rs.P2.State, rs.P2.Teammate.State)
		}
	case "score":
		if rs.WinsText.TextSpriteData != nil {
			fs := m.replaceFormatSpecifiers(rs.WinsText.Text)
			// Total score for P1 side:
			sc := int32(0)
			if len(sys.scoreStart) > 0 {
				sc = int32(sys.scoreStart[0])
			}
			for _, v := range sys.scoreRounds {
				if len(v) > 0 {
					sc += int32(v[0])
				}
			}
			rs.WinsText.TextSpriteData.text = fmt.Sprintf(fs, sc)
		}
		if wouldPlace {
			wi.assignStates(
				rs.P1.Win.State, rs.P1.Teammate.Win.State,
				rs.P2.Win.State, rs.P2.Teammate.Win.State,
			)
		} else {
			wi.assignStates(rs.P1.State, rs.P1.Teammate.State, rs.P2.State, rs.P2.Teammate.State)
		}
	default:
		// dataType == "" (or unknown): no formatting fallback.
		if rs.WinsText.TextSpriteData != nil {
			rs.WinsText.TextSpriteData.text = rs.WinsText.Text
		}
		wi.assignStates(rs.P1.State, rs.P1.Teammate.State, rs.P2.State, rs.P2.Teammate.State)
	}

	wi.stateTime = rs.State.Time
	wi.soundsEnabled = rs.Sounds.Enabled
	wi.keyCancel = rs.Cancel.Key
	wi.time = rs.Show.Time
	wi.fadeOutTime = rs.FadeOut.Time
	wi.fadeIn = rs.FadeIn.FadeData
	wi.fadeOut = rs.FadeOut.FadeData
	return true
}

// Handle win screen mode initialization
func (wi *MotifWin) initWinScreen(m *Motif) bool {
	if sys.home != 1 || !m.WinScreen.Enabled {
		return false
	}

	m.WinBgDef.BGDef.Reset()

	wi.resultsScreen = nil
	wi.resultsBgDef = nil
	wi.resultsKey = ""

	wi.assignStates(m.WinScreen.P1.State, m.WinScreen.P1.Teammate.State, m.WinScreen.P2.State, m.WinScreen.P2.Teammate.State)
	wi.stateTime = m.WinScreen.State.Time
	wi.soundsEnabled = m.WinScreen.Sounds.Enabled

	wi.keyCancel = m.WinScreen.Cancel.Key
	wi.time = m.WinScreen.Pose.Time
	wi.fadeOutTime = m.WinScreen.FadeOut.Time
	wi.fadeIn = m.WinScreen.FadeIn.FadeData
	wi.fadeOut = m.WinScreen.FadeOut.FadeData
	return true
}

// Process the step logic for MotifWin
func (wi *MotifWin) step(m *Motif) {
	if wi.counter > 0 {
		if err := sys.luaLState.DoString("hook.run('game.result')"); err != nil {
			sys.luaLState.RaiseError("Error executing Lua hook: %s\n%v", "game.result", err.Error())
		}
	}
	if wi.endTimer == -1 {
		cancel := sys.esc || sys.button(wi.keyCancel, -1)
		if cancel || wi.counter == wi.time {
			startFadeOut(wi.fadeOut, m.fadeOut, cancel, m.fadePolicy)
			wi.endTimer = wi.counter + m.fadeOut.timeRemaining
		}
	}

	// Handle state transitions
	if !wi.stateDone && wi.counter >= wi.stateTime {
		m.processStateTransitionsBySide(wi.p1State, wi.p1TeammateState, wi.p2State, wi.p2TeammateState)
		wi.stateDone = true
	}

	// Check if the sequence has ended
	if wi.endTimer != -1 && wi.counter >= wi.endTimer {
		if m.fadeOut != nil {
			m.fadeOut.reset()
		}
		wi.active = false
		if !wi.soundsEnabled {
			sys.noSoundFlg = false
		}
		return
	}

	// Increment counter
	wi.counter++
}

func (wi *MotifWin) draw(m *Motif, layerno int16) {
	// If we initialized via results map, draw that (no hardcoded mode names).
	if wi.resultsScreen != nil {
		bg := wi.resultsBgDef
		rs := wi.resultsScreen
		// Background
		if bg != nil {
			if bg.BgClearColor[0] >= 0 && bg.RectData != nil {
				bg.RectData.Draw(layerno)
			}
			if bg.BGDef != nil && layerno <= 1 {
				bg.BGDef.Draw(int32(layerno), 0, 0, 1)
			}
		}
		// Overlay
		if rs.Overlay.RectData != nil {
			rs.Overlay.RectData.Draw(layerno)
		}
		// Text
		if wi.counter >= rs.WinsText.DisplayTime && rs.WinsText.TextSpriteData != nil {
			rs.WinsText.TextSpriteData.Draw(layerno)
		}
		// Top background
		if bg != nil && bg.BGDef != nil && layerno == 2 {
			bg.BGDef.Draw(int32(layerno), 0, 0, 1)
		}
		return
	}
	// Fallback: normal win screen.
	{
		// Background
		if m.WinBgDef.BgClearColor[0] >= 0 {
			m.WinBgDef.RectData.Draw(layerno)
		}
		if layerno <= 1 {
			m.WinBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
		}
		// Overlay
		m.WinScreen.Overlay.RectData.Draw(layerno)
		// Text
		if wi.counter >= m.WinScreen.WinText.DisplayTime {
			m.WinScreen.WinText.TextSpriteData.Draw(layerno)
		}
		if layerno == 2 {
			m.WinBgDef.BGDef.Draw(int32(layerno), 0, 0, 1)
		}
	}
}
