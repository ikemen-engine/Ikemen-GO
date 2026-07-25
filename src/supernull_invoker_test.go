package main

import (
	"strings"
	"testing"
)

func TestVoidCNSDataLooksInvoker(t *testing.T) {
	raw := strings.Repeat("x", 80) + "\n[State 0]\ntype = AssertSpecial\nflag = qBHELLHELLHELLHELLHELLHELLHELLzr@\n"
	if !voidCNSDataLooksInvoker(raw) {
		t.Fatal("expected invoker CNS signature")
	}
	if voidCNSDataLooksInvoker("type = ChangeState\nvalue = 0") {
		t.Fatal("vanilla CNS should not match invoker heuristic")
	}
}

func TestVoidInvokerPayloadLooksLethal(t *testing.T) {
	if !voidInvokerPayloadLooksLethal(`command="uCHELLHELLHELLHELLHELLHELLHELLHELLHELLHELLHELLHELLHELLHELLHELLHEAFA"`) {
		t.Fatal("HELLHELL trigger should be lethal")
	}
}

func TestVoidUnsafeVMActiveForPlayer(t *testing.T) {
	sys.cgi = make([]CharGlobalInfo, 2)
	voidMarkInvokerProfile(0, "test")
	if !voidUnsafeVMActiveForPlayer(0) {
		t.Fatal("invoker profile should enable UnsafeVM")
	}
	if voidUnsafeVMActiveForPlayer(1) {
		t.Fatal("non-invoker slot should not enable UnsafeVM")
	}
}
