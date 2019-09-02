# vsync

Hashicorp Vault one-way secrets sync tool (written in golang).

## Installation

`go get -u github.com/flaccid/vsync/cmd`

## Usage

`vsync --help`

### Wrapper/Helper Commands

#### Requests

Make an arbitrary request to the source vault:

```
vsync request GET /sys/health
```

Or on the destination vault:

```
vsync request -ds GET /sys/health
```

## TODO

- support other auth methods; see https://github.com/lanceplarsen/go-vault-demo/blob/master/client/vault.go
- look at the higher ground; see https://github.com/tuenti/secrets-manager

### Helm Chart

Validate the chart:

`helm lint charts/vsync`

Dry run and print out rendered YAML:

`helm install --dry-run --debug --name vsync charts/vsync`

Install the chart:

`helm install --name vsync charts/vsync`

Or, with a more complete config from env:

```
helm install \
  --name vsync charts/vsync \
  --set vault.source.address="$VAULT_ADDR" \
  --set vault.source.token="$VAULT_TOKEN" \
  --set vault.destination.address="$DESTINATION_VAULT_ADDR" \
  --set vault.destination.token="$DESTINATION_VAULT_TOKEN"
```

Upgrade the chart:

`helm upgrade vsync charts/vsync`

Testing after deployment:

`helm test vsync`

Completely remove the chart:

`helm delete --purge vsync`

## Development

The easiest way to get testing is to just run up a local vault dev server:

`vault server -dev`

Login to this vault, getting the token printed the server's stdout:

`vault login -token=abcd123`

Take note of https://stackoverflow.com/questions/49872480/vault-error-while-writing.
You may need to change to the v1 secrets engine:

```
vault secrets disable secret
vault secrets enable -version=1 -path=secret kv
```

You can use CLI options, but it may be easier to just add some settings to env:

```
export VAULT_ADDR=https://my-source-vault.suf:8200/
export VAULT_TOKEN=my-source-vault-token
export DESTINATION_VAULT_ADDR=http://localhost:8200
export DESTINATION_VAULT_TOKEN=my-destination-vault-token
# export VSYNC_LOG_LEVEL=debug
```

Now try a sync:

`vsync sync-secrets`

## License

- Author: Chris Fordham (<chris@fordham.id.au>)

```text
Copyright 2019, Chris Fordham

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
