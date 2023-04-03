FROM arm32v7/golang:1.20.2-alpine3.17 AS build
RUN apk --no-cache add gcc g++ make ca-certificates
WORKDIR /go/src

COPY . .

RUN GO111MODULE=on go build -o /go/bin/app ./cmd

FROM arm32v7/alpine:3.17
WORKDIR /usr/bin
COPY --from=build /go/bin .
CMD ["app"]