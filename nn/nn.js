// Scalar contains a single numeric value and performs operations on it
const Scalar = (val, refs = [], gradrefs = [], op='') => {
	let x = {
		// val = scalar value
		// grad = scalar gradient
		// refs = references to related scalars
		// gradrefs = references to related scalar gradients
		// op = operation that resulted in this scalar (for debugging)
		val, refs, gradrefs, op, grad: 0,
		// All the math operations we need for a simple neural network
		add: (y) => Scalar(x.val + y.val, [x, y], [1, 1], '+'),
		sub: (y) => Scalar(x.val - y.val, [x, y], [1, -1], '-'),
		mul: (y) => Scalar(x.val * y.val, [x, y], [y.val, x.val], '*'),
		div: (y) => Scalar(x.val / y.val, [x, y], [1/y.val, -x.val / (y.val * y.val)], '/'),
		pow: (n) => Scalar(Math.pow(x.val, n), [x], [n * Math.pow(x.val, n - 1)], `^${n}`),
		relu: () => Scalar(Math.max(x.val, 0), [x], [+(x.val > 0)], 'relu'),
		// Backpropagation
		backward: () => {
			const visited = [];
			const nodes = [];
			const buildGraph = (n) => {
				if (visited.indexOf(n) < 0) {
					visited.push(n);
					n.refs.forEach(buildGraph)
					nodes.push(n);
				}
			};
			buildGraph(x);
			x.grad = 1;
			nodes.reverse();
			nodes.forEach(n => n.refs.forEach((ref, i) => ref.grad += n.grad * n.gradrefs[i]));
		},
		// Debug: print a .dot graph of all related scalar operations
		trace: () => {
			const nodes = [];
			const edges = [];
			const build = (v) => {
				if (nodes.indexOf(v) < 0) {
					nodes.push(v);
					for (let c of v.refs) {
						edges.push([c, v]);
						build(c);
					}
				}
			};
			build(x);
			return `digraph D {` + nodes.map((n, i) => {
				const isParam = n.op.startsWith('w') || n.op == 'b';
				const isZero = n.val == 0;
				return `N${i} [shape=record style="filled" fillcolor="${isParam ? "yellow" : isZero ? "white" : ""}" label="${n.op}|{ data ${n.val.toFixed(2)} | grad ${n.grad.toFixed(2)} }"]`;
			}).join('\n') + edges.map(e => `N${nodes.indexOf(e[0])} -> N${nodes.indexOf(e[1])}`).join('\n') + '\n}';
		},
	};
	return x;
};

// Tests from micrograd
(() => {
	const x = Scalar(-4.0)
	const z = Scalar(2).mul(x).add(Scalar(2)).add(x);
	const q = z.relu().add(z.mul(x));
	const h = z.mul(z).relu()
	const y = h.add(q).add(q.mul(x));
	y.backward()
	console.assert(y.val == -20);
	console.assert(x.grad == 46);
})();

(() => {
	const a = Scalar(-4.0);
	const b = Scalar(2.0);
  let c = a.add(b);
	let d = a.mul(b).add(b.pow(3));
	c = c.add(c.add(Scalar(1)));
	c = c.add(Scalar(1).add(c).sub(a));
	d = d.add(d.mul(Scalar(2)).add(b.add(a).relu()));
	d = d.add(Scalar(3).mul(d).add(b.sub(a).relu()));
	const e = c.sub(d);
	const f = e.pow(2);
	let g = f.div(Scalar(2));
	g = g.add(Scalar(10.0).div(f));
	g.backward()
	console.assert(a.grad|0 == 138);
	console.assert(b.grad|0 == 645);
	console.assert(g.val|0 == 24);
})();

// A single neuron (weights + bias)
const Neuron = (nin) => {
	let n = {
		w: Array(nin).fill(0).map((_, i) => Scalar(Math.random()*2-1, [], [], `w${i}`)),
		b: Scalar(0, [], [], 'b'),
		params: () => [...n.w, n.b],
		eval: (x) => x.map((xi, i) => xi.mul(n.w[i])).reduce((a, xwi) => a.add(xwi), n.b).relu(),
	};
	return n;
};

// A layer of neurons
const Layer = (nin, nout, nonlin = false) => {
	let l = {
		neurons: Array(nout).fill(0).map(() => Neuron(nin, nonlin)),
		params: () => l.neurons.reduce((p, n) => [...p, ...n.params()], []),
		eval: x => l.neurons.map(n => n.eval(x)),
	};
	return l;
};

// A multi-layer perceptron (network)
const MLP = (N) => {
	let mlp = {
		layers: N.slice(1).map((n, i) => Layer(N[i], N[i+1], i < N.length - 1)),
		eval: x => mlp.layers.reduce((a, l) => a = l.eval(a), x),
		params: () => mlp.layers.reduce((p, l) => [...p, ...l.params()], []),
		loss: (Xset, Yset, rate = 0.1) => {
			const out = Xset.map(xs => mlp.eval(xs.map(x => Scalar(x, [], [], 'X'))));
			const err = Yset.map((ys, i) => ys.map((ysi, j) => Scalar(ysi, [], [], 'Y').sub(out[i][j]).pow(2)));
			const totalErr = err.reduce((e, es) => e.add(es.reduce((a, e) => a.add(e), Scalar(0))), Scalar(0)).div(Scalar(err.length));
			mlp.params().map(p => p.grad = 0);
			totalErr.backward();
			mlp.params().map(p => p.val -= p.grad * rate);
			return totalErr;
		},
	};
	return mlp;
};

// Test data generator for the "two moons" problem
const moons = (n) => {
	const unif = (low, high) => Math.random() * (high - low) + low;
  const data = [], labels = [];
  for (let i = 0; i < Math.PI; i += (Math.PI * 2) / n) {
    data.push([Math.cos(i) + unif(-0.1, 0.1) - 0.5, Math.sin(i) + unif(-0.1, 0.1) - 0.3]);
    labels.push([0]);
		data.push([0.5 - Math.cos(i) + unif(-0.1, 0.1), 0.2 - Math.sin(i) + unif(-0.1, 0.1)]);
		labels.push([1]);
  }
  return [data, labels];
}

// Test data generator for XOR problem
const xor = () => {
	return [
			[[0, 0], [0, 1], [1, 0], [1, 1]],
			[[0], [1], [1], [0]],
	];
}

const [X, Y] = moons(200);
const nn = MLP([2,8,4,1])

for (let i = 0; i < 100; i++) {
	const e = nn.loss(X, Y, 0.2);
	console.log(i, e.val);
}

// Visualise the solution for "two-moons" problem
for (let y = -10; y < 10; y++) {
	let s = '';
	for (let x = -20; x < 20; x++) {
		const point = X.find(p => ((p[0]*10)|0) == x && ((p[1]*10)|0) == y);
		let bg = '';
		if (point) {
			bg = (Y[X.indexOf(point)][0] > 0 ? '\x1b[41m' : '\x1b[44m');
		} 
		const n = nn.eval([Scalar(x/10), Scalar(y/10)])[0].val;
		s = s + bg + (n > 0.5 ? '*' : '.') + '\x1b[0m';
	}
	console.log(s);
}
