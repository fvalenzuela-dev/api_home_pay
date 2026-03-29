# HomePay API

Backend REST para HomePay. Gestiona empresas, cuentas, facturas, gastos variables y planes de cuotas por usuario.

## Stack

- **Go 1.23**
- **chi v5** — router HTTP
- **pgx v5** — driver PostgreSQL con pool de conexiones
- **Clerk SDK v2** — autenticación JWT
- **svix** — verificación de webhooks
- **Supabase** — base de datos PostgreSQL cloud (schema `homepay`)

## Requisitos

- Go 1.23+
- Acceso a una instancia de Supabase con el schema `homepay` creado

## Configuración

Copia el archivo de ejemplo y completa las variables:

```bash
cp .env.example .env
```

| Variable | Descripción |
|---|---|
| `DATABASE_URL` | Connection string de Supabase con `search_path=homepay` |
| `CLERK_SECRET_KEY` | Clave secreta de Clerk para validar JWT |
| `CLERK_WEBHOOK_SECRET` | Secreto para verificar firma de webhooks de Clerk |
| `PORT` | Puerto del servidor (default: `8080`) |

## Compilar

```bash
go build ./...
```

Genera el binario en la raíz:

```bash
go build -o bin/api ./cmd/api
```

## Ejecutar

```bash
# Modo desarrollo (recarga manual)
go run ./cmd/api/main.go

# Binario compilado
./bin/api
```

El servidor arranca en `http://localhost:8080` (o el puerto configurado en `PORT`).

## Endpoints

Todas las rutas excepto `/webhooks/clerk` requieren `Authorization: Bearer <token>` de Clerk.

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/webhooks/clerk` | Webhook de Clerk (sin JWT) |
| GET | `/companies` | Lista empresas del usuario |
| POST | `/companies` | Crea empresa |
| PUT | `/companies/{id}` | Edita empresa |
| DELETE | `/companies/{id}` | Soft delete (propaga a cuentas y facturas) |
| GET | `/companies/{companyID}/accounts` | Lista cuentas de una empresa |
| POST | `/companies/{companyID}/accounts` | Crea cuenta |
| PUT | `/companies/{companyID}/accounts/{id}` | Edita cuenta |
| DELETE | `/companies/{companyID}/accounts/{id}` | Soft delete (propaga a facturas) |
| GET | `/accounts/{accountID}/billings` | Lista facturas de una cuenta |
| POST | `/accounts/{accountID}/billings` | Registra factura del mes |
| PUT | `/accounts/{accountID}/billings/{id}` | Actualiza monto pagado |
| GET | `/expenses` | Lista gastos (filtros: `?month=&year=&category=`) |
| POST | `/expenses` | Registra gasto |
| PUT | `/expenses/{id}` | Edita gasto |
| DELETE | `/expenses/{id}` | Soft delete |
| GET | `/installments` | Lista planes de cuotas |
| POST | `/installments` | Crea plan y genera todos los pagos |
| PUT | `/installments/{id}/payments/{paymentID}` | Marca cuota como pagada |
| DELETE | `/installments/{id}` | Soft delete del plan |
| GET | `/dashboard` | Resumen financiero mensual (`?month=&year=`) |

## Estructura del proyecto

```
cmd/api/main.go          — entry point
internal/
  config/                — carga de variables de entorno
  database/              — conexión pgxpool
  middleware/            — auth JWT de Clerk
  models/                — structs del dominio
  repository/            — queries SQL (pgx)
  service/               — lógica de negocio
  handlers/              — handlers HTTP
  router/                — rutas chi
```

## Documentación (Swagger)

La UI interactiva está disponible en `http://localhost:8080/docs/index.html` cuando el servidor está corriendo.

Para regenerar los docs después de modificar handlers o modelos:

```bash
~/go/bin/swag init -g cmd/api/main.go -o docs/
```

> Los archivos `docs/docs.go`, `docs/swagger.json` y `docs/swagger.yaml` son generados automáticamente — no editar a mano.
