This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://github.com/vercel/next.js/tree/canary/packages/create-next-app).

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

## Deployment

This project uses GitHub Actions to automate validation and deployment processes. Deployments to the development environment are automatically triggered on pushes to the `develop` branch, while pushes to the `main` branch trigger deployments to the production environment.

## Versioning

The project version is managed in the `VERSION` file at the root. The CI pipeline enforces a check to ensure that the version is updated during development and pull requests.
