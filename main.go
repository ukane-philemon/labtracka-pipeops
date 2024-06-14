package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	adminapi "github.com/ukane-philemon/labtracka-api/cmd/admin"
	patientapi "github.com/ukane-philemon/labtracka-api/cmd/patient"
	admindb "github.com/ukane-philemon/labtracka-api/db/admin"
	patientdb "github.com/ukane-philemon/labtracka-api/db/patient"
)

func main() {
	var connectionURL string
	flag.StringVar(&connectionURL, "dbConnectionURL", "", "MongoDB Database connection URL")
	var devMode bool
	flag.BoolVar(&devMode, "dev", true, "Specify development environment")
	var adminServer bool
	flag.BoolVar(&adminServer, "admin", false, "Specify server")
	flag.Parse()

	logger := slog.New(slog.Default().Handler())
	logErrorAndExit := func(err error) {
		trace := string(debug.Stack())
		fmt.Println(err)
		fmt.Println(trace)
		os.Exit(1)
	}

	patientDB, err := patientdb.New(context.Background(), devMode, logger.WithGroup("PDB"), connectionURL)
	if err != nil {
		logErrorAndExit(err)
	}

	adminDB, err := admindb.New(context.Background(), devMode, logger.WithGroup("ADB"), connectionURL)
	if err != nil {
		logErrorAndExit(err)
	}

	if !adminServer {
		logger.Info("Attempting to start patient server...")
		startPatientServer(patientDB, adminDB, &patientapi.Config{
			DevMode:      devMode,
			ServerHost:   "localhost",
			ServerPort:   "8080",
			ServerEmail:  "ukanephilemon@gmail.com",
			SMTPHost:     "smtp-relay.brevo.com",
			SMTPPort:     587,
			SMTPUsername: "labtracka@gmail.com",
			SMTPPassword: "", // Add when needed
			SMTPFrom:     "labtracka@gmail.com",
		})
	} else {
		logger.Info("Attempting to start admin server...")
		startAdminServer(patientDB, adminDB, &adminapi.Config{
			DevMode:      devMode,
			ServerHost:   "localhost",
			ServerPort:   "8081",
			ServerEmail:  "ukanephilemon@gmail.com",
			SMTPHost:     "smtp-relay.brevo.com",
			SMTPPort:     587,
			SMTPUsername: "labtracka@gmail.com",
			SMTPPassword: "", // Add when needed
			SMTPFrom:     "labtracka@gmail.com",
		})

	}
}

func startAdminServer(patientDB *patientdb.MongoDB, adminDB *admindb.MongoDB, cfg *adminapi.Config) {
	adminServer, err := adminapi.NewServer(adminDB, patientDB, cfg)
	if err != nil {
		trace := string(debug.Stack())
		fmt.Println(err)
		fmt.Println(trace)
		os.Exit(1)
	}

	adminServer.Run()
}

func startPatientServer(patientDB *patientdb.MongoDB, adminDB *admindb.MongoDB, cfg *patientapi.Config) {
	patientServer, err := patientapi.NewServer(patientDB, adminDB, cfg)
	if err != nil {
		trace := string(debug.Stack())
		fmt.Println(err)
		fmt.Println(trace)
		os.Exit(1)
	}

	patientServer.Run()
}
