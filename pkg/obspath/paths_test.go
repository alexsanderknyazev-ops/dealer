package obspath

import (
	"testing"

	"github.com/dealer/dealer/pkg/obstest"
)

func TestIsProbe(t *testing.T) {
	for _, path := range Probes {
		if !IsProbe(path) {
			t.Fatalf("%s should be probe", path)
		}
	}
	if IsProbe(obstest.APIExamplePath) {
		t.Fatalf("%s should not be probe", obstest.APIExamplePath)
	}
}
