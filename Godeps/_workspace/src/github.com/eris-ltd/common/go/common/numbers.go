package common

import (
	"math/big"
)

var tt256 = new(big.Int).Lsh(big.NewInt(1), 256)
var tt256m1 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
var tt255 = new(big.Int).Lsh(big.NewInt(1), 255)

func U256(x *big.Int) *big.Int {
	//if x.Cmp(Big0) < 0 {
	//		return new(big.Int).Add(tt256, x)
	//	}

	x.And(x, tt256m1)

	return x
}

func S256(x *big.Int) *big.Int {
	if x.Cmp(tt255) < 0 {
		return x
	} else {
		// We don't want to modify x, ever
		return new(big.Int).Sub(x, tt256)
	}
}
