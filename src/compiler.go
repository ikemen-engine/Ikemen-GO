package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const specialSymbols = " !=<>()|&+-*/%,[]^:;{}#\"\t\r\n"

type expFunc func(out *BytecodeExp, in *string) (BytecodeValue, error)

type scFunc func(is IniSection, sc *StateControllerBase, ihp int8) (StateController, error)

type Compiler struct {
	cmdl             *CommandList
	previousOperator string
	reverseOrder     bool
	norange          bool
	token            string
	playerNo         int
	scmap            map[string]scFunc
	block            *StateBlock
	lines            []string
	i                int
	linechan         chan *string
	vars             map[string]uint8
	funcs            map[string]bytecodeFunction
	funcUsed         map[string]bool
	stateNo          int32
	zssMode          bool
}

func newCompiler() *Compiler {
	c := &Compiler{funcs: make(map[string]bytecodeFunction)}
	c.scmap = map[string]scFunc{
		// Mugen state controllers
		"afterimage":         c.afterImage,
		"afterimagetime":     c.afterImageTime,
		"allpalfx":           c.allPalFX,
		"angleadd":           c.angleAdd,
		"angledraw":          c.angleDraw,
		"anglemul":           c.angleMul,
		"angleset":           c.angleSet,
		"appendtoclipboard":  c.appendToClipboard,
		"assertspecial":      c.assertSpecial,
		"attackdist":         c.attackDist,
		"attackmulset":       c.attackMulSet,
		"bgpalfx":            c.bgPalFX,
		"bindtoparent":       c.bindToParent,
		"bindtoroot":         c.bindToRoot,
		"bindtotarget":       c.bindToTarget,
		"changeanim":         c.changeAnim,
		"changeanim2":        c.changeAnim2,
		"changestate":        c.changeState,
		"clearclipboard":     c.clearClipboard,
		"ctrlset":            c.ctrlSet,
		"defencemulset":      c.defenceMulSet,
		"destroyself":        c.destroySelf,
		"displaytoclipboard": c.displayToClipboard,
		"envcolor":           c.envColor,
		"envshake":           c.envShake,
		"explod":             c.explod,
		"explodbindtime":     c.explodBindTime,
		"fallenvshake":       c.fallEnvShake,
		"forcefeedback":      c.forceFeedback,
		"gamemakeanim":       c.gameMakeAnim,
		"gravity":            c.gravity,
		"helper":             c.helper,
		"hitadd":             c.hitAdd,
		"hitby":              c.hitBy,
		"hitdef":             c.hitDef,
		"hitfalldamage":      c.hitFallDamage,
		"hitfallset":         c.hitFallSet,
		"hitfallvel":         c.hitFallVel,
		"hitoverride":        c.hitOverride,
		"hitvelset":          c.hitVelSet,
		"lifeadd":            c.lifeAdd,
		"lifeset":            c.lifeSet,
		"makedust":           c.makeDust,
		"modifyexplod":       c.modifyExplod,
		"movehitreset":       c.moveHitReset,
		"nothitby":           c.notHitBy,
		"null":               c.null,
		"offset":             c.offset,
		"palfx":              c.palFX,
		"parentvaradd":       c.parentVarAdd,
		"parentvarset":       c.parentVarSet,
		"pause":              c.pause,
		"playerpush":         c.playerPush,
		"playsnd":            c.playSnd,
		"posadd":             c.posAdd,
		"posfreeze":          c.posFreeze,
		"posset":             c.posSet,
		"poweradd":           c.powerAdd,
		"powerset":           c.powerSet,
		"projectile":         c.projectile,
		"remappal":           c.remapPal,
		"removeexplod":       c.removeExplod,
		"removetext":         c.removeText,
		"reversaldef":        c.reversalDef,
		"screenbound":        c.screenBound,
		"selfstate":          c.selfState,
		"sndpan":             c.sndPan,
		"sprpriority":        c.sprPriority,
		"statetypeset":       c.stateTypeSet,
		"stopsnd":            c.stopSnd,
		"superpause":         c.superPause,
		"targetbind":         c.targetBind,
		"targetdrop":         c.targetDrop,
		"targetfacing":       c.targetFacing,
		"targetlifeadd":      c.targetLifeAdd,
		"targetpoweradd":     c.targetPowerAdd,
		"targetstate":        c.targetState,
		"targetveladd":       c.targetVelAdd,
		"targetvelset":       c.targetVelSet,
		"trans":              c.trans,
		"turn":               c.turn,
		"varadd":             c.varAdd,
		"varrandom":          c.varRandom,
		"varrangeset":        c.varRangeSet,
		"varset":             c.varSet,
		"veladd":             c.velAdd,
		"velmul":             c.velMul,
		"velset":             c.velSet,
		"victoryquote":       c.victoryQuote,
		"width":              c.width,
		"zoom":               c.zoom,
		// Ikemen state controllers
		"assertcommand":        c.assertCommand,
		"assertinput":          c.assertInput,
		"camera":               c.cameraCtrl,
		"depth":                c.depth,
		"dialogue":             c.dialogue,
		"dizzypointsadd":       c.dizzyPointsAdd,
		"dizzypointsset":       c.dizzyPointsSet,
		"dizzyset":             c.dizzySet,
		"gethitvarset":         c.getHitVarSet,
		"groundleveloffset":    c.groundLevelOffset,
		"guardbreakset":        c.guardBreakSet,
		"guardpointsadd":       c.guardPointsAdd,
		"guardpointsset":       c.guardPointsSet,
		"height":               c.height,
		"lifebaraction":        c.lifebarAction,
		"loadfile":             c.loadFile,
		"loadstate":            c.loadState,
		"mapadd":               c.mapAdd,
		"mapset":               c.mapSet,
		"matchrestart":         c.matchRestart,
		"modifybgctrl":         c.modifyBGCtrl,
		"modifybgctrl3d":       c.modifyBGCtrl3d,
		"modifybgm":            c.modifyBgm,
		"modifyhitdef":         c.modifyHitDef,
		"modifyplayer":         c.modifyPlayer,
		"modifyprojectile":     c.modifyProjectile,
		"modifyreflection":     c.modifyReflection,
		"modifyreversaldef":    c.modifyReversalDef,
		"modifyshadow":         c.modifyShadow,
		"modifysnd":            c.modifySnd,
		"modifystagebg":        c.modifyStageBG,
		"modifystagevar":       c.modifyStageVar,
		"parentmapadd":         c.parentMapAdd,
		"parentmapset":         c.parentMapSet,
		"playbgm":              c.playBgm,
		"printtoconsole":       c.printToConsole,
		"redlifeadd":           c.redLifeAdd,
		"redlifeset":           c.redLifeSet,
		"remapsprite":          c.remapSprite,
		"rootmapadd":           c.rootMapAdd,
		"rootmapset":           c.rootMapSet,
		"rootvaradd":           c.rootVarAdd,
		"rootvarset":           c.rootVarSet,
		"roundtimeadd":         c.roundTimeAdd,
		"roundtimeset":         c.roundTimeSet,
		"savefile":             c.saveFile,
		"savestate":            c.saveState,
		"scoreadd":             c.scoreAdd,
		"shiftinput":           c.shiftInput,
		"tagin":                c.tagIn,
		"tagout":               c.tagOut,
		"targetadd":            c.targetAdd,
		"targetdizzypointsadd": c.targetDizzyPointsAdd,
		"targetguardpointsadd": c.targetGuardPointsAdd,
		"targetredlifeadd":     c.targetRedLifeAdd,
		"targetscoreadd":       c.targetScoreAdd,
		"teammapadd":           c.teamMapAdd,
		"teammapset":           c.teamMapSet,
		"text":                 c.text,
		"transformclsn":        c.transformClsn,
		"transformsprite":      c.transformSprite,
	}
	return c
}

var triggerMap = map[string]int{
	// Redirections
	"player":      0,
	"parent":      0,
	"root":        0,
	"helper":      0,
	"target":      0,
	"partner":     0,
	"enemy":       0,
	"enemynear":   0,
	"playerid":    0,
	"playerindex": 0,
	"p2":          0,
	"stateowner":  0,
	"helperindex": 0,
	// Mugen triggers
	"abs":               1,
	"acos":              1,
	"ailevel":           1,
	"alive":             1,
	"anim":              1,
	"animelem":          1,
	"animelemno":        1,
	"animelemtime":      1,
	"animexist":         1,
	"animtime":          1,
	"asin":              1,
	"atan":              1,
	"authorname":        1,
	"backedge":          1,
	"backedgebodydist":  1,
	"backedgedist":      1,
	"bottomedge":        1,
	"botboundbodydist":  1,
	"botbounddist":      1,
	"camerapos":         1,
	"camerazoom":        1,
	"canrecover":        1,
	"ceil":              1,
	"command":           1,
	"cond":              1,
	"const":             1,
	"const240p":         1,
	"const480p":         1,
	"const720p":         1,
	"cos":               1,
	"ctrl":              1,
	"drawgame":          1,
	"e":                 1,
	"exp":               1,
	"facing":            1,
	"floor":             1,
	"frontedge":         1,
	"frontedgebodydist": 1,
	"frontedgedist":     1,
	"fvar":              1,
	"gameheight":        1,
	"gametime":          1,
	"gamewidth":         1,
	"gethitvar":         1,
	"hitbyattr":         1,
	"hitcount":          1,
	"hitdefattr":        1,
	"hitfall":           1,
	"hitover":           1,
	"hitpausetime":      1,
	"hitshakeover":      1,
	"hitvel":            1,
	"id":                1,
	"ifelse":            1,
	"inguarddist":       1,
	"ishelper":          1,
	"ishometeam":        1,
	"leftedge":          1,
	"life":              1,
	"lifemax":           1,
	"ln":                1,
	"log":               1,
	"lose":              1,
	"loseko":            1,
	"losetime":          1,
	"matchno":           1,
	"matchover":         1,
	"movecontact":       1,
	"moveguarded":       1,
	"movehit":           1,
	"movereversed":      1,
	"movetype":          1,
	"name":              1,
	"numenemy":          1,
	"numexplod":         1,
	"numhelper":         1,
	"numpartner":        1,
	"numproj":           1,
	"numprojid":         1,
	"numstagebg":        1,
	"numtarget":         1,
	"numtext":           1,
	"p1name":            1,
	"p2bodydist":        1,
	"p2dist":            1,
	"p2life":            1,
	"p2movetype":        1,
	"p2name":            1,
	"p2stateno":         1,
	"p2statetype":       1,
	"p3name":            1,
	"p4name":            1,
	"palno":             1,
	"parentdist":        1,
	"pi":                1,
	"playeridexist":     1,
	"pos":               1,
	"power":             1,
	"powermax":          1,
	"prevstateno":       1,
	"projcanceltime":    1,
	"projcontact":       1,
	"projcontacttime":   1,
	"projguarded":       1,
	"projguardedtime":   1,
	"projhit":           1,
	"projhittime":       1,
	"random":            1,
	"rightedge":         1,
	"rootdist":          1,
	"roundno":           1,
	"roundsexisted":     1,
	"roundstate":        1,
	"roundswon":         1,
	"screenheight":      1,
	"screenpos":         1,
	"screenwidth":       1,
	"selfanimexist":     1,
	"sin":               1,
	"stagebgvar":        1,
	"stagevar":          1,
	"stateno":           1,
	"statetype":         1,
	"sysfvar":           1,
	"sysvar":            1,
	"tan":               1,
	"teammode":          1,
	"teamside":          1,
	"tickspersecond":    1,
	"time":              1,
	"timemod":           1,
	"topedge":           1,
	"topboundbodydist":  1,
	"topbounddist":      1,
	"uniqhitcount":      1,
	"var":               1,
	"vel":               1,
	"win":               1,
	"winko":             1,
	"winperfect":        1,
	"wintime":           1,
	// Ikemen triggers
	"ailevelf":           1,
	"airjumpcount":       1,
	"alpha":              1,
	"angle":              1,
	"xangle":             1,
	"yangle":             1,
	"animelemvar":        1,
	"animlength":         1,
	"animplayerno":       1,
	"spriteplayerno":     1,
	"atan2":              1,
	"attack":             1,
	"bgmvar":             1,
	"clamp":              1,
	"clsnoverlap":        1,
	"clsnvar":            1,
	"combocount":         1,
	"consecutivewins":    1,
	"const1080p":         1,
	"decisiveround":      1,
	"defence":            1,
	"deg":                1,
	"displayname":        1,
	"dizzy":              1,
	"dizzypoints":        1,
	"dizzypointsmax":     1,
	"envshakevar":        1,
	"explodvar":          1,
	"fightscreenstate":   1,
	"fightscreenvar":     1,
	"fighttime":          1,
	"firstattack":        1,
	"float":              1,
	"gamemode":           1,
	"gameoption":         1,
	"gamevar":            1,
	"groundangle":        1,
	"guardbreak":         1,
	"guardcount":         1,
	"guardpoints":        1,
	"guardpointsmax":     1,
	"helperid":           1,
	"helperindexexist":   1,
	"helpername":         1,
	"hitoverridden":      1,
	"ikemenversion":      1,
	"incustomanim":       1,
	"incustomstate":      1,
	"index":              1,
	"indialogue":         1,
	"inputtime":          1,
	"introstate":         1,
	"isasserted":         1,
	"isclsnproxy":        1,
	"ishost":             1,
	"lastplayerid":       1,
	"layerno":            1,
	"lerp":               1,
	"localcoord":         1,
	"map":                1,
	"max":                1,
	"memberno":           1,
	"min":                1,
	"motifstate":         1,
	"movecountered":      1,
	"movehitvar":         1,
	"mugenversion":       1,
	"numplayer":          1,
	"offset":             1,
	"outrostate":         1,
	"p5name":             1,
	"p6name":             1,
	"p7name":             1,
	"p8name":             1,
	"palfxvar":           1,
	"pausetime":          1,
	"physics":            1,
	"playerindexexist":   1,
	"playerno":           1,
	"playernoexist":      1,
	"prevanim":           1,
	"prevmovetype":       1,
	"prevstatetype":      1,
	"projclsnoverlap":    1,
	"projvar":            1,
	"rad":                1,
	"randomrange":        1,
	"ratiolevel":         1,
	"receiveddamage":     1,
	"receivedhits":       1,
	"redlife":            1,
	"reversaldefattr":    1,
	"round":              1,
	"roundtime":          1,
	"runorder":           1,
	"scale":              1,
	"score":              1,
	"scoretotal":         1,
	"selfcommand":        1,
	"selfstatenoexist":   1,
	"sign":               1,
	"soundvar":           1,
	"sprpriority":        1,
	"stagebackedgedist":  1,
	"stageconst":         1,
	"stagefrontedgedist": 1,
	"stagetime":          1,
	"standby":            1,
	"teamleader":         1,
	"teamsize":           1,
	"timeelapsed":        1,
	"timeremaining":      1,
	"timetotal":          1,
	"winhyper":           1,
	"winspecial":         1,
	"xshear":             1,
}

func (c *Compiler) tokenizer(in *string) string {
	return strings.ToLower(c.tokenizerCS(in))
}

// Same but case-sensitive
func (*Compiler) tokenizerCS(in *string) string {
	*in = strings.TrimSpace(*in)
	if len(*in) == 0 {
		return ""
	}
	switch (*in)[0] {
	case '=':
		*in = (*in)[1:]
		return "="
	case ':':
		if len(*in) >= 2 && (*in)[1] == '=' {
			*in = (*in)[2:]
			return ":="
		}
		*in = (*in)[1:]
		return ":"
	case ';':
		*in = (*in)[1:]
		return ";"
	case '!':
		if len(*in) >= 2 && (*in)[1] == '=' {
			*in = (*in)[2:]
			return "!="
		}
		*in = (*in)[1:]
		return "!"
	case '>':
		if len(*in) >= 2 && (*in)[1] == '=' {
			*in = (*in)[2:]
			return ">="
		}
		*in = (*in)[1:]
		return ">"
	case '<':
		if len(*in) >= 2 && (*in)[1] == '=' {
			*in = (*in)[2:]
			return "<="
		}
		*in = (*in)[1:]
		return "<"
	case '~':
		*in = (*in)[1:]
		return "~"
	case '&':
		if len(*in) >= 2 && (*in)[1] == '&' {
			*in = (*in)[2:]
			return "&&"
		}
		*in = (*in)[1:]
		return "&"
	case '^':
		if len(*in) >= 2 && (*in)[1] == '^' {
			*in = (*in)[2:]
			return "^^"
		}
		*in = (*in)[1:]
		return "^"
	case '|':
		if len(*in) >= 2 && (*in)[1] == '|' {
			*in = (*in)[2:]
			return "||"
		}
		*in = (*in)[1:]
		return "|"
	case '+':
		*in = (*in)[1:]
		return "+"
	case '-':
		*in = (*in)[1:]
		return "-"
	case '*':
		if len(*in) >= 2 && (*in)[1] == '*' {
			*in = (*in)[2:]
			return "**"
		}
		*in = (*in)[1:]
		return "*"
	case '/':
		*in = (*in)[1:]
		return "/"
	case '%':
		*in = (*in)[1:]
		return "%"
	case ',':
		*in = (*in)[1:]
		return ","
	case '(':
		*in = (*in)[1:]
		return "("
	case ')':
		*in = (*in)[1:]
		return ")"
	case '[':
		*in = (*in)[1:]
		return "["
	case ']':
		*in = (*in)[1:]
		return "]"
	case '"':
		*in = (*in)[1:]
		return "\""
	case '{':
		*in = (*in)[1:]
		return "{"
	case '}':
		*in = (*in)[1:]
		return "}"
	}
	i, ten := 0, false
	for ; i < len(*in); i++ {
		if (*in)[i] == '.' {
			if ten {
				break
			}
			ten = true
		} else if (*in)[i] < '0' || (*in)[i] > '9' {
			break
		}
	}
	if i > 0 && i < len(*in) && ((*in)[i] == 'e' || (*in)[i] == 'E') {
		j := i + 1
		for i++; i < len(*in); i++ {
			if ((*in)[i] < '0' || (*in)[i] > '9') &&
				(i != j || ((*in)[i] != '-' && (*in)[i] != '+')) {
				break
			}
		}
	}
	if i == 0 {
		i = strings.IndexAny(*in, specialSymbols)
		if i < 0 {
			i = len(*in)
		}
	}
	token := (*in)[:i]
	*in = (*in)[i:]
	return token
}

func (*Compiler) isOperator(token string) int {
	switch token {
	case "", ",", ")", "]":
		return -1
	case "||":
		return 1
	case "^^":
		return 2
	case "&&":
		return 3
	case "|":
		return 4
	case "^":
		return 5
	case "&":
		return 6
	case "=", "!=":
		return 7
	case ">", ">=", "<", "<=":
		return 8
	case "+", "-":
		return 9
	case "*", "/", "%":
		return 10
	case "**":
		return 11
	}
	return 0
}

func (c *Compiler) operator(in *string) error {
	if len(c.previousOperator) > 0 {
		if opp := c.isOperator(c.token); opp <= c.isOperator(c.previousOperator) {
			if opp < 0 || ((!c.reverseOrder || c.token[0] != '(') &&
				(c.token[0] < 'A' || c.token[0] > 'Z') &&
				(c.token[0] < 'a' || c.token[0] > 'z')) {
				return Error("Invalid data: " + c.previousOperator)
			}
			*in = c.token + " " + *in
			c.token = c.previousOperator
			c.previousOperator = ""
			c.norange = true
		}
	}
	return nil
}

func (c *Compiler) integer2(in *string) (int32, error) {
	istr := c.token
	c.token = c.tokenizer(in)
	minus := istr == "-"
	if minus {
		istr = c.token
		c.token = c.tokenizer(in)
	}
	for _, c := range istr {
		if c < '0' || c > '9' {
			return 0, Error(istr + " is not an integer")
		}
	}
	i := Atoi(istr)
	if minus {
		i *= -1
	}
	return i, nil
}

func (c *Compiler) number(token string) BytecodeValue {
	f, err := strconv.ParseFloat(token, 64)
	if err != nil && f == 0 {
		return bvNone()
	}
	if strings.Contains(token, ".") {
		c.reverseOrder = false
		return BytecodeValue{VT_Float, f}
	}
	if strings.ContainsAny(token, "Ee") {
		return bvNone()
	}
	c.reverseOrder = false
	if f > math.MaxInt32 {
		return BytecodeValue{VT_Int, float64(math.MaxInt32)}
	}
	if f < math.MinInt32 {
		return BytecodeValue{VT_Int, float64(math.MinInt32)}
	}
	return BytecodeValue{VT_Int, f}
}

func (c *Compiler) attr(text string, hitdef bool) (int32, error) {
	flg := int32(0)
	att := SplitAndTrim(text, ",")
	for _, a := range att[0] {
		switch a {
		case 'S', 's':
			if hitdef {
				flg = int32(ST_S)
			} else {
				flg |= int32(ST_S)
			}
		case 'C', 'c':
			if hitdef {
				flg = int32(ST_C)
			} else {
				flg |= int32(ST_C)
			}
		case 'A', 'a':
			if hitdef {
				flg = int32(ST_A)
			} else {
				flg |= int32(ST_A)
			}
		default:
			if sys.ignoreMostErrors && a < 128 && (a < 'A' || a > 'Z') &&
				(a < 'a' || a > 'z') {
				return flg, nil
			}
			return 0, Error("Invalid attr value: " + string(a))
		}
	}
	//hitdefflg := flg
	for _, a := range att[1:] {
		l := len(a)
		if sys.ignoreMostErrors && l >= 2 {
			a = strings.TrimSpace(a[:2])
		}
		switch strings.ToLower(a) {
		case "na":
			flg |= int32(AT_NA)
		case "nt":
			flg |= int32(AT_NT)
		case "np":
			flg |= int32(AT_NP)
		case "sa":
			flg |= int32(AT_SA)
		case "st":
			flg |= int32(AT_ST)
		case "sp":
			flg |= int32(AT_SP)
		case "ha":
			flg |= int32(AT_HA)
		case "ht":
			flg |= int32(AT_HT)
		case "hp":
			flg |= int32(AT_HP)
		case "aa":
			flg |= int32(AT_AA)
		case "at":
			flg |= int32(AT_AT)
		case "ap":
			flg |= int32(AT_AP)
		case "n":
			flg |= int32(AT_NA | AT_NT | AT_NP)
		case "s":
			flg |= int32(AT_SA | AT_ST | AT_SP)
		case "h", "a":
			flg |= int32(AT_HA | AT_HT | AT_HP)
		default:
			if sys.ignoreMostErrors {
				//if hitdef {
				//	flg = hitdefflg
				//}
				sys.appendToConsole("WARNING: " + sys.cgi[c.playerNo].nameLow + fmt.Sprintf(": Invalid attr value: "+a+" in state %v ", c.stateNo))
				return flg, nil
			}
			return 0, Error("Invalid attr value: " + a)
		}
		//if i == 0 {
		//	hitdefflg = flg
		//}
		if l > 2 {
			break
		}
	}
	//if hitdef {
	//	flg = hitdefflg
	//}
	return flg, nil
}

func (c *Compiler) trgAttr(in *string) (int32, error) {
	flg := int32(0)
	*in = c.token + *in
	i := strings.IndexAny(*in, specialSymbols)
	var att string
	if i >= 0 {
		att = (*in)[:i]
		*in = strings.TrimSpace((*in)[i:])
	} else {
		att = *in
		*in = ""
	}
	for _, a := range att {
		switch a {
		case 'S', 's':
			flg |= int32(ST_S)
		case 'C', 'c':
			flg |= int32(ST_C)
		case 'A', 'a':
			flg |= int32(ST_A)
		default:
			return 0, Error("Invalid attr value: " + att)
		}
	}
	for len(*in) > 0 && (*in)[0] == ',' {
		oldin := *in
		*in = strings.TrimSpace((*in)[1:])
		i := strings.IndexAny(*in, specialSymbols)
		var att string
		if i >= 0 {
			att = (*in)[:i]
			*in = strings.TrimSpace((*in)[i:])
		} else {
			att = *in
			*in = ""
		}
		switch strings.ToLower(att) {
		case "na":
			flg |= int32(AT_NA)
		case "nt":
			flg |= int32(AT_NT)
		case "np":
			flg |= int32(AT_NP)
		case "sa":
			flg |= int32(AT_SA)
		case "st":
			flg |= int32(AT_ST)
		case "sp":
			flg |= int32(AT_SP)
		case "ha":
			flg |= int32(AT_HA)
		case "ht":
			flg |= int32(AT_HT)
		case "hp":
			flg |= int32(AT_HP)
		case "aa":
			flg |= int32(AT_AA)
		case "at":
			flg |= int32(AT_AT)
		case "ap":
			flg |= int32(AT_AP)
		case "n":
			flg |= int32(AT_NA | AT_NT | AT_NP)
		case "s":
			flg |= int32(AT_SA | AT_ST | AT_SP)
		case "h", "a":
			flg |= int32(AT_HA | AT_HT | AT_HP)
		default:
			*in = oldin
			return flg, nil
		}
	}
	return flg, nil
}

func (c *Compiler) checkOpeningParenthesis(in *string) error {
	if c.tokenizer(in) != "(" {
		return Error("Missing '(' after " + c.token)
	}
	c.token = c.tokenizer(in)
	return nil
}

// Same but case-sensitive
func (c *Compiler) checkOpeningParenthesisCS(in *string) error {
	if c.tokenizerCS(in) != "(" {
		return Error("Missing '(' after " + c.token)
	}
	c.token = c.tokenizerCS(in)
	return nil
}

func (c *Compiler) checkClosingParenthesis() error {
	c.reverseOrder = true
	if c.token != ")" {
		return Error("Missing ')' before " + c.token)
	}
	return nil
}

func (c *Compiler) checkEquality(in *string) (not bool, err error) {
	for {
		c.token = c.tokenizer(in)
		if len(c.token) > 0 {
			if c.token == "!=" {
				not = true
				break
			} else if c.token == "=" {
				break
			} else if sys.ignoreMostErrors {
				if c.token[len(c.token)-1] == '=' {
					break
				}
				continue
			}
		}
		return false, Error("Missing '=' or '!='")
	}
	c.token = c.tokenizer(in)
	return
}

