package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const kuuhaktokigou = " !=<>()|&+-*/%,[]^:;{}#\"\t\r\n"

type expFunc func(out *BytecodeExp, in *string) (BytecodeValue, error)
type scFunc func(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error)
type Compiler struct {
	cmdl     *CommandList
	maeOp    string
	usiroOp  bool
	norange  bool
	token    string
	playerNo int
	scmap    map[string]scFunc
	block    *StateBlock
	lines    []string
	i        int
	linechan chan *string
	vars     map[string]uint8
	funcs    map[string]bytecodeFunction
	funcUsed map[string]bool
	stateNo  int32
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
		"forcefeedback":        c.null,
		"null":                 c.null,
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
		"printtoconsole":       c.printToConsole,
		"rankadd":              c.rankAdd,
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
	}
	return c
}

var triggerMap = map[string]int{
	//redirections
	"player":    0,
	"parent":    0,
	"root":      0,
	"helper":    0,
	"target":    0,
	"partner":   0,
	"enemy":     0,
	"enemynear": 0,
	"playerid":  0,
	//vanilla triggers
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
	//new triggers
	"animelemlength":   1,
	"animlength":       1,
	"combocount":       1,
	"consecutivewins":  1,
	"dizzy":            1,
	"dizzypoints":      1,
	"dizzypointsmax":   1,
	"firstattack":      1,
	"gamemode":         1,
	"getplayerid":      1,
	"guardbreak":       1,
	"guardpoints":      1,
	"guardpointsmax":   1,
	"hitoverridden":    1,
	"incustomstate":    1,
	"indialogue":       1,
	"isasserted":       1,
	"localscale":       1,
	"majorversion":     1,
	"map":              1,
	"memberno":         1,
	"movecountered":    1,
	"p5name":           1,
	"p6name":           1,
	"p7name":           1,
	"p8name":           1,
	"pausetime":        1,
	"physics":          1,
	"playerno":         1,
	"rank":             1,
	"ratiolevel":       1,
	"receivedhits":     1,
	"receiveddamage":   1,
	"redlife":          1,
	"roundtype":        1,
	"score":            1,
	"scoretotal":       1,
	"selfstatenoexist": 1,
	"sprpriority":      1,
	"stagebackedge":    1,
	"stageconst":       1,
	"stagefrontedge":   1,
	"stagetime":        1,
	"standby":          1,
	"teamleader":       1,
	"teamsize":         1,
	"timeelapsed":      1,
	"timeremaining":    1,
	"timetotal":        1,
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
		i = strings.IndexAny(*in, kuuhaktokigou)
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
	if len(c.maeOp) > 0 {
		if opp := c.isOperator(c.token); opp <= c.isOperator(c.maeOp) {
			if opp < 0 || ((!c.usiroOp || c.token[0] != '(') &&
				(c.token[0] < 'A' || c.token[0] > 'Z') &&
				(c.token[0] < 'a' || c.token[0] > 'z')) {
				return Error("Invalid data: " + c.maeOp)
			}
			*in = c.token + " " + *in
			c.token = c.maeOp
			c.maeOp = ""
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
		c.usiroOp = false
		return BytecodeValue{VT_Float, f}
	}
	if strings.ContainsAny(token, "Ee") {
		return bvNone()
	}
	c.usiroOp = false
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
	hitdefflg := flg
	for i, a := range att[1:] {
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
			if sys.ignoreMostErrors && sys.cgi[c.playerNo].ver[0] == 1 {
				if hitdef {
					flg = hitdefflg
				}
				return flg, nil
			}
			return 0, Error("Invalid value: " + a)
		}
		if i == 0 {
			hitdefflg = flg
		}
		if l > 2 {
			break
		}
	}
	if hitdef {
		flg = hitdefflg
	}
	return flg, nil
}
func (c *Compiler) trgAttr(in *string) (int32, error) {
	flg := int32(0)
	*in = c.token + *in
	i := strings.IndexAny(*in, kuuhaktokigou)
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
		i := strings.IndexAny(*in, kuuhaktokigou)
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
func (c *Compiler) kakkohiraku(in *string) error {
	if c.tokenizer(in) != "(" {
		return Error("Missing '(' after " + c.token)
	}
	c.token = c.tokenizer(in)
	return nil
}

/* TODO: Case sensitive maps
func (c *Compiler) kakkohirakuCS(in *string) error {
	if c.tokenizerCS(in) != "(" {
		return Error("Missing '(' after " + c.token)
	}
	c.token = c.tokenizerCS(in)
	return nil
}*/
func (c *Compiler) kakkotojiru() error {
	c.usiroOp = true
	if c.token != ")" {
		return Error("Missing ')' before " + c.token)
	}
	return nil
}
func (c *Compiler) kyuushiki(in *string) (not bool, err error) {
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
func (c *Compiler) kyuushikiThroughNeo(_range bool, in *string) {
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
func (c *Compiler) kyuushikiSuperDX(out *BytecodeExp, in *string,
	hissu bool) error {
	comma := c.token == ","
	if comma {
		c.token = c.tokenizer(in)
	}
	var opc OpCode
	hikaku := true
	switch c.token {
	case "<":
		opc = OC_lt
		c.kyuushikiThroughNeo(false, in)
	case ">":
		opc = OC_gt
		c.kyuushikiThroughNeo(false, in)
	case "<=":
		opc = OC_le
		c.kyuushikiThroughNeo(false, in)
	case ">=":
		opc = OC_ge
		c.kyuushikiThroughNeo(false, in)
	default:
		opc = OC_eq
		switch c.token {
		case "!=":
			opc = OC_ne
		case "=":
		default:
			if hissu && !comma {
				return Error("No comparison operator" +
					"\n[ECID 1]\n")
			}
			hikaku = false
		}
		if hikaku {
			c.kyuushikiThroughNeo(true, in)
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
			c.usiroOp = comma || hikaku
			return nil
		}
	}
	ot, oi := c.token, *in
	n, err := c.integer2(in)
	if err != nil {
		if hissu && !hikaku {
			return Error("No comparison operator" +
				"\n[ECID 2]\n")
		}
		if hikaku {
			return err
		}
		n, c.token, *in = 0, ot, oi
	}
	out.appendValue(BytecodeInt(n))
	out.append(opc)
	c.usiroOp = true
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
		if err := c.kakkotojiru(); err != nil {
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
	c.usiroOp, c.norange = true, false
	bv := c.number(c.token)
	if !bv.IsNone() {
		c.token = c.tokenizer(in)
		return bv, nil
	}
	_var := func(sys, f bool) error {
		bv1, err := c.oneArg(out, in, rd, false)
		if err != nil {
			return err
		}
		var oc OpCode
		c.token = c.tokenizer(in)
		set, _else := c.token == ":=", false
		if !bv1.IsNone() && bv1.ToI() >= 0 {
			switch [...]bool{sys, f} {
			case [...]bool{false, false}:
				if bv1.ToI() < int32(NumVar) {
					oc = OC_var0 + OpCode(bv1.ToI()) // OC_st_var0と同じ値
				} else {
					_else = true
				}
			case [...]bool{false, true}:
				if bv1.ToI() < int32(NumFvar) {
					oc = OC_fvar0 + OpCode(bv1.ToI()) // OC_st_fvar0と同じ値
				} else {
					_else = true
				}
			case [...]bool{true, false}:
				if bv1.ToI() < int32(NumSysVar) {
					oc = OC_sysvar0 + OpCode(bv1.ToI()) // OC_st_sysvar0と同じ値
				} else {
					_else = true
				}
			case [...]bool{true, true}:
				if bv1.ToI() < int32(NumSysFvar) {
					oc = OC_sysfvar0 + OpCode(bv1.ToI()) // OC_st_sysfvar0と同じ値
				} else {
					_else = true
				}
			}
		} else {
			_else = true
		}
		if _else {
			out.appendValue(bv1)
		}
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
		if _else {
			switch [...]bool{sys, f} {
			case [...]bool{false, false}:
				oc = OC_var
			case [...]bool{false, true}:
				oc = OC_fvar
			case [...]bool{true, false}:
				oc = OC_sysvar
			case [...]bool{true, true}:
				oc = OC_sysfvar
			}
			if set {
				oc += OC_st_var - OC_var
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
		not, err := c.kyuushiki(in)
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
		not, err := c.kyuushiki(in)
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
	case "root", "parent", "helper", "target", "partner",
		"enemy", "enemynear", "playerid":
		switch c.token {
		case "parent":
			opc = OC_parent
			c.token = c.tokenizer(in)
		case "root":
			opc = OC_root
			c.token = c.tokenizer(in)
		default:
			switch c.token {
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
			}
			c.token = c.tokenizer(in)
			if c.token == "(" {
				c.token = c.tokenizer(in)
				if bv1, err = c.expBoolOr(&be1, in); err != nil {
					return bvNone(), err
				}
				if err := c.kakkotojiru(); err != nil {
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
				case OC_playerid:
					return bvNone(), Error("Missing '(' after playerid")
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
		if err := c.kakkotojiru(); err != nil {
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
		if err := c.kakkohiraku(in); err != nil {
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
		if err := c.kakkotojiru(); err != nil {
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
			i, ok := c.cmdl.Names[c.token]
			if !ok {
				return Error("Command doesn't exist: " + c.token)
			}
			out.appendI32Op(OC_command, int32(i))
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "const":
		if err := c.kakkohiraku(in); err != nil {
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
		case "size.z.width":
			out.append(OC_const_size_z_width)
		case "size.height":
			out.append(OC_const_size_height)
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
		if err := c.kakkohiraku(in); err != nil {
			return bvNone(), err
		}
		switch c.token {
		case "xveladd":
			bv.SetF(0)
		case "yveladd":
			bv.SetF(0)
		case "type":
			bv.SetI(0)
		case "zoff":
			bv.SetF(0)
		case "fall.envshake.dir":
			bv.SetI(0)
		default:
			out.append(OC_ex_)
			switch c.token {
			case "animtype":
				out.append(OC_ex_gethitvar_animtype)
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
			default:
				return bvNone(), Error("Invalid data: " + c.token)
			}
		}
		c.token = c.tokenizer(in)
		if err := c.kakkotojiru(); err != nil {
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
		if sys.cgi[c.playerNo].ver[0] == 1 {
			if err := eqne(hda); err != nil {
				return bvNone(), err
			}
		} else {
			if not, err := c.kyuushiki(in); err != nil {
				if sys.ignoreMostErrors {
					out.appendValue(BytecodeBool(false))
				} else {
					return bvNone(), err
				}
			} else if err := hda(); err != nil {
				return bvNone(), err
			} else if not && !sys.ignoreMostErrors {
				return bvNone(), Error("hitdefattr doesn't support '!=' in this mugenversion")
			}
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
	case "movetype", "p2movetype":
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
			if trname == "p2movetype" {
				out.appendI32Op(OC_p2, 2+Btoi(not))
			}
			out.append(OC_movetype, OpCode(mt>>15))
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
	case "rightedge":
		out.append(OC_rightedge)
	case "roundno":
		out.append(OC_ex_, OC_ex_roundno)
	case "roundsexisted":
		out.append(OC_roundsexisted)
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
	case "statetype", "p2statetype":
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
			if trname == "p2statetype" {
				out.appendI32Op(OC_p2, 2+Btoi(not))
			}
			out.append(OC_statetype, OpCode(st))
			if not {
				out.append(OC_blnot)
			}
			return nil
		}); err != nil {
			return bvNone(), err
		}
	case "stagevar":
		if err := c.kakkohiraku(in); err != nil {
			return bvNone(), err
		}
		svname := c.token
		c.token = c.tokenizer(in)
		if err := c.kakkotojiru(); err != nil {
			return bvNone(), err
		}
		var opc OpCode
		switch svname {
		case "info.name":
			opc = OC_const_stagevar_info_name
		case "info.displayname":
			opc = OC_const_stagevar_info_displayname
		case "info.author":
			opc = OC_const_stagevar_info_author
		default:
			return bvNone(), Error("Invalid data: " + svname)
		}
		if err := nameSub(opc); err != nil {
			return bvNone(), err
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
		if not, err := c.kyuushiki(in); err != nil {
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
		if err = c.kyuushikiSuperDX(&be, in, false); err != nil {
			return bvNone(), err
		}
		out.append(OC_jsf8, OpCode(len(be)))
		out.append(be...)
		return bv, nil
	case "timemod":
		if not, err := c.kyuushiki(in); err != nil {
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
		if err = c.kyuushikiSuperDX(out, in, true); err != nil {
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
			out.append(OC_ex_, OC_ex_p2dist_y)
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
		if err := c.kakkohiraku(in); err != nil {
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
		if err := c.kakkotojiru(); err != nil {
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
		if err := c.kakkohiraku(in); err != nil {
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
		if err := c.kakkotojiru(); err != nil {
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
		if err := c.kakkohiraku(in); err != nil {
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
		if err := c.kakkotojiru(); err != nil {
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
	case "rand":
		if err := c.kakkohiraku(in); err != nil {
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
		if err := c.kakkotojiru(); err != nil {
			return bvNone(), err
		}
		if rd {
			out.append(OC_rdreset)
		}
		out.append(be1...)
		out.appendValue(bv1)
		out.append(be2...)
		out.appendValue(bv2)
		out.append(OC_ex_, OC_ex_rand)
	case "round":
		if err := c.kakkohiraku(in); err != nil {
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
		if err := c.kakkotojiru(); err != nil {
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
	case "ailevelf":
		out.append(OC_ex_, OC_ex_ailevelf)
	case "animelemlength":
		out.append(OC_ex_, OC_ex_animelemlength)
	case "animlength":
		out.append(OC_ex_, OC_ex_animlength)
	case "combocount":
		out.append(OC_ex_, OC_ex_combocount)
	case "consecutivewins":
		out.append(OC_ex_, OC_ex_consecutivewins)
	case "dizzy":
		out.append(OC_ex_, OC_ex_dizzy)
	case "dizzypoints":
		out.append(OC_ex_, OC_ex_dizzypoints)
	case "dizzypointsmax":
		out.append(OC_ex_, OC_ex_dizzypointsmax)
	case "firstattack":
		out.append(OC_ex_, OC_ex_firstattack)
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
		if err := c.kakkohiraku(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_)
		switch c.token {
		case "nostandguard":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nostandguard))
		case "nocrouchguard":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nocrouchguard))
		case "noairguard":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_noairguard))
		case "noshadow":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_noshadow))
		case "invisible":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_invisible))
		case "unguardable":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_unguardable))
		case "nojugglecheck":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nojugglecheck))
		case "noautoturn":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_noautoturn))
		case "nowalk":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nowalk))
		case "nobrake":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nobrake))
		case "nocrouch":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nocrouch))
		case "nostand":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nostand))
		case "nojump":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nojump))
		case "noairjump":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_noairjump))
		case "nohardcodedkeys":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nohardcodedkeys))
		case "nogetupfromliedown":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nogetupfromliedown))
		case "nofastrecoverfromliedown":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nofastrecoverfromliedown))
		case "nofallcount":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nofallcount))
		case "nofalldefenceup":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nofalldefenceup))
		case "noturntarget":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_noturntarget))
		case "noinput":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_noinput))
		case "nopowerbardisplay":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_nopowerbardisplay))
		case "autoguard":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_autoguard))
		case "animfreeze":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_animfreeze))
		case "postroundinput":
			out.appendI32Op(OC_ex_isassertedchar, int32(CSF_postroundinput))
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
		case "noko":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_noko))
		case "nokovelocity":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_nokovelocity))
		case "roundnotskip":
			out.appendI32Op(OC_ex_isassertedglobal, int32(GSF_roundnotskip))
		default:
			return bvNone(), Error("Invalid data: " + c.token)
		}
		c.token = c.tokenizer(in)
		if err := c.kakkotojiru(); err != nil {
			return bvNone(), err
		}
	case "ishost":
		out.append(OC_ex_, OC_ex_ishost)
	case "localscale":
		out.append(OC_ex_, OC_ex_localscale)
	case "majorversion":
		out.append(OC_ex_, OC_ex_majorversion)
	case "map":
		if err := c.kakkohiraku(in); err != nil {
			return bvNone(), err
		}
		out.append(OC_ex_)
		out.appendI32Op(OC_ex_maparray, int32(sys.stringPool[c.playerNo].Add(strings.ToLower(c.token))))
		c.token = c.tokenizer(in)
		if err := c.kakkotojiru(); err != nil {
			return bvNone(), err
		}
	case "memberno":
		out.append(OC_ex_, OC_ex_memberno)
	case "movecountered":
		out.append(OC_ex_, OC_ex_movecountered)
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
	case "rank":
		out.append(OC_ex_, OC_ex_rank)
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
	case "stagebackedge":
		out.append(OC_ex_, OC_ex_stagebackedge)
	case "stageconst":
		if err := c.kakkohiraku(in); err != nil {
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
	case "stagefrontedge":
		out.append(OC_ex_, OC_ex_stagefrontedge)
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
	case "timeleft":
		out.append(OC_ex_, OC_ex_timeremaining) // Only here for backwards compatibility purposes, going to be deprecated once Add004 updates.
	case "timetotal":
		out.append(OC_ex_, OC_ex_timetotal)
	case "drawpalno":
		out.append(OC_ex_, OC_ex_drawpalno)
	case "=", "!=", ">", ">=", "<", "<=", "&", "&&", "^", "^^", "|", "||",
		"+", "*", "**", "/", "%":
		if !sys.ignoreMostErrors || len(c.maeOp) > 0 {
			return bvNone(), Error("Invalid data: " + c.token)
		}
		if rd {
			out.append(OC_rdreset)
		}
		c.maeOp = c.token
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
			if not, err := c.kyuushiki(in); err != nil {
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
			if err = c.kyuushikiSuperDX(&be, in, false); err != nil {
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
			c.usiroOp = true
			if bv.IsNone() {
				out.append(OC_blnot)
			} else {
				out.blnot(&bv)
			}
			c.token = c.tokenizer(in)
		}
	}
	if len(c.maeOp) == 0 {
		if opp := c.isOperator(c.token); opp == 0 {
			if !sys.ignoreMostErrors || !c.usiroOp && c.token == "(" {
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
			if c.usiroOp {
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
		if err := c.kakkotojiru(); err != nil {
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
	defer func(omp string) { c.maeOp = omp }(c.maeOp)
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
func (c *Compiler) paramPostye(is IniSection, sc *StateControllerBase,
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
				pt = PT_F
			case 'b':
				pt = PT_B
			case 'l':
				pt = PT_L
			case 'r':
				pt = PT_R
			case 'n':
				pt = PT_N
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

func (c *Compiler) paramSaveData(is IniSection, sc *StateControllerBase,
	id byte) error {
	return c.stateParam(is, "savedata", func(data string) error {
		if len(data) <= 1 {
			return Error("Value not specified")
		}
		var sv SaveData
		if len(data) >= 2 {
			if strings.ToLower(data[:2]) == "ma" {
				sv = SaveData_map
			} else if strings.ToLower(data[:2]) == "va" {
				sv = SaveData_var
			} else if strings.ToLower(data[:2]) == "fv" {
				sv = SaveData_fvar
			}
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
		if !afterImage || sys.cgi[c.playerNo].ver[0] == 1 {
			if err := c.stateParam(is, prefix+"alpha", func(data string) error {
				b = true
				bes, err := c.exprs(data, VT_Int, 2)
				if err != nil {
					return err
				}
				if tt == TT_add1 {
					exp = make([]BytecodeExp, 4) // 長さ4にする
				} else if tt == TT_add || tt == TT_alpha {
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
					if tt != TT_alpha && tt != TT_add1 && !(tt == TT_add && sys.cgi[c.playerNo].ver[0] == 1) {
						exp[1].append(OC_pop)
					}
				}
				switch tt {
				case TT_alpha, TT_add1:
					if len(bes) <= 1 {
						exp[1].appendValue(BytecodeInt(255))
					}
				case TT_add:
					if sys.cgi[c.playerNo].ver[0] == 1 {
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
		b = false
		if err := c.stateParam(is, "juggle", func(data string) error {
			b = true
			return c.scAdd(sc, stateDef_juggle, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_juggle, sc.iToExp(0))
		}
		if err := c.paramValue(is, sc, "velset",
			stateDef_velset, VT_Float, 3, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", func(data string) error {
			fflg := false
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = true
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, stateDef_anim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
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

func (c *Compiler) hitBySub(is IniSection, sc *StateControllerBase) error {
	attr, two := int32(-1), false
	var err error
	if err = c.stateParam(is, "value", func(data string) error {
		attr, err = c.attr(data, false)
		return err
	}); err != nil {
		return err
	}
	if attr == -1 {
		if err = c.stateParam(is, "value2", func(data string) error {
			two = true
			attr, err = c.attr(data, false)
			return err
		}); err != nil {
			return err
		}
	}
	if attr == -1 {
		return Error("value parameter not specified")
	}
	if err := c.paramValue(is, sc, "time",
		hitBy_time, VT_Int, 1, false); err != nil {
		return err
	}
	if two {
		sc.add(hitBy_value2, sc.iToExp(attr))
	} else {
		sc.add(hitBy_value, sc.iToExp(attr))
	}
	return nil
}
func (c *Compiler) hitBy(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*hitBy)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitBy_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.hitBySub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) notHitBy(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*notHitBy)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitBy_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.hitBySub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) assertSpecial(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*assertSpecial)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			assertSpecial_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		foo := func(data string) error {
			switch strings.ToLower(data) {
			case "nostandguard":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nostandguard)))
			case "nocrouchguard":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nocrouchguard)))
			case "noairguard":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_noairguard)))
			case "noshadow":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_noshadow)))
			case "invisible":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_invisible)))
			case "unguardable":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_unguardable)))
			case "nojugglecheck":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nojugglecheck)))
			case "noautoturn":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_noautoturn)))
			case "nowalk":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nowalk)))
			case "nobrake":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nobrake)))
			case "nocrouch":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nocrouch)))
			case "nostand":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nostand)))
			case "nojump":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nojump)))
			case "noairjump":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_noairjump)))
			case "nohardcodedkeys":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nohardcodedkeys)))
			case "nogetupfromliedown":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nogetupfromliedown)))
			case "nofastrecoverfromliedown":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nofastrecoverfromliedown)))
			case "nofallcount":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nofallcount)))
			case "nofalldefenceup":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nofalldefenceup)))
			case "noturntarget":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_noturntarget)))
			case "noinput":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_noinput)))
			case "nopowerbardisplay":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_nopowerbardisplay)))
			case "autoguard":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_autoguard)))
			case "animfreeze":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_animfreeze)))
			case "postroundinput":
				sc.add(assertSpecial_flag, sc.iToExp(int32(CSF_postroundinput)))
			case "intro":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_intro)))
			case "roundnotover":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_roundnotover)))
			case "nomusic":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_nomusic)))
			case "nobardisplay":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_nobardisplay)))
			case "nobg":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_nobg)))
			case "nofg":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_nofg)))
			case "globalnoshadow":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_globalnoshadow)))
			case "timerfreeze":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_timerfreeze)))
			case "nokosnd":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_nokosnd)))
			case "nokoslow":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_nokoslow)))
			case "noko":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_noko)))
			case "nokovelocity":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_nokovelocity)))
			case "roundnotskip":
				sc.add(assertSpecial_flag_g, sc.iToExp(int32(GSF_roundnotskip)))
			default:
				return Error("Invalid value: " + data)
			}
			return nil
		}
		f := false
		if err := c.stateParam(is, "flag", func(data string) error {
			f = true
			return foo(data)
		}); err != nil {
			return err
		}
		if !f {
			return Error("flag parameter not specified")
		}
		if err := c.stateParam(is, "flag2", func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "flag3", func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) playSnd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*playSnd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			playSnd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		f := false
		if err := c.stateParam(is, "value", func(data string) error {
			f = true
			fflg := false
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' || strings.ToLower(data)[0] == 's' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = strings.ToLower(data)[0] == 'f'
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, playSnd_value, data, VT_Int, 2,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if !f {
			return Error("value parameter not specified")
		}
		if err := c.paramValue(is, sc, "channel",
			playSnd_channel, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "lowpriority",
			playSnd_lowpriority, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pan",
			playSnd_pan, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "abspan",
			playSnd_abspan, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "volume",
			playSnd_volume, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "volumescale",
			playSnd_volumescale, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "freqmul",
			playSnd_freqmul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loop",
			playSnd_loop, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) changeStateSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "redirectid",
		changeState_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "value",
		changeState_value, VT_Int, 1, true); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ctrl",
		changeState_ctrl, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "anim", func(data string) error {
		fflg := false
		if len(data) > 1 {
			if strings.ToLower(data)[0] == 'f' {
				re := regexp.MustCompile("[^a-z]")
				m := re.Split(strings.ToLower(data)[1:], -1)
				if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
					fflg = true
					data = data[1:]
				}
			}
		}
		return c.scAdd(sc, changeState_anim, data, VT_Int, 1,
			sc.iToExp(Btoi(fflg))...)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "readplayerid",
		changeState_readplayerid, VT_Int, 1, false); err != nil {
		return err
	}
	if c.block != nil && c.stateNo >= 0 && c.block.ignorehitpause == -1 {
		c.block.ignorehitpause = sys.cgi[c.playerNo].wakewakaLength
		sys.cgi[c.playerNo].wakewakaLength++
	}
	return nil
}
func (c *Compiler) changeState(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*changeState)(sc), c.stateSec(is, func() error {
		return c.changeStateSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) selfState(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*selfState)(sc), c.stateSec(is, func() error {
		return c.changeStateSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) tagIn(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*tagIn)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			tagIn_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "self", tagIn_self, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partner", tagIn_partner, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stateno", tagIn_stateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partnerstateno", tagIn_partnerstateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ctrl", tagIn_ctrl, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partnerctrl", tagIn_partnerctrl, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "leader", tagIn_leader, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	//if c.block != nil && c.block.ignorehitpause == -1 {
	//	c.block.ignorehitpause = sys.cgi[c.playerNo].wakewakaLength
	//	sys.cgi[c.playerNo].wakewakaLength++
	//}
	return *ret, err
}
func (c *Compiler) tagOut(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*tagOut)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			tagOut_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "self", tagOut_self, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partner", tagOut_partner, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stateno", tagOut_stateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partnerstateno", tagOut_partnerstateno, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	//if c.block != nil && c.block.ignorehitpause == -1 {
	//	c.block.ignorehitpause = sys.cgi[c.playerNo].wakewakaLength
	//	sys.cgi[c.playerNo].wakewakaLength++
	//}
	return *ret, err
}
func (c *Compiler) destroySelf(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*destroySelf)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			destroySelf_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "recursive",
			destroySelf_recursive, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "removeexplods",
			destroySelf_removeexplods, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) changeAnimSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "redirectid",
		changeAnim_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "elem",
		changeAnim_elem, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "value", func(data string) error {
		fflg := false
		if len(data) > 1 {
			if strings.ToLower(data)[0] == 'f' {
				re := regexp.MustCompile("[^a-z]")
				m := re.Split(strings.ToLower(data)[1:], -1)
				if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
					fflg = true
					data = data[1:]
				}
			}
		}
		return c.scAdd(sc, changeAnim_value, data, VT_Int, 1,
			sc.iToExp(Btoi(fflg))...)
	}); err != nil {
		return err
	}
	return nil
}
func (c *Compiler) changeAnim(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*changeAnim)(sc), c.stateSec(is, func() error {
		return c.changeAnimSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) changeAnim2(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*changeAnim2)(sc), c.stateSec(is, func() error {
		return c.changeAnimSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) helper(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*helper)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			helper_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "helpertype", func(data string) error {
			if len(data) == 0 {
				return Error("Value not specified")
			}
			switch strings.ToLower(data)[0] {
			case 'n':
			case 'p':
				sc.add(helper_helpertype, sc.iToExp(1))
			default:
				return Error("Invalid value: " + data)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "name", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(helper_name, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramPostye(is, sc, helper_postype); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ownpal",
			helper_ownpal, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.xscale",
			helper_size_xscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.yscale",
			helper_size_yscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.ground.back",
			helper_size_ground_back, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.ground.front",
			helper_size_ground_front, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.air.back",
			helper_size_air_back, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.air.front",
			helper_size_air_front, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.height",
			helper_size_height, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.proj.doscale",
			helper_size_proj_doscale, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.head.pos",
			helper_size_head_pos, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.mid.pos",
			helper_size_mid_pos, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.shadowoffset",
			helper_size_shadowoffset, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stateno",
			helper_stateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "keyctrl", func(data string) error {
			bes, err := c.exprs(data, VT_Int, 4)
			if err != nil {
				return err
			}
			sc.add(helper_keyctrl, bes)
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			helper_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			helper_pos, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "facing",
			helper_facing, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pausemovetime",
			helper_pausemovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "supermovetime",
			helper_supermovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "remappal",
			helper_remappal, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "extendsmap",
			helper_extendsmap, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "inheritjuggle",
			helper_inheritjuggle, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "immortal",
			helper_immortal, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "kovelocity",
			helper_kovelocity, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "preserve",
			helper_preserve, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) ctrlSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*ctrlSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			ctrlSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", ctrlSet_value, VT_Bool, 1, true)
	})
	return *ret, err
}
func (c *Compiler) explodSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "remappal",
		explod_remappal, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "id",
		explod_id, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "facing",
		explod_facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "vfacing",
		explod_vfacing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pos",
		explod_pos, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "random",
		explod_random, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramPostye(is, sc, explod_postype); err != nil {
		return err
	}
	if err := c.paramSpace(is, sc, explod_space); err != nil {
		return err
	}
	f := false
	if err := c.stateParam(is, "vel", func(data string) error {
		f = true
		return c.scAdd(sc, explod_velocity, data, VT_Float, 2)
	}); err != nil {
		return err
	}
	if !f {
		if err := c.paramValue(is, sc, "velocity",
			explod_velocity, VT_Float, 2, false); err != nil {
			return err
		}
	}
	if err := c.paramValue(is, sc, "accel",
		explod_accel, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "scale",
		explod_scale, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "bindtime",
		explod_bindtime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "removetime",
		explod_removetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "supermove",
		explod_supermove, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "supermovetime",
		explod_supermovetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pausemovetime",
		explod_pausemovetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "sprpriority",
		explod_sprpriority, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "bindid",
		explod_bindid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "ontop", func(data string) error {
		if err := c.scAdd(sc, explod_ontop, data, VT_Bool, 1); err != nil {
			return err
		}
		if c.block != nil {
			sc.add(explod_strictontop, nil)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "under", func(data string) error {
		if err := c.scAdd(sc, explod_under, data, VT_Bool, 1); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "shadow",
		explod_shadow, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "removeongethit",
		explod_removeongethit, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramTrans(is, sc, "", explod_trans, false); err != nil {
		return err
	}
	return nil
}
func (c *Compiler) explod(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error) {
	ret, err := (*explod)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			explod_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ownpal",
			explod_ownpal, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.explodSub(is, sc); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", func(data string) error {
			fflg := false
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = true
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, explod_anim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "angle",
			explod_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			explod_yangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			explod_xangle, VT_Float, 1, false); err != nil {
			return err
		}
		if ihp == 0 {
			sc.add(explod_ignorehitpause, sc.iToExp(0))
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) modifyExplod(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error) {
	ret, err := (*modifyExplod)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			explod_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ownpal",
			explod_ownpal, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.explodSub(is, sc); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", func(data string) error {
			fflg := false
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = true
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, explod_anim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "angle",
			explod_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			explod_yangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			explod_xangle, VT_Float, 1, false); err != nil {
			return err
		}
		if ihp == 0 {
			sc.add(explod_ignorehitpause, sc.iToExp(0))
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) gameMakeAnim(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*gameMakeAnim)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			gameMakeAnim_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			gameMakeAnim_pos, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "random",
			gameMakeAnim_random, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "under",
			gameMakeAnim_under, VT_Bool, 1, false); err != nil {
			return err
		}
		b := false
		anim := func(data string) error {
			b = true
			fflg := true
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 's' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = false
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, gameMakeAnim_anim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}
		if err := c.stateParam(is, "anim", func(data string) error {
			return anim(data)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.stateParam(is, "value", func(data string) error {
				return anim(data)
			}); err != nil {
				return err
			}
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) posSetSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "x",
		posSet_x, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "y",
		posSet_y, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "z",
		posSet_z, VT_Float, 1, false); err != nil {
		return err
	}
	return nil
}
func (c *Compiler) posSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*posSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) posAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*posAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) velSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*velSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) velAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*velAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) velMul(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*velMul)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) palFXSub(is IniSection,
	sc *StateControllerBase, prefix string) error {
	if err := c.paramValue(is, sc, prefix+"time",
		palFX_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"color",
		palFX_color, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"add", func(data string) error {
		bes, err := c.exprs(data, VT_Int, 3)
		if err != nil {
			return err
		}
		if len(bes) < 3 {
			return Error(prefix + "add - not enough arguments")
		}
		sc.add(palFX_add, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"mul", func(data string) error {
		bes, err := c.exprs(data, VT_Int, 3)
		if err != nil {
			return err
		}
		if len(bes) < 3 {
			return Error(prefix + "mul - not enough arguments")
		}
		sc.add(palFX_mul, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"sinadd", func(data string) error {
		bes, err := c.exprs(data, VT_Int, 4)
		if err != nil {
			return err
		}
		if len(bes) < 3 {
			return Error(prefix + "sinadd - not enough arguments")
		}
		sc.add(palFX_sinadd, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"invertall",
		palFX_invertall, VT_Bool, 1, false); err != nil {
		return err
	}
	return nil
}
func (c *Compiler) palFX(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*palFX)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			palFX_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.palFXSub(is, sc, "")
	})
	return *ret, err
}
func (c *Compiler) allPalFX(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*allPalFX)(sc), c.stateSec(is, func() error {
		return c.palFXSub(is, sc, "")
	})
	return *ret, err
}
func (c *Compiler) bgPalFX(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*bgPalFX)(sc), c.stateSec(is, func() error {
		return c.palFXSub(is, sc, "")
	})
	return *ret, err
}
func (c *Compiler) afterImageSub(is IniSection,
	sc *StateControllerBase, ihp int8, prefix string) error {
	if err := c.paramValue(is, sc, "redirectid",
		afterImage_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramTrans(is, sc, prefix,
		afterImage_trans, true); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"time",
		afterImage_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"length",
		afterImage_length, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"timegap",
		afterImage_timegap, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"framegap",
		afterImage_framegap, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palcolor",
		afterImage_palcolor, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palinvertall",
		afterImage_palinvertall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palbright",
		afterImage_palbright, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palcontrast",
		afterImage_palcontrast, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palpostbright",
		afterImage_palpostbright, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"paladd",
		afterImage_paladd, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palmul",
		afterImage_palmul, VT_Float, 3, false); err != nil {
		return err
	}
	if ihp == 0 {
		sc.add(afterImage_ignorehitpause, sc.iToExp(0))
	}
	return nil
}
func (c *Compiler) afterImage(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error) {
	ret, err := (*afterImage)(sc), c.stateSec(is, func() error {
		return c.afterImageSub(is, sc, ihp, "")
	})
	return *ret, err
}
func (c *Compiler) afterImageTime(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*afterImageTime)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			afterImageTime_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "time", func(data string) error {
			b = true
			return c.scAdd(sc, afterImageTime_time, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.stateParam(is, "value", func(data string) error {
				b = true
				return c.scAdd(sc, afterImageTime_time, data, VT_Int, 1)
			}); err != nil {
				return err
			}
			if !b {
				sc.add(afterImageTime_time, sc.iToExp(0))
			}
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) hitDefSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.stateParam(is, "attr", func(data string) error {
		attr, err := c.attr(data, true)
		if err != nil {
			return err
		}
		sc.add(hitDef_attr, sc.iToExp(attr))
		return nil
	}); err != nil {
		return err
	}
	hflg := func(id byte, data string) error {
		var flg int32
		for _, c := range data {
			switch c {
			case 'H', 'h':
				flg |= int32(ST_S)
			case 'L', 'l':
				flg |= int32(ST_C)
			case 'M', 'm':
				flg |= int32(ST_S | ST_C)
			case 'A', 'a':
				flg |= int32(ST_A)
			case 'F', 'f':
				flg |= int32(ST_F)
			case 'D', 'd':
				flg |= int32(ST_D)
			case 'P', 'p':
				flg |= int32(ST_P)
			case '-':
				flg |= int32(MT_MNS)
			case '+':
				flg |= int32(MT_PLS)
			}
		}
		sc.add(id, sc.iToExp(flg))
		return nil
	}
	if err := c.stateParam(is, "guardflag", func(data string) error {
		return hflg(hitDef_guardflag, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "hitflag", func(data string) error {
		return hflg(hitDef_hitflag, data)
	}); err != nil {
		return err
	}
	htyp := func(id byte, data string) error {
		if len(data) == 0 {
			return Error("Value not specified")
		}
		var ht HitType
		switch data[0] {
		case 'H', 'h':
			ht = HT_High
		case 'L', 'l':
			ht = HT_Low
		case 'T', 't':
			ht = HT_Trip
		case 'N', 'n':
			ht = HT_None
		default:
			return Error("Invalid value: " + data)
		}
		sc.add(id, sc.iToExp(int32(ht)))
		return nil
	}
	if err := c.stateParam(is, "ground.type", func(data string) error {
		return htyp(hitDef_ground_type, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "air.type", func(data string) error {
		return htyp(hitDef_air_type, data)
	}); err != nil {
		return err
	}
	reac := func(id byte, data string) error {
		if len(data) == 0 {
			return Error("Value not specified")
		}
		var ra Reaction
		switch data[0] {
		case 'L', 'l':
			ra = RA_Light
		case 'M', 'm':
			ra = RA_Medium
		case 'H', 'h':
			ra = RA_Hard
		case 'B', 'b':
			ra = RA_Back
		case 'U', 'u':
			ra = RA_Up
		case 'D', 'd':
			ra = RA_Diagup
		default:
			return Error("Invalid value: " + data)
		}
		sc.add(id, sc.iToExp(int32(ra)))
		return nil
	}
	if err := c.stateParam(is, "animtype", func(data string) error {
		return reac(hitDef_animtype, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "air.animtype", func(data string) error {
		return reac(hitDef_air_animtype, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "fall.animtype", func(data string) error {
		return reac(hitDef_fall_animtype, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "affectteam", func(data string) error {
		if len(data) == 0 {
			return Error("Value not specified")
		}
		var at int32
		switch data[0] {
		case 'E', 'e':
			at = 1
		case 'B', 'b':
			at = 0
		case 'F', 'f':
			at = -1
		default:
			return Error("Invalid value: " + data)
		}
		sc.add(hitDef_affectteam, sc.iToExp(at))
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "teamside",
		hitDef_teamside, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "id",
		hitDef_id, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "chainid",
		hitDef_chainid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "nochainid",
		hitDef_nochainid, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "kill",
		hitDef_kill, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.kill",
		hitDef_guard_kill, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.kill",
		hitDef_fall_kill, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "hitonce",
		hitDef_hitonce, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.juggle",
		hitDef_air_juggle, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "getpower",
		hitDef_getpower, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "damage",
		hitDef_damage, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "givepower",
		hitDef_givepower, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "numhits",
		hitDef_numhits, VT_Int, 1, false); err != nil {
		return err
	}
	hsnd := func(id byte, data string) error {
		fflg := true
		if len(data) > 1 {
			if strings.ToLower(data)[0] == 'f' || strings.ToLower(data)[0] == 's' {
				re := regexp.MustCompile("[^a-z]")
				m := re.Split(strings.ToLower(data)[1:], -1)
				if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
					fflg = strings.ToLower(data)[0] == 'f'
					data = data[1:]
				}
			}
		}
		return c.scAdd(sc, id, data, VT_Int, 2, sc.iToExp(Btoi(fflg))...)
	}
	if err := c.stateParam(is, "hitsound", func(data string) error {
		return hsnd(hitDef_hitsound, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "guardsound", func(data string) error {
		return hsnd(hitDef_guardsound, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "priority", func(data string) error {
		be, err := c.argExpression(&data, VT_Int)
		if err != nil {
			return err
		}
		at := AT_Hit
		data = strings.TrimSpace(data)
		if c.token == "," && len(data) > 0 {
			switch data[0] {
			case 'H', 'h':
				at = AT_Hit
			case 'M', 'm':
				at = AT_Miss
			case 'D', 'd':
				at = AT_Dodge
			default:
				return Error("Invalid value: " + data)
			}
		}
		sc.add(hitDef_priority, append(sc.beToExp(be), sc.iToExp(int32(at))...))
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p1stateno",
		hitDef_p1stateno, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p2stateno",
		hitDef_p2stateno, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p2getp1state",
		hitDef_p2getp1state, VT_Bool, 1, false); err != nil {
		return err
	}
	b := false
	if err := c.stateParam(is, "p1sprpriority", func(data string) error {
		b = true
		return c.scAdd(sc, hitDef_p1sprpriority, data, VT_Int, 1)
	}); err != nil {
		return err
	}
	if !b {
		if err := c.paramValue(is, sc, "sprpriority",
			hitDef_p1sprpriority, VT_Int, 1, false); err != nil {
			return err
		}
	}
	if err := c.paramValue(is, sc, "p2sprpriority",
		hitDef_p2sprpriority, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "forcestand",
		hitDef_forcestand, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "forcenofall",
		hitDef_forcenofall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.damage",
		hitDef_fall_damage, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.xvelocity",
		hitDef_fall_xvelocity, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.yvelocity",
		hitDef_fall_yvelocity, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.recover",
		hitDef_fall_recover, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.recovertime",
		hitDef_fall_recovertime, VT_Int, 1, false); err != nil {
		return err
	}
	sprk := func(id byte, data string) error {
		fflg := true
		if len(data) > 1 {
			if strings.ToLower(data)[0] == 's' {
				re := regexp.MustCompile("[^a-z]")
				m := re.Split(strings.ToLower(data)[1:], -1)
				if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
					fflg = false
					data = data[1:]
				}
			}
		}
		return c.scAdd(sc, id, data, VT_Int, 1, sc.iToExp(Btoi(fflg))...)
	}
	if err := c.stateParam(is, "sparkno", func(data string) error {
		return sprk(hitDef_sparkno, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "guard.sparkno", func(data string) error {
		return sprk(hitDef_guard_sparkno, data)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "sparkxy",
		hitDef_sparkxy, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.hittime",
		hitDef_down_hittime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p1facing",
		hitDef_p1facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p1getp2facing",
		hitDef_p1getp2facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "mindist",
		hitDef_mindist, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "maxdist",
		hitDef_maxdist, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "snap",
		hitDef_snap, VT_Float, 4, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p2facing",
		hitDef_p2facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.hittime",
		hitDef_air_hittime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall",
		hitDef_fall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.fall",
		hitDef_air_fall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.cornerpush.veloff",
		hitDef_air_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.bounce",
		hitDef_down_bounce, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.velocity",
		hitDef_down_velocity, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.cornerpush.veloff",
		hitDef_down_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ground.hittime",
		hitDef_ground_hittime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.hittime",
		hitDef_guard_hittime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.dist",
		hitDef_guard_dist, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pausetime",
		hitDef_pausetime, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.pausetime",
		hitDef_guard_pausetime, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.velocity",
		hitDef_air_velocity, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "airguard.velocity",
		hitDef_airguard_velocity, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ground.slidetime",
		hitDef_ground_slidetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.slidetime",
		hitDef_guard_slidetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.ctrltime",
		hitDef_guard_ctrltime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "airguard.ctrltime",
		hitDef_airguard_ctrltime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "ground.velocity", func(data string) error {
		in := data
		if c.token = c.tokenizer(&in); c.token == "n" {
			if c.token = c.tokenizer(&in); len(c.token) > 0 && c.token != "," {
				return Error("Invalid data: " + c.token)
			}
		} else {
			in = data
			be, err := c.argExpression(&in, VT_Float)
			if err != nil {
				return err
			}
			sc.add(hitDef_ground_velocity_x, sc.beToExp(be))
		}
		if c.token == "," {
			oldin := in
			if c.token = c.tokenizer(&in); c.token == "n" {
				if c.token = c.tokenizer(&in); len(c.token) > 0 {
					return Error("Invalid data: " + c.token)
				}
			} else {
				in = oldin
				be, err := c.fullExpression(&in, VT_Float)
				if err != nil {
					return err
				}
				sc.add(hitDef_ground_velocity_y, sc.beToExp(be))
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.velocity",
		hitDef_guard_velocity, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ground.cornerpush.veloff",
		hitDef_ground_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.cornerpush.veloff",
		hitDef_guard_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "airguard.cornerpush.veloff",
		hitDef_airguard_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "yaccel",
		hitDef_yaccel, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.palFXSub(is, sc, "palfx."); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.time",
		hitDef_envshake_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.ampl",
		hitDef_envshake_ampl, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.phase",
		hitDef_envshake_phase, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.freq",
		hitDef_envshake_freq, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.time",
		hitDef_fall_envshake_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.ampl",
		hitDef_fall_envshake_ampl, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.phase",
		hitDef_fall_envshake_phase, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.freq",
		hitDef_fall_envshake_freq, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "dizzypoints",
		hitDef_dizzypoints, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guardpoints",
		hitDef_guardpoints, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "redlife",
		hitDef_redlife, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "score",
		hitDef_score, VT_Float, 2, false); err != nil {
		return err
	}
	return nil
}
func (c *Compiler) hitDef(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*hitDef)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitDef_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.hitDefSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) reversalDef(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*reversalDef)(sc), c.stateSec(is, func() error {
		attr := int32(-1)
		var err error
		if err := c.paramValue(is, sc, "redirectid",
			reversalDef_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err = c.stateParam(is, "reversal.attr", func(data string) error {
			attr, err = c.attr(data, false)
			return err
		}); err != nil {
			return err
		}
		if attr == -1 {
			return Error("reversal.attr parameter not specified")
		}
		sc.add(reversalDef_reversal_attr, sc.iToExp(attr))
		return c.hitDefSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) projectile(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error) {
	ret, err := (*projectile)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			projectile_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramPostye(is, sc, projectile_postype); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projid",
			projectile_projid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projremove",
			projectile_projremove, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projremovetime",
			projectile_projremovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projshadow",
			projectile_projshadow, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projmisstime",
			projectile_projmisstime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projhits",
			projectile_projhits, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projpriority",
			projectile_projpriority, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "projhitanim", func(data string) error {
			fflg := false
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = true
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, projectile_projhitanim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "projremanim", func(data string) error {
			fflg := false
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = true
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, projectile_projremanim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "projcancelanim", func(data string) error {
			fflg := false
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = true
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, projectile_projcancelanim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "velocity",
			projectile_velocity, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "velmul",
			projectile_velmul, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "remvelocity",
			projectile_remvelocity, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "accel",
			projectile_accel, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projscale",
			projectile_projscale, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projangle",
			projectile_projangle, VT_Float, 1, false); err != nil {
			return err
		}

		// hitdef部分
		if err := c.hitDefSub(is, sc); err != nil {
			return err
		}

		if err := c.paramValue(is, sc, "offset",
			projectile_offset, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projsprpriority",
			projectile_projsprpriority, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projstagebound",
			projectile_projstagebound, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projedgebound",
			projectile_projedgebound, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "projheightbound",
			projectile_projheightbound, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "projanim", func(data string) error {
			fflg := false
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = true
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, projectile_projanim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "supermovetime",
			projectile_supermovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pausemovetime",
			projectile_pausemovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ownpal",
			projectile_ownpal, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "remappal",
			projectile_remappal, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.afterImageSub(is, sc, ihp, "afterimage."); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "platform",
			projectile_platform, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "platformwidth",
			projectile_platformwidth, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "platformheight",
			projectile_platformheight, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "platformangle",
			projectile_platformangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "platformfence",
			projectile_platformfence, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) width(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*width)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			width_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "edge", func(data string) error {
			b = true
			if len(data) == 0 {
				return nil
			}
			return c.scAdd(sc, width_edge, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "player", func(data string) error {
			b = true
			if len(data) == 0 {
				return nil
			}
			return c.scAdd(sc, width_player, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.paramValue(is, sc, "value",
				width_value, VT_Float, 2, true); err != nil {
				return err
			}
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) sprPriority(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*sprPriority)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			sprPriority_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value",
			sprPriority_value, VT_Int, 1, false)
	})
	return *ret, err
}
func (c *Compiler) varSetSub(is IniSection,
	sc *StateControllerBase, rd OpCode, oc OpCode) error {
	b, v, fv := false, false, false
	var value string
	if err := c.stateParam(is, "value", func(data string) error {
		b = true
		value = data
		return nil
	}); err != nil {
		return err
	}
	if b {
		var ve BytecodeExp
		if err := c.stateParam(is, "v", func(data string) (err error) {
			v = true
			ve, err = c.fullExpression(&data, VT_Int)
			return
		}); err != nil {
			return err
		}
		if !v {
			if err := c.stateParam(is, "fv", func(data string) (err error) {
				fv = true
				ve, err = c.fullExpression(&data, VT_Int)
				return
			}); err != nil {
				return err
			}
		}
		if v || fv {
			if len(ve) == 2 && ve[0] == OC_int8 && int8(ve[1]) >= 0 &&
				(v && ve[1] < NumVar || fv && ve[1] < NumFvar) {
				if oc == OC_st_var {
					if v {
						oc = OC_st_var0 + ve[1]
					} else {
						oc = OC_st_fvar0 + ve[1]
					}
				} else {
					if v {
						oc = OC_st_var0add + ve[1]
					} else {
						oc = OC_st_fvar0add + ve[1]
					}
				}
				ve = nil
			} else if oc == OC_st_var {
				if v {
					oc = OC_st_var
				} else {
					oc = OC_st_fvar
				}
			} else {
				if v {
					oc = OC_st_varadd
				} else {
					oc = OC_st_fvaradd
				}
			}
			var vt ValueType
			if v {
				vt = VT_Int
			} else {
				vt = VT_Float
			}
			in := value
			be, err := c.fullExpression(&in, vt)
			if err != nil {
				return Error(value + "\n" + "value: " + err.Error())
			}
			ve.append(be...)
			if rd != OC_rdreset {
				var tmp BytecodeExp
				tmp.appendI32Op(OC_nordrun, int32(len(ve)))
				ve.append(OC_st_, oc)
				ve = append(tmp, ve...)
				tmp = nil
				tmp.appendI32Op(rd, int32(len(ve)))
				ve = append(tmp, ve...)
			} else {
				ve.append(OC_st_, oc)
			}
			sc.add(varSet_, sc.beToExp(ve))
		}
		return nil
	}
	sys := false
	set := func(data string) error {
		data = strings.TrimSpace(data)
		if data[0] != '(' {
			return Error("Missing '('")
		}
		var be BytecodeExp
		c.token = c.tokenizer(&data)
		bv, err := c.expValue(&be, &data, false)
		if err != nil {
			return err
		}
		_else := false
		if !bv.IsNone() {
			i := bv.ToI()
			if i >= 0 && (!sys && v && i < int32(NumVar) ||
				!sys && fv && i < int32(NumFvar) || sys && v && i < int32(NumSysVar) ||
				sys && fv && i < int32(NumSysFvar)) {
				if v {
					if oc == OC_st_var {
						oc = OC_st_var0 + OpCode(i)
					} else {
						oc = OC_st_var0add + OpCode(i)
					}
					if sys {
						oc += NumVar
					}
				} else {
					if oc == OC_st_var {
						oc = OC_st_fvar0 + OpCode(i)
					} else {
						oc = OC_st_fvar0add + OpCode(i)
					}
					if sys {
						oc += NumFvar
					}
				}
			} else {
				be.appendValue(bv)
				_else = true
			}
		} else {
			_else = true
		}
		if _else {
			if oc == OC_st_var {
				if sys {
					if v {
						oc = OC_st_sysvar
					} else {
						oc = OC_st_sysfvar
					}
				} else {
					if v {
						oc = OC_st_var
					} else {
						oc = OC_st_fvar
					}
				}
			} else {
				if sys {
					if v {
						oc = OC_st_sysvaradd
					} else {
						oc = OC_st_sysfvaradd
					}
				} else {
					if v {
						oc = OC_st_varadd
					} else {
						oc = OC_st_fvaradd
					}
				}
			}
		}
		if len(c.token) == 0 || c.token[len(c.token)-1] != '=' {
			idx := strings.Index(data, "=")
			if idx < 0 {
				return Error("Missing '='")
			}
			data = data[idx+1:]
		}
		var vt ValueType
		if v {
			vt = VT_Int
		} else {
			vt = VT_Float
		}
		ve := be
		be, err = c.fullExpression(&data, vt)
		if err != nil {
			return err
		}
		ve.append(be...)
		if rd != OC_rdreset {
			var tmp BytecodeExp
			tmp.appendI32Op(OC_nordrun, int32(len(ve)))
			ve.append(OC_st_, oc)
			ve = append(tmp, ve...)
			tmp = nil
			tmp.appendI32Op(rd, int32(len(ve)))
			ve = append(tmp, ve...)
		} else {
			ve.append(OC_st_, oc)
		}
		sc.add(varSet_, sc.beToExp(ve))
		return nil
	}
	if err := c.stateParam(is, "var", func(data string) error {
		if data[0] != 'v' {
			return Error(data[:3] + "'v' is not lowercase")
		}
		b = true
		v = true
		return set(data[3:])
	}); err != nil {
		return err
	}
	if b {
		return nil
	}
	if err := c.stateParam(is, "fvar", func(data string) error {
		if rd == OC_rdreset && data[0] != 'f' {
			return Error(data[:4] + "'f' is not lowercase")
		}
		b = true
		fv = true
		return set(data[4:])
	}); err != nil {
		return err
	}
	if b {
		return nil
	}
	if err := c.stateParam(is, "sysvar", func(data string) error {
		if data[3] != 'v' {
			return Error(data[:6] + "'v' is not lowercase")
		}
		b = true
		v = true
		sys = true
		return set(data[6:])
	}); err != nil {
		return err
	}
	if b {
		return nil
	}
	if err := c.stateParam(is, "sysfvar", func(data string) error {
		b = true
		fv = true
		sys = true
		return set(data[7:])
	}); err != nil {
		return err
	}
	if b {
		return nil
	}
	return Error("value parameter not specified")
}
func (c *Compiler) varSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_rdreset, OC_st_var)
	})
	return *ret, err
}
func (c *Compiler) varAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_rdreset, OC_st_varadd)
	})
	return *ret, err
}
func (c *Compiler) parentVarSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_parent, OC_st_var)
	})
	return *ret, err
}
func (c *Compiler) parentVarAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_parent, OC_st_varadd)
	})
	return *ret, err
}
func (c *Compiler) rootVarSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_root, OC_st_var)
	})
	return *ret, err
}
func (c *Compiler) rootVarAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_root, OC_st_varadd)
	})
	return *ret, err
}
func (c *Compiler) turn(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*turn)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			turn_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(turn_, nil)
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetFacing(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetFacing)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetFacing_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetFacing_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetFacing_value, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetBind(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetBind)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetBind_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetBind_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			targetBind_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			targetBind_pos, VT_Float, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) bindToTarget(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*bindToTarget)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			bindToTarget_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			bindToTarget_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			bindToTarget_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "pos", func(data string) error {
			be, err := c.argExpression(&data, VT_Float)
			if err != nil {
				return err
			}
			exp := sc.beToExp(be)
			if c.token != "," {
				sc.add(bindToTarget_pos, exp)
				return nil
			}
			if be, err = c.argExpression(&data, VT_Float); err != nil {
				return err
			}
			exp, data = append(exp, be), strings.TrimSpace(data)
			if c.token != "," || len(data) == 0 {
				sc.add(bindToTarget_pos, exp)
				return nil
			}
			var hmf HMF
			switch data[0] {
			case 'H', 'h':
				hmf = HMF_H
			case 'M', 'm':
				hmf = HMF_M
			case 'F', 'f':
				hmf = HMF_F
			default:
				return Error("Invalid value: " + data)
			}
			sc.add(bindToTarget_pos, append(exp, sc.iToExp(int32(hmf))...))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetLifeAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetLifeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetLifeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetLifeAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			targetLifeAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "kill",
			targetLifeAdd_kill, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetLifeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetState(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetState)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetState_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetState_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetState_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetVelSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetVelSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetVelSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetVelSet_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			targetVelSet_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			targetVelSet_y, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetVelAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetVelAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetVelAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetVelAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			targetVelAdd_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			targetVelAdd_y, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetPowerAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetPowerAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetPowerAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetPowerAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetPowerAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetDrop(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetDrop)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetDrop_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "excludeid",
			targetDrop_excludeid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "keepone",
			targetDrop_keepone, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) lifeAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*lifeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			lifeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			lifeAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "kill",
			lifeAdd_kill, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			lifeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) lifeSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*lifeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			lifeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", lifeSet_value, VT_Int, 1, true)
	})
	return *ret, err
}
func (c *Compiler) powerAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*powerAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			powerAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", powerAdd_value, VT_Int, 1, true)
	})
	return *ret, err
}
func (c *Compiler) powerSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*powerSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			powerSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", powerSet_value, VT_Int, 1, true)
	})
	return *ret, err
}
func (c *Compiler) hitVelSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*hitVelSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitVelSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			hitVelSet_x, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			hitVelSet_y, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) screenBound(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*screenBound)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			screenBound_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", func(data string) error {
			b = true
			return c.scAdd(sc, screenBound_value, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(screenBound_value, sc.iToExp(0))
		}
		b = false
		if err := c.stateParam(is, "movecamera", func(data string) error {
			b = true
			return c.scAdd(sc, screenBound_movecamera, data, VT_Bool, 2)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(screenBound_movecamera, append(sc.iToExp(0), sc.iToExp(0)...))
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) posFreeze(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*posFreeze)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posFreeze_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", func(data string) error {
			b = true
			return c.scAdd(sc, posFreeze_value, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(posFreeze_value, sc.iToExp(1))
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) envShake(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*envShake)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "time",
			envShake_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ampl",
			envShake_ampl, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "phase",
			envShake_phase, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "freq",
			envShake_freq, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) hitOverride(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*hitOverride)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitOverride_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "attr", func(data string) error {
			attr, err := c.attr(data, false)
			if err != nil {
				return err
			}
			sc.add(hitOverride_attr, sc.iToExp(attr))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "slot",
			hitOverride_slot, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stateno",
			hitOverride_stateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			hitOverride_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "forceair",
			hitOverride_forceair, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) pause(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*pause)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			pause_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			pause_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "movetime",
			pause_movetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pausebg",
			pause_pausebg, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "endcmdbuftime",
			pause_endcmdbuftime, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) superPause(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*superPause)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			superPause_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			superPause_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "movetime",
			superPause_movetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pausebg",
			superPause_pausebg, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "endcmdbuftime",
			superPause_endcmdbuftime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "darken",
			superPause_darken, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", func(data string) error {
			fflg := true
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' || strings.ToLower(data)[0] == 's' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = strings.ToLower(data)[0] == 'f'
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, superPause_anim, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			superPause_pos, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "p2defmul",
			superPause_p2defmul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "poweradd",
			superPause_poweradd, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "unhittable",
			superPause_unhittable, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "sound", func(data string) error {
			fflg := true
			if len(data) > 1 {
				if strings.ToLower(data)[0] == 'f' || strings.ToLower(data)[0] == 's' {
					re := regexp.MustCompile("[^a-z]")
					m := re.Split(strings.ToLower(data)[1:], -1)
					if _, ok := triggerMap[m[0]]; ok || m[0] == "" {
						fflg = strings.ToLower(data)[0] == 'f'
						data = data[1:]
					}
				}
			}
			return c.scAdd(sc, superPause_sound, data, VT_Int, 2,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) trans(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*trans)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			trans_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramTrans(is, sc, "", trans_trans, false)
	})
	return *ret, err
}
func (c *Compiler) playerPush(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*playerPush)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			playerPush_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", func(data string) error {
			b = true
			return c.scAdd(sc, playerPush_value, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(playerPush_value, sc.iToExp(1))
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) stateTypeSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*stateTypeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			stateTypeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		statetype := func(data string) error {
			if len(data) == 0 {
				return Error("Value not specified")
			}
			var st StateType
			switch strings.ToLower(data)[0] {
			case 's':
				st = ST_S
			case 'c':
				st = ST_C
			case 'a':
				st = ST_A
			case 'l':
				st = ST_L
			default:
				return Error("Invalid value: " + data)
			}
			sc.add(stateTypeSet_statetype, sc.iToExp(int32(st)))
			return nil
		}
		b := false
		if err := c.stateParam(is, "statetype", func(data string) error {
			b = true
			return statetype(data)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.stateParam(is, "value", func(data string) error {
				return statetype(data)
			}); err != nil {
				return err
			}
		}
		if err := c.stateParam(is, "movetype", func(data string) error {
			if len(data) == 0 {
				return Error("Value not specified")
			}
			var mt MoveType
			switch strings.ToLower(data)[0] {
			case 'i':
				mt = MT_I
			case 'a':
				mt = MT_A
			case 'h':
				mt = MT_H
			default:
				return Error("Invalid value: " + data)
			}
			sc.add(stateTypeSet_movetype, sc.iToExp(int32(mt)))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "physics", func(data string) error {
			if len(data) == 0 {
				return Error("Value not specified")
			}
			var st StateType
			switch strings.ToLower(data)[0] {
			case 's':
				st = ST_S
			case 'c':
				st = ST_C
			case 'a':
				st = ST_A
			case 'n':
				st = ST_N
			default:
				return Error("Invalid value: " + data)
			}
			sc.add(stateTypeSet_physics, sc.iToExp(int32(st)))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) angleDraw(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*angleDraw)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			angleDraw_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			angleDraw_value, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scale",
			angleDraw_scale, VT_Float, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) angleSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*angleSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			angleSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			angleSet_value, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) angleAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*angleAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			angleAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			angleAdd_value, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) angleMul(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*angleMul)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			angleMul_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			angleMul_value, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) envColor(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*envColor)(sc), c.stateSec(is, func() error {
		if err := c.stateParam(is, "value", func(data string) error {
			bes, err := c.exprs(data, VT_Int, 3)
			if err != nil {
				return err
			}
			if len(bes) < 3 {
				return Error("value - not enough arguments")
			}
			sc.add(envColor_value, bes)
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			envColor_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "under",
			envColor_under, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) displayToClipboardSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "redirectid",
		displayToClipboard_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "params", func(data string) error {
		bes, err := c.exprs(data, VT_SFalse, 100000)
		if err != nil {
			return err
		}
		sc.add(displayToClipboard_params, bes)
		return nil
	}); err != nil {
		return err
	}
	b := false
	if err := c.stateParam(is, "text", func(data string) error {
		b = true
		_else := false
		if len(data) >= 2 && data[0] == '"' {
			if i := strings.Index(data[1:], "\""); i >= 0 {
				data, _ = strconv.Unquote(data)
			} else {
				_else = true
			}
		} else {
			_else = true
		}
		if _else {
			return Error("Not enclosed in \"")
		}
		sc.add(displayToClipboard_text,
			sc.iToExp(int32(sys.stringPool[c.playerNo].Add(data))))
		return nil
	}); err != nil {
		return err
	}
	if !b {
		return Error("text parameter not specified")
	}
	return nil
}
func (c *Compiler) displayToClipboard(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*displayToClipboard)(sc), c.stateSec(is, func() error {
		return c.displayToClipboardSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) appendToClipboard(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*appendToClipboard)(sc), c.stateSec(is, func() error {
		return c.displayToClipboardSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) clearClipboard(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*clearClipboard)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			clearClipboard_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(clearClipboard_, nil)
		return nil
	})
	return *ret, err
}
func (c *Compiler) makeDust(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*makeDust)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			makeDust_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "spacing", func(data string) error {
			b = true
			return c.scAdd(sc, makeDust_spacing, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(makeDust_spacing, sc.iToExp(3))
		}
		b = false
		if err := c.stateParam(is, "pos", func(data string) error {
			b = true
			return c.scAdd(sc, makeDust_pos, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(makeDust_pos, sc.iToExp(0))
		}
		if err := c.paramValue(is, sc, "pos2",
			makeDust_pos2, VT_Float, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) attackDist(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*attackDist)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			attackDist_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			attackDist_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) attackMulSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*attackMulSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			attackMulSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			attackMulSet_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) defenceMulSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*defenceMulSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			defenceMulSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			defenceMulSet_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) fallEnvShake(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*fallEnvShake)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			fallEnvShake_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(fallEnvShake_, nil)
		return nil
	})
	return *ret, err
}
func (c *Compiler) hitFallDamage(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*hitFallDamage)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitFallDamage_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(hitFallDamage_, nil)
		return nil
	})
	return *ret, err
}
func (c *Compiler) hitFallVel(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*hitFallVel)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitFallVel_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(hitFallVel_, nil)
		return nil
	})
	return *ret, err
}
func (c *Compiler) hitFallSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*hitFallSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitFallSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", func(data string) error {
			b = true
			return c.scAdd(sc, hitFallSet_value, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(hitFallSet_value, sc.iToExp(-1))
		}
		if err := c.paramValue(is, sc, "xvel",
			hitFallSet_xvel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yvel",
			hitFallSet_yvel, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) varRangeSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*varRangeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varRangeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "first",
			varRangeSet_first, VT_Int, 1, false); err != nil {
			return err
		}
		last := false
		if err := c.stateParam(is, "last", func(data string) error {
			last = true
			return c.scAdd(sc, varRangeSet_last, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", func(data string) error {
			b = true
			if !last {
				sc.add(varRangeSet_last, sc.iToExp(int32(NumVar-1)))
			}
			return c.scAdd(sc, varRangeSet_value, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.stateParam(is, "fvalue", func(data string) error {
				b = true
				if !last {
					sc.add(varRangeSet_last, sc.iToExp(int32(NumFvar-1)))
				}
				return c.scAdd(sc, varRangeSet_fvalue, data, VT_Float, 1)
			}); err != nil {
				return err
			}
			if !b {
				return Error("value parameter not specified")
			}
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) remapPal(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*remapPal)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			remapPal_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "source",
			remapPal_source, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dest",
			remapPal_dest, VT_Int, 2, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) stopSnd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*stopSnd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			stopSnd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "channel",
			stopSnd_channel, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) sndPan(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*sndPan)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			sndPan_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "channel",
			sndPan_channel, VT_Int, 1, true); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pan",
			sndPan_pan, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "abspan",
			sndPan_abspan, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) varRandom(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*varRandom)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varRandom_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "v",
			varRandom_v, VT_Int, 1, true); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "range",
			varRandom_range, VT_Int, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) gravity(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*gravity)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			gravity_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(gravity_, nil)
		return nil
	})
	return *ret, err
}
func (c *Compiler) bindToParentSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "redirectid",
		bindToParent_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "time",
		bindToParent_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "facing",
		bindToParent_facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pos",
		bindToParent_pos, VT_Float, 2, false); err != nil {
		return err
	}
	return nil
}
func (c *Compiler) bindToParent(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*bindToParent)(sc), c.stateSec(is, func() error {
		return c.bindToParentSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) bindToRoot(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*bindToRoot)(sc), c.stateSec(is, func() error {
		return c.bindToParentSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) removeExplod(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*removeExplod)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			removeExplod_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "id", func(data string) error {
			b = true
			return c.scAdd(sc, removeExplod_id, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(removeExplod_id, sc.iToExp(-1))
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) explodBindTime(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*explodBindTime)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			explodBindTime_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			explodBindTime_id, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "time", func(data string) error {
			b = true
			return c.scAdd(sc, explodBindTime_time, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.paramValue(is, sc, "value",
				explodBindTime_time, VT_Int, 1, false); err != nil {
				return err
			}
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) moveHitReset(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*moveHitReset)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			moveHitReset_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(moveHitReset_, nil)
		return nil
	})
	return *ret, err
}
func (c *Compiler) hitAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*hitAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			hitAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) offset(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*offset)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			offset_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			offset_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			offset_y, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) victoryQuote(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*victoryQuote)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			victoryQuote_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			victoryQuote_value, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) zoom(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*zoom)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			zoom_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			zoom_pos, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scale",
			zoom_scale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "lag",
			zoom_lag, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) dialogue(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*dialogue)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			dialogue_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "hidebars",
			dialogue_hidebars, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "force",
			dialogue_force, VT_Bool, 1, false); err != nil {
			return err
		}
		var keys []int
		r, _ := regexp.Compile("^text[0-9]+$")
		for k := range is {
			if r.MatchString(k) {
				re := regexp.MustCompile("[0-9]+")
				submatchall := re.FindAllString(k, -1)
				if len(submatchall) == 1 {
					keys = append(keys, int(Atoi(submatchall[0])))
				}
			}
		}
		sort.Ints(keys)
		for _, key := range keys {
			if err := c.stateParam(is, fmt.Sprintf("text%v", key), func(data string) error {
				if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
					return Error("Not enclosed in \"")
				}
				sc.add(dialogue_text, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) dizzyPointsAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*dizzyPointsAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			dizzyPointsAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", dizzyPointsAdd_value, VT_Int, 1, true)
	})
	return *ret, err
}
func (c *Compiler) dizzyPointsSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*dizzyPointsSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			dizzyPointsSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", dizzyPointsSet_value, VT_Int, 1, true)
	})
	return *ret, err
}
func (c *Compiler) dizzySet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*dizzySet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			dizzySet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", dizzySet_value, VT_Bool, 1, true)
	})
	return *ret, err
}
func (c *Compiler) guardBreakSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*guardBreakSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			guardBreakSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", guardBreakSet_value, VT_Bool, 1, true)
	})
	return *ret, err
}
func (c *Compiler) guardPointsAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*guardPointsAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			guardPointsAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", guardPointsAdd_value, VT_Int, 1, true)
	})
	return *ret, err
}
func (c *Compiler) guardPointsSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*guardPointsSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			guardPointsSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", guardPointsSet_value, VT_Int, 1, true)
	})
	return *ret, err
}

