package main

import (
	"bytes"
	"debug/elf"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	codeCategory    SectionCategory = "code"
	dataCategory    SectionCategory = "data"
	bssCategory     SectionCategory = "bss"
	unknownCategory SectionCategory = "unknown"
)

const (
	ghActionRefName = "GITHUB_REF_NAME"
	ghActionSha     = "GITHUB_SHA"
	ghActionRepo    = "GITHUB_REPOSITORY"
)

const (
	headerRepo = "X-Felf-Repo"
)

type Size struct {
	Text uint64 `json:"text,omitempty"`
	Data uint64 `json:"data,omitempty"`
	Bss  uint64 `json:"bss,omitempty"`
}

type Payload struct {
	Repo string `json:"-"`
	Ref  string `json:"ref,omitempty"`
	Sha  string `json:"sha,omitempty"`
	Size Size   `json:"size"`
}

type SectionCategory string

func newPayload() (*Payload, error) {
	ref := os.Getenv(ghActionRefName)
	if len(ref) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("%s not found", ghActionRefName))
	}

	sha := os.Getenv(ghActionSha)
	if len(sha) != 40 {
		return nil, fmt.Errorf("malformed sha, not 40 characters long")
	}

	repo := os.Getenv(ghActionRepo)
	if len(repo) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("%s not found", ghActionRepo))
	}

	return &Payload{
		Repo: repo,
		Ref:  ref,
		Sha:  sha,
	}, nil
}

func pushPayload(token string, url string, payload *Payload) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Add(headerRepo, payload.Repo)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return client.Do(req)
}

func newSize(file *elf.File) Size {
	result := Size{}
	for _, section := range file.Sections {
		if section.Type == elf.SHT_NULL {
			continue
		}

		switch category(section) {
		case codeCategory:
			result.Text += section.Size
		case dataCategory:
			result.Data += section.Size
		case bssCategory:
			result.Bss += section.Size
		}
	}

	return result
}

func category(section *elf.Section) SectionCategory {
	if (section.Flags & elf.SHF_ALLOC) == 0 {
		return unknownCategory
	}

	if section.Type != elf.SHT_NOBITS {
		if (section.Flags & elf.SHF_WRITE) == 0 {
			return codeCategory
		} else {
			return dataCategory
		}
	} else {
		return bssCategory
	}
}