func (c *Compiler) intRange(in *string) (minop OpCode, maxop OpCode,
	min, max int32, err error) {
	switch c.token {
	case "(":
		minop = OC_gt
	case "[":
		minop = OC_ge
	default:
		err = Error("Range missing '[' or '('")
		return
	}
	var intf func(in *string) (int32, error)
	if sys.ignoreMostErrors {
		intf = func(in *string) (int32, error) {
			c.token = c.tokenizer(in)
			minus := false
			for c.token == "-" || c.token == "+" {
				minus = minus || c.token == "-"
				c.token = c.tokenizer(in)
			}
			if len(c.token) == 0 || c.token[0] < '0' || c.token[0] > '9' {
				return 0, Error("Error reading range number")
			}
			i := Atoi(c.token)
			if minus {
				i *= -1
			}
			return i, nil
		}
	} else {
		intf = c.integer2
	}
	if min, err = intf(in); err != nil {
		return
	}
	if sys.ignoreMostErrors {
		if i := strings.Index(*in, ","); i >= 0 {
			c.token = ","
			*in = (*in)[i+1:]
		}
	} else {
		c.token = c.tokenizer(in)
	}
	if c.token != "," {
		err = Error("Range missing ','")
		return
	}
	if max, err = intf(in); err != nil {
		return
	}
	if sys.ignoreMostErrors {
		if i := strings.IndexAny(*in, "])"); i >= 0 {
			c.token = string((*in)[i])
			*in = (*in)[i+1:]
		}
	} else {
		c.token = c.tokenizer(in)
	}
	switch c.token {
	case ")":
		maxop = OC_lt
	case "]":
		maxop = OC_le
	default:
		err = Error("Range missing ']' or ')'")
		return
	}
	c.token = c.tokenizer(in)
	return
}

func (c *Compiler) compareValues(_range bool, in *string) {
	if sys.ignoreMostErrors {
		i := 0
		for ; i < len(*in); i++ {
			if (*in)[i] >= '0' && (*in)[i] <= '9' || (*in)[i] == '-' ||
				_range && ((*in)[i] == '[' || (*in)[i] == '(') {
				break
			}
		}
		*in = (*in)[i:]
	}
	c.token = c.tokenizer(in)
}

func (c *Compiler) evaluateComparison(out *BytecodeExp, in *string,
	required bool) error {
	comma := c.token == ","
	if comma {
		c.token = c.tokenizer(in)
	}
	var opc OpCode
	compare := true
	switch c.token {
	case "<":
		opc = OC_lt
		c.compareValues(false, in)
	case ">":
		opc = OC_gt
		c.compareValues(false, in)
	case "<=":
		opc = OC_le
		c.compareValues(false, in)
	case ">=":
		opc = OC_ge
		c.compareValues(false, in)
	default:
		opc = OC_eq
		switch c.token {
		case "!=":
			opc = OC_ne
		case "=":
		default:
			if required && !comma {
				return Error("No comparison operator" +
					"\n[ECID 1]\n")
			}
			compare = false
		}
		if compare {
			c.compareValues(true, in)
		}
		if c.token == "[" || c.token == "(" {
			minop, maxop, min, max, err := c.intRange(in)
			if err != nil {
				return err
			}
			if opc == OC_ne {
				if minop == OC_gt {
					minop = OC_le
				} else {
					minop = OC_lt
				}
				if maxop == OC_lt {
					minop = OC_ge
				} else {
					minop = OC_gt
				}
			}
			out.append(OC_dup)
			out.appendValue(BytecodeInt(min))
			out.append(minop)
			out.append(OC_swap)
			out.appendValue(BytecodeInt(max))
			out.append(maxop)
			if opc == OC_ne {
				out.append(OC_blor)
			} else {
				out.append(OC_bland)
			}
			c.reverseOrder = comma || compare
			return nil
		}
	}
	ot, oi := c.token, *in
	n, err := c.integer2(in)
	if err != nil {
		if required && !compare {
			return Error("No comparison operator" +
				"\n[ECID 2]\n")
		}
		if compare {
			return err
		}
		n, c.token, *in = 0, ot, oi
	}
	out.appendValue(BytecodeInt(n))
	out.append(opc)
	c.reverseOrder = true
	return nil
}

func (c *Compiler) oneArg(out *BytecodeExp, in *string,
	rd, appendVal bool, defval ...BytecodeValue) (BytecodeValue, error) {
	var be BytecodeExp
	var bv BytecodeValue
	mae := c.token
	if c.token = c.tokenizer(in); c.token != "(" {
		if len(defval) == 0 || defval[0].IsNone() {
			return bvNone(), Error("Missing '(' after " + mae)
		}
		*in = c.token + " " + *in
		bv = defval[0]
	} else {
		c.token = c.tokenizer(in)
		var err error
		if bv, err = c.expBoolOr(&be, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	}
	if appendVal {
		be.appendValue(bv)
		bv = bvNone()
	}
	if rd && len(be) > 0 {
		out.appendI32Op(OC_nordrun, int32(len(be)))
	}
	out.append(be...)
	return bv, nil
}

// Read with two optional arguments
// Currently only for IsHelper
func (c *Compiler) twoOptArg(out *BytecodeExp, in *string,
	rd, appendVal bool, defval ...BytecodeValue) (BytecodeValue, BytecodeValue, error) {

	var be BytecodeExp
	var bv1, bv2 BytecodeValue
	mae := c.token

	// Validate default compiler values
	if len(defval) < 2 || defval[0].IsNone() || defval[1].IsNone() {
		return bvNone(), bvNone(), Error("Missing default arguments for " + mae)
	}

	// Check for opening parenthesis
	if c.token = c.tokenizer(in); c.token != "(" {
		// Put token back where it was and use default values
		*in = c.token + " " + *in
		bv1 = defval[0]
		bv2 = defval[1]
	} else {
		// Parse first argument
		c.token = c.tokenizer(in)
		var err error
		if bv1, err = c.expBoolOr(&be, in); err != nil {
			return bvNone(), bvNone(), err
		}

		// Check for second argument
		if c.token == "," {
			c.token = c.tokenizer(in)
			if bv2, err = c.expBoolOr(&be, in); err != nil {
				return bvNone(), bvNone(), err
			}
		} else {
			// Use default for second argument only
			bv2 = defval[1]
		}

		// Check for closing parenthesis
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), bvNone(), err
		}
	}

	if appendVal {
		be.appendValue(bv1)
		be.appendValue(bv2)
		bv1 = bvNone()
		bv2 = bvNone()
	}

	if rd && len(be) > 0 {
		out.appendI32Op(OC_nordrun, int32(len(be)))
	}

	out.append(be...)
	return bv1, bv2, nil
}

func (c *Compiler) mathFunc(out *BytecodeExp, in *string, rd bool,
	oc OpCode, f func(*BytecodeValue)) (bv BytecodeValue, err error) {
	var be BytecodeExp
	if bv, err = c.oneArg(&be, in, false, false); err != nil {
		return
	}
	if bv.IsNone() {
		if rd {
			out.append(OC_rdreset)
		}
		out.append(be...)
		out.append(oc)
	} else {
		f(&bv)
	}
	return
}

