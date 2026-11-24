FROM golang:1.27.4-alpine AS build
WORKDIR /mtg-price-checker
# Copy dependencies list
COPY api ./api
WORKDIR /mtg-price-checker/api
RUN go mod download -x
# Build
RUN env GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -ldflags="-s -w" -o main cmd/main.go

# Copy artifacts to a clean image
FROM gcr.io/distroless/static-debian13@sha256:7f8dc7b7524fd10b4b855490a1837aacb4fb0be153f80e181c56902210f6a94f AS final
#FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /mtg-price-checker/api/main /main
ENTRYPOINT [ "/main" ]
