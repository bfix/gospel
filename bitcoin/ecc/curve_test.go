package ecc

import (
	"github.com/bfix/gospel/math"
	"testing"
)

var (
	g  = &Point{c.Gx, c.Gy}
	gm = &Point{c.Gx, c.Gy.Neg()}

	tests = [][]string{
		{"AA5E28D6A97A2479A65527F7290311A3624D4CC0FA1578598EE3C2613BF99522",
			"34F9460F0E4F08393D192B3C5133A6BA099AA0AD9FD54EBCCFACDFA239FF49C6",
			"0B71EA9BD730FD8923F6D25A7A91E7DD7728A960686CB5A901BB419E0F2CA232"},

		{"7E2B897B8CEBC6361663AD410835639826D590F393D90A9538881735256DFAE3",
			"D74BF844B0862475103D96A611CF2D898447E288D34B360BC885CB8CE7C00575",
			"131C670D414C4546B88AC3FF664611B1C38CEB1C21D76369D7A7A0969D61D97D"},

		{"6461E6DF0FE7DFD05329F41BF771B86578143D4DD1F7866FB4CA7E97C5FA945D",
			"E8AECC370AEDD953483719A116711963CE201AC3EB21D3F3257BB48668C6A72F",
			"C25CAF2F0EBA1DDB2F0F3F47866299EF907867B7D27E95B3873BF98397B24EE1"},

		{"376A3A2CDCD12581EFFF13EE4AD44C4044B8A0524C42422A7E1E181E4DEECCEC",
			"14890E61FCD4B0BD92E5B36C81372CA6FED471EF3AA60A3E415EE4FE987DABA1",
			"297B858D9F752AB42D3BCA67EE0EB6DCD1C2B7B0DBE23397E66ADC272263F982"},

		{"1B22644A7BE026548810C378D0B2994EEFA6D2B9881803CB02CEFF865287D1B9",
			"F73C65EAD01C5126F28F442D087689BFA08E12763E0CEC1D35B01751FD735ED3",
			"F449A8376906482A84ED01479BD18882B919C140D638307F0C0934BA12590BDE"},
	}
)

func TestBase(t *testing.T) {
	if !g.IsOnCurve() {
		t.Fatal("base point not on curve")
	}
	if !testInOut(g) {
		t.Fatal("base point serialization failed")
	}
	gT := GetBasePoint()
	if !g.Equals(gT) {
		t.Fatal("GetBasePoint failed")
	}
	p := NewPoint(g.x, g.y)
	if !g.Equals(p) {
		t.Fatal("NewPoint failed")
	}
}

func TestInfinity(t *testing.T) {
	p1 := g.Mult(c.N)
	if !p1.Equals(Inf) {
		t.Fatal("n*G is not infinity")
	}
	if !p1.IsInf() {
		t.Fatal("IsInf failed")
	}
	if !testInOut(p1) {
		t.Fatal("infinity serialization failed")
	}
	p1 = g.Add(gm)
	if !p1.Equals(Inf) {
		t.Fatal("g-g is not infinity")
	}
	p1 = g.Add(Inf)
	if !p1.Equals(g) {
		t.Fatal("g+0 != g")
	}
	p1 = Inf.Mult(math.EIGHT)
	if !p1.Equals(Inf) {
		t.Fatal("8*0 != 0")
	}
}

func TestMult(t *testing.T) {
	p1 := g.Double()
	mult := func(n *math.Int) *Point {
		p := MultBase(n)
		if !p.IsOnCurve() {
			t.Fatal("point not on curve")
		}
		if !testInOut(p) {
			t.Fatal("point serialization failed")
		}
		return p
	}
	p2 := mult(math.TWO)
	if !p1.Equals(p2) {
		t.Fatal("mult failed")
	}
	mult(math.THREE)
	mult(math.SEVEN)
	mult(math.EIGHT)
}

func TestAdd(t *testing.T) {
	p1 := g.Double()
	p2 := g.Add(p1)
	p3 := p1.Add(g)
	if !p2.Equals(p3) {
		t.Fatal("p+g != g+p")
	}
	if !testInOut(p3) {
		t.Fatal("point serialization failed")
	}
	p1 = g.Double().Add(g)
	p2 = g.Mult(math.THREE)
	if !p1.Equals(p2) {
		t.Fatal("G+G+G != 3*G")
	}
	if !testInOut(p3) {
		t.Fatal("point serialization failed")
	}

	for n := 0; n < 32; n++ {
		a := nRnd(math.ZERO)
		b := nRnd(math.ZERO)
		c := a.Add(b)
		p := g.Mult(a)
		q := g.Mult(b)
		r := g.Mult(c)
		p1 = p.Add(q)
		p2 = q.Add(p)
		if !p1.Equals(p2) || !p1.Equals(r) {
			t.Fatal("a*G + b*G != (a+b)*G")
		}
	}
}

func TestDouble(t *testing.T) {
	pnt := g.Double()
	if !pnt.IsOnCurve() {
		t.Fatal("doubled point not on curve")
	}
	if !testInOut(pnt) {
		t.Fatal("point serialization failed")
	}
}

func TestNIST(t *testing.T) {
	for _, set := range tests {
		m := math.NewIntFromHex(set[0])
		x := math.NewIntFromHex(set[1])
		y := math.NewIntFromHex(set[2])
		p1 := &Point{x, y}
		p2 := g.Mult(m)
		if !p1.Equals(p2) {
			t.Fatal("failed nist case")
		}
		if !testInOut(p1) {
			t.Fatal("point serialization failed")
		}
	}
}

func TestCommute(t *testing.T) {
	dp := math.NewIntRndRange(math.THREE, c.N)
	dq := math.NewIntRndRange(math.THREE, c.N)
	p := MultBase(dp)
	q := MultBase(dq)
	p1 := p.Mult(dq)
	p2 := q.Mult(dp)
	if !p1.Equals(p2) {
		t.Fatal("failed commute")
	}
}

func TestInverse(t *testing.T) {
	for i := 0; i < 20; i++ {
		d := math.NewIntRndRange(math.THREE, c.N)
		di := d.ModInverse(c.N)
		x := nMul(d, di)
		if !x.Equals(math.ONE) {
			t.Fatal("failed inverse (1)")
		}
		d2 := d.Rsh(1)
		q := MultBase(d)
		q2 := MultBase(d2)
		if q2.Double().Equals(q) != (d.Bit(0) == 0) {
			t.Fatal("failed inverse (2)")
		}
	}
}

func testInOut(p *Point) bool {
	comprIn := p.x.Bit(0) == 0
	b := p.Bytes(comprIn)
	pp, comprOut, err := NewPointFromBytes(b)
	return (err == nil && pp.Equals(p) && comprIn == comprOut)
}
