# api-home-pay

API REST en Go para gestión de gastos e ingresos del hogar.

**Stack:** Go 1.23 · Gin v1.10.1 · PostgreSQL 17 (Supabase) · Clerk JWT · GCP Cloud Run

---

## Requisitos

- Go 1.23+
- PostgreSQL 17 (o acceso a Supabase)
- Cuenta en [Clerk](https://clerk.com) para autenticación JWT

---

## Configuración

Copia el archivo de ejemplo y completa los valores:

```bash
cp .env.example .env
```

| Variable | Descripción |
|----------|-------------|
| `PORT` | Puerto del servidor (default: `8080`) |
| `CLERK_SECRET_KEY` | Secret key de tu aplicación Clerk |
| `DATABASE_URL` | URL de conexión a PostgreSQL |
| `GIN_MODE` | `debug` en desarrollo, `release` en producción |

> La base de datos debe tener el schema `finances`. La API aplica `SET search_path TO finances` automáticamente al conectar.

---

## Desarrollo

### Instalar dependencias

```bash
go mod download
```

### Ejecutar en modo desarrollo

```bash
go run cmd/api/main.go
```

### Compilar y ejecutar

```bash
go build -o bin/api cmd/api/main.go
./bin/api
```

### Verificar que el servidor está corriendo

```bash
curl http://localhost:8080/health
```

---

## Tests

### Correr todos los tests

```bash
go test ./...
```

### Con cobertura

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Un paquete específico

```bash
go test ./internal/handlers/...
go test ./internal/services/...
go test ./internal/repository/...
```

---

## Documentación (Swagger)

Disponible en: `http://localhost:8080/swagger/index.html`

Para regenerar después de cambiar comentarios `// @...` en los handlers:

```bash
swag init -g cmd/api/main.go
```

---

## Endpoints

Todos los endpoints bajo `/api` requieren autenticación via `Authorization: Bearer <token>`.

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/me` | Usuario autenticado |
| `GET/POST/PUT/DELETE` | `/api/categories` | Categorías de gastos |
| `GET/POST/PUT/DELETE` | `/api/periods` | Períodos (mes/año) |
| `GET/POST/PUT/DELETE` | `/api/companies` | Compañías proveedoras |
| `GET/POST/PUT/DELETE` | `/api/service-accounts` | Cuentas de servicio |
| `GET/POST/PUT/DELETE` | `/api/expenses` | Gastos |
| `PATCH` | `/api/expenses/:id/pay` | Marcar gasto como pagado |
| `GET` | `/api/expenses/pending` | Gastos pendientes |
| `GET/POST/PUT/DELETE` | `/api/incomes` | Ingresos |
| `GET` | `/api/summary/:period_id` | Resumen del período |

---

## Docker

```bash
docker build -t api-home-pay .
docker run -p 8080:8080 --env-file .env api-home-pay
```

---

## CI/CD

| Workflow | Trigger | Acción |
|----------|---------|--------|
| `central-validation` | PR a `develop` | Tests + cobertura (lcov) |
| `docker-gcp-dev` | Manual (`workflow_dispatch`) | Deploy a Cloud Run DEV |
| `docker-gcp-prod` | Push a `main` | Deploy a Cloud Run PROD |
| `continuous-docs` | PR con código | Análisis de docs via Gemini |

---

## Versionado

La versión del proyecto se gestiona en el archivo `VERSION` en la raíz. El pipeline de CI valida que la versión sea actualizada en cada PR.
