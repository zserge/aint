package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestTokens(t *testing.T) {
	tok := LoadTokens("tokens.dat")
	if len(tok) != 50257 {
		t.Fatal(len(tok))
	}
	for _, test := range []struct {
		s string
		t []int
	}{
		{" king", []int{5822}},
		{" queen", []int{16599}},
		{" nuclear", []int{4523}},
		{"Hello, world", []int{15496, 11, 995}},
	} {
		tokens := tok.Encode(test.s)
		if !reflect.DeepEqual(tokens, test.t) {
			t.Fatal(test.s, test.t, tokens)
		}
		if decoded := strings.Join(tok.Decode(tokens), ""); decoded != test.s {
			t.Fatal(test.s, decoded)
		}
	}
}
