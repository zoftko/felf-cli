package main

import (
	"debug/elf"
	"testing"
)

var rgctlMeasurements = Size{
	Text: 940,
	Data: 2,
	Bss:  6,
}

func compareSizes(t *testing.T, expect, got Size) {
	if expect.Text != got.Text {
		t.Errorf("expected text size %d got %d", expect.Text, got.Text)
	}
	if expect.Data != got.Data {
		t.Errorf("expected data size %d got %d", expect.Data, got.Data)
	}
	if expect.Bss != got.Bss {
		t.Errorf("expected bss size %d got %d", expect.Bss, got.Bss)
	}
}

func TestNewMeasurements(t *testing.T) {
	file, _ := elf.Open("testdata/rgctl.elf")
	result := newSize(file)
	compareSizes(
		t,
		rgctlMeasurements,
		result,
	)
}
