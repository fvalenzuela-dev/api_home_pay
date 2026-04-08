# HomePay — PRD Etapa 2: Backend

## Contexto
Este documento describe el API backend de HomePay. Debe leerse junto al PRD Etapa 1 (base de datos) ya que asume que el schema `homepay` en Supabase está creado y funcionando.

---

## Stack
- **Lenguaje:** Go
- **Router:** chi (`github.com/go-chi/chi/v5`)
- **Driver Postgres:** pgx v5 (`github.com/jackc/pgx/v5`) con pool de conexiones (`pgxpool`)
- **Autenticación:** Clerk (`github.com/clerk/clerk-sdk-go/v2`)
- **Verificación de webhooks:** svix (`github.com/svix/svix-webhooks/go`)
- **Variables de entorno:** godotenv (`github.com/joho/godotenv`)
- **Base de datos:** Supabase (Postgres cloud)

---

## Variables de entorno requeridas

```
DATABASE_URL        — connection string de Supabase con search_path=homepay
CLERK_SECRET_KEY    — clave secreta de Clerk para validar JWT
CLERK_WEBHOOK_SECRET — secreto para verificar firma de webhooks de Clerk
PORT                — default 8080
```

---

## Estructura del proyecto

```
homepay-api/
├── cmd/api/main.go
├── internal/
│   ├── config/config.go
│   ├── database/database.go
│   ├── middleware/auth.go
│   ├── models/
│   │   ├── user.go
│   │   ├── company.go
│   │   ├── account.go
│   │   ├── billing.go
│   │   ├── expense.go
│   │   └── installment.go
│   ├── handlers/
│   │   ├── webhook.go
│   │   ├── companies.go
│   │   ├── accounts.go
│   │   ├── billings.go
│   │   ├── expenses.go
│   │   └── installments.go
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── company_repo.go
│   │   ├── account_repo.go
│   │   ├── billing_repo.go
│   │   ├── expense_repo.go
│   │   └── installment_repo.go
│   └── router/router.go
├── .env
├── go.mod
└── go.sum
```

---

## Capas de la aplicación

### `config`
Carga las variables de entorno al iniciar. Si falta alguna variable requerida, el servidor no arranca y loguea el error.

### `database`
Abre un pool de conexiones a Postgres con `pgxpool`. Verifica la conexión con ping al arrancar.

### `middleware/auth`
Intercepta todos los requests en rutas protegidas. Extrae el JWT del header `Authorization: Bearer <token>`, lo valida con Clerk y deja el `auth_user_id` disponible en el contexto del request. Si el token es inválido o no existe, responde 401.

### `models`
Structs de Go que representan las entidades del dominio. Incluyen tags JSON para serialización y mapeo con las columnas de Postgres.

### `repository`
Única capa que habla con Postgres. Recibe contexto y parámetros, ejecuta queries SQL, retorna structs o errores. No sabe nada de HTTP.

### `handlers`
Recibe el request HTTP, valida el input, llama al repository y escribe la respuesta JSON. No contiene lógica de negocio compleja.

### `router`
Monta todas las rutas. Separa las rutas públicas (webhook) de las protegidas (requieren JWT).

---

## Autenticación

Clerk maneja el ciclo completo de auth. El backend solo valida el JWT en cada request:

- El frontend obtiene el JWT de Clerk después del login
- Lo envía en cada request como `Authorization: Bearer <token>`
- El middleware lo valida usando el SDK de Clerk
- El `auth_user_id` (claims.Subject) queda disponible en el contexto
- Todos los queries a Postgres filtran por ese `auth_user_id` — un usuario nunca puede ver datos de otro

---

## Webhook de Clerk

**Ruta:** `POST /webhooks/clerk` (pública, sin JWT)

**Flujo:**
1. Leer el body raw del request
2. Verificar la firma con svix usando `CLERK_WEBHOOK_SECRET` — si falla, responder 401
3. Parsear el evento JSON
4. Actuar según `event.type`:
   - `user.created` → upsert en `homepay.users`
   - `user.updated` → upsert en `homepay.users`
   - `user.deleted` → soft delete en `homepay.users`
5. Responder 200

**Upsert de usuario:** `INSERT INTO homepay.users ... ON CONFLICT (auth_user_id) DO UPDATE SET ...`. Si el usuario ya existe, actualiza `email`, `full_name` y `updated_at`. Si no existe, lo crea. Solo actúa sobre registros con `deleted_at IS NULL`.

