package gopkg

import (
	"context"
	"errors"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/plugins/support"
	"github.com/stretchr/testify/assert"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func mockPluginCacheCheck(fn func(ic *importContext) bool) {
	before := pluginCached
	pluginCached = func(ic *importContext) bool {
		defer func() {
			pluginCached = before
		}()
		return fn(ic)
	}
}

func mockCommandRunner(fn func(cmd *exec.Cmd) error) {
	before := runGoCommand
	runGoCommand = func(cmd *exec.Cmd) error {
		defer func() {
			runGoCommand = before
		}()
		return fn(cmd)
	}
}

type expects struct {
	pkgPath  string
	fileName string
	rebuild  bool
	err      string
	build    bool
}

func TestGetPlugin(t *testing.T) {
	cases := map[string]struct {
		expects
		uri          string
		pluginExists bool
		cmdError     error
	}{
		"compile plugin package if it's not cached": {
			uri: "go://foo/bar",
			expects: expects{
				build:    true,
				pkgPath:  filepath.Join("foo", "bar"),
				fileName: support.AddPluginExtension("bar"),
			},
		},
		"do not build plugin if it's cached": {
			uri:          "go://foo/bar",
			pluginExists: true,
			expects: expects{
				pkgPath:  filepath.Join("foo", "bar"),
				fileName: support.AddPluginExtension("bar"),
			},
		},
		"rebuild plugin if rebuild flag present": {
			uri:          "go://foo/bar?rebuild=true",
			pluginExists: true,
			expects: expects{
				rebuild:  true,
				build:    true,
				pkgPath:  filepath.Join("foo", "bar"),
				fileName: support.AddPluginExtension("bar"),
			},
		},
		"return error on compile failure": {
			uri:      "go://foo/bar",
			cmdError: errors.New("dump error"),
			expects: expects{
				build:    true,
				pkgPath:  filepath.Join("foo", "bar"),
				err:      "failed to build plugin package (dump error)",
				fileName: support.AddPluginExtension("bar"),
			},
		},
	}

	for name, c := range cases {
		t.Run("should "+name, func(t *testing.T) {
			log.UseTestLogger(t)
			uri, err := url.Parse(c.uri)
			if err != nil {
				t.Fatal(err)
			}

			mockPluginCacheCheck(func(ic *importContext) bool {
				defer func() {
					if r := recover(); r != nil {
						t.Fatal(r)
					}
				}()
				assert.Equal(t, c.expects.fileName, ic.fileName)
				assert.Equal(t, c.expects.pkgPath, ic.pkgPath)
				assert.Equal(t, c.expects.rebuild, ic.rebuild)
				return c.pluginExists
			})

			mockCommandRunner(func(cmd *exec.Cmd) error {
				if !c.expects.build {
					t.Fatal("package build unexpected")
				}

				// output file path is penultimate
				outPath := cmd.Args[len(cmd.Args)-2]

				if !strings.Contains(outPath, c.expects.fileName) {
					t.Errorf("'%s' is not in '%s'", c.expects.fileName, outPath)
					t.Fatalf("output path should contain valid filename!")
				}
				return c.cmdError
			})

			out, err := GetPlugin(context.Background(), uri)
			if c.expects.err != "" {
				assert.EqualError(t, err, c.expects.err)
				return
			}

			assert.NoError(t, err)
			if !strings.Contains(out, c.expects.fileName) {
				t.Errorf("'%s' is not in '%s'", c.expects.fileName, out)
				t.Fatalf("library path should contain valid filename!")
			}
		})
	}
}
