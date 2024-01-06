package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

type Markov struct {
	Order int
	Chain map[string][]string
	Start []string
	RNG   func(int) int
}

func NewMarkov(order int) *Markov {
	return &Markov{Order: order, Chain: map[string][]string{}, RNG: rand.Intn}
}

func (m *Markov) Add(input []string) {
	if len(input) < m.Order {
		return
	}
	input = append(input, make([]string, m.Order)...) // pad with empty strings
	m.Start = append(m.Start, strings.Join(input[:m.Order], " "))
	for i := 0; i < len(input)-m.Order; i++ {
		prefix := strings.Join(input[i:i+m.Order], " ")
		m.Chain[prefix] = append(m.Chain[prefix], input[i+m.Order])
	}
}

func (m *Markov) Generate() string {
	w := m.Start[m.RNG(len(m.Start))]
	out := []string{w}
	for {
		candidates := m.Chain[w]
		if len(candidates) == 0 {
			break
		}
		next := candidates[m.RNG(len(candidates))]
		out = append(out, next)
		parts := strings.Fields(w)
		if len(parts) < m.Order {
			break
		}
		w = strings.Join(append(parts[1:m.Order], next), " ")
	}
	return strings.TrimSpace(strings.Join(out, " "))
}

func main() {
	markov := NewMarkov(2)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		markov.Add(words)
	}
	fmt.Println(markov.Generate())
}
