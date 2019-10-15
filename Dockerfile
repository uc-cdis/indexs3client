FROM golang:1.13.1 as build-deps
RUN mkdir -p /indexs3client
WORKDIR /indexs3client
ADD . .
RUN go build -tags netgo -ldflags '-extldflags "-static"' -o indexs3client

# Store only the resulting binary in the final image
# Resulting in significantly smaller docker image size
FROM scratch
COPY --from=build-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-deps /indexs3client/indexs3client /indexs3client

CMD ["/indexs3client"]
