FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /sendyourpicks ./cmd/server/main.go

FROM gcr.io/distroless/static-debian12
COPY --from=build /sendyourpicks /sendyourpicks
ENTRYPOINT ["/sendyourpicks"]
