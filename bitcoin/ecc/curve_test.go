package ecc

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"encoding/hex"
	"fmt"
	"github.com/bfix/gospel/math"
	"math/big"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	public test method

func TestCurve(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("ecc/curve Test")
	fmt.Println("********************************************************")

	g := &point{curve_gx, curve_gy}
	gm := &point{curve_gx, new(big.Int).Neg(curve_gy)}

	fmt.Print("Checking if base point 'g' is on curve: ")
	if !isOnCurve(g) {
		fmt.Printf("Failed")
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(g) {
		t.Fail()
		return
	}

	fmt.Print("Computing infinity '0 = n*g': ")
	p1 := scalarMult(g, curve_n)
	if !isEqual(p1, inf) {
		fmt.Printf("Failed: %s\n", p1.emit())
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(p1) {
		t.Fail()
		return
	}

	fmt.Print("Computing infinity '0 = (-g) + g': ")
	p1 = add(g, gm)
	if !isEqual(p1, inf) {
		fmt.Printf("Failed: %s\n", p1.emit())
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}

	fmt.Println("Checking infinity:")
	fmt.Print("    0+p = p: ")
	p1 = add(g, inf)
	if !isEqual(p1, g) {
		fmt.Printf("Failed: %s != %s\n", p1.emit(), g.emit())
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	fmt.Print("    k*0 = 0: ")
	p1 = scalarMult(inf, math.EIGHT)
	if !isEqual(p1, inf) {
		fmt.Printf("Failed: %s != (0,0)\n", p1.emit())
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}

	fmt.Print("Checking for 'double(x) == scalarMult(x,2)': ")
	p1 = double(g)
	p2 := scalarMult(g, math.TWO)
	if !isEqual(p1, p2) {
		fmt.Println("Failed")
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(p2) {
		t.Fail()
		return
	}

	fmt.Print("Checking for 'p+q = q+p': ")
	p1 = double(g)
	p2 = add(g, p1)
	p3 := add(p1, g)
	if !isEqual(p2, p3) {
		fmt.Println("Failed")
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(p3) {
		t.Fail()
		return
	}

	fmt.Print("Checking for 'add(double(x),x) == scalarMult(x,3)': ")
	p1 = add(double(g), g)
	p2 = scalarMult(g, math.THREE)
	if !isEqual(p1, p2) {
		fmt.Println("Failed")
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(p3) {
		t.Fail()
		return
	}

	fmt.Print("Checking if point '2*g' is on curve: ")
	pnt := double(g)
	if !isOnCurve(pnt) {
		fmt.Println("Failed")
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(pnt) {
		t.Fail()
		return
	}

	fmt.Print("Checking if point '3*g' is on curve: ")
	pnt = scalarMult(g, math.THREE)
	if !isOnCurve(pnt) {
		fmt.Println("Failed")
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(pnt) {
		t.Fail()
		return
	}

	fmt.Print("Checking if point '7*g' is on curve: ")
	pnt = scalarMult(g, math.SEVEN)
	if !isOnCurve(pnt) {
		fmt.Println("Failed")
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(pnt) {
		t.Fail()
		return
	}

	fmt.Print("Checking if point '8*g' is on curve: ")
	pnt = scalarMult(g, math.EIGHT)
	if !isOnCurve(pnt) {
		fmt.Println("Failed")
		t.Fail()
		return
	} else {
		fmt.Println("O.K.")
	}
	if !testInOut(pnt) {
		t.Fail()
		return
	}

	fmt.Println("Checking curve: 'aG + bG = cG if a + b = c")
	fmt.Print("    ")
	failed := false
	for n := 0; n < 32; n++ {
		a := n_rnd(math.ZERO)
		b := n_rnd(math.ZERO)
		c := new(big.Int).Add(a, b)
		p := scalarMult(g, a)
		q := scalarMult(g, b)
		r := scalarMult(g, c)
		p1 = add(p, q)
		p2 = add(q, p)

		if !isEqual(p1, p2) {
			failed = true
			fmt.Print("-")
			fmt.Print("\nFAIL: ")
			p1.emit()
			fmt.Print("\nFAIL: ")
			p2.emit()
			fmt.Println()
		} else {
			if !isEqual(p1, r) {
				failed = true
				fmt.Print("-")
				fmt.Print("\nFAIL: ")
				p1.emit()
				fmt.Print("\nFAIL: ")
				r.emit()
				fmt.Println()
			} else {
				fmt.Print("+")
			}
		}
	}
	if failed {
		t.Fail()
		return
		fmt.Println(" Failed")
	} else {
		fmt.Println(" O.K.")
	}

	fmt.Println("Checking NIST test values:")
	fmt.Print("    ")
	var tests = [][]string{
		[]string{"AA5E28D6A97A2479A65527F7290311A3624D4CC0FA1578598EE3C2613BF99522",
			"34F9460F0E4F08393D192B3C5133A6BA099AA0AD9FD54EBCCFACDFA239FF49C6",
			"0B71EA9BD730FD8923F6D25A7A91E7DD7728A960686CB5A901BB419E0F2CA232"},

		[]string{"7E2B897B8CEBC6361663AD410835639826D590F393D90A9538881735256DFAE3",
			"D74BF844B0862475103D96A611CF2D898447E288D34B360BC885CB8CE7C00575",
			"131C670D414C4546B88AC3FF664611B1C38CEB1C21D76369D7A7A0969D61D97D"},

		[]string{"6461E6DF0FE7DFD05329F41BF771B86578143D4DD1F7866FB4CA7E97C5FA945D",
			"E8AECC370AEDD953483719A116711963CE201AC3EB21D3F3257BB48668C6A72F",
			"C25CAF2F0EBA1DDB2F0F3F47866299EF907867B7D27E95B3873BF98397B24EE1"},

		[]string{"376A3A2CDCD12581EFFF13EE4AD44C4044B8A0524C42422A7E1E181E4DEECCEC",
			"14890E61FCD4B0BD92E5B36C81372CA6FED471EF3AA60A3E415EE4FE987DABA1",
			"297B858D9F752AB42D3BCA67EE0EB6DCD1C2B7B0DBE23397E66ADC272263F982"},

		[]string{"1B22644A7BE026548810C378D0B2994EEFA6D2B9881803CB02CEFF865287D1B9",
			"F73C65EAD01C5126F28F442D087689BFA08E12763E0CEC1D35B01751FD735ED3",
			"F449A8376906482A84ED01479BD18882B919C140D638307F0C0934BA12590BDE"},
	}
	failed = false
	for _, set := range tests {
		m := fromHex(set[0])
		x := fromHex(set[1])
		y := fromHex(set[2])
		p1 = &point{x, y}
		p2 = scalarMult(g, m)

		if !isEqual(p1, p2) {
			failed = true
			fmt.Printf("-\nFAIL: %s\n", p1.emit())
			fmt.Printf("FAIL: %s\n", p2.emit())
		} else {
			fmt.Print("+")
		}
		if !testInOut(p1) {
			t.Fail()
			return
		}
	}
	if failed {
		t.Fail()
		fmt.Println(" Failed")
		return
	} else {
		fmt.Println(" O.K.")
	}
}

///////////////////////////////////////////////////////////////////////
// helper methods: print a point

func (p *point) emit() string {
	return "(" + p.x.String() + "," + p.y.String() + ")"
}

// test binary conversion for point
func testInOut(p *point) bool {
	cmpr := p.x.Bit(0) == 0
	b := pointAsBytes(p, cmpr)
	pp, err := pointFromBytes(b)
	rc := (err == nil && isEqual(pp, p))
	if !rc {
		fmt.Printf (">> %s\n", p.emit()) 
		fmt.Printf (">> %v\n", cmpr) 
		fmt.Printf (">> %s\n", hex.EncodeToString(b)) 
		fmt.Printf (">> %s\n", pp.emit()) 
		fmt.Println ("BinRep() failed!")
	}
	return rc
}
