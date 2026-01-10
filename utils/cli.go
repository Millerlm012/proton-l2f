package utils

import (
	"flag"
	"fmt"
	"io/fs"
	"strings"
)

type Args struct {
	EnvPath      string
	IgnoreLabels []string
}

func ParseFlags() (*Args, error) {
	envPath := flag.String("env-path", "", "The absolute path to the env file on your machine.")
	ignoreLabels := flag.String("ignore", "", "A comma separated list of labels you'd like to ignore when migrating to Proton's folder structure. (e.g. Important,Library,Bills-Utilities )")
	flag.Parse()

	if err := parseEnv(envPath); err != nil {
		return nil, err
	}

	labels, err := parseIgnoreLabels(ignoreLabels)
	if err != nil {
		return nil, err
	}

	return &Args{
		EnvPath:      *envPath,
		IgnoreLabels: labels,
	}, nil
}

func parseEnv(envPath *string) error {
	if envPath == nil {
		return fmt.Errorf("env-path needs to be provided")
	}

	if !fs.ValidPath(*envPath) {
		return fmt.Errorf("%s is not a valid path", *envPath)
	}

	return nil
}

func parseIgnoreLabels(ignoreLabels *string) ([]string, error) {
	if ignoreLabels == nil {
		return nil, nil
	}

	parsed := strings.Split(*ignoreLabels, ",")
	if len(parsed) == 0 {
		return nil, fmt.Errorf("%s is empty - please ignore this flag if you want to migrate all labels to folders", *ignoreLabels)
	}

	return parsed, nil
}
