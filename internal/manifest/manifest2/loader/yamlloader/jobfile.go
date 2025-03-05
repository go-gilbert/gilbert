package yamlloader

import (
	"errors"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
)

type yamlJobFile struct {
	result   *manifest2.JobFile
	version  string
	includes []string
}

func (j *yamlJobFile) appendConsts(newItems map[string]any) error {
	if j.result.Consts == nil {
		j.result.Consts = newItems
		return nil
	}

	copyMapUniq(j.result.Consts, newItems)
	return nil
}

func (j *yamlJobFile) appendInputs(newItems manifest2.Inputs) error {
	if len(newItems) == 0 {
		return errors.New("empty inputs list")
	}

	if j.result.Inputs == nil {
		j.result.Inputs = newItems
		return nil
	}

	return copyMapWithCheck(j.result.Inputs, newItems, func(k string, dup *manifest2.InputDefinition) error {
		loc := dup.Location
		return fmt.Errorf(
			"duplicate input block %q (previous declaration at %s:%s)", k,
			loc.FileName, loc.Range.Start,
		)
	})
}

func (j *yamlJobFile) appendTasks(newItems manifest2.JobGroups) error {
	if len(newItems) == 0 {
		return nil
	}

	if j.result.Tasks == nil {
		j.result.Tasks = newItems
		return nil
	}

	return copyMapWithCheck(j.result.Tasks, newItems, func(k string, dup *manifest2.JobGroup) error {
		loc := dup.Location
		return fmt.Errorf(
			"duplicate task block %q (previous declaration at %s:%s)", k,
			loc.FileName, loc.Range.Start,
		)
	})
}

func (j *yamlJobFile) appendMixins(newItems manifest2.JobGroups) error {
	if len(newItems) == 0 {
		return nil
	}

	if j.result.Mixins == nil {
		j.result.Mixins = newItems
		return nil
	}

	return copyMapWithCheck(j.result.Mixins, newItems, func(k string, dup *manifest2.JobGroup) error {
		loc := dup.Location
		return fmt.Errorf(
			"duplicate mixin block %q (previous declaration at %s:%s)", k,
			loc.FileName, loc.Range.Start,
		)
	})
}

//type yamlJobFile2 struct {
//	Version  ast.Node   `yaml:"version"`
//	Includes []ast.Node `yaml:"includes"`
//	Inputs   ast.Node   `yaml:"inputs"`
//	Tasks    ast.Node   `yaml:"tasks"`
//}
//
//type jobFileLoader struct {
//	fileSet map[string]struct{}
//	queue []string
//	result  *manifest2.JobFile
//	diags   manifest2.Diagnostics
//}
//
//func newJobFileLoader(dst *manifest2.JobFile) *jobFileLoader {
//	return &jobFileLoader{
//		fileSet: make(map[string]struct{}),
//		result:  dst,
//	}
//}
//
//func (l *jobFileLoader) pushDiagnostic(diag *manifest2.Diagnostic) {
//	l.diags = append(l.diags, diag)
//}
//
//func (l *jobFileLoader) resolveFile(fileName string) error {
//	absPath, err := filepath.Abs(fileName)
//	if err != nil {
//		return err
//	}
//	if _, ok := l.fileSet[absPath]; ok {
//		return nil
//	}
//
//	src, err := os.ReadFile(absPath)
//	if err != nil {
//		return err
//	}
//
//	var jobFile yamlJobFile2
//	err = yaml.Unmarshal(src, &jobFile)
//	if err != nil {
//		// TODO: map yaml error to diagnostics
//		l.pushDiagnostic(newSimpleErrorDiagnostic(absPath, err))
//		return errors.New("cannot parse file")
//	}
//
//	_, err = readVersionNode(jobFile.Version)
//	if err != nil {
//		l.pushDiagnostic(newErrDiagnosticFromNode(absPath, jobFile.Version, err))
//		return errors.New("incompatible file version")
//	}
//
//	hasIncludeErrors := false
//	curDir := filepath.Dir(absPath)
//	for _, n := range jobFile.Includes {
//		incPath, err := readStringNode(n)
//		if err != nil {
//			hasIncludeErrors = true
//			l.pushDiagnostic(newErrDiagnosticFromNode(absPath, n, err))
//			continue
//		}
//
//		fpath := filepath.Join(curDir, incPath)
//		if err := l.resolveFile(fpath); err != nil {
//			l.pushDiagnostic(newErrDiagnosticFromNode(absPath, n, err))
//			hasIncludeErrors = true
//		}
//	}
//
//	hasParseErrors := l.appendFile(absPath, jobFile)
//	if hasIncludeErrors || hasParseErrors {
//		return errors.New("file parsed with errors")
//	}
//
//	return nil
//}
//
//func (l *jobFileLoader) appendFile(fileName string, f yamlJobFile2) bool {
//
//}
//
//func (l *jobFileLoader) readTasks(fileName string, node ast.Node) bool {
//	mn, ok := node.(*ast.MappingNode)
//	if !ok {
//		l.pushDiagnostic(newErrDiagnosticFromNode(fileName, node, errors.New("tasks node should be a dictionary")))
//		return false
//	}
//
//	hasError := false
//	for _, n := range mn.Values {
//		task := manifest2.JobGroup{
//			Inputs:   nil,
//			Jobs:     nil,
//			Location: nodeToLocation(fileName, n),
//		}
//
//		if kn, ok := n.Key.(*ast.StringNode); ok {
//			task.Name = kn.Value
//		} else {
//			hasError = true
//			l.pushDiagnostic(newErrDiagnosticFromNode(fileName, node, errors.New("task name should be a string")))
//		}
//
//		if kn, ok :=
//	}
//}
//
//func (l *jobFileLoader) load(fileName string) error {
//	absPath, err := filepath.Abs(fileName)
//	if err != nil {
//		return err
//	}
//
//	if err := l.resolveFile(absPath); err != nil {
//		if len(l.diags) == 0 {
//			return err
//		}
//
//		return l.diags
//	}
//
//	return nil
//}
//
//func LoadJobFile(fileName string) (*manifest2.JobFile, error) {
//	return nil, nil
//}
