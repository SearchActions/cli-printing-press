package generator

import (
	"github.com/mvanhorn/cli-printing-press/internal/spec"
)

type Generator struct {
	Spec      *spec.APISpec
	OutputDir string
}

func New(s *spec.APISpec, outputDir string) *Generator {
	return &Generator{Spec: s, OutputDir: outputDir}
}

func (g *Generator) Generate() error {
	// TODO: implement in Task 3+4
	return nil
}
