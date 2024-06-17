package patient

import (
	"errors"
	"fmt"

	"github.com/ukane-philemon/labtracka-api/internal/files"
	"github.com/ukane-philemon/labtracka-api/internal/validator"
)

type Config struct {
	DevMode bool

	// Server config
	ServerHost string
	ServerPort string
	// Provide email to receive server error.
	ServerEmail string // optional.

	// SMTP config
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	Uploader files.FileUploader
}

// Validate checks that the config values are valid.
func (cfg Config) Validate() error {
	if cfg.ServerHost == "" {
		return errors.New("server host is required")
	}

	if cfg.ServerPort == "" {
		return errors.New("server port is required")
	}

	if cfg.ServerEmail != "" && !validator.IsEmail(cfg.ServerEmail) {
		return fmt.Errorf("%s is not a valid email", cfg.ServerEmail)
	}

	return nil
}

// hasValidSMTPConfig checks if valid SMTP config was provided.
func (cfg Config) hasValidSMTPConfig() bool {
	return cfg.SMTPHost != "" && cfg.SMTPPort > 0 && cfg.SMTPUsername != "" && cfg.SMTPPassword != "" && cfg.SMTPFrom != "" && validator.IsEmail(cfg.SMTPFrom)
}
