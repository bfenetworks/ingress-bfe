FROM golang:1.16-alpine3.14 AS build

WORKDIR /bfe-ingress-controller
COPY . .
RUN build/build.sh

FROM bfenetworks/bfe:v-1.3.0
WORKDIR /
COPY --from=build /bfe-ingress-controller/output/* /

EXPOSE 8080 8443 8421

ENTRYPOINT ["/start.sh", "/bfe-ingress-controller", "/bfe/bin/bfe"]
