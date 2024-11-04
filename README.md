# Loki Data Connector

The Hasura Loki Connector allows for connecting to a Loki database giving you an instant GraphQL API on top of your [Loki data](https://grafana.com/docs/loki/latest/).

This connector is built using the [Go Data Connector SDK](https://github.com/hasura/ndc-sdk-go) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

## Development

### Get started

#### Start Docker services

```sh
make start-ddn
```

#### Introspect and build DDN metadata

```sh
make build-supergraph-test
```

Browse the engine console at http://localhost:3000 and Grafana at http://localhost:3001
