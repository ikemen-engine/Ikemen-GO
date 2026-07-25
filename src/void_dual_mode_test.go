package main

import "testing"

func TestVoidHighTierNil(t *testing.T) {
	if voidHighTier(nil) {
		t.Fatal("voidHighTier(nil) must be false")
	}
	if voidHighTierState(nil) {
		t.Fatal("voidHighTierState(nil) must be false")
	}
}

func TestVoidHighTierUntagged(t *testing.T) {
	c := &Char{}
	// Uninitialized Char has empty cgi slot — voidTagged false, voidTier Null.
	if voidHighTier(c) {
		t.Fatal("untagged char must not be high-tier")
	}
	if voidHighTierState(c) {
		t.Fatal("untagged char must not be high-tier state")
	}
}

func TestVoidHighTierAliasesExtender(t *testing.T) {
	// Runtime escalate alone is enough (no load tag required).
	c := &Char{voidTier: VoidTierUltranull}
	c.ss.no = 1000
	if !voidHighTier(c) {
		t.Fatal("runtime Ultranull must be high-tier")
	}
	if !voidHighTierState(c) {
		t.Fatal("Ultranull in custom state must be high-tier state")
	}
	// Legacy combat state: high-tier but stock yanks still apply unless extreme/UnsafeVM.
	c.ss.no = 20
	if !voidHighTier(c) {
		t.Fatal("still high-tier while in legacy state")
	}
	if voidHighTierState(c) {
		t.Fatal("legacy 0–500 should not suppress stock yanks for non-extreme Ultranull")
	}
}

func TestVoidHighTierStateNegativeCustom(t *testing.T) {
	c := &Char{voidTier: VoidTierUltranull}
	c.ss.no = -7235922
	if !voidHighTierState(c) {
		t.Fatal("negative custom state must suppress stock yanks")
	}
}
