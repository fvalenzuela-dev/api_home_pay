This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://github.com/vercel/next.js/tree/canary/packages/create-next-app).

## Getting Started

First, run the development server:

```bash
npm run dev
```

## CI/CD Pipeline

This project uses GitHub Actions to automate validation and deployment processes. The workflows have been organized for clarity:

- **PR Guard**: Monitors pull requests to ensure they follow the established branching strategy.
- **Central Validation**: Triggered on pushes to `main` and `develop`. It performs:
  - **Version Control**: Validates that the `VERSION` file has been correctly updated.
  - **Build & Test**: Ensures the application builds and passes all automated tests.
  - **Codacy Coverage**: Generates and uploads code coverage reports to Codacy.
- **GCP Deploy (DEV/PROD)**: Manual workflows (workflow_dispatch) to build and deploy Docker images to Google Cloud Platform.

## Versioning

The project version is managed in the `VERSION` file at the root. The CI pipeline enforces a check to ensure that the version is updated during development and pull requests.