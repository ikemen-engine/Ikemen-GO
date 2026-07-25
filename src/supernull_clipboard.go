// IKEMEN:VOID WinMUGEN DisplayToClipboard %n memory-write exploit emulation.
package main

import (
	"fmt"
)

// voidClipboardFormatHasWrite returns true when the format string contains a %n write verb.
func voidClipboardFormatHasWrite(format string) bool {
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		if i+1 < len(format) && format[i+1] == '%' {
			i++
			continue
		}
		j := i + 1
		for j < len(format) && (format[j] == ' ' || format[j] == '0' ||
			format[j] == '-' || format[j] == '+' || format[j] == '#') {
			j++
		}
		for j < len(format) && format[j] >= '0' && format[j] <= '9' {
			j++
		}
		if j < len(format) && format[j] == '.' {
			j++
			for j < len(format) && format[j] >= '0' && format[j] <= '9' {
				j++
			}
		}
		if j < len(format) && format[j] == '*' {
			j++
		}
		if j < len(format) && (format[j] == 'h' || format[j] == 'l' || format[j] == 'L') {
			j++
		}
		if j < len(format) && format[j] == 'n' {
			return true
		}
	}
	return false
}

func voidClipboardParamInt(v interface{}) (int32, bool) {
	switch n := v.(type) {
	case int32:
		return n, true
	case int:
		return int32(n), true
	case int64:
		return int32(n), true
	case float32:
		return int32(n), true
	case float64:
		return int32(n), true
	default:
		return 0, false
	}
}

func voidClipboardParseExploitParams(params []interface{}) (memAddr, writeVal int32) {
	writeVal = 0
	for _, p := range params {
		v, ok := voidClipboardParamInt(p)
		if !ok {
			continue
		}
		if v > 10000 || v < -10000 {
			memAddr = v
		} else if v != 0 || writeVal == 0 {
			writeVal = v
		}
	}
	return memAddr, writeVal
}

// voidClipboardSplitAtPercentN returns the format prefix printed before the first %n verb.
func voidClipboardSplitAtPercentN(format string) (prefix string, ok bool) {
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		if i+1 >= len(format) {
			break
		}
		if format[i+1] == '%' {
			i++
			continue
		}
		j := i + 1
		for j < len(format) && (format[j] == ' ' || format[j] == '0' ||
			format[j] == '-' || format[j] == '+' || format[j] == '#') {
			j++
		}
		if j < len(format) && format[j] == '*' {
			j++
		} else {
			for j < len(format) && format[j] >= '0' && format[j] <= '9' {
				j++
			}
		}
		if j < len(format) && format[j] == '.' {
			j++
			if j < len(format) && format[j] == '*' {
				j++
			} else {
				for j < len(format) && format[j] >= '0' && format[j] <= '9' {
					j++
				}
			}
		}
		if j < len(format) && (format[j] == 'h' || format[j] == 'l' || format[j] == 'L') {
			j++
		}
		if j < len(format) && format[j] == 'n' {
			return format[:i], true
		}
	}
	return "", false
}

// voidClipboardParamsBeforeN counts printf params consumed before the %n address param.
func voidClipboardParamsBeforeN(prefix string) int {
	n := 0
	for i := 0; i < len(prefix); i++ {
		if prefix[i] != '%' {
			continue
		}
		if i+1 >= len(prefix) {
			break
		}
		if prefix[i+1] == '%' {
			i++
			continue
		}
		j := i + 1
		for j < len(prefix) && (prefix[j] == ' ' || prefix[j] == '0' ||
			prefix[j] == '-' || prefix[j] == '+' || prefix[j] == '#') {
			j++
		}
		if j < len(prefix) && prefix[j] == '*' {
			n++
			j++
		} else {
			for j < len(prefix) && prefix[j] >= '0' && prefix[j] <= '9' {
				j++
			}
		}
		if j < len(prefix) && prefix[j] == '.' {
			j++
			if j < len(prefix) && prefix[j] == '*' {
				n++
				j++
			} else {
				for j < len(prefix) && prefix[j] >= '0' && prefix[j] <= '9' {
					j++
				}
			}
		}
		if j < len(prefix) && (prefix[j] == 'h' || prefix[j] == 'l' || prefix[j] == 'L') {
			j++
		}
		if j < len(prefix) && prefix[j] != '%' {
			n++
		}
		i = j
	}
	return n
}

// voidClipboardEvalPercentN emulates WinMUGEN %n: bytes printed so far are written to the next param address.
func voidClipboardEvalPercentN(format string, params []interface{}) (memAddr, nWritten int32, ok bool) {
	prefix, ok := voidClipboardSplitAtPercentN(format)
	if !ok {
		return 0, 0, false
	}
	paramCount := voidClipboardParamsBeforeN(prefix)
	if paramCount >= len(params) {
		return 0, 0, false
	}
	if v, ok2 := voidClipboardParamInt(params[paramCount]); ok2 {
		memAddr = v
	}
	if memAddr == 0 {
		return 0, 0, false
	}
	printParams := params
	if paramCount < len(params) {
		printParams = params[:paramCount]
	}
	out := OldSprintf(prefix, printParams...)
	nWritten = int32(len(out))
	return memAddr, nWritten, true
}

// voidClipboardExploitEmulate intercepts WinMUGEN %n clipboard writes and signals a one-shot KO trigger.
// Ikemen's OldSprintf does not implement %n memory writes, so VOID must emulate them for all tiers.
func voidClipboardExploitEmulate(actor *Char, format string, params []interface{}) bool {
	if !voidRawExecutionActive(actor) && !voidExtremeExploitActive(actor) {
		return false
	}
	if actor == nil || !voidClipboardFormatHasWrite(format) {
		return false
	}
	if !voidActorUsesExploitKO(actor) {
		return false
	}
	voidSignalExploitFrame("clipboard_%n")

	memAddr, nWritten, parsed := voidClipboardEvalPercentN(format, params)
	if !parsed {
		memAddr, nWritten = voidClipboardParseExploitParams(params)
	}

	detail := fmt.Sprintf("format=%q", format)
	for i, p := range params {
		if v, ok := voidClipboardParamInt(p); ok {
			detail += fmt.Sprintf(" param[%v]=%v", i, v)
		}
	}
	if parsed {
		detail += fmt.Sprintf(" nWritten=%v", nWritten)
	}

	if memAddr != 0 {
		voidExploitShadowWrite(memAddr, nWritten)
	}

	opp := voidExploitPrimaryOpponent(actor)
	if opp == nil {
		opp = voidGlobalP2Char()
	}
	if opp == nil {
		if !voidExploitKOCommitted {
			voidExploitDebugLogOpponentLife(actor, nil, memAddr, "clipboard/%n", nWritten, false,
				"displayToClipboard", detail)
		}
		return true
	}

	if !voidExploitKOCommitted {
		if memAddr != 0 {
			opp.voidExploitWhitelistWrite(actor, memAddr, nWritten, 0, VoidVarInt, false)
		}
		voidExploitTriggerKO(actor, opp, memAddr, "displayToClipboard/%n",
			fmt.Sprintf("WinMUGEN %%n write emulated (n=%v) | %s", nWritten, detail))
	}
	return true
}
