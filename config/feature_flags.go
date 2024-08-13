package config

import "os"

const FF_REPO_VIEW = "FF_REPO_VIEW"

func IsFeatureEnabled(name string) bool {
	_, ok := os.LookupEnv(name)
	return ok
}
