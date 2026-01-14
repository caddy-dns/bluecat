Bluecat module for Caddy
===========================

This package contains a DNS provider module for [Caddy](https://github.com/caddyserver/caddy). It can be used to manage DNS records with Bluecat Address Manager.

## Caddy module name

```
dns.providers.bluecat
```

## Config examples

To use this module for the ACME DNS challenge, [configure the ACME issuer in your Caddy JSON](https://caddyserver.com/docs/json/apps/tls/automation/policies/issuer/acme/) like so:

```json
{
	"module": "acme",
	"challenges": {
		"dns": {
			"provider": {
				"name": "bluecat",
				"server_url": "{env.BLUECAT_SERVER_URL}",
				"username": "{env.BLUECAT_USERNAME}",
				"password": "{env.BLUECAT_PASSWORD}",
				"configuration_name": "{env.BLUECAT_CONFIGURATION_NAME}",
				"view_name": "{env.BLUECAT_VIEW_NAME}"
			}
		}
	}
}
```

or with the Caddyfile:

```caddyfile
# globally
{
	acme_dns bluecat {
		server_url {env.BLUECAT_SERVER_URL}
		username {env.BLUECAT_USERNAME}
		password {env.BLUECAT_PASSWORD}
		configuration_name {env.BLUECAT_CONFIGURATION_NAME}  # optional
		view_name {env.BLUECAT_VIEW_NAME}                    # optional
	}
}
```

```caddyfile
# one site
tls {
	dns bluecat {
		server_url {env.BLUECAT_SERVER_URL}
		username {env.BLUECAT_USERNAME}
		password {env.BLUECAT_PASSWORD}
		configuration_name {env.BLUECAT_CONFIGURATION_NAME}  # optional
		view_name {env.BLUECAT_VIEW_NAME}                    # optional
	}
}
```

## Configuration Fields

- **server_url** (required): The base URL of your Bluecat Address Manager server (e.g., `https://bluecat.example.com`)
- **username** (required): Username for authenticating with the Bluecat API
- **password** (required): Password for authenticating with the Bluecat API
- **configuration_name** (optional): Bluecat configuration name (defaults to first available)
- **view_name** (optional): Bluecat view name (defaults to first available)

If you'd rather directly add the config items you can forgo the .env file.