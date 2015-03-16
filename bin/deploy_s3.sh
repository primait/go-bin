#!/bin/sh

GIT_VERSION="$(git rev-parse --verify HEAD | cut -b 1-15)"
make
aws s3 cp s3://prima-deploy/parameters_go.yml parameters_prod.yml
zip -9 -r -q "gobin_$GIT_VERSION.zip" parameters_prod.yml dlx/dlx bin/start.sh supervisord.conf Dockerfile Dockerrun.aws.json
aws s3 cp "gobin_$GIT_VERSION.zip" s3://prima-deploy/gobin/"gobin_$GIT_VERSION.zip"
aws elasticbeanstalk create-application-version --application-name prima-go --version-label prima-go-$GIT_VERSION --description "Deploy master" --source-bundle S3Bucket=prima-deploy,S3Key="gobin/gobin_$GIT_VERSION.zip" --no-auto-create-application
aws elasticbeanstalk update-environment --environment-name prima-go-prod --version-label prima-go-$GIT_VERSION
rm "gobin_$GIT_VERSION.zip"
