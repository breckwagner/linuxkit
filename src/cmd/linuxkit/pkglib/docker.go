package pkglib

// Thin wrappers around Docker CLI invocations

//go:generate ./gen

import (
	"fmt"
	"os"
	"os/exec"
)

const debugDockerCommands = false

const dctEnableEnv = "DOCKER_CONTENT_TRUST=1"

type dockerRunner struct {
	dct   bool
	cache bool
}

func newDockerRunner(dct, cache bool) dockerRunner {
	return dockerRunner{dct: dct, cache: cache}
}

func isExecErrNotFound(err error) bool {
	eerr, ok := err.(*exec.Error)
	if !ok {
		return false
	}
	return eerr.Err == exec.ErrNotFound
}

func (dr dockerRunner) command(args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if dr.dct {
		cmd.Env = append(cmd.Env, dctEnableEnv)
	}
	if debugDockerCommands {
		var dct string
		if dr.dct {
			dct = dctEnableEnv + " "
		}
		fmt.Fprintf(os.Stderr, "+ %s%v\n", dct, cmd.Args)
	}
	err := cmd.Run()
	if isExecErrNotFound(err) {
		return fmt.Errorf("linuxkit pkg requires docker to be installed")
	}
	return err
}

func (dr dockerRunner) pull(img string) (bool, error) {
	err := dr.command("pull", img)
	if err == nil {
		return true, nil
	}
	switch err.(type) {
	case *exec.ExitError:
		return false, nil
	default:
		return false, err
	}
}

func (dr dockerRunner) push(img string) error {
	return dr.command("push", img)
}

func (dr dockerRunner) pushWithManifest(img, suffix string) error {
	if err := dr.push(img + suffix); err != nil {
		return err
	}

	var dctArg string
	if dr.dct {
		dctArg = "1"
	}

	cmd := exec.Command("/bin/sh", "-c", manifestPushScript, "manifest-push-script", img, dctArg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if debugDockerCommands {
		fmt.Fprintf(os.Stderr, "+ %v\n", cmd.Args)
	}
	return cmd.Run()
}

func (dr dockerRunner) tag(ref, tag string) error {
	return dr.command("tag", ref, tag)
}

func (dr dockerRunner) build(tag, pkg string, opts ...string) error {
	args := []string{"build"}
	if !dr.cache {
		args = append(args, "--no-cache")
	}
	args = append(args, opts...)
	args = append(args, "-t", tag, pkg)
	return dr.command(args...)
}
