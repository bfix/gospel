package math

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var (
	// ZERO as number "0"
	ZERO = NewInt(0)
	// ONE as number "1"
	ONE = NewInt(1)
	// TWO as number "2"
	TWO = NewInt(2)
	// THREE as number "3"
	THREE = NewInt(3)
	// FOUR as number "4"
	FOUR = NewInt(4)
	// FIVE as number "5"
	FIVE = NewInt(5)
	// SIX as number "6"
	SIX = NewInt(6)
	// SEVEN as number "7"
	SEVEN = NewInt(7)
	// EIGHT as number "8"
	EIGHT = NewInt(8)
)

// Int is an integer of arbitrary size
type Int struct {
	v *big.Int
}

// NewInt returns a new Int from an intrinsic int64
func NewInt(v int64) *Int {
	return &Int{v: big.NewInt(v)}
}

// NewIntFromString converts a string representation of an integer
func NewIntFromString(s string) *Int {
	v := new(big.Int)
	if err := v.UnmarshalText([]byte(s)); err != nil {
		panic(err)
	}
	return &Int{v}
}

// NewIntFromHex converts a hexadecimal string into an unsigned integer.
func NewIntFromHex(s string) *Int {
	val, _ := new(big.Int).SetString(s, 16)
	return &Int{v: val}
}

// NewIntFromBytes converts a binary array into an unsigned integer.
func NewIntFromBytes(buf []byte) *Int {
	return &Int{v: new(big.Int).SetBytes(buf)}
}

// NewIntRnd creates a new random value between [0,j[
func NewIntRnd(j *Int) *Int {
	r, err := rand.Int(rand.Reader, j.v)
	if err != nil {
		panic(err)
	}
	return &Int{v: r}
}

// NewIntRndBits creates a new random value with a max. bitlength.
func NewIntRndBits(n int) *Int {
	return NewIntRnd(TWO.Pow(n))
}

// NewIntRndRange returns a random integer value within given range.
func NewIntRndRange(lower, upper *Int) *Int {
	return lower.Add(NewIntRnd(upper.Sub(lower).Add(ONE)))
}

// NewIntRndPrime generates a new random prime number between [0,j[
func NewIntRndPrime(j *Int) *Int {
	r := NewIntRnd(j)
	if r.Bit(0) == 0 {
		r = r.Add(ONE)
	}
	for {
		if r.ProbablyPrime(128) {
			return r
		}
		r = r.Add(TWO)
	}
}

// NewIntRndPrimeBits generates a new random prime number with a maximum
// bitlength
func NewIntRndPrimeBits(n int) *Int {
	return NewIntRndPrime(TWO.Pow(n))
}

// Bytes returns a byte array representation of the integer.
func (i *Int) Bytes() []byte {
	return i.v.Bytes()
}

// String converts an Int to a string representation.
func (i *Int) String() string {
	return i.v.String()
}

// ProbablyPrime checks if an Int is prime. The chances this is wrong
// are less than 2^(-n).
func (i *Int) ProbablyPrime(n int) bool {
	return i.v.ProbablyPrime(n)
}

// Add two Ints
func (i *Int) Add(j *Int) *Int {
	return &Int{v: new(big.Int).Add(i.v, j.v)}
}

// Sub substracts two Ints
func (i *Int) Sub(j *Int) *Int {
	return &Int{v: new(big.Int).Sub(i.v, j.v)}
}

// Mul multiplies two Ints
func (i *Int) Mul(j *Int) *Int {
	return &Int{v: new(big.Int).Mul(i.v, j.v)}
}

// Div divides two Int (no fraction)
func (i *Int) Div(j *Int) *Int {
	return &Int{v: new(big.Int).Div(i.v, j.v)}
}

// DivMod returns the quotient and modulus of two Ints.
func (i *Int) DivMod(j *Int) (*Int, *Int) {
	return &Int{v: new(big.Int).Div(i.v, j.v)}, &Int{v: new(big.Int).Mod(i.v, j.v)}
}

// Mod returns the modulus of two Ints
func (i *Int) Mod(j *Int) *Int {
	return &Int{v: new(big.Int).Mod(i.v, j.v)}
}

// ModSign returns a signed modulus of two Ints.
func (i *Int) ModSign(j *Int) *Int {
	k := i.Mod(j)
	if k.Mul(TWO).Cmp(j) > 0 {
		k = k.Sub(j)
	}
	return k
}

// BitLen returns the number of bits in an Int.
func (i *Int) BitLen() int {
	return i.v.BitLen()
}

// Sign returns the sign of an Int.
func (i *Int) Sign() int {
	return i.v.Sign()
}

// ModInverse returns the multiplicative inverse of i in the ring ℤ/jℤ.
func (i *Int) ModInverse(j *Int) *Int {
	return &Int{v: new(big.Int).ModInverse(i.v, j.v)}
}

// Cmp returns the comparision between two Ints.
func (i *Int) Cmp(j *Int) int {
	return i.v.Cmp(j.v)
}

// Equals check if two Ints are equal.
func (i *Int) Equals(j *Int) bool {
	return i.v.Cmp(j.v) == 0
}

// GCD return the greatest common divisor of two Ints.
func (i *Int) GCD(j *Int) *Int {
	return &Int{v: new(big.Int).GCD(nil, nil, i.v, j.v)}
}

// LCM returns the least common multiplicative of two Ints.
func (i *Int) LCM(j *Int) *Int {
	g := i.GCD(j)
	return i.Mul(j).Div(g)
}

