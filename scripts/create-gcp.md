# ==========================================
# 1. VARIABLES - AJUSTA ESTOS VALORES
# ==========================================
PROJECT_ID="xxxxxxxx-dev-deploy"
REGION="us-central1"
POOL_ID="github-frontend-pol-dev"
PROVIDER_ID="github-frontend-provider-dev"
REPO_FULL_NAME="fvalenzuela-dev/template-nextjs-pro"
SA_NAME="github-deploy-sa-frontend-dev"
REPO_ARTIFACT="frontend-react-repo-dev"

# Obtener el número de proyecto automáticamente
PROJECT_NUMBER=$(gcloud projects describe ${PROJECT_ID} --format='value(projectNumber)')
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# ==========================================
# 2. IDENTIDAD Y FEDERACIÓN (WIF)
# ==========================================

# Crear Service Account
gcloud iam service-accounts create ${SA_NAME} --project=${PROJECT_ID}

# Crear Pool
gcloud iam workload-identity-pools create ${POOL_ID} \
    --location="global" --project=${PROJECT_ID} || true

# Crear Provider
gcloud iam workload-identity-pools providers create-oidc ${PROVIDER_ID} \
    --location="global" \
    --workload-identity-pool=${POOL_ID} \
    --issuer-uri="https://token.actions.githubusercontent.com" \
    --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
    --attribute-condition="attribute.repository_owner == 'fvalenzuela-dev'" \
    --project=${PROJECT_ID} || true

# Vincular SA con el Repositorio específico
gcloud iam service-accounts add-iam-policy-binding ${SA_EMAIL} \
    --role="roles/iam.workloadIdentityUser" \
    --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/attribute.repository/${REPO_FULL_NAME}" \
    --project=${PROJECT_ID}

# ==========================================
# 3. PERMISOS DE RECURSOS (Artifact + Cloud Run)
# ==========================================

# Crear e integrar Artifact Registry
gcloud artifacts repositories create ${REPO_ARTIFACT} \
    --repository-format=docker --location=${REGION} --project=${PROJECT_ID} || true

gcloud artifacts repositories add-iam-policy-binding ${REPO_ARTIFACT} \
    --location=${REGION} \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/artifactregistry.writer" \
    --project=${PROJECT_ID}

# Permisos para administrar Cloud Run
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/run.admin"

# Permiso para usar la cuenta de servicio por defecto de Compute (necesario para desplegar)
gcloud iam service-accounts add-iam-policy-binding ${PROJECT_NUMBER}-compute@developer.gserviceaccount.com \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/iam.serviceAccountUser" \
    --project=${PROJECT_ID}

echo "--------------------------------------------------------"
echo "CONFIGURACIÓN COMPLETADA"
echo "gcp_workload_identity_provider: projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/providers/${PROVIDER_ID}"
echo "Service Account: ${SA_EMAIL}"
echo "--------------------------------------------------------"