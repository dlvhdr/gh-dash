package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type SectionConfig struct {
	Title   string
	Filters string
	Repos   []string
}

const SectionsFileName = "sections.yml"

func ParseSectionsConfig() ([]SectionConfig, error) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(pwd, SectionsFileName))
	if err != nil {
		panic(err)
	}
	var sections []SectionConfig
	err = yaml.Unmarshal([]byte(data), &sections)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	return sections, nil
}
