const relu = x => x < 0 ? 0 : x;
// Run the network for the given input array X
const nn = X => {
  // Define the parameters for the network
  const W1 = [[-1, -1], [-0.7, -0.7]], W2 = [-1.8, 1.3];
  const b1 = [1, 1.3], b2 = 0;
  // Calculate the output
  const hidden = b1.map((b, i) => relu(X.reduce((a, x, j) => a+x*W1[i][j], 0) + b))
  return relu(hidden.reduce((a, x, i) => a+x*W2[i], 0) + b2);
};

// Predict XOR function:
console.log('0^0', nn([0, 0])); // 0^0 => 0
console.log('0^1', nn([0, 1])); // 0^1 => 0.9916
console.log('1^0', nn([1, 0])); // 1^0 => 0.9916
console.log('1^1', nn([1, 1])); // 1^1 => 0

