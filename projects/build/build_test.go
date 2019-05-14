package build

import (
	"fmt"
	"testing"
)

func TestBuild(t *testing.T) {
	out, err := Build("wenzhenglin/test", "")
	if err != nil {
		t.Error("build err", err)
		return
	}
	fmt.Println("output", out)
}
