FROM golang:1.21
WORKDIR /app
COPY . .

#Run
RUN go get -v
RUN go mod tidy
RUN go build -v . 
EXPOSE 8080
CMD ["./crochess_backend"]
