// codesign signs a dxvm binary using the macOS codesign utility.
package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
)

//go:embed dxvm.entitlements
var entitlements []byte

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s path\n\n%[1]s signs a dxvm binary using the macOS codesign utility.\n", os.Args[0])
		os.Exit(2)
	}

	exe, err := exec.LookPath("/usr/bin/codesign")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: /usr/bin/codesign not found (did you run xcode-select --install)")
		os.Exit(1)
	}

	entitlements, err := entitlementsPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: create entitlements file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(entitlements)

	const adhocIdentity = "-"
	cmd := exec.CommandContext(ctx, exe, "--force", "--entitlements", entitlements, "--sign", adhocIdentity, "bin/dxvm")
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: run %s: %v\n", exe, err)
		os.Exit(1)
	}
}

func entitlementsPath() (string, error) {
	f, err := os.CreateTemp("", "dxvm-codesign-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: /usr/bin/codesign not found (did you run xcode-select --install)")
		os.Exit(1)
	}
	if _, err := f.Write(entitlements); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}
