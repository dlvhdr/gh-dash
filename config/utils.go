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