// rd means Redirect
func (c *Compiler) expValue(out *BytecodeExp, in *string,
	rd bool) (BytecodeValue, error) {
	c.reverseOrder, c.norange = true, false
	bv := c.number(c.token)
	if !bv.IsNone() {
		c.token = c.tokenizer(in)
		return bv, nil
	}
	_var := func(sys, f bool) error {
		_, err := c.oneArg(out, in, rd, true)
		if err != nil {
			return err
		}
		var oc OpCode
		c.token = c.tokenizer(in)
		set := c.token == ":="
		if set {
			c.token = c.tokenizer(in)
			var be2 BytecodeExp
			bv2, err := c.expEqne(&be2, in)
			if err != nil {
				return err
			}
			be2.appendValue(bv2)
			if rd {
				out.appendI32Op(OC_nordrun, int32(len(be2)))
			}
			out.append(be2...)
			out.append(OC_st_)
		}
		switch [...]bool{sys, f} {
		case [...]bool{false, false}:
			oc = OC_var
			if set {
				oc = OC_st_var
			}
		case [...]bool{false, true}:
			oc = OC_fvar
			if set {
				oc = OC_st_fvar
			}
		case [...]bool{true, false}:
			oc = OC_sysvar
			if set {
				oc = OC_st_sysvar
			}
		case [...]bool{true, true}:
			oc = OC_sysfvar
			if set {
				oc = OC_st_sysfvar
			}
		}
		out.append(oc)
		return nil
	}
	text := func() error {
		i := strings.Index(*in, "\"")
		if c.token != "\"" || i < 0 {
			return Error("Not enclosed in \"")
		}
		c.token = (*in)[:i]
		*in = (*in)[i+1:]
		return nil
	}
	eqne := func(f func() error) error { // Equal, not equal
		not, err := c.checkEquality(in)
		if err != nil {
			return err
		}
		if err := f(); err != nil {
			return err
		}
		if not {
			out.append(OC_blnot)
		}
		return nil
	}
	eqne2 := func(f func(not bool) error) error {
		not, err := c.checkEquality(in)
		if err != nil {
			return err
		}
		if err := f(not); err != nil {
			return err
		}
		return nil
	}
	nameSub := func(opct, opc OpCode) error {
		return eqne(func() error {
			if err := text(); err != nil {
				return err
			}
			out.append(opct)
			out.appendI32Op(opc, int32(sys.stringPool[c.playerNo].Add(
				strings.ToLower(c.token))))
			return nil
		})
	}
	// Parses a flag. Returns flag and error.
	flagSub := func() (int32, error) {
		flg := int32(0)
		base := c.token
		for _, ch := range base {
			switch ch {
			case 'H', 'h':
				flg |= int32(HF_H)
			case 'L', 'l':
				flg |= int32(HF_L)
			case 'M', 'm':
				flg |= int32(HF_H | HF_L)
			case 'A', 'a':
				flg |= int32(HF_A)
			case 'F', 'f':
				flg |= int32(HF_F)
			case 'D', 'd':
				flg |= int32(HF_D)
			case 'P', 'p':
				flg |= int32(HF_P)
			default:
				return flg, Error("Invalid flags: " + base)
			}
		}
		// peek ahead to see if we have signs in the flag
		if len(*in) > 0 {
			switch (*in)[0] {
			case '+':
				// move forward
				flg |= int32(HF_PLS)
				*in = (*in)[1:]
			case '-':
				// move forward
				flg |= int32(HF_MNS)
				*in = (*in)[1:]
			}
		}
		return flg, nil
	}
	var be1, be2, be3 BytecodeExp
	var bv1, bv2, bv3 BytecodeValue
	var n int32
	var be BytecodeExp
	var opc OpCode
	var err error
	switch c.token {
	case "":
		return bvNone(), Error("Nothing assigned")
	// Redirections without arguments
	case "root", "parent", "p2", "stateowner":
		switch c.token {
		case "root":
			opc = OC_root
		case "parent":
			opc = OC_parent
		case "p2":
			opc = OC_p2
		case "stateowner":
			opc = OC_stateowner
		}
		c.token = c.tokenizer(in)
		if c.token != "," {
			switch opc {
			case OC_partner, OC_enemy, OC_enemynear:
				be1.appendValue(BytecodeInt(0))
			case OC_root:
				return bvNone(), Error("Missing ',' after Root")
			case OC_parent:
				return bvNone(), Error("Missing ',' after Parent")
			case OC_p2:
				return bvNone(), Error("Missing ',' after P2")
			case OC_stateowner:
				return bvNone(), Error("Missing ',' after StateOwner")
			default:
				return bvNone(), Error("Missing ','")
			}
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expValue(&be2, in, true); err != nil {
			return bvNone(), err
		}
		be2.appendValue(bv2)
		out.appendI32Op(opc, int32(len(be2)))
		out.append(be2...)
		return bvNone(), nil
	// Redirections with 1 argument
	case "partner", "enemy", "enemynear", "playerid", "player", "playerindex", "helperindex":
		switch c.token {
		case "player":
			opc = OC_player
		case "partner":
			opc = OC_partner
		case "enemy":
			opc = OC_enemy
		case "enemynear":
			opc = OC_enemynear
		case "playerid":
			opc = OC_playerid
		case "playerindex":
			opc = OC_playerindex
		case "helperindex":
			opc = OC_helperindex
		}
		c.token = c.tokenizer(in)
		if c.token == "(" {
			c.token = c.tokenizer(in)
			if bv1, err = c.expBoolOr(&be1, in); err != nil {
				return bvNone(), err
			}
			if err := c.checkClosingParenthesis(); err != nil {
				return bvNone(), err
			}
			c.token = c.tokenizer(in)
			be1.appendValue(bv1)
		} else {
			switch opc {
			case OC_partner, OC_enemy, OC_enemynear:
				be1.appendValue(BytecodeInt(0)) // Argument is optional for these
			case OC_player:
				return bvNone(), Error("Missing '(' after Player")
			case OC_playerid:
				return bvNone(), Error("Missing '(' after PlayerID")
			case OC_playerindex:
				return bvNone(), Error("Missing '(' after PlayerIndex")
			case OC_helperindex:
				return bvNone(), Error("Missing '(' after HelperIndex")
			default:
				return bvNone(), Error("Missing '('")
			}
		}
		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		out.append(be1...)
		if c.token != "," {
			switch opc {
			case OC_partner:
				return bvNone(), Error("Missing ',' after Partner")
			case OC_enemy:
				return bvNone(), Error("Missing ',' after Enemy")
			case OC_enemynear:
				return bvNone(), Error("Missing ',' after EnemyNear")
			case OC_player:
				return bvNone(), Error("Missing ',' after Player")
			case OC_playerid:
				return bvNone(), Error("Missing ',' after PlayerID")
			case OC_playerindex:
				return bvNone(), Error("Missing ',' after PlayerIndex")
			case OC_helperindex:
				return bvNone(), Error("Missing ',' after HelperIndex")
			default:
				return bvNone(), Error("Missing ','")
			}
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expValue(&be2, in, true); err != nil {
			return bvNone(), err
		}
		be2.appendValue(bv2)
		out.appendI32Op(opc, int32(len(be2)))
		out.append(be2...)
		return bvNone(), nil
	// Redirections with 2 arguments
	case "helper", "target":
		switch c.token {
		case "helper":
			opc = OC_helper
		case "target":
			opc = OC_target
		}
		c.token = c.tokenizer(in)
		if c.token == "(" {
			c.token = c.tokenizer(in)
			// Read the first argument (ID)
			if bv1, err = c.expBoolOr(&be1, in); err != nil {
				return bvNone(), err
			}
			be1.appendValue(bv1)
			// Check if there's a second argument
			if c.token == "," {
				c.token = c.tokenizer(in)
				if bv2, err = c.expBoolOr(&be1, in); err != nil {
					return bvNone(), err
				}
				be1.appendValue(bv2)
			} else {
				// If not, default index to 0
				be1.appendValue(BytecodeInt(0))
			}
			if err := c.checkClosingParenthesis(); err != nil {
				return bvNone(), err
			}
			c.token = c.tokenizer(in)
		} else {
			// Default to ID -1 and index 0 if no arguments are provided
			be1.appendValue(BytecodeInt(-1))
			be1.appendValue(BytecodeInt(0))
		}
		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		out.append(be1...)
		if c.token != "," {
			switch opc {
			case OC_helper:
				return bvNone(), Error("Missing ',' after Helper")
			case OC_target:
				return bvNone(), Error("Missing ',' after Target")
			default:
				return bvNone(), Error("Missing ','")
			}
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expValue(&be2, in, true); err != nil {
			return bvNone(), err
		}
		be2.appendValue(bv2)
		out.appendI32Op(opc, int32(len(be2)))
		out.append(be2...)
		return bvNone(), nil
	case "-":
		if len(*in) > 0 && (((*in)[0] >= '0' && (*in)[0] <= '9') ||
			(*in)[0] == '.') {
			c.token += c.tokenizer(in)
			bv = c.number(c.token)
			if bv.IsNone() {
				return bvNone(), Error("Invalid data: " + c.token)
			}
		} else {
			c.token = c.tokenizer(in)
			if bv, err = c.expValue(&be1, in, false); err != nil {
				return bvNone(), err
			}
			if bv.IsNone() {
				if rd {
					//out.append(OC_rdreset)
					return bvNone(), Error("'-' operator cannot be used within a trigger redirection")
				}
				out.append(be1...)
				out.append(OC_neg)
			} else {
				out.neg(&bv)
			}
			return bv, nil
		}
	case "~":
		c.token = c.tokenizer(in)
		if bv, err = c.expValue(&be1, in, false); err != nil {
			return bvNone(), err
		}
		if bv.IsNone() {
			if rd {
				//out.append(OC_rdreset)
				return bvNone(), Error("'~' operator cannot be used within a trigger redirection")
			}
			out.append(be1...)
			out.append(OC_not)
		} else {
			out.not(&bv)
		}
		return bv, nil
	case "!":
		c.token = c.tokenizer(in)
		if bv, err = c.expValue(&be1, in, false); err != nil {
			return bvNone(), err
		}
		if bv.IsNone() {
			if rd {
				//out.append(OC_rdreset)
				// Ikemen used to allow operators in the middle of a redirection and make them just cancel the redirection
				// Mugen's compiler crashes instead. This seems safer because of user error
				return bvNone(), Error("'!' operator cannot be used within a trigger redirection")
			}
			out.append(be1...)
			out.append(OC_blnot)
		} else {
			out.blnot(&bv)
		}
		return bv, nil
	case "(":
		c.token = c.tokenizer(in)
		if bv, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if bv.IsNone() {
			if rd {
				//out.append(OC_rdreset)
				return bvNone(), Error("Parentheses cannot be used within a trigger redirection")
			}
			out.append(be1...)
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	case "var":
		return bvNone(), _var(false, false)
	case "fvar":
		return bvNone(), _var(false, true)
	case "sysvar":
		return bvNone(), _var(true, false)
	case "sysfvar":
		return bvNone(), _var(true, true)
	case "ifelse", "cond":
		cond := c.token == "cond"
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			if cond {
				return bvNone(), Error("Missing ',' in Cond")
			} else {
				return bvNone(), Error("Missing ',' in IfElse")
			}
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			if cond {
				return bvNone(), Error("Missing ',' in Cond")
			} else {
				return bvNone(), Error("Missing ',' in IfElse")
			}
		}
		c.token = c.tokenizer(in)
		if bv3, err = c.expBoolOr(&be3, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if bv1.IsNone() || bv2.IsNone() || bv3.IsNone() {
			if cond {
				be3.appendValue(bv3)
				be2.appendValue(bv2)
				if len(be3) > int(math.MaxUint8-1) {
					be2.appendI32Op(OC_jmp, int32(len(be3)+1))
				} else {
					be2.append(OC_jmp8, OpCode(len(be3)+1))
				}
				be1.appendValue(bv1)
				if len(be2) > int(math.MaxUint8-1) {
					be1.appendI32Op(OC_jz, int32(len(be2)+1))
				} else {
					be1.append(OC_jz8, OpCode(len(be2)+1))
				}
				be1.append(OC_pop)
				be1.append(be2...)
				be1.append(OC_pop)
				be1.append(be3...)
				if rd {
					out.appendI32Op(OC_run, int32(len(be1)))
				}
				out.append(be1...)
			} else {
				if rd {
					out.append(OC_rdreset)
				}
				out.append(be1...)
				out.appendValue(bv1)
				out.append(be2...)
				out.appendValue(bv2)
				out.append(be3...)
				out.appendValue(bv3)
				out.append(OC_ifelse)
			}
		} else {
			if bv1.ToB() {
				bv = bv2
			} else {
				bv = bv3
			}
		}
	case "ailevel":
		out.append(OC_ailevel)
	case "alive":
		out.append(OC_alive)
	case "anim":
		out.append(OC_anim)
	case "animelemno":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_animelemno)
	case "animelemtime":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_animelemtime)
	case "animexist":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_animexist)
	case "animtime":
		out.append(OC_animtime)
	case "authorname":
		if err := nameSub(OC_const_, OC_const_authorname); err != nil {
			return bvNone(), err
		}
	case "backedge":
		out.append(OC_backedge)
	case "backedgebodydist":
		out.append(OC_backedgebodydist)
	case "backedgedist":
		out.append(OC_backedgedist)
	case "bgmvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		vname := c.token
		c.token = c.tokenizer(in)
		opct := OC_ex_
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		isStr := false
		switch vname {
		case "filename":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_filename
			isStr = true
		case "length":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_length
		case "loop":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_loop
		case "loopcount":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_loopcount
		case "loopend":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_loopend
		case "loopstart":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_loopstart
		case "position":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_position
		case "startposition":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_startposition
		case "volume":
			opct = OC_ex2_
			opc = OC_ex2_bgmvar_volume
		default:
			return bvNone(), Error("Invalid BGMVar argument: " + vname)
		}
		if isStr {
			if err := nameSub(opct, opc); err != nil {
				return bvNone(), err
			}
		} else {
			out.append(opct)
			out.append(opc)
		}
	case "bottomedge":
		out.append(OC_bottomedge)
	case "botboundbodydist":
		out.append(OC_ex2_, OC_ex2_botboundbodydist)
	case "botbounddist":
		out.append(OC_ex2_, OC_ex2_botbounddist)
	case "camerapos":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_camerapos_x)
		case "y":
			out.append(OC_camerapos_y)
		default:
			return bvNone(), Error("Invalid CameraPos argument: " + c.token)
		}
	case "camerazoom":
		out.append(OC_camerazoom)
	case "canrecover":
		out.append(OC_canrecover)
	case "clsnoverlap":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		c1type := c.token
		switch c1type {
		case "clsn1":
			bv1 = BytecodeInt(1)
		case "clsn2":
			bv1 = BytecodeInt(2)
		case "size":
			bv1 = BytecodeInt(3)
		default:
			return bvNone(), Error("Invalid collision box type: " + c1type)
		}
		c.token = c.tokenizer(in)
		if c.token != "," {
			return bvNone(), Error("Missing ',' in ClsnOverlap")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in ClsnOverlap")
		}
		c.token = c.tokenizer(in)
		c2type := c.token
		switch c2type {
		case "clsn1":
			bv3 = BytecodeInt(1)
		case "clsn2":
			bv3 = BytecodeInt(2)
		case "size":
			bv3 = BytecodeInt(3)
		default:
			return bvNone(), Error("Invalid collision box type: " + c2type)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		be2.appendValue(bv2)
		be1.appendValue(bv1)
		if len(be2) > int(math.MaxUint8-1) {
			be1.appendI32Op(OC_jz, int32(len(be2)+1))
		} else {
			be1.append(OC_jz8, OpCode(len(be2)+1))
		}
		be1.append(be2...)
		be1.appendValue(bv3)
		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		out.append(be1...)
		out.append(OC_ex_, OC_ex_clsnoverlap)
	case "clsnvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		ctype := c.token
		switch ctype {
		case "size":
			bv1 = BytecodeInt(3)
		case "clsn1":
			bv1 = BytecodeInt(1)
		case "clsn2":
			bv1 = BytecodeInt(2)
		}
		c.token = c.tokenizer(in)

		if c.token != "," {
			return bvNone(), Error("Missing ',' in ClsnVar")
		}
		c.token = c.tokenizer(in)

		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		c.token = c.tokenizer(in)
		vname := c.token

		switch vname {
		case "back":
			opc = OC_ex2_clsnvar_left
		case "top":
			opc = OC_ex2_clsnvar_top
		case "front":
			opc = OC_ex2_clsnvar_right
		case "bottom":
			opc = OC_ex2_clsnvar_bottom
		default:
			return bvNone(), Error(fmt.Sprint("Invalid ClsnVar argument: %s", vname))
		}
		c.token = c.tokenizer(in)

		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		be2.appendValue(bv2)
		be1.appendValue(bv1)
		if len(be2) > int(math.MaxUint8-1) {
			be1.appendI32Op(OC_jz, int32(len(be2)+1))
		} else {
			be1.append(OC_jz8, OpCode(len(be2)+1))
		}
		be1.append(be2...)
		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		// Just in case anybody else bangs their head against a wall with redirects:
		// it is imperative that the be1.append(opcodetype, opcode) comes after the
		// rd out.appendI32Op(OC_nordrun, int32(len(be1)))
		be1.append(OC_ex2_, opc)
		out.append(be1...)
	case "command", "selfcommand":
		opc := OC_command
		if c.token == "selfcommand" {
			out.append(OC_ex_)
			opc = OC_ex_selfcommand
		}
		if err := eqne(func() error {
			if err := text(); err != nil {
				return err
			}
			_, ok := c.cmdl.Names[c.token]
			if !ok {
				return Error("Command doesn't exist: " + c.token)
			}
			out.appendI32Op(opc, int32(sys.stringPool[c.playerNo].Add(c.token)))
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "const":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_const_)
		switch c.token {
		case "data.life":
			out.append(OC_const_data_life)
		case "data.power":
			out.append(OC_const_data_power)
		case "data.dizzypoints":
			out.append(OC_const_data_dizzypoints)
		case "data.guardpoints":
			out.append(OC_const_data_guardpoints)
		case "data.attack":
			out.append(OC_const_data_attack)
		case "data.defence":
			out.append(OC_const_data_defence)
		case "data.fall.defence_up":
			out.append(OC_const_data_fall_defence_up)
		case "data.fall.defence_mul":
			out.append(OC_const_data_fall_defence_mul)
		case "data.liedown.time":
			out.append(OC_const_data_liedown_time)
		case "data.airjuggle":
			out.append(OC_const_data_airjuggle)
		case "data.sparkno":
			out.append(OC_const_data_sparkno)
		case "data.guard.sparkno":
			out.append(OC_const_data_guard_sparkno)
		case "data.hitsound.channel":
			out.append(OC_const_data_hitsound_channel)
		case "data.guardsound.channel":
			out.append(OC_const_data_guardsound_channel)
		case "data.ko.echo":
			out.append(OC_const_data_ko_echo)
		case "data.volume":
			out.append(OC_const_data_volume)
		case "data.intpersistindex":
			out.append(OC_const_data_intpersistindex)
		case "data.floatpersistindex":
			out.append(OC_const_data_floatpersistindex)
		case "size.xscale":
			out.append(OC_const_size_xscale)
		case "size.yscale":
			out.append(OC_const_size_yscale)
		case "size.ground.back":
			out.append(OC_const_size_ground_back)
		case "size.ground.front":
			out.append(OC_const_size_ground_front)
		case "size.air.back":
			out.append(OC_const_size_air_back)
		case "size.air.front":
			out.append(OC_const_size_air_front)
		case "size.height":
			out.append(OC_const_size_height)
		case "size.attack.dist", "size.attack.dist.width.front": // Optional new syntax for consistency
			out.append(OC_const_size_attack_dist_width_front)
		case "size.attack.dist.width.back":
			out.append(OC_const_size_attack_dist_width_back)
		case "size.attack.dist.height.top":
			out.append(OC_const_size_attack_dist_height_top)
		case "size.attack.dist.height.bottom":
			out.append(OC_const_size_attack_dist_height_bottom)
		case "size.attack.dist.depth.top":
			out.append(OC_const_size_attack_dist_depth_top)
		case "size.attack.dist.depth.bottom":
			out.append(OC_const_size_attack_dist_depth_bottom)
		case "size.attack.depth.top":
			out.append(OC_const_size_attack_depth_top)
		case "size.attack.depth.bottom":
			out.append(OC_const_size_attack_depth_bottom)
		case "size.proj.attack.dist", "size.proj.attack.dist.width.front": // Optional new syntax for consistency
			out.append(OC_const_size_proj_attack_dist_width_front)
		case "size.proj.attack.dist.width.back":
			out.append(OC_const_size_proj_attack_dist_width_back)
		case "size.proj.attack.dist.height.top":
			out.append(OC_const_size_proj_attack_dist_height_top)
		case "size.proj.attack.dist.height.bottom":
			out.append(OC_const_size_proj_attack_dist_height_bottom)
		case "size.proj.attack.dist.depth.top":
			out.append(OC_const_size_proj_attack_dist_depth_top)
		case "size.proj.attack.dist.depth.bottom":
			out.append(OC_const_size_proj_attack_dist_depth_bottom)
		case "size.proj.doscale":
			out.append(OC_const_size_proj_doscale)
		case "size.head.pos.x":
			out.append(OC_const_size_head_pos_x)
		case "size.head.pos.y":
			out.append(OC_const_size_head_pos_y)
		case "size.mid.pos.x":
			out.append(OC_const_size_mid_pos_x)
		case "size.mid.pos.y":
			out.append(OC_const_size_mid_pos_y)
		case "size.shadowoffset":
			out.append(OC_const_size_shadowoffset)
		case "size.draw.offset.x":
			out.append(OC_const_size_draw_offset_x)
		case "size.draw.offset.y":
			out.append(OC_const_size_draw_offset_y)
		case "size.depth.top":
			out.append(OC_const_size_depth_top)
		case "size.depth.bottom":
			out.append(OC_const_size_depth_bottom)
		case "size.weight":
			out.append(OC_const_size_weight)
		case "size.pushfactor":
			out.append(OC_const_size_pushfactor)
		case "velocity.air.gethit.airrecover.add.x":
			out.append(OC_const_velocity_air_gethit_airrecover_add_x)
		case "velocity.air.gethit.airrecover.add.y":
			out.append(OC_const_velocity_air_gethit_airrecover_add_y)
		case "velocity.air.gethit.airrecover.back":
			out.append(OC_const_velocity_air_gethit_airrecover_back)
		case "velocity.air.gethit.airrecover.down":
			out.append(OC_const_velocity_air_gethit_airrecover_down)
		case "velocity.air.gethit.airrecover.fwd":
			out.append(OC_const_velocity_air_gethit_airrecover_fwd)
		case "velocity.air.gethit.airrecover.mul.x":
			out.append(OC_const_velocity_air_gethit_airrecover_mul_x)
		case "velocity.air.gethit.airrecover.mul.y":
			out.append(OC_const_velocity_air_gethit_airrecover_mul_y)
		case "velocity.air.gethit.airrecover.up":
			out.append(OC_const_velocity_air_gethit_airrecover_up)
		case "velocity.air.gethit.groundrecover.x":
			out.append(OC_const_velocity_air_gethit_groundrecover_x)
		case "velocity.air.gethit.groundrecover.y":
			out.append(OC_const_velocity_air_gethit_groundrecover_y)
		case "velocity.air.gethit.ko.add.x":
			out.append(OC_const_velocity_air_gethit_ko_add_x)
		case "velocity.air.gethit.ko.add.y":
			out.append(OC_const_velocity_air_gethit_ko_add_y)
		case "velocity.air.gethit.ko.ymin":
			out.append(OC_const_velocity_air_gethit_ko_ymin)
		case "velocity.airjump.back.x":
			out.append(OC_const_velocity_airjump_back_x)
		case "velocity.airjump.down.x":
			out.append(OC_const_velocity_airjump_down_x)
		case "velocity.airjump.down.y":
			out.append(OC_const_velocity_airjump_down_y)
		case "velocity.airjump.down.z":
			out.append(OC_const_velocity_airjump_down_z)
		case "velocity.airjump.fwd.x":
			out.append(OC_const_velocity_airjump_fwd_x)
		case "velocity.airjump.neu.x":
			out.append(OC_const_velocity_airjump_neu_x)
		case "velocity.airjump.up.x":
			out.append(OC_const_velocity_airjump_up_x)
		case "velocity.airjump.up.y":
			out.append(OC_const_velocity_airjump_up_y)
		case "velocity.airjump.up.z":
			out.append(OC_const_velocity_airjump_up_z)
		case "velocity.airjump.y":
			out.append(OC_const_velocity_airjump_y)
		case "velocity.ground.gethit.ko.add.x":
			out.append(OC_const_velocity_ground_gethit_ko_add_x)
		case "velocity.ground.gethit.ko.add.y":
			out.append(OC_const_velocity_ground_gethit_ko_add_y)
		case "velocity.ground.gethit.ko.xmul":
			out.append(OC_const_velocity_ground_gethit_ko_xmul)
		case "velocity.ground.gethit.ko.ymin":
			out.append(OC_const_velocity_ground_gethit_ko_ymin)
		case "velocity.jump.back.x":
			out.append(OC_const_velocity_jump_back_x)
		case "velocity.jump.down.x":
			out.append(OC_const_velocity_jump_down_x)
		case "velocity.jump.down.y":
			out.append(OC_const_velocity_jump_down_y)
		case "velocity.jump.down.z":
			out.append(OC_const_velocity_jump_down_z)
		case "velocity.jump.fwd.x":
			out.append(OC_const_velocity_jump_fwd_x)
		case "velocity.jump.neu.x":
			out.append(OC_const_velocity_jump_neu_x)
		case "velocity.jump.up.x":
			out.append(OC_const_velocity_jump_up_x)
		case "velocity.jump.up.y":
			out.append(OC_const_velocity_jump_up_y)
		case "velocity.jump.up.z":
			out.append(OC_const_velocity_jump_up_z)
		case "velocity.jump.y":
			out.append(OC_const_velocity_jump_y)
		case "velocity.run.back.x":
			out.append(OC_const_velocity_run_back_x)
		case "velocity.run.back.y":
			out.append(OC_const_velocity_run_back_y)
		case "velocity.run.down.x":
			out.append(OC_const_velocity_run_down_x)
		case "velocity.run.down.y":
			out.append(OC_const_velocity_run_down_y)
		case "velocity.run.down.z":
			out.append(OC_const_velocity_run_down_z)
		case "velocity.run.fwd.x":
			out.append(OC_const_velocity_run_fwd_x)
		case "velocity.run.fwd.y":
			out.append(OC_const_velocity_run_fwd_y)
		case "velocity.run.up.x":
			out.append(OC_const_velocity_run_up_x)
		case "velocity.run.up.y":
			out.append(OC_const_velocity_run_up_y)
		case "velocity.run.up.z":
			out.append(OC_const_velocity_run_up_z)
		case "velocity.runjump.back.x":
			out.append(OC_const_velocity_runjump_back_x)
		case "velocity.runjump.back.y":
			out.append(OC_const_velocity_runjump_back_y)
		case "velocity.runjump.down.x":
			out.append(OC_const_velocity_runjump_down_x)
		case "velocity.runjump.down.y":
			out.append(OC_const_velocity_runjump_down_y)
		case "velocity.runjump.down.z":
			out.append(OC_const_velocity_runjump_down_z)
		case "velocity.runjump.fwd.x":
			out.append(OC_const_velocity_runjump_fwd_x)
		case "velocity.runjump.up.x":
			out.append(OC_const_velocity_runjump_up_x)
		case "velocity.runjump.up.y":
			out.append(OC_const_velocity_runjump_up_y)
		case "velocity.runjump.up.z":
			out.append(OC_const_velocity_runjump_up_z)
		case "velocity.runjump.y":
			out.append(OC_const_velocity_runjump_y)
		case "velocity.walk.back.x":
			out.append(OC_const_velocity_walk_back_x)
		case "velocity.walk.down.x":
			out.append(OC_const_velocity_walk_down_x)
		case "velocity.walk.down.y":
			out.append(OC_const_velocity_walk_down_y)
		case "velocity.walk.down.z":
			out.append(OC_const_velocity_walk_down_z)
		case "velocity.walk.fwd.x":
			out.append(OC_const_velocity_walk_fwd_x)
		case "velocity.walk.up.x":
			out.append(OC_const_velocity_walk_up_x)
		case "velocity.walk.up.y":
			out.append(OC_const_velocity_walk_up_y)
		case "velocity.walk.up.z":
			out.append(OC_const_velocity_walk_up_z)
		case "movement.airjump.num":
			out.append(OC_const_movement_airjump_num)
		case "movement.airjump.height":
			out.append(OC_const_movement_airjump_height)
		case "movement.yaccel":
			out.append(OC_const_movement_yaccel)
		case "movement.stand.friction":
			out.append(OC_const_movement_stand_friction)
		case "movement.crouch.friction":
			out.append(OC_const_movement_crouch_friction)
		case "movement.stand.friction.threshold":
			out.append(OC_const_movement_stand_friction_threshold)
		case "movement.crouch.friction.threshold":
			out.append(OC_const_movement_crouch_friction_threshold)
		case "movement.air.gethit.groundlevel":
			out.append(OC_const_movement_air_gethit_groundlevel)
		case "movement.air.gethit.groundrecover.ground.threshold":
			out.append(OC_const_movement_air_gethit_groundrecover_ground_threshold)
		case "movement.air.gethit.groundrecover.groundlevel":
			out.append(OC_const_movement_air_gethit_groundrecover_groundlevel)
		case "movement.air.gethit.airrecover.threshold":
			out.append(OC_const_movement_air_gethit_airrecover_threshold)
		case "movement.air.gethit.airrecover.yaccel":
			out.append(OC_const_movement_air_gethit_airrecover_yaccel)
		case "movement.air.gethit.trip.groundlevel":
			out.append(OC_const_movement_air_gethit_trip_groundlevel)
		case "movement.down.bounce.offset.x":
			out.append(OC_const_movement_down_bounce_offset_x)
		case "movement.down.bounce.offset.y":
			out.append(OC_const_movement_down_bounce_offset_y)
		case "movement.down.bounce.yaccel":
			out.append(OC_const_movement_down_bounce_yaccel)
		case "movement.down.bounce.groundlevel":
			out.append(OC_const_movement_down_bounce_groundlevel)
		case "movement.down.gethit.offset.x":
			out.append(OC_const_movement_down_gethit_offset_x)
		case "movement.down.gethit.offset.y":
			out.append(OC_const_movement_down_gethit_offset_y)
		case "movement.down.friction.threshold":
			out.append(OC_const_movement_down_friction_threshold)
		default:
			out.appendI32Op(OC_const_constants, int32(sys.stringPool[c.playerNo].Add(
				strings.ToLower(c.token))))
			//return bvNone(), Error("Invalid data: " + c.token)
		}
		*in = strings.TrimSpace(*in)
		if len(*in) == 0 || (!sys.ignoreMostErrors && (*in)[0] != ')') {
			return bvNone(), Error("Missing ')' before " + c.token)
		}
		*in = (*in)[1:]
	case "const240p":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_const240p)
	case "const480p":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_const480p)
	case "const720p":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_const720p)
	case "const1080p":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_const1080p)
	case "ctrl":
		out.append(OC_ctrl)
	case "displayname":
		if err := nameSub(OC_const_, OC_const_displayname); err != nil {
			return bvNone(), err
		}
	case "drawgame":
		out.append(OC_ex_, OC_ex_drawgame)
	case "drawpal":
		c.token = c.tokenizer(in)
		switch c.token {
		case "group":
			out.append(OC_ex2_, OC_ex2_drawpal_group)
		case "index":
			out.append(OC_ex2_, OC_ex2_drawpal_index)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "explodvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in ExplodVar")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in ExplodVar")
		}
		c.token = c.tokenizer(in)

		vname := c.token

		switch vname {
		case "accel":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_explodvar_accel_x
			case "y":
				opc = OC_ex2_explodvar_accel_y
			case "z":
				opc = OC_ex2_explodvar_accel_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ExplodVar accel argument: %s", c.token))
			}
		case "angle":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_explodvar_angle_x
			case "y":
				opc = OC_ex2_explodvar_angle_y
			case ")":
				opc = OC_ex2_explodvar_angle
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ExplodVar angle argument: %s", c.token))
			}
		case "anim":
			opc = OC_ex2_explodvar_anim
		case "animelem":
			opc = OC_ex2_explodvar_animelem
		case "animelemtime":
			opc = OC_ex2_explodvar_animelemtime
		case "animplayerno":
			opc = OC_ex2_explodvar_animplayerno
		case "spriteplayerno":
			opc = OC_ex2_explodvar_spriteplayerno
		case "bindtime":
			opc = OC_ex2_explodvar_bindtime
		case "drawpal":
			c.token = c.tokenizer(in)

			switch c.token {
			case "group":
				opc = OC_ex2_explodvar_drawpal_group
			case "index":
				opc = OC_ex2_explodvar_drawpal_index
			default:
				return bvNone(), Error("Invalid data: " + c.token)
			}
		case "facing":
			opc = OC_ex2_explodvar_facing
		case "friction":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_explodvar_friction_x
			case "y":
				opc = OC_ex2_explodvar_friction_y
			case "z":
				opc = OC_ex2_explodvar_friction_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ExplodVar friction argument: %s", c.token))
			}
		case "id":
			opc = OC_ex2_explodvar_id
		case "layerno":
			opc = OC_ex2_explodvar_layerno
		case "pausemovetime":
			opc = OC_ex2_explodvar_pausemovetime
		case "pos":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_explodvar_pos_x
			case "y":
				opc = OC_ex2_explodvar_pos_y
			case "z":
				opc = OC_ex2_explodvar_pos_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ExplodVar pos argument: %s", c.token))
			}
		case "removetime":
			opc = OC_ex2_explodvar_removetime
		case "scale":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_explodvar_scale_x
			case "y":
				opc = OC_ex2_explodvar_scale_y
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ExplodVar scale argument: %s", c.token))
			}
		case "sprpriority":
			opc = OC_ex2_explodvar_sprpriority
		case "time":
			opc = OC_ex2_explodvar_time
		case "vel":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_explodvar_vel_x
			case "y":
				opc = OC_ex2_explodvar_vel_y
			case "z":
				opc = OC_ex2_explodvar_vel_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ExplodVar vel argument: %s", c.token))
			}
		case "xshear":
			opc = OC_ex2_explodvar_xshear
		default:
			return bvNone(), Error(fmt.Sprint("Invalid ExplodVar angle argument: %s", vname))
		}
		if opc != OC_ex2_explodvar_angle {
			c.token = c.tokenizer(in)

			if err := c.checkClosingParenthesis(); err != nil {
				return bvNone(), err
			}
		}

		be2.appendValue(bv2)
		be1.appendValue(bv1)
		if len(be2) > int(math.MaxUint8-1) {
			be1.appendI32Op(OC_jz, int32(len(be2)+1))
		} else {
			be1.append(OC_jz8, OpCode(len(be2)+1))
		}
		be1.append(be2...)
		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		// Just in case anybody else bangs their head against a wall with redirects:
		// it is imperative that the be1.append(opcodetype, opcode) comes after the
		// rd out.appendI32Op(OC_nordrun, int32(len(be1)))
		be1.append(OC_ex2_, opc)
		out.append(be1...)
	case "facing":
		out.append(OC_facing)
	case "frontedge":
		out.append(OC_frontedge)
	case "frontedgebodydist":
		out.append(OC_frontedgebodydist)
	case "frontedgedist":
		out.append(OC_frontedgedist)
	case "gameheight":
		out.append(OC_gameheight)
	case "gameoption":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_const_)
		out.appendI32Op(OC_const_gameoption, int32(sys.stringPool[c.playerNo].Add(
			strings.ToLower(c.token))))
		*in = strings.TrimSpace(*in)
		if len(*in) == 0 || (!sys.ignoreMostErrors && (*in)[0] != ')') {
			return bvNone(), Error("Missing ')' before " + c.token)
		}
		*in = (*in)[1:]
	case "gametime":
		out.append(OC_gametime)
	case "gamewidth":
		out.append(OC_gamewidth)
	case "gethitvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		opct := OC_ex_
		isFlag := 0
		switch c.token {
		case "animtype":
			opc = OC_ex_gethitvar_animtype
		case "air.animtype":
			opc = OC_ex_gethitvar_air_animtype
		case "ground.animtype":
			opc = OC_ex_gethitvar_ground_animtype
		case "fall.animtype":
			opc = OC_ex_gethitvar_fall_animtype
		case "type":
			opc = OC_ex_gethitvar_type
		case "airtype":
			opc = OC_ex_gethitvar_airtype
		case "groundtype":
			opc = OC_ex_gethitvar_groundtype
		case "damage":
			opc = OC_ex_gethitvar_damage
		case "guardcount":
			opc = OC_ex_gethitvar_guardcount
		case "hitcount":
			opc = OC_ex_gethitvar_hitcount
		case "fallcount":
			opc = OC_ex_gethitvar_fallcount
		case "hitshaketime":
			opc = OC_ex_gethitvar_hitshaketime
		case "hittime":
			opc = OC_ex_gethitvar_hittime
		case "slidetime":
			opc = OC_ex_gethitvar_slidetime
		case "ctrltime":
			opc = OC_ex_gethitvar_ctrltime
		case "recovertime", "down.recovertime": // Added second term for consistency
			opc = OC_ex_gethitvar_down_recovertime
		case "xoff":
			opc = OC_ex_gethitvar_xoff
		case "yoff":
			opc = OC_ex_gethitvar_yoff
		case "zoff":
			opc = OC_ex_gethitvar_zoff
		case "xvel":
			opc = OC_ex_gethitvar_xvel
		case "yvel":
			opc = OC_ex_gethitvar_yvel
		case "zvel":
			opc = OC_ex_gethitvar_zvel
		case "xaccel":
			opc = OC_ex_gethitvar_xaccel
		case "yaccel":
			opc = OC_ex_gethitvar_yaccel
		case "zaccel":
			opc = OC_ex_gethitvar_zaccel
		case "xveladd":
			opc = OC_ex_gethitvar_xveladd
		case "yveladd":
			opc = OC_ex_gethitvar_yveladd
		case "hitid", "chainid":
			opc = OC_ex_gethitvar_chainid
		case "guarded":
			opc = OC_ex_gethitvar_guarded
		case "isbound":
			opc = OC_ex_gethitvar_isbound
		case "fall":
			opc = OC_ex_gethitvar_fall
		case "fall.damage":
			opc = OC_ex_gethitvar_fall_damage
		case "fall.xvel":
			opc = OC_ex_gethitvar_fall_xvel
		case "fall.yvel":
			opc = OC_ex_gethitvar_fall_yvel
		case "fall.zvel":
			opc = OC_ex_gethitvar_fall_zvel
		case "fall.recover":
			opc = OC_ex_gethitvar_fall_recover
		case "fall.time":
			opc = OC_ex_gethitvar_fall_time
		case "fall.recovertime":
			opc = OC_ex_gethitvar_fall_recovertime
		case "fall.kill":
			opc = OC_ex_gethitvar_fall_kill
		case "fall.envshake.time":
			opc = OC_ex_gethitvar_fall_envshake_time
		case "fall.envshake.freq":
			opc = OC_ex_gethitvar_fall_envshake_freq
		case "fall.envshake.ampl":
			opc = OC_ex_gethitvar_fall_envshake_ampl
		case "fall.envshake.phase":
			opc = OC_ex_gethitvar_fall_envshake_phase
		case "fall.envshake.mul":
			opc = OC_ex_gethitvar_fall_envshake_mul
		case "fall.envshake.dir":
			opct = OC_ex2_
			opc = OC_ex2_gethitvar_fall_envshake_dir
		case "attr":
			opc = OC_ex_gethitvar_attr
			isFlag = 1
		case "dizzypoints":
			opc = OC_ex_gethitvar_dizzypoints
		case "guardpoints":
			opc = OC_ex_gethitvar_guardpoints
		case "id":
			opc = OC_ex_gethitvar_id
		case "playerno":
			opc = OC_ex_gethitvar_playerno
		case "redlife":
			opc = OC_ex_gethitvar_redlife
		case "score":
			opc = OC_ex_gethitvar_score
		case "hitdamage":
			opc = OC_ex_gethitvar_hitdamage
		case "guarddamage":
			opc = OC_ex_gethitvar_guarddamage
		case "power":
			opc = OC_ex_gethitvar_power
		case "hitpower":
			opc = OC_ex_gethitvar_hitpower
		case "guardpower":
			opc = OC_ex_gethitvar_guardpower
		case "kill":
			opc = OC_ex_gethitvar_kill
		case "priority":
			opc = OC_ex_gethitvar_priority
		case "facing":
			opc = OC_ex_gethitvar_facing
		case "ground.velocity.x":
			opc = OC_ex_gethitvar_ground_velocity_x
		case "ground.velocity.y":
			opc = OC_ex_gethitvar_ground_velocity_y
		case "ground.velocity.z":
			opc = OC_ex_gethitvar_ground_velocity_z
		case "air.velocity.x":
			opc = OC_ex_gethitvar_air_velocity_x
		case "air.velocity.y":
			opc = OC_ex_gethitvar_air_velocity_y
		case "air.velocity.z":
			opc = OC_ex_gethitvar_air_velocity_z
		case "down.velocity.x":
			opc = OC_ex_gethitvar_down_velocity_x
		case "down.velocity.y":
			opc = OC_ex_gethitvar_down_velocity_y
		case "down.velocity.z":
			opc = OC_ex_gethitvar_down_velocity_z
		case "guard.velocity.x":
			opc = OC_ex_gethitvar_guard_velocity_x
		case "guard.velocity.y":
			opc = OC_ex_gethitvar_guard_velocity_y
		case "guard.velocity.z":
			opc = OC_ex_gethitvar_guard_velocity_z
		case "airguard.velocity.x":
			opc = OC_ex_gethitvar_airguard_velocity_x
		case "airguard.velocity.y":
			opc = OC_ex_gethitvar_airguard_velocity_y
		case "airguard.velocity.z":
			opc = OC_ex_gethitvar_airguard_velocity_z
		case "frame":
			opc = OC_ex_gethitvar_frame
		case "down.recover":
			opc = OC_ex_gethitvar_down_recover
		case "guardflag":
			opc = OC_ex_gethitvar_guardflag
			isFlag = 2
		default:
			return bvNone(), Error("Invalid GetHitVar argument: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		switch isFlag {
		case 1:
			// attr
			hda := func() error {
				if attr, err := c.trgAttr(in); err != nil {
					return err
				} else {
					out.append(opct)
					out.appendI32Op(opc, attr)
				}
				return nil
			}
			if err := eqne(hda); err != nil {
				return bvNone(), err
			}
		case 2:
			// hit/guard flag
			hgf := func() error {
				if flg, err := flagSub(); err != nil {
					return err
				} else {
					out.append(opct)
					out.appendI32Op(opc, flg)
					return nil
				}
			}
			if err := eqne(hgf); err != nil {
				return bvNone(), err
			}
		case 0:
			// no flag
			out.append(opct, opc)
		}
		// no-op (for y/xveladd and fall.envshake.dir)
	case "groundlevel":
		out.append(OC_ex2_, OC_ex2_groundlevel)
	case "guardcount":
		out.append(OC_ex_, OC_ex_guardcount)
	case "helperindexexist":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_helperindexexist)
	case "hitcount":
		out.append(OC_hitcount)
	case "hitbyattr":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if attr, err := c.trgAttr(in); err != nil {
			return bvNone(), err
		} else {
			out.append(OC_ex2_)
			out.appendI32Op(OC_ex2_hitbyattr, attr)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	case "hitdefattr":
		hda := func() error {
			if attr, err := c.trgAttr(in); err != nil {
				return err
			} else {
				out.appendI32Op(OC_hitdefattr, attr)
			}
			return nil
		}
		if err := eqne(hda); err != nil {
			//if sys.cgi[c.playerNo].ikemenverF > 0 || !sys.ignoreMostErrors {
			if c.zssMode || !sys.ignoreMostErrors {
				return bvNone(), err
			}
			sys.appendToConsole("WARNING: " + sys.cgi[c.playerNo].nameLow + fmt.Sprintf(": HitDefAttr Missing '=' or '!=' "+" in state %v ", c.stateNo))
			out.appendValue(BytecodeBool(false))
		}
	case "hitdefvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		param := c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		isFlag := false
		switch param {
		case "guard.dist.depth.bottom":
			opc = OC_ex2_hitdefvar_guard_dist_depth_bottom
		case "guard.dist.depth.top":
			opc = OC_ex2_hitdefvar_guard_dist_depth_top
		case "guard.dist.height.bottom":
			opc = OC_ex2_hitdefvar_guard_dist_height_bottom
		case "guard.dist.height.top":
			opc = OC_ex2_hitdefvar_guard_dist_height_top
		case "guard.dist.width.back":
			opc = OC_ex2_hitdefvar_guard_dist_width_back
		case "guard.dist.width.front":
			opc = OC_ex2_hitdefvar_guard_dist_width_front
		case "guard.pausetime":
			opc = OC_ex2_hitdefvar_guard_pausetime
		case "guard.shaketime":
			opc = OC_ex2_hitdefvar_guard_shaketime
		case "guard.sparkno":
			opc = OC_ex2_hitdefvar_guard_sparkno
		case "guarddamage":
			opc = OC_ex2_hitdefvar_guarddamage
		case "guardflag":
			opc = OC_ex2_hitdefvar_guardflag
			isFlag = true
		case "guardsound.group":
			opc = OC_ex2_hitdefvar_guardsound_group
		case "guardsound.number":
			opc = OC_ex2_hitdefvar_guardsound_number
		case "hitdamage":
			opc = OC_ex2_hitdefvar_hitdamage
		case "hitflag":
			opc = OC_ex2_hitdefvar_hitflag
			isFlag = true
		case "hitsound.group":
			opc = OC_ex2_hitdefvar_hitsound_group
		case "hitsound.number":
			opc = OC_ex2_hitdefvar_hitsound_number
		case "id":
			opc = OC_ex2_hitdefvar_id
		case "p1stateno":
			opc = OC_ex2_hitdefvar_p1stateno
		case "p2stateno":
			opc = OC_ex2_hitdefvar_p2stateno
		case "pausetime":
			opc = OC_ex2_hitdefvar_pausetime
		case "priority":
			opc = OC_ex2_hitdefvar_priority
		case "shaketime":
			opc = OC_ex2_hitdefvar_shaketime
		case "sparkno":
			opc = OC_ex2_hitdefvar_sparkno
		case "sparkx":
			opc = OC_ex2_hitdefvar_sparkx
		case "sparky":
			opc = OC_ex2_hitdefvar_sparky
		default:
			return bvNone(), Error("Invalid HitDefVar argument: " + c.token)
		}
		if isFlag {
			if err := eqne(func() error {
				if flg, err := flagSub(); err != nil {
					return err
				} else {
					out.append(OC_ex2_)
					out.appendI32Op(opc, flg)
					return nil
				}
			}); err != nil {
				return bvNone(), err
			}
		} else {
			out.append(OC_ex2_)
			out.append(opc)
		}
	case "hitfall":
		out.append(OC_hitfall)
	case "hitover":
		out.append(OC_hitover)
	case "hitpausetime":
		out.append(OC_hitpausetime)
	case "hitshakeover":
		out.append(OC_hitshakeover)
	case "hitvel":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_hitvel_x)
		case "y":
			out.append(OC_hitvel_y)
		case "z":
			out.append(OC_hitvel_z)
		default:
			return bvNone(), Error("Invalid HitVel argument: " + c.token)
		}
	case "id":
		out.append(OC_id)
	case "inguarddist":
		out.append(OC_inguarddist)
	case "ishelper":
		if _, _, err := c.twoOptArg(out, in, rd, true, BytecodeInt(math.MinInt32), BytecodeInt(math.MinInt32)); err != nil {
			return bvNone(), err
		}
		out.append(OC_ishelper)
	case "ishometeam":
		out.append(OC_ex_, OC_ex_ishometeam)
	case "isclsnproxy":
		out.append(OC_ex2_, OC_ex2_isclsnproxy)
	case "index":
		out.append(OC_ex2_, OC_ex2_index)
	case "layerno":
		out.append(OC_ex2_, OC_ex2_layerno)
	case "leftedge":
		out.append(OC_leftedge)
	case "life", "p2life":
		if c.token == "p2life" {
			out.appendI32Op(OC_p2, 1)
		}
		out.append(OC_life)
	case "lifemax":
		out.append(OC_lifemax)
	case "lose":
		out.append(OC_ex_, OC_ex_lose)
	case "loseko":
		out.append(OC_ex_, OC_ex_loseko)
	case "losetime":
		out.append(OC_ex_, OC_ex_losetime)
	case "matchno":
		out.append(OC_ex_, OC_ex_matchno)
	case "matchover":
		out.append(OC_ex_, OC_ex_matchover)
	case "movecontact":
		out.append(OC_movecontact)
	case "moveguarded":
		out.append(OC_moveguarded)
	case "movehit":
		out.append(OC_movehit)
	case "movereversed":
		out.append(OC_movereversed)
	case "movetype", "p2movetype", "prevmovetype":
		trname := c.token
		if err := eqne2(func(not bool) error {
			if len(c.token) == 0 {
				return Error(trname + " trigger requires a comparison")
			}
			var mt MoveType
			switch c.token[0] {
			case 'i':
				mt = MT_I
			case 'a':
				mt = MT_A
			case 'h':
				mt = MT_H
			default:
				return Error("Invalid MoveType: " + c.token)
			}
			if trname == "prevmovetype" {
				out.append(OC_ex_, OC_ex_prevmovetype, OpCode(mt>>15))
			} else {
				if trname == "p2movetype" {
					out.appendI32Op(OC_p2, 2+Btoi(not))
				}
				out.append(OC_movetype, OpCode(mt>>15))
			}
			if not {
				out.append(OC_blnot)
			}
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "palfxvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex2_)
		switch c.token {
		case "time":
			out.append(OC_ex2_palfxvar_time)
		case "add.r":
			out.append(OC_ex2_palfxvar_addr)
		case "add.g":
			out.append(OC_ex2_palfxvar_addg)
		case "add.b":
			out.append(OC_ex2_palfxvar_addb)
		case "mul.r":
			out.append(OC_ex2_palfxvar_mulr)
		case "mul.g":
			out.append(OC_ex2_palfxvar_mulg)
		case "mul.b":
			out.append(OC_ex2_palfxvar_mulb)
		case "color":
			out.append(OC_ex2_palfxvar_color)
		case "hue":
			out.append(OC_ex2_palfxvar_hue)
		case "invertall":
			out.append(OC_ex2_palfxvar_invertall)
		case "invertblend":
			out.append(OC_ex2_palfxvar_invertblend)
		case "bg.time":
			out.append(OC_ex2_palfxvar_bg_time)
		case "bg.add.r":
			out.append(OC_ex2_palfxvar_bg_addr)
		case "bg.add.g":
			out.append(OC_ex2_palfxvar_bg_addg)
		case "bg.add.b":
			out.append(OC_ex2_palfxvar_bg_addb)
		case "bg.mul.r":
			out.append(OC_ex2_palfxvar_bg_mulr)
		case "bg.mul.g":
			out.append(OC_ex2_palfxvar_bg_mulg)
		case "bg.mul.b":
			out.append(OC_ex2_palfxvar_bg_mulb)
		case "bg.color":
			out.append(OC_ex2_palfxvar_bg_color)
		case "bg.hue":
			out.append(OC_ex2_palfxvar_bg_hue)
		case "bg.invertall":
			out.append(OC_ex2_palfxvar_bg_invertall)
		case "all.time":
			out.append(OC_ex2_palfxvar_all_time)
		case "all.add.r":
			out.append(OC_ex2_palfxvar_all_addr)
		case "all.add.g":
			out.append(OC_ex2_palfxvar_all_addg)
		case "all.add.b":
			out.append(OC_ex2_palfxvar_all_addb)
		case "all.mul.r":
			out.append(OC_ex2_palfxvar_all_mulr)
		case "all.mul.g":
			out.append(OC_ex2_palfxvar_all_mulg)
		case "all.mul.b":
			out.append(OC_ex2_palfxvar_all_mulb)
		case "all.color":
			out.append(OC_ex2_palfxvar_all_color)
		case "all.hue":
			out.append(OC_ex2_palfxvar_all_hue)
		case "all.invertall":
			out.append(OC_ex2_palfxvar_all_invertall)
		case "all.invertblend":
			out.append(OC_ex2_palfxvar_all_invertblend)
		default:
			return bvNone(), Error("Invalid PalFXVar argument: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	case "name", "p1name", "p2name", "p3name", "p4name", "p5name", "p6name", "p7name", "p8name":
		opc := OC_const_name
		switch c.token {
		case "p2name":
			opc = OC_const_p2name
		case "p3name":
			opc = OC_const_p3name
		case "p4name":
			opc = OC_const_p4name
		case "p5name":
			opc = OC_const_p5name
		case "p6name":
			opc = OC_const_p6name
		case "p7name":
			opc = OC_const_p7name
		case "p8name":
			opc = OC_const_p8name
		}
		if err := nameSub(OC_const_, opc); err != nil {
			return bvNone(), err
		}
	case "numenemy":
		out.append(OC_numenemy)
	case "numexplod":
		if _, err := c.oneArg(out, in, rd, true, BytecodeInt(-1)); err != nil {
			return bvNone(), err
		}
		out.append(OC_numexplod)
	case "numhelper":
		if _, err := c.oneArg(out, in, rd, true, BytecodeInt(-1)); err != nil {
			return bvNone(), err
		}
		out.append(OC_numhelper)
	case "numpartner":
		out.append(OC_numpartner)
	case "numproj":
		out.append(OC_numproj)
	case "numprojid":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_numprojid)
	case "numstagebg":
		if _, err := c.oneArg(out, in, rd, true, BytecodeInt(-1)); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex2_, OC_ex2_numstagebg)
	case "numtarget":
		if _, err := c.oneArg(out, in, rd, true, BytecodeInt(-1)); err != nil {
			return bvNone(), err
		}
		out.append(OC_numtarget)
	case "numtext":
		if _, err := c.oneArg(out, in, rd, true, BytecodeInt(-1)); err != nil {
			return bvNone(), err
		}
		out.append(OC_numtext)
	case "palno":
		out.append(OC_palno)
	case "pos":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_pos_x)
		case "y":
			out.append(OC_pos_y)
		case "z":
			out.append(OC_ex_, OC_ex_pos_z)
		default:
			return bvNone(), Error("Invalid Pos argument: " + c.token)
		}
	case "power":
		out.append(OC_power)
	case "powermax":
		out.append(OC_powermax)
	case "playeridexist":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_playeridexist)
	case "prevanim":
		out.append(OC_ex_, OC_ex_prevanim)
	case "prevstateno":
		out.append(OC_prevstateno)
	case "projcanceltime":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_projcanceltime)
	case "projcontacttime":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_projcontacttime)
	case "projclsnoverlap":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' after index in ProjClsnOverlap")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' after Target_PlayerID in ProjClsnOverlap")
		}
		c.token = c.tokenizer(in)
		c2type := c.token
		var bv3 BytecodeValue
		switch c2type {
		case "clsn1":
			bv3 = BytecodeInt(1)
		case "clsn2":
			bv3 = BytecodeInt(2)
		case "size":
			bv3 = BytecodeInt(3)
		default:
			return bvNone(), Error("Invalid collision box type: " + c2type)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		be2.appendValue(bv2)
		be1.appendValue(bv1)

		be1.append(be2...)
		be1.appendValue(bv3)
		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		out.append(be1...)
		out.append(OC_ex2_, OC_ex2_projclsnoverlap)

	case "projguardedtime":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_projguardedtime)
	case "projhittime":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_projhittime)
	case "projvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in ProjVar")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in ProjVar")
		}
		c.token = c.tokenizer(in)

		vname := c.token
		isFlag := false

		switch vname {
		case "accel":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_projvar_accel_x
			case "y":
				opc = OC_ex2_projvar_accel_y
			case "z":
				opc = OC_ex2_projvar_accel_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ProjVar accel argument: %s", c.token))
			}
		case "angle":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_projvar_projxangle
			case "y":
				opc = OC_ex2_projvar_projyangle
			case ")":
				opc = OC_ex2_projvar_projangle
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ProjVar angle argument: %s", c.token))
			}
		case "anim":
			opc = OC_ex2_projvar_projanim
		case "animelem":
			opc = OC_ex2_projvar_animelem
		case "drawpal":
			c.token = c.tokenizer(in)

			switch c.token {
			case "group":
				opc = OC_ex2_projvar_drawpal_group
			case "index":
				opc = OC_ex2_projvar_drawpal_index
			default:
				return bvNone(), Error("Invalid data: " + c.token)
			}
		case "facing":
			opc = OC_ex2_projvar_facing
		case "guardflag":
			opc = OC_ex2_projvar_guardflag
			isFlag = true
		case "highbound":
			opc = OC_ex2_projvar_highbound
		case "hitflag":
			opc = OC_ex2_projvar_hitflag
			isFlag = true
		case "lowbound":
			opc = OC_ex2_projvar_lowbound
		case "pausemovetime":
			opc = OC_ex2_projvar_pausemovetime
		case "pos":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_projvar_pos_x
			case "y":
				opc = OC_ex2_projvar_pos_y
			case "z":
				opc = OC_ex2_projvar_pos_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ProjVar angle argument: %s", c.token))
			}
		case "projcancelanim":
			opc = OC_ex2_projvar_projcancelanim
		case "projedgebound":
			opc = OC_ex2_projvar_projedgebound
		case "projhitanim":
			opc = OC_ex2_projvar_projhitanim
		case "projhits":
			opc = OC_ex2_projvar_projhits
		case "projhitsmax":
			opc = OC_ex2_projvar_projhitsmax
		case "projid":
			opc = OC_ex2_projvar_projid
		case "projlayerno":
			opc = OC_ex2_projvar_projlayerno
		case "projmisstime":
			opc = OC_ex2_projvar_projmisstime
		case "projpriority":
			opc = OC_ex2_projvar_projpriority
		case "projremanim":
			opc = OC_ex2_projvar_projremanim
		case "projremove":
			opc = OC_ex2_projvar_projremove
		case "projremovetime":
			opc = OC_ex2_projvar_projremovetime
		case "projsprpriority":
			opc = OC_ex2_projvar_projsprpriority
		case "projstagebound":
			opc = OC_ex2_projvar_projstagebound
		case "remvelocity":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_projvar_remvelocity_x
			case "y":
				opc = OC_ex2_projvar_remvelocity_y
			case "z":
				opc = OC_ex2_projvar_remvelocity_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ProjVar remvelocity argument: %s", c.token))
			}
		case "scale":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_projvar_projscale_x
			case "y":
				opc = OC_ex2_projvar_projscale_y
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ProjVar scale argument: %s", c.token))
			}
		case "shadow":
			c.token = c.tokenizer(in)

			switch c.token {
			case "r":
				opc = OC_ex2_projvar_projshadow_r
			case "g":
				opc = OC_ex2_projvar_projshadow_g
			case "b":
				opc = OC_ex2_projvar_projshadow_b
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ProjVar shadow argument: %s", c.token))
			}
		case "supermovetime":
			opc = OC_ex2_projvar_supermovetime
		case "teamside":
			opc = OC_ex2_projvar_teamside
		case "time":
			opc = OC_ex2_projvar_time
		case "vel":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_projvar_vel_x
			case "y":
				opc = OC_ex2_projvar_vel_y
			case "z":
				opc = OC_ex2_projvar_vel_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ProjVar vel argument: %s", c.token))
			}
		case "velmul":
			c.token = c.tokenizer(in)

			switch c.token {
			case "x":
				opc = OC_ex2_projvar_velmul_x
			case "y":
				opc = OC_ex2_projvar_velmul_y
			case "z":
				opc = OC_ex2_projvar_velmul_z
			default:
				return bvNone(), Error(fmt.Sprint("Invalid ProjVar velmul argument: %s", c.token))
			}
		case "xshear":
			opc = OC_ex2_projvar_projxshear
		default:
			return bvNone(), Error(fmt.Sprint("Invalid ProjVar argument: %s", vname))
		}
		if opc != OC_ex2_projvar_projangle {
			c.token = c.tokenizer(in)

			if err := c.checkClosingParenthesis(); err != nil {
				return bvNone(), err
			}
		}

		// If bv1 is ever 0 Ikemen crashes.
		// I do not know why this happens.
		// It happened with clsnVar.
		idx := bv1.ToI()
		if idx >= 0 {
			bv1.SetI(idx + 1)
		}

		bv3 := BytecodeInt(0)
		if isFlag {
			if err := eqne2(func(not bool) error {
				if flg, err := flagSub(); err != nil {
					return err
				} else {
					if not {
						bv3 = BytecodeInt(^flg)
					} else {
						bv3 = BytecodeInt(flg)
					}
				}
				return nil
			}); err != nil {
				return bvNone(), err
			}
		}

		be3.appendValue(bv3)
		be2.appendValue(bv2)
		be1.appendValue(bv1)

		if len(be2) > int(math.MaxUint8-1) {
			be1.appendI32Op(OC_jz, int32(len(be2)+1))
		} else {
			be1.append(OC_jz8, OpCode(len(be2)+1))
		}
		be1.append(be2...)
		be1.append(be3...)

		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		// Just in case anybody else bangs their head against a wall with redirects:
		// it is imperative that the be1.append(opcodetype, opcode) comes after the
		// rd out.appendI32Op(OC_nordrun, int32(len(be1)))
		be1.append(OC_ex2_, opc)
		out.append(be1...)
	case "random":
		out.append(OC_random)
	case "reversaldefattr":
		hda := func() error {
			if attr, err := c.trgAttr(in); err != nil {
				return err
			} else {
				out.append(OC_ex_)
				out.appendI32Op(OC_ex_reversaldefattr, attr)
			}
			return nil
		}
		if err := eqne(hda); err != nil {
			return bvNone(), err
		}
	case "rightedge":
		out.append(OC_rightedge)
	case "runorder":
		out.append(OC_ex2_, OC_ex2_runorder)
	case "roundno":
		out.append(OC_ex_, OC_ex_roundno)
	case "roundsexisted":
		out.append(OC_ex_, OC_ex_roundsexisted)
	case "roundstate":
		out.append(OC_roundstate)
	case "roundswon":
		out.append(OC_roundswon)
	case "introstate":
		out.append(OC_ex2_, OC_ex2_introstate)
	case "outrostate":
		out.append(OC_ex2_, OC_ex2_outrostate)
	case "screenheight":
		out.append(OC_screenheight)
	case "screenpos":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_screenpos_x)
		case "y":
			out.append(OC_screenpos_y)
		default:
			return bvNone(), Error("Invalid ScreenPos argument: " + c.token)
		}
	case "screenwidth":
		out.append(OC_screenwidth)
	case "selfanimexist":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_selfanimexist)
	case "soundvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in SoundVar")
		}
		c.token = c.tokenizer(in)

		vname := c.token
		switch vname {
		case "group":
			opc = OC_ex2_soundvar_group
		case "number":
			opc = OC_ex2_soundvar_number
		case "freqmul":
			opc = OC_ex2_soundvar_freqmul
		case "isplaying":
			opc = OC_ex2_soundvar_isplaying
		case "length":
			opc = OC_ex2_soundvar_length
		case "loopcount":
			opc = OC_ex2_soundvar_loopcount
		case "loopend":
			opc = OC_ex2_soundvar_loopend
		case "loopstart":
			opc = OC_ex2_soundvar_loopstart
		case "pan":
			opc = OC_ex2_soundvar_pan
		case "position":
			opc = OC_ex2_soundvar_position
		case "priority":
			opc = OC_ex2_soundvar_priority
		case "startposition":
			opc = OC_ex2_soundvar_startposition
		case "volumescale":
			opc = OC_ex2_soundvar_volumescale
		default:
			return bvNone(), Error(fmt.Sprint("Invalid SoundVar argument: %s", vname))
		}

		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}

		// If bv1 is ever 0 Ikemen crashes.
		// I do not know why this happens.
		// It happened with clsnVar.
		idx := bv1.ToI()
		if idx >= 0 {
			bv1.SetI(idx + 1)
		}

		be2.appendValue(bv2)
		be1.appendValue(bv1)

		if len(be2) > int(math.MaxUint8-1) {
			be1.appendI32Op(OC_jz, int32(len(be2)+1))
		} else {
			be1.append(OC_jz8, OpCode(len(be2)+1))
		}
		be1.append(be2...)

		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		// Just in case anybody else bangs their head against a wall with redirects:
		// it is imperative that the be1.append(opcodetype, opcode) comes after the
		// rd out.appendI32Op(OC_nordrun, int32(len(be1)))
		be1.append(OC_ex2_, opc)
		out.append(be1...)
	case "stateno", "p2stateno":
		if c.token == "p2stateno" {
			out.appendI32Op(OC_p2, 1)
		}
		out.append(OC_stateno)
	case "statetype", "p2statetype", "prevstatetype":
		trname := c.token
		if err := eqne2(func(not bool) error {
			if len(c.token) == 0 {
				return Error(trname + " trigger requires a comparison")
			}
			var st StateType
			switch c.token[0] {
			case 's':
				st = ST_S
			case 'c':
				st = ST_C
			case 'a':
				st = ST_A
			case 'l':
				st = ST_L
			default:
				return Error("Invalid StateType: " + c.token)
			}
			if trname == "prevstatetype" {
				out.append(OC_ex_, OC_ex_prevstatetype, OpCode(st))
			} else {
				if trname == "p2statetype" {
					out.appendI32Op(OC_p2, 2+Btoi(not))
				}
				out.append(OC_statetype, OpCode(st))
			}
			if not {
				out.append(OC_blnot)
			}
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "stagebgvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		// First argument
		bv1, err := c.expBoolOr(&be1, in)
		if err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in StageBGVar")
		}
		// Second argument
		c.token = c.tokenizer(in)
		bv2, err := c.expBoolOr(&be2, in)
		if err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in StageBGVar")
		}
		// Third argument
		c.token = c.tokenizer(in)
		vname := c.token
		var opc OpCode
		switch vname {
		case "actionno":
			opc = OC_ex2_stagebgvar_actionno
		case "delta.x":
			opc = OC_ex2_stagebgvar_delta_x
		case "delta.y":
			opc = OC_ex2_stagebgvar_delta_y
		case "id":
			opc = OC_ex2_stagebgvar_id
		case "layerno":
			opc = OC_ex2_stagebgvar_layerno
		case "pos.x":
			opc = OC_ex2_stagebgvar_pos_x
		case "pos.y":
			opc = OC_ex2_stagebgvar_pos_y
		case "start.x":
			opc = OC_ex2_stagebgvar_start_x
		case "start.y":
			opc = OC_ex2_stagebgvar_start_y
		case "tile.x":
			opc = OC_ex2_stagebgvar_tile_x
		case "tile.y":
			opc = OC_ex2_stagebgvar_tile_y
		case "velocity.x":
			opc = OC_ex2_stagebgvar_velocity_x
		case "velocity.y":
			opc = OC_ex2_stagebgvar_velocity_y
		default:
			return bvNone(), Error("Invalid StageBGVar argument: " + vname)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		// Output
		be2.appendValue(bv2)
		be1.appendValue(bv1)
		if len(be2) > int(math.MaxUint8-1) {
			be1.appendI32Op(OC_jz, int32(len(be2)+1))
		} else {
			be1.append(OC_jz8, OpCode(len(be2)+1))
		}
		be1.append(be2...)
		be1.append(OC_ex2_, opc)
		out.append(be1...)
	case "stagevar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		svname := c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		isStr := false
		switch svname {
		case "info.author":
			opc = OC_const_stagevar_info_author
			isStr = true
		case "info.displayname":
			opc = OC_const_stagevar_info_displayname
			isStr = true
		case "info.ikemenversion":
			opc = OC_const_stagevar_info_ikemenversion
		case "info.mugenversion":
			opc = OC_const_stagevar_info_mugenversion
		case "info.name":
			opc = OC_const_stagevar_info_name
			isStr = true
		case "camera.boundleft":
			opc = OC_const_stagevar_camera_boundleft
		case "camera.boundright":
			opc = OC_const_stagevar_camera_boundright
		case "camera.boundhigh":
			opc = OC_const_stagevar_camera_boundhigh
		case "camera.boundlow":
			opc = OC_const_stagevar_camera_boundlow
		case "camera.verticalfollow":
			opc = OC_const_stagevar_camera_verticalfollow
		case "camera.floortension":
			opc = OC_const_stagevar_camera_floortension
		case "camera.tensionhigh":
			opc = OC_const_stagevar_camera_tensionhigh
		case "camera.tensionlow":
			opc = OC_const_stagevar_camera_tensionlow
		case "camera.tension":
			opc = OC_const_stagevar_camera_tension
		case "camera.tensionvel":
			opc = OC_const_stagevar_camera_tensionvel
		case "camera.cuthigh":
			opc = OC_const_stagevar_camera_cuthigh
		case "camera.cutlow":
			opc = OC_const_stagevar_camera_cutlow
		case "camera.startzoom":
			opc = OC_const_stagevar_camera_startzoom
		case "camera.zoomout":
			opc = OC_const_stagevar_camera_zoomout
		case "camera.zoomin":
			opc = OC_const_stagevar_camera_zoomin
		case "camera.zoomindelay":
			opc = OC_const_stagevar_camera_zoomindelay
		case "camera.zoominspeed":
			opc = OC_const_stagevar_camera_zoominspeed
		case "camera.zoomoutspeed":
			opc = OC_const_stagevar_camera_zoomoutspeed
		case "camera.yscrollspeed":
			opc = OC_const_stagevar_camera_yscrollspeed
		case "camera.ytension.enable":
			opc = OC_const_stagevar_camera_ytension_enable
		case "camera.autocenter":
			opc = OC_const_stagevar_camera_autocenter
		case "camera.lowestcap":
			opc = OC_const_stagevar_camera_lowestcap
		case "playerinfo.leftbound":
			opc = OC_const_stagevar_playerinfo_leftbound
		case "playerinfo.rightbound":
			opc = OC_const_stagevar_playerinfo_rightbound
		case "playerinfo.topbound":
			opc = OC_const_stagevar_playerinfo_topbound
		case "playerinfo.botbound":
			opc = OC_const_stagevar_playerinfo_botbound
		case "playerinfo.p1startx":
			opc = OC_const_stagevar_playerinfo_p1startx
		case "playerinfo.p2startx":
			opc = OC_const_stagevar_playerinfo_p2startx
		case "playerinfo.p1starty":
			opc = OC_const_stagevar_playerinfo_p1starty
		case "playerinfo.p2starty":
			opc = OC_const_stagevar_playerinfo_p2starty
		case "playerinfo.p1startz":
			opc = OC_const_stagevar_playerinfo_p1startz
		case "playerinfo.p2startz":
			opc = OC_const_stagevar_playerinfo_p2startz
		case "playerinfo.p1facing":
			opc = OC_const_stagevar_playerinfo_p1facing
		case "playerinfo.p2facing":
			opc = OC_const_stagevar_playerinfo_p2facing
		case "scaling.topz":
			opc = OC_const_stagevar_scaling_topz
		case "scaling.botz":
			opc = OC_const_stagevar_scaling_botz
		case "scaling.topscale":
			opc = OC_const_stagevar_scaling_topscale
		case "scaling.botscale":
			opc = OC_const_stagevar_scaling_botscale
		case "bound.screenleft":
			opc = OC_const_stagevar_bound_screenleft
		case "bound.screenright":
			opc = OC_const_stagevar_bound_screenright
		case "stageinfo.localcoord.x":
			opc = OC_const_stagevar_stageinfo_localcoord_x
		case "stageinfo.localcoord.y":
			opc = OC_const_stagevar_stageinfo_localcoord_y
		case "stageinfo.zoffset":
			opc = OC_const_stagevar_stageinfo_zoffset
		case "stageinfo.zoffsetlink":
			opc = OC_const_stagevar_stageinfo_zoffsetlink
		case "stageinfo.xscale":
			opc = OC_const_stagevar_stageinfo_xscale
		case "stageinfo.yscale":
			opc = OC_const_stagevar_stageinfo_yscale
		case "shadow.intensity":
			opc = OC_const_stagevar_shadow_intensity
		case "shadow.color.r":
			opc = OC_const_stagevar_shadow_color_r
		case "shadow.color.g":
			opc = OC_const_stagevar_shadow_color_g
		case "shadow.color.b":
			opc = OC_const_stagevar_shadow_color_b
		case "shadow.yscale":
			opc = OC_const_stagevar_shadow_yscale
		case "shadow.fade.range.begin":
			opc = OC_const_stagevar_shadow_fade_range_begin
		case "shadow.fade.range.end":
			opc = OC_const_stagevar_shadow_fade_range_end
		case "shadow.xshear":
			opc = OC_const_stagevar_shadow_xshear
		case "shadow.offset.x":
			opc = OC_const_stagevar_shadow_offset_x
		case "shadow.offset.y":
			opc = OC_const_stagevar_shadow_offset_y
		case "reflection.intensity":
			opc = OC_const_stagevar_reflection_intensity
		case "reflection.yscale":
			opc = OC_const_stagevar_reflection_yscale
		case "reflection.offset.x":
			opc = OC_const_stagevar_reflection_offset_x
		case "reflection.offset.y":
			opc = OC_const_stagevar_reflection_offset_y
		case "reflection.xshear":
			opc = OC_const_stagevar_reflection_xshear
		case "reflection.color.r":
			opc = OC_const_stagevar_reflection_color_r
		case "reflection.color.g":
			opc = OC_const_stagevar_reflection_color_g
		case "reflection.color.b":
			opc = OC_const_stagevar_reflection_color_b
		default:
			return bvNone(), Error("Invalid StageVar argument: " + svname)
		}
		if isStr {
			if err := nameSub(OC_const_, opc); err != nil {
				return bvNone(), err
			}
		} else {
			out.append(OC_const_)
			out.append(opc)
		}
	case "teammode":
		if err := eqne(func() error {
			if len(c.token) == 0 {
				return Error("TeamMode trigger requires a comparison")
			}
			var tm TeamMode
			switch c.token {
			case "single":
				tm = TM_Single
			case "simul":
				tm = TM_Simul
			case "turns":
				tm = TM_Turns
			case "tag":
				tm = TM_Tag
			default:
				return Error("Invalid TeamMode: " + c.token)
			}
			out.append(OC_teammode, OpCode(tm))
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "teamside":
		out.append(OC_teamside)
	case "tickspersecond":
		out.append(OC_ex_, OC_ex_tickspersecond)
	case "time", "statetime":
		out.append(OC_time)
	case "topedge":
		out.append(OC_topedge)
	case "topboundbodydist":
		out.append(OC_ex2_, OC_ex2_topboundbodydist)
	case "topbounddist":
		out.append(OC_ex2_, OC_ex2_topbounddist)
	case "uniqhitcount":
		out.append(OC_uniqhitcount)
	case "vel":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_vel_x)
		case "y":
			out.append(OC_vel_y)
		case "z":
			out.append(OC_ex_, OC_ex_vel_z)
		default:
			return bvNone(), Error("Invalid Vel argument: " + c.token)
		}
	case "win":
		out.append(OC_ex_, OC_ex_win)
	case "winko":
		out.append(OC_ex_, OC_ex_winko)
	case "wintime":
		out.append(OC_ex_, OC_ex_wintime)
	case "winperfect":
		out.append(OC_ex_, OC_ex_winperfect)
	case "winspecial":
		out.append(OC_ex_, OC_ex_winspecial)
	case "winhyper":
		out.append(OC_ex_, OC_ex_winhyper)
	case "animelem":
		if not, err := c.checkEquality(in); err != nil {
			return bvNone(), err
		} else if not && !sys.ignoreMostErrors {
			return bvNone(), Error("AnimElem doesn't support '!='")
		}
		if c.token == "-" {
			return bvNone(), Error("'-' should not be used")
		}
		if n, err = c.integer2(in); err != nil {
			return bvNone(), err
		}
		if n <= 0 {
			return bvNone(), Error("AnimElem must be greater than 0")
		}
		be1.appendValue(BytecodeInt(n))
		if rd {
			out.appendI32Op(OC_nordrun, int32(len(be1)))
		}
		out.append(be1...)
		out.append(OC_animelemtime)
		if err = c.evaluateComparison(&be, in, false); err != nil {
			return bvNone(), err
		}
		out.append(OC_jsf8, OpCode(len(be)))
		out.append(be...)
		return bv, nil
	case "timemod":
		if not, err := c.checkEquality(in); err != nil {
			return bvNone(), err
		} else if not && !sys.ignoreMostErrors {
			return bvNone(), Error("TimeMod doesn't support '!='")
		}
		if c.token == "-" {
			return bvNone(), Error("'-' should not be used")
		}
		if n, err = c.integer2(in); err != nil {
			return bvNone(), err
		}
		if n <= 0 {
			return bvNone(), Error("TimeMod must be greater than 0")
		}
		out.append(OC_time)
		out.appendValue(BytecodeInt(n))
		out.append(OC_mod)
		if err = c.evaluateComparison(out, in, true); err != nil {
			return bvNone(), err
		}
		return bv, nil
	case "p2dist":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_p2dist_x)
		case "y":
			out.append(OC_ex_, OC_ex_p2dist_y)
		case "z":
			out.append(OC_ex_, OC_ex_p2dist_z)
		default:
			return bvNone(), Error("Invalid P2Dist argument: " + c.token)
		}
	case "p2bodydist":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_p2bodydist_x)
		case "y":
			out.append(OC_ex_, OC_ex_p2bodydist_y)
		case "z":
			out.append(OC_ex_, OC_ex_p2bodydist_z)
		default:
			return bvNone(), Error("Invalid P2BodyDist argument: " + c.token)
		}
	case "rootdist":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_rootdist_x)
		case "y":
			out.append(OC_ex_, OC_ex_rootdist_y)
		case "z":
			out.append(OC_ex_, OC_ex_rootdist_z)
		default:
			return bvNone(), Error("Invalid RootDist argument: " + c.token)
		}
	case "parentdist":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_parentdist_x)
		case "y":
			out.append(OC_ex_, OC_ex_parentdist_y)
		case "z":
			out.append(OC_ex_, OC_ex_parentdist_z)
		default:
			return bvNone(), Error("Invalid ParentDist argument: " + c.token)
		}
	case "pi":
		bv = BytecodeFloat(float32(math.Pi))
	case "e":
		bv = BytecodeFloat(float32(math.E))
	case "abs":
		if bv, err = c.mathFunc(out, in, rd, OC_abs, out.abs); err != nil {
			return bvNone(), err
		}
	case "exp":
		if bv, err = c.mathFunc(out, in, rd, OC_exp, out.exp); err != nil {
			return bvNone(), err
		}
	case "ln":
		if bv, err = c.mathFunc(out, in, rd, OC_ln, out.ln); err != nil {
			return bvNone(), err
		}
	case "log":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in Log")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if bv1.IsNone() || bv2.IsNone() {
			if rd {
				out.append(OC_rdreset)
			}
			out.append(be1...)
			out.appendValue(bv1)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(OC_log)
		} else {
			out.log(&bv1, bv2)
			bv = bv1
		}
	case "cos":
		if bv, err = c.mathFunc(out, in, rd, OC_cos, out.cos); err != nil {
			return bvNone(), err
		}
	case "sin":
		if bv, err = c.mathFunc(out, in, rd, OC_sin, out.sin); err != nil {
			return bvNone(), err
		}
	case "tan":
		if bv, err = c.mathFunc(out, in, rd, OC_tan, out.tan); err != nil {
			return bvNone(), err
		}
	case "acos":
		if bv, err = c.mathFunc(out, in, rd, OC_acos, out.acos); err != nil {
			return bvNone(), err
		}
	case "asin":
		if bv, err = c.mathFunc(out, in, rd, OC_asin, out.asin); err != nil {
			return bvNone(), err
		}
	case "atan":
		if bv, err = c.mathFunc(out, in, rd, OC_atan, out.atan); err != nil {
			return bvNone(), err
		}
	case "floor":
		if bv, err = c.mathFunc(out, in, rd, OC_floor, out.floor); err != nil {
			return bvNone(), err
		}
	case "ceil":
		if bv, err = c.mathFunc(out, in, rd, OC_ceil, out.ceil); err != nil {
			return bvNone(), err
		}
	case "float":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_float)
	case "max":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in Max")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if bv1.IsNone() || bv2.IsNone() {
			if rd {
				out.append(OC_rdreset)
			}
			out.append(be1...)
			out.appendValue(bv1)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(OC_ex_, OC_ex_max)
		} else {
			out.max(&bv1, bv2)
			bv = bv1
		}
	case "min":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in Min")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if bv1.IsNone() || bv2.IsNone() {
			if rd {
				out.append(OC_rdreset)
			}
			out.append(be1...)
			out.appendValue(bv1)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(OC_ex_, OC_ex_min)
		} else {
			out.min(&bv1, bv2)
			bv = bv1
		}
	case "randomrange":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in RandomRange")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if rd {
			out.append(OC_rdreset)
		}
		out.append(be1...)
		out.appendValue(bv1)
		out.append(be2...)
		out.appendValue(bv2)
		out.append(OC_ex_, OC_ex_randomrange)
	case "round":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in Round")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if bv1.IsNone() || bv2.IsNone() {
			if rd {
				out.append(OC_rdreset)
			}
			out.append(be1...)
			out.appendValue(bv1)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(OC_ex_, OC_ex_round)
		} else {
			out.round(&bv1, bv2)
			bv = bv1
		}
	case "clamp":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in Clamp")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in Clamp")
		}
		c.token = c.tokenizer(in)
		if bv3, err = c.expBoolOr(&be3, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if bv1.IsNone() || bv2.IsNone() || bv3.IsNone() {
			if rd {
				out.append(OC_rdreset)
			}
			out.append(be1...)
			out.appendValue(bv1)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(be3...)
			out.appendValue(bv3)
			out.append(OC_ex_, OC_ex_clamp)
		} else {
			out.clamp(&bv1, bv2, bv3)
			bv = bv1
		}
	case "atan2":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in ATan2")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if bv1.IsNone() || bv2.IsNone() {
			if rd {
				out.append(OC_rdreset)
			}
			out.append(be1...)
			out.appendValue(bv1)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(OC_ex_, OC_ex_atan2)
		} else {
			out.atan2(&bv1, bv2)
			bv = bv1
		}
	case "sign":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_sign)
	case "rad":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_rad)
	case "deg":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_deg)
	case "lerp":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in Lerp")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ',' in Lerp")
		}
		c.token = c.tokenizer(in)
		if bv3, err = c.expBoolOr(&be3, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		if bv1.IsNone() || bv2.IsNone() || bv3.IsNone() {
			if rd {
				out.append(OC_rdreset)
			}
			out.append(be1...)
			out.appendValue(bv1)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(be3...)
			out.appendValue(bv3)
			out.append(OC_ex_, OC_ex_lerp)
		} else {
			out.lerp(&bv1, bv2, bv3)
			bv = bv1
		}
	case "ailevelf":
		out.append(OC_ex_, OC_ex_ailevelf)
	case "airjumpcount":
		out.append(OC_ex_, OC_ex_airjumpcount)
	case "animelemvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_)
		switch c.token {
		case "alphadest":
			out.append(OC_ex_animelemvar_alphadest)
		case "angle":
			out.append(OC_ex_animelemvar_angle)
		case "alphasource":
			out.append(OC_ex_animelemvar_alphasource)
		case "group":
			out.append(OC_ex_animelemvar_group)
		case "hflip":
			out.append(OC_ex_animelemvar_hflip)
		case "image":
			out.append(OC_ex_animelemvar_image)
		case "time":
			out.append(OC_ex_animelemvar_time)
		case "vflip":
			out.append(OC_ex_animelemvar_vflip)
		case "xoffset":
			out.append(OC_ex_animelemvar_xoffset)
		case "xscale":
			out.append(OC_ex_animelemvar_xscale)
		case "yoffset":
			out.append(OC_ex_animelemvar_yoffset)
		case "yscale":
			out.append(OC_ex_animelemvar_yscale)
		case "numclsn1":
			out.append(OC_ex_animelemvar_numclsn1)
		case "numclsn2":
			out.append(OC_ex_animelemvar_numclsn2)
		default:
			return bvNone(), Error("Invalid AnimElemVar argument: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	case "animlength":
		out.append(OC_ex_, OC_ex_animlength)
	case "animplayerno":
		out.append(OC_ex_, OC_ex_animplayerno)
	case "spriteplayerno":
		out.append(OC_ex_, OC_ex_spriteplayerno)
	case "attack":
		out.append(OC_ex_, OC_ex_attack)
	case "combocount":
		out.append(OC_ex_, OC_ex_combocount)
	case "consecutivewins":
		out.append(OC_ex_, OC_ex_consecutivewins)
	case "debugmode":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex2_)
		switch c.token {
		case "accel":
			out.append(OC_ex2_debugmode_accel)
		case "clsndisplay":
			out.append(OC_ex2_debugmode_clsndisplay)
		case "debugdisplay":
			out.append(OC_ex2_debugmode_debugdisplay)
		case "lifebarhide":
			out.append(OC_ex2_debugmode_lifebarhide)
		case "wireframedisplay":
			out.append(OC_ex2_debugmode_wireframedisplay)
		case "roundreset":
			out.append(OC_ex2_debugmode_roundreset)
		default:
			return bvNone(), Error("Invalid Debug trigger argument: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	case "decisiveround":
		out.append(OC_ex_, OC_ex_decisiveround)
	case "defence":
		out.append(OC_ex_, OC_ex_defence)
	case "dizzy":
		out.append(OC_ex_, OC_ex_dizzy)
	case "dizzypoints":
		out.append(OC_ex_, OC_ex_dizzypoints)
	case "dizzypointsmax":
		out.append(OC_ex_, OC_ex_dizzypointsmax)
	case "envshakevar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		opct := OC_ex_
		switch c.token {
		case "time":
			opc = OC_ex_envshakevar_time
		case "freq":
			opc = OC_ex_envshakevar_freq
		case "ampl":
			opc = OC_ex_envshakevar_ampl
		case "dir":
			opct = OC_ex2_
			opc = OC_ex2_envshakevar_dir
		default:
			return bvNone(), Error("Invalid EnvShakeVar argument: " + c.token)
		}
		out.append(opct, opc)
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	case "fightscreenstate":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		fssname := c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		switch fssname {
		case "fightdisplay":
			opc = OC_ex2_fightscreenstate_fightdisplay
		case "kodisplay":
			opc = OC_ex2_fightscreenstate_kodisplay
		case "rounddisplay":
			opc = OC_ex2_fightscreenstate_rounddisplay
		case "windisplay":
			opc = OC_ex2_fightscreenstate_windisplay
		default:
			return bvNone(), Error("Invalid FightScreenState argument: " + fssname)
		}
		out.append(OC_ex2_)
		out.append(opc)
	case "fightscreenvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		fsvname := c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		isStr := false
		switch fsvname {
		case "info.author":
			opc = OC_ex_fightscreenvar_info_author
			isStr = true
		case "info.localcoord.x":
			opc = OC_ex_fightscreenvar_info_localcoord_x
		case "info.localcoord.y":
			opc = OC_ex_fightscreenvar_info_localcoord_y
		case "info.name":
			opc = OC_ex_fightscreenvar_info_name
			isStr = true
		case "round.ctrl.time":
			opc = OC_ex_fightscreenvar_round_ctrl_time
		case "round.over.hittime":
			opc = OC_ex_fightscreenvar_round_over_hittime
		case "round.over.time":
			opc = OC_ex_fightscreenvar_round_over_time
		case "round.over.waittime":
			opc = OC_ex_fightscreenvar_round_over_waittime
		case "round.over.wintime":
			opc = OC_ex_fightscreenvar_round_over_wintime
		case "round.slow.time":
			opc = OC_ex_fightscreenvar_round_slow_time
		case "round.start.waittime":
			opc = OC_ex_fightscreenvar_round_start_waittime
		case "round.callfight.time":
			opc = OC_ex_fightscreenvar_round_callfight_time
		case "time.framespercount":
			opc = OC_ex_fightscreenvar_time_framespercount
		default:
			return bvNone(), Error("Invalid FightScreenVar argument: " + fsvname)
		}
		if isStr {
			if err := nameSub(OC_ex_, opc); err != nil {
				return bvNone(), err
			}
		} else {
			out.append(OC_ex_)
			out.append(opc)
		}
	case "fighttime":
		out.append(OC_ex_, OC_ex_fighttime)
	case "firstattack":
		out.append(OC_ex_, OC_ex_firstattack)
	case "gamemode":
		if err := nameSub(OC_ex_, OC_ex_gamemode); err != nil {
			return bvNone(), err
		}
	case "gamevar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		svname := c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		switch svname {
		case "introtime":
			opc = OC_ex2_gamevar_introtime
		case "outrotime":
			opc = OC_ex2_gamevar_outrotime
		case "pausetime":
			opc = OC_ex2_gamevar_pausetime
		case "slowtime":
			opc = OC_ex2_gamevar_slowtime
		case "superpausetime":
			opc = OC_ex2_gamevar_superpausetime
		default:
			return bvNone(), Error("Invalid GameVar argument: " + svname)
		}
		out.append(OC_ex2_)
		out.append(opc)
	case "groundangle":
		out.append(OC_ex_, OC_ex_groundangle)
	case "guardbreak":
		out.append(OC_ex_, OC_ex_guardbreak)
	case "guardpoints":
		out.append(OC_ex_, OC_ex_guardpoints)
	case "guardpointsmax":
		out.append(OC_ex_, OC_ex_guardpointsmax)
	case "helperid":
		out.append(OC_ex_, OC_ex_helperid)
	case "helpername":
		if err := nameSub(OC_ex_, OC_ex_helpername); err != nil {
			return bvNone(), err
		}
	case "hitoverridden":
		out.append(OC_ex_, OC_ex_hitoverridden)
	case "ikemenversion":
		out.append(OC_ex_, OC_ex_ikemenversion)
	case "incustomanim":
		out.append(OC_ex_, OC_ex_incustomanim)
	case "incustomstate":
		out.append(OC_ex_, OC_ex_incustomstate)
	case "indialogue":
		out.append(OC_ex_, OC_ex_indialogue)
	case "inputtime":
		if err := c.checkOpeningParenthesisCS(in); err != nil {
			return bvNone(), err
		}
		key := c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		switch key {
		case "B":
			out.append(OC_ex_, OC_ex_inputtime_B)
		case "D":
			out.append(OC_ex_, OC_ex_inputtime_D)
		case "F":
			out.append(OC_ex_, OC_ex_inputtime_F)
		case "U":
			out.append(OC_ex_, OC_ex_inputtime_U)
		case "L":
			out.append(OC_ex_, OC_ex_inputtime_L)
		case "R":
			out.append(OC_ex_, OC_ex_inputtime_R)
		case "N":
			out.append(OC_ex_, OC_ex_inputtime_N)
		case "a":
			out.append(OC_ex_, OC_ex_inputtime_a)
		case "b":
			out.append(OC_ex_, OC_ex_inputtime_b)
		case "c":
			out.append(OC_ex_, OC_ex_inputtime_c)
		case "x":
			out.append(OC_ex_, OC_ex_inputtime_x)
		case "y":
			out.append(OC_ex_, OC_ex_inputtime_y)
		case "z":
			out.append(OC_ex_, OC_ex_inputtime_z)
		case "s":
			out.append(OC_ex_, OC_ex_inputtime_s)
		case "d":
			out.append(OC_ex_, OC_ex_inputtime_d)
		case "w":
			out.append(OC_ex_, OC_ex_inputtime_w)
		case "m":
			out.append(OC_ex_, OC_ex_inputtime_m)
		default:
			return bvNone(), Error("Invalid InputTime argument: " + key)
		}
	case "isasserted":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_)
		switch c.token {
		// Mugen char flags
		case "invisible":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_invisible))
		case "noairguard":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noairguard))
		case "noautoturn":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noautoturn))
		case "nocrouchguard":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nocrouchguard))
		case "nojugglecheck":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nojugglecheck))
		case "noko":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noko))
		case "noshadow":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noshadow))
		case "nostandguard":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nostandguard))
		case "nowalk":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nowalk))
		case "unguardable":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_unguardable))
		// Mugen global flags
		case "globalnoshadow":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_globalnoshadow))
		case "intro":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_intro))
		case "nobardisplay":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nobardisplay))
		case "nobg":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nobg))
		case "nofg":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nofg))
		case "nokoslow":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nokoslow))
		case "nokosnd":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nokosnd))
		case "nomusic":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nomusic))
		case "roundnotover":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_roundnotover))
		case "timerfreeze":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_timerfreeze))
		// Ikemen char flags
		case "animatehitpause":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_animatehitpause))
		case "animfreeze":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_animfreeze))
		case "autoguard":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_autoguard))
		case "drawunder":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_drawunder))
		case "noaibuttonjam":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noaibuttonjam))
		case "noaicheat":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noaicheat))
		case "noailevel":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noailevel))
		case "noairjump":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noairjump))
		case "nobrake":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nobrake))
		case "nocombodisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nocombodisplay))
		case "nocrouch":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nocrouch))
		case "nodizzypointsdamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nodizzypointsdamage))
		case "nofacedisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofacedisplay))
		case "nofacep2":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofacep2))
		case "nofallcount":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofallcount))
		case "nofalldefenceup":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofalldefenceup))
		case "nofallhitflag":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofallhitflag))
		case "nofastrecoverfromliedown":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofastrecoverfromliedown))
		case "nogetupfromliedown":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nogetupfromliedown))
		case "noguardbardisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noguardbardisplay))
		case "noguarddamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noguarddamage))
		case "noguardko":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noguardko))
		case "noguardpointsdamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noguardpointsdamage))
		case "nohardcodedkeys":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nohardcodedkeys))
		case "nohitdamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nohitdamage))
		case "noinput":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noinput))
		case "nointroreset":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nointroreset))
		case "nojump":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nojump))
		case "nokofall":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nokofall))
		case "nokovelocity":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nokovelocity))
		case "nolifebaraction":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nolifebaraction))
		case "nolifebardisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nolifebardisplay))
		case "nomakedust":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nomakedust))
		case "nonamedisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nonamedisplay))
		case "nopowerbardisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nopowerbardisplay))
		case "noredlifedamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noredlifedamage))
		case "nostand":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nostand))
		case "nostunbardisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nostunbardisplay))
		case "noturntarget":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noturntarget))
		case "nowinicondisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nowinicondisplay))
		case "postroundinput":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_postroundinput))
		case "projtypecollision":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_projtypecollision))
		case "runfirst":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_runfirst))
		case "runlast":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_runlast))
		case "sizepushonly":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_sizepushonly))
		// Ikemen global flags
		case "camerafreeze":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_camerafreeze))
		case "globalnoko":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_globalnoko))
		case "roundfreeze":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_roundfreeze))
		case "roundnotskip":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_roundnotskip))
		case "skipfightdisplay":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_skipfightdisplay))
		case "skipkodisplay":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_skipkodisplay))
		case "skiprounddisplay":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_skiprounddisplay))
		case "skipwindisplay":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_skipwindisplay))
		default:
			return bvNone(), Error("Invalid AssertSpecial flag: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	case "ishost":
		out.append(OC_ex_, OC_ex_ishost)
	case "jugglepoints":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_jugglepoints)
	case "lastplayerid":
		out.append(OC_ex_, OC_ex_lastplayerid)
	case "localcoord":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_localcoord_x)
		case "y":
			out.append(OC_ex_, OC_ex_localcoord_y)
		default:
			return bvNone(), Error("Invalid LocalCoord argument: " + c.token)
		}
	case "map":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		var m string = c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		c.token = c.tokenizer(in)
		if c.token == ":=" {
			c.token = c.tokenizer(in)
			bv2, err := c.expEqne(&be2, in)
			if err != nil {
				return bvNone(), err
			}
			be2.appendValue(bv2)
			if rd {
				out.appendI32Op(OC_nordrun, int32(len(be2)))
			}
			out.append(be2...)
			out.append(OC_st_)
			out.appendI32Op(OC_st_map, int32(sys.stringPool[c.playerNo].Add(strings.ToLower(m))))
		} else {
			out.append(OC_ex_)
			out.appendI32Op(OC_ex_maparray, int32(sys.stringPool[c.playerNo].Add(strings.ToLower(m))))
		}
		return bvNone(), nil
	case "memberno":
		out.append(OC_ex_, OC_ex_memberno)
	case "motifstate":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		msname := c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
		switch msname {
		case "continuescreen":
			opc = OC_ex2_motifstate_continuescreen
		case "victoryscreen":
			opc = OC_ex2_motifstate_victoryscreen
		case "winscreen":
			opc = OC_ex2_motifstate_winscreen
		default:
			return bvNone(), Error("Invalid MotifState argument: " + msname)
		}
		out.append(OC_ex2_)
		out.append(opc)
	case "movecountered":
		out.append(OC_ex_, OC_ex_movecountered)
	case "movehitvar":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_)
		switch c.token {
		case "cornerpush":
			out.append(OC_ex_movehitvar_cornerpush)
		case "frame":
			out.append(OC_ex_movehitvar_frame)
		case "id":
			out.append(OC_ex_movehitvar_id)
		case "overridden":
			out.append(OC_ex_movehitvar_overridden)
		case "playerno":
			out.append(OC_ex_movehitvar_playerno)
		case "sparkx":
			out.append(OC_ex_movehitvar_spark_x)
		case "sparky":
			out.append(OC_ex_movehitvar_spark_y)
		case "uniqhit":
			out.append(OC_ex_movehitvar_uniqhit)
		default:
			return bvNone(), Error("Invalid MoveHitVar argument: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingParenthesis(); err != nil {
			return bvNone(), err
		}
	case "mugenversion":
		out.append(OC_ex_, OC_ex_mugenversion)
	case "numplayer":
		out.append(OC_ex_, OC_ex_numplayer)
	case "pausetime":
		out.append(OC_ex_, OC_ex_pausetime)
	case "physics":
		if err := eqne(func() error {
			if len(c.token) == 0 {
				return Error("Physics trigger requires a comparison")
			}
			var st StateType
			switch c.token[0] {
			case 's':
				st = ST_S
			case 'c':
				st = ST_C
			case 'a':
				st = ST_A
			case 'n':
				st = ST_N
			default:
				return Error("Invalid Physics type: " + c.token)
			}
			out.append(OC_ex_, OC_ex_physics, OpCode(st))
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "playerindexexist":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_playerindexexist)
	case "playerno":
		out.append(OC_ex_, OC_ex_playerno)
	case "playernoexist":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_playernoexist)
	case "ratiolevel":
		out.append(OC_ex_, OC_ex_ratiolevel)
	case "receiveddamage":
		out.append(OC_ex_, OC_ex_receiveddamage)
	case "receivedhits":
		out.append(OC_ex_, OC_ex_receivedhits)
	case "redlife":
		out.append(OC_ex_, OC_ex_redlife)
	case "roundtime":
		out.append(OC_ex_, OC_ex_roundtime)
	case "score":
		out.append(OC_ex_, OC_ex_score)
	case "scoretotal":
		out.append(OC_ex_, OC_ex_scoretotal)
	case "selfstatenoexist":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_selfstatenoexist)
	case "sprpriority":
		out.append(OC_ex_, OC_ex_sprpriority)
	case "stagebackedgedist", "stagebackedge": // Latter is deprecated
		out.append(OC_ex_, OC_ex_stagebackedgedist)
	case "stageconst":
		if err := c.checkOpeningParenthesis(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_const_)
		out.appendI32Op(OC_const_stage_constants, int32(sys.stringPool[c.playerNo].Add(
			strings.ToLower(c.token))))
		*in = strings.TrimSpace(*in)
		if len(*in) == 0 || (!sys.ignoreMostErrors && (*in)[0] != ')') {
			return bvNone(), Error("StageConst missing ')' before " + c.token)
		}
		*in = (*in)[1:]
	case "stagefrontedgedist", "stagefrontedge": // Latter is deprecated
		out.append(OC_ex_, OC_ex_stagefrontedgedist)
	case "stagetime":
		out.append(OC_ex_, OC_ex_stagetime)
	case "standby":
		out.append(OC_ex_, OC_ex_standby)
	case "teamleader":
		out.append(OC_ex_, OC_ex_teamleader)
	case "teamsize":
		out.append(OC_ex_, OC_ex_teamsize)
	case "timeelapsed":
		out.append(OC_ex_, OC_ex_timeelapsed)
	case "timeremaining":
		out.append(OC_ex_, OC_ex_timeremaining)
	case "timetotal":
		out.append(OC_ex_, OC_ex_timetotal)
	case "angle":
		out.append(OC_ex_, OC_ex_angle)
	case "XAngle":
		out.append(OC_ex2_, OC_ex2_angle_x)
	case "YAngle":
		out.append(OC_ex2_, OC_ex2_angle_y)
	case "scale":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_scale_x)
		case "y":
			out.append(OC_ex_, OC_ex_scale_y)
		case "z":
			out.append(OC_ex_, OC_ex_scale_z)
		default:
			return bvNone(), Error("Invalid Scale trigger argument: " + c.token)
		}
	case "offset":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_offset_x)
		case "y":
			out.append(OC_ex_, OC_ex_offset_y)
		default:
			return bvNone(), Error("Invalid Offset trigger argument: " + c.token)
		}
	case "alpha":
		c.token = c.tokenizer(in)
		switch c.token {
		case "source":
			out.append(OC_ex_, OC_ex_alpha_s)
		case "dest":
			out.append(OC_ex_, OC_ex_alpha_d)
		default:
			return bvNone(), Error("Invalid Alpha trigger argument: " + c.token)
		}
	case "xshear":
		out.append(OC_ex2_, OC_ex2_xshear)
	case "=", "!=", ">", ">=", "<", "<=", "&", "&&", "^", "^^", "|", "||",
		"+", "*", "**", "/", "%":
		if !sys.ignoreMostErrors || len(c.previousOperator) > 0 {
			return bvNone(), Error("Invalid data: " + c.token)
		}
		if rd {
			//out.append(OC_rdreset)
			return bvNone(), Error("'" + c.token + "' operator cannot be used within a trigger redirection")
		}
		c.previousOperator = c.token
		c.token = c.tokenizer(in)
		return c.expValue(out, in, false)
	default:
		l := len(c.token)
		if l >= 7 && c.token[:7] == "projhit" || l >= 11 &&
			(c.token[:11] == "projguarded" || c.token[:11] == "projcontact") {
			trname, opc, id := c.token, OC_projhittime, int32(0)
			if trname[:7] == "projhit" {
				id = Atoi(trname[7:])
				trname = trname[:7]
			} else {
				id = Atoi(trname[11:])
				trname = trname[:11]
				if trname == "projguarded" {
					opc = OC_projguardedtime
				} else {
					opc = OC_projcontacttime
				}
			}
			if not, err := c.checkEquality(in); err != nil {
				return bvNone(), err
			} else if not && !sys.ignoreMostErrors {
				return bvNone(), Error(trname + " doesn't support '!='")
			}
			if c.token == "-" {
				return bvNone(), Error("'-' should not be used")
			}
			if n, err = c.integer2(in); err != nil {
				return bvNone(), err
			}
			be1.appendValue(BytecodeInt(id))
			if rd {
				out.appendI32Op(OC_nordrun, int32(len(be1)))
			}
			out.append(be1...)
			out.append(opc)
			out.appendValue(BytecodeInt(0))
			out.append(OC_eq)
			be.append(OC_pop)
			be.appendValue(BytecodeInt(0))
			if err = c.evaluateComparison(&be, in, false); err != nil {
				return bvNone(), err
			}
			out.append(OC_jz8, OpCode(len(be)))
			out.append(be...)
			if n == 0 {
				out.append(OC_blnot)
			}
			return bv, nil
		} else if len(c.token) >= 2 && c.token[0] == '$' && c.token != "$_" {
			vi, ok := c.vars[c.token[1:]]
			if !ok {
				return bvNone(), Error(c.token + " is not defined")
			}
			out.append(OC_localvar, OpCode(vi))
		} else {
			return bvNone(), Error("Invalid data: " + c.token)
		}
	}
	c.token = c.tokenizer(in)
	return bv, nil
}

