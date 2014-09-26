package cfm

import (
	"testing"
)

var testConf string = `
g1 1;
g2 2;
g3 3;


#asdfasdfadf
#asdfasdf

b {
	b1 1;
	b2 aaaa;
	b3 3; #asdfafda
}

#asdfadsfg
`

type testGConf struct {
	G1 int
	G2 int
	G3 int
}

type testBConf struct {
	B1 int
	B2 string
	B3 int
}

var gCmds = []Command{
	{"g1", "G1", 0, CommandSetInt},
	{"g2", "G2", 0, CommandSetInt},
	{"g3", "G3", 0, CommandSetInt},
}

var bCmds = []Command{
	{"b1", "B1", 0, CommandSetInt},
	{"b2", "B2", 0, CommandSetString},
	{"b3", "B3", 0, CommandSetInt},
}

var (
	gc *testGConf
	bc *testBConf
)

func TestParse(t *testing.T) {

	gc = new(testGConf)
	bc = new(testBConf)

	rc := NewRootContext()
	rc.AddConf(gc)
	rc.AddCommand(gCmds)

	tc := NewContext("b", CTX_ROOT_NAME)
	tc.AddCommand(bCmds)
	tc.AddConf(bc)

	c := new(Config)
	c.allContexts = contexts
	c.content = []byte(testConf)

	if err := c.Parse(); err != nil {
		t.Logf("%v", err)
	}

	if gc.G1 != 1 || gc.G2 != 2 || gc.G3 != 3 {
		t.Logf("parse err, %v", gc)
	}

	if bc.B1 != 1 || bc.B2 != "aaaa" || bc.B3 != 3 {
		t.Fatalf("Parse err, %v", bc)
	}
}
