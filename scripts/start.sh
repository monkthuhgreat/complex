#!/usr/bin/env bash
set -e

NAMESPACE="fago"
K8S_DIR="manifests"

echo "Синтез аминокислот..."
docker build -t local/fago-gateway:v1 ./gateway
docker build -t local/fago-service:v1 ./service

echo "Смешивание первичного бульона..."
docker save local/fago-gateway:v1 | sudo k3s ctr images import -
docker save local/fago-service:v1 | sudo k3s ctr images import -

echo "Размещение гнезд..."
sudo k3s kubectl apply -f ${K8S_DIR}/Namespace.yaml

echo "Заточка копий..."
sudo k3s kubectl create secret generic fago-env \
  --from-env-file=.env \
  -n ${NAMESPACE} \
  --dry-run=client -o yaml | sudo k3s kubectl apply -f -

echo "Прокладка дорог..."
sudo k3s kubectl apply -Rf ${K8S_DIR}/

echo "Калибровка терраформеров..."
sudo k3s kubectl rollout restart deployment -n ${NAMESPACE}

echo "Готово! Можете эксплуатировать"