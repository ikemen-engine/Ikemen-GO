package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const specialSymbols = " !=<>()|&+-*/%,[]^:;{}#\"\t\r\n"

type expFunc func(out *BytecodeExp, in *string) (BytecodeValue, error)
type scFunc func(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error)
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
}

func newCompiler() *Compiler {
	c := &Compiler{funcs: make(map[string]bytecodeFunction)}
	c.scmap = map[string]scFunc{
		"hitby":                c.hitBy,
		"nothitby":             c.notHitBy,
		"assertspecial":        c.assertSpecial,
		"playsnd":              c.playSnd,
		"changestate":          c.changeState,
		"selfstate":            c.selfState,
		"tagin":                c.tagIn,
		"tagout":               c.tagOut,
		"destroyself":          c.destroySelf,
		"changeanim":           c.changeAnim,
		"changeanim2":          c.changeAnim2,
		"helper":               c.helper,
		"ctrlset":              c.ctrlSet,
		"explod":               c.explod,
		"modifyexplod":         c.modifyExplod,
		"gamemakeanim":         c.gameMakeAnim,
		"posset":               c.posSet,
		"posadd":               c.posAdd,
		"velset":               c.velSet,
		"veladd":               c.velAdd,
		"velmul":               c.velMul,
		"palfx":                c.palFX,
		"allpalfx":             c.allPalFX,
		"bgpalfx":              c.bgPalFX,
		"afterimage":           c.afterImage,
		"afterimagetime":       c.afterImageTime,
		"hitdef":               c.hitDef,
		"reversaldef":          c.reversalDef,
		"projectile":           c.projectile,
		"width":                c.width,
		"sprpriority":          c.sprPriority,
		"varset":               c.varSet,
		"varadd":               c.varAdd,
		"parentvarset":         c.parentVarSet,
		"parentvaradd":         c.parentVarAdd,
		"rootvarset":           c.rootVarSet,
		"rootvaradd":           c.rootVarAdd,
		"turn":                 c.turn,
		"targetfacing":         c.targetFacing,
		"targetbind":           c.targetBind,
		"bindtotarget":         c.bindToTarget,
		"targetlifeadd":        c.targetLifeAdd,
		"targetstate":          c.targetState,
		"targetvelset":         c.targetVelSet,
		"targetveladd":         c.targetVelAdd,
		"targetpoweradd":       c.targetPowerAdd,
		"targetdrop":           c.targetDrop,
		"lifeadd":              c.lifeAdd,
		"lifeset":              c.lifeSet,
		"poweradd":             c.powerAdd,
		"powerset":             c.powerSet,
		"hitvelset":            c.hitVelSet,
		"screenbound":          c.screenBound,
		"posfreeze":            c.posFreeze,
		"envshake":             c.envShake,
		"hitoverride":          c.hitOverride,
		"pause":                c.pause,
		"superpause":           c.superPause,
		"trans":                c.trans,
		"playerpush":           c.playerPush,
		"statetypeset":         c.stateTypeSet,
		"angledraw":            c.angleDraw,
		"angleset":             c.angleSet,
		"angleadd":             c.angleAdd,
		"anglemul":             c.angleMul,
		"envcolor":             c.envColor,
		"displaytoclipboard":   c.displayToClipboard,
		"appendtoclipboard":    c.appendToClipboard,
		"clearclipboard":       c.clearClipboard,
		"makedust":             c.makeDust,
		"attackdist":           c.attackDist,
		"attackmulset":         c.attackMulSet,
		"defencemulset":        c.defenceMulSet,
		"fallenvshake":         c.fallEnvShake,
		"hitfalldamage":        c.hitFallDamage,
		"hitfallvel":           c.hitFallVel,
		"hitfallset":           c.hitFallSet,
		"varrangeset":          c.varRangeSet,
		"remappal":             c.remapPal,
		"stopsnd":              c.stopSnd,
		"sndpan":               c.sndPan,
		"varrandom":            c.varRandom,
		"gravity":              c.gravity,
		"bindtoparent":         c.bindToParent,
		"bindtoroot":           c.bindToRoot,
		"removeexplod":         c.removeExplod,
		"explodbindtime":       c.explodBindTime,
		"movehitreset":         c.moveHitReset,
		"hitadd":               c.hitAdd,
		"hitscaleset":          c.hitScaleSet,
		"offset":               c.offset,
		"victoryquote":         c.victoryQuote,
		"zoom":                 c.zoom,
		"forcefeedback":        c.forceFeedback,
		"null":                 c.null,
		"assertcommand":        c.assertCommand,
		"assertinput":          c.assertInput,
		"dialogue":             c.dialogue,
		"dizzypointsadd":       c.dizzyPointsAdd,
		"dizzypointsset":       c.dizzyPointsSet,
		"dizzyset":             c.dizzySet,
		"guardbreakset":        c.guardBreakSet,
		"guardpointsadd":       c.guardPointsAdd,
		"guardpointsset":       c.guardPointsSet,
		"lifebaraction":        c.lifebarAction,
		"loadfile":             c.loadFile,
		"mapset":               c.mapSet,
		"mapadd":               c.mapAdd,
		"parentmapset":         c.parentMapSet,
		"parentmapadd":         c.parentMapAdd,
		"rootmapset":           c.rootMapSet,
		"rootmapadd":           c.rootMapAdd,
		"teammapset":           c.teamMapSet,
		"teammapadd":           c.teamMapAdd,
		"matchrestart":         c.matchRestart,
		"modifybgctrl":         c.modifyBGCtrl,
		"playbgm":              c.playBgm,
		"printtoconsole":       c.printToConsole,
		"redlifeadd":           c.redLifeAdd,
		"redlifeset":           c.redLifeSet,
		"remapsprite":          c.remapSprite,
		"roundtimeadd":         c.roundTimeAdd,
		"roundtimeset":         c.roundTimeSet,
		"savefile":             c.saveFile,
		"scoreadd":             c.scoreAdd,
		"targetdizzypointsadd": c.targetDizzyPointsAdd,
		"targetguardpointsadd": c.targetGuardPointsAdd,
		"targetredlifeadd":     c.targetRedLifeAdd,
		"targetscoreadd":       c.targetScoreAdd,
		"text":                 c.text,
		"modifystagevar":       c.modifyStageVar,
		"camera":               c.cameraCtrl,
		"height":               c.height,
		"modifychar":           c.modifyChar,
	}
	return c
}

