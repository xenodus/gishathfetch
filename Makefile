AWS_ACCOUNT_ID := 206363131200
AWS_REGION := ap-southeast-1
ECR_REPO := mtg-price-scrapper
ECR_IMAGE := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(ECR_REPO):latest
LAMBDA_ROLE := arn:aws:iam::$(AWS_ACCOUNT_ID):role/lambda-mtg
LAMBDA_FUNCTIONS := mtg-price-scrapper mtg-price-ck-refresh mtg-analytics-keywords-export
ANALYTICS_KEYWORDS_EXPORT_RULE := analytics-keywords-export-daily
ANALYTICS_KEYWORDS_EXPORT_LAMBDA := mtg-analytics-keywords-export
ANALYTICS_KEYWORDS_EXPORT_ACTION := analytics-keywords-export-run
ANALYTICS_KEYWORDS_EXPORT_SCHEDULE := rate(1 hour)

deploy: deploy-common docker-tag docker-push lambda-update frontend-update

deploy-common: docker-build aws-login

docker-build:
	docker buildx build --platform linux/amd64 --provenance=false --load -t $(ECR_REPO) .

docker-tag:
	docker tag $(ECR_REPO) $(ECR_IMAGE)

docker-push:
	export AWS_PAGER="" && docker push $(ECR_IMAGE)

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
	aws s3 sync frontend/dist s3://gishathfetch.com \
		--exclude ".well-known/http-message-signatures-directory" \
		--exclude "robots.txt" \
		--exclude "ads.txt" \
		--exclude "analytics/*"
	aws s3 sync frontend/dist/assets s3://gishathfetch.com/assets --delete
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
      --code ImageUri=$(ECR_IMAGE) \
      --role $(LAMBDA_ROLE)

lambda-update:
	@for fn in $(LAMBDA_FUNCTIONS); do \
		echo "Updating Lambda $$fn"; \
		export AWS_PAGER="" && aws lambda update-function-code \
			--function-name $$fn \
			--image-uri $(ECR_IMAGE) \
			--region $(AWS_REGION) \
			--output text > /dev/null; \
	done

aws-login:
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

# Hourly GA4 keyword export: updates the existing EventBridge rule schedule and target.
eventbridge-analytics-keywords-hourly:
	export AWS_PAGER="" && aws events put-rule \
		--name $(ANALYTICS_KEYWORDS_EXPORT_RULE) \
		--schedule-expression "$(ANALYTICS_KEYWORDS_EXPORT_SCHEDULE)" \
		--state ENABLED \
		--description "Hourly GA4 top-search-keywords export to S3" \
		--region $(AWS_REGION)
	export AWS_PAGER="" && aws events put-targets \
		--rule $(ANALYTICS_KEYWORDS_EXPORT_RULE) \
		--targets "Id=1,Arn=arn:aws:lambda:$(AWS_REGION):$(AWS_ACCOUNT_ID):function:$(ANALYTICS_KEYWORDS_EXPORT_LAMBDA),Input={\"action\":\"$(ANALYTICS_KEYWORDS_EXPORT_ACTION)\"}" \
		--region $(AWS_REGION)
	export AWS_PAGER="" && aws lambda add-permission \
		--function-name $(ANALYTICS_KEYWORDS_EXPORT_LAMBDA) \
		--statement-id $(ANALYTICS_KEYWORDS_EXPORT_RULE) \
		--action lambda:InvokeFunction \
		--principal events.amazonaws.com \
		--source-arn arn:aws:events:$(AWS_REGION):$(AWS_ACCOUNT_ID):rule/$(ANALYTICS_KEYWORDS_EXPORT_RULE) \
		--region $(AWS_REGION) || true

test:
	cd api && go clean -testcache && go test -mod=vendor -failfast -timeout 5m ./...
