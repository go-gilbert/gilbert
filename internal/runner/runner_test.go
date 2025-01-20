package runner

import (
	"errors"
	"testing"
	"time"

	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/runner/job"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/internal/support/test"
	"github.com/stretchr/testify/assert"
)

type results struct {
	done       bool
	cancel     bool
	startTime  time.Time
	endTime    time.Time
	cancelTime time.Time
}

func TestTaskRunner_Run(t *testing.T) {
	cases := map[string]struct {
		skip     bool
		taskName string
		m        manifest.Manifest
		err      string
		before   func(t *testing.T, tr *TaskRunner, hs *HandlerSet, r *results)
		after    func(t *testing.T, tr *TaskRunner, l *test.Log, r *results)
	}{
		"error if task not exists": {
			taskName: "foo",
			err:      `task "foo" doesn't exists`,
		},
		"error if job is empty": {
			taskName: "foo",
			err:      "no task handler defined",
			m:        manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{manifest.Job{}}}},
		},
		"error if action not exists": {
			taskName: "foo",
			err:      `no such action handler: "foo"`,
			m:        manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{manifest.Job{ActionName: "foo"}}}},
			after: func(t *testing.T, tr *TaskRunner, l *test.Log, r *results) {
				l.AssertMessage("task context was not set")
			},
		},
		"error if action returned error": {
			taskName: "foo",
			err:      "fail",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				manifest.Job{ActionName: testAction, Params: manifest.ActionParams{"err": "fail"}, Async: true},
				manifest.Job{ActionName: testAction, Params: manifest.ActionParams{"err": "fail"}},
			}}},
		},
		"error if action factory returned error": {
			taskName: "foo",
			err:      "fail",
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testBadAction", func(*scope.Scope, manifest.ActionParams) (ActionHandler, error) {
					return nil, errors.New("foo")
				})
			},
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				manifest.Job{ActionName: "testBadAction"},
			}}},
		},
		"wait until async task complete": {
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				manifest.Job{ActionName: "testAsync", Async: true},
			}}},
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testAsync", func(*scope.Scope, manifest.ActionParams) (ActionHandler, error) {
					return &asyncTestHandle{data: r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log, r *results) {
				assert.True(t, r.done)
			},
		},
		"respect exec condition": {
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				manifest.Job{ActionName: "testTimeout", Condition: "badcommand"},
			}}},
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testTimeout", func(*scope.Scope, manifest.ActionParams) (ActionHandler, error) {
					return &asyncTestHandle{data: r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log, r *results) {
				assert.Falsef(t, r.done, "task shouldn't start")
			},
		},
		"skip job if condition expression is bad": {
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				manifest.Job{ActionName: "testBadConditionHook", Condition: "{{bad}} condition"},
			}}},
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testBadConditionHook", func(*scope.Scope, manifest.ActionParams) (ActionHandler, error) {
					return &asyncTestHandle{data: r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log, r *results) {
				assert.Falsef(t, r.done, "task shouldn't start")
			},
		},
		"run job if expression returns OK result": {
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				manifest.Job{ActionName: "testOKConditionHook", Condition: "echo {{msg}}", Vars: manifest.Vars{"msg": "hello"}},
			}}},
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testOKConditionHook", func(*scope.Scope, manifest.ActionParams) (ActionHandler, error) {
					return &asyncTestHandle{data: r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log, r *results) {
				assert.Truef(t, r.done, "task was not started")
			},
		},
		"respect timeout": {
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				manifest.Job{ActionName: "testTimeout", Delay: manifest.Period(800)},
			}}},
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testTimeout", func(*scope.Scope, manifest.ActionParams) (ActionHandler, error) {
					return &asyncTestHandle{data: r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log, r *results) {
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
			taskName: "foo",
			m: manifest.Manifest{Tasks: manifest.TaskSet{"foo": manifest.Task{
				manifest.Job{ActionName: "testDeadline", Deadline: manifest.Period(10)},
			}}},
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testDeadline", func(*scope.Scope, manifest.ActionParams) (ActionHandler, error) {
					return &asyncTestHandle{data: r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log, r *results) {
				assert.Truef(t, r.cancel, "task didn't receive cancel callback on deadline")
				diff := uint((r.cancelTime.Sub(r.startTime).Seconds()) * 1000)
				const deadlineTime = uint(10)
				const permMargin = 2 // On some CI's like AppVeyor, cancel signal can take additional second
				if (diff < deadlineTime-permMargin) || diff > (deadlineTime+permMargin) {
					t.Fatalf("was canceled too late (%d msec), want %d", diff, deadlineTime)
				}
			},
		},
		"execute mixins": {
			taskName: "foo",
			m: manifest.Manifest{
				Mixins: manifest.Mixins{
					"mx1": manifest.Mixin{
						manifest.Job{Description: "start {{foo}}", ActionName: "testMixinExec1"},
					},
				},
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						manifest.Job{ActionName: testAction},
						manifest.Job{MixinName: "mx1", Vars: manifest.Vars{"foo": "bar"}},
					},
				},
			},
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testMixinExec1", func(sc *scope.Scope, ap manifest.ActionParams) (ActionHandler, error) {
					vars := sc.Vars()
					assert.NotEmptyf(t, vars["foo"], "parent job variables wasn't passed to mixin")
					return &asyncTestHandle{data: r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log, r *results) {
				assert.Truef(t, r.done, "task from mixin was not processed")
			},
		},
		"report error if mixin job failed": {
			taskName: "foo",
			err:      "mixin job fail",
			m: manifest.Manifest{
				Mixins: manifest.Mixins{
					"mx1": manifest.Mixin{
						manifest.Job{ActionName: testAction, Params: manifest.ActionParams{"err": "mixin job fail"}},
					},
				},
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						manifest.Job{ActionName: testAction},
						manifest.Job{MixinName: "mx1"},
					},
				},
			},
		},
		"report error if mixin doesn't exists": {
			taskName: "foo",
			err:      `mixin "mx1" doesn't exists`,
			m: manifest.Manifest{
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						manifest.Job{ActionName: testAction},
						manifest.Job{MixinName: "mx1"},
					},
				},
			},
		},
		"not fail if expression in mixin job desc is malformed": {
			taskName: "foo",
			m: manifest.Manifest{
				Mixins: manifest.Mixins{
					"mx1": manifest.Mixin{
						manifest.Job{Description: "bad {{expr}}", ActionName: testAction},
					},
				},
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						manifest.Job{MixinName: "mx1"},
					},
				},
			},
		},
		"check if subtask exists": {
			taskName: "t1",
			err:      `task "t2" doesn't exists`,
			m: manifest.Manifest{
				Tasks: manifest.TaskSet{
					"t1": manifest.Task{
						manifest.Job{TaskName: "t2"},
					},
				},
			},
		},
		"execute subtask": {
			taskName: "foo",
			m: manifest.Manifest{
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						manifest.Job{ActionName: testAction},
						manifest.Job{TaskName: "bar", Vars: manifest.Vars{"foo": "bar"}},
					},
					"bar": manifest.Task{
						manifest.Job{ActionName: testAction, Async: true},
						manifest.Job{Description: "start {{foo}}", ActionName: "testSubTaskExec1"},
					},
				},
			},
			before: func(t *testing.T, _ *TaskRunner, hs *HandlerSet, r *results) {
				_ = hs.HandleFunc("testSubTaskExec1", func(sc *scope.Scope, ap manifest.ActionParams) (ActionHandler, error) {
					vars := sc.Vars()
					assert.NotEmptyf(t, vars["foo"], "parent job variables wasn't passed to mixin")
					return &asyncTestHandle{data: r}, nil
				})
			},
			after: func(t *testing.T, _ *TaskRunner, _ *test.Log, r *results) {
				assert.Truef(t, r.done, "subtask was not processed")
			},
		},
		"return subtask errors": {
			taskName: "foo",
			err:      `task "foo" returned an error on step 2: fail (sub-task step 1)`,
			m: manifest.Manifest{
				Tasks: manifest.TaskSet{
					"foo": manifest.Task{
						manifest.Job{ActionName: testAction},
						manifest.Job{TaskName: "bar"},
					},
					"bar": manifest.Task{
						manifest.Job{ActionName: testAction, Params: manifest.ActionParams{"err": "fail"}},
					},
				},
			},
		},
	}

	for name, c := range cases {
		if c.skip {
			continue
		}

		tc := c
		t.Run("should "+name, func(t *testing.T) {
			r := &results{}
			l := &test.Log{T: t}
			handlers := NewHandlerSet(ActionHandlers{
				testAction: newTestAction,
			})

			tr := NewTaskRunner(Config{
				Logger:   l,
				Handlers: handlers,
				Manifest: &tc.m,
			})
			r.startTime = time.Now()
			if c.before != nil {
				c.before(t, tr, handlers, r)
			}
			err := tr.Run(tc.taskName, nil)
			if tc.err != "" {
				test.AssertErrorContains(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
			if c.after != nil {
				c.after(t, tr, l, r)
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

func (t *asyncTestHandle) Call(*job.RunContext, *TaskRunner) error {
	time.Sleep(time.Millisecond * 100)
	t.data.done = true
	t.data.endTime = time.Now()
	return nil
}

func (t *asyncTestHandle) Cancel(*job.RunContext) error {
	t.data.cancel = true
	t.data.cancelTime = time.Now()
	return nil
}

func newTestAction(_ *scope.Scope, p manifest.ActionParams) (ActionHandler, error) {
	ac := &testActionHandler{}
	err := p.Unmarshal(ac)
	return ac, err
}

type testActionHandler struct {
	Err     string          `mapstructure:"err"`
	Timeout manifest.Period `mapstructure:"timeout"`
}

func (t *testActionHandler) Call(_ *job.RunContext, _ *TaskRunner) error {
	if t.Timeout == 0 {
		t.Timeout = 100
	}

	time.Sleep(t.Timeout.ToDuration())
	if t.Err == "" {
		return nil
	}

	return errors.New(t.Err)
}

func (t *testActionHandler) Cancel(_ *job.RunContext) error {
	return nil
}
