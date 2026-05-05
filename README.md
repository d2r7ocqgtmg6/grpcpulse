# grpcpulse

Lightweight health-check daemon for gRPC services that exposes Prometheus-compatible metrics.

---

## Installation

```bash
go install github.com/yourorg/grpcpulse@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/grpcpulse.git
cd grpcpulse
go build -o grpcpulse ./cmd/grpcpulse
```

---

## Usage

Define your targets in a YAML config file:

```yaml
# config.yaml
interval: 15s
metrics_port: 9090
targets:
  - name: auth-service
    address: localhost:50051
  - name: order-service
    address: localhost:50052
```

Run the daemon:

```bash
grpcpulse --config config.yaml
```

Metrics will be available at `http://localhost:9090/metrics` in Prometheus format.

```
# EXAMPLE METRICS
grpcpulse_check_duration_seconds{target="auth-service"} 0.003
grpcpulse_check_status{target="auth-service"} 1
grpcpulse_check_status{target="order-service"} 0
```

Scrape the endpoint with Prometheus by adding it to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: grpcpulse
    static_configs:
      - targets: ["localhost:9090"]
```

---

## License

[MIT](LICENSE)