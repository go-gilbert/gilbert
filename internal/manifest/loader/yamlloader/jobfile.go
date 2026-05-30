package yamlloader

import (
	"errors"
	"fmt"
	"maps"

	"github.com/hashicorp/go-set/v3"

	"github.com/go-gilbert/gilbert/internal/manifest"
)

type includeDecl struct {
	filePath string
	location *manifest.ReferenceLocation
}

// yamlJobFile resembles a final job file accumulated from all imports
type yamlJobFile struct {
	// result is destination job file where all yamls decoded into.
	result *manifest.JobFile

	// version is YAML file version.
	version string

	// includes is list of files included by a file.
	includes []*includeDecl

	// builtinNamespaces contains reserved namespaces that can't be used
	// for importing custom plugins.
	builtinNamespaces *set.Set[string]
}

func (j *yamlJobFile) appendPlugins(newItems manifest.PluginImports) error {
	// Includes only permit merging plugins with same URL and import slug.
	aliasByURLs := make(map[string]string, len(j.result.Plugins))
	dst := j.result.Plugins

	if len(dst) == 0 {
		for k, plug := range newItems {
			if j.builtinNamespaces.Contains(k) {
				return fmt.Errorf("import namespace %q is reserved", k)
			}

			if prevAlias, ok := aliasByURLs[plug.URI]; ok {
				prevLoc := dst[prevAlias].Location
				return fmt.Errorf(
					"plugin was already imported under a different alias (previous declaration at %s:%s)",
					prevLoc.FileName, prevLoc.Range.Start,
				)
			}

			aliasByURLs[plug.URI] = k
		}

		j.result.Plugins = newItems
		return nil
	}

	for k, imp := range newItems {
		if j.builtinNamespaces.Contains(k) {
			return fmt.Errorf("import namespace %q is reserved", k)
		}

		// Check if same plugin was imported with a different alias.
		prevAlias, ok := aliasByURLs[imp.URI]
		if ok && prevAlias != k {
			prevLoc := dst[prevAlias].Location
			return fmt.Errorf(
				"plugin was already imported under a different alias (previous declaration at %s:%s)",
				prevLoc.FileName, prevLoc.Range.Start,
			)
		}

		// Check if same alias used by different plugin.
		dup, ok := dst[k]
		if ok {
			if dup.URI == imp.URI {
				continue
			}

			return fmt.Errorf(
				"import alias already taken by a different plugin %q (previous declaration at %s:%s)",
				k, dup.Location.FileName, dup.Location.Range.Start,
			)
		}

		dst[k] = imp
		aliasByURLs[imp.URI] = k
	}

	return nil
}

func (j *yamlJobFile) appendConsts(newItems map[string]any) error {
	if len(j.result.Consts) == 0 {
		j.result.Consts = newItems
		return nil
	}

	copyMapUniq(j.result.Consts, newItems)
	return nil
}

func (j *yamlJobFile) appendEnv(newItems map[string]*manifest.LazyValue) error {
	if len(j.result.Env) == 0 {
		j.result.Env = newItems
		return nil
	}

	// TODO: throw warning about duplicate env declaration
	maps.Copy(j.result.Env, newItems)
	return nil
}

func (j *yamlJobFile) appendInputs(newItems manifest.Inputs) error {
	if len(newItems) == 0 {
		return errors.New("empty inputs list")
	}

	if len(j.result.Inputs) == 0 {
		j.result.Inputs = newItems
		return nil
	}

	return copyMapWithCheck(j.result.Inputs, newItems, func(k string, dup *manifest.InputDefinition) error {
		loc := dup.Location
		return fmt.Errorf(
			"duplicate input block %q (previous declaration at %s:%s)", k,
			loc.FileName, loc.Range.Start,
		)
	})
}

func (j *yamlJobFile) appendTasks(newItems manifest.JobGroups) error {
	if len(newItems) == 0 {
		return nil
	}

	if j.result.Tasks == nil {
		j.result.Tasks = newItems
		return nil
	}

	return copyMapWithCheck(j.result.Tasks, newItems, func(k string, dup *manifest.JobGroup) error {
		loc := dup.Location
		return fmt.Errorf(
			"duplicate task block %q (previous declaration at %s:%s)", k,
			loc.FileName, loc.Range.Start,
		)
	})
}

func (j *yamlJobFile) appendMixins(newItems manifest.JobGroups) error {
	if len(newItems) == 0 {
		return nil
	}

	if j.result.Mixins == nil {
		j.result.Mixins = newItems
		return nil
	}

	return copyMapWithCheck(j.result.Mixins, newItems, func(k string, dup *manifest.JobGroup) error {
		loc := dup.Location
		return fmt.Errorf(
			"duplicate mixin block %q (previous declaration at %s:%s)", k,
			loc.FileName, loc.Range.Start,
		)
	})
}
