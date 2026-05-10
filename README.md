# HomePay API

Backend REST para HomePay. Gestiona categorías, empresas, cuentas, facturas, gastos variables y planes de cuotas por usuario.

## Stack

- **Go 1.25**
- **chi v5** — router HTTP
- **pgx v5** — driver PostgreSQL con pool de conexiones
- **Clerk SDK v2** — autenticación JWT
- **svix v1.89** — verificación de webhooks
- **Supabase** — base de datos PostgreSQL cloud (schema `homepay`)
- **swaggo** — documentación Swagger autogenerada
- **Docker** — contenedor de producción

## Requisitos

- Go 1.26.1+
- Docker y Docker Compose
- Acceso a una instancia de Supabase con el schema `homepay` creado

## Configuración

Crea un archivo `.env` en la raíz con las siguientes variables:

| Variable | Descripción |
|---|---|
| `DATABASE_URL` | Connection string de Supabase con `search_path=homepay` |
| `CLERK_SECRET_KEY` | Clave secreta de Clerk para validar JWT |
| `CLERK_WEBHOOK_SECRET` | Secreto de firma de webhooks de Clerk (`whsec_...`) |
| `PORT` | Puerto del servidor (default: `8080`) |

## Ejecutar en desarrollo

```bash
go run ./cmd/api/main.go
```

El servidor arranca en `http://localhost:8080`.

## Ejecutar en Docker

```bash
docker compose up --build
```

La app queda expuesta en `http://localhost:8082`.

## Compilar binario

```bash
go build -o bin/api ./cmd/api
```

## Endpoints

Todas las rutas excepto `/webhooks/clerk` requieren `Authorization: Bearer <token>` de Clerk.

### Webhooks
| Método | Ruta | Descripción |
|---|---|---|
| POST | `/webhooks/clerk` | Sincroniza usuarios desde Clerk (sin JWT) |

### Categorías
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/categories` | Lista categorías del usuario |
| GET | `/categories/{id}` | Obtiene categoría por ID |
| POST | `/categories` | Crea categoría |
| PUT | `/categories/{id}` | Edita categoría |
| DELETE | `/categories/{id}` | Soft delete |

### Empresas
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/companies` | Lista empresas del usuario |
| GET | `/companies/{id}` | Obtiene empresa por ID |
| POST | `/companies` | Crea empresa |
| PUT | `/companies/{id}` | Edita empresa |
| DELETE | `/companies/{id}` | Soft delete (propaga a cuentas y facturas) |

### Cuentas
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/companies/{companyID}/accounts` | Lista cuentas de una empresa |
| GET | `/companies/{companyID}/accounts/{id}` | Obtiene cuenta por ID |
| POST | `/companies/{companyID}/accounts` | Crea cuenta |
| PUT | `/companies/{companyID}/accounts/{id}` | Edita cuenta |
| DELETE | `/companies/{companyID}/accounts/{id}` | Soft delete (propaga a facturas) |

### Facturas
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/accounts/{accountID}/billings` | Lista facturas de una cuenta |
| GET | `/accounts/{accountID}/billings/{id}` | Obtiene factura por ID |
| POST | `/accounts/{accountID}/billings` | Registra factura del mes |
| PUT | `/accounts/{accountID}/billings/{id}` | Actualiza monto pagado |

### Gastos variables
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/expenses` | Lista gastos (filtros: `?month=&year=&company_id=`) |
| GET | `/expenses/{id}` | Obtiene gasto por ID |
| POST | `/expenses` | Registra gasto |
| PUT | `/expenses/{id}` | Edita gasto |
| DELETE | `/expenses/{id}` | Soft delete |

### Planes de cuotas
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/installments` | Lista planes con sus cuotas |
| GET | `/installments/{id}` | Obtiene plan con sus cuotas por ID |
| POST | `/installments` | Crea plan y genera todos los pagos |
| PUT | `/installments/{id}/payments/{paymentID}` | Marca cuota como pagada |
| DELETE | `/installments/{id}` | Soft delete del plan |

### Dashboard
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/dashboard` | Resumen financiero mensual (`?month=&year=`) |

## Documentación Swagger

Disponible en `/docs/` cuando el servidor está corriendo.

Para regenerar tras modificar handlers o modelos:

```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go --output docs
```

> `docs/docs.go`, `docs/swagger.json` y `docs/swagger.yaml` son generados — no editar a mano.

## Estructura del proyecto

```
cmd/api/main.go          — entry point, wiring de dependencias
internal/
  config/                — carga de variables de entorno (.env)
  database/              — conexión pgxpool a Supabase
  middleware/            — validación JWT de Clerk
  models/                — structs del dominio y request/response
  repository/            — queries SQL directas con pgx
  service/               — lógica de negocio
  handlers/              — handlers HTTP
  router/                — rutas chi
docs/                    — Swagger autogenerado
Dockerfile               — imagen multi-stage para producción
docker-compose.yml       — orquestación local
```

## Arquitectura

Flujo de una request: `Handler → Service → Repository → Supabase`

- **Handler**: decodifica request, llama al service, escribe response
- **Service**: valida reglas de negocio, orquesta repositorios
- **Repository**: ejecuta SQL, mapea filas a structs

Todos los deletes son **soft delete** (`deleted_at = NOW()`). Las categorías son por usuario. La categoría de un gasto variable se hereda de la empresa asociada.

## Deployment a GCP Cloud Run

### Requisitos

- Google Cloud SDK (`gcloud`)
- Workflows de GitHub configurados con Workload Identity Federation
- Secrets configurados en GitHub:
  - `GCP_PROJECT_ID_DEV` / `GCP_PROJECT_ID_PROD`
  - `GCP_WORKLOAD_IDENTITY_PROVIDER_DEV` / `GCP_WORKLOAD_IDENTITY_PROVIDER_PROD`
  - `GCP_SERVICE_ACCOUNT_DEV` / `GCP_SERVICE_ACCOUNT_PROD`

### Variables de entorno en Cloud Run

| Variable | Descripción |
|---|---|
| `DATABASE_URL` | Connection string de Supabase con `search_path=homepay` |
| `CLERK_SECRET_KEY` | Clave secreta de Clerk para validar JWT |
| `CLERK_WEBHOOK_SECRET` | Secreto de firma de webhooks de Clerk (`whsec_...`) |
| `PORT` | Puerto del servidor (default: `8080`) |

### Health check

El endpoint `GET /health/ready` se usa como readiness probe de Cloud Run. Retorna `200` si la base de datos está accesible, `503` si no.

### Workflows

- `develop` → `docker-gcp-dev.yml` → `api-home-pay-go-dev`
- `main` → `docker-gcp-prod.yml` → `api-home-pay-go-prod`
