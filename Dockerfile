FROM quay.io/cdis/golang:1.17-bullseye as build-deps

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR $GOPATH/src/github.com/uc-cdis/indexs3client/

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GITCOMMIT=$(git rev-parse HEAD) \
    GITVERSION=$(git describe --always --tags) \
    && go build \
    -ldflags="-X 'github.com/uc-cdis/indexs3client/handlers/version.GitCommit=${GITCOMMIT}' -X 'github.com/uc-cdis/indexs3client/handlers/version.GitVersion=${GITVERSION}'" \
    -o /indexs3client

FROM scratch
USER nobody
COPY --from=build-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-deps /indexs3client /indexs3client
CMD ["/indexs3client"]
