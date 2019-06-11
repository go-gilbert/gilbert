package runner

import (
	"errors"
	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/actions"
	"github.com/go-gilbert/gilbert/manifest"
	"github.com/go-gilbert/gilbert/support/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if err := actions.HandleFunc(testAction, newTestAction); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

type results struct {
	done       bool
	cancel     bool
	startTime  time.Time
	endTime    time.Time
	cancelTime time.Time
}

func TestTaskRunner_Run(t *testing.T) {
	r := results{}
	cases := map[string]struct {
		skip     bool
		taskName string
		m        manifest.Manifest
		err      string
		before   func(t *testing.T, tr *TaskRunner, l *test.Log)
		after    func(t *testing.T, tr *TaskRunner, l *test.Log)
	}{
		"error if task not exists": {
			skip:     true,
			taskName: "foo",
			err:      "task 'foo' doesn't exists",
		},
		"error if job is empty": {
			skip:     true,
			taskName: "foo",
			err:      "no task handler defined",
			m:        manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{sdk.Job{}}}},
		},
		"error if action not exists": {
			skip:     true,
			taskName: "foo",
			err:      "no such action handler: 'foo'",
			m:        manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{sdk.Job{ActionName: "foo"}}}},
			after: func(t *testing.T, tr *TaskRunner, l *test.Log) {
				l.AssertMessage("task context was not set")
			},
		},
		"error if action returned error": {
			skip:     true,
			taskName: "foo",
			err:      "fail",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				sdk.Job{ActionName: testAction, Params: sdk.ActionParams{"err": "fail"}, Async: true},
				sdk.Job{ActionName: testAction, Params: sdk.ActionParams{"err": "fail"}},
			}}},
		},
		"error if action factory returned error": {
			skip:     true,
			taskName: "foo",
			err:      "fail",
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testBadAction", func(sdk.ScopeAccessor, sdk.ActionParams) (sdk.ActionHandler, error) {
					return nil, errors.New("foo")
				})
			},
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				sdk.Job{ActionName: "testBadAction"},
			}}},
		},
		"wait until async task complete": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				sdk.Job{ActionName: "testAsync", Async: true},
			}}},
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testAsync", func(sdk.ScopeAccessor, sdk.ActionParams) (sdk.ActionHandler, error) {
					return &asyncTestHandle{data: &r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				assert.True(t, r.done)
			},
		},
		"respect exec condition": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				sdk.Job{ActionName: "testTimeout", Condition: "badcommand"},
			}}},
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testTimeout", func(sdk.ScopeAccessor, sdk.ActionParams) (sdk.ActionHandler, error) {
					return &asyncTestHandle{data: &r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				assert.Falsef(t, r.done, "task shouldn't start")
			},
		},
		"skip job if condition expression is bad": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				sdk.Job{ActionName: "testBadConditionHook", Condition: "{{bad}} condition"},
			}}},
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testBadConditionHook", func(sdk.ScopeAccessor, sdk.ActionParams) (sdk.ActionHandler, error) {
					return &asyncTestHandle{data: &r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				assert.Falsef(t, r.done, "task shouldn't start")
			},
		},
		"run job if expression returns OK result": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				sdk.Job{ActionName: "testOKConditionHook", Condition: "echo {{msg}}", Vars: sdk.Vars{"msg": "hello"}},
			}}},
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testOKConditionHook", func(sdk.ScopeAccessor, sdk.ActionParams) (sdk.ActionHandler, error) {
					return &asyncTestHandle{data: &r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				assert.Truef(t, r.done, "task was not started")
			},
		},
		"respect timeout": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				sdk.Job{ActionName: "testTimeout", Delay: sdk.Period(800)},
			}}},
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testTimeout", func(sdk.ScopeAccessor, sdk.ActionParams) (sdk.ActionHandler, error) {
					return &asyncTestHandle{data: &r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				assert.Truef(t, r.done, "task didn't finished")
				diff := uint((r.endTime.Sub(r.startTime).Seconds()) * 1000)
				const expectedDiff = 800 + 100
				const permMargin = 100
				if (diff < expectedDiff-permMargin) || diff > (expectedDiff+permMargin) {
					t.Fatalf("task timeout is invalid (%d msec), want %d", diff, expectedDiff)
				}
			},
		},
		"respect deadline": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				sdk.Job{ActionName: "testDeadline", Deadline: sdk.Period(10)},
			}}},
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testDeadline", func(sdk.ScopeAccessor, sdk.ActionParams) (sdk.ActionHandler, error) {
					return &asyncTestHandle{data: &r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				assert.Truef(t, r.cancel, "task didn't receive cancel callback on deadline")
				diff := uint((r.cancelTime.Sub(r.startTime).Seconds()) * 1000)
				const deadlineTime = uint(10)
				assert.Equalf(t, deadlineTime, diff, "task cancel time mismatch")
			},
		},
		"execute mixins": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{
				Mixins: manifest.Mixins{
					"mx1": manifest.Mixin{
						sdk.Job{Description: "start {{foo}}", ActionName: "testMixinExec1"},
					},
				},
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						sdk.Job{ActionName: testAction},
						sdk.Job{MixinName: "mx1", Vars: sdk.Vars{"foo": "bar"}},
					},
				},
			},
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testMixinExec1", func(sc sdk.ScopeAccessor, ap sdk.ActionParams) (sdk.ActionHandler, error) {
					vars := sc.Vars()
					assert.NotEmptyf(t, vars["foo"], "parent job variables wasn't passed to mixin")
					return &asyncTestHandle{data: &r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				assert.Truef(t, r.done, "task from mixin was not processed")
			},
		},
		"report error if mixin job failed": {
			skip:     true,
			taskName: "foo",
			err:      "mixin job fail",
			m: manifest.Manifest{
				Mixins: manifest.Mixins{
					"mx1": manifest.Mixin{
						sdk.Job{ActionName: testAction, Params: sdk.ActionParams{"err": "mixin job fail"}},
					},
				},
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						sdk.Job{ActionName: testAction},
						sdk.Job{MixinName: "mx1"},
					},
				},
			},
		},
		"report error if mixin doesn't exists": {
			skip:     true,
			taskName: "foo",
			err:      "mixin 'mx1' doesn't exists",
			m: manifest.Manifest{
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						sdk.Job{ActionName: testAction},
						sdk.Job{MixinName: "mx1"},
					},
				},
			},
		},
		"not fail if expression in mixin job desc is malformed": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{
				Mixins: manifest.Mixins{
					"mx1": manifest.Mixin{
						sdk.Job{Description: "bad {{expr}}", ActionName: testAction},
					},
				},
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						sdk.Job{MixinName: "mx1"},
					},
				},
			},
		},
		"check if subtask exists": {
			skip:     true,
			taskName: "t1",
			err:      "task 't2' doesn't exists",
			m: manifest.Manifest{
				Tasks: manifest.TaskSet{
					"t1": manifest.Task{
						sdk.Job{TaskName: "t2"},
					},
				},
			},
		},
		"execute subtask": {
			skip:     true,
			taskName: "foo",
			m: manifest.Manifest{
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						sdk.Job{ActionName: testAction},
						sdk.Job{TaskName: "bar", Vars: sdk.Vars{"foo": "bar"}},
					},
					"bar": manifest.Task{
						sdk.Job{ActionName: testAction, Async: true},
						sdk.Job{Description: "start {{foo}}", ActionName: "testSubTaskExec1"},
					},
				},
			},
			before: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				_ = actions.HandleFunc("testSubTaskExec1", func(sc sdk.ScopeAccessor, ap sdk.ActionParams) (sdk.ActionHandler, error) {
					vars := sc.Vars()
					assert.NotEmptyf(t, vars["foo"], "parent job variables wasn't passed to mixin")
					return &asyncTestHandle{data: &r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log) {
				assert.Truef(t, r.done, "subtask was not processed")
			},
		},
		"return subtask errors": {
			skip:     true,
			taskName: "foo",
			err:      "task 'foo' returned an error on step 2: fail (sub-task step 1)",
			m: manifest.Manifest{
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						sdk.Job{ActionName: testAction},
						sdk.Job{TaskName: "bar"},
					},
					"bar": manifest.Task{
						sdk.Job{ActionName: testAction, Params: sdk.ActionParams{"err": "fail"}},
					},
				},
			},
		},
	}

	for name, c := range cases {
		if c.skip {
			//continue
		}
		r = results{}
		l := &test.Log{T: t}
		t.Run("should "+name, func(t *testing.T) {
			tr := NewTaskRunner(&c.m, "", l)
			r.startTime = time.Now()
			if c.before != nil {
				c.before(t, tr, l)
			}
			err := tr.Run(c.taskName, nil)
			if c.err != "" {
				test.AssertErrorContains(t, err, c.err)
			} else {
				assert.NoError(t, err)
			}
			if c.after != nil {
				c.after(t, tr, l)
			}
		})
	}
}

