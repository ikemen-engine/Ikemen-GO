package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVoidShellIsDllName(t *testing.T) {
	for _, name := range []string{"VoidShell.Dll", "voidshell.dll", "VOIDSHELL.DLL", "path/VoidShell.dll"} {
		if !voidShellIsDllName(name) {
			t.Fatalf("expected VoidShell dll match for %q", name)
		}
	}
	for _, name := range []string{"kernel32.dll", "VoidShell.exe", "shell.dll"} {
		if voidShellIsDllName(name) {
			t.Fatalf("did not expect VoidShell match for %q", name)
		}
	}
}

func TestVoidShellCharDirDetect(t *testing.T) {
	dir := t.TempDir()
	def := filepath.Join(dir, "char.def")
	if err := os.WriteFile(def, []byte("name = Test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if voidShellCharDirHasDll(def) {
		t.Fatal("no DLL yet — must be false")
	}
	dll := filepath.Join(dir, "VoidShell.Dll")
	if err := os.WriteFile(dll, []byte("MZ"), 0644); err != nil {
		t.Fatal(err)
	}
	if !voidShellCharDirHasDll(def) {
		t.Fatal("VoidShell.Dll beside DEF must detect")
	}
}

func TestVoidShellMarkAffiliatesWithoutSecretary(t *testing.T) {
	dir := t.TempDir()
	def := filepath.Join(dir, "guest.def")
	if err := os.WriteFile(def, []byte("name = Guest\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "VoidShell.Dll"), []byte("MZ"), 0644); err != nil {
		t.Fatal(err)
	}
	sys.cgi[0] = newCharGlobalInfo()
	sys.cgi[0].def = def
	sys.cgi[0].displayname = "Guest"
	sys.cgi[0].voidTier = VoidTierNull
	voidMarkVoidShellAtLoad(0, def)
	if !sys.cgi[0].voidShell {
		t.Fatal("must set voidShell")
	}
	if sys.cgi[0].voidTier < VoidTierUltranull {
		t.Fatalf("expected Ultranull+, got %v", sys.cgi[0].voidTier)
	}
	if sys.cgi[0].voidTier >= VoidTierSecretary {
		t.Fatal("VoidShell alone must not escalate to Secretary")
	}
}

func TestVoidShellPlayerArtsGuestToNativeNoOp(t *testing.T) {
	sys.cgi[0] = newCharGlobalInfo()
	sys.cgi[0].voidShell = true
	sys.cgi[0].voidTier = VoidTierUltranull
	sys.cgi[0].supernullChar = true
	sys.cgi[1] = newCharGlobalInfo()
	sys.cgi[1].voidTier = VoidTierNull

	actor := &Char{playerNo: 0, voidTier: VoidTierUltranull}
	normal := &Char{playerNo: 1}
	if voidShellPlayerArtsAllow(actor, normal) {
		t.Fatal("guest→native Player Arts must Safe No-Op")
	}
	sys.cgi[1].voidShell = true
	sys.cgi[1].voidTier = VoidTierUltranull
	sys.cgi[1].supernullChar = true
	guest := &Char{playerNo: 1, voidTier: VoidTierUltranull}
	if !voidShellPlayerArtsAllow(actor, guest) {
		t.Fatal("guest→guest Player Arts must allow")
	}
}

func TestVoidShellSoulkeeperRestoresWipedStates(t *testing.T) {
	sys.cgi[0] = newCharGlobalInfo()
	sys.cgi[0].voidShell = true
	sys.cgi[0].voidTier = VoidTierUltranull
	sys.cgi[0].supernullChar = true
	sys.cgi[0].displayname = "ShellGuest"
	sys.cgi[0].states = map[int32]StateBytecode{
		-1:   {},
		0:    {},
		1000: {},
	}
	voidShellResetMatch()
	voidShellSnapshotPlayerCache(0)
	delete(sys.cgi[0].states, 1000)
	delete(sys.cgi[0].states, -1)

	c := &Char{playerNo: 0, helperIndex: 0, voidTier: VoidTierUltranull}
	sys.chars[0] = []*Char{c}
	voidShellSoulkeeperTick(c)
	if _, ok := sys.cgi[0].states[1000]; !ok {
		t.Fatal("Soulkeeper must restore wiped state 1000")
	}
	if _, ok := sys.cgi[0].states[-1]; !ok {
		t.Fatal("Soulkeeper must restore wiped state -1")
	}
	c.setSCF(SCF_disabled)
	voidShellSoulkeeperTick(c)
	if c.scf(SCF_disabled) {
		t.Fatal("Soulkeeper must clear remote SCF_disabled")
	}
	// Standby must not be cleared.
	c.setSCF(SCF_disabled)
	c.setSCF(SCF_standby)
	voidShellSoulkeeperTick(c)
	if !c.scf(SCF_disabled) {
		t.Fatal("Soulkeeper must not fight Turns/standby disable")
	}
	voidShellResetMatch()
	sys.chars[0] = nil
}

func TestVoidShellNormalsUnaffected(t *testing.T) {
	sys.cgi[2] = newCharGlobalInfo()
	sys.cgi[2].voidTier = VoidTierNull
	c := &Char{playerNo: 2}
	if voidShellActive(c) {
		t.Fatal("normal must not be VoidShell-active")
	}
	if voidShellRootActive(c) {
		t.Fatal("normal root must not be VoidShell-active")
	}
}
