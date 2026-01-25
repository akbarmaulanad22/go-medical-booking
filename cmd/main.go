package main

import (
	"go-template-clean-architecture/cmd/bootstrap"

	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize application with all dependencies
	app, err := bootstrap.New()
	if err != nil {
		logrus.Fatalf("Failed to initialize application: %v", err)
	}

	// Run the application
	app.Run()
}
