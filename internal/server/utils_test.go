package server

import "testing"

var str = "Fri, 12 Jul 2024 13:00:00 +0200"

func TestParseTime(t *testing.T) {
	_, err := ParseTime(str)
	if err != nil {
		t.Fatal("shouldn't fail")
	}
}
