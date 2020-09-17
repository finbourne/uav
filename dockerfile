############################
# STEP 1 build executable binary
############################

FROM golang:1.15.2-alpine as builder

# Create uavuser user and group.
RUN addgroup -S -g 18631 uav-user && \
    adduser -S -D -u 18631 -G uav-user uav

RUN mkdir /workdir
WORKDIR /workdir
COPY . .

RUN go get -d -v
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags="-w -s" -o /go/bin/uav

############################
# STEP 2 build a small image
############################
FROM scratch 

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/uav /uav

USER uav

ENTRYPOINT ["/uav"]