func (c *Compiler) contiguousOperator(in *string) error {
	*in = strings.TrimSpace(*in)
	if len(*in) > 0 {
		switch (*in)[0] {
		default:
			if len(*in) < 2 || (*in)[:2] != "!=" {
				break
			}
			fallthrough
		case '=', '<', '>', '|', '&', '+', '*', '/', '%', '^':
			return Error("Contiguous operator: " + c.tokenizer(in))
		}
	}
	return nil
}

func (c *Compiler) expPostNot(out *BytecodeExp, in *string) (BytecodeValue, error) {
	bv, err := c.expValue(out, in, false)
	if err != nil {
		return bvNone(), err
	}
	if sys.ignoreMostErrors {
		for c.token == "!" {
			c.reverseOrder = true
			if bv.IsNone() {
				out.append(OC_blnot)
			} else {
				out.blnot(&bv)
			}
			c.token = c.tokenizer(in)
		}
	}
	if len(c.previousOperator) == 0 {
		if opp := c.isOperator(c.token); opp == 0 {
			if !sys.ignoreMostErrors || !c.reverseOrder && c.token == "(" {
				return bvNone(), Error("No comparison operator" +
					"\n" +
					"Token = '" + c.token + "' String = '" + *in + "'" +
					"\n[ECID 3]\n")
			}
			oldtoken, oldin := c.token, *in
			var dummyout BytecodeExp
			if _, err := c.expValue(&dummyout, in, false); err != nil {
				return bvNone(), err
			}
			if c.reverseOrder {
				if c.isOperator(c.token) <= 0 {
					return bvNone(), Error("No comparison operator" +
						"\n[ECID 4]\n")
				}
				if err := c.contiguousOperator(in); err != nil {
					return bvNone(), err
				}
				oldin = oldin[:len(oldin)-len(*in)]
				*in = oldtoken + " " + oldin[:strings.LastIndex(oldin, c.token)] +
					" " + *in
			}
		} else if opp > 0 {
			if err := c.contiguousOperator(in); err != nil {
				return bvNone(), err
			}
		}
	}
	return bv, nil
}

