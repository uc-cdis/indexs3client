FROM golang:1.10 as build-deps

WORKDIR /indexs3client
ENV GOPATH=/indexs3client

COPY . /indexs3client

RUN go build -ldflags "-linkmode external -extldflags -static"

# Store only the resulting binary in the final image
# Resulting in significantly smaller docker image size
FROM scratch
COPY --from=build-deps /indexs3client/indexs3client /indexs3client
CMD ["/indexs3client"]
