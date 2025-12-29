# Patbin

A modern, minimalist pastebin built with Go and Gin framework. Features syntax highlighting, dark/light themes, optional authentication, and a responsive UI.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## Features

- **Syntax Highlighting** - Auto-detect via URL extension (e.g., `/abc123.go`, `/abc123.py`)
- **Dark/Light Mode** - System-aware with manual toggle
- **Optional Authentication** - Anonymous pastes + login for edit/delete
- **Public/Private Pastes** - Control visibility of your pastes
- **Expiring Pastes** - Set TTL: 1 hour, 1 day, 1 week, or never
- **Burn After Read** - Self-destructing pastes
- **Fork Pastes** - Create copies of existing pastes
- **User Profiles** - Shareable list of public pastes
- **Line Numbers** - Click to link to specific lines
- **Mobile-First Design** - Responsive, touch-friendly UI
- **Copy to Clipboard** - One-click copying
- **Keyboard Shortcuts** - `Ctrl+Enter` to submit, `Ctrl+S` to save

## Quick Start

### Prerequisites
- Go 1.21 or higher

### Run Locally

```bash
git clone https://github.com/AmarnathCJD/PatBin.git
cd PatBin
go mod download
go run main.go
```

Visit http://localhost:8080

### Using Docker

```bash
docker build -t patbin .
docker run -p 8080:8080 -v patbin-data:/app/data patbin
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `JWT_SECRET` | `patbin-super-secret...` | JWT signing key |
| `DB_PATH` | `patbin.db` | SQLite database path |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/paste` | Create new paste |
| `GET` | `/api/paste/:id` | Get paste |
| `PUT` | `/api/paste/:id` | Update paste (auth) |
| `DELETE` | `/api/paste/:id` | Delete paste (auth) |
| `POST` | `/api/paste/:id/fork` | Fork a paste |
| `POST` | `/api/auth/register` | Create account |
| `POST` | `/api/auth/login` | Login |
| `POST` | `/api/auth/logout` | Logout |

## Syntax Highlighting

Access pastes with file extension:
- `/abc123.go` - Go
- `/abc123.py` - Python  
- `/abc123.js` - JavaScript
- `/abc123.rs` - Rust

## License

MIT License
