package script

import (
	"encoding/hex"
	"fmt"
	"testing"
)

const (
	verbose = false
)

var (
	s = []string{
		"483045022074f35af390c41ef1f5395d11f6041cf55a6d7dab0acdac8ee746c1" +
			"f2de7a43b3022100b3dc3d916b557d378268a856b8f9a98b9afaf45442f5c9d7" +
			"26fce343de835a58012102c34538fc933799d972f55752d318c0328ca2bacccd" +
			"5c7482119ea9da2df70a2f76a9145e4ff47ceb3a51cdf7ddd80afc4acc5a692d" +
			"ac2d88ac",
		"004830450221009b65fcd0b0e3fcf038cc3ce8d1857e1b1e8e9050b56f9640fb" +
			"b80c0d2a65853f022039c215893a821de3927e513a417a811fc6cc5775dce809" +
			"c54483484994b9accf01483045022100db3b83a3b4462cfe63b2eab8e80e876b" +
			"ef93d801195893b19813bf83b7884de30220208fd3383df4a14a1f579499889b" +
			"537fedd2726e6e8a085b1d5f669d1f7eb80b014c695221021013a39b3e05020d" +
			"abd5e8942f03a65fa69ca7ce3e329c58f1f82b9515005a562102bf573d06fbe0" +
			"509a0ae897f58e746c9316b63b8a9b355a95339fd5fb51efdb852103035670a7" +
			"49d943639eb7bfc65e99167df83f0f98a0e251ac1387d5a5c015a3bb53ae",
	}
)

var (
	r = NewRuntime()
)

func TestExec(t *testing.T) {
	if verbose {
		r.CbStep = func(stack *Stack, stmt *Statement, rc int) {
			fmt.Println("==============================")
			fmt.Println("Statement: " + stmt.String())
			fmt.Printf("RC: %s\n", RcString[rc])
			fmt.Println("Stack:")
			for i, v := range stack.d {
				fmt.Printf("   %d: %s\n", stack.Len()-i-1, hex.EncodeToString(v.Bytes()))
			}
		}
	}
	scr, rc := Parse(s[0])
	if rc != RcOK {
		t.Fatal(fmt.Sprintf("Parse failed: rc=%s", RcString[rc]))
	}
	ok, rc := r.exec(scr)
	if rc != RcOK {
		if rc == RcNoTransaction {
			if verbose {
				fmt.Println("No transaction available")
			}
			return
		}
		t.Fatal(fmt.Sprintf("Exec failed: rc=%s", RcString[rc]))
	}
	if verbose {
		fmt.Printf("Result: %v\n", ok)
	}
}

func TestTemplate(t *testing.T) {
	scr, rc := Parse(s[0])
	if rc != RcOK {
		t.Fatal(fmt.Sprintf("Parse failed: rc=%s", RcString[rc]))
	}
	tpl, rc := scr.GetTemplate()
	if rc != RcOK {
		t.Fatal(fmt.Sprintf("GetTemplate failed: rc=%s", RcString[rc]))
	}
	if verbose {
		fmt.Printf("Template: %s\n", hex.EncodeToString(tpl))
	}
}
