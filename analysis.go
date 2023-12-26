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
)

type Size struct {
	Text uint64
	Data uint64
	Bss  uint64
}

type Payload struct {
	Ref  string
	Sha  string
	Size Size
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

	return &Payload{
		Ref: ref,
		Sha: sha,
	}, nil
}

func pushPayload(token string, url string, payload *Payload) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
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

	if section.Type == elf.SHT_PROGBITS {
		if (section.Flags & elf.SHF_WRITE) == 0 {
			return codeCategory
		} else {
			return dataCategory
		}
	} else if section.Type == elf.SHT_NOBITS {
		return bssCategory
	}

	return unknownCategory
}
