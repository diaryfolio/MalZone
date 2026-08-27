.PHONY: check design-check go-test helm-check build-poc image-poc import-poc deploy-poc e2e-poc

MALZONE_NAMESPACE ?= malzone-system
K3D_CLUSTER ?= cks
POC_IMAGE ?= malzone-poc:dev
TARGETARCH ?= arm64

check: design-check go-test helm-check

design-check:
	python3 -m unittest discover -s tests -p 'test_*.py'

go-test:
	go test ./...

helm-check:
	helm lint charts/malzone
	helm template malzone charts/malzone --namespace $(MALZONE_NAMESPACE) >/dev/null

build-poc:
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=$(TARGETARCH) go build -trimpath -ldflags='-s -w' -o build/malzone ./cmd/malzone

image-poc: build-poc
	docker build --platform linux/$(TARGETARCH) -t $(POC_IMAGE) .

import-poc: image-poc
	k3d image import --cluster $(K3D_CLUSTER) $(POC_IMAGE)

deploy-poc: import-poc
	kubectl create namespace $(MALZONE_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	kubectl label namespace $(MALZONE_NAMESPACE) pod-security.kubernetes.io/enforce=restricted --overwrite
	kubectl apply -f charts/malzone/crds/malzone.io_analyses.yaml
	helm upgrade --install malzone charts/malzone --namespace $(MALZONE_NAMESPACE) --set image.repository=$(word 1,$(subst :, ,$(POC_IMAGE))) --set image.tag=$(word 2,$(subst :, ,$(POC_IMAGE)))
	kubectl -n $(MALZONE_NAMESPACE) rollout restart deployment/malzone-api deployment/malzone-operator
	kubectl -n $(MALZONE_NAMESPACE) rollout status deployment/malzone-api deployment/malzone-operator --timeout=120s

e2e-poc:
	MALZONE_NAMESPACE=$(MALZONE_NAMESPACE) scripts/poc_e2e.sh
