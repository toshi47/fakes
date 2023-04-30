package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"api/auth_manager"
	"api/rest"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Warnf("failed to load environment: %v", err)
	}

	logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Fatalf("failed to set log level: %v", err)
	}
	switch log.Level(logLevel) {
	case log.PanicLevel, log.FatalLevel, log.ErrorLevel, log.WarnLevel, log.InfoLevel, log.DebugLevel, log.TraceLevel:
		break
	default:
		log.Fatalf("unknown log level: %v", logLevel)
	}

	log.SetLevel(log.Level(logLevel))
	log.SetReportCaller(true)

	authmgr, err := auth_manager.New(auth_manager.EmailInfo{
		Address:  os.Getenv("EMAIL_ADDRESS"),
		Password: os.Getenv("EMAIL_PASSWORD"),
		SmtpHost: os.Getenv("SMTP_HOST"),
		SmtpPort: os.Getenv("SMTP_PORT"),
	})
	if err != nil {
		log.Fatalf("failed to create auth mgr: %v", err)
	}

	server, err := rest.NewServer(os.Getenv("API_ADDRESS"), os.Getenv("NETWORK_ADDRESS"), []byte(os.Getenv("COOKIE_STORE_HASH_KEY")), authmgr)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	server.Start()
}
