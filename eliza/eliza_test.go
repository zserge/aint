package main

import (
	"sort"
	"strings"
	"testing"
)

func TestReplace(t *testing.T) {
	mapping := map[string]string{
		"foo":       "bar",
		"bar":       "f o o",
		"war":       "peace",
		"freedome":  "slavery",
		"ignorance": "strength",
	}
	for _, test := range []struct {
		Text   string
		Result string
	}{
		{"", ""},
		{"war is peace", "peace is peace"},
		{"foo", "bar"},
		{"foo bar", "bar f o o"},
		{"foo bar bar", "bar f o o f o o"},
		{"das rhabarberbarbarabarbarbarenbartbarbierbier konnte man nur an einer ganz bestimmten bar kaufen", "das rhabarberbarbarabarbarbarenbartbarbierbier konnte man nur an einer ganz bestimmten f o o kaufen"},
	} {
		result := replace(strings.Fields(test.Text), mapping)
		if strings.Join(result, " ") != test.Result {
			t.Error(test.Text, "->", test.Result, "!=", result)
		}
	}
}

func TestMatch(t *testing.T) {
	syn := map[string][]string{"letter": []string{"a", "b", "c"}, "number": []string{"1", "2", "3"}}
	post := map[string]string{"a": "z", "z": "a"}

	for _, test := range []struct {
		Pattern string
		Words   string
		Match   bool
		Groups  []string
	}{
		{"", "", true, nil},
		{"a", "a", true, nil},
		{"a", "z", false, nil},
		{"a", "b", false, nil},
		{"*", "a", true, []string{"z"}},
		{"*", "a b", true, []string{"z b"}},
		{"* c", "a b c", true, []string{"z b"}},
		{"* c *", "a b c d e", true, []string{"z b", "d e"}},
		{"* /number", "a b a 1", true, []string{"z b z", "1"}},
		{"* /number", "a b a b", false, nil},
	} {
		g, ok := match(strings.Fields(test.Pattern), strings.Fields(test.Words), syn, post)
		if ok != test.Match {
			t.Error(test, ok)
		} else if len(g) != len(test.Groups) {
			t.Error(test, g)
		} else {
			for i, m := range g {
				if m != test.Groups[i] {
					t.Error(test, g)
				}
			}
		}
	}
}

func TestEliza(t *testing.T) {
	sort.Slice(keywords, func(i, j int) bool { return keywords[i].Rank > keywords[j].Rank })
	for _, msg := range []struct {
		Input  string
		Output string
	}{
		// A dialogue from Jan 1966 Weizenbaum's paper about Eliza:
		{"Men are all alike", "In what way ?"},
		{"They're always bugging us about something or other", "Can you think of a specific example ?"},
		{"Well, my boyfriend made me come here", "Lets discuss further why your boyfriend made you come here."},
		{"He says I'm depressed much of the time.", "I am sorry to hear that you are depressed."},
		{"It's true. I am unhappy", "Do you think coming here will help you not to be unhappy ?"},
		{"I need some help", "What would it mean to you if you got some help ?"},
		{"Perhaps I could learn to get along with my mother", "Tell me more about your family."},
		{"My mother takes care of me", "Who else in your family takes care of you ?"},
		{"My father", "Your father ?"},
		{"You are like my father in some ways", "What resemblence do you see ?"},
		{"You are not very aggressive", "What makes you think I am not very aggressive ?"},
		{"You don't argue with me", "Why do you think I don't argue with you ?"},
		{"You are afraid of me", "Does it please you to believe I am afraid of you ?"},
		{"My father is afraid of me", "What else comes to your mind when you think of your father ?"},
		{"Bullies", "Lets discuss further why your boyfriend made you come here."},
	} {
		out := respond(msg.Input)
		if out != msg.Output {
			t.Error(msg.Input, msg.Output, out)
		}
	}
}
