FROM golang:1.21-alpine3.18 as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -v -o go_crochess_backend . 

#Run
FROM alpine:3.18
COPY --from=build /app/go_crochess_backend /
COPY --from=build /app/.config.toml /
COPY --from=build /app/src/database/migrations/*.sql /src/database/migrations/
EXPOSE 8080
CMD ["./go_crochess_backend"]
