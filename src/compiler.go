package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const kuuhaktokigou = " !=<>()|&+-*/%,[]^|:\"\t\r\n"

type ExpFunc func(out *BytecodeExp, in *string) (BytecodeValue, error)
type Compiler struct {
	cmdl    *CommandList
	valCnt  int
	maeOp   string
	usiroOp bool
	norange bool
	token   string
}

func newCompiler() *Compiler {
	return &Compiler{}
}
func (_ *Compiler) tokenizer(in *string) string {
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
func (_ *Compiler) isOperator(token string) int {
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
func (c *Compiler) operator(in *string) (string, error) {
	if len(c.maeOp) > 0 {
		if opp := c.isOperator(c.token); opp <= c.isOperator(c.maeOp) {
			if opp < 0 || ((!c.usiroOp || c.token[0] != '(') &&
				(c.token[0] < 'A' || c.token[0] > 'Z') &&
				(c.token[0] < 'a' || c.token[0] > 'z')) {
				return "", Error(c.maeOp + "が不正です")
			}
			*in = c.token + " " + *in
			c.token = c.maeOp
			c.maeOp = ""
			c.norange = true
		}
	}
	return c.token, nil
}
func (c *Compiler) number(token string) BytecodeValue {
	f, err := strconv.ParseFloat(token, 64)
	if err != nil && f == 0 {
		return BytecodeNaN()
	}
	if strings.Index(token, ".") >= 0 {
		c.usiroOp = false
		return BytecodeValue{VT_Float, f}
	}
	if strings.IndexAny(token, "Ee") >= 0 {
		return BytecodeNaN()
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
func (c *Compiler) expValue(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	c.usiroOp, c.norange = true, false
	bv := c.number(c.token)
	if !bv.IsNaN() {
		c.valCnt++
		c.token = c.tokenizer(in)
		return bv, nil
	}
	if !sys.ignoreMostErrors {
		defer func() { c.usiroOp = false }()
	}
	unimplemented()
	c.valCnt++
	c.token = c.tokenizer(in)
	return bv, nil
}
func (c *Compiler) renzikuEnzansihaError(in *string) error {
	*in = strings.TrimSpace(*in)
	if len(*in) > 0 {
		switch (*in)[0] {
		default:
			if len(*in) < 2 || (*in)[:2] != "!=" {
				break
			}
			fallthrough
		case '=', '<', '>', '|', '&', '+', '*', '/', '%', '^':
			return Error(c.tokenizer(in) + "が不正です")
		}
	}
	return nil
}
func (c *Compiler) expPostNot(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expValue(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	for c.token == "!" {
		c.usiroOp = true
		if bv.IsNaN() {
			out.append(OC_blnot)
		} else {
			bv.t = VT_Bool
			if bv.v != 0 {
				bv.v = 0
			} else {
				bv.v = 1
			}
		}
		c.token = c.tokenizer(in)
	}

	if len(c.maeOp) == 0 {
		opp := c.isOperator(c.token)
		if opp == 0 {
			if !c.usiroOp && c.token == "(" {
				return BytecodeNaN(), Error("演算子がありません")
			}
			oldin := *in
			var dummyout BytecodeExp
			if _, err := c.expValue(&dummyout, in); err != nil {
				return BytecodeNaN(), err
			}
			if c.isOperator(c.token) <= 0 {
				return BytecodeNaN(), Error("演算子がありません")
			}
			if err := c.renzikuEnzansihaError(in); err != nil {
				return BytecodeNaN(), err
			}
			oldin = oldin[:len(oldin)-len(*in)]
			*in = "(" + oldin[:strings.LastIndex(oldin, c.token)] + *in
		} else if opp > 0 {
			if err := c.renzikuEnzansihaError(in); err != nil {
				return BytecodeNaN(), err
			}
		}
	}
	return bv, nil
}
func (c *Compiler) expPow(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expPostNot(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expMldv(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expPow(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expAdsb(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expMldv(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expGrls(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expAdsb(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expEqu(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expGrls(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expAnd(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expEqu(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expXor(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expAnd(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expOr(out *BytecodeExp, in *string) (BytecodeValue, error) {
	bv, err := c.expXor(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expBoolAnd(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expOr(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expBoolXor(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expBoolAnd(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) expBoolOr(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	defer func(ovc int, omp string) {
		c.valCnt, c.maeOp = ovc, omp
	}(c.valCnt, c.maeOp)
	bv, err := c.expBoolXor(out, in)
	if err != nil {
		return BytecodeNaN(), err
	}
	unimplemented()
	return bv, nil
}
func (c *Compiler) typedExp(ef ExpFunc, out *BytecodeExp, in *string,
	vt ValueType) (float64, error) {
	c.token = c.tokenizer(in)
	var be BytecodeExp
	bv, err := ef(&be, in)
	if err != nil {
		return 0, err
	}
	if len(be) == 0 && vt != VT_Variant {
		if vt == VT_Bool {
			if bv.v != 0 {
				bv.v = 1
			} else {
				bv.v = 0
			}
		}
		return bv.v, nil
	}
	out.append(be...)
	out.AppendValue(bv)
	return math.NaN(), nil
}
func (c *Compiler) argExpression(in *string,
	vt ValueType) (BytecodeExp, float64, error) {
	var be BytecodeExp
	v, err := c.typedExp(c.expBoolOr, &be, in, vt)
	if err != nil {
		return nil, 0, err
	}
	if len(c.token) > 0 && c.token != "," {
		return nil, 0, Error(c.token + "が不正です")
	}
	return be, v, nil
}
func (c *Compiler) fullExpression(in *string,
	vt ValueType) (BytecodeExp, float64, error) {
	var be BytecodeExp
	v, err := c.typedExp(c.expBoolOr, &be, in, vt)
	if err != nil {
		return nil, 0, err
	}
	if len(c.token) > 0 {
		return nil, 0, Error(c.token + "が不正です")
	}
	return be, v, nil
}
func (c *Compiler) parseSection(lines []string, i *int,
	sctrl func(name, data string) error) (IniSection, error) {
	is := NewIniSection()
	for ; *i < len(lines); (*i)++ {
		line := strings.ToLower(strings.TrimSpace(
			strings.SplitN(lines[*i], ";", 2)[0]))
		if len(line) > 0 && line[0] == '[' {
			(*i)--
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
				return nil, Error(name + "が重複しています")
			}
			if sctrl != nil {
				switch name {
				case "type", "persistent", "ignorehitpause":
				default:
					if len(name) < 7 || name[:7] != "trigger" {
						is[name] = data
						continue
					}
				}
				if err := sctrl(name, data); err != nil {
					return nil, err
				}
			} else {
				is[name] = data
			}
		}
	}
	return is, nil
}
func (c *Compiler) stateSec(is IniSection, f func() error) error {
	if err := f(); err != nil {
		return err
	}
	if !sys.ignoreMostErrors {
		var str string
		for k, _ := range is {
			if len(str) > 0 {
				str += ", "
			}
			str += k
		}
		if len(str) > 0 {
			return Error(str + "は無効なキー名です")
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
func (c *Compiler) scAdd(sc *StateControllerBase, id byte,
	data string, vt ValueType, numArg int) error {
	bes, vs := []BytecodeExp{}, []float64{}
	for n := 1; n <= numArg; n++ {
		var be BytecodeExp
		var v float64
		var err error
		if n < numArg {
			be, v, err = c.argExpression(&data, vt)
		} else {
			be, v, err = c.fullExpression(&data, vt)
		}
		if err != nil {
			return err
		}
		bes = append(bes, be)
		vs = append(vs, v)
		if n < numArg && c.token != "," {
			break
		}
	}
	cns := true
	for i, v := range vs {
		if math.IsNaN(v) {
			cns = false
		} else {
			bes[i].AppendValue(BytecodeValue{vt, v})
		}
	}
	if cns {
		if vt == VT_Float {
			floats := make([]float32, len(vs))
			for i := range floats {
				floats[i] = float32(vs[i])
			}
			sc.add(id+SCID_const, sc.fToExp(floats...))
		} else {
			ints := make([]int32, len(vs))
			for i := range ints {
				ints[i] = int32(vs[i])
			}
			sc.add(id+SCID_const, sc.iToExp(ints...))
		}
	} else {
		sc.add(id, bes)
	}
	return nil
}
func (c *Compiler) stateDef(is IniSection, sbc *StateBytecode) error {
	return c.stateSec(is, func() error {
		var sc StateControllerBase
		if err := c.stateParam(is, "type", func(data string) error {
			if len(data) == 0 {
				return Error("値が指定されていません")
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
				return Error(data + "が無効な値です")
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "movetype", func(data string) error {
			if len(data) == 0 {
				return Error("値が指定されていません")
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
				return Error(data + "が無効な値です")
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "physics", func(data string) error {
			if len(data) == 0 {
				return Error("値が指定されていません")
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
				return Error(data + "が無効な値です")
			}
			return nil
		}); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "hitcountpersist", func(data string) error {
			b = true
			be, v, err := c.fullExpression(&data, VT_Bool)
			if err != nil {
				return err
			}
			if math.IsNaN(v) {
				sc.add(stateDef_hitcountpersist, sc.beToExp(be))
			} else if v == 0 { // falseのときだけクリアする
				sc.add(stateDef_hitcountpersist_c, nil)
			}
			return nil
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_hitcountpersist_c, nil)
		}
		b = false
		if err := c.stateParam(is, "movehitpersist", func(data string) error {
			b = true
			be, v, err := c.fullExpression(&data, VT_Bool)
			if err != nil {
				return err
			}
			if math.IsNaN(v) {
				sc.add(stateDef_movehitpersist, sc.beToExp(be))
			} else if v == 0 { // falseのときだけクリアする
				sc.add(stateDef_movehitpersist_c, nil)
			}
			return nil
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_movehitpersist_c, nil)
		}
		b = false
		if err := c.stateParam(is, "hitdefpersist", func(data string) error {
			b = true
			be, v, err := c.fullExpression(&data, VT_Bool)
			if err != nil {
				return err
			}
			if math.IsNaN(v) {
				sc.add(stateDef_hitdefpersist, sc.beToExp(be))
			} else if v == 0 { // falseのときだけクリアする
				sc.add(stateDef_hitdefpersist_c, nil)
			}
			return nil
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_hitdefpersist_c, nil)
		}
		if err := c.stateParam(is, "sprpriority", func(data string) error {
			return c.scAdd(&sc, stateDef_sprpriority, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "facep2", func(data string) error {
			be, v, err := c.fullExpression(&data, VT_Bool)
			if err != nil {
				return err
			}
			if math.IsNaN(v) {
				sc.add(stateDef_facep2, sc.beToExp(be))
			} else if v != 0 {
				sc.add(stateDef_facep2_c, nil)
			}
			return nil
		}); err != nil {
			return err
		}
		b = false
		if err := c.stateParam(is, "juggle", func(data string) error {
			b = true
			return c.scAdd(&sc, stateDef_juggle, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(stateDef_juggle_c, sc.iToExp(0))
		}
		if err := c.stateParam(is, "velset", func(data string) error {
			return c.scAdd(&sc, stateDef_velset, data, VT_Float, 3)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", func(data string) error {
			return c.scAdd(&sc, stateDef_anim, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "ctrl", func(data string) error {
			return c.scAdd(&sc, stateDef_ctrl, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "poweradd", func(data string) error {
			return c.scAdd(&sc, stateDef_poweradd, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		sbc.stateDef = stateDef(sc)
		return nil
	})
}
func (c *Compiler) stateCompile(bc *Bytecode, filename, def string) error {
	var lines []string
	if err := LoadFile(&filename, def, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines = SplitAndTrim(str, "\n")
		return nil
	}); err != nil {
		return err
	}
	i := 0
	errmes := func(err error) error {
		return Error(fmt.Sprintf("%v:%v:\n%v", filename, i+1, err.Error()))
	}
	existInThisFile := make(map[int32]bool)
	for ; i < len(lines); i++ {
		line := strings.ToLower(strings.TrimSpace(
			strings.SplitN(lines[i], ";", 2)[0]))
		if len(line) < 11 || line[0] != '[' || line[len(line)-1] != ']' ||
			line[1:10] != "statedef " {
			continue
		}
		n := Atoi(line[11:])
		if existInThisFile[n] {
			continue
		}
		existInThisFile[n] = true
		i++
		is, err := c.parseSection(lines, &i, nil)
		if err != nil {
			return errmes(err)
		}
		sbc := newStateBytecode()
		if err := c.stateDef(is, sbc); err != nil {
			return errmes(err)
		}
		unimplemented()
	}
	return nil
}
func (c *Compiler) Compile(n int, def string) (*Bytecode, error) {
	bc := newBytecode()
	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	lines, i, cmd, stcommon := SplitAndTrim(str, "\n"), 0, "", ""
	var st [11]string
	info, files := true, true
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				var v0, v1 int32 = 0, 0
				is.ReadI32("mugenversion", &v0, &v1)
				sys.cgi[n].ver = [2]int16{I32ToI16(v0), I32ToI16(v1)}
			}
		case "files":
			if files {
				files = false
				cmd, stcommon = is["cmd"], is["stcommon"]
				st[0] = is["st"]
				for i := 1; i < len(st); i++ {
					st[i] = is[fmt.Sprintf("st%d", i-1)]
				}
			}
		}
	}
	if err := LoadFile(&cmd, def, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines, i = SplitAndTrim(str, "\n"), 0
		return nil
	}); err != nil {
		return nil, err
	}
	if sys.chars[n][0].cmd == nil {
		sys.chars[n][0].cmd = make([]CommandList, MaxSimul*2)
		b := newCommandBuffer()
		for i := range sys.chars[n][0].cmd {
			sys.chars[n][0].cmd[i] = *NewCommandList(b)
		}
	}
	c.cmdl = &sys.chars[n][0].cmd[n]
	remap, defaults, ckr := true, true, NewCommandKeyRemap()
	var cmds []IniSection
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "remap":
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
					}
				}
				rm("x", &ckr.x, &ckr.nx)
				rm("y", &ckr.y, &ckr.ny)
				rm("z", &ckr.z, &ckr.nz)
				rm("a", &ckr.a, &ckr.na)
				rm("b", &ckr.b, &ckr.nb)
				rm("c", &ckr.c, &ckr.nc)
				rm("s", &ckr.s, &ckr.ns)
			}
		case "defaults":
			if defaults {
				defaults = false
				is.ReadI32("command.time", &c.cmdl.DefaultTime)
				var i32 int32
				if is.ReadI32("command.buffer.time", &i32) {
					c.cmdl.DefaultBufferTime = Max(1, i32)
				}
			}
		default:
			if len(name) >= 7 && name[:7] == "command" {
				cmds = append(cmds, is)
			}
		}
	}
	for _, is := range cmds {
		cm, err := ReadCommand(is["name"], is["command"], ckr)
		if err != nil {
			return nil, Error(cmd + ":\nname = " + is["name"] +
				"\ncommand = " + is["command"] + "\n" + err.Error())
		}
		is.ReadI32("time", &cm.time)
		var i32 int32
		if is.ReadI32("buffer.time", &i32) {
			cm.buftime = Max(1, i32)
		}
		c.cmdl.Add(*cm)
	}
	for _, s := range st {
		if len(s) > 0 {
			if err := c.stateCompile(bc, s, def); err != nil {
				return nil, err
			}
		}
	}
	if err := c.stateCompile(bc, cmd, def); err != nil {
		return nil, err
	}
	if len(stcommon) > 0 {
		if err := c.stateCompile(bc, stcommon, def); err != nil {
			return nil, err
		}
	}
	return bc, nil
}
