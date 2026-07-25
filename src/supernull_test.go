package main

import "testing"

func TestSupernullCharAtValid(t *testing.T) {
	sys.supernullMode = true
	pn := 0
	sys.chars[pn] = []*Char{newChar(pn, 0)}
	ch := sys.charAt(pn)
	if ch == nil {
		t.Fatal("expected valid char")
	}
}

func TestSupernullCharAtInvalid(t *testing.T) {
	sys.supernullMode = true
	ch := sys.charAt(999)
	if ch == nil {
		t.Fatal("expected sentinel char in supernull mode")
	}
	if ch.name != "SupernullSentinel" {
		t.Fatalf("got sentinel name %q", ch.name)
	}
}

func TestSupernullEnsurePersistentGrow(t *testing.T) {
	sys.supernullMode = true
	c := newChar(0, 0)
	c.name = "Test"
	c.ss.no = 5900
	ps := []int32{0}
	out := sys.supernullEnsurePersistent(ps, 50, c)
	if len(out) <= 50 {
		t.Fatalf("expected grown slice, got len=%v", len(out))
	}
}

func TestSupernullEffectOpponentLifeZero(t *testing.T) {
	sys.supernullMode = true
	sys.chars[0] = []*Char{newChar(0, 0)}
	sys.chars[1] = []*Char{newChar(1, 0)}
	sys.chars[0][0].teamside = 0
	sys.chars[1][0].teamside = 1
	sys.chars[1][0].life = 1000
	sys.supernullApplyEffect(sys.chars[0][0], "opponent_life_zero")
	if sys.chars[1][0].life != 0 {
		t.Fatalf("expected opponent life 0, got %v", sys.chars[1][0].life)
	}
}

func TestSupernullCurFrameClamp(t *testing.T) {
	sys.supernullMode = true
	a := &Animation{
		frames:  []AnimFrame{{Time: 1}, {Time: 1}},
		curelem: 99,
	}
	f := a.curFrame()
	if f == nil {
		t.Fatal("expected frame")
	}
}

func TestVoidunsafePersistentStub(t *testing.T) {
	buf := []int32{1, 2, 3}
	if got := voidunsafePersistentGet(buf, 1); got != 2 {
		t.Fatalf("got %v", got)
	}
	if got := voidunsafePersistentGet(buf, 99); got != 0 {
		t.Fatalf("expected 0 for OOB, got %v", got)
	}
}

func TestSupernullFingerprintDedup(t *testing.T) {
	supernullFingerprints = nil
	c := newChar(0, 0)
	c.name = "FPTest"
	sys.supernullRecordFingerprint("test", c, 100, 0, "detail")
	sys.supernullRecordFingerprint("test", c, 100, 0, "detail")
	if len(supernullFingerprints) != 1 || supernullFingerprints[0].Count != 2 {
		t.Fatalf("dedup failed: %+v", supernullFingerprints)
	}
}

func TestParseVoidVersion(t *testing.T) {
	ver, f := ParseVoidVersion("1.0.0")
	if ver[0] != 1 || f < 1.0 {
		t.Fatalf("parse failed: %v %v", ver, f)
	}
}

func TestVoidClipboardEvalPercentN(t *testing.T) {
	addr, n, ok := voidClipboardEvalPercentN(`%.*d%n%d`, []interface{}{int32(96), int32(0), int32(4931646)})
	if !ok || addr != 4931646 || n <= 0 {
		t.Fatalf("eval failed: ok=%v addr=%v n=%v", ok, addr, n)
	}
	if _, _, ok := voidClipboardSplitAtPercentN(`%%n`); ok {
		t.Fatal("expected no split on escaped %n")
	}
}
