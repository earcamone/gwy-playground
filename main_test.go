package main

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMainAsExternalProcess is an integration-style test that
// mimics clients downloading gapi and using it for the first
// time, apart from ensuring it runs, handles a requests and
// exits gracefully when a SIGTERM is received.
func TestMainAsExternalProcess(t *testing.T) {
	// Build the binary to run it as a separate process
	cmd := exec.Command("go", "build", "-o", "testapi", "main.go")
	err := cmd.Run()

	assert.NoError(t, err, "Failed to build main.go")
	defer os.Remove("testapi") // Clean up binary after test

	// Run gapi binary
	cmd = exec.Command("./testapi")

	// Capture stdout and stderr
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err = cmd.Start()
	assert.NoError(t, err, "Failed to start the binary")

	// Ensure process is killed if test fails
	defer cmd.Process.Kill()

	// Wait 2 seconds for the server to start
	time.Sleep(2 * time.Second)

	// 1. Check if "Running API" is in the output
	output := outBuf.String()
	assert.Contains(t, output, "Running API", "Expected 'Running API' in output")

	// 2. Send GET request to "/" and expect 404
	resp, err := http.Get("http://localhost:8080/")
	assert.NoError(t, err, "GET request should not fail")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected 404 Not Found")
	resp.Body.Close()

	// 3. Send SIGTERM to the process
	err = cmd.Process.Signal(os.Signal(syscall.SIGTERM))
	assert.NoError(t, err, "Failed to send SIGTERM")

	// Wait 2 seconds for graceful shutdown
	time.Sleep(2 * time.Second)

	// 4. Check for shutdown message
	output = outBuf.String()
	assert.Contains(t, output, "shutting down HTTP server gracefully",
		"Expected shutdown message in output")

	// Ensure the process exits
	err = cmd.Wait()
}
