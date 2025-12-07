# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Installer les dépendances de build (nécessaire pour sqlite)
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source
COPY . .

# Compiler le binaire
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o gip .

# Stage 2: Runtime
FROM alpine:latest

# Installer les librairies nécessaires pour SQLite
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copier le binaire depuis le builder
COPY --from=builder /app/gip .

# Exposer le port par défaut
EXPOSE 8080

# Créer un volume pour la base de données
VOLUME ["/data"]

# Commande par défaut
ENTRYPOINT ["./gip"]
CMD ["serve", "--database", "/data/data.db"]
