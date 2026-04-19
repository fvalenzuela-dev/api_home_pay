# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Comandos

```bash
# Desarrollo
go run ./cmd/api/main.go

# Compilar
go build ./...
go build -o bin/api ./cmd/api

# Docker local (puerto 8082)
docker compose up --build

# Regenerar Swagger (obligatorio tras cambiar handlers o modelos)
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go --output docs
```

No hay tests automatizados en este proyecto aún.

## Arquitectura

Tres capas estrictas: `Handler → Service → Repository → Supabase (pgx)`

El entry point (`cmd/api/main.go`) instancia manualmente todos los repos, services y handlers, y los pasa al router. No hay DI framework.

```
Handler    — decodifica request, llama service, escribe response
Service    — valida reglas de negocio, orquesta repos
Repository — SQL directo con pgx v5, una interfaz por entidad
```

## Convenciones críticas

**Base de datos**
- Schema siempre prefijado: `homepay.tabla`
- Todos los deletes son soft delete: `SET deleted_at = NOW()` — nunca `DELETE FROM`
- Queries de lectura siempre filtran `WHERE deleted_at IS NULL`
- IDs de negocio: `UUID`; categorías usan `SMALLINT` (identity)
- `auth_user_id` es el ID de Clerk (`user_2abc...`), no UUID propio

**Repositorios**
- Patrón consistente: constante `xCols` con columnas + función `scanX(pgx.Row, *models.X)`
- `GetByID` retorna `(nil, nil)` cuando no encuentra — nunca retorna error por not found

**Handlers**
- `authUserID := middleware.GetAuthUserID(r)` para obtener el usuario del JWT
- `decode(r, &req)` para deserializar body
- `writeJSON` / `writeError` / `writeInternalError` para responder

**Swagger**
- Anotaciones en cada handler (`@Summary`, `@Router`, etc.)
- Después de cualquier cambio regenerar con el comando de arriba
- No editar `docs/` manualmente
- El `@host` fue eliminado del main para que Swagger use el host del request dinámicamente

## Entidades y relaciones

```
users           — sincronizado desde Clerk vía webhook POST /webhooks/clerk
categories      — por usuario (auth_user_id), soft delete
companies       — por usuario, FK a categories
  accounts      — por empresa, billing_day + auto_accumulate
    billings    — factura mensual en formato YYYYMM (ej: 202603)
variable_expenses — por usuario, company_id opcional (categoría heredada de empresa)
installment_plans — por usuario
  installment_payments — generados todos al crear el plan
```

## Variables de entorno

| Variable | Descripción |
|---|---|
| `DATABASE_URL` | Supabase connection string con `?search_path=homepay` |
| `CLERK_SECRET_KEY` | Clave secreta Clerk para validar JWT |
| `CLERK_WEBHOOK_SECRET` | Secreto de firma de webhooks (`whsec_...`) — pasar con prefijo, el SDK ya lo stripea |
| `PORT` | Puerto HTTP (default `8080`) |

## Restricciones

**Qué NO hacer:**
- NO usar `DELETE FROM` — siempre soft delete con `deleted_at`
- NO usar Gin, Echo u otro router — solo chi v5
- NO usar `database/sql` — solo pgx v5
- NO agregar migraciones SQL — el schema lo administra el usuario por fuera
- NO hardcodear el host en Swagger
- NO crear helpers innecesarios — si algo se usa una sola vez, va inline
- NO usar sesiones HTTP — solo JWT via Clerk
- NO hacer commits sin pasar `go build ./...` antes
