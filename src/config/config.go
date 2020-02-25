package config

import (
	"os"
	"strconv"
)

func ternary(value string, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

var (
	// IsMaster master
	IsMaster = os.Getenv("MASTER") == "1"

	// Verbose verbose
	Verbose = os.Getenv("VERBOSE") == "1"

	// HTTPPort - http port to run
	HTTPPort, _  = strconv.Atoi(ternary(os.Getenv("HTTP_PORT"), "8088"))
	DBConnection = ternary(os.Getenv("DB_CONNECTION"), "test:test@tcp(127.0.0.1:3306)/personal_project")
)
