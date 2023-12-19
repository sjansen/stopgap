package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lmika/goseq/seqdiagram"
	"github.com/alecthomas/kingpin/v2"
)

var (
	files = kingpin.Arg("INPUT...", "").Required().ExistingFiles()
)

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	kingpin.Parse()

	for _, src := range *files {
		ext := strings.ToLower(filepath.Ext(src))
		dst := strings.TrimSuffix(src, ext) + ".svg"

		r, err := os.Open(src)
		if err != nil {
			die(err)
		}

		diagram, err := seqdiagram.ParseDiagram(r, src)
		if err != nil {
			die(err)
		}

		w, err := os.Create(dst)
		if err != nil {
			die(err)
		}

		err = diagram.WriteSVG(w)
		if err != nil {
			die(err)
		}
	}
}
