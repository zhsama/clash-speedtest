# Clash SpeedTest Frontend

This is the frontend for Clash SpeedTest, built with Astro 5, React, Tailwind CSS, shadcn/ui, and TypeScript.

## Prerequisites

- Node.js 18+
- Go backend API server running on port 8090

## Setup

1. Install dependencies:
```bash
npm install
```

2. Start the backend API server:
```bash
cd ../backend
go run api-server/main.go
```

3. Start the frontend development server:
```bash
npm run dev
```

4. Open http://localhost:4321 in your browser

## Build

```bash
npm run build
```

## Features

- **Configuration File Input**: Support for local files and HTTP(S) subscription URLs
- **Proxy Filtering**: Regular expression based filtering
- **Speed Test Parameters**:
  - Download/Upload test size configuration
  - Concurrent connections setting
  - Timeout configuration
  - Latency and speed thresholds
- **Advanced Options**:
  - Stash compatibility mode
  - Node renaming with location info
- **Results Display**:
  - Real-time testing progress
  - Sortable results table
  - Color-coded performance indicators
  - Export results as JSON

## API Endpoints

The frontend expects the Go backend API server to be running on `http://localhost:8090` with the following endpoints:

- `POST /api/test` - Run speed test with configuration
- `GET /api/health` - Health check endpoint

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── SpeedTest.tsx      # Main speed test component
│   │   └── ui/                # shadcn/ui components
│   ├── layouts/
│   │   └── Layout.astro       # Base layout
│   ├── pages/
│   │   └── index.astro        # Home page
│   ├── lib/
│   │   └── utils.ts           # Utility functions
│   └── styles/
│       └── global.css         # Global styles
├── public/                    # Static assets
├── astro.config.mjs          # Astro configuration
├── tailwind.config.mjs       # Tailwind configuration
├── tsconfig.json             # TypeScript configuration
├── biome.json                # Biome linter configuration
└── package.json

```
