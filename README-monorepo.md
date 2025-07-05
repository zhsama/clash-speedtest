# Clash SpeedTest Monorepo

This is a monorepo containing both the backend (Go) and frontend (Astro) for Clash SpeedTest, managed with Turborepo.

## Project Structure

```
clash-speedtest/
├── backend/                # Go backend with CLI and API server
│   ├── main.go            # CLI tool
│   ├── api-server/        # HTTP API server
│   ├── speedtester/       # Core speed testing logic
│   └── download-server/   # Optional speed test server
├── frontend/              # Astro + React frontend
│   ├── src/
│   │   ├── components/    # React components with shadcn/ui
│   │   ├── pages/         # Astro pages
│   │   └── styles/        # Global styles
│   └── public/            # Static assets
├── turbo.json            # Turborepo configuration
└── package.json          # Root workspace configuration
```

## Prerequisites

- Node.js 18+
- Go 1.21+
- npm 10+

## Quick Start

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Run both frontend and backend in development mode:**
   ```bash
   npm run dev
   ```

   This will start:
   - Frontend at http://localhost:4321
   - Backend API at http://localhost:8090

3. **Run specific apps:**
   ```bash
   # Frontend only
   npm run frontend:dev

   # Backend CLI (interactive mode)
   npm run backend:dev

   # Backend API server only
   npm run api:dev
   ```

## Available Commands

### Root Commands

- `npm run dev` - Start all apps in development mode
- `npm run build` - Build all apps
- `npm run lint` - Lint all apps
- `npm run format` - Format all apps
- `npm run clean` - Clean all build artifacts

### Backend Commands

Run these with `npm run <command> --workspace=backend`:

- `dev` - Run CLI in interactive mode
- `api:dev` - Run API server
- `build` - Build CLI binary
- `build:api` - Build API server binary
- `test` - Run Go tests
- `lint` - Run Go vet
- `format` - Format Go code

### Frontend Commands

Run these with `npm run <command> --workspace=frontend`:

- `dev` - Start Astro dev server
- `build` - Build for production
- `preview` - Preview production build
- `lint` - Run Biome linter
- `format` - Format with Biome

## Features

- **Turborepo**: Efficient monorepo builds with caching
- **Go Backend**: Fast proxy testing with Mihomo/Clash core
- **Astro Frontend**: Modern web app with React components
- **shadcn/ui**: Beautiful UI components
- **TypeScript**: Type-safe frontend development
- **Biome**: Fast linting and formatting for frontend

## Development Workflow

1. Make changes to either backend or frontend
2. Turbo will automatically detect changes and rebuild only affected packages
3. Use `npm run dev` to run everything in watch mode

## Building for Production

```bash
# Build everything
npm run build

# Backend binaries will be in backend/dist/
# Frontend static files will be in frontend/dist/
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Run `npm run lint` and `npm run format`
5. Submit a pull request

## License

[MIT License](LICENSE)