func (c *Compiler) expPow(out *BytecodeExp, in *string) (BytecodeValue, error) {
	bv, err := c.expPostNot(out, in)
	if err != nil {
		return bvNone(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return bvNone(), err
		}
		if c.token == "**" {
			c.token = c.tokenizer(in)
			var be BytecodeExp
			bv2, err := c.expPostNot(&be, in)
			if err != nil {
				return bvNone(), err
			}
			if bv.IsNone() || bv2.IsNone() {
				out.appendValue(bv)
				out.append(be...)
				out.appendValue(bv2)
				out.append(OC_pow)
				bv = bvNone()
			} else {
				out.pow(&bv, bv2, c.playerNo)
			}
		} else {
			break
		}
	}
	return bv, nil
}

func (c *Compiler) expMldv(out *BytecodeExp, in *string) (BytecodeValue, error) {
	bv, err := c.expPow(out, in)
	if err != nil {
		return bvNone(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return bvNone(), err
		}
		switch c.token {
		case "*":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expPow, out.mul, OC_mul)
		case "/":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expPow, out.div, OC_div)
		case "%":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expPow, out.mod, OC_mod)
		default:
			return bv, nil
		}
		if err != nil {
			return bvNone(), err
		}
	}
}

func (c *Compiler) expAdsb(out *BytecodeExp, in *string) (BytecodeValue, error) {
	bv, err := c.expMldv(out, in)
	if err != nil {
		return bvNone(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return bvNone(), err
		}
		switch c.token {
		case "+":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expMldv, out.add, OC_add)
		case "-":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expMldv, out.sub, OC_sub)
		default:
			return bv, nil
		}
		if err != nil {
			return bvNone(), err
		}
	}
}

