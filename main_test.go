package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestExitCodes(t *testing.T) {
	old := os.Args
	t.Cleanup(func() { os.Args = old })

	tests := map[string]struct {
		env  map[string]string
		args []string
		code int
	}{
		"NoArgs":      {args: []string{"cmd"}, code: 2},
		"NoSuchFile":  {args: []string{"cmd", "testdata/nonexistent.elf"}, code: 74},
		"OnlyMeasure": {args: []string{"cmd", "--only-measure", "testdata/rgctl.elf"}, code: 0},
		"DryRun": {env: map[string]string{
			ghActionSha:     "7eb306e0a1a8c77922acb685a21e4f9854a6bce3",
			ghActionRepo:    "zoftko/felf",
			ghActionRefName: "dev",
		},
			args: []string{"cmd", "--dry-run", "testdata/rgctl.elf"},
			code: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Args = test.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			for key, val := range test.env {
				t.Setenv(key, val)
			}

			if status := cli(); status != test.code {
				t.Errorf("%s: expected exit code %d, got %d", name, test.code, status)
			}
		})
	}
}

func TestPush(t *testing.T) {
	old := os.Args
	t.Cleanup(func() { os.Args = old })

	ref := "main"
	repo := "zoftko/felf-cli"
	token := "magicalmisterytoken"
	path := "/api/measurements"
	sha := "7eb306e0a1a8c77922acb685a21e4f9854a6bce3"
	tests := map[string]struct {
		env      map[string]string
		response int
		code     int
	}{
		"MissingToken": {
			env: map[string]string{
				ghActionSha:     sha,
				ghActionRefName: ref,
				ghActionRepo:    repo,
			},
			response: 200,
			code:     4,
		},
		"MissingGithubSha": {
			env: map[string]string{
				ghActionSha:     "",
				ghActionRefName: ref,
			},
			response: 200,
			code:     3,
		},
		"MissingGithubRef": {
			env: map[string]string{
				ghActionRefName: "",
			},
			response: -1,
			code:     3,
		},
		"MissingGithubRepo": {
			env: map[string]string{
				ghActionRefName: ref,
				ghActionSha:     sha,
				ghActionRepo:    "",
			},
			response: -1,
			code:     3,
		},
		"ResponseNon200": {
			env: map[string]string{
				ghActionSha:     sha,
				ghActionRefName: ref,
				ghActionRepo:    repo,
				envToken:        token,
			},
			response: 401,
			code:     1,
		},
		"Response200": {
			env: map[string]string{
				ghActionSha:     sha,
				ghActionRefName: ref,
				ghActionRepo:    repo,
				envToken:        token,
			},
			response: 200,
			code:     0,
		},
		"MissingURL": {
			env: map[string]string{
				ghActionSha:     sha,
				ghActionRefName: ref,
				ghActionRepo:    repo,
				envToken:        token,
				envUrl:          "",
			},
			response: -1,
			code:     5,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Args = []string{"cmd", "testdata/rgctl.elf"}
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if content := request.Header.Get("Content-Type"); content != "application/json" {
					t.Errorf("expected Content-Type: %s, got: %s", "application/json", content)
				}

				if method := request.Method; method != http.MethodPost {
					t.Errorf("expected %s, got %s", method, http.MethodPost)
				}

				if repoHeader := request.Header.Get(headerRepo); repoHeader != repo {
					t.Errorf("expected %s: %s, got %s", headerRepo, repo, repoHeader)
				}

				expectedAuth := fmt.Sprintf("Bearer %s", token)
				if authorization := request.Header.Get("Authorization"); authorization != expectedAuth {
					t.Errorf("expected Authorization: %s, got: %s", expectedAuth, authorization)
				}

				var payload Payload
				_ = json.NewDecoder(request.Body).Decode(&payload)
				if payload.Sha != sha {
					t.Errorf("expected sha: %s, got :%s", sha, payload.Sha)
				}
				if payload.Ref != ref {
					t.Errorf("expected ref: %s, got :%s", ref, payload.Ref)
				}

				writer.WriteHeader(test.response)
			}))
			t.Cleanup(func() { server.Close() })

			t.Setenv("FELF_URL", server.URL+path)
			for env, val := range test.env {
				t.Setenv(env, val)
			}

			if code := cli(); code != test.code {
				t.Errorf("expected exit code %d, got: %d", test.code, code)
			}
		})
	}
}
