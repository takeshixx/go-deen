package misc

import (
	"bytes"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"flag"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/takeshixx/deen/pkg/types"
)

const binInspectListLimit = 40

func byteEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	var counts [256]int
	for _, b := range data {
		counts[b]++
	}
	var e float64
	for _, count := range counts {
		if count == 0 {
			continue
		}
		p := float64(count) / float64(len(data))
		e -= p * math.Log2(p)
	}
	return e
}

func printStringList(w io.Writer, label, entryLabel string, values []string) {
	fmt.Fprintf(w, "%s: %d\n", label, len(values))
	for i, value := range values {
		if i >= binInspectListLimit {
			fmt.Fprintf(w, "%s-truncated: %d\n", label, len(values)-i)
			return
		}
		fmt.Fprintf(w, "%s: %s\n", entryLabel, value)
	}
}

func inspectELF(data []byte, w io.Writer) error {
	f, err := elf.NewFile(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(w, "format: ELF\n")
	fmt.Fprintf(w, "class: %s\n", f.Class)
	fmt.Fprintf(w, "data: %s\n", f.Data)
	fmt.Fprintf(w, "type: %s\n", f.Type)
	fmt.Fprintf(w, "machine: %s\n", f.Machine)
	fmt.Fprintf(w, "entry: 0x%x\n", f.Entry)
	fmt.Fprintf(w, "programs: %d\n", len(f.Progs))
	for _, p := range f.Progs {
		fmt.Fprintf(w, "program: %-14s off=0x%x vaddr=0x%x filesz=%d memsz=%d flags=%s\n",
			p.Type, p.Off, p.Vaddr, p.Filesz, p.Memsz, p.Flags)
	}
	fmt.Fprintf(w, "sections: %d\n", len(f.Sections))
	for _, s := range f.Sections {
		sectionData, _ := s.Data()
		fmt.Fprintf(w, "section: %-18s type=%-14s addr=0x%x off=0x%x size=%d entropy=%.2f\n",
			s.Name, s.Type, s.Addr, s.Offset, s.Size, byteEntropy(sectionData))
	}
	if libs, err := f.ImportedLibraries(); err == nil {
		printStringList(w, "libraries", "library", libs)
	}
	if symbols, err := f.ImportedSymbols(); err == nil {
		fmt.Fprintf(w, "imports: %d\n", len(symbols))
		for i, sym := range symbols {
			if i >= binInspectListLimit {
				fmt.Fprintf(w, "imports-truncated: %d\n", len(symbols)-i)
				break
			}
			library := sym.Library
			if sym.Version != "" {
				library = strings.TrimSpace(library + "@" + sym.Version)
			}
			if library == "" {
				fmt.Fprintf(w, "import: %s\n", sym.Name)
			} else {
				fmt.Fprintf(w, "import: %s (%s)\n", sym.Name, library)
			}
		}
	}
	return nil
}

func inspectPE(data []byte, w io.Writer) error {
	f, err := pe.NewFile(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(w, "format: PE\n")
	fmt.Fprintf(w, "machine: 0x%04x\n", f.Machine)
	fmt.Fprintf(w, "characteristics: 0x%04x\n", f.Characteristics)
	if oh, ok := f.OptionalHeader.(*pe.OptionalHeader32); ok {
		fmt.Fprintf(w, "entry: 0x%x\n", oh.AddressOfEntryPoint)
		fmt.Fprintf(w, "image-base: 0x%x\n", oh.ImageBase)
		fmt.Fprintf(w, "subsystem: %d\n", oh.Subsystem)
	}
	if oh, ok := f.OptionalHeader.(*pe.OptionalHeader64); ok {
		fmt.Fprintf(w, "entry: 0x%x\n", oh.AddressOfEntryPoint)
		fmt.Fprintf(w, "image-base: 0x%x\n", oh.ImageBase)
		fmt.Fprintf(w, "subsystem: %d\n", oh.Subsystem)
	}
	fmt.Fprintf(w, "sections: %d\n", len(f.Sections))
	for _, s := range f.Sections {
		sectionData, _ := s.Data()
		fmt.Fprintf(w, "section: %-8s va=0x%x raw=0x%x size=%d entropy=%.2f characteristics=0x%08x\n",
			s.Name, s.VirtualAddress, s.Offset, s.Size, byteEntropy(sectionData), s.Characteristics)
	}
	if libs, err := f.ImportedLibraries(); err == nil {
		printStringList(w, "libraries", "library", libs)
	}
	if symbols, err := f.ImportedSymbols(); err == nil {
		printStringList(w, "imports", "import", symbols)
	}
	return nil
}

func inspectMachO(data []byte, w io.Writer) error {
	if fat, err := macho.NewFatFile(bytes.NewReader(data)); err == nil {
		defer fat.Close()
		fmt.Fprintf(w, "format: Mach-O universal\n")
		fmt.Fprintf(w, "arches: %d\n", len(fat.Arches))
		for _, arch := range fat.Arches {
			fmt.Fprintf(w, "arch: cpu=%s subtype=%#x offset=0x%x size=%d\n",
				arch.Cpu, uint32(arch.SubCpu), arch.Offset, arch.Size)
		}
		return nil
	}

	f, err := macho.NewFile(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(w, "format: Mach-O\n")
	fmt.Fprintf(w, "cpu: %s\n", f.Cpu)
	fmt.Fprintf(w, "type: %s\n", f.Type)
	fmt.Fprintf(w, "flags: 0x%x\n", uint32(f.Flags))
	fmt.Fprintf(w, "sections: %d\n", len(f.Sections))
	for _, s := range f.Sections {
		sectionData, _ := s.Data()
		fmt.Fprintf(w, "section: %-16s segment=%-16s addr=0x%x off=0x%x size=%d entropy=%.2f\n",
			s.Name, s.Seg, s.Addr, s.Offset, s.Size, byteEntropy(sectionData))
	}
	if libs, err := f.ImportedLibraries(); err == nil {
		printStringList(w, "libraries", "library", libs)
	}
	if symbols, err := f.ImportedSymbols(); err == nil {
		printStringList(w, "imports", "import", symbols)
	}
	return nil
}

// NewPluginBinInspect creates a binary structure inspector for executable formats.
func NewPluginBinInspect() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "bininspect"
	p.Aliases = []string{"bin", "binary"}
	p.Category = "misc"
	p.Description = "Inspect executable binary structure for ELF, PE and Mach-O files."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return fmt.Errorf("empty input")
		}
		if err := inspectELF(data, w); err == nil {
			return nil
		}
		if err := inspectPE(data, w); err == nil {
			return nil
		}
		if err := inspectMachO(data, w); err == nil {
			return nil
		}
		return fmt.Errorf("unsupported binary format")
	}
	return p
}
