package config

import "time"

// version should be initialized by CI/CD workflow with release branch version using -ldflags,
// if you are checking this file from within its repo, you have one in ".github" directory :P
var version = "unknown: check CI/CD workflow"

const (
	GracefulTimeout = time.Second * 30
)
