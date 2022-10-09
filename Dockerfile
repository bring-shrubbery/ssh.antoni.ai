FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY data/* ./data/

RUN go build -o ssh-antoni

EXPOSE 22

CMD [ "./ssh-antoni" ]