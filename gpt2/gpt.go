package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unsafe"
)

const (
	ContextSize = 1024
)

var (
	WordVecSize int
	NumLayers   int
	NumHeads    int
)

type Model struct {
	dir    string
	lnf_g  []float32
	lnf_b  []float32
	wte    []float32 // word token embeddings
	wpe    []float32 // word position embeddings
	layers []Layer
}

type Layer struct {
	ln1_b        []float32
	ln1_g        []float32
	ln2_b        []float32
	ln2_g        []float32
	mlp_cfc_b    []float32
	mlp_cfc_w    []float32
	mlp_cproj_b  []float32
	mlp_cproj_w  []float32
	attn_cattn_b []float32
	attn_cattn_w []float32
	attn_cproj_b []float32
	attn_cproj_w []float32
	k            []float32
	v            []float32
}

type Tokens []string

func LoadTokens(filename string) (tokens Tokens) {
	b, _ := ioutil.ReadFile(filename)
	for len(b) > 0 {
		i := bytes.IndexByte(b, 0) // tokens are null-terminated
		tokens = append(tokens, string(b[:i]))
		b = b[i+1:]
	}
	return tokens
}

func (tokens Tokens) Find(s string) (index, overlap int) {
	for i, t := range tokens {
		j := 0
		for ; j < len(s) && j < len(t) && s[j] == t[j]; j++ {
		}
		if j > overlap || (j == overlap && j == len(t)) {
			overlap, index = j, i
		}
	}
	return index, overlap
}

func (tokens Tokens) Encode(s string) (context []int) {
	for len(s) > 0 {
		t, n := tokens.Find(s)
		if t < 0 {
			return context
		}
		context = append(context, t)
		s = s[n:]
	}
	return context
}

func (tokens Tokens) Decode(t []int) (words []string) {
	for _, ti := range t {
		words = append(words, tokens[ti])
	}
	return words
}

func LoadModel(dir string) (m Model) {
	m.dir = dir
	m.lnf_g = m.read("lnf_g.dat")
	m.lnf_b = m.read("lnf_b.dat")
	m.wte = m.read("wte.dat")
	m.wpe = m.read("wpe.dat")
	m.layers = make([]Layer, NumLayers)
	for i := range m.layers {
		l := &m.layers[i]
		l.ln1_g = m.read(fmt.Sprintf("h%d_ln1_g.dat", i))
		l.ln1_b = m.read(fmt.Sprintf("h%d_ln1_b.dat", i))
		l.ln2_g = m.read(fmt.Sprintf("h%d_ln2_g.dat", i))
		l.ln2_b = m.read(fmt.Sprintf("h%d_ln2_b.dat", i))
		l.mlp_cfc_w = m.read(fmt.Sprintf("h%d_mlp_cfc_w.t", i))
		l.mlp_cfc_b = m.read(fmt.Sprintf("h%d_mlp_cfc_b.dat", i))
		l.mlp_cproj_w = m.read(fmt.Sprintf("h%d_mlp_cproj_w.t", i))
		l.mlp_cproj_b = m.read(fmt.Sprintf("h%d_mlp_cproj_b.dat", i))
		l.attn_cproj_w = m.read(fmt.Sprintf("h%d_attn_cproj_w.t", i))
		l.attn_cproj_b = m.read(fmt.Sprintf("h%d_attn_cproj_b.dat", i))
		l.attn_cattn_w = m.read(fmt.Sprintf("h%d_attn_cattn_w.t", i))
		l.attn_cattn_b = m.read(fmt.Sprintf("h%d_attn_cattn_b.dat", i))
		l.k = make([]float32, ContextSize*WordVecSize)
		l.v = make([]float32, ContextSize*WordVecSize)
	}
	return m
}

func (m Model) read(filename string) []float32 {
	b, _ := ioutil.ReadFile(m.dir + "/" + filename)
	return unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), len(b)/4)
}

func (m Model) WordVec(token int) []float32 {
	return m.wte[WordVecSize*token : WordVecSize*(token+1)]
}

