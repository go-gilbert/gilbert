package tasks

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/x1unix/gilbert/log"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner"
	"github.com/x1unix/gilbert/scope"
	"os"
	"os/signal"
)

var (
	r *runner.TaskRunner
)

func getManifest(dir string) (*manifest.Manifest, error) {
	return manifest.FromDirectory(dir)
}

// RunTask is a handler for 'run' command
func RunTask(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return fmt.Errorf("no task specified")
	}

	task := c.Args()[0]

	r, err = getRunner()
	if err != nil {
		return err
	}

	return runTask(task)
}

func getRunner() (*runner.TaskRunner, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get current working directory, %v", err)
	}

	m, err := getManifest(dir)
	if err != nil {
		return nil, err
	}

	if err := importProjectPlugins(m, dir); err != nil {
		return nil, err
	}

	return runner.NewTaskRunner(m, dir, log.Default), nil
}

func importProjectPlugins(m *manifest.Manifest, cwd string) error {
	s := scope.CreateScope(cwd, m.Vars)
	for _, uri := range m.Plugins {
		expanded, err := s.ExpandVariables(uri)
		if err != nil {
			return fmt.Errorf("failed to load plugins from manifest, %s", err)
		}

		if err := plugins.Import(expanded); err != nil {
			return err
		}
	}

	return nil
}

func handleShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Default.Log("Shutting down...")
		r.Stop()
	}
}

func runTask(taskName string) error {
	go handleShutdown()
	if err := r.RunTask(taskName); err != nil {
		return err
	}

	log.Default.Successf("Task '%s' ran successfully\n", taskName)
	return nil
}
