package wcwidth

import (
	"testing"
)

// test cases copied from https://github.com/mattn/go-runewidth/raw/master/runewidth_test.go

var stringwidthtests = []struct {
	in    string
	out   int
	eaout int
}{
	{"■㈱の世界①", 10, 12},
	{"スター☆", 7, 8},
	{"つのだ☆HIRO", 11, 12},
}

func BenchmarkStringWidth(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringWidth(stringwidthtests[i%len(stringwidthtests)].in)
	}
}

func TestStringWidth(t *testing.T) {
	for _, tt := range stringwidthtests {
		if out := StringWidth(tt.in); out != tt.out {
			t.Errorf("StringWidth(%q) = %d, want %d", tt.in, out, tt.out)
		}
	}
}

var runewidthtests = []struct {
	in  rune
	out int
}{
	{'世', 2},
	{'界', 2},
	{'ｾ', 1},
	{'ｶ', 1},
	{'ｲ', 1},
	{'☆', 1}, // double width in ambiguous
	{'☺', 1},
	{'☻', 1},
	{'♥', 1},
	{'♦', 1},
	{'♣', 1},
	{'♠', 1},
	{'♂', 1},
	{'♀', 1},
	{'♪', 1},
	{'♫', 1},
	{'☼', 1},
	{'↕', 1},
	{'‼', 1},
	{'↔', 1},
	{'\x00', 0},
	{'\x01', 0},
	{'\u0300', 0},
	{'\u2028', 0},
	{'\u2029', 0},
	{'a', 1}, // ASCII classified as "na" (narrow)
	{'⟦', 1}, // non-ASCII classified as "na" (narrow)
	{'👁', 1},
	{'\u0301', 0}, // Combining acute accent
	{'a', 1},
	{'Ω', 1},
	{'好', 2},
	{'か', 2},
}

func BenchmarkRuneWidth(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RuneWidth(runewidthtests[i%len(runewidthtests)].in)
	}
}

func TestRuneWidth(t *testing.T) {
	for i, tt := range runewidthtests {
		if out := RuneWidth(tt.in); out != tt.out {
			t.Errorf("case %d: RuneWidth(%q) = %d, want %d", i, tt.in, out, tt.out)
		}
	}
}

func TestZeroWidthJoiner(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"👩", 2},
		{"👩\u200d", 2},
		{"👩\u200d🍳", 4},
		{"\u200d🍳", 2},
		{"👨\u200d👨", 4},
		{"👨\u200d👨\u200d👧", 6},
		{"🏳️\u200d🌈", 3},
		{"あ👩\u200d🍳い", 8},
		{"あ\u200d🍳い", 6},
		{"あ\u200dい", 4},
		{"abc", 3},
		{"你好", 4},
	}

	for _, tt := range tests {
		if got := StringWidth(tt.in); got != tt.want {
			t.Errorf("StringWidth(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
