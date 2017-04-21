package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func getType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "map"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case int:
		return "int"
	case float64:
		return "float64"
	case bool:
		return "bool"
	default:
		return "%"
	}
}

func compare(a, b interface{}, depth int, w io.Writer) bool {
	at := getType(a)
	bt := getType(b)
	fmt.Fprintf(w, "%d| %s\n", depth, at)
	if at != bt {
		fmt.Fprintf(w, "Type mismatch: %s != %s\n", at, bt)
		return false
	}
	switch at {
	case "array":
		aa := a.([]interface{})
		ba := b.([]interface{})
		for i, v := range aa {
			fmt.Fprintf(w, "%d| [%d]\n", depth, i)
			if !compare(v, ba[i], depth+1, w) {
				return false
			}
		}
	case "map":
		am := a.(map[string]interface{})
		bm := b.(map[string]interface{})
		for k, v := range am {
			fmt.Fprintf(w, "%d| ['%s']\n", depth, k)
			x, ok := bm[k]
			if !ok {
				fmt.Fprintf(w, "Key: %s=%v\n", k, v)
				return false
			}
			if !compare(v, x, depth+1, w) {
				return false
			}
		}
	case "string":
		as := a.(string)
		bs := b.(string)
		fmt.Fprintf(w, "%d|   ='%s'\n", depth, as)
		return as == bs
	case "int":
		ai := a.(int)
		bi := b.(int)
		fmt.Fprintf(w, "%d|   =%d\n", depth, ai)
		return ai == bi
	case "float64":
		af := a.(float64)
		bf := b.(float64)
		fmt.Fprintf(w, "%d|   =%f\n", depth, af)
		return af == bf
	case "bool":
		ab := a.(bool)
		bb := b.(bool)
		fmt.Fprintf(w, "%d|   =%v\n", depth, ab)
		return ab == bb
	default:
		panic("compare")
	}
	return true
}

func prepare(i interface{}, w io.Writer) (interface{}, bool) {
	if getType(i) != "%" {
		return i, true
	}
	b, err := json.Marshal(i)
	if err != nil {
		fmt.Fprintln(w, "ERROR: "+err.Error())
		return nil, false
	}
	var ii interface{}
	if err = json.Unmarshal(b, &ii); err != nil {
		fmt.Fprintln(w, "ERROR: "+err.Error())
		return nil, false
	}
	return ii, true
}

func checkJSON(a, b interface{}) (bool, string) {
	buf := new(bytes.Buffer)
	am, ok := prepare(a, buf)
	if !ok {
		return false, buf.String()
	}
	bm, ok := prepare(b, buf)
	if !ok {
		return false, buf.String()
	}
	rc := compare(am, bm, 0, buf)
	if !rc {
		fmt.Fprintf(buf, "IN: %v\n", am)
		fmt.Fprintf(buf, "OUT: %v\n", bm)
	}
	return rc, buf.String()
}
