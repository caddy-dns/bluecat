package bluecat

import (
	"context"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/bluecat"
	"github.com/libdns/libdns"
)

// Provider lets Caddy read and manipulate DNS records hosted by Bluecat Address Manager.
type Provider struct {
	// ServerURL is the URL of the Bluecat Address Manager server
	ServerURL string `json:"server_url,omitempty"`
	// Username is the API username
	Username string `json:"username,omitempty"`
	// Password is the API password
	Password string `json:"password,omitempty"`
	// ConfigurationName is accepted for compatibility but not currently
	// applied to zone lookups by libdns/bluecat. A zone name that matches
	// more than one zone fails with an ambiguity error instead of guessing.
	ConfigurationName string `json:"configuration_name,omitempty"`
	// ViewName limits zone lookups to one DNS view (optional)
	ViewName string `json:"view_name,omitempty"`
	// DeployDelay controls how long to wait after the last DNS record write
	// before issuing a QuickDeploy to Bluecat. Writes are debounced per zone,
	// so concurrent ACME DNS-01 challenges collapse into a single deploy.
	//
	// Accepts a Go duration string, e.g. "5s", "30s". Defaults to 5 seconds
	// when unset. A negative value ("-1" in the Caddyfile) is a legacy alias
	// for DisableDeploy.
	DeployDelay caddy.Duration `json:"deploy_delay,omitempty"`
	// MaxDeployDelay caps how long debouncing can postpone a deploy, so a
	// steady stream of writes during bulk issuance can't delay it forever.
	// Defaults to four times DeployDelay or 30 seconds, whichever is larger.
	MaxDeployDelay caddy.Duration `json:"max_deploy_delay,omitempty"`
	// DisableDeploy suppresses automatic deployment entirely. Records are
	// written to Bluecat but not pushed to the DNS servers.
	DisableDeploy bool `json:"disable_deploy,omitempty"`

	provider *bluecat.Provider
}

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "dns.providers.bluecat",
		New: func() caddy.Module {
			return &Provider{}
		},
	}
}

// Provision sets up the module. Implements caddy.Provisioner.
func (p *Provider) Provision(ctx caddy.Context) error {
	logger := ctx.Logger(p)

	// Apply replacements to the configuration fields
	repl := caddy.NewReplacer()
	p.ServerURL = repl.ReplaceAll(p.ServerURL, "")
	p.Username = repl.ReplaceAll(p.Username, "")
	p.Password = repl.ReplaceAll(p.Password, "")
	p.ConfigurationName = repl.ReplaceAll(p.ConfigurationName, "")
	p.ViewName = repl.ReplaceAll(p.ViewName, "")

	// Initialize the embedded provider with the configuration
	p.provider = newBluecatProvider(p)
	p.provider.Logger = ctx.Slogger()

	logger.Info("Bluecat DNS provider provisioned")

	return nil
}

// newBluecatProvider builds the libdns provider from the module config.
// libdns/bluecat treats a negative DeployDelay as "use the default", not
// "disable", so the legacy -1 sentinel is translated to DisableDeploy here.
func newBluecatProvider(p *Provider) *bluecat.Provider {
	deployDelay := time.Duration(p.DeployDelay)
	disableDeploy := p.DisableDeploy
	if deployDelay < 0 {
		deployDelay = 0
		disableDeploy = true
	}

	return &bluecat.Provider{
		ServerURL:         p.ServerURL,
		Username:          p.Username,
		Password:          p.Password,
		ConfigurationName: p.ConfigurationName,
		ViewName:          p.ViewName,
		DeployDelay:       deployDelay,
		MaxDeployDelay:    time.Duration(p.MaxDeployDelay),
		DisableDeploy:     disableDeploy,
	}
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
//	bluecat {
//	    server_url <url>
//	    username <username>
//	    password <password>
//	    configuration_name <name>  // optional
//	    view_name <name>           // optional
//	    deploy_delay <duration>    // optional, e.g. "5s" (default), "30s"
//	    max_deploy_delay <duration> // optional, default max(4*deploy_delay, 30s)
//	    disable_deploy             // optional, never QuickDeploy
//	}
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "server_url":
				if d.NextArg() {
					p.ServerURL = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "username":
				if d.NextArg() {
					p.Username = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "password":
				if d.NextArg() {
					p.Password = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "configuration_name":
				if d.NextArg() {
					p.ConfigurationName = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "view_name":
				if d.NextArg() {
					p.ViewName = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "deploy_delay":
				if d.NextArg() {
					val := d.Val()
					if val == "-1" {
						p.DeployDelay = -1
					} else {
						dur, err := caddy.ParseDuration(val)
						if err != nil {
							return d.Errf("invalid deploy_delay duration %q: %v", val, err)
						}
						p.DeployDelay = caddy.Duration(dur)
					}
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "max_deploy_delay":
				if !d.NextArg() {
					return d.ArgErr()
				}
				dur, err := caddy.ParseDuration(d.Val())
				if err != nil {
					return d.Errf("invalid max_deploy_delay duration %q: %v", d.Val(), err)
				}
				p.MaxDeployDelay = caddy.Duration(dur)
				if d.NextArg() {
					return d.ArgErr()
				}
			case "disable_deploy":
				p.DisableDeploy = true
				if d.NextArg() {
					return d.ArgErr()
				}
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}

	if p.ServerURL == "" {
		return d.Err("missing server URL")
	}
	if p.Username == "" {
		return d.Err("missing username")
	}
	if p.Password == "" {
		return d.Err("missing password")
	}

	return nil
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	return p.provider.GetRecords(ctx, zone)
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.provider.AppendRecords(ctx, zone, records)
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.provider.SetRecords(ctx, zone, records)
}

// DeleteRecords deletes the specified records from the zone.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.provider.DeleteRecords(ctx, zone, records)
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
