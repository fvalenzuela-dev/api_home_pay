# PRD v2 - Mejoras a api-home-pay

> Este documento cubre **solo los cambios nuevos** a aplicar sobre el proyecto existente.
> Para el PRD completo del proyecto, ver `prd.md`.

---

## 1. Resumen de Cambios

| # | Cambio | Prioridad |
|---|--------|-----------|
| 1 | Implementar logging estructurado | Alta |
| 2 | Actualizar workflows CI/CD con nombres correctos de servicios GCP | Alta |

---

## 2. Logging Estructurado

### Objetivo
Agregar logging estructurado en formato JSON para integración con GCP Cloud Logging y facilitar debugging en producción.

### Requisitos
- **Formato:** JSON estructurado.
- **Niveles:** `debug`, `info`, `warn`, `error`.
- **Biblioteca:** `log/slog` (estándar de Go 1.21+, no requiere dependencia externa).

### Qué se debe loggear
| Evento | Nivel | Contexto |
|--------|-------|----------|
| Request entrante | `info` | método, path, status, duración, request_id |
| Respuesta enviada | `info` | status code, duración |
| Error de autenticación | `warn` | tipo de error (token expirado, inválido) |
| Error de base de datos | `error` | query, error message |
| Error de validación/business | `warn` | path, error message |
| Start/stop del servidor | `info` | puerto |
| Intento de eliminar con datos asociados | `warn` | entidad, user_id |

### Qué NO se debe loggear
- Tokens JWT completos
- Secret keys o passwords
- Datos sensibles del usuario (solo user_id)

### Archivos a modificar
- `cmd/api/main.go` — inicializar logger, log de start/stop
- `internal/middleware/` — middleware de logging de requests
- `internal/handlers/` — logs de errores de negocio
- `internal/repository/` — logs de errores de DB

---

## 3. Workflows CI/CD

### Problema actual
Los workflows `docker-gcp-dev.yml` y `docker-gcp-prod.yml` tienen nombres de servicios heredados de otro proyecto:
- DEV: `mi-app-next-dev` → debe ser `api-home-pay-dev`
- PROD: `prod-deploy` → debe ser `api-home-pay-prod`

### Cambios
| Workflow | Campo | Valor Actual | Valor Nuevo |
|----------|-------|-------------|-------------|
| docker-gcp-dev.yml | `image_name` | `mi-app-next-dev` | `api-home-pay-dev` |
| docker-gcp-dev.yml | `cloudrun_service` | `mi-app-next-dev-deploy` | `api-home-pay-dev` |
| docker-gcp-dev.yml | `artifact_registry_repository` | `frontend-react-repo-dev` | `api-home-pay-repo-dev` |
| docker-gcp-prod.yml | `image_name` | `mi-app-next-prod` | `api-home-pay-prod` |
| docker-gcp-prod.yml | `cloudrun_service` | `prod-deploy` | `api-home-pay-prod` |
| docker-gcp-prod.yml | `artifact_registry_repository` | `app-frontend-react-repo` | `api-home-pay-repo` |
