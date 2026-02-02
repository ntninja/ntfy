package cmd

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"gopkg.in/yaml.v2"
	"heckel.io/ntfy/v2/util"
	"os"
)

// initConfigFileInputSourceFunc is like altsrc.InitInputSourceWithContext and altsrc.NewYamlSourceFromFlagFunc, but checks
// if the config flag is exists and only loads it if it does. If the flag is set and the file exists, it fails.
func initConfigFileInputSourceFunc(configFlag string, flags []cli.Flag, next cli.BeforeFunc) cli.BeforeFunc {
	return func(context *cli.Context) error {
		configFile := context.String(configFlag)
		if context.IsSet(configFlag) && !util.FileExists(configFile) {
			return fmt.Errorf("config file %s does not exist", configFile)
		} else if !context.IsSet(configFlag) && !util.FileExists(configFile) {
			return nil
		}
		inputSource, err := newYamlSourceFromFile(configFile, flags)
		if err != nil {
			return err
		}
		if err := altsrc.ApplyInputSourceValues(context, inputSource, flags); err != nil {
			return err
		}
		if next != nil {
			if err := next(context); err != nil {
				return err
			}
		}
		return nil
	}
}

func parseSingleYamlFile(rawConfig map[any]any, path string, flags []cli.Flag) error {
	// Parse values from files
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(b, rawConfig); err != nil {
		return err
	}

	// Resolve alias values based on flag configuration
	for _, f := range flags {
		flagName := f.Names()[0]
		for _, flagAlias := range f.Names()[1:] {
			if _, ok := rawConfig[flagAlias]; ok {
				rawConfig[flagName] = rawConfig[flagAlias]
			}
		}
	}

	return nil
}

// newYamlSourceFromFile creates a new Yaml InputSourceContext from a filepath.
//
// This function also maps aliases, so a .yml file can contain short options, or options with underscores
// instead of dashes. See https://github.com/binwiederhier/ntfy/issues/255.
func newYamlSourceFromFile(path string, flags []cli.Flag) (altsrc.InputSourceContext, error) {
	// Parse original source file
	rawConfig := make(map[any]any)
	if err := parseSingleYamlFile(rawConfig, path, flags); err != nil {
		return nil, err
	}

	// Process includes
	if includeOpt, ok := rawConfig["include"]; ok {
		// Extract `string` or `[]string`, erroring on wrong types
		var paths []string
		if path, ok := includeOpt.(string); ok {
			paths = append(paths, path)
		} else if maybePaths, ok := includeOpt.([]any); ok {
			for _, maybePath := range maybePaths {
				if path, ok := maybePath.(string); ok {
					paths = append(paths, path)
				} else {
					return nil, errors.New("config item “include” must be of type `string` or `[]string`")
				}
			}
		} else {
			return nil, errors.New("config item “include” must be of type `string` or `[]string`")
		}

		// Process YAML file at each path, so that each at the end `rawConfig`
		// will contain values from the last YAML file with highest precedence
		// and from the originally referenced file with lowest precedence
		for _, path := range paths {
			if err := parseSingleYamlFile(rawConfig, path, flags); err != nil {
				return nil, err
			}
		}

		delete(rawConfig, "include")
	}

	return altsrc.NewMapInputSource(path, rawConfig), nil
}