// Pow raises an Int to power n.
func (i *Int) Pow(n int) *Int {
	return &Int{v: new(big.Int).Exp(i.v, big.NewInt(int64(n)), nil)}
}

// ModPow returns the modular exponentiation of an Int as (i^n mod m).
func (i *Int) ModPow(n, m *Int) *Int {
	return &Int{v: new(big.Int).Exp(i.v, n.v, m.v)}
}

// Bit returns the bit value of an Int at a given position.
func (i *Int) Bit(n int) uint {
	return i.v.Bit(n)
}

// Rsh returns the right shifted value of an Int.
func (i *Int) Rsh(n uint) *Int {
	return &Int{v: new(big.Int).Rsh(i.v, n)}
}

// Lsh returns the left shifted value of an Int.
func (i *Int) Lsh(n uint) *Int {
	return &Int{v: new(big.Int).Lsh(i.v, n)}
}

// NthRoot computes the n.th root of an Int. If upper is set, the
// result is raised to the next highest value.
func (i *Int) NthRoot(n int, upper bool) *Int {
	r := ZERO
	b := i.v.BitLen()
	if n < b {
		for s := TWO.Pow(b/n - 1); s.Cmp(ZERO) > 0; r = r.Add(s) {
			if t := r.Pow(n); t.Cmp(i) > 0 {
				r = r.Sub(s)
				s = s.Div(TWO)
			}
		}
	}
	if r.Mul(r).Cmp(i) < 0 && upper {
		r = r.Add(ONE)
	}
	return r
}

// Legendre computes (i\p)
func (i *Int) Legendre(p *Int) int {
	if i.Mod(p).Equals(ZERO) {
		return 0
	}
	k := p.Sub(ONE).Div(TWO)
	x := i.ModPow(k, p)
	if x.Equals(ONE) {
		return 1
	}
	return -1
}

// ExtendedEuclid computes the factors (x,y) for (a,b) where the
// following equation is satisfied: x*a + b*y = gcd(a,b)
func (i *Int) ExtendedEuclid(j *Int) [2]*Int {
	var impl func(a, b *Int) [2]*Int
	impl = func(a, b *Int) [2]*Int {
		t := a.Mod(b)
		if t.Equals(ZERO) {
			return [2]*Int{ZERO, ONE}
		}
		r := impl(b, t)
		return [2]*Int{r[1], r[0].Sub(r[1].Mul(a.Div(b)))}
	}
	if i.Cmp(j) < 0 {
		r := impl(j, i)
		xc := j.Sub(r[1].Abs())
		yc := i.Sub(r[0].Abs())
		if r[0].Cmp(ZERO) > 0 {
			return [2]*Int{xc, yc.Neg()}
		}
		return [2]*Int{xc.Neg(), yc}
	}
	return impl(i, j)
}

// Abs returns the unsigned value of an Int.
func (i *Int) Abs() *Int {
	return &Int{v: new(big.Int).Abs(i.v)}
}

// Neg flips the sign of an Int.
func (i *Int) Neg() *Int {
	return &Int{v: new(big.Int).Neg(i.v)}
}

// Int64 returns the int64 value of an Int.
func (i *Int) Int64() int64 {
	return i.v.Int64()
}

// SqrtModP computes the square root of a quadratic residue mod p
// It uses the Shanks-Tonelli algorithm to compute the square root
// see (http://en.wikipedia.org/wiki/Shanks%E2%80%93Tonelli_algorithm)
func SqrtModP(n, p *Int) (*Int, error) {
	// check if a solution is possible
	if n.Legendre(p) != 1 {
		return nil, errors.New("No quadratic residue")
	}
	// 1. Factor out powers of 2 from p − 1, defining Q and S as:
	//    p − 1 = Q*2^S with Q odd
	S := 0
	Q := p.Sub(ONE)
	for Q.Bit(0) == 0 {
		S++
		Q = Q.Div(TWO)
	}
	if S == 1 {
		return n.ModPow(p.Add(ONE).Div(FOUR), p), nil
	}
	// 2. Select a z such that the Legendre(z/p) = − 1 (that is, z is a
	//    quadratic non-residue modulo p), and set c ≡ z^Q
	z := ONE
	for z.Legendre(p) != -1 {
		z = z.Add(ONE)
	}
	c := z.ModPow(Q, p)
	// 3. Let R ≡ n^((Q+1)/2), t ≡ n^Q, M = S.
	R := n.ModPow(Q.Add(ONE).Div(TWO), p)
	t := n.ModPow(Q, p)
	M := S
	// 4. Loop...
	for {
		// 4.1. If t ≡ 1, return R.
		if t.Mod(p).Equals(ONE) {
			break
		}
		// 4.2. Otherwise, find the lowest i, 0 < i < M, such that t^(2^i) ≡ 1;
		//      e.g. via repeated squaring.
		for i := 1; i < M; i++ {
			if t.ModPow(TWO.Pow(i), p).Equals(ONE) {
				// 4.3. Let b ≡ c^(2^(M-i-1)), and set R ≡ R*b, t ≡ t*b^2,
				//      c ≡ b^2 and M = i
				b := c.ModPow(TWO.Pow(M-i-1), p)
				R = R.Mul(b).Mod(p)
				t = t.Mul(b.Pow(2)).Mod(p)
				c = b.ModPow(TWO, p)
				M = i
				break
			}
		}
	}
	return R, nil
}
