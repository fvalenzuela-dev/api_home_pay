# PRD — Apertura de Periodo

## Contexto

HomePay maneja cuentas recurrentes (servicios, tarjetas, suscripciones) agrupadas por empresa. Cada mes, el usuario abre un **periodo** (formato `YYYYMM`) para generar los billings del mes. La apertura es manual, disparada desde el frontend.

## Objetivo

Exponer un endpoint que, dado un periodo como `202605`, genere automáticamente un `account_billing` por cada cuenta activa del usuario, aplicando la lógica de carry-over para deudas del periodo anterior.

---

## Endpoints

### 1. Abrir periodo

```
POST /periods/{period}/open
Authorization: Bearer <clerk-jwt>
```

**Path param**: `period` — entero YYYYMM (ej: `202605`)

**Response 200**:
```json
{
  "period": 202605,
  "created": 12,
  "skipped": 2
}
```

- `created`: billings nuevos generados
- `skipped`: cuentas que ya tenían billing para ese periodo (idempotencia)

**Errores**:
- `400` — periodo inválido (formato, mes fuera de rango)
- `401` — sin autenticación

---

### 2. Consultar billings de un periodo

```
GET /periods/{period}/billings?status=all|paid|unpaid&page=1&page_size=20
Authorization: Bearer <clerk-jwt>
```

**Path param**: `period` — entero YYYYMM

**Query params**:

| Param | Valores | Default | Descripción |
|-------|---------|---------|-------------|
| `status` | `all`, `paid`, `unpaid` | `all` | Filtra por estado de pago |
| `page` | entero ≥ 1 | `1` | Paginación |
| `page_size` | entero 1–100 | `20` | Tamaño de página |

**Response 200**:
```json
{
  "data": [
    {
      "id": "uuid",
      "account_id": "uuid",
      "period": 202605,
      "amount_billed": 15000.00,
      "amount_paid": 0.00,
      "is_paid": false,
      "paid_at": null,
      "carried_from": "uuid-del-billing-anterior",
      "created_at": "2026-05-01T00:00:00Z"
    }
  ],
  "total": 14,
  "page": 1,
  "page_size": 20
}
```

**Errores**:
- `400` — periodo o status inválido
- `401` — sin autenticación

**Notas**:
- Solo retorna billings del usuario autenticado (join con companies por `auth_user_id`)
- El campo `carried_from` permite al frontend resolver la cadena de deuda
- `GetAllByPeriod` ya existe en el repo — solo necesita el parámetro `is_paid *bool` adicional

---

## Lógica de negocio

### 1. Obtener cuentas activas del usuario

Todas las `accounts` cuyas `companies` tengan `auth_user_id = <usuario>` y `deleted_at IS NULL` en ambas tablas.

### 2. Por cada cuenta — decisión de billing

Para cada cuenta:

**A) Ya existe billing para el periodo solicitado** → saltar (sumar a `skipped`).

**B) No existe billing** → crear uno nuevo:

1. Calcular el **periodo anterior** (`P-1`):
   - `202605` → `202604`
   - `202601` → `202512` ← requiere lógica de rollover de mes/año

2. Buscar el billing del periodo anterior para esta cuenta:
   - Si **existe y no está pagado** (`is_paid = false`):
     - `amount_billed = billing_anterior.amount_billed` (se arrastra el monto completo)
     - `carried_from = billing_anterior.id`
     - El encadenamiento hacia atrás se preserva a través de `carried_from` del billing anterior (linked list implícita)
   - Si **no existe** o **está pagado**:
     - `amount_billed = 0`
     - `carried_from = null`

3. El billing nuevo siempre arranca con `is_paid = false`, `amount_paid = 0`.

4. El frontend actualiza el `amount_billed` real después (puede quedar en 0 hasta que el usuario lo cargue).

### 3. Responder con el resumen

```
{ period, created, skipped }
```

---

## Carry-over — detalle

El campo `carried_from` apunta al billing impago del periodo inmediatamente anterior. Encadenando ese campo hacia atrás se puede reconstruir toda la cadena de deuda:

```
202603 billing (impago) ← carried_from
  202604 billing (impago) ← carried_from
    202605 billing (nuevo, amount = suma acumulada de 202604)
```

**Importante**: el `amount_billed` del nuevo billing es el `amount_billed` del periodo anterior (no la suma de toda la cadena). La cadena completa se resuelve en el frontend siguiendo los `carried_from`.

---

## Cálculo del periodo anterior

```go
func previousPeriod(period int) int {
    year := period / 100
    month := period % 100
    month--
    if month == 0 {
        month = 12
        year--
    }
    return year*100 + month
}
```

---

## Validación del periodo

- Debe ser un entero de 6 dígitos
- Mes entre 01 y 12
- Año razonable (ej: 2020–2100)

---

## Idempotencia

Si se llama dos veces con el mismo periodo, la segunda llamada retorna `created: 0, skipped: N`. No se sobreescriben billings existentes.

---

## Capas afectadas

| Capa | Cambio |
|------|--------|
| `models` | Nuevo tipo `OpenPeriodResponse` |
| `repository/billing_repo` | `GetByAccountAndPeriod(ctx, accountID, period)` — busca billing existente para esa cuenta+periodo |
| `repository/account_repo` | `GetAllByUser(ctx, authUserID)` — todas las cuentas activas del usuario (join con companies) |
| `service/billing_service` | `OpenPeriod(ctx, authUserID, period)` — orquesta toda la lógica |
| `handlers/billing_handler` | `POST /periods/{period}/open` y `GET /periods/{period}/billings` |
| `router` | Registrar la nueva ruta |
| `docs` | Regenerar Swagger |

---

## Métodos de repo existentes reutilizables

- `CreateCarryOver(ctx, accountID, period, amount, carriedFrom)` — ya existe, crear billing con carry
- `Create(ctx, accountID, req)` — para billings sin carry
- `GetUnpaidByAccount(ctx, accountID)` — retorna el impago más reciente (puede servir pero retorna cualquier periodo, no el anterior específico)
- `GetAllByPeriod(ctx, authUserID, period, pagination)` — ya existe, necesita extenderse con filtro `is_paid`

**Nuevos métodos necesarios en BillingRepository**:
```go
GetByAccountAndPeriod(ctx context.Context, accountID string, period int) (*models.AccountBilling, error)
```

**Método a extender en BillingRepository**:
```go
// Agregar parámetro isPaid *bool — nil = todos, true = pagados, false = impagos
GetAllByPeriod(ctx context.Context, authUserID string, period int, isPaid *bool, p models.PaginationParams) ([]models.AccountBilling, int, error)
```

**Nuevo método necesario en AccountRepository**:
```go
GetAllByUser(ctx context.Context, authUserID string) ([]models.Account, error)
```

---

## Notas de implementación

- La operación puede generar N inserts (uno por cuenta). Hacerlos en una transacción para que sea atómica.
- Si un insert individual falla, hacer rollback total y retornar 500.
- No hay background jobs — es síncrono, el usuario espera la respuesta.
