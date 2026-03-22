# Product Requirements Document (PRD): api-home-pay

## 1. Visión General
**api-home-pay** es una API REST robusta desarrollada en **Go (Golang)** diseñada para la gestión integral de gastos y pagos del hogar. El sistema permite a los usuarios registrar cuentas, categorizarlas, dar seguimiento a fechas de vencimiento y marcar pagos realizados, garantizando la seguridad de los datos mediante autenticación moderna.

---

## 2. Objetivos del Proyecto
* **Centralización:** Un solo punto para ver deudas de luz, agua, internet, alquiler, etc.
* **Control Financiero:** Permitir al usuario saber cuánto debe y cuánto ha pagado en el mes.
* **Seguridad:** Implementar un sistema de acceso privado donde cada usuario solo accede a su propia información.
* **Escalabilidad:** Código limpio en Go que permita crecer a futuro (ej. recordatorios por email).

---

## 3. Especificaciones Técnicas (Stack)
* **Lenguaje:** Go 1.23+
* **Framework:** Gin Gonic v1.10.1 (Alta velocidad y facilidad de uso).
* **Base de Datos:** PostgreSQL 17 via Supabase (Relacional, ideal para transacciones financieras).
* **Autenticación:** Clerk SDK Go v2 (JSON Web Tokens + User Management completo).
* **Documentación:** Swagger / OpenAPI.

### Detalles de Autenticación con Clerk
* **SDK:** `github.com/clerk/clerk-sdk-go/v2`
* **Configuración:** Requiere `CLERK_SECRET_KEY` desde el Dashboard de Clerk.
* **Features:**
  - Gestión completa de usuarios (crear, obtener, listar)
  - Soporte para Organizaciones y Memberships
  - Validación de tokens JWT integrada
  - Manejo de errores con `clerk.APIErrorResponse`
* **Endpoints protegidos:** Middleware de Clerk valida JWT en todas las rutas `/api/*`.

---

## 4. Arquitectura de Datos (Entidades)

BEGIN;


