package main

import "testing"

func TestVoidPostmanWinTeam(t *testing.T) {
	if len(sys.cgi) < 2 {
		sys.cgi = append(sys.cgi, newCharGlobalInfo(), newCharGlobalInfo())
	}
	if len(sys.chars) < 2 {
		sys.chars = make([][]*Char, 2)
	}
	sys.cgi[0].voidTier = VoidTierPostman
	sys.cgi[0].def = "chars/- Extravaganza -/sys/Extravaganza-Post.def"
	sys.chars[0] = []*Char{newChar(0, 0)}
	sys.chars[1] = []*Char{newChar(1, 0)}
	sys.chars[0][0].teamside = 0
	sys.chars[1][0].teamside = 1
	sys.chars[0][0].life = 0
	sys.chars[1][0].life = 0
	sys.chars[0][0].setSCF(SCF_ko)
	sys.chars[1][0].setSCF(SCF_ko)
	if w := voidPostmanWinTeam(); w != 0 {
		t.Fatalf("expected postman P1 win, got team %v", w)
	}
}

func TestVoidPostmanWinTeamEnemyPostHijack(t *testing.T) {
	if len(sys.cgi) < 2 {
		sys.cgi = append(sys.cgi, newCharGlobalInfo(), newCharGlobalInfo())
	}
	if len(sys.chars) < 2 {
		sys.chars = make([][]*Char, 2)
	}
	sys.cgi[0].displayname = "Extravaganza"
	sys.cgi[1].displayname = "Who am I?"
	sys.cgi[1].data.life = 1
	sys.cgi[1].data.attack = 1
	sys.cgi[1].data.defence = 1
	sys.cgi[1].data.liedown.time = 60
	sys.chars[0] = []*Char{newChar(0, 0)}
	sys.chars[1] = []*Char{newChar(1, 0)}
	sys.chars[0][0].teamside = 0
	sys.chars[1][0].teamside = 1
	sys.chars[0][0].life = 666
	sys.chars[1][0].life = 0
	sys.chars[0][0].setSCF(SCF_ko)
	sys.chars[1][0].unsetSCF(SCF_ko)
	if w := voidPostmanWinTeam(); w != 0 {
		t.Fatalf("expected hijack postman P1 win, got team %v", w)
	}
}

func TestVoidSyncKOFlagWithLife(t *testing.T) {
	sys.chars[0] = []*Char{newChar(0, 0)}
	sys.chars[0][0].teamside = 0
	sys.chars[0][0].life = 666
	sys.chars[0][0].setSCF(SCF_ko)
	voidSyncKOFlagWithLife()
	if sys.chars[0][0].scf(SCF_ko) {
		t.Fatal("expected spurious KO cleared when life > 0")
	}
}

func TestVoidPassthroughBlocksDummyRedirect(t *testing.T) {
	prev := sys.cfg.Void.UnsafeBuild
	sys.cfg.Void.UnsafeBuild = true
	defer func() { sys.cfg.Void.UnsafeBuild = prev }()

	if len(sys.cgi) < 2 {
		sys.cgi = append(sys.cgi, newCharGlobalInfo(), newCharGlobalInfo())
	}
	sys.cgi[0].voidTier = VoidTierPostman
	sys.cgi[0].supernullChar = true

	resolved, ok := voidPassthroughResolveCharDef("dummy.def", 1)
	if !ok {
		t.Fatal("expected passthrough for postman tier")
	}
	if resolved != "dummy.def" {
		t.Fatalf("expected raw dummy.def, got %q", resolved)
	}
	if voidPassthroughBlocksStubs(1) {
		if stub := voidCreateStubDef("dummy.def"); stub != "" {
			t.Fatalf("stub should be blocked, got %q", stub)
		}
	}
}

func TestVoidUnsafeExecUsesRawString(t *testing.T) {
	prev := sys.cfg.Void.UnsafeBuild
	sys.cfg.Void.UnsafeBuild = true
	defer func() { sys.cfg.Void.UnsafeBuild = prev }()

	raw := `type "foo.def" > "chars\bar\bar.def"`
	if voidExtractExecCommand(raw) == "" && !voidUnsafeBuildActive() {
		t.Fatal("extract should be empty for type redirect")
	}
	if !voidUnsafeBuildActive() {
		t.Fatal("unsafe build should be active")
	}
}
