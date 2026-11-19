package tools

import "testing"

func TestRequire(t *testing.T) {
	if err := Require("go"); err != nil {
		t.Errorf("Require(go) = %v, want nil", err)
	}

	if err := Require("kl-tool-that-does-not-exist"); err == nil {
		t.Error("Require() = nil, want error")
	}
}
