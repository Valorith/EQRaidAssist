package scanner

import (
	"testing"
)

func TestScanRaid(t *testing.T) {

	isStarted = true
	err := scanRaid()
	if err != nil {
		t.Fatalf("scanRaid: %s", err)
	}

}
