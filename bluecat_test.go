package bluecat

import (
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestProvisionWithEnvVars(t *testing.T) {
	p := Provider{}
	p.ServerURL = "{env.SERVER_URL}"
	p.Username = "{env.USERNAME}"
	p.Password = "{env.PASSWORD}"

	// Note: In a real test, you'd set up the Caddy context properly
	// This is just a basic compilation test
}

func TestUnmarshalCaddyfile(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		shouldErr bool
	}{
		{
			name: "valid config",
			config: `bluecat {
				server_url https://bluecat.example.com
				username admin
				password secret
			}`,
			shouldErr: false,
		},
		{
			name: "missing server_url",
			config: `bluecat {
				username admin
				password secret
			}`,
			shouldErr: true,
		},
		{
			name: "missing username",
			config: `bluecat {
				server_url https://bluecat.example.com
				password secret
			}`,
			shouldErr: true,
		},
		{
			name: "missing password",
			config: `bluecat {
				server_url https://bluecat.example.com
				username admin
			}`,
			shouldErr: true,
		},
		{
			name: "with optional config",
			config: `bluecat {
				server_url https://bluecat.example.com
				username admin
				password secret
				configuration_name MyConfig
				view_name MyView
				deployment_batch_window 5s
			}`,
			shouldErr: false,
		},
		{
			name: "invalid directive",
			config: `bluecat {
				server_url https://bluecat.example.com
				username admin
				password secret
				invalid_field value
			}`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispenser := caddyfile.NewTestDispenser(tt.config)
			p := Provider{}

			err := p.UnmarshalCaddyfile(dispenser)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestUnmarshalCaddyfileValues(t *testing.T) {
	config := `bluecat {
		server_url https://bluecat.example.com
		username testuser
		password testpass
		configuration_name TestConfig
		view_name TestView
		deployment_batch_window 5s
	}`

	dispenser := caddyfile.NewTestDispenser(config)
	p := Provider{}

	err := p.UnmarshalCaddyfile(dispenser)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if p.ServerURL != "https://bluecat.example.com" {
		t.Errorf("Expected ServerURL to be 'https://bluecat.example.com', got '%s'", p.ServerURL)
	}
	if p.Username != "testuser" {
		t.Errorf("Expected Username to be 'testuser', got '%s'", p.Username)
	}
	if p.Password != "testpass" {
		t.Errorf("Expected Password to be 'testpass', got '%s'", p.Password)
	}
	if p.ConfigurationName != "TestConfig" {
		t.Errorf("Expected ConfigurationName to be 'TestConfig', got '%s'", p.ConfigurationName)
	}
	if p.ViewName != "TestView" {
		t.Errorf("Expected ViewName to be 'TestView', got '%s'", p.ViewName)
	}
	if p.DeploymentBatchWindow != "5s" {
		t.Errorf("Expected DeploymentBatchWindow to be '5s', got '%s'", p.DeploymentBatchWindow)
	}
}

func TestParseDeploymentBatchWindow(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "empty uses default", input: "", want: 0},
		{name: "valid duration", input: "5s", want: 5 * time.Second},
		{name: "invalid duration", input: "later", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDeploymentBatchWindow(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
