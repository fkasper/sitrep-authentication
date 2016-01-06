#! /bin/bash
cd $HOME/sitrep-builds/src/github.com/fkasper/sitrep-biometrics

docker build -t $EXTERNAL_REGISTRY_ENDPOINT/$PROJECT_NAME:$CIRCLE_SHA1 .
mkdir -p ~/docker; docker save $EXTERNAL_REGISTRY_ENDPOINT/$PROJECT_NAME:$CIRCLE_SHA1 > ~/docker/image.tar

echo $GCLOUD_KEY | base64 --decode > gcloud.p12
gcloud auth activate-service-account $GCLOUD_EMAIL --key-file gcloud.p12
ssh-keygen -f ~/.ssh/google_compute_engine -N ""



gcloud docker push $EXTERNAL_REGISTRY_ENDPOINT/$PROJECT_NAME > /dev/null