func (c *Compiler) expGrls(out *BytecodeExp, in *string) (BytecodeValue, error) {
	bv, err := c.expAdsb(out, in)
	if err != nil {
		return bvNone(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return bvNone(), err
		}
		switch c.token {
		case ">":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expAdsb, out.gt, OC_gt)
		case ">=":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expAdsb, out.ge, OC_ge)
		case "<":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expAdsb, out.lt, OC_lt)
		case "<=":
			c.token = c.tokenizer(in)
			err = c.expOneOpSub(out, in, &bv, c.expAdsb, out.le, OC_le)
		default:
			return bv, nil
		}
		if err != nil {
			return bvNone(), err
		}
	}
}

func (c *Compiler) expRange(out *BytecodeExp, in *string,
	bv *BytecodeValue, opc OpCode) (bool, error) {
	open := c.token
	oldin := *in
	c.token = c.tokenizer(in)
	var be2, be3 BytecodeExp
	bv2, err := c.expBoolOr(&be2, in)
	if err != nil {
		return false, err
	}
	if c.token != "," {
		if open != "(" {
			return false, Error("Missing ','")
		}
		if err := c.checkClosingParenthesis(); err != nil {
			return false, err
		}
		c.token = open
		*in = oldin
		return false, nil
	}
	c.token = c.tokenizer(in)
	bv3, err := c.expBoolOr(&be3, in)
	if err != nil {
		return false, err
	}
	close := c.token
	if close != "]" && close != ")" {
		return false, Error("Missing ']' or ')'")
	}
	c.token = c.tokenizer(in)
	if bv.IsNone() || bv2.IsNone() || bv3.IsNone() {
		var op1, op2, op3 OpCode
		if opc == OC_ne {
			if open == "(" {
				op1 = OC_le
			} else {
				op1 = OC_lt
			}
			if close == ")" {
				op2 = OC_ge
			} else {
				op2 = OC_gt
			}
			op3 = OC_blor
		} else {
			if open == "(" {
				op1 = OC_gt
			} else {
				op1 = OC_ge
			}
			if close == ")" {
				op2 = OC_lt
			} else {
				op2 = OC_le
			}
			op3 = OC_bland
		}
		out.appendValue(*bv)
		out.append(OC_dup)
		out.append(be2...)
		out.appendValue(bv2)
		out.append(op1)
		out.append(OC_swap)
		out.append(be3...)
		out.appendValue(bv3)
		out.append(op2)
		out.append(op3)
		*bv = bvNone()
	} else {
		tmp := *bv
		if open == "(" {
			out.gt(&tmp, bv2)
		} else {
			out.ge(&tmp, bv2)
		}
		if close == ")" {
			out.lt(bv, bv3)
		} else {
			out.le(bv, bv3)
		}
		bv.SetB(tmp.ToB() && bv.ToB())
		if opc == OC_ne {
			bv.SetB(!bv.ToB())
		}
	}
	return true, nil
}

func (c *Compiler) expEqne(out *BytecodeExp, in *string) (BytecodeValue, error) {
	bv, err := c.expGrls(out, in)
	if err != nil {
		return bvNone(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return bvNone(), err
		}
		var opc OpCode
		switch c.token {
		case "=":
			opc = OC_eq
		case "!=":
			opc = OC_ne
		default:
			return bv, nil
		}
		c.token = c.tokenizer(in)
		switch c.token {
		case "[", "(":
			if !c.norange {
				if ok, err := c.expRange(out, in, &bv, opc); err != nil {
					return bvNone(), err
				} else if ok {
					break
				}
			}
			fallthrough
		default:
			switch opc {
			case OC_eq:
				err = c.expOneOpSub(out, in, &bv, c.expGrls, out.eq, opc)
			case OC_ne:
				err = c.expOneOpSub(out, in, &bv, c.expGrls, out.ne, opc)
			}
			if err != nil {
				return bvNone(), err
			}
		}
	}
}

func (*Compiler) expOneOpSub(out *BytecodeExp, in *string, bv *BytecodeValue,
	ef expFunc, opf func(v1 *BytecodeValue, v2 BytecodeValue),
	opc OpCode) error {
	var be BytecodeExp
	bv2, err := ef(&be, in)
	if err != nil {
		return err
	}
	if bv.IsNone() || bv2.IsNone() {
		out.appendValue(*bv)
		out.append(be...)
		out.appendValue(bv2)
		out.append(opc)
		*bv = bvNone()
	} else {
		opf(bv, bv2)
	}
	return nil
}

func (c *Compiler) expOneOp(out *BytecodeExp, in *string, ef expFunc,
	opt string, opf func(v1 *BytecodeValue, v2 BytecodeValue),
	opc OpCode) (BytecodeValue, error) {
	bv, err := ef(out, in)
	if err != nil {
		return bvNone(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return bvNone(), err
		}
		if c.token == opt {
			c.token = c.tokenizer(in)
			if err := c.expOneOpSub(out, in, &bv, ef, opf, opc); err != nil {
				return bvNone(), err
			}
		} else {
			return bv, nil
		}
	}
}

func (c *Compiler) expAnd(out *BytecodeExp, in *string) (BytecodeValue, error) {
	return c.expOneOp(out, in, c.expEqne, "&", out.and, OC_and)
}

func (c *Compiler) expXor(out *BytecodeExp, in *string) (BytecodeValue, error) {
	return c.expOneOp(out, in, c.expAnd, "^", out.xor, OC_xor)
}

func (c *Compiler) expOr(out *BytecodeExp, in *string) (BytecodeValue, error) {
	return c.expOneOp(out, in, c.expXor, "|", out.or, OC_or)
}

func (c *Compiler) expBoolAnd(out *BytecodeExp, in *string) (BytecodeValue, error) {
	if c.block != nil {
		return c.expOneOp(out, in, c.expOr, "&&", out.bland, OC_bland)
	}
	bv, err := c.expOr(out, in)
	if err != nil {
		return bvNone(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return bvNone(), err
		}
		if c.token == "&&" {
			c.token = c.tokenizer(in)
			var be BytecodeExp
			bv2, err := c.expBoolAnd(&be, in)
			if err != nil {
				return bvNone(), err
			}
			if bv.IsNone() || bv2.IsNone() {
				out.appendValue(bv)
				be.appendValue(bv2)
				if len(be) > int(math.MaxUint8-1) {
					out.appendI32Op(OC_jz, int32(len(be)+1))
				} else {
					out.append(OC_jz8, OpCode(len(be)+1))
				}
				out.append(OC_pop)
				out.append(be...)
				bv = bvNone()
			} else {
				out.bland(&bv, bv2)
			}
		} else {
			break
		}
	}
	return bv, nil
}

func (c *Compiler) expBoolXor(out *BytecodeExp, in *string) (BytecodeValue, error) {
	return c.expOneOp(out, in, c.expBoolAnd, "^^", out.blxor, OC_blxor)
}

func (c *Compiler) expBoolOr(out *BytecodeExp, in *string) (BytecodeValue, error) {
	defer func(omp string) { c.previousOperator = omp }(c.previousOperator)
	if c.block != nil {
		return c.expOneOp(out, in, c.expBoolXor, "||", out.blor, OC_blor)
	}
	bv, err := c.expBoolXor(out, in)
	if err != nil {
		return bvNone(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return bvNone(), err
		}
		if c.token == "||" {
			c.token = c.tokenizer(in)
			var be BytecodeExp
			bv2, err := c.expBoolOr(&be, in)
			if err != nil {
				return bvNone(), err
			}
			if bv.IsNone() || bv2.IsNone() {
				out.appendValue(bv)
				be.appendValue(bv2)
				if len(be) > int(math.MaxUint8-1) {
					out.appendI32Op(OC_jnz, int32(len(be)+1))
				} else {
					out.append(OC_jnz8, OpCode(len(be)+1))
				}
				out.append(OC_pop)
				out.append(be...)
				bv = bvNone()
			} else {
				out.blor(&bv, bv2)
			}
		} else {
			break
		}
	}
	return bv, nil
}

func (c *Compiler) typedExp(ef expFunc, in *string,
	vt ValueType) (BytecodeExp, error) {
	c.token = c.tokenizer(in)
	var be BytecodeExp
	bv, err := ef(&be, in)
	if err != nil {
		return nil, err
	}
	if !bv.IsNone() {
		switch vt {
		case VT_Float:
			bv.SetF(bv.ToF())
		case VT_Int:
			bv.SetI(bv.ToI())
		case VT_Bool:
			bv.SetB(bv.ToB())
		}
		be.appendValue(bv)
	}
	return be, nil
}

func (c *Compiler) argExpression(in *string, vt ValueType) (BytecodeExp, error) {
	be, err := c.typedExp(c.expBoolOr, in, vt)
	if err != nil {
		return nil, err
	}
	if len(c.token) > 0 {
		if c.token != "," {
			return nil, Error("Invalid data: " + c.token)
		}
		oldin := *in
		if c.tokenizer(in) == "" {
			c.token = ""
		} else {
			*in = oldin
		}
	}
	return be, nil
}

func (c *Compiler) fullExpression(in *string, vt ValueType) (BytecodeExp, error) {
	be, err := c.typedExp(c.expBoolOr, in, vt)
	if err != nil {
		return nil, err
	}
	if len(c.token) > 0 {
		return nil, Error("Invalid data: " + c.token)
	}
	return be, nil
}

func (c *Compiler) parseSection(
	sctrl func(name, data string) error) (IniSection, bool, error) {
	is := NewIniSection()
	_type, persistent, ignorehitpause := true, true, true
	for ; c.i < len(c.lines); c.i++ {
		line := strings.TrimSpace(strings.SplitN(c.lines[c.i], ";", 2)[0])
		if len(line) > 0 && line[0] == '[' {
			c.i--
			break
		}
		var name, data string
		if len(line) >= 3 && strings.ToLower(line[:3]) == "var" {
			name, data = "var", line
		} else if len(line) >= 3 && strings.ToLower(line[:3]) == "map" {
			name, data = "map", line
		} else if len(line) >= 4 && strings.ToLower(line[:4]) == "fvar" {
			name, data = "fvar", line
		} else if len(line) >= 6 && strings.ToLower(line[:6]) == "sysvar" {
			name, data = "sysvar", line
		} else if len(line) >= 7 && strings.ToLower(line[:7]) == "sysfvar" {
			name, data = "sysfvar", line
		} else {
			ia := strings.IndexAny(line, "= \t")
			if ia > 0 {
				name = strings.ToLower(line[:ia])
				ia = strings.Index(line, "=")
				if ia >= 0 {
					data = strings.TrimSpace(line[ia+1:])
				}
			}
		}
		if len(name) > 0 {
			_, ok := is[name]
			if ok && (len(name) < 7 || name[:7] != "trigger") {
				if sys.ignoreMostErrors {
					continue
				}
				return nil, false, Error(name + " is duplicated")
			}
			if sctrl != nil {
				switch name {
				case "type":
					if !_type {
						continue
					}
					_type = false
				case "persistent":
					if !persistent {
						continue
					}
					persistent = false
				case "ignorehitpause":
					if !ignorehitpause {
						continue
					}
					ignorehitpause = false
				default:
					if len(name) < 7 || name[:7] != "trigger" {
						is[name] = data
						continue
					}
				}
				if err := sctrl(name, data); err != nil {
					return nil, false, err
				}
			} else {
				is[name] = data
			}
		}
	}
	return is, !ignorehitpause, nil
}

func (c *Compiler) stateSec(is IniSection, f func() error) error {
	if err := f(); err != nil {
		return err
	}
	if !sys.ignoreMostErrors {
		var str string
		for k := range is {
			if len(str) > 0 {
				str += ", "
			}
			str += k
		}
		if len(str) > 0 {
			return Error("Invalid key name: " + str)
		}
	}
	return nil
}

func (c *Compiler) stateParam(is IniSection, name string, mandatory bool, f func(string) error) error {
	data, ok := is[name]
	if ok {
		if err := f(data); err != nil {
			return Error(data + "\n" + name + ": " + err.Error())
		}
		delete(is, name)
	} else if mandatory {
		return Error(name + " not specified")
	}
	return nil
}

// Returns FX prefix from a data string while removing prefix from the data
func (c *Compiler) getDataPrefix(data *string, ffxDefault bool) (prefix string) {
	if len(*data) > 1 {
		str := strings.ToLower(*data)

		// Find the longest matching valid prefix at the start of the string
		// The length check allows "FFF" to be used even though "F" is reserved
		longestMatch := ""
		// Check "F" and "S" reserved prefixes
		if strings.HasPrefix(str, "f") {
			longestMatch = "f"
		}
		if strings.HasPrefix(str, "s") {
			longestMatch = "s"
		}
		// Check common FX prefixes currently in use
		for p := range sys.ffx {
			if strings.HasPrefix(str, p) && len(p) > len(longestMatch) {
				longestMatch = p
			}
		}

		if longestMatch != "" {
			// Get the substring after the matched prefix
			rest := str[len(longestMatch):]

			// Split by any sequence of non-letter characters to isolate tokens
			re := regexp.MustCompile("[^a-z]+")
			tokens := re.Split(rest, -1)

			nextToken := ""
			if len(tokens) > 0 {
				nextToken = tokens[0]
			}

			// Remove prefix only if next token is empty or a known trigger
			if nextToken == "" || triggerMap[nextToken] != 0 {
				prefix = longestMatch
				*data = (*data)[len(longestMatch):]
			}
		}
	}

	if ffxDefault && prefix == "" {
		prefix = "f"
	}

	return
}

func (c *Compiler) exprs(data string, vt ValueType,
	numArg int) ([]BytecodeExp, error) {
	bes := []BytecodeExp{}
	for n := 1; n <= numArg; n++ {
		var be BytecodeExp
		var err error
		if n < numArg {
			be, err = c.argExpression(&data, vt)
		} else {
			be, err = c.fullExpression(&data, vt)
		}
		if err != nil {
			return nil, err
		}
		bes = append(bes, be)
		if c.token != "," {
			break
		}
	}
	return bes, nil
}

func (c *Compiler) scAdd(sc *StateControllerBase, id byte,
	data string, vt ValueType, numArg int, topbe ...BytecodeExp) error {
	bes, err := c.exprs(data, vt, numArg)
	if err != nil {
		return err
	}
	sc.add(id, append(topbe, bes...))
	return nil
}

// ParamValue adds the parameter immediately, unlike StateParam which only reads it
func (c *Compiler) paramValue(is IniSection, sc *StateControllerBase,
	paramname string, id byte, vt ValueType, numArg int, mandatory bool) error {
	found := false
	if err := c.stateParam(is, paramname, false, func(data string) error {
		found = true
		return c.scAdd(sc, id, data, vt, numArg)
	}); err != nil {
		return err
	}
	if mandatory && !found {
		return Error(paramname + " not specified")
	}
	return nil
}

func (c *Compiler) paramAnimtype(is IniSection, sc *StateControllerBase, paramName string, id byte) error {
	return c.stateParam(is, paramName, false, func(data string) error {
		if len(data) == 0 {
			return Error(paramName + " not specified")
		}
		var ra Reaction
		dataLower := strings.ToLower(data)
		//if sys.cgi[c.playerNo].ikemenver[0] == 0 && sys.cgi[c.playerNo].ikemenver[1] == 0 {
		if !c.zssMode {
			// CNS: first letter is enough
			switch dataLower[0] {
			case 'l':
				ra = RA_Light
			case 'm':
				ra = RA_Medium
			case 'h':
				ra = RA_Hard
			case 'b':
				ra = RA_Back
			case 'u':
				ra = RA_Up
			case 'd':
				ra = RA_Diagup
			default:
				return Error("Invalid " + paramName + ": " + data)
			}
		} else {
			// ZSS: require full word
			switch dataLower {
			case "light":
				ra = RA_Light
			case "medium":
				ra = RA_Medium
			case "hard":
				ra = RA_Hard
			case "heavy":
				ra = RA_Hard
			case "back":
				ra = RA_Back
			case "up":
				ra = RA_Up
			case "diagup":
				ra = RA_Diagup
			default:
				return Error("Invalid " + paramName + ": " + data)
			}
		}
		sc.add(id, sc.iToExp(int32(ra)))
		return nil
	})
}

func (c *Compiler) paramHittype(is IniSection, sc *StateControllerBase, paramName string, id byte) error {
	return c.stateParam(is, paramName, false, func(data string) error {
		if len(data) == 0 {
			return Error(paramName + " not specified")
		}
		var ht HitType
		dataLower := strings.ToLower(data)
		//if sys.cgi[c.playerNo].ikemenver[0] == 0 && sys.cgi[c.playerNo].ikemenver[1] == 0 {
		if !c.zssMode {
			// CNS: first letter is enough
			switch dataLower[0] {
			case 'h':
				ht = HT_High
			case 'l':
				ht = HT_Low
			case 't':
				ht = HT_Trip
			case 'n':
				ht = HT_None
			default:
				return Error("Invalid " + paramName + ": " + data)
			}
		} else {
			// ZSS: require full word
			switch dataLower {
			case "high":
				ht = HT_High
			case "low":
				ht = HT_Low
			case "trip":
				ht = HT_Trip
			case "none":
				ht = HT_None
			default:
				return Error("Invalid " + paramName + ": " + data)
			}
		}
		sc.add(id, sc.iToExp(int32(ht)))
		return nil
	})
}

func (c *Compiler) paramPostype(is IniSection, sc *StateControllerBase, id byte) error {
	return c.stateParam(is, "postype", false, func(data string) error {
		if len(data) == 0 {
			return Error("postype not specified")
		}
		var pt PosType
		dataLower := strings.ToLower(data)
		//if sys.cgi[c.playerNo].ikemenver[0] == 0 && sys.cgi[c.playerNo].ikemenver[1] == 0 {
		if !c.zssMode {
			// CNS: first letter is enough
			if len(dataLower) >= 2 && dataLower[:2] == "p2" {
				pt = PT_P2
			} else {
				switch dataLower[0] {
				case 'p':
					pt = PT_P1
				case 'f':
					pt = PT_Front
				case 'b':
					pt = PT_Back
				case 'l':
					pt = PT_Left
				case 'r':
					pt = PT_Right
				case 'n':
					pt = PT_None
				default:
					return Error("Invalid postype: " + data)
				}
			}
		} else {
			// ZSS: require full word
			switch dataLower {
			case "p1":
				pt = PT_P1
			case "p2":
				pt = PT_P2
			case "front":
				pt = PT_Front
			case "back":
				pt = PT_Back
			case "left":
				pt = PT_Left
			case "right":
				pt = PT_Right
			case "none":
				pt = PT_None
			default:
				return Error("Invalid postype: " + data)
			}
		}
		sc.add(id, sc.iToExp(int32(pt)))
		return nil
	})
}

