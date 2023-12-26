package main

import (
	"debug/elf"
	"flag"
	"log"
)

const (
	CODE    SectionCategory = "code"
	DATA    SectionCategory = "data"
	BSS     SectionCategory = "bss"
	UNKNOWN SectionCategory = "unknown"
)

type Measurements struct {
	textSize uint64
	dataSize uint64
	bssSize  uint64
}

type SectionCategory string

func main() { cli() }

func cli() Measurements {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("Only a single positional argument is supported")
	}

	file, err := elf.Open(args[0])
	if err != nil {
		log.Fatalf(err.Error())
	}

	result := Measurements{}
	for _, section := range file.Sections {
		if section.Type == elf.SHT_NULL {
			continue
		}

		switch sectionCategory(section) {
		case CODE:
			result.textSize += section.Size
		case DATA:
			result.dataSize += section.Size
		case BSS:
			result.bssSize += section.Size
		}
	}

	return result
}

func sectionCategory(section *elf.Section) SectionCategory {
	if (section.Flags & elf.SHF_ALLOC) == 0 {
		return UNKNOWN
	}

	if section.Type == elf.SHT_PROGBITS {
		if (section.Flags & elf.SHF_WRITE) == 0 {
			return CODE
		} else {
			return DATA
		}
	} else if section.Type == elf.SHT_NOBITS {
		return BSS
	}

	return UNKNOWN
}
