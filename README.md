# LLM Proxy

A lightweight, high-performance LLM API Proxy designed for individual use. Aggregates OpenAI, Anthropic, and Gemini into a single OpenAI-compatible endpoint.

## Features

- **Unified API**: OpenAI-compatible chat completion endpoint.
- **Multi-Provider Support**: Built-in adapters for OpenAI, Anthropic, and Gemini.
- **Key Rotation**: Round-robin load balancing across multiple API keys per provider.
- **Secure**: JWT-based admin panel and Bearer token (`sk-xxx`) client authentication.
- **Lightweight**: Built with Go, Gin, and SQLite (pure Go, no CGO).
- **Embedded Migrations**: Simple deployment with automated SQL schema setup.

## Getting Started

### 1. Build
```bash
make build
```

### 2. Run
```bash
./bin/llm-proxy
```
The server uses environment variables for configuration (optional):
- `PORT`: Server port (default: 8080)
- `DB_PATH`: Path to SQLite DB (default: data/llm-proxy.db)
On the first run, it will generate a random admin password and JWT secret. **Make sure to save them!**

## API Reference

### Client API (sk-xxx Auth)
- `POST /v1/chat/completions`: Chat completion (OpenAI format)
- `GET /v1/models`: List available models

### Admin API (JWT Auth)
- `POST /auth/login`: Admin login
- `GET /admin/providers`: Manage providers
- `POST /admin/tokens`: Generate new client access tokens
- `GET /admin/logs`: View request logs

## License
MIT
