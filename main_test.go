package main

import (
	"reflect"
	"testing"
)

func TestGetAtoms(t *testing.T) {
	input := []string{
		"system",
		"33760",
		"    1GOLD   AUS    1   0.084   0.498   1.100  0.0000  0.0000  0.0000",
		"    1GOLD   AUC    2   0.117   0.513   1.040  0.5329 -2.6386 -0.3969",
		"    1GOLD   AUS    3   0.084   0.970   1.100  0.0000  0.0000  0.0000",
		"    1GOLD   AUC    4   0.043   0.964   1.157 -2.2278 -3.5080 -1.8500",
		"    1GOLD   AUS    5   0.084   1.432   1.100  0.0000  0.0000  0.0000",
		"    1GOLD   AUC    6   0.066   1.365   1.102 -1.4282  0.3398 -0.6124",
		"    2PEC      C 6721   0.151   3.314   8.277 -0.0826 -0.2064 -0.1188",
		"    2PEC     C1 6722   0.231   3.398   8.176 -0.1558  0.5928  0.2292",
		"    2PEC      H 6723   0.046   3.333   8.254  2.4064  1.1910 -1.5404",
		"   12SOL     OW 8261   0.355   1.460   8.787 -0.3360 -0.6447 -0.0165",
		"   12SOL    HW1 8262   0.441   1.492   8.758  0.0606 -1.5891  0.0822",
		"   12SOL    HW2 8263   0.326   1.403   8.715 -1.2787  0.3683 -0.4391",
		"   5.48067   5.55646   9.09602",
	}
	expected := make(Atoms)
	expected["GOLD"] = []Atom{
		{
			resNum:   "1",
			atomName: "AUS",
			atomNum:  "1",
		},
		{
			resNum:   "1",
			atomName: "AUC",
			atomNum:  "2",
		},
		{
			resNum:   "1",
			atomName: "AUS",
			atomNum:  "3",
		},
		{
			resNum:   "1",
			atomName: "AUC",
			atomNum:  "4",
		},
		{
			resNum:   "1",
			atomName: "AUS",
			atomNum:  "5",
		},
		{
			resNum:   "1",
			atomName: "AUC",
			atomNum:  "6",
		},
	}
	expected["PEC"] = []Atom{
		{
			resNum:   "2",
			atomName: "C",
			atomNum:  "6721",
		},
		{
			resNum:   "2",
			atomName: "C1",
			atomNum:  "6722",
		},
		{
			resNum:   "2",
			atomName: "H",
			atomNum:  "6723",
		},
	}
	expected["SOL"] = []Atom{
		{
			resNum:   "12",
			atomName: "OW",
			atomNum:  "8261",
		},
		{
			resNum:   "12",
			atomName: "HW1",
			atomNum:  "8262",
		},
		{
			resNum:   "12",
			atomName: "HW2",
			atomNum:  "8263",
		},
	}

	actual := getAtoms(input)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
	}
}
