FROM golang:1.23.4-alpine AS builder

WORKDIR /app/builder

RUN apk add --no-cache npm

RUN echo "Build-time variable BUILD_TIME_VAR is set to: $BUILD_TIME_VAR"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN npm install

RUN npm run dev

RUN go build -o /app/builder/mentoref ./cmd/main.go

FROM alpine:3.19 AS runner

WORKDIR /app

ENV DB_PORT="5432"
ENV DB_NAME="mentoref"
ENV DB_USER="mentoref_user"

COPY --from=builder /app/builder/mentoref /app/mentoref

COPY --from=builder /app/builder/web/css /app/web/css
COPY --from=builder /app/builder/web/templates /app/web/templates

EXPOSE 3000

CMD ["./mentoref"]
