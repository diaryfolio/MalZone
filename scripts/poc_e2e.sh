#!/usr/bin/env bash
set -euo pipefail

POC_NAMESPACE="${MALZONE_NAMESPACE:-malzone-system}"
POC_PORT="${MALZONE_POC_PORT:-18080}"
POC_NAME="poc-e2e-$(date +%s)"
POC_CANCEL_NAME="${POC_NAME}-cancel"
POC_BASE_URL="http://127.0.0.1:${POC_PORT}"
POC_FORWARD_LOG="/tmp/malzone-poc-port-forward.log"

cleanup() {
  kubectl -n "${POC_NAMESPACE}" delete analysis "${POC_NAME}" --ignore-not-found --wait=false >/dev/null 2>&1 || true
  kubectl -n "${POC_NAMESPACE}" delete analysis "${POC_CANCEL_NAME}" --ignore-not-found --wait=false >/dev/null 2>&1 || true
  if [[ -n "${POC_FORWARD_PID:-}" ]]; then
    kill "${POC_FORWARD_PID}" >/dev/null 2>&1 || true
    wait "${POC_FORWARD_PID}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

kubectl -n "${POC_NAMESPACE}" wait deployment/malzone-api deployment/malzone-operator \
  --for=condition=Available --timeout=120s
kubectl -n "${POC_NAMESPACE}" port-forward service/malzone-api "${POC_PORT}:8080" \
  >"${POC_FORWARD_LOG}" 2>&1 &
POC_FORWARD_PID=$!

for _ in $(seq 1 30); do
  if curl --fail --silent "${POC_BASE_URL}/readyz" >/dev/null; then
    break
  fi
  sleep 1
done
curl --fail --silent "${POC_BASE_URL}/readyz" >/dev/null

test "$(kubectl auth can-i create jobs.batch --as "system:serviceaccount:${POC_NAMESPACE}:malzone-api" -n "${POC_NAMESPACE}")" = "no"
test "$(kubectl auth can-i list analyses.malzone.io --as "system:serviceaccount:${POC_NAMESPACE}:malzone-api" -n default)" = "no"
test "$(kubectl auth can-i get secrets --as "system:serviceaccount:${POC_NAMESPACE}:malzone-operator" -n "${POC_NAMESPACE}")" = "no"
test "$(kubectl auth can-i list analyses.malzone.io --as "system:serviceaccount:${POC_NAMESPACE}:malzone-operator" -n default)" = "no"
test "$(kubectl auth can-i get pods --as "system:serviceaccount:${POC_NAMESPACE}:malzone-runner" -n "${POC_NAMESPACE}")" = "no"

POC_REJECT_STATUS=$(curl --silent --output /dev/null --write-out '%{http_code}' \
  -H 'Content-Type: application/json' \
  -d '{"sample":{"kind":"command","content":"whoami"}}' \
  "${POC_BASE_URL}/api/v1alpha1/analyses")
test "${POC_REJECT_STATUS}" = "422"

curl --fail --silent \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"${POC_NAME}\",\"sample\":{\"kind\":\"canary\",\"content\":\"hello-malzone\"},\"timeoutSeconds\":2}" \
  "${POC_BASE_URL}/api/v1alpha1/analyses" >/dev/null

POC_PHASE=""
POC_RESPONSE=""
for _ in $(seq 1 60); do
  POC_RESPONSE=$(curl --fail --silent "${POC_BASE_URL}/api/v1alpha1/analyses/${POC_NAME}")
  POC_PHASE=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("status", {}).get("phase", ""))' <<<"${POC_RESPONSE}")
  if [[ "${POC_PHASE}" == "Succeeded" || "${POC_PHASE}" == "Failed" || "${POC_PHASE}" == "Cancelled" ]]; then
    break
  fi
  sleep 1
done
test "${POC_PHASE}" = "Succeeded"

python3 -c '
import json, sys
analysis = json.load(sys.stdin)
status = analysis["status"]
result = status["result"]
assert status["cleanupVerified"] is True
assert result["serviceAccountTokenAbsent"] is True
assert result["kubernetesApiDenied"] is True
assert len(result["sha256"]) == 64
' <<<"${POC_RESPONSE}"

POC_RESIDUE=$(kubectl -n "${POC_NAMESPACE}" get jobs,pods \
  -l "malzone.io/analysis=${POC_NAME}" -o name)
test -z "${POC_RESIDUE}"

curl --fail --silent \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"${POC_CANCEL_NAME}\",\"sample\":{\"kind\":\"canary\",\"content\":\"cancel-malzone\"},\"timeoutSeconds\":30}" \
  "${POC_BASE_URL}/api/v1alpha1/analyses" >/dev/null

for _ in $(seq 1 30); do
  POC_CANCEL_PHASE=$(curl --fail --silent "${POC_BASE_URL}/api/v1alpha1/analyses/${POC_CANCEL_NAME}" \
    | python3 -c 'import json,sys; print(json.load(sys.stdin).get("status", {}).get("phase", ""))')
  if [[ "${POC_CANCEL_PHASE}" == "Running" ]]; then
    break
  fi
  sleep 1
done
test "${POC_CANCEL_PHASE}" = "Running"

POC_CANCEL_STARTED=$(date +%s)
curl --fail --silent -X POST "${POC_BASE_URL}/api/v1alpha1/analyses/${POC_CANCEL_NAME}/cancel" >/dev/null
POC_CANCEL_RESPONSE=""
for _ in $(seq 1 60); do
  POC_CANCEL_RESPONSE=$(curl --fail --silent "${POC_BASE_URL}/api/v1alpha1/analyses/${POC_CANCEL_NAME}")
  POC_CANCEL_PHASE=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("status", {}).get("phase", ""))' <<<"${POC_CANCEL_RESPONSE}")
  if [[ "${POC_CANCEL_PHASE}" == "Cancelled" ]]; then
    break
  fi
  sleep 1
done
test "${POC_CANCEL_PHASE}" = "Cancelled"
POC_CANCEL_SECONDS=$(($(date +%s) - POC_CANCEL_STARTED))
test "${POC_CANCEL_SECONDS}" -lt 15
python3 -c '
import json, sys
status = json.load(sys.stdin)["status"]
assert status["cleanupVerified"] is True
' <<<"${POC_CANCEL_RESPONSE}"

POC_CANCEL_RESIDUE=$(kubectl -n "${POC_NAMESPACE}" get jobs,pods \
  -l "malzone.io/analysis=${POC_CANCEL_NAME}" -o name)
test -z "${POC_CANCEL_RESIDUE}"

echo "POC end-to-end test passed: success, prompt cancellation (${POC_CANCEL_SECONDS}s), cleanup, RBAC, token, and network denial checks"
