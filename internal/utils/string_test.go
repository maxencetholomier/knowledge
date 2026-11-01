package utils

import "testing"

func TestCreateTimestampIsUniqueAndIncreasing(t *testing.T) {
	first := CreateTimestamp()
	second := CreateTimestamp()
	if len(first) != 14 || second <= first {
		t.Errorf("got %q then %q", first, second)
	}
}
