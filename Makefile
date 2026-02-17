deploy: deploy-common docker-tag docker-push lambda-update frontend-update

deploy-staging: deploy-common docker-tag-staging docker-push-staging lambda-update-staging frontend-update-staging

deploy-common: test docker-build aws-login

docker-build:
	docker buildx build --platform linux/amd64 --provenance=false -t mtg-price-scrapper .

docker-tag:
	docker tag mtg-price-scrapper 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:latest

docker-tag-staging:
	docker tag mtg-price-scrapper 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:staging

docker-push:
	export AWS_PAGER="" && docker push 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:latest

docker-push-staging:
	export AWS_PAGER="" && docker push 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:staging

# Legacy web-update targets removed

frontend-dev:
	cd frontend && npm install && npm run dev

frontend-build:
	cd frontend && npm install && npm run build

frontend-update: frontend-build
	aws s3 sync frontend/dist s3://gishathfetch.com
	export AWS_PAGER="" && aws cloudfront create-invalidation --distribution-id E3NPGUM21YCN36 --paths "/*"

# Legacy web-update-staging targets removed

frontend-update-staging: frontend-build
	aws s3 sync frontend/dist s3://staging.gishathfetch.com
	export AWS_PAGER="" && aws cloudfront create-invalidation --distribution-id E33AK6HADX83U0 --paths "/*"

lambda-create:
	export AWS_PAGER="" && aws lambda create-function \
      --function-name mtg-price-scrapper \
      --package-type Image \
      --code ImageUri=206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:latest \
      --role arn:aws:iam::206363131200:role/lambda-mtg

lambda-create-staging:
	export AWS_PAGER="" && aws lambda create-function \
      --function-name mtg-price-scrapper-staging \
      --package-type Image \
      --code ImageUri=206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:staging \
      --role arn:aws:iam::206363131200:role/lambda-mtg

lambda-update:
	export AWS_PAGER="" && aws lambda update-function-code \
      --function-name mtg-price-scrapper \
      --image-uri 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:latest

lambda-update-staging:
	export AWS_PAGER="" && aws lambda update-function-code \
      --function-name mtg-price-scrapper-staging \
      --image-uri 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:staging

aws-login:
	aws ecr get-login-password --region ap-southeast-1 | docker login --username AWS --password-stdin 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com

test:
	cd api && go clean -testcache && go test -mod=vendor -failfast -timeout 5m ./... || (echo "\n\033[0;31mTESTS FAILED\033[0m. Continue deployment? [y/N] "; read ans; [ $${ans:-N} = y ])