func (m Model) runLayer(x []float32, layer, slot int) {
	l := m.layers[layer]
	xn := norm(x, l.ln1_b, l.ln1_g)
	q := make([]float32, WordVecSize)
	for i := 0; i < WordVecSize*3; i++ {
		a := lin(xn, l.attn_cattn_w[WordVecSize*i:WordVecSize*(i+1)], l.attn_cattn_b[i])
		if i < WordVecSize {
			q[i] = a
		} else if i < WordVecSize*2 {
			l.k[slot*WordVecSize+(i-WordVecSize)] = a
		} else {
			l.v[(i-WordVecSize*2)*ContextSize+slot] = a
		}
	}

	const headSize = 64
	tmp := make([]float32, WordVecSize)
	for h := 0; h < NumHeads; h++ {
		att := make([]float32, slot+1)
		for i := 0; i <= slot; i++ {
			att[i] = lin(q[h*headSize:(h+1)*headSize], l.k[i*WordVecSize+h*headSize:], 0) / 8
		}
		att = softmax(att)
		for j := 0; j < headSize; j++ {
			tmp[h*headSize+j] = lin(att, l.v[(j+h*headSize)*ContextSize:], 0)
		}
	}
	for i := 0; i < WordVecSize; i++ {
		x[i] += lin(tmp, l.attn_cproj_w[WordVecSize*i:], l.attn_cproj_b[i])
	}
	xn = norm(x, l.ln2_b, l.ln2_g)
	mlp := make([]float32, WordVecSize*4)
	for i := 0; i < WordVecSize*4; i++ {
		mlp[i] = gelu(lin(xn, l.mlp_cfc_w[WordVecSize*i:], l.mlp_cfc_b[i]))
	}
	for i := 0; i < WordVecSize; i++ {
		x[i] += lin(mlp, l.mlp_cproj_w[WordVecSize*4*i:], l.mlp_cproj_b[i])
	}
}

func (m Model) Run(context []int, slot int) []float32 {
	x := make([]float32, WordVecSize)
	wv := m.WordVec(context[slot])
	for i := range x {
		x[i] = m.wpe[i+WordVecSize*slot]
		x[i] += wv[i]
	}
	for i := range m.layers {
		m.runLayer(x, i, slot)
	}
	return norm(x, m.lnf_b, m.lnf_g)
}

func gelu(x float32) float32 {
	return 0.5 * x * (1 + float32(math.Tanh(0.7978845676080871*float64(x+0.044715*x*x*x))))
}

func softmax(x []float32) []float32 {
	out := make([]float32, len(x))
	max, sum := float32(math.Inf(-1)), float32(0)
	for i := range x {
		if x[i] > max {
			max = x[i]
		}
	}
	for i := range x {
		x[i] = float32(math.Exp(float64(x[i] - max)))
		sum += x[i]
	}
	for i := range x {
		out[i] = x[i] / sum
	}
	return out
}

func lin(x, w []float32, b float32) float32 {
	for i := range x {
		b += x[i] * w[i]
	}
	return b
}

func norm(x, beta, gamma []float32) []float32 {
	mean, sqmean := float32(0.0), float32(0.0)
	for i := range x {
		mean += x[i]
	}
	mean = mean / float32(len(x))
	for _, xi := range x {
		sqmean += (xi - mean) * (xi - mean)
	}
	sqmean = float32(math.Max(float64(sqmean/float32(len(x))), 0.0000001))
	m := float32(math.Sqrt(1.0 / float64(sqmean)))
	out := make([]float32, len(x))
	for i, xi := range x {
		out[i] = (xi-mean)*m*gamma[i] + beta[i]
	}
	return out
}

func main() {
	seed := flag.Int64("seed", int64(time.Now().UnixNano()), "PRNG seed")
	modelDir := flag.String("m", "124M", "GPT-2 model to run (124M, 355M, 774M)")
	choice := flag.Int("c", 20, "choice of words to pick from")
	numWords := flag.Int("w", 1, "number of words to generate in a sentence")
	flag.Parse()

	switch *modelDir {
	case "124M": // 124M model
		WordVecSize, NumLayers, NumHeads = 768, 12, 12
	case "335M": // 335M model
		WordVecSize, NumLayers, NumHeads = 1024, 24, 16
	case "774M": // 774M model
		WordVecSize, NumLayers, NumHeads = 1280, 36, 20
	default:
		fmt.Println("unknown model")
		return
	}

	rand.Seed(*seed)

	tokens := LoadTokens(filepath.Join(*modelDir, "tokens.dat"))
	prompt := strings.Join(flag.Args(), " ")
	context := tokens.Encode(strings.Join(flag.Args(), " "))
	fmt.Println("> " + strings.Join(tokens.Decode(context), " "))

	out := []float32{}
	candidates := make([]int, *choice)

	model := LoadModel(*modelDir)
	for i := range context {
		out = model.Run(context, i)
	}

	for i := 0; i < *numWords; i++ {
		match := make([]struct {
			token int
			prob  float32
		}, len(tokens))
		for i := range match {
			match[i].prob = lin(out, model.WordVec(i), 0)
			match[i].token = i
		}
		sort.Slice(match, func(i, j int) bool { return match[j].prob < match[i].prob })
		for i := range candidates {
			candidates[i] = match[i].token
		}
		fmt.Println("\x1b[0;2m" + strings.Join(tokens.Decode(candidates), " "))
		r := rand.Float64()
		n := candidates[(int)((r*r)*float64(len(candidates)))]

		context = append(context, n)
		out = model.Run(context, len(context)-1)
		prompt = prompt + tokens[n]
		fmt.Println("\x1b[0;1m" + prompt)
	}
}
