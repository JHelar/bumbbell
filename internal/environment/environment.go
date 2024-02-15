package environment

import (
	"log"
	"os"
)

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

var environment Environment
var serverPort string

func GetEnvironment() Environment {
	if environment != "" {
		return environment
	}

	env := os.Getenv("ENVIRONMENT")

	switch env {
	case "development":
		environment = Development
	case "production":
		environment = Production
	default:
		log.Println("Environment not set, defaulting to development")
		environment = Development
	}

	return environment
}

func GetServerPort() string {
	if serverPort != "" {
		return serverPort
	}

	port := os.Getenv("PORT")

	if port == "" {
		log.Println("Port not set, defaulting to 8080")
		port = "8080"
	}

	serverPort = ":" + port

	return serverPort
}
