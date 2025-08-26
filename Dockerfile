FROM golang:1.25.0-alpine AS build
WORKDIR /mtg-price-checker
# Copy dependencies list
COPY api ./api
WORKDIR /mtg-price-checker/api
RUN go mod download -x
# Build
RUN env GOOS=linux GOARCH=amd64 GOEXPERIMENT=greenteagc go build -tags lambda.norpc -ldflags="-s -w" -o main cmd/main.go
# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /mtg-price-checker/api/main /main
ENTRYPOINT [ "/main" ]
