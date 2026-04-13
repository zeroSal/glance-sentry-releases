# Glance Sentry Releases

A [Glance](https://github.com/glanceapp/glance) plugin that displays Sentry releases with adoption metrics directly on your dashboard.

This project provides a `@widget.yml` configuration for Glance and a proxy server that fetches data from the Sentry API.

## Features

- Fetches project list from Sentry API
- Retrieves release health data with adoption metrics
- Serves aggregated data via HTTP proxy

## Requirements

- Go 1.21+
- Sentry account with API access

## Setup

1. Set environment variables:

```bash
export SENTRY_ORG=your-org-slug
export SENTRY_AUTH_TOKEN=your-auth-token
```

2. Run the server:

```bash
go run main.go serve
```

The server listens on `127.0.0.1:8099` by default. Override with `GLANCE_SENTRY_PORT`.

## Glance Widget Configuration

This project includes `@widget.yml` - a Glance widget configuration that displays your Sentry releases directly on your Glance dashboard.

### Adding to Glance

1. Copy `widget.yml` into your Glance dashboard's widgets folder
2. Reference it in your Glance config file:

```yaml
- $include: widget.yml
```

> **Important:** Do NOT add the `cache` attribute to this widget in your Glance config. The caching is handled directly by the proxy server, not by Glance.

The widget fetches data from the local proxy server (`http://127.0.0.1:8099/`) and displays:

- Number of monitored projects
- Each project with its current release version
- Adoption percentage with visual progress bar
- New issues count
- Release creation date

### Environment Variables for Widget

The widget requires these environment variables set in your Glance configuration:

- `SENTRY_ORG` - Your Sentry organization slug
- `GLANCE_SENTRY_PORT` - Port where the proxy server runs (default: `8099`)
- `GLANCE_SENTRY_HOST` - Host where the proxy server runs (default: `127.0.0.1`)

## Usage

```bash
# Start the proxy server
go run main.go serve

# Fetch releases with adoption data
curl http://127.0.0.1:8099/
```

### Environment Variables

| Variable                 | Required | Default     | Description               |
| ------------------------ | -------- | ----------- | ------------------------- |
| `SENTRY_ORG`             | Yes      | -           | Sentry organization slug  |
| `SENTRY_AUTH_TOKEN`      | Yes      | -           | Sentry API token          |
| `GLANCE_SENTRY_PORT`     | No       | `8099`      | Server binding port       |
| `GLANCE_SENTRY_HOST`     | No       | `127.0.0.1` | Server binding address    |
| `CACHE_INTERVAL_MINUTES` | No       | 5           | How often to refresh data |
