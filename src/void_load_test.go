package main

import (
	"os"
	"strings"
	"testing"
)

func TestVoidIsGarbageSelectPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"Kuro Aonix", false},
		{"%Eg3%83d%9z6j%E3%28c3%Bx3%gEe3%8a2%B2Ffg%En3", true},
		{"", true},
		{"Animation", false},
	}
	for _, c := range cases {
		if got := voidIsGarbageSelectPath(c.path); got != c.want {
			t.Fatalf("voidIsGarbageSelectPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestVoidPermissiveCharLoadAlwaysOnInVoidBuild(t *testing.T) {
	if !voidExploitsEnabled {
		t.Skip("voidExploitsEnabled is false in this build")
	}
	if !voidPermissiveCharLoad() {
		t.Fatal("voidPermissiveCharLoad should be true when voidExploitsEnabled is true")
	}
}

func TestVoidPermissiveCharName(t *testing.T) {
	if got := voidPermissiveCharName("chars/foo/bar.def"); got != "bar" {
		t.Fatalf("got %q want bar", got)
	}
	if got := voidPermissiveCharName("Aestheticdistortionist.zip"); got != "Aestheticdistortionist" {
		t.Fatalf("got %q want Aestheticdistortionist", got)
	}
}

func TestVoidCreateStubAsset(t *testing.T) {
	if !voidPermissiveCharLoad() {
		t.Skip("permissive load disabled in this build")
	}
	stub := voidCreateStubAsset("chars/GF-Orochi/GF-Orochi.def", "cns_!).txt", "cns")
	if stub == "" {
		t.Fatal("expected stub cns path")
	}
	if _, err := os.Stat(stub); err != nil {
		t.Fatalf("stub file missing: %v", err)
	}
	data, err := os.ReadFile(stub)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[Statedef 0]") {
		t.Fatalf("stub cns missing statedef: %q", string(data))
	}
}

func TestVoidEnsureDummyDef(t *testing.T) {
	if !voidPermissiveCharLoad() {
		t.Skip("permissive load disabled in this build")
	}
	if sys.cfg.Void.UnsafeBuild {
		t.Skip("UnsafeBuild redirects dummy.def without stub creation")
	}
	if !voidIsDummyDefRequest("dummy.def") {
		t.Fatal("dummy.def should match")
	}
	d := voidEnsureDummyDef()
	if d == "" || FileExist(d) == "" {
		t.Fatalf("expected dummy def, got %q", d)
	}
	resolved := voidResolveCharDefForLoad("dummy.def", 1)
	if FileExist(resolved) == "" {
		t.Fatalf("resolve dummy.def failed: %q", resolved)
	}
}
