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

	env := os.Environ()
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

func GetViewTabs(d Defaults) []Tab {
	tabs := []Tab{}

	for _, tab := range d.Tabs {
		if isSectionDisabled(d.DisableViews, ViewType(tab.Name)) {
			continue
		}
		tabs = append(tabs, tab)
	}

	return tabs
}

func isSectionDisabled(s []ViewType, str ViewType) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
