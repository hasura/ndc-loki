# Concepts

## Model

> [!NOTE]
> See [Loki's Log queries](https://grafana.com/docs/loki/latest/query/log_queries/).

Model is a structured definition of a prepared LogQL query to define:

- Available labels.
- Pipelines to pre-filter, parse fields and format the log line.

The equivalent query collections always prepend the prepared query to every request.

For example, the model for nginx access logs with the following content is:

```log
172.21.0.7 - - [04/Nov/2024:16:59:05 +0000] "GET /loki/api/v1/query_range HTTP/1.1" 200 831 "-" "Go-http-client/1.1"
```

```yaml
metadata:
  models:
    nginx_log:
      pipelines:
        - type: line_filter
          operator: _regex
          value: ^[0-9]+\.
        - type: pattern
          pattern: <remote_addr> - <remote_user> [<time_local>] "<request>" <status> <body_bytes_sent> "<http_referer>" "<http_user_agent>" <request_length> <request_time> [<proxy_upstream_name>] [<proxy_alternative_upstream_name>] <upstream_addr> <upstream_response_length> <upstream_response_time> <upstream_status> <req_id>
          fields:
            remote_addr: {}
            remote_user: {}
            time_local: {}
            request: {}
            status: {}
            body_bytes_sent: {}
            http_referer: {}
            http_user_agent: {}
            request_length: {}
            request_time: {}
      labels:
        container_name:
          source: container
          filter:
            operator: _eq
            value: ndc-loki-gateway-1
            static: false
        service_name:
          filter:
            operator: _eq
            value: ndc-loki-gateway-1
            static: true
```

The model use the `pattern` parser to parse fields from the raw log line. Labels and fields will be available in the `nginx_log` collection. All GraphQL query request of this model will have a prefix:

```
{container="ndc-loki-gateway-1", service_name="ndc-loki-gateway-1"} |~ `^[0-9]+\.` | pattern `<remote_addr> - <remote_user> [<time_local>] "<request>" <status> <body_bytes_sent> "<http_referer>" "<http_user_agent>" <request_length> <request_time> [<proxy_upstream_name>] [<proxy_alternative_upstream_name>] <upstream_addr> <upstream_response_length> <upstream_response_time> <upstream_status> <req_id>`
```

Query predicates which are inputted in the GraphQL query are [label filter expressions](https://grafana.com/docs/loki/latest/query/log_queries/#label-filter-expression) and are appends to the prepared query. For example, with the following GraphQL query, the LogQL query will be:

```graphql
query GetNginxSuccessRequests($from: Timestamp!, $to: Timestamp!) {
  nginx_log(
    args: {}
    where: { timestamp: { _gt: $from, _lt: $to }, status: { _eq: "200" } }
  ) {
    status
    log_lines {
      timestamp
      value
    }
  }
}
```

```log
{container="ndc-loki-gateway-1", service_name="ndc-loki-gateway-1"} |~ `^[0-9]+\.` | pattern ... | status = `200`
```

Log lines are returned as a list of timestamp and value pairs in the `log_lines` array.

## Aggregation

> [!NOTE]
> See [Loki's Metric queries](https://grafana.com/docs/loki/latest/query/metric_queries/).

Aggregation query names are defined with `<model_name>_aggregate` format. The query requires a non-empty `aggrerations` array argument that is a list of composable aggregate functions.

```graphql
query GetNginxHTTPRequestsOverTime($from: Timestamp!, $to: Timestamp!) {
  nginx_log_aggregate(
    args: {
      aggregations: [{ count_over_time: "1m" }, { sum: { by: [status] } }]
    }
    where: { timestamp: { _gt: $from, _lt: $to } }
  ) {
    status
    metric_values {
      timestamp
      value
    }
  }
}
```

```log
sum by (status) (count_over_time({container="ndc-loki-gateway-1", service_name="ndc-loki-gateway-1"} |~ `^[0-9]+\.` | pattern ... [1m]))
```

Metric values are returned as a list of `timestamp` and `value` pairs in the `metric_values` array.

## Flat Results

By default, time series results are grouped by a set of unique labels ([source](https://grafana.com/docs/loki/latest/reference/loki-http-api/#examples-2)). However, they aren't compatible with Grafana because the Wild GraphQL Data Source plugin only understands the flat time series data. Therefore Loki connector supports a `flat` option to flatten results into the root array.

**Log**

```graphql
query GetNginxSuccessRequests($from: Timestamp!, $to: Timestamp!) {
  nginx_log(
    args: { flat: true }
    where: { timestamp: { _gt: $from, _lt: $to }, status: { _eq: "200" } }
  ) {
    status
    timestamp
    log_line
  }
}
```

**Aggregation**

```graphql
query GetNginxHTTPRequestsOverTime($from: Timestamp!, $to: Timestamp!) {
  nginx_log_aggregate(
    args: {
      flat: true
      aggregations: [{ count_over_time: "1m" }, { sum: { by: [status] } }]
    }
    where: { timestamp: { _gt: $from, _lt: $to } }
  ) {
    status
    timestamp
    metric_value
  }
}
```

The `flat` option can be globally set in runtime settings in the configuration or explicitly set in the request argument.

> See also: [Grafana Integration](./grafana.md).

## Native Query

When simple queries don't meet your need you can define native queries with prepared variables with the `${<name>}` template. Native queries are defined as collections.

```yaml
metadata:
  native_operations:
    queries:
      hasura_log_count:
        type: metric
        query: 'count by (type) (rate({service_name="ndc-loki-graphql-engine-1", container="ndc-loki-graphql-engine-1"} | json level="level", type="type" | type = `${type}` | level =~ `${level}` [$range]))'
        labels:
          type: {}
        arguments:
          range:
            type: Duration
          type:
            type: String
          level:
            type: String
```

A native query is defined as a `stream` or `metric` query only. So the `type` is required.

Arguments of the native query will be required in the `args` object. Boolean expressions in `where` are also supported. However, they are are used to filter results after the query was executed. It's useful for filtering permissions.

```gql
{
  hasura_log_count(
    args: { step: "1m", type: "http-log", level: "info|error" }
    where: { timestamp: { _gt: "2024-10-11T00:00:00Z" } }
  ) {
    type
    values {
      timestamp
      value
    }
  }
}
```
