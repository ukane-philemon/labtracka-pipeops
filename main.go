package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	patientapi "github.com/ukane-philemon/labtracka-api/cmd/patient"
	patientdb "github.com/ukane-philemon/labtracka-api/db/patient"
)

func main() {
	connectionURL := os.Args[0]
	logger := slog.New(slog.Default().Handler())

	userDB, err := patientdb.New(context.Background(), logger.WithGroup("CDB"), connectionURL)
	if err != nil {
		trace := string(debug.Stack())
		fmt.Println(err)
		fmt.Println(trace)
		os.Exit(1)
	}

	cfg := &patientapi.Config{
		DevMode:      false,
		ServerHost:   "localhost",
		ServerPort:   "8080",
		ServerEmail:  "ukanephilemon@gmail.com",
		SMTPHost:     "smtp-relay.brevo.com",
		SMTPPort:     587,
		SMTPUsername: "labtracka@gmail.com",
		SMTPPassword: "vAn03c2wjxb47PI8",
		SMTPFrom:     "labtracka@gmail.com",
	}

	customerServer, err := patientapi.NewServer(userDB, userDB, cfg)
	if err != nil {
		trace := string(debug.Stack())
		fmt.Println(err)
		fmt.Println(trace)
		os.Exit(1)
	}

	customerServer.Run()
}
