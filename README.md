[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/Lavaburn/sensu-squadcast-webhook-handler)
![Go Test](https://github.com/Lavaburn/sensu-squadcast-webhook-handler/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/Lavaburn/sensu-squadcast-webhook-handler/workflows/goreleaser/badge.svg)

# Sensu Squadcast Webhook Handler

The Sensu Go Squadcast Webhook handler is a [Sensu Event Handler][1] that sends event data to
a [Squadcast](https://www.squadcast.com/) endpoint using the Incident Webhook. Refer [Squadcast documentation](https://support.squadcast.com/docs/apiv2) for more details.

Reason for using the Incident Webhook URL and not the build-in Sensu Go URL is to avoid the pre-defined indident description.
This handler allows you to specify a template file (which allows Markdown) instead.  

## Configuration

### Asset registration

Run the following command to add the asset `sensu-squadcast-webhook-handler`.

```shell
sensuctl asset add Lavaburn/sensu-squadcast-webhook-handler
```

### Handler definition

Example Sensu Go handler definition:

**squadcast-webhook-handler.yaml**

```yaml
type: Handler
api_version: core/v2
metadata:
  name: squadcast-webhook
  namespace: default
spec:
  command: sensu-squadcast-webhook-handler
  env_vars:
  - SENSU_SQUADCAST_APIURL=<Squadcast Alert Source Url>
  runtime_assets:
  - Lavaburn/sensu-squadcast-webhook-handler
  filters:
  - is_incident
  timeout: 10
  type: pipe
```

Run the following to create the handler:

```shell
sensuctl create -f squadcast-webhook-handler.yaml
```

Example Sensu Go check definition:

```yaml
api_version: core/v2
type: CheckConfig
metadata:
  namespace: default
  name: health-check
spec:
  command: check-http -u http://localhost:8080/health
  subscriptions:
  - test
  publish: true
  interval: 10
  handlers:
  - squadcast-webhook
```

## Usage examples

Help:

```
The Sensu Go Squadcast Webhook handler sends Sensu events to Squadcast


Usage:
  sensu-squadcast-webhook-handler [flags]


Flags:
  -h, --help                help for sensu-squadcast-webhook-handler
  -a, --api-url string      The URL for the Squadcast API (Incident Webhook)
  -m, --message string      The template to use for the message (and de-duplication) (default "{{.Entity.Name}}/{{.Check.Name}}")
  -d, --description string  The template to use for the description (default "{{.Check.Output}}")
  -t, --template string   	The template file to use for the description (Overwrites -d argument)
```

[1]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