var triggerMap = map[string]int{
	// redirections
	"player":      0,
	"parent":      0,
	"root":        0,
	"helper":      0,
	"target":      0,
	"partner":     0,
	"enemy":       0,
	"enemynear":   0,
	"playerid":    0,
	"p2":          0,
	"stateowner":  0,
	"helperindex": 0,
	// mugen triggers
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
	"helpername":        1,
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
	"ishost":            1,
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
	"movetype":          1,
	"movereversed":      1,
	"name":              1,
	"numenemy":          1,
	"numexplod":         1,
	"numhelper":         1,
	"numpartner":        1,
	"numproj":           1,
	"numprojid":         1,
	"numtarget":         1,
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
	"pos":               1,
	"power":             1,
	"powermax":          1,
	"playeridexist":     1,
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
	"screenpos":         1,
	"screenheight":      1,
	"screenwidth":       1,
	"selfanimexist":     1,
	"sin":               1,
	"stateno":           1,
	"statetype":         1,
	"stagevar":          1,
	"sysfvar":           1,
	"sysvar":            1,
	"tan":               1,
	"teammode":          1,
	"teamside":          1,
	"tickspersecond":    1,
	"time":              1,
	"timemod":           1,
	"topedge":           1,
	"uniqhitcount":      1,
	"var":               1,
	"vel":               1,
	"win":               1,
	"winko":             1,
	"wintime":           1,
	"winperfect":        1,
	// expanded triggers
	"ailevelf":           1,
	"airjumpcount":       1,
	"animelemlength":     1,
	"animlength":         1,
	"attack":             1,
	"alpha":              1,
	"angle":              1,
	"atan2":              1,
	"bgmlength":          1,
	"bgmposition":        1,
	"clamp":              1,
	"combocount":         1,
	"consecutivewins":    1,
	"deg":                1,
	"defence":            1,
	"dizzy":              1,
	"dizzypoints":        1,
	"dizzypointsmax":     1,
	"drawpalno":          1,
	"envshakevar":        1,
	"fighttime":          1,
	"firstattack":        1,
	"float":              1,
	"framespercount":     1,
	"gamemode":           1,
	"getplayerid":        1,
	"groundangle":        1,
	"guardbreak":         1,
	"guardpoints":        1,
	"guardpointsmax":     1,
	"hitoverridden":      1,
	"incustomstate":      1,
	"indialogue":         1,
	"isasserted":         1,
	"lastplayerid":       1,
	"lerp":               1,
	"localscale":         1,
	"majorversion":       1,
	"map":                1,
	"max":                1,
	"memberno":           1,
	"min":                1,
	"movecountered":      1,
	"mugenversion":       1,
	"offset":             1,
	"p5name":             1,
	"p6name":             1,
	"p7name":             1,
	"p8name":             1,
	"pausetime":          1,
	"physics":            1,
	"playerno":           1,
	"prevanim":           1,
	"prevmovetype":       1,
	"prevstatetype":      1,
	"rad":                1,
	"ratiolevel":         1,
	"randomrange":        1,
	"receivedhits":       1,
	"receiveddamage":     1,
	"redlife":            1,
	"reversaldefattr":    1,
	"round":              1,
	"roundtype":          1,
	"scale":              1,
	"sign":               1,
	"score":              1,
	"scoretotal":         1,
	"selfstatenoexist":   1,
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
}

func (c *Compiler) tokenizer(in *string) string {
	return strings.ToLower(c.tokenizerCS(in))
}

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
			return 0, Error("Invalid value: " + string(a))
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
			if sys.ignoreMostErrors && sys.cgi[c.playerNo].mugenver[0] == 1 {
				//if hitdef {
				//	flg = hitdefflg
				//}
				return flg, nil
			}
			return 0, Error("Invalid value: " + a)
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
			return 0, Error("Invalid attribute value: " + att)
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
func (c *Compiler) checkOpeningBracket(in *string) error {
	if c.tokenizer(in) != "(" {
		return Error("Missing '(' after " + c.token)
	}
	c.token = c.tokenizer(in)
	return nil
}