CREATE TABLE IF NOT EXISTS finances.categories
(
    id serial NOT NULL,
    name character varying(50) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT categories_pkey PRIMARY KEY (id),
    CONSTRAINT categories_name_key UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS finances.companies
(
    id serial NOT NULL,
    name character varying(100) COLLATE pg_catalog."default" NOT NULL,
    website_url character varying(255) COLLATE pg_catalog."default",
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT companies_pkey PRIMARY KEY (id),
    CONSTRAINT companies_name_key UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS finances.expenses
(
    id serial NOT NULL,
    category_id integer NOT NULL,
    period_id integer NOT NULL,
    account_id integer,
    description character varying(255) COLLATE pg_catalog."default" NOT NULL,
    due_date date,
    current_amount numeric(12, 2) NOT NULL DEFAULT 0.00,
    amount_paid numeric(12, 2) NOT NULL DEFAULT 0.00,
    current_installment integer DEFAULT 1,
    total_installments integer DEFAULT 1,
    installment_group_id uuid,
    is_recurring boolean DEFAULT false,
    notes text COLLATE pg_catalog."default",
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT expenses_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS finances.incomes
(
    id serial NOT NULL,
    period_id integer NOT NULL,
    description character varying(255) COLLATE pg_catalog."default" NOT NULL,
    amount numeric(12, 2) NOT NULL DEFAULT 0.00,
    is_recurring boolean DEFAULT true,
    received_at date DEFAULT CURRENT_DATE,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT incomes_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS finances.periods
(
    id serial NOT NULL,
    month_number integer,
    year_number integer NOT NULL,
    CONSTRAINT periods_pkey PRIMARY KEY (id),
    CONSTRAINT periods_month_number_year_number_key UNIQUE (month_number, year_number)
);

CREATE TABLE IF NOT EXISTS finances.service_accounts
(
    id serial NOT NULL,
    company_id integer,
    account_identifier character varying(100) COLLATE pg_catalog."default" NOT NULL,
    alias character varying(100) COLLATE pg_catalog."default",
    CONSTRAINT service_accounts_pkey PRIMARY KEY (id),
    CONSTRAINT service_accounts_company_id_account_identifier_key UNIQUE (company_id, account_identifier)
);

ALTER TABLE IF EXISTS finances.expenses
    ADD CONSTRAINT expenses_account_id_fkey FOREIGN KEY (account_id)
    REFERENCES finances.service_accounts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE SET NULL;


ALTER TABLE IF EXISTS finances.expenses
    ADD CONSTRAINT expenses_category_id_fkey FOREIGN KEY (category_id)
    REFERENCES finances.categories (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS finances.expenses
    ADD CONSTRAINT expenses_period_id_fkey FOREIGN KEY (period_id)
    REFERENCES finances.periods (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS finances.incomes
    ADD CONSTRAINT incomes_period_id_fkey FOREIGN KEY (period_id)
    REFERENCES finances.periods (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION;


ALTER TABLE IF EXISTS finances.service_accounts
    ADD CONSTRAINT service_accounts_company_id_fkey FOREIGN KEY (company_id)
    REFERENCES finances.companies (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE;

END;

---

## 5. Requisitos Funcionales (Endpoints)

### A. Autenticación y Seguridad
* Clerk proporciona autenticación completa via SDK Go v2.
* **Middleware:** `clerk.WithHeaderAuthorization()` valida tokens JWT en todas las rutas `/api/*`.
* **Obtener usuario actual:** `clerk.UserIDFromContext(ctx)` extrae el `user_id` del token.
* **Flujo:** Frontend maneja login/registro via Clerk SDK → Backend valida JWT en cada request.

### B. CRUD de Períodos (Periods)
* `GET /api/periods`: Listar períodos disponibles.
* `POST /api/periods`: Crear nuevo período (mes/año).
* `GET /api/periods/:id`: Detalle de un período específico.
* `PUT /api/periods/:id`: Actualizar período.
* `DELETE /api/periods/:id`: Eliminar período (solo si no tiene gastos ni ingresos asociados).

### C. CRUD de Categorías (Categories)
* `GET /api/categories`: Listar todas las categorías.
* `POST /api/categories`: Crear nueva categoría.
* `GET /api/categories/:id`: Detalle de una categoría específica.
* `PUT /api/categories/:id`: Editar nombre de categoría.
* `DELETE /api/categories/:id`: Eliminar categoría (solo si no tiene gastos asociados).

### D. CRUD de Compañías (Companies)
* `GET /api/companies`: Listar compañías/proveedores.
* `POST /api/companies`: Registrar nueva compañía.
* `GET /api/companies/:id`: Detalle de una compañía.
* `PUT /api/companies/:id`: Actualizar datos de la compañía.
* `DELETE /api/companies/:id`: Eliminar compañía (solo si no tiene cuentas de servicio asociadas).

### E. CRUD de Cuentas de Servicio (Service Accounts)
* `GET /api/service-accounts`: Listar cuentas de servicio.
* `POST /api/service-accounts`: Crear nueva cuenta de servicio asociada a una compañía.
* `GET /api/service-accounts/:id`: Detalle de una cuenta de servicio.
* `PUT /api/service-accounts/:id`: Actualizar cuenta de servicio.
* `DELETE /api/service-accounts/:id`: Eliminar cuenta de servicio.
* **Filtros:** Por compañía (`?company_id=123`).

### F. CRUD de Gastos (Expenses)
* `GET /api/expenses`: Listar gastos con filtros (período, categoría, cuenta de servicio, estado de pago).
* `POST /api/expenses`: Registrar nuevo gasto/cuenta.
* `GET /api/expenses/:id`: Detalle de un gasto específico.
* `PUT /api/expenses/:id`: Modificar datos del gasto.
* `DELETE /api/expenses/:id`: Eliminar registro de gasto.
* `PATCH /api/expenses/:id/pay`: Marcar gasto como pagado (actualiza `amount_paid`).
* **Campos especiales:**
  - Soporte para gastos recurrentes (`is_recurring`)
  - Soporte para cuotas (`current_installment`, `total_installments`, `installment_group_id`)
  - Fecha de vencimiento (`due_date`)

### G. CRUD de Ingresos (Incomes)
* `GET /api/incomes`: Listar ingresos con filtros (período, recurrente).
* `POST /api/incomes`: Registrar nuevo ingreso.
* `GET /api/incomes/:id`: Detalle de un ingreso específico.
* `PUT /api/incomes/:id`: Modificar datos del ingreso.
* `DELETE /api/incomes/:id`: Eliminar registro de ingreso.
* **Campos especiales:**
  - Soporte para ingresos recurrentes (`is_recurring`)
  - Fecha de recepción (`received_at`)

### H. Endpoints de Reportes y Resumen
* `GET /api/summary/:period_id`: Resumen financiero de un período (total ingresos, total gastos, balance).
* `GET /api/expenses/pending`: Listar gastos pendientes de pago (vencidos o próximos a vencer).

---

## 6. Reglas de Negocio y Seguridad
1.  **Aislamiento de Datos:** El `user_id` extraído del token JWT de Clerk debe filtrar todas las consultas (Un usuario no puede ver las cuentas de otro).
2.  **Validación de Montos:** No se permiten montos negativos.
3.  **Autenticación:** Todos los endpoints protegidos requieren token JWT válido de Clerk.
4.  **Manejo de Errores:** Respuestas estandarizadas:
    ```json
    {
      "status": "error",
      "message": "Descripción del error",
      "code": 400
    }
    ```
5.  **Persistencia:** Uso de migraciones para mantener la estructura de base de datos sincronizada.

---

## 7. Roadmap de Desarrollo
1.  **Fase 1:** Configuración de entorno Go 1.23 + conexión a Supabase (PostgreSQL 17).
2.  **Fase 2:** Implementación de autenticación con Clerk SDK Go v2.
3.  **Fase 3:** CRUD de Períodos, Categorías, Compañías, Cuentas de Servicio, Gastos e Ingresos.
4.  **Fase 4:** Lógica de reportes y resúmenes financieros.
5.  **Fase 5:** Documentación con Swagger/OpenAPI.