**Campos que llegan de Clerk:**
- `data.id` → `auth_user_id`
- `data.email_addresses[0].email_address` → `email`
- `data.first_name` + `data.last_name` → `full_name`

---

## Endpoints

Todas las rutas siguientes requieren JWT válido en el header. Todos los queries filtran automáticamente por `auth_user_id` y `deleted_at IS NULL`.

### Empresas — `/companies`

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/companies` | Lista todas las empresas activas del usuario |
| POST | `/companies` | Crea una empresa |
| PUT | `/companies/{id}` | Edita nombre o categoría |
| DELETE | `/companies/{id}` | Soft delete — marca `deleted_at`. Propaga a sus accounts y billings |

**Regla de negocio DELETE:** al hacer soft delete de una empresa, el backend debe marcar `deleted_at = NOW()` también en todas sus `accounts` activas y en todas las `account_billings` activas de esas cuentas.

---

### Cuentas — `/companies/{companyID}/accounts`

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/companies/{companyID}/accounts` | Lista cuentas activas de esa empresa |
| POST | `/companies/{companyID}/accounts` | Crea una cuenta |
| PUT | `/companies/{companyID}/accounts/{id}` | Edita nombre, billing_day o auto_accumulate |
| DELETE | `/companies/{companyID}/accounts/{id}` | Soft delete — propaga a sus billings activas |

---

### Facturas — `/accounts/{accountID}/billings`

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/accounts/{accountID}/billings` | Lista facturas de esa cuenta |
| POST | `/accounts/{accountID}/billings` | Registra la factura del mes |
| PUT | `/accounts/{accountID}/billings/{id}` | Actualiza monto pagado o marca como pagada |

**Regla de negocio — acumulación:** al iniciar un nuevo mes, si una `account_billing` no está pagada (`is_paid = FALSE`) y su cuenta tiene `auto_accumulate = TRUE`, el backend debe crear un nuevo registro para el mes siguiente con `carried_from = id_de_la_factura_impaga`. El registro original no se modifica.

**Regla de negocio — marcar como pagada:** cuando `amount_paid >= amount_billed`, actualizar `is_paid = TRUE` y `paid_at = fecha actual`.

---

### Gastos variables — `/expenses`

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/expenses` | Lista gastos del usuario. Soporta filtro por `?month=&year=` y `?category=` |
| POST | `/expenses` | Registra un gasto |
| PUT | `/expenses/{id}` | Edita descripción, monto, categoría o fecha |
| DELETE | `/expenses/{id}` | Soft delete |

---

### Cuotas — `/installments`

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/installments` | Lista planes activos del usuario |
| POST | `/installments` | Crea un plan y genera todos sus pagos individuales |
| PUT | `/installments/{id}/payments/{paymentID}` | Marca una cuota como pagada |
| DELETE | `/installments/{id}` | Soft delete del plan y sus pagos |

**Regla de negocio — crear plan:** al crear un `installment_plan`, el backend genera automáticamente todos los registros en `installment_payments` (uno por cuota), calculando `due_date` para cada uno a partir de `start_date`.

**Regla de negocio — pagar cuota:** al marcar una cuota como pagada, incrementar `installment_plans.installments_paid`. Si `installments_paid = total_installments`, marcar `is_completed = TRUE`.

---

### Dashboard — `/dashboard`

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/dashboard?month=&year=` | Retorna resumen financiero del mes solicitado |

**Datos que retorna:**
- Total de facturas del mes (suma de `account_billings.amount_billed`)
- Total pagado vs pendiente
- Total de gastos variables del mes agrupados por categoría
- Total de cuotas activas del mes
- Lista de compromisos pendientes del mes (facturas no pagadas + cuotas del mes no pagadas)

---

## Convenciones de respuesta

**Éxito:** HTTP 200/201 con JSON `{ "data": ... }`

**Error de validación:** HTTP 400 con JSON `{ "error": "descripción del error" }`

**No autorizado:** HTTP 401 con JSON `{ "error": "no autorizado" }`

**No encontrado:** HTTP 404 con JSON `{ "error": "no encontrado" }`

**Error interno:** HTTP 500 con JSON `{ "error": "error interno" }`

---

## Consideraciones de seguridad

- Todos los queries incluyen el `auth_user_id` del JWT en el WHERE — nunca se confía en un ID de usuario que venga en el body o en la URL
- El endpoint `/webhooks/clerk` no lleva middleware de JWT — usa verificación de firma propia de Clerk
- Las variables de entorno nunca se loguean ni se exponen en respuestas de error