func (c *Compiler) paramSpace(is IniSection, sc *StateControllerBase, id byte) error {
	return c.stateParam(is, "space", false, func(data string) error {
		if len(data) == 0 {
			return Error("space not specified")
		}
		var spc Space
		switch strings.ToLower(data) {
		case "stage":
			spc = Space_stage
		case "screen":
			spc = Space_screen
		default:
			//if sys.cgi[c.playerNo].ikemenverF > 0 && !sys.ignoreMostErrors {
			if c.zssMode && !sys.ignoreMostErrors {
				return Error("Invalid space type: " + data)
			} else {
				sys.appendToConsole("WARNING: " + sys.cgi[c.playerNo].nameLow + fmt.Sprintf(": Invalid space type: "+data+" in state %v ", c.stateNo))
			}
		}
		sc.add(id, sc.iToExp(int32(spc)))
		return nil
	})
}

func (c *Compiler) paramProjection(is IniSection, sc *StateControllerBase, key string, id byte) error {
	return c.stateParam(is, key, false, func(data string) error {
		var proj Projection

		switch strings.ToLower(strings.TrimSpace(data)) {
		case "orthographic":
			proj = Projection_Orthographic
		case "perspective":
			proj = Projection_Perspective
		case "perspective2":
			proj = Projection_Perspective2
		default:
			return Error("invalid projection type: " + data)
		}

		sc.add(id, sc.iToExp(int32(proj)))
		return nil
	})
}

func (c *Compiler) paramSaveData(is IniSection, sc *StateControllerBase, id byte) error {
	return c.stateParam(is, "savedata", false, func(data string) error {
		if len(data) <= 1 {
			return Error("savedata not specified")
		}
		var sv SaveData
		switch strings.ToLower(data) {
		case "map":
			sv = SaveData_map
		case "var":
			sv = SaveData_var
		case "fvar":
			sv = SaveData_fvar
		default:
			return Error("Invalid savedata type: " + data)
		}
		sc.add(id, sc.iToExp(int32(sv)))
		return nil
	})
}

// Parse trans and alpha together
func (c *Compiler) paramTrans(is IniSection, sc *StateControllerBase,
	prefix string, id byte, afterImage bool) error {

	return c.stateParam(is, prefix+"trans", false, func(data string) error {
		if len(data) == 0 {
			return Error("trans type not specified")
		}

		// Defaults
		tt := TT_default
		defsrc, defdst := int32(255), int32(0)

		// Parse the trans type and set default alpha
		data = strings.ToLower(data)
		switch data {
		case "default":
			tt = TT_default
			defsrc, defdst = 255, 0
		case "none":
			tt = TT_none
			defsrc, defdst = 255, 0
		case "add":
			tt = TT_add
			defsrc, defdst = 255, 255
		case "add1":
			tt = TT_add
			defsrc, defdst = 255, 128
		case "addalpha":
			tt = TT_add
			defsrc, defdst = 255, 0 // In Mugen it defaults to this before reading the alpha
		case "sub":
			tt = TT_sub
			defsrc, defdst = 255, 255
		default:
			// In Mugen, afterimages ignore invalid parameter names
			if !afterImage || c.zssMode || !sys.ignoreMostErrors {
				return Error("Invalid trans type: " + data)
			} else {
				sys.appendToConsole("WARNING: " + sys.cgi[c.playerNo].nameLow + fmt.Sprintf(": Invalid trans type: "+data+" in state %v ", c.stateNo))
				return nil
			}
		}

		exp := make([]BytecodeExp, 3)

		// Parse custom alpha
		_ = c.stateParam(is, prefix+"alpha", false, func(data string) error {
			vals, err := c.exprs(data, VT_Int, 2)
			if err != nil {
				return err
			}
			if len(vals) > 0 {
				exp[0] = vals[0]
			}
			if len(vals) > 1 {
				exp[1] = vals[1]
			}
			return nil
		})

		// Use custom or default alpha
		if exp[0] == nil {
			exp[0] = sc.iToExp(defsrc)[0]
		}
		if exp[1] == nil {
			exp[1] = sc.iToExp(defdst)[0]
		}

		// Always use trans type
		exp[2] = sc.iToExp(int32(tt))[0]

		sc.add(id, exp)
		return nil
	})
}

