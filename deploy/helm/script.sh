# #!/bin/bash
# for i in $(seq 1 10); do
#   helm upgrade --install edgeorch-site-$i ./edgeorch-lo \
#     -n site-$i \
#     --create-namespace \
#     --set site.id=site-$i \
#     --set lo.replicas=1 \
#     --set era.replicas=4
# done


#!/bin/bash
set -e

CHART_PATH=./edgeorch-lo
NUM_SITES=10
LO_REPLICAS=1
ERA_REPLICAS=4

echo "Deploying $NUM_SITES sites..."

for i in $(seq 1 $NUM_SITES); do
  SITE="site-$i"
  RELEASE="edgeorch-$SITE"

  echo "Deploying $SITE"

  kubectl create namespace $SITE --dry-run=client -o yaml | kubectl apply -f -

  helm upgrade --install $RELEASE $CHART_PATH \
    -n $SITE \
    --set site.id=$SITE \
    --set lo.replicas=$LO_REPLICAS \
    --set era.replicas=$ERA_REPLICAS \
    --set site.name=$SITE \
    --set era.lo.host=$RELEASE-lo \

done

echo "All sites deployed successfully."

