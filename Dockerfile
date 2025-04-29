# syntax=docker/dockerfile:1

##################
# Building Stage #
##################

FROM golang:1.23.4 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Default VERSION to "develop" for local builds;
# CI provides the version from the release branch
ARG VERSION=develop

# Extract module from go.mod and build the -X path
RUN MODULE=$(grep '^module' go.mod | awk '{print $2}') && \
    LDFLAGS="-X $MODULE/api/config.version=$VERSION" && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "$LDFLAGS" -o /output

#################
# Release Stage #
#################

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /output /output

ARG VERSION=dev
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.description="API powered by github.com/earcamone/gapi"

EXPOSE 8080

USER nonroot:nonroot

HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/output"]