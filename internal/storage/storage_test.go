package storage

import (
	"github.com/go-gilbert/gilbert/internal/support/test"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func homeDirEnv() string {
	if runtime.GOOS == "windows" {
		return "USERPROFILE"
	}

	return "HOME"
}

func TestHomeDir(t *testing.T) {
	t.Log(os.Unsetenv(StoreVarName))
	hdir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot run test, user dir unavailable: %s", err)
		return
	}

	envName := homeDirEnv()
	cases := []struct {
		name string
		err  string
		want string
		mod  func()
	}{
		{
			name: "return default cache dir",
			want: filepath.Join(hdir, homeDirName),
		},
		{
			name: "override path from env var",
			want: "testdata",
			mod: func() {
				_ = os.Setenv(StoreVarName, "testdata")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.mod != nil {
				c.mod()
			}

			defer func() {
				t.Log(os.Unsetenv(StoreVarName))
				t.Log(os.Setenv(envName, hdir))
			}()
			result, err := home()
			if c.err != "" {
				test.AssertErrorContains(t, err, c.err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, c.want, result)
		})
	}
}

func TestPath(t *testing.T) {
	t.Log(os.Unsetenv(StoreVarName))
	t.Log(os.Setenv(StoreVarName, "testdata"))
	_, err := Path(Type(48))
	assert.Error(t, err)

	val, err := Path(Root, "foo")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join("testdata", "foo"), val)
}

func TestLocalPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip(err)
		return
	}

	_, err = LocalPath(Type(48))
	assert.Error(t, err)

	val, err := LocalPath(Root, "foo")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(cwd, homeDirName, "foo"), val)
}

func TestDelete(t *testing.T) {
	t.Log(os.Unsetenv(StoreVarName))
	t.Log(os.Setenv(StoreVarName, "testdata"))
	err := Delete(Type(48))
	assert.Error(t, err)

	if err := os.Mkdir("testdata", 0777); err != nil {
		t.Skip(err)
		return
	}

	err = Delete(Root)
	assert.NoErrorf(t, err, "error: %s", err)
}