///////////////////
// Test Fixtures //
///////////////////

const testAction = "testAction"

type asyncTestHandle struct {
	data *results
}

func (t *asyncTestHandle) Call(sdk.JobContextAccessor, sdk.JobRunner) error {
	time.Sleep(time.Millisecond * 100)
	t.data.done = true
	t.data.endTime = time.Now()
	return nil
}

func (t *asyncTestHandle) Cancel(sdk.JobContextAccessor) error {
	t.data.cancel = true
	t.data.cancelTime = time.Now()
	return nil
}

func newTestAction(_ sdk.ScopeAccessor, p sdk.ActionParams) (sdk.ActionHandler, error) {
	ac := &testActionHandler{}
	err := p.Unmarshal(ac)
	return ac, err
}

type testActionHandler struct {
	Err     string     `mapstructure:"err"`
	Timeout sdk.Period `mapstructure:"timeout"`
}

func (t *testActionHandler) Call(sdk.JobContextAccessor, sdk.JobRunner) error {
	if t.Timeout == 0 {
		t.Timeout = sdk.Period(100)
	}

	time.Sleep(t.Timeout.ToDuration())
	if t.Err == "" {
		return nil
	}

	return errors.New(t.Err)
}

func (t *testActionHandler) Cancel(sdk.JobContextAccessor) error {
	return nil
}
