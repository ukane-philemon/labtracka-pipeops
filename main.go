package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	patientapi "github.com/ukane-philemon/labtracka-api/cmd/patient"
	admindb "github.com/ukane-philemon/labtracka-api/db/admin"
	patientdb "github.com/ukane-philemon/labtracka-api/db/patient"
)

func main() {
	var connectionURL string
	flag.StringVar(&connectionURL, "dbURL", "", "Database connection URL")
	flag.Parse()

	logger := slog.New(slog.Default().Handler())
	logErrorAndExit := func(err error) {
		trace := string(debug.Stack())
		fmt.Println(err)
		fmt.Println(trace)
		os.Exit(1)
	}

	userDB, err := patientdb.New(context.Background(), logger.WithGroup("CDB"), connectionURL)
	if err != nil {
		logErrorAndExit(err)
	}

	adminDB, err := admindb.New(context.Background(), logger.WithGroup("ADB"), connectionURL)
	if err != nil {
		logErrorAndExit(err)
	}

	cfg := &patientapi.Config{
		DevMode:      false,
		ServerHost:   "localhost",
		ServerPort:   "8080",
		ServerEmail:  "ukanephilemon@gmail.com",
		SMTPHost:     "smtp-relay.brevo.com",
		SMTPPort:     587,
		SMTPUsername: "labtracka@gmail.com",
		SMTPPassword: "", // Add when needed
		SMTPFrom:     "labtracka@gmail.com",
	}

	customerServer, err := patientapi.NewServer(userDB, adminDB, cfg)
	if err != nil {
		trace := string(debug.Stack())
		fmt.Println(err)
		fmt.Println(trace)
		os.Exit(1)
	}

	customerServer.Run()
}
