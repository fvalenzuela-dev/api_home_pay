# Project README

This is a Next.js project.

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

## CI/CD Pipeline

This project uses GitHub Actions to automate validation, documentation, and deployment processes.

### Validation & Quality
- **Central Validation**: Triggered on pushes to feature branches (including `qa`) and pull requests targeting the `develop` branch. It validates the `VERSION` file, ensures the application builds, and passes all automated tests.
- **Continuous Documentation**: Monitors pull requests (except those targeting `main`) to ensure that documentation stays in sync with code changes by analyzing changes in `.js`, `.md`, and `.yml` files.

### Deployment
- **GCP Deploy (DEV)**: Automatically triggered on pushes to the `develop` branch. Deploys the Docker image to the development environment on Google Cloud Platform.
- **GCP Deploy (PROD)**: Automatically triggered on pushes to the `main` branch. Deploys the Docker image to the production environment on Google Cloud Platform.

## Versioning

The project version is managed in the `VERSION` file at the root. The CI pipeline enforces a check to ensure that the version is updated during development.