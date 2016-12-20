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
			return 0, Error(istr + "が整数でありません")
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
	return nil
}
func (c *Compiler) kyuushiki(in *string) (not bool, err error) {
	for {
		if c.token == "!=" {
			not = true
			break
		} else if len(c.token) > 0 {
			if c.token[len(c.token)-1] == '=' {
				break
			}
		} else {
			return false, Error("'='か'!='がありません")
		}
		c.token = c.tokenizer(in)
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
		err = Error("'['か'('がありません")
		return
	}
	integer := func() (int32, error) {
		c.token = c.tokenizer(in)
		minus := false
		for c.token == "-" || c.token == "+" {
			minus = minus || c.token == "-"
			c.token = c.tokenizer(in)
		}
		if len(c.token) == 0 || c.token[0] < '0' || c.token[0] > '9' {
			return 0, Error("数字の読み込みエラーです")
		}
		i := Atoi(c.token)
		if minus {
			i *= -1
		}
		return i, nil
	}
	if min, err = integer(); err != nil {
		return
	}
	i := strings.Index(*in, ",")
	if i < 0 {
		err = Error("','がありません")
		return
	}
	*in = (*in)[i+1:]
	if max, err = integer(); err != nil {
		return
	}
	i = strings.IndexAny(*in, "])")
	if i < 0 {
		err = Error("']'か')'がありません")
		return
	}
	if (*in)[i] == ')' {
		maxop = OC_lt
	} else {
		maxop = OC_le
	}
	*in = (*in)[i+1:]
	c.token = c.tokenizer(in)
	return
}
func (c *Compiler) kyuushikiThroughNeo(_range bool, in *string) {
	i := 0
	for ; i < len(*in); i++ {
		if (*in)[i] >= '0' && (*in)[i] <= '9' || (*in)[i] == '-' ||
			_range && ((*in)[i] == '[' || (*in)[i] == '(') {
			break
		}
	}
	*in = (*in)[i:]
	c.token = c.tokenizer(in)
}
func (c *Compiler) kyuushikiSuperDX(out *BytecodeExp, in *string,
	hissu bool) error {
	comma := c.token == ","
	if comma {
		c.token = c.tokenizer(in)
	}
	var opc OpCode
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
		hikaku := false
		switch c.token {
		case "!=":
			opc = OC_ne
			hikaku = true
		case "=":
			hikaku = true
		default:
			if hissu && !comma {
				return Error("比較演算子がありません")
			}
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
	n, err := c.integer2(in)
	if err != nil {
		return err
	}
	out.appendValue(BytecodeInt(n))
	out.append(opc)
	c.usiroOp = true
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
	var be1, be2, be3 BytecodeExp
	var bv1, bv2, bv3 BytecodeValue
	var n int32
	var be BytecodeExp
	var err error
	switch c.token {
	case "-":
		if len(*in) > 0 && (((*in)[0] >= '0' && (*in)[0] <= '9') || (*in)[0] == '.') {
			c.token += c.tokenizer(in)
			bv = c.number(c.token)
			if bv.IsSF() {
				return BytecodeSF(), Error(c.token + "が不正です")
			}
		} else {
			c.token = c.tokenizer(in)
			bv, err = c.expValue(out, in)
			if bv.IsSF() {
				out.append(OC_neg)
			} else {
				out.neg(&bv)
			}
		}
	case "~":
		c.token = c.tokenizer(in)
		bv, err = c.expValue(out, in)
		if bv.IsSF() {
			out.append(OC_not)
		} else {
			out.not(&bv)
		}
	case "!":
		c.token = c.tokenizer(in)
		bv, err = c.expValue(out, in)
		if bv.IsSF() {
			out.append(OC_blnot)
		} else {
			out.blnot(&bv)
		}
	case "time":
		out.append(OC_time)
	case "alive":
		out.append(OC_alive)
	case "ifelse":
		if c.tokenizer(in) != "(" {
			return BytecodeSF(), Error(c.token + "の次に'('がありません")
		}
		c.token = c.tokenizer(in)
		if bv1, err = c.expBoolOr(&be1, in); err != nil {
			return BytecodeSF(), err
		}
		if c.token != "," {
			return BytecodeSF(), Error("','がありません")
		}
		c.token = c.tokenizer(in)
		if bv2, err = c.expBoolOr(&be2, in); err != nil {
			return BytecodeSF(), err
		}
		if c.token != "," {
			return BytecodeSF(), Error("','がありません")
		}
		c.token = c.tokenizer(in)
		if bv3, err = c.expBoolOr(&be3, in); err != nil {
			return BytecodeSF(), err
		}
		if err := c.kakkotojiru(in); err != nil {
			return BytecodeSF(), err
		}
		if bv1.IsSF() || bv2.IsSF() || bv3.IsSF() {
			out.append(be1...)
			out.appendValue(bv1)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(be3...)
			out.appendValue(bv3)
			out.append(OC_ifelse)
		} else {
			if bv1.ToB() {
				bv = bv2
			} else {
				bv = bv3
			}
		}
	case "random":
		out.append(OC_random)
	case "roundstate":
		out.append(OC_roundstate)
	case "matchover":
		out.append(OC_ex_, OC_ex_matchover)
	case "anim":
		out.append(OC_anim)
	case "animtime":
		out.append(OC_animtime)
	case "animelem":
		if _, err = c.kyuushiki(in); err != nil {
			return BytecodeSF(), err
		}
		if c.token == "-" {
			return BytecodeSF(), Error("マイナスが付くとエラーです")
		}
		if n, err = c.integer2(in); err != nil {
			return BytecodeSF(), err
		}
		if n <= 0 {
			return BytecodeSF(), Error("animelemのは0より大きくなければいけません")
		}
		out.appendValue(BytecodeInt(n))
		out.append(OC_animelemtime)
		if err = c.kyuushikiSuperDX(&be, in, false); err != nil {
			return BytecodeSF(), err
		}
		out.append(OC_jsf8, OpCode(len(be)))
		out.append(be...)
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
			if bv.IsSF() || bv2.IsSF() {
				out.appendValue(bv)
				out.append(be...)
				out.appendValue(bv2)
				out.append(OC_pow)
				bv = BytecodeSF()
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
		c.token = c.tokenizer(in)
		if bv.IsSF() || bv2.IsSF() {
			out.appendValue(*bv)
			out.append(be2...)
			out.appendValue(bv2)
			out.append(opc)
			*bv = BytecodeSF()
		} else {
			switch opc {
			case OC_eq:
				out.eq(bv, bv2)
			case OC_ne:
				out.ne(bv, bv2)
			}
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
	if bv.IsSF() || bv2.IsSF() || bv3.IsSF() {
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
	if bv.IsSF() || bv2.IsSF() {
		out.appendValue(*bv)
		out.append(be...)
		out.appendValue(bv2)
		out.append(opc)
		*bv = BytecodeSF()
	} else {
		opf(bv, bv2)
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
	data string, vt ValueType, numArg int, topbe ...BytecodeExp) error {
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
		return Error(paramname + "が指定されていません")
	}
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
		if err := c.paramValue(is, sc, "anim",
			stateDef_anim, VT_Int, 1, false); err != nil {
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
		return Error("valueが指定されていません")
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
		foo := func(data string) error {
			switch data {
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
			default:
				return Error(data + "が無効な値です")
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
			return Error("flagが指定されていません")
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
}
func (c *Compiler) playSnd(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return playSnd(*sc), c.stateSec(is, func() error {
		f := false
		if err := c.stateParam(is, "value", func(data string) error {
			f = true
			fflg := false
			if len(data) > 0 {
				switch data[0] {
				case 'F', 'f':
					fflg = true
					data = data[1:]
				case 'S', 's':
					data = data[1:]
				}
			}
			return c.scAdd(sc, playSnd_value, data, VT_Int, 2,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if !f {
			return Error("valueが指定されていません")
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
		var volname string
		if sys.cgi[c.playerNo].ver[0] == 1 {
			volname = "volumescale"
		} else {
			volname = "volume"
		}
		if err := c.paramValue(is, sc, volname,
			playSnd_volume, VT_Int, 1, false); err != nil {
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
}
func (c *Compiler) changeStateSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "value",
		changeState_value, VT_Int, 1, true); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ctrl",
		changeState_ctrl, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "anim",
		changeState_anim, VT_Int, 1, false); err != nil {
		return err
	}
	return nil
}
func (c *Compiler) changeState(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return changeState(*sc), c.stateSec(is, func() error {
		return c.changeStateSub(is, sc)
	})
}
func (c *Compiler) selfState(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return selfState(*sc), c.stateSec(is, func() error {
		return c.changeStateSub(is, sc)
	})
}
func (c *Compiler) tagIn(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return tagIn(*sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "stateno",
			tagIn_stateno, VT_Int, 1, true); err != nil {
			return err
		}
		f := false
		if err := c.stateParam(is, "partnerstateno", func(data string) error {
			f = true
			return c.scAdd(sc, tagIn_partnerstateno, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !f {
			sc.add(tagIn_partnerstateno, sc.iToExp(-1))
		}
		return nil
	})
}
func (c *Compiler) tagOut(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return tagOut(*sc), c.stateSec(is, func() error {
		sc.add(tagOut_, nil)
		return nil
	})
}
func (c *Compiler) destroySelf(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return destroySelf(*sc), c.stateSec(is, func() error {
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
}
func (c *Compiler) changeAnimSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "elem",
		changeAnim_elem, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "value",
		changeAnim_value, VT_Int, 1, true); err != nil {
		return err
	}
	return nil
}
func (c *Compiler) changeAnim(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return changeAnim(*sc), c.stateSec(is, func() error {
		return c.changeAnimSub(is, sc)
	})
}
func (c *Compiler) changeAnim2(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return changeAnim2(*sc), c.stateSec(is, func() error {
		return c.changeAnimSub(is, sc)
	})
}
func (c *Compiler) helper(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return helper(*sc), c.stateSec(is, func() error {
		if err := c.stateParam(is, "helpertype", func(data string) error {
			if len(data) == 0 {
				return Error("値が指定されていません")
			}
			switch strings.ToLower(data)[0] {
			case 'n':
			case 'p':
				sc.add(helper_helpertype, sc.iToExp(1))
			default:
				return Error(data + "が無効な値です")
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "name", func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("\"で囲まれていません")
			}
			sc.add(helper_name, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "postype", func(data string) error {
			if len(data) == 0 {
				return Error("値が指定されていません")
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
					return Error(data + "が無効な値です")
				}
			}
			sc.add(helper_postype, sc.iToExp(int32(pt)))
			return nil
		}); err != nil {
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
		if err := c.paramValue(is, sc, "keyctrl",
			helper_keyctrl, VT_Bool, 1, false); err != nil {
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
		return nil
	})
}
func (c *Compiler) powerAdd(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return powerAdd(*sc), c.stateSec(is, func() error {
		return c.paramValue(is, sc, "value", powerAdd_value, VT_Int, 1, true)
	})
}
func (c *Compiler) ctrlSet(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return ctrlSet(*sc), c.stateSec(is, func() error {
		return c.paramValue(is, sc, "value", ctrlSet_value, VT_Bool, 1, true)
	})
}
func (c *Compiler) hitDefSub(is IniSection,
	sc *StateControllerBase) error {
	unimplemented()
	return nil
}
func (c *Compiler) hitDef(is IniSection, sbc *StateBytecode,
	sc *StateControllerBase) (StateController, error) {
	return hitDef(*sc), c.stateSec(is, func() error {
		return c.hitDefSub(is, sc)
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
				i--
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
					case "playsnd":
						scf = c.playSnd
					case "changestate":
						scf = c.changeState
					case "selfstate":
						scf = c.selfState
					case "tagin":
						scf = c.tagIn
					case "tagout":
						scf = c.tagOut
					case "destroyself":
						scf = c.destroySelf
					case "changeanim":
						scf = c.changeAnim
					case "changeanim2":
						scf = c.changeAnim2
					case "helper":
						scf = c.helper
					case "poweradd":
						scf = c.powerAdd
					case "ctrlset":
						scf = c.ctrlSet
					case "hitdef":
						scf = c.hitDef
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
					if !ok || tn < 1 || tn > 65536 {
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
				texp.append(e...)
				if i < len(triggerall)-1 {
					texp.append(OC_jz8, 0)
					texp.append(OC_pop)
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
