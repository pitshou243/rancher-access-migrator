FROM golang:1.24.0-alpine3.19 AS build
RUN apk add -U --no-cache make git bash
COPY . /src/rancher-access-migrator
WORKDIR /src/rancher-access-migrator
RUN ls -l
RUN make build

FROM alpine AS package
COPY --from=build /src/rancher-access-migrator/bin /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/rancher-access-migrator"]
