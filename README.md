# Safi-Budget: Smart Personal Finance CLI

A robust, standalone Command-Line Interface (CLI) personal budgeting utility engineered in Go. The application assists users in optimizing their net monthly income utilizing a structural 50/30/20 financial allocation framework (Needs, Wants, Savings) while protecting financial data via a local persistent ledger file system.

## Project Architecture & Design Principles
The application operates entirely independent of external shared architectures or databases, ensuring strict localized personal data isolation.

1. **Memory State Layer (`main.go`):** Manages dynamic runtime memory allocations, validation boundaries, safety overrides, and algorithmic splits.
2. **Persistence State Layer (`income.json` / `expenses.json`):** Holds the flat-file database schema, recording real-time mutations to prevent historical data loss across runtime sessions.

---

## Core Features & Technical Solutions

### 1. Terminal Buffer Blockade Override
To permanently resolve standard command-line interface input lockups or deadlocks common in native Windows console hosting environments, this application implements a structured processing loop using standard readers. The system captures input text directly from standard input, dynamically truncates trailing hidden carriage returns (`\r`, `\n`), and safely parses numerical data conversions without utilizing raw pointers that cause memory freezes.

### 2. Envelope Budgeting System
The baseline 50% "Needs" capital allocation is strictly partitioned into explicit operational sub-envelopes to maintain tracking precision:
- **Rent Envelope:** Tracks housing metrics.
- **Food Envelope:** Logs essential sustenance costs.
- **Transport Envelope:** Tracks commuting fees.
- **Utilities Envelope:** Monitors mandatory public services bills.

### 3. Real-Time Cascade Waterfall Engine
If a core budget bucket (like Needs) falls into a deficit from overspending, the math engine automatically cannibalizes liquidity from the lifestyle bucket (Wants) first, and breaks into emergency reserves (Savings) next. This prevents fake positive indicators and forces the console to reflect true liquid reality.

### 4. Secure Ledger Purging Protocol
Includes an isolated maintenance runtime option (Option 4) that safely flushes local JSON file data back to a clean slate without requiring manual directory deletion or risking schema corruption.

---

## Getting Started

### Prerequisites
- Go 1.18 or higher installed on your local environment.

### Run locally
To launch the app locally, run:

```bash
go run ./cmd/safi-budget
```

Then open http://localhost:8080 in your browser.

### Deploy for sharing
This project is now ready for deployment on platforms like Render, Railway, Fly.io, or any Docker host.

#### Option 1: Docker
```bash
docker build -t safi-budget .
docker run -p 8080:8080 safi-budget
```

#### Option 2: Render
1. Push this repository to GitHub.
2. Create a new Web Service on Render.
3. Connect the repo and use the included render.yaml file.

### Installation for peers
If you want peers to use it easily, the best approach is:
- host it online with a cloud platform,
- share the public URL,
- optionally package it as a desktop app later if you want a native install experience.

For a personal-use desktop install, you could also wrap this web app into a simple Electron or Tauri app later, but the current version is already suitable for web-based sharing and review.