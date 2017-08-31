package dqrack

import (
	"fmt"
	"testing"
)

func TestTagParser(t *testing.T) {
	// parse tags like dq:"someKey,etc" into component parts
	to := parseDqTag("key,omitempty,inline,prefix")
	if to.Name != "key" {
		t.Error("tag wasn't parsed correctly:", to)
		return
	}

	fmt.Println(to)

	to = parseDqTag("-,omitempty,inline,prefix")
	if !to.Ignore {
		t.Error("tag wasn't parsed correctly to ignore:", to)
		return
	}

	fmt.Println(to)
}

func TestEdgeValue(t *testing.T) {

}
