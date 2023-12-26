package main

import (
	"os"
	"testing"
)

func compareMeasurements(t *testing.T, expect, got Measurements) {
	if expect.textSize != got.textSize {
		t.Errorf("expected text size %d got %d", expect.textSize, got.textSize)
	}
	if expect.dataSize != got.dataSize {
		t.Errorf("expected data size %d got %d", expect.dataSize, got.dataSize)
	}
	if expect.bssSize != got.bssSize {
		t.Errorf("expected bss size %d got %d", expect.bssSize, got.bssSize)
	}
}

func TestCliMeasurements(t *testing.T) {
	defer func(old []string) { os.Args = old }(os.Args)
	os.Args = []string{"cmd", "testdata/rgctl.elf"}

	result := cli()
	compareMeasurements(
		t,
		Measurements{
			textSize: 940,
			dataSize: 2,
			bssSize:  6,
		},
		result,
	)
}