/*
TODO: Case sensitive maps

	func (c *Compiler) checkOpeningBracketCS(in *string) error {
		if c.tokenizerCS(in) != "(" {
			return Error("Missing '(' after " + c.token)
		}
		c.token = c.tokenizerCS(in)
		return nil
	}
*/
func (c *Compiler) checkClosingBracket() error {
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
		err = Error("Missing '[' or '('")
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
				return 0, Error("Error reading number")
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
		err = Error("Missing ','")
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
		err = Error("Missing ']' or ')'")
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
		if err := c.checkClosingBracket(); err != nil {
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
	eqne := func(f func() error) error {
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
	nameSub := func(opc OpCode) error {
		return eqne(func() error {
			if err := text(); err != nil {
				return err
			}
			out.append(OC_const_)
			out.appendI32Op(opc, int32(sys.stringPool[c.playerNo].Add(
				strings.ToLower(c.token))))
			return nil
		})
	}
	nameSubEx := func(opc OpCode) error {
		return eqne(func() error {
			if err := text(); err != nil {
				return err
			}
			out.append(OC_ex_)
			out.appendI32Op(opc, int32(sys.stringPool[c.playerNo].Add(
				strings.ToLower(c.token))))
			return nil
		})
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
	case "root", "player", "parent", "helper", "target", "partner",
		"enemy", "enemynear", "playerid", "p2", "stateowner", "helperindex":
		switch c.token {
		case "parent":
			opc = OC_parent
			c.token = c.tokenizer(in)
		case "root":
			opc = OC_root
			c.token = c.tokenizer(in)
		case "p2":
			opc = OC_p2
			c.token = c.tokenizer(in)
		case "stateowner":
			opc = OC_stateowner
			c.token = c.tokenizer(in)
		default:
			switch c.token {
			case "player":
				opc = OC_player
			case "helper":
				opc = OC_helper
			case "target":
				opc = OC_target
			case "partner":
				opc = OC_partner
			case "enemy":
				opc = OC_enemy
			case "enemynear":
				opc = OC_enemynear
			case "playerid":
				opc = OC_playerid
			case "helperindex":
				opc = OC_helperindex
			}
			c.token = c.tokenizer(in)
			if c.token == "(" {
				c.token = c.tokenizer(in)
				if bv1, err = c.expBoolOr(&be1, in); err != nil {
					return bvNone(), err
				}
				if err := c.checkClosingBracket(); err != nil {
					return bvNone(), err
				}
				c.token = c.tokenizer(in)
				be1.appendValue(bv1)
			} else {
				switch opc {
				case OC_helper, OC_target:
					be1.appendValue(BytecodeInt(-1))
				case OC_partner, OC_enemy, OC_enemynear:
					be1.appendValue(BytecodeInt(0))
				case OC_player:
					return bvNone(), Error("Missing '(' after player")
				case OC_playerid:
					return bvNone(), Error("Missing '(' after playerid")
				case OC_helperindex:
					return bvNone(), Error("Missing '(' after helperindex")
				}
			}
			if rd {
				out.appendI32Op(OC_nordrun, int32(len(be1)))
			}
			out.append(be1...)
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
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
					out.append(OC_rdreset)
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
				out.append(OC_rdreset)
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
				out.append(OC_rdreset)
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
				out.append(OC_rdreset)
			}
			out.append(be1...)
		}
		if err := c.checkClosingBracket(); err != nil {
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
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv3, err = c.expBoolOr(&be3, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
		if err := nameSub(OC_const_authorname); err != nil {
			return bvNone(), err
		}
	case "backedge":
		out.append(OC_backedge)
	case "backedgebodydist":
		out.append(OC_backedgebodydist)
	case "backedgedist":
		out.append(OC_backedgedist)
	case "bgmlength":
		out.append(OC_ex_, OC_ex_bgmlength)
	case "bgmposition":
		out.append(OC_ex_, OC_ex_bgmposition)
	case "bottomedge":
		out.append(OC_bottomedge)
	case "camerapos":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_camerapos_x)
		case "y":
			out.append(OC_camerapos_y)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "camerazoom":
		out.append(OC_camerazoom)
	case "canrecover":
		out.append(OC_canrecover)
	case "command":
		if err := eqne(func() error {
			if err := text(); err != nil {
				return err
			}
			_, ok := c.cmdl.Names[c.token]
			if !ok {
				return Error("Command doesn't exist: " + c.token)
			}
			i := sys.stringPool[c.playerNo].Add(c.token)
			out.appendI32Op(OC_command, int32(i))
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "const":
		if err := c.checkOpeningBracket(in); err != nil {
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
		case "size.height", "size.height.stand": // Latter is also accepted for consistency's sake
			out.append(OC_const_size_height_stand)
		case "size.height.crouch":
			out.append(OC_const_size_height_crouch)
		case "size.height.air.top":
			out.append(OC_const_size_height_air_top)
		case "size.height.air.bottom":
			out.append(OC_const_size_height_air_bottom)
		case "size.height.down":
			out.append(OC_const_size_height_down)
		case "size.attack.dist":
			out.append(OC_const_size_attack_dist)
		case "size.attack.z.width.back":
			out.append(OC_const_size_attack_z_width_back)
		case "size.attack.z.width.front":
			out.append(OC_const_size_attack_z_width_front)
		case "size.proj.attack.dist":
			out.append(OC_const_size_proj_attack_dist)
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
		case "size.z.width":
			out.append(OC_const_size_z_width)
		case "size.z.enable":
			out.append(OC_const_size_z_enable)
		case "size.classicpushbox":
			out.append(OC_const_size_classicpushbox)
		case "velocity.walk.fwd.x":
			out.append(OC_const_velocity_walk_fwd_x)
		case "velocity.walk.back.x":
			out.append(OC_const_velocity_walk_back_x)
		case "velocity.walk.up.x":
			out.append(OC_const_velocity_walk_up_x)
		case "velocity.walk.down.x":
			out.append(OC_const_velocity_walk_down_x)
		case "velocity.run.fwd.x":
			out.append(OC_const_velocity_run_fwd_x)
		case "velocity.run.fwd.y":
			out.append(OC_const_velocity_run_fwd_y)
		case "velocity.run.back.x":
			out.append(OC_const_velocity_run_back_x)
		case "velocity.run.back.y":
			out.append(OC_const_velocity_run_back_y)
		case "velocity.run.up.x":
			out.append(OC_const_velocity_run_up_x)
		case "velocity.run.up.y":
			out.append(OC_const_velocity_run_up_y)
		case "velocity.run.down.x":
			out.append(OC_const_velocity_run_down_x)
		case "velocity.run.down.y":
			out.append(OC_const_velocity_run_down_y)
		case "velocity.jump.y":
			out.append(OC_const_velocity_jump_y)
		case "velocity.jump.neu.x":
			out.append(OC_const_velocity_jump_neu_x)
		case "velocity.jump.back.x":
			out.append(OC_const_velocity_jump_back_x)
		case "velocity.jump.fwd.x":
			out.append(OC_const_velocity_jump_fwd_x)
		case "velocity.jump.up.x":
			out.append(OC_const_velocity_jump_up_x)
		case "velocity.jump.down.x":
			out.append(OC_const_velocity_jump_down_x)
		case "velocity.runjump.back.x":
			out.append(OC_const_velocity_runjump_back_x)
		case "velocity.runjump.back.y":
			out.append(OC_const_velocity_runjump_back_y)
		case "velocity.runjump.y":
			out.append(OC_const_velocity_runjump_y)
		case "velocity.runjump.fwd.x":
			out.append(OC_const_velocity_runjump_fwd_x)
		case "velocity.runjump.up.x":
			out.append(OC_const_velocity_runjump_up_x)
		case "velocity.runjump.down.x":
			out.append(OC_const_velocity_runjump_down_x)
		case "velocity.airjump.y":
			out.append(OC_const_velocity_airjump_y)
		case "velocity.airjump.neu.x":
			out.append(OC_const_velocity_airjump_neu_x)
		case "velocity.airjump.back.x":
			out.append(OC_const_velocity_airjump_back_x)
		case "velocity.airjump.fwd.x":
			out.append(OC_const_velocity_airjump_fwd_x)
		case "velocity.airjump.up.x":
			out.append(OC_const_velocity_airjump_up_x)
		case "velocity.airjump.down.x":
			out.append(OC_const_velocity_airjump_down_x)
		case "velocity.air.gethit.groundrecover.x":
			out.append(OC_const_velocity_air_gethit_groundrecover_x)
		case "velocity.air.gethit.groundrecover.y":
			out.append(OC_const_velocity_air_gethit_groundrecover_y)
		case "velocity.air.gethit.airrecover.mul.x":
			out.append(OC_const_velocity_air_gethit_airrecover_mul_x)
		case "velocity.air.gethit.airrecover.mul.y":
			out.append(OC_const_velocity_air_gethit_airrecover_mul_y)
		case "velocity.air.gethit.airrecover.add.x":
			out.append(OC_const_velocity_air_gethit_airrecover_add_x)
		case "velocity.air.gethit.airrecover.add.y":
			out.append(OC_const_velocity_air_gethit_airrecover_add_y)
		case "velocity.air.gethit.airrecover.back":
			out.append(OC_const_velocity_air_gethit_airrecover_back)
		case "velocity.air.gethit.airrecover.fwd":
			out.append(OC_const_velocity_air_gethit_airrecover_fwd)
		case "velocity.air.gethit.airrecover.up":
			out.append(OC_const_velocity_air_gethit_airrecover_up)
		case "velocity.air.gethit.airrecover.down":
			out.append(OC_const_velocity_air_gethit_airrecover_down)
		case "velocity.air.gethit.ko.add.x":
			out.append(OC_const_velocity_air_gethit_ko_add_x)
		case "velocity.air.gethit.ko.add.y":
			out.append(OC_const_velocity_air_gethit_ko_add_y)
		case "velocity.air.gethit.ko.ymin":
			out.append(OC_const_velocity_air_gethit_ko_ymin)
		case "velocity.ground.gethit.ko.xmul":
			out.append(OC_const_velocity_ground_gethit_ko_xmul)
		case "velocity.ground.gethit.ko.add.x":
			out.append(OC_const_velocity_ground_gethit_ko_add_x)
		case "velocity.ground.gethit.ko.add.y":
			out.append(OC_const_velocity_ground_gethit_ko_add_y)
		case "velocity.ground.gethit.ko.ymin":
			out.append(OC_const_velocity_ground_gethit_ko_ymin)
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
	case "ctrl":
		out.append(OC_ctrl)
	case "drawgame":
		out.append(OC_ex_, OC_ex_drawgame)
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
	case "gametime":
		out.append(OC_gametime)
	case "gamewidth":
		out.append(OC_gamewidth)
	case "gethitvar":
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		switch c.token {
		case "xveladd":
			bv.SetF(0)
		case "yveladd":
			bv.SetF(0)
		case "zoff":
			bv.SetF(0)
		case "fall.envshake.dir":
			bv.SetI(0)
		default:
			out.append(OC_ex_)
			switch c.token {
			case "animtype":
				out.append(OC_ex_gethitvar_animtype)
			case "air.animtype":
				out.append(OC_ex_gethitvar_air_animtype)
			case "ground.animtype":
				out.append(OC_ex_gethitvar_ground_animtype)
			case "fall.animtype":
				out.append(OC_ex_gethitvar_fall_animtype)
			case "type":
				out.append(OC_ex_gethitvar_type)
			case "airtype":
				out.append(OC_ex_gethitvar_airtype)
			case "groundtype":
				out.append(OC_ex_gethitvar_groundtype)
			case "damage":
				out.append(OC_ex_gethitvar_damage)
			case "hitcount":
				out.append(OC_ex_gethitvar_hitcount)
			case "fallcount":
				out.append(OC_ex_gethitvar_fallcount)
			case "hitshaketime":
				out.append(OC_ex_gethitvar_hitshaketime)
			case "hittime":
				out.append(OC_ex_gethitvar_hittime)
			case "slidetime":
				out.append(OC_ex_gethitvar_slidetime)
			case "ctrltime":
				out.append(OC_ex_gethitvar_ctrltime)
			case "recovertime":
				out.append(OC_ex_gethitvar_recovertime)
			case "xoff":
				out.append(OC_ex_gethitvar_xoff)
			case "yoff":
				out.append(OC_ex_gethitvar_yoff)
			case "xvel":
				out.append(OC_ex_gethitvar_xvel)
			case "yvel":
				out.append(OC_ex_gethitvar_yvel)
			case "yaccel":
				out.append(OC_ex_gethitvar_yaccel)
			case "hitid", "chainid":
				out.append(OC_ex_gethitvar_chainid)
			case "guarded":
				out.append(OC_ex_gethitvar_guarded)
			case "isbound":
				out.append(OC_ex_gethitvar_isbound)
			case "fall":
				out.append(OC_ex_gethitvar_fall)
			case "fall.damage":
				out.append(OC_ex_gethitvar_fall_damage)
			case "fall.xvel":
				out.append(OC_ex_gethitvar_fall_xvel)
			case "fall.yvel":
				out.append(OC_ex_gethitvar_fall_yvel)
			case "fall.recover":
				out.append(OC_ex_gethitvar_fall_recover)
			case "fall.time":
				out.append(OC_ex_gethitvar_fall_time)
			case "fall.recovertime":
				out.append(OC_ex_gethitvar_fall_recovertime)
			case "fall.kill":
				out.append(OC_ex_gethitvar_fall_kill)
			case "fall.envshake.time":
				out.append(OC_ex_gethitvar_fall_envshake_time)
			case "fall.envshake.freq":
				out.append(OC_ex_gethitvar_fall_envshake_freq)
			case "fall.envshake.ampl":
				out.append(OC_ex_gethitvar_fall_envshake_ampl)
			case "fall.envshake.phase":
				out.append(OC_ex_gethitvar_fall_envshake_phase)
			case "fall.envshake.mul":
				out.append(OC_ex_gethitvar_fall_envshake_mul)
			case "attr":
				out.append(OC_ex_gethitvar_attr)
			case "dizzypoints":
				out.append(OC_ex_gethitvar_dizzypoints)
			case "guardpoints":
				out.append(OC_ex_gethitvar_guardpoints)
			case "id":
				out.append(OC_ex_gethitvar_id)
			case "playerno":
				out.append(OC_ex_gethitvar_playerno)
			case "redlife":
				out.append(OC_ex_gethitvar_redlife)
			case "score":
				out.append(OC_ex_gethitvar_score)
			case "hitdamage":
				out.append(OC_ex_gethitvar_hitdamage)
			case "guarddamage":
				out.append(OC_ex_gethitvar_guarddamage)
			case "hitpower":
				out.append(OC_ex_gethitvar_hitpower)
			case "guardpower":
				out.append(OC_ex_gethitvar_guardpower)
			case "kill":
				out.append(OC_ex_gethitvar_kill)
			case "priority":
				out.append(OC_ex_gethitvar_priority)
			default:
				return bvNone(), Error("Invalid data: " + c.token)
			}
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingBracket(); err != nil {
			return bvNone(), err
		}
	case "hitcount":
		out.append(OC_hitcount)
	case "hitdefattr":
		hda := func() error {
			if attr, err := c.trgAttr(in); err != nil {
				return err
			} else {
				out.appendI32Op(OC_hitdefattr, attr)
			}
			return nil
		}
		// if sys.cgi[c.playerNo].mugenver[0] == 1 {
		// if err := eqne(hda); err != nil {
		// return bvNone(), err
		// }
		// } else {
		// if not, err := c.checkEquality(in); err != nil {
		// if sys.ignoreMostErrors {
		// out.appendValue(BytecodeBool(false))
		// } else {
		// return bvNone(), err
		// }
		// } else if err := hda(); err != nil {
		// return bvNone(), err
		// } else if not && !sys.ignoreMostErrors {
		// return bvNone(), Error("hitdefattr doesn't support '!=' in this mugenversion")
		// }
		// }
		if err := eqne(hda); err != nil {
			return bvNone(), err
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
			bv = BytecodeFloat(0)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "id":
		out.append(OC_id)
	case "inguarddist":
		out.append(OC_inguarddist)
	case "ishelper":
		if _, err := c.oneArg(out, in, rd, true, BytecodeInt(math.MinInt32)); err != nil {
			return bvNone(), err
		}
		out.append(OC_ishelper)
	case "ishometeam":
		out.append(OC_ex_, OC_ex_ishometeam)
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
				return Error(trname + " value is not specified")
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
				return Error("Invalid value: " + c.token)
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
		if err := nameSub(opc); err != nil {
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
	case "numtarget":
		if _, err := c.oneArg(out, in, rd, true, BytecodeInt(-1)); err != nil {
			return bvNone(), err
		}
		out.append(OC_numtarget)
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
			return bvNone(), Error("Invalid data: " + c.token)
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
	case "roundno":
		out.append(OC_ex_, OC_ex_roundno)
	case "roundsexisted":
		out.append(OC_ex_, OC_ex_roundsexisted)
	case "roundstate":
		out.append(OC_roundstate)
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
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "screenwidth":
		out.append(OC_screenwidth)
	case "selfanimexist":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_selfanimexist)
	case "stateno", "p2stateno":
		if c.token == "p2stateno" {
			out.appendI32Op(OC_p2, 1)
		}
		out.append(OC_stateno)
	case "statetype", "p2statetype", "prevstatetype":
		trname := c.token
		if err := eqne2(func(not bool) error {
			if len(c.token) == 0 {
				return Error(trname + " value is not specified")
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
				return Error("Invalid value: " + c.token)
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
	case "stagevar":
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		svname := c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingBracket(); err != nil {
			return bvNone(), err
		}
		isStr := false
		switch svname {
		case "info.name":
			opc = OC_const_stagevar_info_name
			isStr = true
		case "info.displayname":
			opc = OC_const_stagevar_info_displayname
			isStr = true
		case "info.author":
			opc = OC_const_stagevar_info_author
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
		case "camera.startzoom":
			opc = OC_const_stagevar_camera_startzoom
		case "camera.zoomout":
			opc = OC_const_stagevar_camera_zoomout
		case "camera.zoomin":
			opc = OC_const_stagevar_camera_zoomin
		case "camera.ytension.enable":
			opc = OC_const_stagevar_camera_ytension_enable
		case "playerinfo.leftbound":
			opc = OC_const_stagevar_playerinfo_leftbound
		case "playerinfo.rightbound":
			opc = OC_const_stagevar_playerinfo_rightbound
		case "scaling.topscale":
			opc = OC_const_stagevar_scaling_topscale
		case "bound.screenleft":
			opc = OC_const_stagevar_bound_screenleft
		case "bound.screenright":
			opc = OC_const_stagevar_bound_screenright
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
		case "reflection.intensity":
			opc = OC_const_stagevar_reflection_intensity
		default:
			return bvNone(), Error("Invalid data: " + svname)
		}
		if isStr {
			if err := nameSub(opc); err != nil {
				return bvNone(), err
			}
		} else {
			out.append(OC_const_)
			out.append(opc)
		}
	case "teammode":
		if err := eqne(func() error {
			if len(c.token) == 0 {
				return Error("teammode value is not specified")
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
				return Error("Invalid value: " + c.token)
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
			return bvNone(), Error("Invalid data: " + c.token)
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
			return bvNone(), Error("animelem doesn't support '!='")
		}
		if c.token == "-" {
			return bvNone(), Error("'-' should not be used")
		}
		if n, err = c.integer2(in); err != nil {
			return bvNone(), err
		}
		if n <= 0 {
			return bvNone(), Error("animelem must be greater than 0")
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
			return bvNone(), Error("timemod doesn't support '!='")
		}
		if c.token == "-" {
			return bvNone(), Error("'-' should not be used")
		}
		if n, err = c.integer2(in); err != nil {
			return bvNone(), err
		}
		if n <= 0 {
			return bvNone(), Error("timemod must be greater than 0")
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
			bv = BytecodeFloat(0)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "p2bodydist":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_p2bodydist_x)
		case "y":
			out.append(OC_ex_, OC_ex_p2bodydist_y)
		case "z":
			bv = BytecodeFloat(0)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "rootdist":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_rootdist_x)
		case "y":
			out.append(OC_ex_, OC_ex_rootdist_y)
		case "z":
			bv = BytecodeFloat(0)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "parentdist":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_parentdist_x)
		case "y":
			out.append(OC_ex_, OC_ex_parentdist_y)
		case "z":
			bv = BytecodeFloat(0)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
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
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
	case "randomrange", "rand": // rand is deprecated, kept for backward compatibility
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv3, err = c.expBoolOr(&be3, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return bvNone(), err
		}
		if c.token != "," {
			return bvNone(), Error("Missing ','")
		}
		c.token = c.tokenizer(in)
		if bv3, err = c.expBoolOr(&be3, in); err != nil {
			return bvNone(), err
		}
		if err := c.checkClosingBracket(); err != nil {
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
	case "animelemlength":
		out.append(OC_ex_, OC_ex_animelemlength)
	case "animframe":
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_)
		switch c.token {
		case "alphadest":
			out.append(OC_ex_animframe_alphadest)
		case "angle":
			out.append(OC_ex_animframe_angle)
		case "alphasource":
			out.append(OC_ex_animframe_alphasource)
		case "group":
			out.append(OC_ex_animframe_group)
		case "hflip":
			out.append(OC_ex_animframe_hflip)
		case "image":
			out.append(OC_ex_animframe_image)
		case "time":
			out.append(OC_ex_animframe_time)
		case "vflip":
			out.append(OC_ex_animframe_vflip)
		case "xoffset":
			out.append(OC_ex_animframe_xoffset)
		case "xscale":
			out.append(OC_ex_animframe_xscale)
		case "yoffset":
			out.append(OC_ex_animframe_yoffset)
		case "yscale":
			out.append(OC_ex_animframe_yscale)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingBracket(); err != nil {
			return bvNone(), err
		}
	case "animlength":
		out.append(OC_ex_, OC_ex_animlength)
	case "attack":
		out.append(OC_ex_, OC_ex_attack)
	case "combocount":
		out.append(OC_ex_, OC_ex_combocount)
	case "consecutivewins":
		out.append(OC_ex_, OC_ex_consecutivewins)
	case "defence":
		out.append(OC_ex_, OC_ex_defence)
	case "dizzy":
		out.append(OC_ex_, OC_ex_dizzy)
	case "dizzypoints":
		out.append(OC_ex_, OC_ex_dizzypoints)
	case "dizzypointsmax":
		out.append(OC_ex_, OC_ex_dizzypointsmax)
	case "envshakevar":
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_)
		switch c.token {
		case "time":
			out.append(OC_ex_envshakevar_time)
		case "freq":
			out.append(OC_ex_envshakevar_freq)
		case "ampl":
			out.append(OC_ex_envshakevar_ampl)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingBracket(); err != nil {
			return bvNone(), err
		}
	case "fighttime":
		out.append(OC_ex_, OC_ex_fighttime)
	case "firstattack":
		out.append(OC_ex_, OC_ex_firstattack)
	case "framespercount":
		out.append(OC_ex_, OC_ex_framespercount)
	case "gamemode":
		if err := nameSubEx(OC_ex_gamemode); err != nil {
			return bvNone(), err
		}
	case "getplayerid":
		if _, err := c.oneArg(out, in, rd, true); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_, OC_ex_getplayerid)
	case "groundangle":
		out.append(OC_ex_, OC_ex_groundangle)
	case "guardbreak":
		out.append(OC_ex_, OC_ex_guardbreak)
	case "guardpoints":
		out.append(OC_ex_, OC_ex_guardpoints)
	case "guardpointsmax":
		out.append(OC_ex_, OC_ex_guardpointsmax)
	case "helpername":
		if err := nameSubEx(OC_ex_helpername); err != nil {
			return bvNone(), err
		}
	case "hitoverridden":
		out.append(OC_ex_, OC_ex_hitoverridden)
	case "incustomstate":
		out.append(OC_ex_, OC_ex_incustomstate)
	case "indialogue":
		out.append(OC_ex_, OC_ex_indialogue)
	case "isasserted":
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_)
		switch c.token {
		case "nostandguard":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nostandguard))
		case "nocrouchguard":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nocrouchguard))
		case "noairguard":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noairguard))
		case "noshadow":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noshadow))
		case "invisible":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_invisible))
		case "unguardable":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_unguardable))
		case "nojugglecheck":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nojugglecheck))
		case "noautoturn":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noautoturn))
		case "nowalk":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nowalk))
		case "nobrake":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nobrake))
		case "nocrouch":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nocrouch))
		case "nostand":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nostand))
		case "nojump":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nojump))
		case "noairjump":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noairjump))
		case "nohardcodedkeys":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nohardcodedkeys))
		case "nogetupfromliedown":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nogetupfromliedown))
		case "nofastrecoverfromliedown":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofastrecoverfromliedown))
		case "nofallcount":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofallcount))
		case "nofalldefenceup":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nofalldefenceup))
		case "noturntarget":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noturntarget))
		case "noinput":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noinput))
		case "nopowerbardisplay":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nopowerbardisplay))
		case "autoguard":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_autoguard))
		case "animfreeze":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_animfreeze))
		case "postroundinput":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_postroundinput))
		case "nohitdamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nohitdamage))
		case "noguarddamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noguarddamage))
		case "nodizzypointsdamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nodizzypointsdamage))
		case "noguardpointsdamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noguardpointsdamage))
		case "noredlifedamage":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noredlifedamage))
		case "nomakedust":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nomakedust))
		case "noko":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noko))
		case "noguardko":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noguardko))
		case "nokovelocity":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nokovelocity))
		case "noailevel":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_noailevel))
		case "nointroreset":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_nointroreset))
		case "immovable":
			out.appendI64Op(OC_ex_isassertedchar, int64(ASF_immovable))
		case "intro":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_intro))
		case "roundnotover":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_roundnotover))
		case "nomusic":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nomusic))
		case "nobardisplay":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nobardisplay))
		case "nobg":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nobg))
		case "nofg":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nofg))
		case "globalnoshadow":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_globalnoshadow))
		case "timerfreeze":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_timerfreeze))
		case "nokosnd":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nokosnd))
		case "nokoslow":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nokoslow))
		case "globalnoko":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_noko))
		case "roundnotskip":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_roundnotskip))
		case "roundfreeze":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_roundfreeze))
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.checkClosingBracket(); err != nil {
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
	case "localscale":
		out.append(OC_ex_, OC_ex_localscale)
	case "majorversion":
		out.append(OC_ex_, OC_ex_majorversion)
	case "map":
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		var m string = c.token
		c.token = c.tokenizer(in)
		if err := c.checkClosingBracket(); err != nil {
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
	case "movecountered":
		out.append(OC_ex_, OC_ex_movecountered)
	case "mugenversion":
		out.append(OC_ex_, OC_ex_mugenversion)
	case "pausetime":
		out.append(OC_ex_, OC_ex_pausetime)
	case "physics":
		if err := eqne(func() error {
			if len(c.token) == 0 {
				return Error("physics value not specified")
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
				return Error("Invalid value: " + c.token)
			}
			out.append(OC_ex_, OC_ex_physics, OpCode(st))
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "playerno":
		out.append(OC_ex_, OC_ex_playerno)
	case "ratiolevel":
		out.append(OC_ex_, OC_ex_ratiolevel)
	case "receiveddamage":
		out.append(OC_ex_, OC_ex_receiveddamage)
	case "receivedhits":
		out.append(OC_ex_, OC_ex_receivedhits)
	case "redlife":
		out.append(OC_ex_, OC_ex_redlife)
	case "roundtype":
		out.append(OC_ex_, OC_ex_roundtype)
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
	case "stagebackedgedist", "stagebackedge": //Latter is deprecated
		out.append(OC_ex_, OC_ex_stagebackedgedist)
	case "stageconst":
		if err := c.checkOpeningBracket(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_const_)
		out.appendI32Op(OC_const_stage_constants, int32(sys.stringPool[c.playerNo].Add(
			strings.ToLower(c.token))))
		*in = strings.TrimSpace(*in)
		if len(*in) == 0 || (!sys.ignoreMostErrors && (*in)[0] != ')') {
			return bvNone(), Error("Missing ')' before " + c.token)
		}
		*in = (*in)[1:]
	case "stagefrontedgedist", "stagefrontedge": //Latter is deprecated
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
	case "timeremaining", "timeleft": // timeleft is deprecated, kept for backward compatibility
		out.append(OC_ex_, OC_ex_timeremaining)
	case "timetotal":
		out.append(OC_ex_, OC_ex_timetotal)
	case "drawpalno":
		out.append(OC_ex_, OC_ex_drawpalno)
	case "angle":
		out.append(OC_ex_, OC_ex_angle)
	case "scale":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_scale_x)
		case "y":
			out.append(OC_ex_, OC_ex_scale_y)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "offset":
		c.token = c.tokenizer(in)
		switch c.token {
		case "x":
			out.append(OC_ex_, OC_ex_offset_x)
		case "y":
			out.append(OC_ex_, OC_ex_offset_y)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "alpha":
		c.token = c.tokenizer(in)
		switch c.token {
		case "source":
			out.append(OC_ex_, OC_ex_alpha_s)
		case "dest":
			out.append(OC_ex_, OC_ex_alpha_d)
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
	case "=", "!=", ">", ">=", "<", "<=", "&", "&&", "^", "^^", "|", "||",
		"+", "*", "**", "/", "%":
		if !sys.ignoreMostErrors || len(c.previousOperator) > 0 {
			return bvNone(), Error("Invalid data: " + c.token)
		}
		if rd {
			out.append(OC_rdreset)
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
func (c *Compiler) renzokuEnzansihaError(in *string) error {
	*in = strings.TrimSpace(*in)
	if len(*in) > 0 {
		switch (*in)[0] {
		default:
			if len(*in) < 2 || (*in)[:2] != "!=" {
				break
			}
			fallthrough
		case '=', '<', '>', '|', '&', '+', '*', '/', '%', '^':
			return Error("Invalid data: " + c.tokenizer(in))
		}
	}
	return nil
}
func (c *Compiler) expPostNot(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
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
				if err := c.renzokuEnzansihaError(in); err != nil {
					return bvNone(), err
				}
				oldin = oldin[:len(oldin)-len(*in)]
				*in = oldtoken + " " + oldin[:strings.LastIndex(oldin, c.token)] +
					" " + *in
			}
		} else if opp > 0 {
			if err := c.renzokuEnzansihaError(in); err != nil {
				return bvNone(), err
			}
		}
	}
	return bv, nil
}
func (c *Compiler) expPow(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
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
func (c *Compiler) expMldv(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
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
func (c *Compiler) expAdsb(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
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
func (c *Compiler) expGrls(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
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
		if err := c.checkClosingBracket(); err != nil {
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
func (c *Compiler) expEqne(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
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
func (c *Compiler) expAnd(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	return c.expOneOp(out, in, c.expEqne, "&", out.and, OC_and)
}
func (c *Compiler) expXor(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	return c.expOneOp(out, in, c.expAnd, "^", out.xor, OC_xor)
}
func (c *Compiler) expOr(out *BytecodeExp, in *string) (BytecodeValue, error) {
	return c.expOneOp(out, in, c.expXor, "|", out.or, OC_or)
}
func (c *Compiler) expBoolAnd(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
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
func (c *Compiler) expBoolXor(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	return c.expOneOp(out, in, c.expBoolAnd, "^^", out.blxor, OC_blxor)
}
func (c *Compiler) expBoolOr(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
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
func (c *Compiler) argExpression(in *string, vt ValueType) (BytecodeExp,
	error) {
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
func (c *Compiler) fullExpression(in *string, vt ValueType) (BytecodeExp,
	error) {
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
func (c *Compiler) stateParam(is IniSection, name string,
	f func(string) error) error {
	data, ok := is[name]
	if ok {
		if err := f(data); err != nil {
			return Error(data + "\n" + name + ": " + err.Error())
		}
		delete(is, name)
	}
	return nil
}

// Returns FX prefix from a data string, removes prefix from the data
func (c *Compiler) getDataPrefix(data *string, ffxDefault bool) (prefix string) {
	if len(*data) > 1 {
		// Check prefix
		re := regexp.MustCompile(sys.ffxRegexp)
		prefix = re.FindString(strings.ToLower(*data))
		if prefix != "" {
			// Remove prefix from data string
			re = regexp.MustCompile("[^a-z]")
			m := re.Split(strings.ToLower(*data)[len(prefix):], -1)
			if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
				*data = (*data)[len(prefix):]
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
func (c *Compiler) paramValue(is IniSection, sc *StateControllerBase,
	paramname string, id byte, vt ValueType, numArg int, mandatory bool) error {
	f := false
	if err := c.stateParam(is, paramname, func(data string) error {
		f = true
		return c.scAdd(sc, id, data, vt, numArg)
	}); err != nil {
		return err
	}
	if mandatory && !f {
		return Error(paramname + " not specified")
	}
	return nil
}
func (c *Compiler) paramPostype(is IniSection, sc *StateControllerBase,
	id byte) error {
	return c.stateParam(is, "postype", func(data string) error {
		if len(data) == 0 {
			return Error("Value not specified")
		}
		var pt PosType
		if len(data) >= 2 && strings.ToLower(data[:2]) == "p2" {
			pt = PT_P2
		} else {
			switch strings.ToLower(data)[0] {
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
				return Error("Invalid value: " + data)
			}
		}
		sc.add(id, sc.iToExp(int32(pt)))
		return nil
	})
}

func (c *Compiler) paramSpace(is IniSection, sc *StateControllerBase,
	id byte) error {
	return c.stateParam(is, "space", func(data string) error {
		if len(data) <= 1 {
			return Error("Value not specified")
		}
		var sp Space
		if len(data) >= 2 {
			if strings.ToLower(data[:2]) == "st" {
				sp = Space_stage
			} else if strings.ToLower(data[:2]) == "sc" {
				sp = Space_screen
			}
		}
		sc.add(id, sc.iToExp(int32(sp)))
		return nil
	})
}

func (c *Compiler) paramProjection(is IniSection, sc *StateControllerBase,
	id byte) error {
	return c.stateParam(is, "projection", func(data string) error {
		if len(data) <= 1 {
			return Error("Value not specified")
		}
		var proj Projection
		if len(data) >= 2 {
			if strings.ToLower(data[:2]) == "or" {
				proj = Projection_Orthographic
			} else if strings.ToLower(data[:2]) == "pe" {
				if data[len(data)-1] != '2' {
					proj = Projection_Perspective
				} else {
					proj = Projection_Perspective2
				}

			}
		}
		sc.add(id, sc.iToExp(int32(proj)))
		return nil
	})
}

func (c *Compiler) paramSaveData(is IniSection, sc *StateControllerBase,
	id byte) error {
	return c.stateParam(is, "savedata", func(data string) error {
		if len(data) <= 1 {
			return Error("Value not specified")
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
			return Error("Invalid value: " + data)
		}
		sc.add(id, sc.iToExp(int32(sv)))
		return nil
	})
}

func (c *Compiler) paramTrans(is IniSection, sc *StateControllerBase,
	prefix string, id byte, afterImage bool) error {
	return c.stateParam(is, prefix+"trans", func(data string) error {
		if len(data) == 0 {
			return Error("Value not specified")
		}
		tt := TT_default
		data = strings.ToLower(data)
		switch data {
		case "none":
			tt = TT_none
		case "add1":
			tt = TT_add1
		case "sub":
			tt = TT_sub
		default:
			_error := false
			if afterImage {
				if len(data) >= 3 && data[:3] == "add" {
					tt = TT_add
				} else {
					_error = true
				}
			} else {
				switch data {
				case "default":
					tt = TT_default
				case "add":
					tt = TT_add
				case "addalpha", "alpha":
					tt = TT_alpha
				default:
					_error = true
				}
			}
			if _error && (!afterImage || !sys.ignoreMostErrors) {
				return Error("Invalid value: " + data)
			}
		}
		var exp []BytecodeExp
		b := false
		if !afterImage || sys.cgi[c.playerNo].mugenver[0] == 1 {
			if err := c.stateParam(is, prefix+"alpha", func(data string) error {
				b = true
				bes, err := c.exprs(data, VT_Int, 2)
				if err != nil {
					return err
				}
				// TODO: Based on my tests add1 doesn't need special alpha[1] handling
				// Remove unused code if there won't be regression.
				//if tt == TT_add1 {
				//	exp = make([]BytecodeExp, 4) // 長さ4にする
				//} else if tt == TT_add || tt == TT_alpha {
				if tt == TT_add || tt == TT_alpha || tt == TT_add1 {
					exp = make([]BytecodeExp, 3) // 長さ3にする
				} else {
					exp = make([]BytecodeExp, 2)
				}
				exp[0] = bes[0]
				if len(exp) == 2 {
					exp[0].append(OC_pop)
					switch tt {
					case TT_none:
						exp[0].appendValue(BytecodeInt(255))
					case TT_sub:
						exp[0].appendValue(BytecodeInt(1))
					default:
						exp[0].appendValue(BytecodeInt(-1))
					}
				}
				if len(bes) > 1 {
					exp[1] = bes[1]
					if tt != TT_alpha && tt != TT_add1 && !(tt == TT_add && sys.cgi[c.playerNo].mugenver[0] == 1) {
						exp[1].append(OC_pop)
					}
				}
				switch tt {
				case TT_alpha, TT_add1:
					if len(bes) <= 1 {
						exp[1].appendValue(BytecodeInt(255))
					}
				case TT_add:
					if sys.cgi[c.playerNo].mugenver[0] == 1 {
						if len(bes) <= 1 {
							exp[1].appendValue(BytecodeInt(255))
						}
					} else {
						exp[1].appendValue(BytecodeInt(255))
					}
				case TT_sub:
					exp[1].appendValue(BytecodeInt(255))
				default:
					exp[1].appendValue(BytecodeInt(0))
				}
				return nil
			}); err != nil {
				return err
			}
		}
		if !b {
			switch tt {
			case TT_none:
				exp = sc.iToExp(255, 0)
			case TT_add:
				exp = sc.iToExp(255, 255)
			case TT_add1:
				exp = sc.iToExp(255, ^255)
			case TT_sub:
				exp = sc.iToExp(1, 255)
			default:
				exp = sc.iToExp(-1, 0)
			}
		}
		sc.add(id, exp)
		return nil
	})
}

// Interprets an IniSection of statedef properties and sets them to a StateBytecode
func (c *Compiler) stateDef(is IniSection, sbc *StateBytecode) error {
	return c.stateSec(is, func() error {
		sc := newStateControllerBase()
		if err := c.stateParam(is, "type", func(data string) error {
			if len(data) == 0 {
				return Error("Value not specified")
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
				return Error("Invalid value: " + data)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "movetype", func(data string) error {
			if len(data) == 0 {
				return Error("Value not specified")
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
				return Error("Invalid value: " + data)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "physics", func(data string) error {
			if len(data) == 0 {
				return Error("Value not specified")
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
				return Error("Invalid value: " + data)
			}
			return nil
		}); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "hitcountpersist", func(data string) error {
			b = true
			return c.scAdd(sc, stateDef_hitcountpersist, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_hitcountpersist, sc.iToExp(0))
		}
		b = false
		if err := c.stateParam(is, "movehitpersist", func(data string) error {
			b = true
			return c.scAdd(sc, stateDef_movehitpersist, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_movehitpersist, sc.iToExp(0))
		}
		b = false
		if err := c.stateParam(is, "hitdefpersist", func(data string) error {
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
		if err := c.stateParam(is, "juggle", func(data string) error {
			return c.scAdd(sc, stateDef_juggle, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "velset",
			stateDef_velset, VT_Float, 3, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", func(data string) error {
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

		// Do the string was closed?
		if inString != 2 {
			if inString%2 != 0 {
				return nil, Error("String not closed.")
			} else if inString > 2 { // Do we have more than 1 string without using ','?
				return nil, Error("Lack of ',' separator.")
			} else {
				return nil, Error("Unknown string array error.")
			}
		} else if formatError {
			return nil, Error("Wrong format on string array.")
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
	zss := HasExtension(filename, ".zss")
	fnz := filename
	// Load state file
	if err := LoadFile(&filename, dirs, func(filename string) error {
		var err error
		// If this is a zss file
		if zss {
			b, err := ioutil.ReadFile(filename)
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
			b, err := ioutil.ReadFile(filename)
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
		/* Find a statedef, skipping over other lines until finding one */
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
			// Flag if this trigger can never be true
			allUtikiri := false
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
							allUtikiri = true
						}
					} else if !allUtikiri {
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
					} else if !allUtikiri && trexist[tn] >= 0 {
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
				return errmes(Error("type parameter not specified"))
			}
			if len(trexist) == 0 || (!allUtikiri && trexist[0] == 0) {
				return errmes(Error("Missing trigger1"))
			}

			/* Create trigger bytecode */
			var texp BytecodeExp
			for _, e := range triggerall {
				texp.append(e...)
				texp.append(OC_jz8, 0)
				texp.append(OC_pop)
			}
			if allUtikiri {
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
				if !allUtikiri {
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
		return Error("Missing token")
	}
	return Error("Unexpected token: " + c.token)
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
		return "", Error("Not enclosed in \"")
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
			s, *line = (*line)[:i], ""
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
						return nil, Error("Duplicated name: " + name)
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
	if t == "+" && len(*line) == 2 && (*line)[0] == '1' {
		c.scan(line)
		return int32(-10), err
	}
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
				return Error("persistent cannot be used in a function")
			}
			if c.stateNo < 0 {
				return Error("persistent cannot be used in a negative state")
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
				return Error("persistent(1) is meaningless")
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
					return Error("For needs more than one expression")
				} else {
					// For only has begin/end expressions, so we stop compiling the header
					break
				}
			}
			if c.token == ";" && i < 1 {
				return Error("Misplaced ;")
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
			//helperはステコンとリダイレクトの両方で使う名称なのでチェックする
			if c.token == "helper" && ((*line)[0] == ',' || (*line)[0] == '(') {
				ok = false
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
						return errmes(Error("Duplicated name: " + r))
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

	/* Load initial data from definition file */
	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	lines, i, cmd, stcommon := SplitAndTrim(str, "\n"), 0, "", ""
	var st [11]string
	info, files := true, true
	for i < len(lines) {
		// Parse each ini section
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			// Read info section for the mugen/ikemen version of the character
			if info {
				info = false
				var ok bool
				var str string
				sys.cgi[pn].mugenver = [2]uint16{}
				if str, ok = is["mugenversion"]; ok {
					for i, s := range SplitAndTrim(str, ".") {
						if i >= len(sys.cgi[pn].mugenver) {
							break
						}
						if v, err := strconv.ParseUint(s, 10, 16); err == nil {
							sys.cgi[pn].mugenver[i] = uint16(v)
						} else {
							break
						}
					}
				}
				sys.cgi[pn].ikemenver = [3]uint16{}
				if str, ok = is["ikemenversion"]; ok {
					for i, s := range SplitAndTrim(str, ".") {
						if i >= len(sys.cgi[pn].ikemenver) {
							break
						}
						if v, err := strconv.ParseUint(s, 10, 16); err == nil {
							sys.cgi[pn].ikemenver[i] = uint16(v)
						} else {
							break
						}
					}
				}
				// Ikemen characters adopt Mugen 1.1 version as a safeguard
				if sys.cgi[pn].ikemenver[0] != 0 || sys.cgi[pn].ikemenver[1] != 0 {
					sys.cgi[pn].mugenver[0] = 1
					sys.cgi[pn].mugenver[1] = 1
				}
			}
		case "files":
			// Read files section to find the command and state filenames
			if files {
				files = false
				cmd, stcommon = is["cmd"], is["stcommon"]
				st[0] = is["st"]
				for i := 1; i < len(st); i++ {
					st[i] = is[fmt.Sprintf("st%v", i-1)]
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
	for _, s := range sys.commonCmd {
		if err := LoadFile(&s, []string{def, sys.motifDir, sys.lifebar.def, "", "data/"}, func(filename string) error {
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
	lines, i = SplitAndTrim(str, "\n"), 0

	// Initialize command list data
	if sys.chars[pn][0].cmd == nil {
		sys.chars[pn][0].cmd = make([]CommandList, MaxSimul*2+MaxAttachedChar)
		b := NewCommandBuffer()
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
			// Read controller remap
			if remap {
				remap = false
				rm := func(name string, k, nk *CommandKey) {
					switch strings.ToLower(is[name]) {
					case "x":
						*k, *nk = CK_x, CK_rx
					case "y":
						*k, *nk = CK_y, CK_ry
					case "z":
						*k, *nk = CK_z, CK_rz
					case "a":
						*k, *nk = CK_a, CK_ra
					case "b":
						*k, *nk = CK_b, CK_rb
					case "c":
						*k, *nk = CK_c, CK_rc
					case "s":
						*k, *nk = CK_s, CK_rs
					case "d":
						*k, *nk = CK_d, CK_rd
					case "w":
						*k, *nk = CK_w, CK_rw
					case "m":
						*k, *nk = CK_m, CK_rm
					}
				}
				rm("x", &ckr.x, &ckr.nx)
				rm("y", &ckr.y, &ckr.ny)
				rm("z", &ckr.z, &ckr.nz)
				rm("a", &ckr.a, &ckr.na)
				rm("b", &ckr.b, &ckr.nb)
				rm("c", &ckr.c, &ckr.nc)
				rm("s", &ckr.s, &ckr.ns)
				rm("d", &ckr.d, &ckr.nd)
				rm("w", &ckr.w, &ckr.nw)
				rm("m", &ckr.m, &ckr.nm)
			}
		case "defaults":
			// Read default command time and buffer time
			if defaults {
				defaults = false
				is.ReadI32("command.time", &c.cmdl.DefaultTime)
				var i32 int32
				if is.ReadI32("command.buffer.time", &i32) {
					c.cmdl.DefaultBufferTime = Max(1, i32)
				}
			}
		default:
			// Read input commands
			if len(name) >= 7 && name[:7] == "command" {
				cmds = append(cmds, is)
			}
		}
	}
	// Parse input commands
	for _, is := range cmds {
		name, _, err := is.getText("name")
		if err != nil {
			return nil, Error(fmt.Sprintf("%v:\nname: %v\n%v",
				cmd, name, err.Error()))
		}
		cm, err := ReadCommand(name, is["command"], ckr)
		if err != nil {
			return nil, Error(cmd + ":\nname = " + is["name"] +
				"\ncommand = " + is["command"] + "\n" + err.Error())
		}
		cm.time, cm.buftime = c.cmdl.DefaultTime, c.cmdl.DefaultBufferTime
		is.ReadI32("time", &cm.time)
		var i32 int32
		if is.ReadI32("buffer.time", &i32) {
			cm.buftime = Max(1, i32)
		}
		c.cmdl.Add(*cm)
	}

	/* Compile states */
	sys.stringPool[pn].Clear()
	sys.cgi[pn].wakewakaLength = 0
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
	for _, s := range sys.commonStates {
		if err := c.stateCompile(states, s, []string{def, sys.motifDir, sys.lifebar.def, "", "data/"},
			false, constants); err != nil {
			return nil, err
		}
	}
	return states, nil
}
