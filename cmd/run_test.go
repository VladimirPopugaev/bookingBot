package cmd

import (
	"booking_bot/internal/domain"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *domain.Config
		port    int
		wantErr bool
	}{
		{
			name: "tls disabled allows empty cert files",
			cfg: &domain.Config{
				HTTP: domain.HTTPConfig{
					Host:       "0.0.0.0",
					TLSEnabled: false,
				},
			},
			port: 8080,
		},
		{
			name: "tls enabled requires cert file",
			cfg: &domain.Config{
				HTTP: domain.HTTPConfig{
					Host:       "0.0.0.0",
					TLSEnabled: true,
					KeyFile:    "server.key",
				},
			},
			port:    8080,
			wantErr: true,
		},
		{
			name: "tls enabled requires key file",
			cfg: &domain.Config{
				HTTP: domain.HTTPConfig{
					Host:       "0.0.0.0",
					TLSEnabled: true,
					CertFile:   "server.crt",
				},
			},
			port:    8080,
			wantErr: true,
		},
		{
			name: "invalid port",
			cfg: &domain.Config{
				HTTP: domain.HTTPConfig{
					Host: "0.0.0.0",
				},
			},
			port:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateConfig(tt.cfg, tt.port)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
