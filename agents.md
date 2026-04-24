# HomePay API — Contexto para agentes

Instrucciones y contexto para Claude Code u otros agentes que trabajen en este repositorio.

## Descripción del proyecto

API REST en Go para HomePay, una app de gestión de finanzas personales. Permite al usuario registrar empresas a las que paga, sus cuentas (servicios), facturas mensuales, gastos variables y planes de cuotas.

## Stack tecnológico

- **Go 1.25** — lenguaje
- **chi v5** — router HTTP (no Gin, no Echo)
- **pgx v5** — driver PostgreSQL con pgxpool (no database/sql)
- **Clerk SDK v2** — autenticación JWT
- **svix v1.89** — verificación de webhooks de Clerk
- **Supabase** — PostgreSQL cloud, schema `homepay`
- **swaggo** — Swagger autogenerado desde anotaciones en handlers
- **Docker** — multi-stage build

## Arquitectura

Tres capas estrictas: `Handler → Service → Repository`

```
cmd/api/main.go       — entry point, instancia y conecta todas las dependencias
internal/
  config/             — godotenv + validación de vars requeridas
  database/           — pgxpool.Connect
  middleware/         — auth.go: valida JWT Clerk, inyecta auth_user_id en el contexto
  models/             — structs de dominio + request/response bodies
  repository/         — SQL con pgx, una interfaz + implementación por entidad
  service/            — lógica de negocio, orquesta repos
  handlers/           — HTTP handlers, usan service
  router/             — monta rutas chi
```

## Convenciones del proyecto

### Base de datos
- Schema siempre prefijado: `homepay.tabla`
- Todos los deletes son **soft delete**: `SET deleted_at = NOW()`
- Nunca usar `DELETE FROM`
- Las queries de lectura siempre filtran `WHERE deleted_at IS NULL`
- IDs de tablas de negocio: `UUID DEFAULT gen_random_uuid()`
- ID de categorías: `SMALLINT GENERATED ALWAYS AS IDENTITY` (no UUID)
- `auth_user_id` es el ID de Clerk (`user_2abc...`), no un UUID propio

### Autenticación
- El middleware de Clerk valida el JWT y pone el `auth_user_id` en el contexto
- En handlers se obtiene con: `authUserID := middleware.GetAuthUserID(r)`
- Todas las rutas excepto `POST /webhooks/clerk` requieren JWT

### Repositorios
- Cada entidad tiene su propia interfaz en `repository/`
- Se usa `pgx.Row` para escanear filas individuales y `pgx.Rows` para colecciones
- Patrón de scan: función `scanX(row pgx.Row, x *models.X) error`
- Constante `xCols` con los nombres de columnas para reutilizar en queries
- `GetByID` siempre retorna `(nil, nil)` cuando no encuentra (no error)

### Servicios
- Cada servicio tiene su interfaz definida en el mismo archivo
- Validan reglas de negocio antes de llamar al repo
- No acceden directamente a la DB

### Handlers
- Decodifican body con `decode(r, &req)`
- Responden con `writeJSON(w, status, data)` o `writeError(w, status, msg)`
- Errores internos: `writeInternalError(w, r, err)` (loggea + responde 500)

### Swagger
- Anotaciones en cada handler con `// @Summary`, `// @Router`, etc.
- Después de cualquier cambio en handlers o modelos, regenerar:
  ```bash
  go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go --output docs
  ```
- No editar `docs/` a mano

### Modelos
- Requests de creación: `CreateXRequest`
- Requests de actualización: `UpdateXRequest` con campos puntero (`*string`, `*int`) para PATCH semántico
- Soft delete nunca expuesto en responses (`json:"deleted_at,omitempty"`)

## Relaciones entre entidades

```
users (Clerk)
  └── categories (por usuario)
  └── companies (por usuario)
        └── accounts
              └── account_billings
  └── variable_expenses (company_id opcional → categoría heredada)
  └── installment_plans
        └── installment_payments
```

## Variables de entorno requeridas

| Variable | Descripción |
|---|---|
| `DATABASE_URL` | Connection string Supabase con `?search_path=homepay` |
| `CLERK_SECRET_KEY` | Clave secreta Clerk para validar JWT |
| `CLERK_WEBHOOK_SECRET` | Secreto de firma de webhooks (`whsec_...`) |
| `PORT` | Puerto HTTP (default `8080`) |

## Comandos frecuentes

```bash
# Desarrollo
go run ./cmd/api/main.go

# Docker local (puerto 8082)
docker compose up --build

# Regenerar Swagger
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go --output docs

# Compilar
go build -o bin/api ./cmd/api
```

## Qué NO hacer

- No usar `DELETE FROM` — siempre soft delete con `deleted_at`
- No usar Gin ni otro router — solo chi v5
- No usar `database/sql` — solo pgx v5
- No agregar migraciones SQL — el schema lo administra el usuario por fuera
- No hardcodear el host en Swagger — el `@host` fue eliminado para que sea dinámico
- No crear helpers innecesarios — si algo se usa una sola vez, va inline
