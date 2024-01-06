package main

import (
	"math/rand"
	"strings"
	"testing"
)

func TestMarkov(t *testing.T) {
	m := NewMarkov(2)
	m.Add(strings.Split("Mary had a little lamb little lamb little lamb", " "))
	if len(m.Start) != 1 {
		t.Error(m.Start)
	}
	if len(m.Chain) != 6 {
		t.Error(len(m.Chain))
	}
	for _, prefix := range []string{"Mary had", "had a", "a little", "little lamb", "lamb little", "lamb "} {
		if _, ok := m.Chain[prefix]; !ok {
			t.Error(prefix, m.Chain)
		}
	}
	m.Add(strings.Split("Old McDonald had a farm", " "))
	if len(m.Chain) != 10 {
		t.Error(len(m.Chain))
	}
	if len(m.Chain["had a"]) != 2 {
		t.Error(m.Chain["had a"])
	}
	rand.Seed(12)
	if s := m.Generate(); s != "Old McDonald had a little lamb little lamb" {
		t.Error(s)
	}
	if s := m.Generate(); s != "Mary had a farm" {
		t.Error(s)
	}
}