// Interprets an IniSection of statedef properties and sets them to a StateBytecode
func (c *Compiler) stateDef(is IniSection, sbc *StateBytecode) error {
	return c.stateSec(is, func() error {
		sc := newStateControllerBase()
		if err := c.stateParam(is, "type", false, func(data string) error {
			if len(data) == 0 {
				return Error("statetype not specified")
			}
			switch strings.ToLower(data)[0] {
			case 's':
				sbc.stateType = ST_S
			case 'c':
				sbc.stateType = ST_C
			case 'a':
				sbc.stateType = ST_A
			case 'l':
				sbc.stateType = ST_L
			case 'u':
				sbc.stateType = ST_U
			default:
				return Error("Invalid statetype: " + data)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "movetype", false, func(data string) error {
			if len(data) == 0 {
				return Error("movetype not specified")
			}
			switch strings.ToLower(data)[0] {
			case 'i':
				sbc.moveType = MT_I
			case 'a':
				sbc.moveType = MT_A
			case 'h':
				sbc.moveType = MT_H
			case 'u':
				sbc.moveType = MT_U
			default:
				return Error("Invalid movetype: " + data)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "physics", false, func(data string) error {
			if len(data) == 0 {
				return Error("physics not specified")
			}
			switch strings.ToLower(data)[0] {
			case 's':
				sbc.physics = ST_S
			case 'c':
				sbc.physics = ST_C
			case 'a':
				sbc.physics = ST_A
			case 'n':
				sbc.physics = ST_N
			case 'u':
				sbc.physics = ST_U
			default:
				return Error("Invalid physics type: " + data)
			}
			return nil
		}); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "hitcountpersist", false, func(data string) error {
			b = true
			return c.scAdd(sc, stateDef_hitcountpersist, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_hitcountpersist, sc.iToExp(0))
		}
		b = false
		if err := c.stateParam(is, "movehitpersist", false, func(data string) error {
			b = true
			return c.scAdd(sc, stateDef_movehitpersist, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_movehitpersist, sc.iToExp(0))
		}
		b = false
		if err := c.stateParam(is, "hitdefpersist", false, func(data string) error {
			b = true
			return c.scAdd(sc, stateDef_hitdefpersist, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_hitdefpersist, sc.iToExp(0))
		}
		if err := c.paramValue(is, sc, "sprpriority",
			stateDef_sprpriority, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "facep2",
			stateDef_facep2, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "juggle", false, func(data string) error {
			return c.scAdd(sc, stateDef_juggle, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "velset",
			stateDef_velset, VT_Float, 3, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", false, func(data string) error {
			prefix := c.getDataPrefix(&data, false)
			return c.scAdd(sc, stateDef_anim, data, VT_Int, 1,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ctrl",
			stateDef_ctrl, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "poweradd",
			stateDef_poweradd, VT_Int, 1, false); err != nil {
			return err
		}
		sbc.stateDef = stateDef(*sc)
		return nil
	})
}

// Parses multiple strings separated by ','
func cnsStringArray(arg string) ([]string, error) {
	// Split the plain text string array into substrings,
	var strArray = strings.Split(arg, ",")
	// If "1" it means we are inside a string,
	var inString = 0
	// The array that we return with parsed strings.
	var fullStrArray []string = make([]string, len(strArray))
	// When comes the inevitable moment a when user makes a typo.
	var formatError = false

	// Iterate the string array.
	for i, values := range strArray {
		for _, char := range values {
			if char == '"' { // Open/close string.
				inString++
			} else if inString == 1 { // Add any char to the array if we are inside a string.
				fullStrArray[i] += string(char)
			} else if char != ' ' { // If anything that is not whitespace is outside the declaration is bad syntax.
				formatError = true
			}
		}

		// Was the string closed?
		if inString != 2 {
			if inString%2 != 0 {
				return nil, Error("String not closed")
			} else if inString > 2 { // Do we have more than 1 string without using ','?
				return nil, Error("Lack of ',' separator")
			} else {
				return nil, Error("Unknown string array error")
			}
		} else if formatError {
			return nil, Error("Wrong format on string array")
		} else { // All's good.
			inString = 0
		}
	} // Return the parsed string array,
	return fullStrArray, nil
}

// Compile a state file
func (c *Compiler) stateCompile(states map[int32]StateBytecode,
	filename string, dirs []string, negoverride bool, constants map[string]float32) error {
	var str string
	c.zssMode = HasExtension(filename, ".zss")
	fnz := filename

	// Load state file
	if err := LoadFile(&filename, dirs, func(filename string) error {
		var err error
		// If this is a zss file
		if c.zssMode {
			b, err := LoadText(filename)
			if err != nil {
				return err
			}
			str = string(b)
			return c.stateCompileZ(states, fnz, str, constants)
		}

		// Try reading as an st file
		str, err = LoadText(filename)
		return err
	}); err != nil {
		// If filename doesn't exist, see if a zss file exists
		fnz += ".zss"
		if err := LoadFile(&fnz, dirs, func(filename string) error {
			b, err := LoadText(filename)
			if err != nil {
				return err
			}
			str = string(b)
			return nil
		}); err == nil {
			return c.stateCompileZ(states, fnz, str, constants)
		}
		return err
	}

	c.lines, c.i = SplitAndTrim(str, "\n"), 0
	errmes := func(err error) error {
		return Error(fmt.Sprintf("%v:%v:\n%v", filename, c.i+1, err.Error()))
	}
	// Keep a map of states that have already been found in this file
	existInThisFile := make(map[int32]bool)
	c.vars = make(map[string]uint8)
	// Loop through state file lines
	for ; c.i < len(c.lines); c.i++ {
		// Find a statedef, skipping over other lines until finding one
		// Get the current line, without comments
		line := strings.ToLower(strings.TrimSpace(
			strings.SplitN(c.lines[c.i], ";", 2)[0]))
		// If this is not a line starting a statedef, continue to the next line
		if len(line) < 11 || line[0] != '[' || line[len(line)-1] != ']' ||
			line[1:10] != "statedef " {
			continue
		}

		// Parse state number
		line = line[10:]
		var err error
		if c.stateNo, err = c.scanStateDef(&line, constants); err != nil {
			return errmes(err)
		}

		// Skip if this state has already been added
		if existInThisFile[c.stateNo] {
			continue
		}
		existInThisFile[c.stateNo] = true

		c.i++
		// Parse the statedef properties
		is, _, err := c.parseSection(nil)
		if err != nil {
			return errmes(err)
		}
		sbc := newStateBytecode(c.playerNo)
		if _, ok := states[c.stateNo]; ok && c.stateNo < 0 {
			*sbc = states[c.stateNo]
		}
		// Interpret the statedef properties
		if err := c.stateDef(is, sbc); err != nil {
			return errmes(err)
		}

		// Continue looping through state file lines to define the current state
		for c.i++; c.i < len(c.lines); c.i++ {
			// Get the current line, without comments
			line := strings.ToLower(strings.TrimSpace(
				strings.SplitN(c.lines[c.i], ";", 2)[0]))
			// If this is not a line starting an sctrl, continue to the next line
			if line == "" || line[0] != '[' || line[len(line)-1] != ']' {
				continue
			}
			if len(line) < 7 || line[1:7] != "state " {
				c.i--
				break
			}
			c.i++

			// Create this sctrl and get its properties
			c.block = newStateBlock()
			sc := newStateControllerBase()
			var scf scFunc
			var triggerall []BytecodeExp
			// Flag if following triggers can never be true because of triggerall = 0
			allTerminated := false
			var trigger [][]BytecodeExp
			var trexist []int8
			// Parse each line of the sctrl to get triggers and settings
			is, ihp, err := c.parseSection(func(name, data string) error {
				switch name {
				case "type":
					var ok bool
					scf, ok = c.scmap[strings.ToLower(data)]
					if !ok {
						return Error("Invalid state controller: " + data)
					}
				case "persistent":
					if c.stateNo >= 0 {
						c.block.persistent = Atoi(data)
						if c.block.persistent > 128 {
							c.block.persistent = 1
						} else if c.block.persistent != 1 {
							if c.block.persistent <= 0 {
								c.block.persistent = math.MaxInt32
							}
							c.block.persistentIndex = int32(len(sbc.ctrlsps))
							sbc.ctrlsps = append(sbc.ctrlsps, 0)
						}
					}
				case "ignorehitpause":
					ih := Atoi(data) != 0
					c.block.ignorehitpause = Btoi(ih) - 2
					c.block.ctrlsIgnorehitpause = ih
				case "triggerall":
					be, err := c.fullExpression(&data, VT_Bool)
					if err != nil {
						return err
					}
					// If triggerall = 0 is encountered, flag it
					if len(be) == 2 && be[0] == OC_int8 {
						if be[1] == 0 {
							allTerminated = true
						}
					} else if !allTerminated {
						triggerall = append(triggerall, be)
					}
				default:
					// Get the trigger number
					tn, ok := readDigit(name[7:])
					if !ok || tn < 1 || tn > 65536 {
						if sys.ignoreMostErrors {
							break
						}
						return Error("Invalid trigger name: " + name)
					}
					// Add more entries to the trigger collection if needed
					if len(trigger) < int(tn) {
						trigger = append(trigger, make([][]BytecodeExp,
							int(tn)-len(trigger))...)
					}
					if len(trexist) < int(tn) {
						trexist = append(trexist, make([]int8, int(tn)-len(trexist))...)
					}
					tn--
					// Parse trigger condition into a bytecode expression
					be, err := c.fullExpression(&data, VT_Bool)
					if err != nil {
						if sys.ignoreMostErrors {
							_break := false
							for i := 0; i < int(tn); i++ {
								if trexist[i] == 0 {
									_break = true
									break
								}
							}
							if _break {
								break
							}
						}
						return err
					}
					// If trigger is a constant int value
					if len(be) == 2 && be[0] == OC_int8 {
						// If trigger is always false (0)
						if be[1] == 0 {
							// trexist == -1 means this specific trigger set can never be true
							trexist[tn] = -1
						} else if trexist[tn] == 0 {
							trexist[tn] = 1
						}
					} else if !allTerminated && trexist[tn] >= 0 {
						trigger[tn] = append(trigger[tn], be)
						trexist[tn] = 1
					}
				}
				return nil
			})
			if err != nil {
				return errmes(err)
			}

			// Check that the sctrl has a valid type parameter
			if scf == nil {
				return errmes(Error("State controller type not specified"))
			}
			if len(trexist) == 0 || (!allTerminated && trexist[0] == 0) {
				return errmes(Error("Missing trigger1"))
			}

			// Create trigger bytecode
			var texp BytecodeExp
			for _, e := range triggerall {
				texp.append(e...)
				texp.append(OC_jz8, 0)
				texp.append(OC_pop)
			}
			if allTerminated {
				if len(texp) > 0 {
					texp.appendValue(BytecodeBool(false))
				}
			} else {
				for i, tr := range trigger {
					if trexist[i] == 0 {
						break
					}
					var te BytecodeExp
					if trexist[i] < 0 {
						te.append(OC_pop)
						te.appendValue(BytecodeBool(false))
					}
					oldlen := len(te)
					for j := len(tr) - 1; j >= 0; j-- {
						tmp := tr[j]
						if j < len(tr)-1 {
							if len(te) > int(math.MaxUint8-1) {
								tmp.appendI32Op(OC_jz, int32(len(te)+1))
							} else {
								tmp.append(OC_jz8, OpCode(len(te)+1))
							}
							tmp.append(OC_pop)
						}
						te = append(tmp, te...)
					}
					if len(te) == oldlen {
						te = nil
					}
					if len(te) == 0 {
						if trexist[i] > 0 {
							if len(texp) > 0 {
								texp.appendValue(BytecodeBool(true))
								texp.append(OC_jmp8, 0)
							}
							break
						}
						if len(texp) > 0 && (i == len(trigger)-1 || trexist[i+1] == 0) {
							texp.appendValue(BytecodeBool(false))
						}
					} else {
						texp.append(te...)
						if i < len(trigger)-1 && trexist[i+1] != 0 {
							texp.append(OC_jnz8, 0)
							texp.append(OC_pop)
						}
					}
				}
			}
			c.block.trigger = texp

			// Ignorehitpause
			_ihp := int8(-1)
			if ihp {
				_ihp = int8(Btoi(c.block.ignorehitpause >= -1))
			}

			// For this sctrl type, call the function to construct the sctrl
			sctrl, err := scf(is, sc, _ihp)
			if err != nil {
				return errmes(err)
			}

			// Check if the triggers can ever be true before appending the new sctrl
			appending := true
			if len(c.block.trigger) == 0 {
				appending = false
				if !allTerminated {
					for _, te := range trexist {
						if te >= 0 {
							if te > 0 {
								appending = true
							}
							break
						}
					}
				}
			}
			if appending {
				// If the trigger is always true
				if len(c.block.trigger) == 0 && c.block.persistentIndex < 0 &&
					c.block.ignorehitpause < -1 {
					if _, ok := sctrl.(NullStateController); !ok {
						sbc.block.ctrls = append(sbc.block.ctrls, sctrl)
					}
				} else {
					if _, ok := sctrl.(NullStateController); !ok {
						c.block.ctrls = append(c.block.ctrls, sctrl)
					}
					sbc.block.ctrls = append(sbc.block.ctrls, *c.block)
					if c.block.ignorehitpause >= -1 {
						sbc.block.ignorehitpause = -1
					}
				}
			}
		}

		// Skip appending if already declared. Exception for negative states present in CommonStates and files belonging to char flagged with ikemenversion
		if _, ok := states[c.stateNo]; !ok || (!negoverride && c.stateNo < 0) {
			states[c.stateNo] = *sbc
		}
	}
	return nil
}

func (c *Compiler) wrongClosureToken() error {
	if c.token == "" {
		return Error("Missing closure token")
	}
	return Error("Unexpected closure token: " + c.token)
}

func (c *Compiler) nextLine() (string, bool) {
	s := <-c.linechan
	if s == nil {
		return "", false
	}
	return *s, true
}

func (c *Compiler) scan(line *string) string {
	for {
		c.token = c.tokenizer(line)
		if len(c.token) > 0 {
			if c.token[0] != '#' {
				break
			}
		}
		var ok bool
		*line, ok = c.nextLine()
		if !ok {
			break
		}
	}
	return c.token
}

func (c *Compiler) needToken(t string) error {
	if c.token != t {
		if c.token == "" {
			return Error("Missing token: " + t)
		}
		return Error(fmt.Sprintf("Wrong token: expected %v, got %v", t, c.token))
	}
	return nil
}

func (c *Compiler) readString(line *string) (string, error) {
	i := strings.Index(*line, "\"")
	if i < 0 {
		return "", Error("String not enclosed in \"")
	}
	s := (*line)[:i]
	*line = (*line)[i+1:]
	return s, nil
}

func (c *Compiler) readSentenceLine(line *string) (s string, assign bool,
	err error) {
	c.token = ""
	offset := 0
	for {
		i := strings.IndexAny((*line)[offset:], ":;#\"{}")
		if i < 0 {
			s, *line = *line, ""
			return
		}
		i += offset
		switch (*line)[i] {
		case ':', ';', '{', '}':
			if (*line)[i] == ':' && len(*line) > i+1 && (*line)[i+1] == '=' {
				assign = true
				offset = i + 1
				continue
			}
			c.token = (*line)[i : i+1]
			s, *line = (*line)[:i], (*line)[i+1:]
		case '#':
			s, *line = (*line)[:i], "" // Ignore the rest as a comment
		case '"':
			tmp := (*line)[i+1:]
			if _, err := c.readString(&tmp); err != nil {
				return "", false, err
			}
			offset = len(*line) - len(tmp)
			continue
		}
		break
	}
	return
}

func (c *Compiler) readSentence(line *string) (s string, a bool, err error) {
	if s, a, err = c.readSentenceLine(line); err != nil {
		return
	}
	for c.token == "" {
		var ok bool
		*line, ok = c.nextLine()
		if !ok {
			break
		}
		if sen, ass, err := c.readSentenceLine(line); err != nil {
			return "", false, err
		} else {
			s += "\n" + sen
			a = a || ass
		}
	}
	return strings.TrimSpace(s), a, nil
}

func (c *Compiler) statementEnd(line *string) error {
	c.token = c.tokenizer(line)
	if len(c.token) > 0 && c.token[0] != '#' {
		return c.wrongClosureToken()
	}
	c.token, *line = "", ""
	return nil
}

func (c *Compiler) readKeyValue(is IniSection, end string,
	line *string) error {
	name := c.scan(line)
	if name == "" || name == ":" {
		return c.wrongClosureToken()
	}
	if name == end {
		return nil
	}
	c.scan(line)
	if err := c.needToken(":"); err != nil {
		return err
	}
	data, _, err := c.readSentence(line)
	if err != nil {
		return err
	}
	is[name] = data
	return nil
}

func (c *Compiler) varNameCheck(nm string) (err error) {
	if (nm[0] < 'a' || nm[0] > 'z') && nm[0] != '_' {
		return Error("Invalid name: " + nm)
	}
	for _, c := range nm[1:] {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_' {
			return Error("Invalid name: " + nm)
		}
	}
	return nil
}

func (c *Compiler) varNames(end string, line *string) ([]string, error) {
	names, name := []string{}, c.scan(line)
	if name != end {
		for {
			if name == "" || name == "," || name == end {
				return nil, c.wrongClosureToken()
			}
			if err := c.varNameCheck(name); err != nil {
				return nil, err
			}
			if name != "_" {
				for _, nm := range names {
					if nm == name {
						return nil, Error("Duplicate name: " + name)
					}
				}
			}
			names = append(names, name)
			c.scan(line)
			if c.token == "," {
				name = c.scan(line)
			} else {
				if err := c.needToken(end); err != nil {
					return nil, err
				}
				break
			}
		}
	}
	return names, nil
}

func (c *Compiler) inclNumVars(numVars *int32) error {
	*numVars++
	if *numVars > 256 {
		return Error("Exceeded 256 local variable limit")
	}
	return nil
}

func (c *Compiler) scanI32(line *string) (int32, error) {
	t := c.scan(line)
	if t == "" {
		return 0, c.wrongClosureToken()
	}
	if t == "-" && len(*line) > 0 && (*line)[0] >= '0' && (*line)[0] <= '9' {
		t += c.scan(line)
	}
	v, err := strconv.ParseInt(t, 10, 32)
	return int32(v), err
}

func (c *Compiler) scanStateDef(line *string, constants map[string]float32) (int32, error) {
	t := c.scan(line)
	if t == "" {
		return 0, c.wrongClosureToken()
	}
	var err error
	// StateDef using constants
	if t == "const" {
		c.scan(line)
		k := c.scan(line)
		c.scan(line)
		v, ok := constants[k]
		if !ok {
			err = Error(fmt.Sprintf("StateDef constant not found: %v", k))
		}
		return int32(v), err
	}
	// Special +1 case
	if t == "+" {
		nextToken := c.scan(line)
		if nextToken == "1" {
			return int32(-10), nil
		}
		t += nextToken
	}
	// Negative states
	if t == "-" && len(*line) > 0 && (*line)[0] >= '0' && (*line)[0] <= '9' {
		t += c.scan(line)
	}
	v := Atoi(t)
	return v, err
}

// Sets attributes to a StateBlock, like IgnoreHitPause, Persistent
func (c *Compiler) blockAttribSet(line *string, bl *StateBlock, sbc *StateBytecode,
	inheritIhp, nestedInLoop bool) error {
	// Inherit ignorehitpause/loop attr from parent block
	if inheritIhp {
		bl.ignorehitpause, bl.ctrlsIgnorehitpause = -1, true
		// Avoid re-reading ignorehitpause
		if c.token == "ignorehitpause" {
			c.scan(line)
		}
	}
	bl.nestedInLoop = nestedInLoop
	for {
		switch c.token {
		case "ignorehitpause":
			if bl.ignorehitpause >= -1 {
				return c.wrongClosureToken()
			}
			bl.ignorehitpause, bl.ctrlsIgnorehitpause = -1, true
			c.scan(line)
			continue
		case "persistent":
			if sbc == nil {
				return Error("Persistent cannot be used in a function")
			}
			if c.stateNo < 0 {
				return Error("Persistent cannot be used in a negative state")
			}
			if bl.persistentIndex >= 0 {
				return c.wrongClosureToken()
			}
			c.scan(line)
			if err := c.needToken("("); err != nil {
				return err
			}
			var err error
			if bl.persistent, err = c.scanI32(line); err != nil {
				return err
			}
			c.scan(line)
			if err := c.needToken(")"); err != nil {
				return err
			}
			if bl.persistent == 1 {
				return Error("Persistent(1) is meaningless") // TODO: Do we really need to crash here?
			}
			if bl.persistent <= 0 {
				bl.persistent = math.MaxInt32
			}
			bl.persistentIndex = int32(len(sbc.ctrlsps))
			sbc.ctrlsps = append(sbc.ctrlsps, 0)
			c.scan(line)
			continue
		}
		break
	}
	return nil
}

func (c *Compiler) subBlock(line *string, root bool,
	sbc *StateBytecode, numVars *int32, inheritIhp, nestedInLoop bool) (*StateBlock, error) {
	bl := newStateBlock()
	if err := c.blockAttribSet(line, bl, sbc, inheritIhp, nestedInLoop); err != nil {
		return nil, err
	}
	compileMain, compileElse := true, false
	switch c.token {
	case "{":
	case "if":
		compileElse = true
		expr, _, err := c.readSentence(line)
		if err != nil {
			return nil, err
		}
		otk := c.token
		if bl.trigger, err = c.fullExpression(&expr, VT_Bool); err != nil {
			return nil, err
		}
		c.token = otk
		if err := c.needToken("{"); err != nil {
			return nil, err
		}
	case "switch":
		compileMain = false
		if err := c.switchBlock(line, bl, sbc, numVars); err != nil {
			return nil, err
		}
	case "for", "while":
		if err := c.loopBlock(line, root, bl, sbc, numVars); err != nil {
			return nil, err
		}
	default:
		return nil, c.wrongClosureToken()
	}
	if compileMain {
		if err := c.stateBlock(line, bl, false,
			sbc, &bl.ctrls, numVars); err != nil {
			return nil, err
		}
	}
	if root {
		if len(bl.trigger) > 0 {
			if c.token = c.tokenizer(line); c.token != "else" {
				if len(c.token) == 0 || c.token[0] == '#' {
					c.token, *line = "", ""
				} else {
					return nil, c.wrongClosureToken()
				}
				c.scan(line)
			}
		} else {
			if err := c.statementEnd(line); err != nil {
				return nil, err
			}
			c.scan(line)
		}
	} else {
		c.scan(line)
	}
	if compileElse && len(bl.trigger) > 0 && c.token == "else" {
		c.scan(line)
		var err error
		if bl.elseBlock, err = c.subBlock(line, root,
			sbc, numVars, inheritIhp || bl.ctrlsIgnorehitpause, nestedInLoop); err != nil {
			return nil, err
		}
		if bl.elseBlock.ignorehitpause >= -1 {
			bl.ignorehitpause = -1
		}
	}
	return bl, nil
}

func (c *Compiler) switchBlock(line *string, bl *StateBlock,
	sbc *StateBytecode, numVars *int32) error {
	// In this implementation of switch, we convert the statement to an if-elseif-else chain of blocks
	header, _, err := c.readSentence(line)
	if err != nil {
		return err
	}
	if err := c.needToken("{"); err != nil {
		return err
	}
	c.scan(line)
	compileCaseBlock := func(sbl *StateBlock, expr *string) error {
		if err := c.blockAttribSet(line, sbl, sbc,
			bl != nil && bl.ctrlsIgnorehitpause, bl != nil && bl.nestedInLoop); err != nil {
			return err
		}
		otk := c.token
		if sbl.trigger, err = c.fullExpression(expr, VT_Bool); err != nil {
			return err
		}
		c.token = otk
		// Compile the inner block for this case
		if err := c.stateBlock(line, sbl, false,
			sbc, &sbl.ctrls, numVars); err != nil {
			return err
		}
		return nil
	}
	// Start examining the cases
	var readNextCase func(*StateBlock) (*StateBlock, error)
	readNextCase = func(def *StateBlock) (*StateBlock, error) {
		expr := ""
		switch c.token {
		case "case":
		case "default":
			if def != nil {
				return nil, Error("Default already defined")
			}
			c.scan(line)
			expr = "1"
			def = newStateBlock()
			if err := compileCaseBlock(def, &expr); err != nil {
				return nil, err
			}
			// See if default is the last case defined in this switch statement,
			// return default block if that's the case
			if c.token == "}" {
				return def, nil
			}
		default:
			return nil, Error("Expected case or default")
		}
		// We loop through all possible expressions in this case, separated by ;
		// Creating an equality/or expression string in the process
		for {
			caseValue, _, err := c.readSentence(line)
			if err != nil {
				return nil, err
			}
			// We create an equality expression that looks like this: header = caseValue
			// and we append it to the case block expression. Colon at the end is also removed
			expr += header + " = " + caseValue
			if c.token == ";" {
				// We'll have another expression to test for this case, so we append an OR operator
				expr += " || "
				continue
			}
			// We finished reading the case, check for colon existence
			if err := c.needToken(":"); err != nil {
				return nil, err
			}
			break
		}
		// Create a new state block for this case
		sbl := newStateBlock()
		if err := compileCaseBlock(sbl, &expr); err != nil {
			return nil, err
		}
		// Switch has finished
		if c.token == "}" {
			// Assign default block as the latest else in the chain
			if def != nil {
				sbl.elseBlock = def
			}
			// If not, we have another case to check
		} else if sbl.elseBlock, err = readNextCase(def); err != nil {
			return nil, err
		}
		return sbl, nil
	}
	if sbl, err := readNextCase(nil); err != nil {
		return err
	} else {
		if bl != nil && sbl.ignorehitpause >= -1 {
			bl.ignorehitpause = -1
		}
		bl.ctrls = append(bl.ctrls, *sbl)
	}
	return nil
}

func (c *Compiler) loopBlock(line *string, root bool, bl *StateBlock,
	sbc *StateBytecode, numVars *int32) error {
	bl.loopBlock = true
	bl.nestedInLoop = true
	switch c.token {
	case "for":
		bl.forLoop = true
		i := 0
		tmp := *line
		nm := c.scan(&tmp)
		if (nm[0] >= 'a' && nm[0] <= 'z') || nm[0] == '_' {
			// Local variable assignation from for header
			names, err := c.varNames("=", line)
			if err != nil {
				return err
			}
			if len(names) > 0 {
				var tmp []StateController
				if err := c.letAssign(line, root, &tmp, numVars, names, false); err != nil {
					return err
				}
				bl.forCtrlVar = tmp[0].(varAssign)
				bl.forAssign = true
				i = 1
			}
		}
		// Compile header expressions
		for ; i < 3; i++ {
			if c.token == "{" {
				if i < 2 {
					return Error("For loop needs more than one expression")
				} else {
					// For only has begin/end expressions, so we stop compiling the header
					break
				}
			}
			if c.token == ";" && i < 1 {
				return Error("Misplaced ';' in for loop")
			}
			expr, _, err := c.readSentence(line)
			if err != nil {
				return err
			}
			otk := c.token
			if bl.forExpression[i], err = c.fullExpression(&expr, VT_Int); err != nil {
				return err
			}
			c.token = otk
		}
		if err := c.needToken("{"); err != nil {
			return err
		}
		// Default increment value: i++
		if bl.forExpression[2] == nil {
			var be BytecodeExp
			be.appendValue(BytecodeInt(1))
			bl.forExpression[2] = be
		}
	case "while":
		expr, _, err := c.readSentence(line)
		if err != nil {
			return err
		}
		otk := c.token
		if bl.trigger, err = c.fullExpression(&expr, VT_Bool); err != nil {
			return err
		}
		c.token = otk
		if err := c.needToken("{"); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) callFunc(line *string, root bool,
	ctrls *[]StateController, ret []uint8) error {
	var cf callFunction
	var ok bool
	cf.bytecodeFunction, ok = c.funcs[c.scan(line)]
	cf.ret = ret
	if !ok {
		if c.token == "" || c.token == "(" {
			return c.wrongClosureToken()
		}
		return Error("Undefined function: " + c.token)
	}
	c.funcUsed[c.token] = true
	if len(ret) > 0 && len(ret) != int(cf.numRets) {
		return Error(fmt.Sprintf("Mismatch in number of assignments and return values: %v = %v",
			len(ret), cf.numRets))
	}
	c.scan(line)
	if err := c.needToken("("); err != nil {
		return err
	}
	expr, _, err := c.readSentence(line)
	if err != nil {
		return err
	}
	otk := c.token
	if cf.numArgs == 0 {
		c.token = c.tokenizer(&expr)
		if c.token == "" {
			c.token = otk
		}
		if err := c.needToken(")"); err != nil {
			return err
		}
	} else {
		for i := 0; i < int(cf.numArgs); i++ {
			var be BytecodeExp
			if i < int(cf.numArgs)-1 {
				if be, err = c.argExpression(&expr, VT_SFalse); err != nil {
					return err
				}
				if c.token == "" {
					c.token = otk
				}
				if err := c.needToken(","); err != nil {
					return err
				}
			} else {
				if be, err = c.typedExp(c.expBoolOr, &expr, VT_SFalse); err != nil {
					return err
				}
				if c.token == "" {
					c.token = otk
				}
				if err := c.needToken(")"); err != nil {
					return err
				}
			}
			cf.arg.append(be...)
		}
	}
	if c.token = c.tokenizer(&expr); c.token != "" {
		return c.wrongClosureToken()
	}
	c.token = otk
	if err := c.needToken(";"); err != nil {
		return err
	}
	if root {
		if err := c.statementEnd(line); err != nil {
			return err
		}
	}
	*ctrls = append(*ctrls, cf)
	c.scan(line)
	return nil
}

func (c *Compiler) letAssign(line *string, root bool,
	ctrls *[]StateController, numVars *int32, names []string, endLine bool) error {
	varis := make([]uint8, len(names))
	for i, n := range names {
		vi, ok := c.vars[n]
		if !ok {
			vi = uint8(*numVars)
			c.vars[n] = vi
			if err := c.inclNumVars(numVars); err != nil {
				return err
			}
		}
		varis[i] = vi
	}
	switch c.scan(line) {
	case "call":
		if err := c.callFunc(line, root, ctrls, varis); err != nil {
			return err
		}
	default:
		otk := c.token
		expr, _, err := c.readSentence(line)
		if err != nil {
			return err
		}
		expr = otk + " " + expr
		otk = c.token
		for i, n := range names {
			var be BytecodeExp
			if i < len(names)-1 {
				be, err = c.argExpression(&expr, VT_SFalse)
				if err != nil {
					return err
				}
				if c.token == "" {
					c.token = otk
				}
				if err := c.needToken(","); err != nil {
					return err
				}
			} else {
				if be, err = c.fullExpression(&expr, VT_SFalse); err != nil {
					return err
				}
			}
			if n == "_" {
				*ctrls = append(*ctrls, StateExpr(be))
			} else {
				*ctrls = append(*ctrls, varAssign{vari: varis[i], be: be})
			}
		}
		c.token = otk
		if err := c.needToken(";"); err != nil {
			return err
		}
		if endLine {
			if root {
				if err := c.statementEnd(line); err != nil {
					return err
				}
			}
			c.scan(line)
		}
	}
	return nil
}

func (c *Compiler) stateBlock(line *string, bl *StateBlock, root bool,
	sbc *StateBytecode, ctrls *[]StateController, numVars *int32) error {
	c.scan(line)
	for {
		switch c.token {
		case "varset", "varadd", "parentvarset", "parentvaradd", "rootvarset", "rootvaradd":
		// Break
		case "", "[":
			if !root {
				return c.wrongClosureToken()
			}
			return nil
		case "}", "case", "default":
			if root {
				return c.wrongClosureToken()
			}
			return nil
		case "for", "if", "ignorehitpause", "persistent", "switch", "while":
			if sbl, err := c.subBlock(line, root, sbc, numVars,
				bl != nil && bl.ctrlsIgnorehitpause, bl != nil && bl.nestedInLoop); err != nil {
				return err
			} else {
				if bl != nil && sbl.ignorehitpause >= -1 {
					bl.ignorehitpause = -1
				}
				*ctrls = append(*ctrls, *sbl)
			}
			continue
		case "call":
			if err := c.callFunc(line, root, ctrls, nil); err != nil {
				return err
			}
			continue
		case "break", "continue":
			if bl.nestedInLoop {
				switch c.token {
				case "break":
					*ctrls = append(*ctrls, LoopBreak{})
				case "continue":
					*ctrls = append(*ctrls, LoopContinue{})
				}
				c.scan(line)
				if err := c.needToken(";"); err != nil {
					return err
				}
				if root {
					if err := c.statementEnd(line); err != nil {
						return err
					}
				}
				c.scan(line)
			} else {
				return Error(fmt.Sprintf("%v can only be used inside a loop block", c.token))
			}
			continue
		case "let":
			names, err := c.varNames("=", line)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				return c.wrongClosureToken()
			}
			if err := c.letAssign(line, root, ctrls, numVars, names, true); err != nil {
				return err
			}
			continue
		default:
			scf, ok := c.scmap[c.token]
			// Check the usage of the name 'helper' since it is used in both the State Controller and Redirect
			if c.token == "helper" {
				// peek ahead to see if this is a redirect
				c.scan(line)
				if len(c.token) > 0 {
					if c.token[0] == ',' || c.token[0] == '(' {
						ok = false
					}
				}
				// reset things to "undo" the peek ahead
				*line = (c.token + (*line))
				c.token = "helper"
			}
			if ok {
				scname := c.token
				c.scan(line)
				if err := c.needToken("{"); err != nil {
					return err
				}
				is, sc := NewIniSection(), newStateControllerBase()
				if err := c.readKeyValue(is, "}", line); err != nil {
					return err
				}
				for c.token != "}" {
					switch c.token {
					case ";":
						if err := c.readKeyValue(is, "}", line); err != nil {
							return err
						}
					default:
						return c.wrongClosureToken()
					}
				}
				if root {
					if err := c.statementEnd(line); err != nil {
						return err
					}
				}
				if scname == "explod" || scname == "modifyexplod" {
					if err := c.paramValue(is, sc, "ignorehitpause",
						explod_ignorehitpause, VT_Bool, 1, false); err != nil {
						return err
					}
				}
				if sctrl, err := scf(is, sc, -1); err != nil {
					return err
				} else {
					*ctrls = append(*ctrls, sctrl)
				}
				c.scan(line)
				continue
			} else {
				otk := c.token
				expr, assign, err := c.readSentence(line)
				if err != nil {
					return err
				}
				expr = otk + " " + expr
				otk = c.token
				if stex, err := c.fullExpression(&expr, VT_SFalse); err != nil {
					return err
				} else {
					*ctrls = append(*ctrls, StateExpr(stex))
				}
				c.token = otk
				if err := c.needToken(";"); err != nil {
					return err
				}
				if !assign {
					return Error("Expression with unused value")
				}
				if root {
					if err := c.statementEnd(line); err != nil {
						return err
					}
				}
				c.scan(line)
				continue
			}
		}
		break
	}
	return c.wrongClosureToken()
}

// Compile a ZSS state
func (c *Compiler) stateCompileZ(states map[int32]StateBytecode,
	filename, src string, constants map[string]float32) error {
	defer func(oime bool) {
		sys.ignoreMostErrors = oime
	}(sys.ignoreMostErrors)
	sys.ignoreMostErrors = false
	c.block = nil
	c.lines, c.i = SplitAndTrim(src, "\n"), 0
	c.linechan = make(chan *string)
	endchan := make(chan bool, 1)
	stop := func() int {
		if c.linechan == nil {
			return 0
		}
		endchan <- true
		lineOffset := 1
		for {
			if sp := <-c.linechan; sp != nil && *sp == "\n" {
				close(endchan)
				close(c.linechan)
				c.linechan = nil
				return c.i + lineOffset
			}
			lineOffset--
		}
	}
	defer stop()
	go func() {
		i := c.i
		for {
			select {
			case <-endchan:
				str := "\n"
				c.linechan <- &str
				return
			default:
			}
			var sp *string
			if i < len(c.lines) {
				str := strings.TrimSpace(c.lines[i])
				sp = &str
				c.i = i
				i++
			}
			c.linechan <- sp
		}
	}()
	errmes := func(err error) error {
		return Error(fmt.Sprintf("%v:%v:\n%v", filename, stop(), err.Error()))
	}
	existInThisFile := make(map[int32]bool)
	funcExistInThisFile := make(map[string]bool)
	var line string
	c.token = ""
	for {
		if c.token == "" {
			c.scan(&line)
			if c.token == "" {
				break
			}
		}
		if c.token != "[" {
			return errmes(c.wrongClosureToken())
		}
		switch c.scan(&line) {
		case "":
			return errmes(c.wrongClosureToken())
		case "statedef":
			var err error
			if c.stateNo, err = c.scanStateDef(&line, constants); err != nil {
				return errmes(err)
			}
			c.scan(&line)
			if existInThisFile[c.stateNo] {
				if c.stateNo == -10 {
					return errmes(Error(fmt.Sprintf("State +1 overloaded")))
				} else {
					return errmes(Error(fmt.Sprintf("State %v overloaded", c.stateNo)))
				}
			}
			existInThisFile[c.stateNo] = true
			is := NewIniSection()
			for c.token != "]" {
				switch c.token {
				case ";":
					if err := c.readKeyValue(is, "]", &line); err != nil {
						return errmes(err)
					}
				default:
					return errmes(c.wrongClosureToken())
				}
			}
			sbc := newStateBytecode(c.playerNo)
			if _, ok := states[c.stateNo]; ok && c.stateNo < 0 {
				*sbc = states[c.stateNo]
			}
			c.vars = make(map[string]uint8)
			if err := c.stateDef(is, sbc); err != nil {
				return errmes(err)
			}
			if err := c.statementEnd(&line); err != nil {
				return errmes(err)
			}
			if err := c.stateBlock(&line, &sbc.block, true,
				sbc, &sbc.block.ctrls, &sbc.numVars); err != nil {
				return errmes(err)
			}
			if _, ok := states[c.stateNo]; !ok || c.stateNo < 0 {
				states[c.stateNo] = *sbc
			}
		case "function":
			name := c.scan(&line)
			if name == "" || name == "(" || name == "]" {
				return errmes(c.wrongClosureToken())
			}
			if err := c.varNameCheck(name); err != nil {
				return errmes(err)
			}
			if funcExistInThisFile[name] {
				return errmes(Error("Function already defined in the same file: " + name))
			}
			funcExistInThisFile[name] = true
			c.scan(&line)
			if err := c.needToken("("); err != nil {
				return errmes(err)
			}
			fun := bytecodeFunction{}
			c.vars = make(map[string]uint8)
			if args, err := c.varNames(")", &line); err != nil {
				return errmes(err)
			} else {
				for _, a := range args {
					c.vars[a] = uint8(fun.numVars)
					if err := c.inclNumVars(&fun.numVars); err != nil {
						return errmes(err)
					}
				}
				fun.numArgs = int32(len(args))
			}
			if rets, err := c.varNames("]", &line); err != nil {
				return errmes(err)
			} else {
				for _, r := range rets {
					if r == "_" {
						return errmes(Error("The return value name is _"))
					} else if _, ok := c.vars[r]; ok {
						return errmes(Error("Duplicate name: " + r))
					} else {
						c.vars[r] = uint8(fun.numVars)
					}
					if err := c.inclNumVars(&fun.numVars); err != nil {
						return errmes(err)
					}
				}
				fun.numRets = int32(len(rets))
			}
			if err := c.stateBlock(&line, nil, true,
				nil, &fun.ctrls, &fun.numVars); err != nil {
				return errmes(err)
			}
			if _, ok := c.funcs[name]; ok {
				continue
				//return errmes(Error("Function already defined in other file: " + name))
			}
			c.funcs[name] = fun
			//c.funcUsed[name] = true
		default:
			return errmes(Error("Unrecognized section (group) name: " + c.token))
		}
	}
	return nil
}

// Compile a character definition file
func (c *Compiler) Compile(pn int, def string, constants map[string]float32) (map[int32]StateBytecode, error) {
	c.playerNo = pn
	states := make(map[int32]StateBytecode)

	// Load initial data from definition file
	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	lines, i, cmd, stcommon := SplitAndTrim(str, "\n"), 0, "", ""
	var st []string
	info, files := true, true
	for i < len(lines) {
		// Parse each ini section
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			// Read info section for the Mugen/Ikemen version of the character
			if info {
				info = false
				var ok bool
				var str string
				// Clear then read MugenVersion
				sys.cgi[pn].mugenver = [2]uint16{}
				sys.cgi[pn].mugenverF = 0
				if str, ok = is["mugenversion"]; ok {
					sys.cgi[pn].mugenver, sys.cgi[pn].mugenverF = parseMugenVersion(str)
				}
				// Clear then read IkemenVersion
				sys.cgi[pn].ikemenver = [3]uint16{}
				sys.cgi[pn].ikemenverF = 0
				if str, ok = is["ikemenversion"]; ok {
					sys.cgi[pn].ikemenver, sys.cgi[pn].ikemenverF = parseIkemenVersion(str)
				}
				// Ikemen characters adopt Mugen 1.1 version as a safeguard
				if sys.cgi[pn].ikemenver[0] != 0 || sys.cgi[pn].ikemenver[1] != 0 {
					sys.cgi[pn].mugenver[0] = 1
					sys.cgi[pn].mugenver[1] = 1
					sys.cgi[pn].mugenverF = 1.1
				}
			}
		case "files":
			// Read files section to find the command and state filenames
			if files {
				files = false
				cmd, stcommon = decodeShiftJIS(is["cmd"]), decodeShiftJIS(is["stcommon"])
				re := regexp.MustCompile(`^st[0-9]*$`)
				// Sorted starting with "st" and followed by "st<num>" in natural order
				for _, v := range SortedKeys(is) {
					if re.MatchString(v) {
						st = append(st, decodeShiftJIS(is[v]))
					}
				}

			}
		}
	}

	// Load the command file
	str = ""
	if len(cmd) > 0 {
		if err := LoadFile(&cmd, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			var err error
			str, err = LoadText(filename)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	for _, key := range SortedKeys(sys.cfg.Common.Cmd) {
		for _, v := range sys.cfg.Common.Cmd[key] {
			if err := LoadFile(&v, []string{def, sys.motifDir, sys.lifebar.def, "", "data/"}, func(filename string) error {
				txt, err := LoadText(filename)
				if err != nil {
					return err
				}
				str += "\n" + txt
				return nil
			}); err != nil {
				return nil, err
			}
		}
	}
	lines, i = SplitAndTrim(str, "\n"), 0

	// Initialize command list data
	if sys.chars[pn][0].cmd == nil {
		sys.chars[pn][0].cmd = make([]CommandList, MaxPlayerNo)
		b := NewInputBuffer()
		for i := range sys.chars[pn][0].cmd {
			sys.chars[pn][0].cmd[i] = *NewCommandList(b)
		}
	}
	c.cmdl = &sys.chars[pn][0].cmd[pn]
	remap, defaults, ckr := true, true, NewCommandKeyRemap()

	var cmds []IniSection
	for i < len(lines) {
		// Read ini sections of command file
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "remap":
			// Read button remapping
			if remap {
				remap = false
				rm := func(name string, k *CommandKey) {
					switch strings.ToLower(is[name]) {
					case "x":
						*k = CK_x
					case "y":
						*k = CK_y
					case "z":
						*k = CK_z
					case "a":
						*k = CK_a
					case "b":
						*k = CK_b
					case "c":
						*k = CK_c
					case "s":
						*k = CK_s
					case "d":
						*k = CK_d
					case "w":
						*k = CK_w
					case "m":
						*k = CK_m
					}
				}
				rm("x", &ckr.x)
				rm("y", &ckr.y)
				rm("z", &ckr.z)
				rm("a", &ckr.a)
				rm("b", &ckr.b)
				rm("c", &ckr.c)
				rm("s", &ckr.s)
				rm("d", &ckr.d)
				rm("w", &ckr.w)
				rm("m", &ckr.m)
			}
		case "defaults":
			// Read default command parameters
			if defaults {
				defaults = false
				is.ReadI32("command.time", &c.cmdl.DefaultTime)
				is.ReadI32("command.steptime", &c.cmdl.DefaultStepTime)
				is.ReadBool("command.autogreater", &c.cmdl.DefaultAutoGreater)
				var i32 int32
				if is.ReadI32("command.buffer.time", &i32) {
					c.cmdl.DefaultBufferTime = Max(1, i32)
				}
				is.ReadBool("command.buffer.hitpause", &c.cmdl.DefaultBufferHitpause)
				is.ReadBool("command.buffer.pauseend", &c.cmdl.DefaultBufferPauseEnd)
				is.ReadBool("command.buffer.shared", &c.cmdl.DefaultBufferShared)
			}
		default:
			// Get command sections
			if len(name) >= 7 && name[:7] == "command" {
				cmds = append(cmds, is)
			}
		}
	}
	// Parse commands
	for _, is := range cmds {
		cm := newCommand()

		// Get name
		name, _, err := is.getText("name")
		if err != nil {
			return nil, Error(fmt.Sprintf("%v:\nname: %v\n%v",
				cmd, name, err.Error()))
		}
		cm.name = name

		// Default parameters
		cm.maxtime = c.cmdl.DefaultTime
		cm.maxbuftime = c.cmdl.DefaultBufferTime
		cm.maxsteptime = c.cmdl.DefaultStepTime
		cm.autogreater = c.cmdl.DefaultAutoGreater
		cm.buffer_hitpause = c.cmdl.DefaultBufferHitpause
		cm.buffer_pauseend = c.cmdl.DefaultBufferPauseEnd
		cm.buffer_shared = c.cmdl.DefaultBufferShared

		// Read specific parameters
		is.ReadI32("time", &cm.maxtime)
		is.ReadI32("steptime", &cm.maxsteptime)
		if cm.maxsteptime <= 0 {
			cm.maxsteptime = cm.maxtime // Default steptime to overall time
		}
		is.ReadBool("autogreater", &cm.autogreater)
		var i32 int32
		if is.ReadI32("buffer.time", &i32) {
			cm.maxbuftime = Max(1, i32)
		}
		is.ReadBool("buffer.hitpause", &cm.buffer_hitpause)
		is.ReadBool("buffer.pauseend", &cm.buffer_pauseend)
		is.ReadBool("buffer.shared", &cm.buffer_shared)

		// Parse the command string and populate steps
		err = cm.ReadCommandSymbols(is["command"], ckr)
		if err != nil {
			if sys.ignoreMostErrors && sys.cgi[pn].ikemenver[0] == 0 && sys.cgi[pn].ikemenver[1] == 0 {
				// Mugen characters ignore command definition errors
			} else {
				return nil, Error(cmd + ":\nname = " + is["name"] +
					"\ncommand = " + is["command"] + "\n" + err.Error())
			}
		}

		c.cmdl.Add(*cm)
	}

	// Compile states
	sys.stringPool[pn].Clear()
	sys.cgi[pn].hitPauseToggleFlagCount = 0
	c.funcUsed = make(map[string]bool)
	// Compile state files
	for _, s := range st {
		if len(s) > 0 {
			if err := c.stateCompile(states, s, []string{def, "", sys.motifDir, "data/"},
				sys.cgi[pn].ikemenver[0] == 0 &&
					sys.cgi[pn].ikemenver[1] == 0, constants); err != nil {
				return nil, err
			}
		}
	}
	// Compile states in command file
	if len(cmd) > 0 {
		if err := c.stateCompile(states, cmd, []string{def, "", sys.motifDir, "data/"},
			sys.cgi[pn].ikemenver[0] == 0 &&
				sys.cgi[pn].ikemenver[1] == 0, constants); err != nil {
			return nil, err
		}
	}
	// Compile states in stcommon state file
	if len(stcommon) > 0 {
		if err := c.stateCompile(states, stcommon, []string{def, "", sys.motifDir, "data/"},
			sys.cgi[pn].ikemenver[0] == 0 &&
				sys.cgi[pn].ikemenver[1] == 0, constants); err != nil {
			return nil, err
		}
	}
	// Compile common states
	for _, key := range SortedKeys(sys.cfg.Common.States) {
		for _, v := range sys.cfg.Common.States[key] {
			if err := c.stateCompile(states, v, []string{def, sys.motifDir, sys.lifebar.def, "", "data/"},
				false, constants); err != nil {
				return nil, err
			}
		}
	}
	return states, nil
}
