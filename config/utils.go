package config

import (
	"fmt"
	"os"
)

func (cfg Config) GetFullScreenDiffPagerEnv() []string {
	diff := cfg.Pager.Diff
	if diff == "" {
		diff = "less"
	}
	if diff == "delta" {
		diff = "delta --paging always"
	}

	var env = os.Environ()
	env = append(
		env,
		"LESS=CRX",
		fmt.Sprintf(
			"GH_PAGER=%s",
			diff,
		),
	)

	return env
}

func (cfg PrsSectionConfig) ToSectionConfig() SectionConfig {
	return SectionConfig{
		Title:   cfg.Title,
		Filters: cfg.Filters,
		Limit:   cfg.Limit,
		Type:    cfg.Type,
	}
}

func (cfg IssuesSectionConfig) ToSectionConfig() SectionConfig {
	return SectionConfig{
		Title:   cfg.Title,
		Filters: cfg.Filters,
		Limit:   cfg.Limit,
	}
}

func MergeColumnConfigs(defaultCfg, sectionCfg ColumnConfig) ColumnConfig {
	colCfg := defaultCfg
	if sectionCfg.Width != nil {
		colCfg.Width = sectionCfg.Width
	}
	if sectionCfg.Hidden != nil {
		colCfg.Hidden = sectionCfg.Hidden
	}
	return colCfg
}
