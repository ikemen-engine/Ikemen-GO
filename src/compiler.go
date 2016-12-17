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
	cmdl     *CommandList
	maeOp    string
	usiroOp  bool
	norange  bool
	token    string
	playerNo int
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
func (c *Compiler) operator(in *string) error {
	if len(c.maeOp) > 0 {
		if opp := c.isOperator(c.token); opp <= c.isOperator(c.maeOp) {
			if opp < 0 || ((!c.usiroOp || c.token[0] != '(') &&
				(c.token[0] < 'A' || c.token[0] > 'Z') &&
				(c.token[0] < 'a' || c.token[0] > 'z')) {
				return Error(c.maeOp + "が不正です")
			}
			*in = c.token + " " + *in
			c.token = c.maeOp
			c.maeOp = ""
			c.norange = true
		}
	}
	return nil
}
func (c *Compiler) number(token string) BytecodeValue {
	f, err := strconv.ParseFloat(token, 64)
	if err != nil && f == 0 {
		return BytecodeSF()
	}
	if strings.Index(token, ".") >= 0 {
		c.usiroOp = false
		return BytecodeValue{VT_Float, f}
	}
	if strings.IndexAny(token, "Ee") >= 0 {
		return BytecodeSF()
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
			return 0, Error(string(a) + "が無効な値です")
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
			flg |= int32(AT_NA | AT_SA | AT_HA)
		case "at":
			flg |= int32(AT_NT | AT_ST | AT_HT)
		case "ap":
			flg |= int32(AT_NP | AT_SP | AT_HP)
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
			return 0, Error(a + "が無効な値です")
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
func (c *Compiler) kakkotojiru(in *string) error {
	if c.token != ")" {
		return Error(c.token + "の前に')'がありません")
	}
	c.token = c.tokenizer(in)
	return nil
}
func (c *Compiler) expValue(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	c.usiroOp, c.norange = true, false
	bv := c.number(c.token)
	if !bv.IsSF() {
		c.token = c.tokenizer(in)
		return bv, nil
	}
	if !sys.ignoreMostErrors {
		defer func() { c.usiroOp = false }()
	}
	switch c.token {
	case "time":
		out.append(OC_time)
	default:
		println(c.token)
		unimplemented()
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
			return Error(c.tokenizer(in) + "が不正です")
		}
	}
	return nil
}
func (c *Compiler) expPostNot(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expValue(out, in)
	if err != nil {
		return BytecodeSF(), err
	}
	for c.token == "!" {
		c.usiroOp = true
		if bv.IsSF() {
			out.append(OC_blnot)
		} else {
			out.blnot(&bv)
		}
		c.token = c.tokenizer(in)
	}
	if len(c.maeOp) == 0 {
		opp := c.isOperator(c.token)
		if opp == 0 {
			if !c.usiroOp && c.token == "(" {
				return BytecodeSF(), Error("演算子がありません")
			}
			oldtoken, oldin := c.token, *in
			var dummyout BytecodeExp
			if _, err := c.expValue(&dummyout, in); err != nil {
				return BytecodeSF(), err
			}
			if c.isOperator(c.token) <= 0 {
				return BytecodeSF(), Error("演算子がありません")
			}
			if err := c.renzokuEnzansihaError(in); err != nil {
				return BytecodeSF(), err
			}
			oldin = oldin[:len(oldin)-len(*in)]
			*in = oldtoken + " " + oldin[:strings.LastIndex(oldin, c.token)] + " " +
				*in
		} else if opp > 0 {
			if err := c.renzokuEnzansihaError(in); err != nil {
				return BytecodeSF(), err
			}
		}
	}
	return bv, nil
}
func (c *Compiler) expPow(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expPostNot(out, in)
	if err != nil {
		return BytecodeSF(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return BytecodeSF(), err
		}
		if c.token == "**" {
			c.token = c.tokenizer(in)
			var be BytecodeExp
			bv2, err := c.expPostNot(&be, in)
			if err != nil {
				return BytecodeSF(), err
			}
			if !bv.IsSF() && !bv2.IsSF() {
				out.pow(&bv, bv2, c.playerNo)
			} else {
				out.appendValue(bv)
				out.append(be...)
				out.appendValue(bv2)
				out.append(OC_pow)
				bv = BytecodeSF()
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
		return BytecodeSF(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return BytecodeSF(), err
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
			return BytecodeSF(), err
		}
	}
}
func (c *Compiler) expAdsb(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expMldv(out, in)
	if err != nil {
		return BytecodeSF(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return BytecodeSF(), err
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
			return BytecodeSF(), err
		}
	}
}
func (c *Compiler) expGrls(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expAdsb(out, in)
	if err != nil {
		return BytecodeSF(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return BytecodeSF(), err
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
			return BytecodeSF(), err
		}
	}
}
func (c *Compiler) expRange(out *BytecodeExp, in *string,
	bv *BytecodeValue, opc OpCode) error {
	open := c.token
	var be2, be3 BytecodeExp
	bv2, err := c.expBoolOr(&be2, in)
	if err != nil {
		return err
	}
	if c.token != "," {
		if open != "(" {
			return Error(",がありません")
		}
		if err := c.kakkotojiru(in); err != nil {
			return err
		}
		if !bv.IsSF() && !bv2.IsSF() {
			switch opc {
			case OC_eq:
				out.eq(bv, bv2)
			case OC_ne:
				out.ne(bv, bv2)
			}
		} else {
			out.appendValue(*bv)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(opc)
			*bv = BytecodeSF()
		}
		return nil
	}
	c.token = c.tokenizer(in)
	bv3, err := c.expBoolOr(&be3, in)
	close := c.token
	if close != "]" && close != ")" {
		return Error("]か)がありません")
	}
	c.token = c.tokenizer(in)
	if !bv.IsSF() && !bv2.IsSF() && !bv3.IsSF() {
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
	} else {
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
		*bv = BytecodeSF()
	}
	return nil
}
func (c *Compiler) expEqne(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	bv, err := c.expGrls(out, in)
	if err != nil {
		return BytecodeSF(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return BytecodeSF(), err
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
				if err = c.expRange(out, in, &bv, opc); err != nil {
					return BytecodeSF(), err
				}
				break
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
				return BytecodeSF(), err
			}
		}
	}
}
func (_ *Compiler) expOneOpSub(out *BytecodeExp, in *string, bv *BytecodeValue,
	ef ExpFunc, opf func(v1 *BytecodeValue, v2 BytecodeValue),
	opc OpCode) error {
	var be BytecodeExp
	bv2, err := ef(&be, in)
	if err != nil {
		return err
	}
	if !bv.IsSF() && !bv2.IsSF() {
		opf(bv, bv2)
	} else {
		out.appendValue(*bv)
		out.append(be...)
		out.appendValue(bv2)
		out.append(opc)
		*bv = BytecodeSF()
	}
	return nil
}
func (c *Compiler) expOneOp(out *BytecodeExp, in *string, ef ExpFunc,
	opt string, opf func(v1 *BytecodeValue, v2 BytecodeValue),
	opc OpCode) (BytecodeValue, error) {
	bv, err := ef(out, in)
	if err != nil {
		return BytecodeSF(), err
	}
	for {
		if err := c.operator(in); err != nil {
			return BytecodeSF(), err
		}
		if c.token == opt {
			c.token = c.tokenizer(in)
			if err := c.expOneOpSub(out, in, &bv, ef, opf, opc); err != nil {
				return BytecodeSF(), err
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
	return c.expOneOp(out, in, c.expOr, "&&", out.bland, OC_bland)
}
func (c *Compiler) expBoolXor(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	return c.expOneOp(out, in, c.expBoolAnd, "^^", out.blxor, OC_blxor)
}
func (c *Compiler) expBoolOr(out *BytecodeExp, in *string) (BytecodeValue,
	error) {
	defer func(omp string) { c.maeOp = omp }(c.maeOp)
	return c.expOneOp(out, in, c.expBoolXor, "||", out.blor, OC_blor)
}
func (c *Compiler) typedExp(ef ExpFunc, in *string,
	vt ValueType) (BytecodeExp, error) {
	c.token = c.tokenizer(in)
	var be BytecodeExp
	bv, err := ef(&be, in)
	if err != nil {
		return nil, err
	}
	if !bv.IsSF() {
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
	if len(c.token) > 0 && c.token != "," {
		return nil, Error(c.token + "が不正です")
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
		return nil, Error(c.token + "が不正です")
	}
	return be, nil
}
func (c *Compiler) parseSection(lines []string, i *int,
	sctrl func(name, data string) error) (IniSection, error) {
	is := NewIniSection()
	_type, persistent, ignorehitpause := true, true, true
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
			return err
		}
		bes = append(bes, be)
		if c.token != "," {
			break
		}
	}
	sc.add(id, bes)
	return nil
}
func (c *Compiler) stateDef(is IniSection, sbc *StateBytecode) error {
	return c.stateSec(is, func() error {
		sc := newStateControllerBase(c.playerNo)
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
		if err := c.stateParam(is, "sprpriority", func(data string) error {
			return c.scAdd(sc, stateDef_sprpriority, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "facep2", func(data string) error {
			return c.scAdd(sc, stateDef_facep2, data, VT_Bool, 1)
		}); err != nil {
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
		if err := c.stateParam(is, "velset", func(data string) error {
			return c.scAdd(sc, stateDef_velset, data, VT_Float, 3)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", func(data string) error {
			return c.scAdd(sc, stateDef_anim, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "ctrl", func(data string) error {
			return c.scAdd(sc, stateDef_ctrl, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "poweradd", func(data string) error {
			return c.scAdd(sc, stateDef_poweradd, data, VT_Int, 1)
		}); err != nil {
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
		return Error("valueが指定されていません")
	}
	if err = c.stateParam(is, "time", func(data string) error {
		return c.scAdd(sc, hitBy_time, data, VT_Int, 1)
	}); err != nil {
		return err
	}
	if two {
		sc.add(hitBy_value2, sc.iToExp(attr))
	} else {
		sc.add(hitBy_value, sc.iToExp(attr))
	}
	return nil
}
func (c *Compiler) hitBy(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return hitBy(*sc), c.stateSec(is, func() error {
		return c.hitBySub(is, sc)
	})
}
func (c *Compiler) notHitBy(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return notHitBy(*sc), c.stateSec(is, func() error {
		return c.hitBySub(is, sc)
	})
}
func (c *Compiler) assertSpecial(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return assertSpecial(*sc), c.stateSec(is, func() error {
		unimplemented()
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
		for i++; i < len(lines); i++ {
			line := strings.ToLower(strings.TrimSpace(
				strings.SplitN(lines[i], ";", 2)[0]))
			if len(line) < 7 || line[0] != '[' || line[len(line)-1] != ']' ||
				line[1:7] != "state " {
				break
			}
			i++
			sc := newStateControllerBase(c.playerNo)
			var scf func(is IniSection, sbc *StateBytecode,
				sc *StateControllerBase) (StateController, error)
			var triggerall []BytecodeExp
			allUtikiri := false
			var trigger [][]BytecodeExp
			var trexist []int8
			is, err := c.parseSection(lines, &i, func(name, data string) error {
				switch name {
				case "type":
					switch data {
					case "hitby":
						scf = c.hitBy
					case "nothitby":
						scf = c.notHitBy
					case "assertspecial":
						scf = c.assertSpecial
					default:
						println(data)
						unimplemented()
					}
				case "persistent":
					if n >= 0 {
						sc.persistent = Atoi(data)
						if sc.persistent > 128 {
							sc.persistent = 1
						} else if sc.persistent <= 0 {
							sc.persistent = math.MaxInt32
						}
					}
				case "ignorehitpause":
					sc.ignorehitpause = Atoi(data) != 0
				case "triggerall":
					be, err := c.fullExpression(&data, VT_Bool)
					if err != nil {
						return err
					}
					if len(be) == 2 && be[0] == OC_int8 {
						if be[1] == 0 {
							allUtikiri = true
						}
					} else if !allUtikiri {
						triggerall = append(triggerall, be)
					}
				default:
					tn, ok := readDigit(name[7:])
					if !ok || tn <= 0 || tn > 65536 {
						errmes(Error("トリガー名 (" + name + ") が不正です"))
					}
					if len(trigger) < int(tn) {
						trigger = append(trigger, make([][]BytecodeExp,
							int(tn)-len(trigger))...)
						trexist = append(trexist, make([]int8, int(tn)-len(trexist))...)
					}
					tn--
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
					if len(be) == 2 && be[0] == OC_int8 {
						if be[1] == 0 {
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
			if scf == nil {
				return errmes(Error("typeが指定されていません"))
			}
			if len(trexist) == 0 || trexist[0] == 0 {
				return errmes(Error("trigger1がありません"))
			}
			var texp BytecodeExp
			for i, e := range triggerall {
				if len(e) > 0 {
					texp.append(e...)
					if i < len(triggerall)-1 {
						texp.append(OC_jz8, 0)
						texp.append(OC_pop)
					}
				}
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
						te.appendValue(BytecodeBool(false))
					}
					oldlen := len(te)
					for j := len(tr) - 1; j >= 0; j-- {
						tmp := tr[j]
						if len(tmp) > 0 {
							if j < len(tr)-1 {
								if len(te) > int(math.MaxUint8-1) {
									tmp.appendJmp(OC_jz, int32(len(te)+1))
									tmp.append(OC_pop)
								} else {
									tmp.append(OC_jz8, OpCode(len(te)+1))
									tmp.append(OC_pop)
								}
							}
							te = append(tmp, te...)
						}
					}
					if len(te) == oldlen {
						te = nil
					}
					if len(te) == 0 {
						if trexist[i] > 0 {
							if len(texp) > 0 {
								texp.appendValue(BytecodeBool(true))
							}
							break
						}
					} else {
						texp.append(te...)
						if i < len(trigger)-1 {
							texp.append(OC_jnz8, 0)
							texp.append(OC_pop)
						}
					}
				}
			}
			if len(texp) > 0 {
				sc.add(SCID_trigger, sc.beToExp(texp))
			}
			sctrl, err := scf(is, sbc, sc)
			if err != nil {
				return errmes(err)
			}
			appending := true
			if len(texp) > 0 {
			} else if allUtikiri {
				appending = false
			} else {
				appending = false
				for _, te := range trexist {
					if te >= 0 {
						if te > 0 {
							appending = true
						}
						break
					}
				}
			}
			if appending {
				sbc.ctrls = append(sbc.ctrls, sctrl)
			}
		}
		_, ok := bc.states[n]
		if !ok {
			bc.states[n] = *sbc
		}
	}
	return nil
}
func (c *Compiler) Compile(n int, def string) (*Bytecode, error) {
	c.playerNo = n
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
