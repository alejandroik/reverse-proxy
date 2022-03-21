# reverse-proxy

Is a simple reverse-proxy written in go, it provides rate-limiting and statistics/metrics through Prometheus/Grafana.

# Install

### Pre-compiled executables

Get them [here](https://github.com/alejandroik/reverse-proxy/releases).

## Configuration

Create a config file in the in the same directory as the binary. reverse-proxy supports all the popular file extensions thanks to viper.

```bash
touch config.yaml
```

Define Limiters by endpoint, using `rate_limit` to limit all requests to a specific endpoint, `client_rate_limit` to limit by client IP, or a combination of both.

```yaml
- endpoint: "/categories"
    rate_config:
      rate_limit: 1
      client_rate_limit: 1
      clean_interval: 5
```

A global limiter can be defined for all requests to the proxy with:

```yaml
- endpoint: "/"
    rate_config:
      client_rate_limit: 10
```

Example configuration file

```yaml
server:
  port: "8080"
  remote_host: https://remote-host.com

Limiters:
  - endpoint: "/"
    rate_config:
      client_rate_limit: 10

  - endpoint: "/categories"
    rate_config:
      rate_limit: 1
      client_rate_limit: 1
      clean_interval: 5

  - endpoint: "/items"
    rate_config:
      rate_limit: 50
      client_rate_limit: 1
      clean_interval: 5
```

## Usage

Start the proxy

```bash
./reverse-proxy
2022-03-21T12:19:17.220-0300	INFO	limiter/limiter.go:57	[Limiter] Started limiter for /
2022-03-21T12:19:17.220-0300	INFO	limiter/limiter.go:57	[Limiter] Started limiter for /categories
2022-03-21T12:19:17.220-0300	INFO	limiter/limiter.go:57	[Limiter] Started limiter for /items
2022-03-21T12:19:17.220-0300	INFO	reverse-proxy/main.go:32	Listening on 8080
```

## License

[MIT](https://choosealicense.com/licenses/mit/)
