package bluecat

import (
	"context"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/bluecat"
	"github.com/libdns/libdns"
	"go.uber.org/zap"
)

// Provider lets Caddy read and manipulate DNS records hosted by Bluecat Address Manager.
type Provider struct{ *bluecat.Provider }

func init() {
	fmt.Println("[DEBUG] bluecat: Initializing Bluecat DNS module")
	caddy.RegisterModule(Provider{})
	fmt.Println("[DEBUG] bluecat: Module registered successfully")
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	fmt.Println("[DEBUG] bluecat: CaddyModule() called")
	return caddy.ModuleInfo{
		ID: "dns.providers.bluecat",
		New: func() caddy.Module {
			fmt.Println("[DEBUG] bluecat: Creating new Provider instance")
			provider := &Provider{new(bluecat.Provider)}
			fmt.Printf("[DEBUG] bluecat: Provider created at %p, embedded provider at %p\n", provider, provider.Provider)
			return provider
		},
	}
}

// Provision sets up the module. Implements caddy.Provisioner.
func (p *Provider) Provision(ctx caddy.Context) error {
	logger := ctx.Logger(p)
	logger.Info("Provisioning Bluecat DNS provider")
	fmt.Printf("[DEBUG] bluecat: Provision() called on provider at %p, embedded at %p\n", p, p.Provider)

	if p.Provider == nil {
		fmt.Println("[DEBUG] bluecat: ERROR - embedded Provider is nil!")
		logger.Error("embedded Provider is nil")
		return fmt.Errorf("embedded Provider is nil")
	}

	fmt.Printf("[DEBUG] bluecat: Before replacement - ServerURL: '%s', Username: '%s', Password: '%s'\n",
		p.Provider.ServerURL, p.Provider.Username,
		func() string {
			if p.Provider.Password != "" {
				return "***SET***"
			}
			return "***EMPTY***"
		}())

	repl := caddy.NewReplacer()
	p.Provider.ServerURL = repl.ReplaceAll(p.Provider.ServerURL, "")
	p.Provider.Username = repl.ReplaceAll(p.Provider.Username, "")
	p.Provider.Password = repl.ReplaceAll(p.Provider.Password, "")
	p.Provider.ConfigurationName = repl.ReplaceAll(p.Provider.ConfigurationName, "")
	p.Provider.ViewName = repl.ReplaceAll(p.Provider.ViewName, "")

	fmt.Printf("[DEBUG] bluecat: After replacement - ServerURL: '%s', Username: '%s', Password: '%s', ConfigName: '%s', ViewName: '%s'\n",
		p.Provider.ServerURL, p.Provider.Username,
		func() string {
			if p.Provider.Password != "" {
				return "***SET***"
			}
			return "***EMPTY***"
		}(), p.Provider.ConfigurationName, p.Provider.ViewName)

	logger.Info("Bluecat DNS provider provisioned",
		zap.String("server_url", p.Provider.ServerURL),
		zap.String("username", p.Provider.Username),
		zap.String("configuration", p.Provider.ConfigurationName),
		zap.String("view", p.Provider.ViewName))

	fmt.Println("[DEBUG] bluecat: Provision() completed successfully")
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
	fmt.Println("[DEBUG] bluecat: UnmarshalCaddyfile() called")

	if p.Provider == nil {
		fmt.Println("[DEBUG] bluecat: ERROR - embedded Provider is nil in UnmarshalCaddyfile!")
		return fmt.Errorf("embedded Provider is nil")
	}

	for d.Next() {
		fmt.Println("[DEBUG] bluecat: Processing directive block")
		if d.NextArg() {
			fmt.Println("[DEBUG] bluecat: ERROR - unexpected argument")
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			directive := d.Val()
			fmt.Printf("[DEBUG] bluecat: Processing subdirective: %s\n", directive)
			switch directive {
			case "server_url":
				if d.NextArg() {
					p.Provider.ServerURL = d.Val()
					fmt.Printf("[DEBUG] bluecat: Set ServerURL to: %s\n", p.Provider.ServerURL)
				}
				if d.NextArg() {
					fmt.Println("[DEBUG] bluecat: ERROR - too many args for server_url")
					return d.ArgErr()
				}
			case "username":
				if d.NextArg() {
					p.Provider.Username = d.Val()
					fmt.Printf("[DEBUG] bluecat: Set Username to: %s\n", p.Provider.Username)
				}
				if d.NextArg() {
					fmt.Println("[DEBUG] bluecat: ERROR - too many args for username")
					return d.ArgErr()
				}
			case "password":
				if d.NextArg() {
					p.Provider.Password = d.Val()
					fmt.Println("[DEBUG] bluecat: Set Password (value redacted)")
				}
				if d.NextArg() {
					fmt.Println("[DEBUG] bluecat: ERROR - too many args for password")
					return d.ArgErr()
				}
			case "configuration_name":
				if d.NextArg() {
					p.Provider.ConfigurationName = d.Val()
					fmt.Printf("[DEBUG] bluecat: Set ConfigurationName to: %s\n", p.Provider.ConfigurationName)
				}
				if d.NextArg() {
					fmt.Println("[DEBUG] bluecat: ERROR - too many args for configuration_name")
					return d.ArgErr()
				}
			case "view_name":
				if d.NextArg() {
					p.Provider.ViewName = d.Val()
					fmt.Printf("[DEBUG] bluecat: Set ViewName to: %s\n", p.Provider.ViewName)
				}
				if d.NextArg() {
					fmt.Println("[DEBUG] bluecat: ERROR - too many args for view_name")
					return d.ArgErr()
				}
			default:
				fmt.Printf("[DEBUG] bluecat: ERROR - unrecognized subdirective: %s\n", directive)
				return d.Errf("unrecognized subdirective '%s'", directive)
			}
		}
	}

	fmt.Println("[DEBUG] bluecat: Validating configuration...")
	if p.Provider.ServerURL == "" {
		fmt.Println("[DEBUG] bluecat: ERROR - missing server URL")
		return d.Err("missing server URL")
	}
	if p.Provider.Username == "" {
		fmt.Println("[DEBUG] bluecat: ERROR - missing username")
		return d.Err("missing username")
	}
	if p.Provider.Password == "" {
		fmt.Println("[DEBUG] bluecat: ERROR - missing password")
		return d.Err("missing password")
	}

	fmt.Printf("[DEBUG] bluecat: Configuration validated. ServerURL=%s, Username=%s, ConfigurationName=%s, ViewName=%s\n",
		p.Provider.ServerURL, p.Provider.Username, p.Provider.ConfigurationName, p.Provider.ViewName)
	fmt.Println("[DEBUG] bluecat: UnmarshalCaddyfile() completed successfully")
	return nil
}

