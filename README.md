# Loki Data Connector

The Hasura Loki Connector allows you to connect to a Loki database giving you an instant GraphQL API on top of your [Loki data](https://grafana.com/docs/loki/latest/).

This connector is built using the [Go Data Connector SDK](https://github.com/hasura/ndc-sdk-go) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

## Get Started

Check out [the official quickstart documentation of Hasura](https://hasura.io/docs/3.0/getting-started/quickstart). By default, the connector serves [Loki HTTP APIs](https://grafana.com/docs/loki/latest/reference/loki-http-api/). You can execute LogQL queries directly.

### Add models

See [Concepts](./docs/concepts.md) to know how to define models.

## Documentation

- [Concepts](./docs/concepts.md)
- [Integrate Loki Connector with Grafana](./docs/grafana.md)

## Contributing

Check out our [contributing guide](./docs/contributing.md) for more details.

## License

Loki Connector is available under the [Apache License 2.0](./LICENSE).
