package vm

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Code-Hex/vz/v3"
)

func scriptedConsole(ctx context.Context, logger *slog.Logger, prompt string, script []string) (*vz.VirtioConsoleDeviceSerialPortConfiguration, error) {
	stdinr, stdinw, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %v", err)
	}
	stdoutr, stdoutw, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %v", err)
	}

	go func() {
		var idle *time.Timer
		idleDur := time.Second
		sawPrompt := false
		doneWriting := false
		scanner := bufio.NewScanner(io.TeeReader(stdoutr, os.Stdout))
		for scanner.Scan() && ctx.Err() == nil {
			logger.Debug("install console", "stdout", scanner.Text())

			if doneWriting {
				continue
			}
			if idle != nil {
				doneWriting = !idle.Reset(idleDur)
				continue
			}

			sawPrompt = sawPrompt || strings.Contains(scanner.Text(), prompt)
			if !sawPrompt {
				continue
			}
			idle = time.AfterFunc(idleDur, sync.OnceFunc(func() {
				_, err := stdinw.WriteString(strings.Join(script, " && ") + "\n")
				if err != nil {
					logger.Error("error writing to VM standard input", "err", err)
				}
				stdinw.Close()
			}))
		}
		if err := scanner.Err(); err != nil {
			logger.Error("error reading install console stdout", "err", err)
		}
	}()

	attach, err := vz.NewFileHandleSerialPortAttachment(stdinr, stdoutw)
	if err != nil {
		return nil, fmt.Errorf("create serial port attachment: %v", err)
	}
	config, err := vz.NewVirtioConsoleDeviceSerialPortConfiguration(attach)
	if err != nil {
		return nil, fmt.Errorf("create serial port configuration: %v", err)
	}
	return config, nil
}