// Parse hitScaleSet ini section.
func (c *Compiler) hitScaleSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitScaleSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitScaleSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		// Parse affects
		if err := c.stateParam(is, "affects", func(data string) error {
			// We do really need to add string support.
			var arrayData []string
			var err2 error

			if arrayData, err2 = cnsStringArray(data); err2 != nil {
				return err2
			}

			// Send values in the array to hitScaleSet.run().
			for _, str := range arrayData {
				switch str {
				case "damage":
					sc.add(hitScaleSet_affects_damage, sc.bToExp(true))
				case "hitTime":
					sc.add(hitScaleSet_affects_hitTime, sc.bToExp(true))
				case "pauseTime":
					sc.add(hitScaleSet_affects_pauseTime, sc.bToExp(true))
				default:
					return Error("Invalid 'affects' value.")
				}
			}

			return nil
		}); err != nil {
			return err
		}

		if err := c.paramValue(is, sc, "id",
			hitScaleSet_id, VT_Int, 1, false); err != nil {
			return err
		}
		// Parse reset, valid values are 0, 1 and 2.
		// If the value is not valid throw a error.
		if err := c.stateParam(is, "reset", func(data string) error {
			var reset = Atoi(strings.TrimSpace(data))

			if reset < 0 && reset > 2 {
				sc.add(hitScaleSet_reset, sc.iToExp(reset))
				return nil
			} else {
				return Error(`Invalid "reset" value.`)
			}
		}); err != nil {
			return err
		}

		if err := c.paramValue(is, sc, "force",
			hitScaleSet_force, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "mul",
			hitScaleSet_mul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "add",
			hitScaleSet_add, VT_Int, 1, false); err != nil {
			return err
		}
		// The only valid values of addType are "mulFirst" and "addFirst"
		if err := c.stateParam(is, "addType", func(data string) error {
			var push = 0
			// Change sting to lowecase and remove quotes
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			} else {
				data, _ = strconv.Unquote(data)
			}

			if data == "addFirst" {
				push = 1
			} else if data == "mulFirst" {
				push = 2
			} else {
				return Error(`Invalid "addType" value.`)
			}

			sc.add(hitScaleSet_addType, sc.iToExp(int32(push)))

			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "min",
			hitScaleSet_min, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "max",
			hitScaleSet_max, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			hitScaleSet_time, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
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
				return nil, Error("WARNING: Unknown string array error.")
			}
		} else if formatError {
			return nil, Error("Wrong format on string array.")
		} else { // All's good.
			inString = 0
		}
	} // Return the parsed string array,
	return fullStrArray, nil
}

