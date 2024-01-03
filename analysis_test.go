package main

import (
	"debug/elf"
	"net/http"
	"net/http/httptest"
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

func TestPushRedirect(t *testing.T) {
	path := "/api/analysis"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == path {
			http.Redirect(writer, request, "/ok", 302)
		}
		if request.URL.Path == "/ok" {
			writer.WriteHeader(200)
		}
	}))
	t.Cleanup(func() { server.Close() })

	payload := Payload{
		Repo: "zoftko/felf-cli",
	}
	response, _ := pushPayload("token", server.URL+path, &payload)

	if response.StatusCode != 302 {
		t.Errorf("expected http %d, got %d", 302, response.StatusCode)
	}
}
