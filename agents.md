# Agent Instructions: api-home-pay

## Contexto del Proyecto

**api-home-pay** es una API REST en Go para gestión de gastos e ingresos del hogar.

- **Stack:** Go 1.23+, Gin Gonic, PostgreSQL (Supabase), Clerk SDK v2
- **Dominio:** Finanzas personales - gastos, ingresos, períodos, categorías, cuentas de servicio
- **Autenticación:** JWT via Clerk - cada usuario solo ve sus datos
- **Documentación:** Swagger/OpenAPI

## Entidades del Dominio

| Entidad | Descripción | Relaciones |
|---------|-------------|------------|
| `periods` | Mes/año de referencia | Tiene expenses e incomes |
| `categories` | Categorías de gastos | Agrupa expenses |
| `companies` | Compañías proveedoras | Tiene service_accounts |
| `service_accounts` | Cuentas de servicio (ej: "Luz - Casa") | Pertenece a company, usada por expenses |
| `expenses` | Gastos/cuentas por pagar | Tiene category, period, service_account |
| `incomes` | Ingresos | Tiene period |

## Comandos Disponibles

### Desarrollo

```bash
# Iniciar servidor en modo desarrollo
/dev
# o
/run

# Ejecutar tests
/test

# Verificar estilo de código (fmt + vet)
/lint

# Compilar para producción
/build
```

### Estructura del Proyecto

```bash
# Crear estructura base del proyecto
/init-structure

# Agregar nuevo endpoint CRUD completo
/add-crud <entidad>
# Ejemplo: /add-crud expense
```

## Convenciones de Go (OBLIGATORIAS)

### Estructura de Paquetes
```
.
├── cmd/api/           # Entry point
├── internal/
│   ├── config/        # Configuración (env vars)
│   ├── handlers/      # HTTP handlers (Gin)
│   ├── middleware/    # Clerk JWT, logging, etc.
│   ├── models/        # Structs de entidades
│   ├── repository/    # Acceso a DB (Supabase/PostgreSQL)
│   ├── services/      # Lógica de negocio
│   └── utils/         # Helpers
├── docs/              # Swagger/OpenAPI
└── tests/             # Tests de integración
```

### Reglas de Código

1. **Nombres:**
   - Tipos exportados: `PascalCase` (ej: `ExpenseHandler`)
   - Funciones privadas: `camelCase` (ej: `validateExpense`)
   - Archivos: `snake_case.go` (ej: `expense_handler.go`)
   - Paquetes: lowercase, cortos (ej: `handlers`, no `http_handlers`)

2. **Interfaces:**
   - Repositorios DEBEN tener interfaces (para testing)
   - Ejemplo: `type ExpenseRepository interface { ... }`

3. **Errores:**
   - Usar `fmt.Errorf("contexto: %w", err)` para wrap errors
   - NUNCA retornar errores crudos al cliente
   - Respuesta estándar:
     ```go
     {
       "status": "error",
       "message": "descripción",
       "code": 400
     }
     ```

4. **Handlers Gin:**
   - Un archivo por entidad (ej: `expense_handler.go`)
   - Funciones: `func(c *gin.Context)`
   - Extraer user_id: `userID, _ := c.Get("user_id")`

5. **Middleware Clerk:**
   - Usar `clerk.WithHeaderAuthorization()` en rutas protegidas
   - Extraer user_id: `clerk.UserIDFromContext(c.Request.Context())`

6. **Base de Datos:**
   - Usar `database/sql` con PostgreSQL driver
   - Queries parametrizadas SIEMPRE (prepared statements)

### Ejemplo de Handler

```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type ExpenseHandler struct {
    service ExpenseService
}

func NewExpenseHandler(s ExpenseService) *ExpenseHandler {
    return &ExpenseHandler{service: s}
}

func (h *ExpenseHandler) List(c *gin.Context) {
    userID, _ := c.Get("user_id")
    
    expenses, err := h.service.ListByUser(userID.(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status": "error",
            "message": "Error al obtener gastos",
            "code": 500,
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data": expenses,
    })
}
```

## Workflows Comunes

### 1. Agregar un Nuevo Endpoint

```
1. Definir route en internal/routes/
2. Crear handler en internal/handlers/
3. Implementar service en internal/services/
4. Agregar método en repository
5. Agregar tests
6. Actualizar docs/swagger
```

### 2. Agregar una Nueva Entidad

```
1. Crear model en internal/models/
2. Crear repository + interface
3. Crear service + interface
4. Crear handler
5. Crear routes
6. Agregar tests
```

### 3. Flujo de Autenticación

```
1. Cliente envía Bearer token de Clerk
2. Middleware valida JWT con Clerk SDK
3. Extrae user_id del contexto
4. Todos los queries filtran por user_id
5. Nunca confiar en user_id del body
```

## Variables de Entorno Requeridas

```env
PORT=8080
CLERK_SECRET_KEY=sk_test_...
DATABASE_URL=postgresql://...
GIN_MODE=release  # o debug en dev
```

## Testing

- Tests unitarios: `*_test.go` junto al código
- Tests de integración: carpeta `/tests/`
- Usar testify/assert para assertions
- Mockear repositorios con interfaces

## Documentación API

- Swagger docs en `/docs/swagger.yaml`
- Endpoint: `GET /swagger/index.html`
- Usar swag CLI para generar docs desde comentarios

## Reglas de Seguridad Críticas

1. **NUNCA** hardcodear secretos
2. **SIEMPRE** validar JWT en endpoints protegidos
3. **SIEMPRE** filtrar queries por user_id
4. **NUNCA** exponer detalles de errores de DB al cliente
5. Usar prepared statements (SQL injection prevention)
6. Validar todos los inputs (usar go-playground/validator)

## Recursos Útiles

- [Go by Example](https://gobyexample.com/)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [Clerk Go SDK](https://github.com/clerk/clerk-sdk-go)
- [PostgreSQL Driver](https://github.com/lib/pq)
- [Testify](https://github.com/stretchr/testify)

## Notas de Desarrollo

- Go no tiene clases, usa structs + métodos
- Composición sobre herencia
- Errores explícitos, no excepciones
- Goroutines para concurrencia (cuando sea necesario)
- Defer para cleanup de recursos