func (c *Compiler) lifebarAction(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*lifebarAction)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			lifebarAction_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "top",
			lifebarAction_top, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "timemul",
			lifebarAction_timemul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			lifebarAction_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "anim",
			lifebarAction_anim, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "spr",
			lifebarAction_spr, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "snd",
			lifebarAction_snd, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "text", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(lifebarAction_text, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) loadFile(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*loadFile)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			loadFile_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "path", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(loadFile_path, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramSaveData(is, sc, loadFile_saveData); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

// TODO: Remove boilderplate from the Map's Compiler.
func (c *Compiler) mapSetSub(is IniSection, sc *StateControllerBase) error {
	err := c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			mapSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "map", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(mapSet_mapArray, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			mapSet_value, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return err
}
func (c *Compiler) mapSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "0", VT_Int, 1)

	return *ret, err
}
func (c *Compiler) mapAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "1", VT_Int, 1)

	return *ret, err
}
func (c *Compiler) parentMapSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "2", VT_Int, 1)

	return *ret, err
}
func (c *Compiler) parentMapAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "3", VT_Int, 1)

	return *ret, err
}
func (c *Compiler) rootMapSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "4", VT_Int, 1)

	return *ret, err
}
func (c *Compiler) rootMapAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "5", VT_Int, 1)

	return *ret, err
}
func (c *Compiler) teamMapSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "6", VT_Int, 1)

	return *ret, err
}
func (c *Compiler) teamMapAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "7", VT_Int, 1)

	return *ret, err
}
func (c *Compiler) matchRestart(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*matchRestart)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "reload",
			matchRestart_reload, VT_Bool, MaxSimul*2+MaxAttachedChar, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "stagedef", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_stagedef, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p1def", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_p1def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p2def", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_p2def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p3def", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_p3def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p4def", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_p4def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p5def", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_p5def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p6def", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_p6def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p7def", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_p7def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p8def", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(matchRestart_p8def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) printToConsole(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*printToConsole)(sc), c.stateSec(is, func() error {
		return c.displayToClipboardSub(is, sc)
	})
	return *ret, err
}
func (c *Compiler) rankAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*rankAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			rankAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "icon", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(rankAdd_icon, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "type", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(rankAdd_type, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "max",
			rankAdd_max, VT_Float, 1, true); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			rankAdd_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) redLifeAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*redLifeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			redLifeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			redLifeAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			redLifeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) redLifeSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*redLifeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			redLifeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", redLifeSet_value, VT_Int, 1, true)
	})
	return *ret, err
}
func (c *Compiler) remapSprite(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*remapSprite)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			remapSprite_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reset",
			remapSprite_reset, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "preset", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(remapSprite_preset, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "source",
			remapSprite_source, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dest",
			remapSprite_dest, VT_Int, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) roundTimeAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*roundTimeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			roundTimeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			roundTimeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) roundTimeSet(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*roundTimeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			roundTimeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", roundTimeSet_value, VT_Int, 1, true)
	})
	return *ret, err
}
func (c *Compiler) saveFile(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*saveFile)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			saveFile_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "path", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Not enclosed in \"")
			}
			sc.add(saveFile_path, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramSaveData(is, sc, saveFile_saveData); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) scoreAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*scoreAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			scoreAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			scoreAdd_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetDizzyPointsAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetDizzyPointsAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetDizzyPointsAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetDizzyPointsAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetDizzyPointsAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetGuardPointsAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetGuardPointsAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetGuardPointsAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetGuardPointsAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetGuardPointsAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetRedLifeAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetRedLifeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetRedLifeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetRedLifeAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			targetRedLifeAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetRedLifeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) targetScoreAdd(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	ret, err := (*targetScoreAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetScoreAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetScoreAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetScoreAdd_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) null(is IniSection, sc *StateControllerBase,
	_ int8) (StateController, error) {
	return nullStateController, nil
}

// Compile a state file
func (c *Compiler) stateCompile(states map[int32]StateBytecode,
	filename, def string) error {
	var str string
	zss := HasExtension(filename, "^\\.[Zz][Ss][Ss]$")
	fnz := filename
	// Load state file
	if err := LoadFile(&filename, def, func(filename string) error {
		var err error
		// If this is a zss file
		if zss {
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			str = string(b)
			return c.stateCompileZ(states, fnz, str)
		}

		// Try reading as an st file
		str, err = LoadText(filename)
		return err
	}); err != nil {
		// If filename doesn't exist, see if a zss file exists
		fnz += ".zss"
		if err := LoadFile(&fnz, def, func(filename string) error {
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			str = string(b)
			return nil
		}); err == nil {
			return c.stateCompileZ(states, fnz, str)
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

		c.stateNo = Atoi(line[10:])

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

		if _, ok := states[c.stateNo]; !ok || c.stateNo < 0 {
			states[c.stateNo] = *sbc
		}
	}
	return nil
}

func (c *Compiler) yokisinaiToken() error {
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
		i := strings.IndexAny((*line)[offset:], ";#\"{}")
		if i < 0 {
			assign = assign || strings.Contains((*line)[offset:], ":=")
			s, *line = *line, ""
			return
		}
		i += offset
		assign = assign || strings.Contains((*line)[offset:i], ":=")
		switch (*line)[i] {
		case ';', '{', '}':
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
		return c.yokisinaiToken()
	}
	c.token, *line = "", ""
	return nil
}
func (c *Compiler) readKeyValue(is IniSection, end string,
	line *string) error {
	name := c.scan(line)
	if name == "" || name == ":" {
		return c.yokisinaiToken()
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
				return nil, c.yokisinaiToken()
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
		return 0, c.yokisinaiToken()
	}
	if t == "-" && len(*line) > 0 && (*line)[0] >= '0' && (*line)[0] <= '9' {
		t += c.scan(line)
	}
	v, err := strconv.ParseInt(t, 10, 32)
	return int32(v), err
}
func (c *Compiler) subBlock(line *string, root bool,
	sbc *StateBytecode, numVars *int32) (*StateBlock, error) {
	bl := newStateBlock()
	for {
		switch c.token {
		case "ignorehitpause":
			if bl.ignorehitpause >= -1 {
				return nil, c.yokisinaiToken()
			}
			bl.ignorehitpause, bl.ctrlsIgnorehitpause = -1, true
			c.scan(line)
			continue
		case "persistent":
			if sbc == nil {
				return nil, Error("persistent cannot be used in a function")
			}
			if c.stateNo < 0 {
				return nil, Error("persistent cannot be used in a negative state")
			}
			if bl.persistentIndex >= 0 {
				return nil, c.yokisinaiToken()
			}
			c.scan(line)
			if err := c.needToken("("); err != nil {
				return nil, err
			}
			var err error
			if bl.persistent, err = c.scanI32(line); err != nil {
				return nil, err
			}
			c.scan(line)
			if err := c.needToken(")"); err != nil {
				return nil, err
			}
			if bl.persistent == 1 {
				return nil, Error("persistent(1) is meaningless")
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
	switch c.token {
	case "{":
	case "if":
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
	default:
		return nil, c.yokisinaiToken()
	}
	if err := c.stateBlock(line, bl, false,
		sbc, &bl.ctrls, numVars); err != nil {
		return nil, err
	}
	if root {
		if len(bl.trigger) > 0 {
			if c.token = c.tokenizer(line); c.token != "else" {
				if len(c.token) == 0 || c.token[0] == '#' {
					c.token, *line = "", ""
				} else {
					return nil, c.yokisinaiToken()
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
	if len(bl.trigger) > 0 && c.token == "else" {
		c.scan(line)
		var err error
		if bl.elseBlock, err = c.subBlock(line, root,
			sbc, numVars); err != nil {
			return nil, err
		}
		if bl.elseBlock.ignorehitpause >= -1 {
			bl.ignorehitpause = -1
		}
	}
	return bl, nil
}
func (c *Compiler) callFunc(line *string, root bool,
	ctrls *[]StateController, ret []uint8) error {
	var cf callFunction
	var ok bool
	cf.bytecodeFunction, ok = c.funcs[c.scan(line)]
	cf.ret = ret
	if !ok {
		if c.token == "" || c.token == "(" {
			return c.yokisinaiToken()
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
		return c.yokisinaiToken()
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
func (c *Compiler) stateBlock(line *string, bl *StateBlock, root bool,
	sbc *StateBytecode, ctrls *[]StateController, numVars *int32) error {
	c.scan(line)
	for {
		switch c.token {
		case "varset", "varadd", "parentvarset", "parentvaradd", "rootvarset", "rootvaradd":
		// Break
		case "", "[":
			if !root {
				return c.yokisinaiToken()
			}
			return nil
		case "}":
			if root {
				return c.yokisinaiToken()
			}
			return nil
		case "if", "ignorehitpause", "persistent":
			if sbl, err := c.subBlock(line, root, sbc, numVars); err != nil {
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
		case "let":
			names, err := c.varNames("=", line)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				return c.yokisinaiToken()
			}
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
				if root {
					if err := c.statementEnd(line); err != nil {
						return err
					}
				}
				c.scan(line)
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
						return c.yokisinaiToken()
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
	return c.yokisinaiToken()
}
func (c *Compiler) stateCompileZ(states map[int32]StateBytecode,
	filename, src string) error {
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
			return errmes(c.yokisinaiToken())
		}
		switch c.scan(&line) {
		case "":
			return errmes(c.yokisinaiToken())
		case "statedef":
			var err error
			if c.stateNo, err = c.scanI32(&line); err != nil {
				return errmes(err)
			}
			c.scan(&line)
			if existInThisFile[c.stateNo] {
				return errmes(Error(fmt.Sprintf("State %v overloaded", c.stateNo)))
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
					return errmes(c.yokisinaiToken())
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
				return errmes(c.yokisinaiToken())
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
func (c *Compiler) Compile(pn int, def string) (map[int32]StateBytecode,
	error) {
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
			// Read info section for the mugen version of the character
			if info {
				info = false
				sys.cgi[pn].ver = [2]uint16{}
				str, ok := is["mugenversion"]
				if ok {
					for i, s := range SplitAndTrim(str, ".") {
						if i >= len(sys.cgi[pn].ver) {
							break
						}
						if v, err := strconv.ParseUint(s, 10, 16); err == nil {
							sys.cgi[pn].ver[i] = uint16(v)
						} else {
							break
						}
					}
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
	if err := LoadFile(&cmd, def, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		str = str + sys.commonCmd
		lines, i = SplitAndTrim(str, "\n"), 0
		return nil
	}); err != nil {
		return nil, err
	}

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
						*k, *nk = CK_x, CK_nx
					case "y":
						*k, *nk = CK_y, CK_ny
					case "z":
						*k, *nk = CK_z, CK_nz
					case "a":
						*k, *nk = CK_a, CK_na
					case "b":
						*k, *nk = CK_b, CK_nb
					case "c":
						*k, *nk = CK_c, CK_nc
					case "s":
						*k, *nk = CK_s, CK_ns
					case "d":
						*k, *nk = CK_d, CK_nd
					case "w":
						*k, *nk = CK_w, CK_nw
					case "m":
						*k, *nk = CK_m, CK_nm
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
			if err := c.stateCompile(states, s, def); err != nil {
				return nil, err
			}
		}
	}
	// Compile states in command file
	if err := c.stateCompile(states, cmd, def); err != nil {
		return nil, err
	}
	// Compile states in common state file
	if len(stcommon) > 0 {
		if err := c.stateCompile(states, stcommon, def); err != nil {
			return nil, err
		}
	}
	// Compile common states from config
	for _, s := range sys.commonStates {
		if err := c.stateCompile(states, s, def); err != nil {
			return nil, err
		}
	}
	return states, nil
}
