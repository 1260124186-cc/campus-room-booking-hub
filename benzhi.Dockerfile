FROM golang:1.26 AS build

ENV GOTOOLCHAIN=local
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build ./...

FROM golang:1.26

ENV GOTOOLCHAIN=local
WORKDIR /app
COPY --from=build /src /app
EXPOSE 8080
CMD ["go", "run", "./cmd/server"]
