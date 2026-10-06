package bluecat

import (
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
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
				deploy_delay 5s
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
		deploy_delay 5s
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
	if p.DeployDelay != caddy.Duration(5e9) {
		t.Errorf("Expected DeployDelay to be 5s, got %v", p.DeployDelay)
	}
}

func TestDeployDelayParsing(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		shouldErr bool
	}{
		{name: "valid duration", config: `bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		deploy_delay 30s
	}`, shouldErr: false},
		{name: "disable deploy", config: `bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		deploy_delay -1
	}`, shouldErr: false},
		{name: "invalid duration", config: `bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		deploy_delay later
	}`, shouldErr: true},
		{name: "max deploy delay", config: `bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		max_deploy_delay 1m
	}`, shouldErr: false},
		{name: "max deploy delay missing value", config: `bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		max_deploy_delay
	}`, shouldErr: true},
		{name: "invalid max deploy delay", config: `bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		max_deploy_delay soon
	}`, shouldErr: true},
		{name: "disable deploy flag", config: `bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		disable_deploy
	}`, shouldErr: false},
		{name: "disable deploy takes no argument", config: `bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		disable_deploy yes
	}`, shouldErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispenser := caddyfile.NewTestDispenser(tt.config)
			p := Provider{}
			err := p.UnmarshalCaddyfile(dispenser)
			if tt.shouldErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestNewBluecatProvider checks the module config reaches libdns/bluecat with
// the right deploy settings. In particular, the legacy "deploy_delay -1" must
// disable deploys: libdns/bluecat itself treats a negative delay as "use the
// default", which would silently turn deploys back on.
func TestNewBluecatProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		wantDelay   time.Duration
		wantMax     time.Duration
		wantDisable bool
	}{
		{name: "defaults", config: ``},
		{name: "explicit delays", config: `
		deploy_delay 10s
		max_deploy_delay 1m`, wantDelay: 10 * time.Second, wantMax: time.Minute},
		{name: "disable_deploy", config: `
		disable_deploy`, wantDisable: true},
		{name: "legacy -1 disables deploys", config: `
		deploy_delay -1`, wantDisable: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispenser := caddyfile.NewTestDispenser(`bluecat {
		server_url https://bluecat.example.com
		username u
		password p
		configuration_name C
		view_name V` + tt.config + `
	}`)
			p := Provider{}
			if err := p.UnmarshalCaddyfile(dispenser); err != nil {
				t.Fatalf("UnmarshalCaddyfile: %v", err)
			}

			bp := newBluecatProvider(&p)
			if bp.DeployDelay != tt.wantDelay {
				t.Errorf("DeployDelay = %v, want %v", bp.DeployDelay, tt.wantDelay)
			}
			if bp.MaxDeployDelay != tt.wantMax {
				t.Errorf("MaxDeployDelay = %v, want %v", bp.MaxDeployDelay, tt.wantMax)
			}
			if bp.DisableDeploy != tt.wantDisable {
				t.Errorf("DisableDeploy = %v, want %v", bp.DisableDeploy, tt.wantDisable)
			}
			if bp.ServerURL != "https://bluecat.example.com" || bp.Username != "u" || bp.Password != "p" ||
				bp.ConfigurationName != "C" || bp.ViewName != "V" {
				t.Errorf("connection settings not passed through: %+v", bp)
			}
		})
	}
}
