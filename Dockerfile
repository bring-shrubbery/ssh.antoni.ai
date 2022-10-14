FROM golang:1.19-alpine

ENV PORT 22
ENV GOOS=linux
ENV GOARCH=arm64

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o ssh-antoni-ai .

CMD [ "./ssh-antoni-ai" ]
