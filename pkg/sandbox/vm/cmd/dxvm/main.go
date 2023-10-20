package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"time"

	"go.jetpack.io/pkg/sandbox/vm"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dataDir := "./dxvm"
	devboxDir, devboxDirFound, err := findDevboxDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "no devbox.json found, using %s for state: %v\n", dataDir, err)
	} else if !devboxDirFound {
		fmt.Fprintf(os.Stderr, "no devbox.json found, using %s for state: searched up to %s\n", dataDir, devboxDir)
	} else {
		dataDir = filepath.Join(devboxDir, ".devbox", "vm")
	}

	dxvm := vm.VM{}
	flag.StringVar(&dxvm.HostDataDir, "datadir", dataDir, "`path` to the directory for saving VM state")
	flag.BoolVar(&dxvm.Install, "install", false, "mount NixOS install image")
	flag.Parse()

	if dxvm.Install {
		slog.Debug("downloading the NixOS installer, this make take a few minutes")
	} else if devboxDirFound {
		dxvm.SharedDirectories = append(dxvm.SharedDirectories, vm.SharedDirectory{
			Path:     devboxDir,
			HomeDir:  true,
			ReadOnly: false,
		})
		fmt.Fprintln(os.Stderr, "booting virtual machine")
	}

	go func() {
		<-ctx.Done()
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := dxvm.Stop(ctx); err != nil {
			slog.Error("stop virtual machine", "err", err)
		}
	}()
	if err := dxvm.Run(ctx); err != nil {
		slog.Error("run virtual machine install", "err", err)
		os.Exit(1)
	}

	// Restart if we just finished bootrapping a new VM.
	if dxvm.Install {
		fmt.Fprintln(os.Stderr, "virtual machine created successfully")
		dxvm.Install = false
		if err := dxvm.Run(ctx); err != nil {
			slog.Error("run virtual machine", "err", err)
			os.Exit(1)
		}
	}
}

func findDevboxDir() (dir string, found bool, err error) {
	dir = "."
	if wd, err := os.Getwd(); err == nil {
		dir = wd
	}

	home, _ := os.UserHomeDir()
	vol := filepath.VolumeName(dir)
	for {
		// Refuse to go past the user's home directory or search root.
		if dir == "" || dir == "/" || dir == home || dir == vol {
			return dir, false, nil
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return "", false, nil
		}
		_, found := slices.BinarySearchFunc(entries, "devbox.json", func(e fs.DirEntry, t string) int {
			return cmp.Compare(e.Name(), t)
		})
		if found {
			return dir, true, nil
		}
		dir = filepath.Dir(dir)
	}
}
