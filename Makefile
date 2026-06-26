deploy: deploy-common docker-tag docker-push lambda-update frontend-update

deploy-common: docker-build aws-login

docker-build:
	docker buildx build --platform linux/amd64 --provenance=false --load -t mtg-price-scrapper .

docker-tag:
	docker tag mtg-price-scrapper 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:latest

docker-push:
	export AWS_PAGER="" && docker push 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:latest

frontend-dev:
	cd frontend && npm install && npm run dev

frontend-build: generate-signature-directory
	cd frontend && npm install && npm run build

SIGNATURE_DIRECTORY_BIN=.cache/signature-directory

generate-signature-directory:
	@if [ -n "$$WEB_BOT_AUTH_PRIVATE_KEY" ] || { [ -n "$$WEB_BOT_AUTH_PRIVATE_KEY_FILE" ] && [ -s "$$WEB_BOT_AUTH_PRIVATE_KEY_FILE" ]; }; then \
		mkdir -p frontend/public/.well-known .cache && \
		cd api && go build -mod=vendor -o ../$(SIGNATURE_DIRECTORY_BIN) ./cmd/signature-directory && \
		../$(SIGNATURE_DIRECTORY_BIN) -out ../frontend/public/.well-known/http-message-signatures-directory; \
	else \
		echo "Skipping signature directory generation: Web Bot Auth private key not configured"; \
	fi

frontend-update: frontend-build
	aws s3 sync frontend/dist s3://gishathfetch.com --exclude ".well-known/http-message-signatures-directory"
	@if [ -f frontend/dist/.well-known/http-message-signatures-directory ]; then \
		export AWS_PAGER="" && aws s3 cp frontend/dist/.well-known/http-message-signatures-directory \
			s3://gishathfetch.com/.well-known/http-message-signatures-directory \
			--content-type "application/http-message-signatures-directory+json" \
			--cache-control "max-age=86400"; \
	fi
	export AWS_PAGER="" && aws cloudfront create-invalidation --distribution-id E3NPGUM21YCN36 --paths "/*"

lambda-create:
	export AWS_PAGER="" && aws lambda create-function \
      --function-name mtg-price-scrapper \
      --package-type Image \
      --code ImageUri=206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:latest \
      --role arn:aws:iam::206363131200:role/lambda-mtg

lambda-update:
	export AWS_PAGER="" && aws lambda update-function-code \
      --function-name mtg-price-scrapper \
      --image-uri 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com/mtg-price-scrapper:latest \
      --output text > /dev/null

aws-login:
	aws ecr get-login-password --region ap-southeast-1 | docker login --username AWS --password-stdin 206363131200.dkr.ecr.ap-southeast-1.amazonaws.com

test:
	cd api && go clean -testcache && go test -mod=vendor -failfast -timeout 5m ./...
