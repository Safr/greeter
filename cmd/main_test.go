package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"
)

var binaryName string

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		binaryName = "greeter.exe"
	} else {
		binaryName = "greeter"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryName)
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		err = os.Remove(binaryName)
		if err != nil {
			log.Fatalf("Error removing built binary: %v", err)
		}
	}()
	m.Run()
}

func TestApplication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := path.Join(curDir, binaryName)
	t.Log(binaryPath)

	tests := []struct {
		args                []string
		input               string
		expectedOutputLines []string
		expectedExitCode    int
	}{
		{
			args:             []string{},
			expectedExitCode: 1,
			expectedOutputLines: []string{
				"Must specify a number greater than 0",
			},
		},
		{
			args:                []string{"-h"},
			expectedExitCode:    1,
			expectedOutputLines: []string{},
		},
		{
			args:                []string{"a"},
			expectedExitCode:    1,
			expectedOutputLines: []string{},
		},
		{
			args:             []string{"-n", "2"},
			input:            "jane doe",
			expectedExitCode: 0,
			expectedOutputLines: []string{
				"Your name please? Press the Enter key when done.",
				"Nice to meet you jane doe",
			},
		},
	}

	byteBuf := new(bytes.Buffer)
	for _, tc := range tests {
		t.Logf("Executing:%v %v\n", binaryPath, tc.args)
		cmd := exec.CommandContext(ctx, binaryPath, tc.args...)
		cmd.Stdout = byteBuf
		if len(tc.input) != 0 {
			cmd.Stdin = strings.NewReader(tc.input)
		}
		err := cmd.Run()

		if err != nil && tc.expectedExitCode == 0 {
			t.Fatalf("Expected application to exit without an error. Got: %v", err)
		}

		if cmd.ProcessState.ExitCode() != tc.expectedExitCode {
			t.Log(byteBuf.String())
			t.Fatalf("Expected application to have exit code: %v. Got: %v", tc.expectedExitCode, cmd.ProcessState.ExitCode())

		}

		output := byteBuf.String()
		lines := strings.Split(output, "\n")
		for num := range tc.expectedOutputLines {
			t.Log("lines[num] ", lines[num])
			t.Log("tc.expectedOutputLines[num] ", tc.expectedOutputLines[num])
			if lines[num] != tc.expectedOutputLines[num] {
				t.Fatalf("Expected output line to be:%v, Got:%v", tc.expectedOutputLines[num], lines[num])
			}
		}
		byteBuf.Reset()
	}
}
