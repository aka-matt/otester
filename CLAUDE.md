# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**otester** is a desktop API debugging tool built with Go + Wails (Windows-first, cross-platform). It reads API endpoint configurations from `config.json` in the working directory, supports Microsoft OAuth 2.0 Client Credentials authentication, and provides a Fluent Design acrylic UI for constructing/sending HTTP requests and viewing responses.

## Tech Stack

- **Backend**: Go, Wails framework, standard `net/http`, `encoding/json`, `context.Context`
- **Frontend**: React or Vue 3 + TypeScript (TBD - see PRD section 13)
- **UI**: CSS Variables for theming, `backdrop-filter` for acrylic effect, Monaco Editor or CodeMirror for JSON editing
- **Icons**: Fluent UI System Icons or Lucide (linear style only)

## Architecture

### Directory Structure (per PRD section 13.2)

```
/
├─ main.go
├─ config.json
├─ internal/
│  ├─ app/app.go
│  ├─ config/{loader.go, model.go, validator.go}
│  ├─ httpclient/{client.go, request_builder.go, response.go}
│  ├─ oauth/{microsoft.go, cache.go, model.go}
│  ├─ security/redact.go
│  └─ history/store.go
├─ frontend/
│  ├─ src/
│  │  ├─ components/
│  │  ├─ views/
│  │  ├─ stores/
│  │  ├─ themes/
│  │  └─ types/
│  └─ package.json
└─ README.md
```

### Key Constraints

1. **Configuration**: Must read `config.json` from the current working directory (not user home or system dirs)
2. **OAuth flow**: Client Credentials only (no Authorization Code, Device Code)
3. **Token storage**: In-memory only, never persisted to disk
4. **Security**: `client_secret` and full `access_token` must never appear in logs or UI; secrets masked in UI by default
5. **No direct API calls from frontend**: All HTTP requests (including OAuth token requests) go through Go backend

### Wails Backend Interface (per PRD section 13.3)

The Go backend exposes these methods to the frontend:
- `LoadConfig() (*ConfigView, error)` - load config, secrets sanitized
- `ReloadConfig() (*ConfigView, error)` - re-read config.json
- `ValidateConfig() (*ValidationResult, error)`
- `SendRequest(input RequestInput) (*ResponseOutput, error)` - async HTTP request
- `CancelRequest(requestID string) error`
- `ClearTokenCache() error`
- `GetTokenStatus(profileID string) (*TokenStatus, error)`
- `GetAppInfo() (*AppInfo, error)`
- `OpenConfigDirectory() error`
- `OpenLogDirectory() error`

### OAuth Token Caching

Tokens are cached in memory using a key composed of `token_url + client_id + scope`. Uses `singleflight` to prevent concurrent refreshes for the same profile. Tokens are refreshed `refresh_before_expiry_seconds` (default 60s) before expiry.

### Three-Layout UI (per PRD section 8.2)

```
┌─────────────────────────────────────────────────────┐
│  Top menu / Theme toggle / Config status / Send      │
├─────────────┬────────────────────────┬──────────────┤
│  Endpoint   │  Request Editor        │  Response    │
│  List       │  URL / Params / Body   │  Status      │
│  (hideable) │  Headers / Auth        │  & Content   │
├─────────────┴────────────────────────┴──────────────┤
│  Status bar: config path, network status, version    │
└─────────────────────────────────────────────────────┘
```

## Configuration Schema

See PRD section 10 for full schema. Key points:

- `app` section: name, title, default_timeout_seconds (default 30s), max_response_body_bytes (default 10MB)
- `oauth_profiles`: Microsoft Client Credentials profiles with `org_id_uuid`, `client_id`, `client_secret`, `scope`
- `endpoints`: API endpoints with method, url (supports `{{variable}}` substitution), auth (oauth2 or none)
- `variables`: Global substitution variables accessible in endpoint URLs, headers, body

## Theme System

Three modes: `light`, `dark`, `system` (follows OS). Theme preference stored in frontend Local Storage, not in `config.json`. Acrylic effect uses `backdrop-filter: blur(18px) saturate(135%)` with graceful degradation on platforms without support.

## Important PRD Decisions (not to be re-litigated)

- OAuth Scope is per-profile (not always Microsoft Graph) - see section 24 item 2
- TLS cert validation can be disabled via `allow_insecure_tls` but UI shows warning - section 24 item 7
- Business API requests may follow redirects with security policy applied - section 24 item 8
- Config is read-only in v1 (no runtime modification/saving) - section 24 item 5
- Request history not persisted to disk in v1 - section 24 item 6
- `client_secret` stored in plaintext in config.json - section 24 item 1

## When Implementing

1. Follow the milestone order in PRD section 22: M1 (foundation) → M2 (HTTP debugging) → M3 (OAuth) → M4 ( polish) → M5 (release)
2. Implement error codes from PRD section 17
3. Log levels: Debug, Info, Warn, Error - never log secrets (section 16)
4. Accessibility: keyboard navigation, shortcuts (section 20), sufficient color contrast
