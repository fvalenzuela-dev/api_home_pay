# Agent Instructions: api-home-pay

## Contexto del Proyecto

**api-home-pay** es una API REST en Go para gestión de gastos e ingresos del hogar.

- **Stack:** Go 1.25.8, Gin v1.10.1, PostgreSQL 17 (Supabase), Clerk SDK v2
- **Dominio:** Finanzas personales - gastos, ingresos, períodos, categorías, cuentas de servicio
- **Autenticación:** JWT via Clerk — extracción manual con `jwt.Verify` + Bearer token
- **Logging:** `log/slog` con JSON handler a stdout (capturado por GCP Cloud Logging en producción)
- **Documentación:** Swagger/OpenAPI (swaggo), endpoint `GET /swagger/index.html`
- **Despliegue:** GCP Cloud Run via GitHub Actions

## Entidades del Dominio

| Entidad | Descripción | user_id | Relaciones |
|---------|-------------|---------|------------|
| `periods` | Mes/año de referencia | ❌ compartido | Tiene expenses e incomes |
| `categories` | Categorías de gastos | ❌ compartido | Agrupa expenses |
| `companies` | Compañías proveedoras | ❌ compartido | Tiene service_accounts |
| `service_accounts` | Cuentas de servicio (ej: "Luz - Casa") | ✅ por usuario | Pertenece a company, usada por expenses |
| `expenses` | Gastos/cuentas por pagar | ✅ por usuario | Tiene category, period, service_account |
| `incomes` | Ingresos | ✅ por usuario | Tiene period |

> **Importante:** `categories`, `periods` y `companies` son catálogos compartidos — NO filtrar por `user_id` en sus queries.

## Base de Datos

- **Schema:** `finances` (no `public`) — requiere `SET search_path TO finances` al conectar
- **Driver:** `github.com/lib/pq`
- **Queries:** parametrizadas SIEMPRE (`$1`, `$2`, ...)
- **Conexión:** ver `internal/repository/db.go` — aplica search_path después del ping

## Autenticación (Clerk)

```go
// Middleware extrae Bearer token y valida con jwt.Verify
// user_id queda en el contexto Gin:
userID, ok := c.Get("user_id")

// NUNCA usar:
// clerk.SessionClaimsFromContext() — obsoleto en este proyecto
// clerk.UserIDFromContext()        — obsoleto en este proyecto
```

- Tokens Clerk expiran en 60 segundos por diseño — el frontend maneja el refresh automático
- La seguridad Swagger se llama `BearerAuth` (sin espacios) — debe coincidir en definición y endpoints

## Estructura del Proyecto

```
.
├── cmd/api/           # Entry point (main.go)
├── internal/
│   ├── config/        # Configuración (env vars)
│   ├── handlers/      # HTTP handlers (Gin) + DTOs de request
│   ├── logger/        # Inicialización de slog JSON
│   ├── middleware/    # ClerkAuth, Logging, Response
│   ├── models/        # Structs de entidades + DTOs (dto.go)
│   ├── repository/    # Acceso a DB (interfaces + implementaciones)
│   ├── services/      # Lógica de negocio (interfaces + implementaciones)
│   └── utils/         # Helpers (response helpers)
├── docs/              # Swagger/OpenAPI (generado con swag init)
└── .github/workflows/ # CI/CD (central-validation, docker-gcp-dev/prod, continuous-docs)
```

## Convenciones de Código

### Nombres
- Tipos exportados: `PascalCase` (ej: `ExpenseHandler`)
- Funciones privadas: `camelCase` (ej: `validateExpense`)
- Archivos: `snake_case.go` (ej: `expense_handler.go`)
- Paquetes: lowercase, cortos (ej: `handlers`)

### DTOs de Request
Los handlers usan DTOs separados para input, definidos en `internal/models/dto.go`:

```go
// Usar DTOs con binding:"required" — Gin rechaza automáticamente si faltan campos
type CreateExpenseRequest struct {
    CategoryID       int     `json:"category_id" binding:"required"`
    PeriodID         int     `json:"period_id" binding:"required"`
    Description      string  `json:"description" binding:"required"`
    CurrentAmount    float64 `json:"current_amount" binding:"required"`
}
```

### Errores y Logging

```go
// En repositorios — loguear antes de retornar error:
slog.Error("db error: failed to create expense", "error", err)
return fmt.Errorf("failed to create expense: %w", err)

// En handlers — loguear según severidad:
slog.Error("[ExpenseHandler.Create] error", "error", err)  // 500
slog.Warn("[ExpenseHandler.Create] not found", "id", id)   // 404
```

### Ejemplo de Handler

```go
func (h *ExpenseHandler) Create(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
        return
    }

    var req models.CreateExpenseRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
        return
    }

    expense := &models.Expense{...} // mapear desde req
    if err := h.service.Create(userID.(string), expense); err != nil {
        slog.Error("[ExpenseHandler.Create] error", "error", err)
        utils.ErrorResponse(c, http.StatusInternalServerError, "internal server error")
        return
    }

    utils.SuccessResponse(c, expense)
}
```

## CI/CD

### Workflows Activos

| Workflow | Trigger | Descripción |
|----------|---------|-------------|
| `central-validation.yml` | PR / push | Tests Go + cobertura (lcov via gcov2lcov) |
| `docker-gcp-dev.yml` | Push a develop | Build y deploy a Cloud Run DEV |
| `docker-gcp-prod.yml` | Push a main | Build y deploy a Cloud Run PROD |
| `continuous-docs.yml` | PR con código | Análisis de documentación via Gemini (salta si PR > 50 archivos) |

### Regenerar Swagger

```bash
swag init -g cmd/api/main.go
```

Siempre correr después de cambiar comentarios `// @...` en handlers o main.go.

## Testing

- Tests unitarios: `*_test.go` junto al código
- Handlers: usar `testify/mock` con interfaces de servicio
- Repositorios: usar `go-sqlmock` para simular DB
- **No mockear** validaciones de binding — `c.ShouldBindJSON` rechaza antes de llamar al servicio

## Variables de Entorno

```env
PORT=8080
CLERK_SECRET_KEY=sk_test_...
DATABASE_URL=postgresql://...?search_path=finances
GIN_MODE=release  # o debug en dev
```

## Reglas de Seguridad

1. **NUNCA** hardcodear secretos
2. **SIEMPRE** validar JWT en endpoints protegidos
3. Filtrar por `user_id` solo en entidades que lo tienen (expenses, incomes, service_accounts)
4. **NUNCA** exponer detalles de errores de DB al cliente
5. Usar prepared statements (prevención SQL injection)
6. Validar inputs con `binding:"required"` en DTOs

## Recursos

- [Gin Documentation](https://gin-gonic.com/docs/)
- [Clerk Go SDK v2](https://github.com/clerk/clerk-sdk-go)
- [swaggo/swag](https://github.com/swaggo/swag)
- [lib/pq](https://github.com/lib/pq)
- [testify](https://github.com/stretchr/testify)
- [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)
