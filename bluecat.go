package bluecat

import (
	"context"
	"net/netip"

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
	// ConfigurationName is the name of the configuration to use
	ConfigurationName string `json:"configuration_name,omitempty"`
	// ViewName is the name of the view to use
	ViewName string `json:"view_name,omitempty"`

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
	p.provider = &bluecat.Provider{
		ServerURL:         p.ServerURL,
		Username:          p.Username,
		Password:          p.Password,
		ConfigurationName: p.ConfigurationName,
		ViewName:          p.ViewName,
	}

	logger.Info("Bluecat DNS provider provisioned")

	return nil
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
//	bluecat {
//	    server_url <url>
//	    username <username>
//	    password <password>
//	    configuration_name <name>  // optional
//	    view_name <name>           // optional
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
	// Convert generic libdns.Record to concrete types for proper handling by libdns/bluecat
	converted := make([]libdns.Record, len(records))
	for i, rec := range records {
		converted[i] = convertToConcreteType(rec)
	}
	return p.provider.AppendRecords(ctx, zone, converted)
}

// convertToConcreteType converts a generic libdns.Record to its concrete type
// based on the Type field. This is necessary because certmagic creates generic
// Record structs, but libdns/bluecat needs concrete types for proper type switching.
func convertToConcreteType(rec libdns.Record) libdns.Record {
	// If it's already a concrete type, return as-is to preserve ProviderData
	switch r := rec.(type) {
	case libdns.TXT, libdns.Address, libdns.CNAME, libdns.MX, libdns.NS, libdns.SRV:
		return r
	}

	// Otherwise, convert based on the Type field
	// Note: libdns.RR doesn't have ProviderData field, so we can't preserve it
	// This is a limitation of how certmagic stores/returns records
	rr := rec.RR()
	
	switch rr.Type {
	case "TXT":
		return libdns.TXT{
			Name: rr.Name,
			TTL:  rr.TTL,
			Text: rr.Data,
		}
	case "A", "AAAA":
		// Parse IP address from Data field
		if ip, err := netip.ParseAddr(rr.Data); err == nil {
			return libdns.Address{
				Name: rr.Name,
				TTL:  rr.TTL,
				IP:   ip,
			}
		}
	case "CNAME":
		return libdns.CNAME{
			Name:   rr.Name,
			TTL:    rr.TTL,
			Target: rr.Data,
		}
	}

	// Return original if we can't convert
	return rec
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
