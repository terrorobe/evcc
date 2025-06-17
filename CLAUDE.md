# EVCC - Claude Development Guide

## Project Overview

**evcc** is an extensible EV Charge Controller and home energy management system written in Go with a Vue.js frontend. It provides local energy management without relying on cloud services, featuring extensive device support and integration capabilities.

### Key Features
- EV charger control and monitoring
- Solar inverter and battery system integration
- Energy meter support
- Vehicle integration (SoC, remote charge control)
- Tariff management and smart charging
- Home automation system integration

## Architecture

### Backend (Go)
- **Main Language**: Go 1.24+
- **Framework**: Custom HTTP server with Gorilla Mux router
- **Database**: SQLite with GORM
- **Real-time**: WebSocket and MQTT support
- **Device Communication**: Modbus, HTTP, MQTT, WebSocket, EEBus, OCPP

### Frontend (Vue.js)
- **Framework**: Vue 3 with TypeScript
- **Build Tool**: Vite
- **UI**: Bootstrap 5
- **Charts**: Chart.js
- **State Management**: Vue stores
- **Internationalization**: Vue I18n (25+ languages)

## Directory Structure

```
├── main.go                 # Application entry point
├── cmd/                    # CLI commands and configuration
├── core/                   # Core business logic
│   ├── loadpoint/         # Charging point management
│   ├── site/              # Site-wide energy management
│   └── vehicle/           # Vehicle integration
├── charger/               # Charger device implementations
├── meter/                 # Energy meter implementations  
├── vehicle/               # Vehicle API integrations
├── tariff/                # Energy tariff providers
├── server/                # HTTP server and APIs
├── assets/                # Frontend source code
│   ├── js/               # Vue.js application
│   └── css/              # Stylesheets
├── templates/             # Device configuration templates
└── util/                  # Utility packages
```

## Development Commands

### Prerequisites Setup
```bash
make install          # Install Go tools
make install-ui       # Install Node.js dependencies (npm ci)
```

### Development Workflow
```bash
make ui              # Build frontend (npm run build)
make build           # Build Go binary
make default         # Build UI and binary
make all             # Full build with linting and tests
```

### Testing & Quality
```bash
make test            # Run Go tests
make test-ui         # Run frontend tests (npm test)
make lint            # Run Go linter (golangci-lint)
make lint-ui         # Run frontend linting (npm run lint)
```

### Useful Development Commands
```bash
make clean           # Clean build artifacts
make assets          # Generate embedded assets (go generate)
make docs            # Generate template documentation
```

## Device Templates

EVCC uses YAML templates to define device configurations in `/templates/definition/`:
- `charger/` - EV charger definitions
- `meter/` - Energy meter definitions
- `vehicle/` - Vehicle API definitions
- `tariff/` - Tariff provider definitions

Test new templates with:
```bash
evcc --template-type charger --template new-charger-template.yaml
```

## Technology Stack

### Backend Dependencies
- **HTTP**: Gorilla Mux, HTTP utilities
- **Database**: GORM with SQLite
- **Device Communication**: 
  - Modbus (custom fork)
  - MQTT (Eclipse Paho)
  - OCPP (custom fork)
  - EEBus integration
- **Authentication**: OAuth2, JWT
- **Monitoring**: InfluxDB client, Prometheus metrics

### Frontend Dependencies
- **Vue 3** with Composition API
- **TypeScript** for type safety
- **Vite** for fast development builds
- **Bootstrap 5** for UI components
- **Chart.js** for energy visualization
- **Axios** for HTTP requests
- **MQTT.js** for real-time updates

## Supported Integrations

### Chargers (100+ models)
ABB, Alfen, Easee, go-e, KEBA, Tesla, Wallbe, and many more including OCPP support

### Energy Meters (100+ models) 
SMA, Fronius, Victron, Shelly, Tasmota, and many inverter/battery systems

### Vehicles (20+ brands)
Tesla, BMW, Audi, VW, Mercedes, Hyundai, Kia, and more via cloud APIs

### Home Automation
Home Assistant, OpenHAB, MQTT, REST APIs

## Development Environment Requirements

- **Go**: 1.24+ (currently using 1.24.4)
- **Node.js**: 22+ (currently using 24.2.0)
- **npm**: 10+ (currently using 11.3.0)

## Common Development Tasks

1. **Adding a new charger**: Create template in `templates/definition/charger/`
2. **Adding vehicle support**: Implement in `vehicle/` directory
3. **Frontend changes**: Work in `assets/js/` directory
4. **API changes**: Modify `server/` and update frontend accordingly
5. **Device communication**: Use existing plugins or extend `plugin/` directory

## Build Process

The project uses a multi-stage build:
1. Frontend build (`npm run build`) creates optimized assets
2. Go build embeds assets and creates single binary
3. Assets are embedded using `go generate` and `embed.go` files

## Testing

- **Go tests**: Standard Go testing with testify
- **Frontend tests**: Vitest with happy-DOM
- **E2E tests**: Playwright for end-to-end testing
- **CI/CD**: GitHub Actions with comprehensive testing

This codebase is well-structured for an energy management system with extensive device support and a modern web interface.