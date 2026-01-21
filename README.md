# Stock Viewer Backend

API REST en Go para visualizar y analizar recomendaciones de stocks.

## Tecnologías

- **Go 1.24+**
- **Gin-Gonic** - Framework HTTP
- **GORM** - ORM
- **CockroachDB** - Base de datos
- **Swagger** - Documentación de API
- **Docker** - Containerización

## Requisitos

- Go 1.24+
- Docker y Docker Compose
- Make (opcional)

## Inicio Rápido

### 1. Clonar y configurar

```bash
cd go-stock-viewer-back
cp env.template .env
# Edit .env and add your KARENAI_TOKEN and BASIC_AUTH_PASSWORD
```

### 2. Ejecutar con Docker

```bash
chmod +x scripts/run_local.sh
./scripts/run_local.sh
```

### 3. Verificar

- API: http://localhost:9000/ping
- Swagger: http://localhost:9000/swagger/index.html
- CockroachDB Admin: http://localhost:8081

## Endpoints

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/ping` | Health check |
| GET | `/health` | Health check detallado |
| GET | `/api/v1/stocks` | Listar stocks con filtros |
| GET | `/api/v1/stocks/:id` | Obtener stock por ID |
| GET | `/api/v1/stocks/search` | Buscar stocks |
| GET | `/api/v1/stocks/filters` | Obtener filtros disponibles |
| GET | `/api/v1/recommendations` | Obtener recomendaciones |
| POST | `/api/v1/sync` | Sincronizar datos (Auth requerida) |

## Autenticación

El endpoint `/api/v1/sync` requiere Basic Authentication:

```bash
curl -X POST http://localhost:9000/api/v1/sync \
  -u $BASIC_AUTH_USER:$BASIC_AUTH_PASSWORD
```

## Estructura del Proyecto

```
go-stock-viewer-back/
├── src/
│   ├── cmd/api/              # Entry point
│   │   └── main.go
│   └── stockviewer/          # Código principal
│       ├── types.go          # Tipos y entidades
│       ├── errors.go         # Errores personalizados
│       ├── config/           # Configuración
│       ├── httpapi/          # Controladores HTTP
│       ├── stocks/           # Servicio de stocks
│       ├── recommendation/   # Servicio de recomendaciones
│       ├── integrations/     # Clientes externos
│       │   └── karenai/
│       └── mocks/            # Mocks para testing
├── scripts/
│   └── run_local.sh
├── docs/                     # Swagger docs (generados)
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## Testing

```bash
go test ./src/... -v
```

## Variables de Entorno

| Variable | Descripción | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_PORT` | Puerto del servidor | 8080 | No |
| `GIN_MODE` | Modo de Gin | debug | No |
| `DB_HOST` | Host de CockroachDB | cockroachdb | No |
| `DB_PORT` | Puerto de CockroachDB | 26257 | No |
| `DB_USER` | Usuario de DB | root | No |
| `DB_PASSWORD` | Password de DB | - | No |
| `DB_NAME` | Nombre de la DB | stockviewer | No |
| `KARENAI_BASE_URL` | URL de la API externa | https://api.karenai.click | No |
| `KARENAI_TOKEN` | Token de autenticación | - | **Yes** |
| `BASIC_AUTH_USER` | Usuario para auth básica | admin | No |
| `BASIC_AUTH_PASSWORD` | Password para auth básica | - | **Yes** (Required, no default) |

> ⚠️ **Security Note**: 
> - Never commit sensitive values like `KARENAI_TOKEN` and `BASIC_AUTH_PASSWORD` to version control
> - `BASIC_AUTH_PASSWORD` is **required** and has no default value - the application will fail to start if not set
> - Use the `env.template` file as a reference and create your own `.env` file locally
> - Rotate passwords regularly for security