// GetRecords lists all the records in the zone.
// This wrapper adds debugging around the embedded provider's GetRecords method.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	fmt.Printf("[DEBUG] bluecat: GetRecords() called for zone: %s\n", zone)
	if p.Provider == nil {
		fmt.Println("[DEBUG] bluecat: ERROR - Provider is nil!")
		return nil, fmt.Errorf("provider not initialized")
	}

	// Verify credentials
	fmt.Printf("[DEBUG] bluecat: CREDENTIALS - ServerURL: '%s', Username: '%s', Password length: %d\n",
		p.Provider.ServerURL, p.Provider.Username, len(p.Provider.Password))

	fmt.Printf("[DEBUG] bluecat: Calling embedded Provider.GetRecords() for zone: %s\n", zone)
	records, err := p.Provider.GetRecords(ctx, zone)
	if err != nil {
		fmt.Printf("[DEBUG] bluecat: GetRecords() ERROR: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] bluecat: GetRecords() returned %d records\n", len(records))
	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
// This wrapper adds debugging around the embedded provider's AppendRecords method.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	fmt.Printf("[DEBUG] bluecat: ====== AppendRecords() called for zone: %s with %d records ======\n", zone, len(records))
	fmt.Printf("[DEBUG] bluecat: Provider instance at %p, embedded at %p\n", p, p.Provider)
	fmt.Printf("[DEBUG] bluecat: Context: %T\n", ctx)
	for i, rec := range records {
		rr := rec.RR()
		fmt.Printf("[DEBUG] bluecat: Record %d: Name='%s', Type='%s', TTL=%v, Data='%s'\n",
			i, rr.Name, rr.Type, rr.TTL, rr.Data)
	}

	if p.Provider == nil {
		fmt.Println("[DEBUG] bluecat: ERROR - Provider is nil!")
		return nil, fmt.Errorf("provider not initialized")
	}

	// Verify credentials are set in the embedded provider
	fmt.Printf("[DEBUG] bluecat: CREDENTIALS CHECK - ServerURL: '%s'\n", p.Provider.ServerURL)
	fmt.Printf("[DEBUG] bluecat: CREDENTIALS CHECK - Username: '%s'\n", p.Provider.Username)
	fmt.Printf("[DEBUG] bluecat: CREDENTIALS CHECK - Password: '%s'\n", func() string {
		if p.Provider.Password == "" {
			return "***EMPTY***"
		}
		if len(p.Provider.Password) < 4 {
			return "***TOO_SHORT***"
		}
		return "***SET_LENGTH_" + fmt.Sprintf("%d", len(p.Provider.Password)) + "***"
	}())
	fmt.Printf("[DEBUG] bluecat: CREDENTIALS CHECK - ConfigurationName: '%s'\n", p.Provider.ConfigurationName)
	fmt.Printf("[DEBUG] bluecat: CREDENTIALS CHECK - ViewName: '%s'\n", p.Provider.ViewName)

	fmt.Printf("[DEBUG] bluecat: About to call embedded Provider.AppendRecords()...\n")
	fmt.Printf("[DEBUG] bluecat: EXACT PARAMETERS:\n")
	fmt.Printf("[DEBUG] bluecat:   zone: '%s' (length: %d, hex: % x)\n", zone, len(zone), []byte(zone))
	fmt.Printf("[DEBUG] bluecat:   records count: %d\n", len(records))
	for i, rec := range records {
		rr := rec.RR()
		fmt.Printf("[DEBUG] bluecat:   Record[%d]: Name='%s' (len:%d, hex:% x)\n", i, rr.Name, len(rr.Name), []byte(rr.Name))
		fmt.Printf("[DEBUG] bluecat:   Record[%d]: Type='%s', TTL=%v\n", i, rr.Type, rr.TTL)
		fmt.Printf("[DEBUG] bluecat:   Record[%d]: Data='%s' (len:%d, hex:% x)\n", i, rr.Data, len(rr.Data), []byte(rr.Data))
	}

	fmt.Printf("[DEBUG] bluecat: Calling embedded Provider.AppendRecords() for zone: %s\n", zone)
	created, err := p.Provider.AppendRecords(ctx, zone, records)
	if err != nil {
		fmt.Printf("[DEBUG] bluecat: AppendRecords() ERROR: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] bluecat: AppendRecords() successfully created %d records\n", len(created))
	return created, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// This wrapper adds debugging around the embedded provider's SetRecords method.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	fmt.Printf("[DEBUG] bluecat: SetRecords() called for zone: %s with %d records\n", zone, len(records))
	if p.Provider == nil {
		fmt.Println("[DEBUG] bluecat: ERROR - Provider is nil!")
		return nil, fmt.Errorf("provider not initialized")
	}

	fmt.Printf("[DEBUG] bluecat: Calling embedded Provider.SetRecords() for zone: %s\n", zone)
	updated, err := p.Provider.SetRecords(ctx, zone, records)
	if err != nil {
		fmt.Printf("[DEBUG] bluecat: SetRecords() ERROR: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] bluecat: SetRecords() successfully updated %d records\n", len(updated))
	return updated, nil
}

// DeleteRecords deletes the specified records from the zone.
// This wrapper adds debugging around the embedded provider's DeleteRecords method.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	fmt.Printf("[DEBUG] bluecat: ====== DeleteRecords() called for zone: %s with %d records ======\n", zone, len(records))
	for i, rec := range records {
		rr := rec.RR()
		fmt.Printf("[DEBUG] bluecat: Record %d to delete: Name='%s', Type='%s'\n", i, rr.Name, rr.Type)
	}

	if p.Provider == nil {
		fmt.Println("[DEBUG] bluecat: ERROR - Provider is nil!")
		return nil, fmt.Errorf("provider not initialized")
	}

	fmt.Printf("[DEBUG] bluecat: Calling embedded Provider.DeleteRecords() for zone: %s\n", zone)
	deleted, err := p.Provider.DeleteRecords(ctx, zone, records)
	if err != nil {
		fmt.Printf("[DEBUG] bluecat: DeleteRecords() ERROR: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] bluecat: DeleteRecords() successfully deleted %d records\n", len(deleted))
	return deleted, nil
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
