package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/earcamone/gwy-playground/api"
	"github.com/earcamone/gwy-playground/api/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	PASSWORD = "1234"
)

func main() {
	// Just some snippet to trigger secrets alert

	// Slack credentials
	slackToken := "xoxb-123456789012-1234567890123-abcdef1234567890abcdef12"
	slackWebhook := "https://hooks.slack.com/services/T12345678/B12345678/abcdef1234567890abcdef12"

	// AWS credentials
	awsAccessKeyID := "AKIAIOSFODNN7EXAMPLE"
	awsSecretAccessKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

	// Just to use them and avoid "unused variable" lint errors
	fmt.Println("Slack Token:", slackToken)
	fmt.Println("Slack Webhook:", slackWebhook)
	fmt.Println("AWS Access Key:", awsAccessKeyID)
	fmt.Println("AWS Secret Key:", awsSecretAccessKey)

	//
	// API initialization: You can customize config and dependencies with New(),
	// check available config options in api/config package. If no custom options
	// are specified, config package will be able to set proper default values.
	//
	// If you are setting configuration symbols from within environment variables,
	// I recommend you injecting this logic inside this package to maintain the
	// main file as clean as possible.
	//
	// NOTE: it's initialized this way instead of classic ListenAndServe(port)
	// snippet to allow the call of Shutdown() method in following shutdown scheme.
	//

	c := config.New()

	server := &http.Server{
		Addr:    c.Address,
		Handler: api.New(c),
	}

	//
	// Run blocking API in thread so we can listen for SIGKILL (and similar) signals,
	// triggering a server graceful shutdown, ending active connections gracefully
	// instead of abruptly closing (TCP: RST) them up inside WaitForShutdown().
	//
	// PER AWS DOCUMENTATION: When a task is stopped, a SIGTERM signal is sent to each
	// container’s entry process, usually PID 1. After a timeout has lapsed, the process
	// will be sent a SIGKILL signal. By default, there is a 30 seconds delay between the
	// delivery of SIGTERM and SIGKILL signals. These signals are sent to container entry
	// point process
	//

	go func() {
		fmt.Printf("Running API (%s) on %s...\n", c.Version, c.Address)
		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("API internal error un-expected shutdown: %v", err)
		}
	}()

	WaitForShutdown(server, c)
}

func WaitForShutdown(server *http.Server, c config.Config) {
	// Block until SIGINT, SIGTERM or SIGKILL signals are received

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-signalCh

	// Call client supplied shutdown function if client specified one

	if c.ShutdownFn != nil {
		go c.ShutdownFn()
	}

	// Interrupt signal received, give out N seconds to shut down API gracefully

	ctx, cancel := context.WithTimeout(context.Background(), c.ShutdownTimeout)
	defer cancel()

	log.Printf("Received %s, shutting down HTTP server gracefully (TO: %v)\n", sig, c.ShutdownTimeout)
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %s", err)
	}
}
