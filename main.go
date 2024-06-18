package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	adminapi "github.com/ukane-philemon/labtracka-api/cmd/admin"
	patientapi "github.com/ukane-philemon/labtracka-api/cmd/patient"
	admindb "github.com/ukane-philemon/labtracka-api/db/admin"
	patientdb "github.com/ukane-philemon/labtracka-api/db/patient"
	"github.com/ukane-philemon/labtracka-api/internal/files"
)

var wg sync.WaitGroup

func main() {
	const dbURLKey = "DB_URL"
	connectionURL := os.Getenv(dbURLKey)
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

	cloud, err := files.NewCloudinaryClient()
	if err != nil {
		logErrorAndExit(err)
	}

	wg.Add(1)
	go func() {
		logger.Info("Attempting to start patient server...")
		startPatientServer(patientDB, adminDB, &patientapi.Config{
			DevMode:      devMode,
			ServerHost:   "0.0.0.0",
			ServerPort:   "8080",
			ServerEmail:  "ukanephilemon@gmail.com",
			SMTPHost:     "smtp-relay.brevo.com",
			SMTPPort:     587,
			SMTPUsername: "labtracka@gmail.com",
			SMTPPassword: "", // Add when needed
			SMTPFrom:     "labtracka@gmail.com",
			Uploader:     cloud,
		})
	}()

	if adminServer {
		wg.Add(1)
		go func() {
			logger.Info("Attempting to start admin server...")
			startAdminServer(patientDB, adminDB, &adminapi.Config{
				DevMode:      devMode,
				ServerHost:   "0.0.0.0",
				ServerPort:   "8081",
				ServerEmail:  "ukanephilemon@gmail.com",
				SMTPHost:     "smtp-relay.brevo.com",
				SMTPPort:     587,
				SMTPUsername: "labtracka@gmail.com",
				SMTPPassword: "", // Add when needed
				SMTPFrom:     "labtracka@gmail.com",
			})
		}()
	}

	wg.Wait()
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
