FROM golang:1.22.3

WORKDIR /app/menhera_api

RUN go install github.com/cosmtrek/air@v1.40.4

EXPOSE 3012
# サービスの開始コマンド
CMD ["make", "start"]