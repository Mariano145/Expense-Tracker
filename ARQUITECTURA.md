# Arquitectura del Proyecto - Billetera

> **Nota:** Este archivo es para referencia personal. NO se sube a git (está en .gitignore).

## Estructura General

```
billetera/
├── backend/              # API en Go (net/http)
├── frontend/             # App en Next.js (React)
├── openspec/             # Artefactos SDD (specs, diseños, tareas)
├── .github/workflows/    # CI/CD con GitHub Actions
├── docker-compose.yml    # Servicios locales (PostgreSQL)
├── Makefile              # Comandos útiles (migraciones, DB)
└── README.md             # Documentación pública del proyecto
```

---

## Backend (Go)

### Estructura

```
backend/
├── cmd/
│   └── server/
│       └── main.go       # Punto de entrada del servidor HTTP
├── internal/
│   ├── config/           # Configuración (variables de entorno)
│   │   └── config.go     # Carga y valida env vars
│   ├── database/         # Conexión a PostgreSQL
│   │   └── postgres.go   # pgxpool setup + health check
│   ├── model/            # Modelos de dominio
│   │   ├── user.go       # User struct
│   │   ├── refresh_token.go  # RefreshToken struct
│   │   └── errors.go     # Sentinel errors
│   ├── repository/       # Acceso a base de datos
│   │   ├── user_repository.go
│   │   ├── user_repository_test.go
│   │   ├── refresh_token_repository.go
│   │   └── refresh_token_repository_test.go
│   ├── handler/          # Handlers de HTTP (controladores)
│   │   ├── health.go     # Endpoint /health
│   │   └── health_test.go
│   └── router/
│       └── router.go     # Configuración de rutas
├── migrations/           # Migraciones SQL (golang-migrate)
│   ├── 000001_baseline.up.sql
│   ├── 000001_baseline.down.sql
│   ├── 000002_create_users.up.sql
│   ├── 000002_create_users.down.sql
│   ├── 000003_create_refresh_tokens.up.sql
│   ├── 000003_create_refresh_tokens.down.sql
│   └── README.md
├── .env.example          # Plantilla de variables de entorno
├── go.mod                # Dependencias de Go
├── go.sum                # Checksums de dependencias
└── README.md             # Docs del backend
```

### ¿Qué hace cada cosa?

**`cmd/server/main.go`**
- Es el punto de entrada de la aplicación
- Inicia el servidor HTTP en el puerto configurado
- Configura el graceful shutdown (cierre limpio)

**`internal/handler/`**
- Acá van los handlers (controladores) de HTTP
- Cada handler procesa una petición HTTP y devuelve una respuesta
- Ejemplo: `health.go` maneja `GET /health` y devuelve `{"status":"ok"}`
- Los tests van junto al código: `health_test.go`

**`internal/router/`**
- Configura todas las rutas de la API
- Usa `http.ServeMux` (nativo de Go 1.22+)
- Acá registrás: `GET /expenses`, `POST /expenses`, etc.

**`migrations/`**
- Migraciones SQL para la base de datos
- Cada migración tiene dos archivos:
  - `XXXXXX_nombre.up.sql` → aplica el cambio
  - `XXXXXX_nombre.down.sql` → revierte el cambio
- Se ejecutan con `make migrate-up` y `make migrate-down`

**`go.mod` y `go.sum`**
- `go.mod`: lista las dependencias del proyecto
- `go.sum`: checksums para verificar integridad
- Se actualizan con `go mod tidy`

---

## Patrones de Arquitectura y Diseño

### Arquitectura de 3 Capas (Clean Architecture simplificada)

```
handler → service → repository → database
   ↓         ↓          ↓
  HTTP    Business    Queries
         Logic        SQL
```

**Capas:**

1. **Handler** (capa de presentación)
   - Recibe requests HTTP
   - Valida input
   - Devuelve responses
   - NO tiene lógica de negocio
   - Mapea errores a códigos HTTP

2. **Service** (capa de negocio)
   - Lógica de negocio (validar password, generar tokens, detectar robo)
   - NO sabe de HTTP
   - Usa interfaces de repository (no implementaciones concretas)
   - Acá van las reglas del dominio

3. **Repository** (capa de datos)
   - Queries a la base de datos
   - NO tiene lógica de negocio
   - Implementa interfaces definidas en el mismo paquete
   - Devuelve modelos de dominio

**Ventajas:**
- Cada capa es testeable independientemente
- Podés mockear el repository para testear el service sin PostgreSQL
- Fácil de mantener y extender

### Interfaces en cada límite

```go
// repository/user_repository.go
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// Implementación concreta
type postgresUserRepository struct {
    pool *pgxpool.Pool
}
```

**Ventaja:** en los tests del service, podés crear un mock que implementa la interfaz sin necesidad de PostgreSQL.

### Dependency Injection manual

```go
// main.go
func main() {
    cfg := config.Load()
    pool := database.NewPostgresDB(cfg.DatabaseURL)
    
    userRepo := repository.NewPostgresUserRepository(pool)
    tokenRepo := repository.NewPostgresRefreshTokenRepository(pool)
    
    authService := service.NewAuthService(userRepo, tokenRepo, cfg)
    authHandler := handler.NewAuthHandler(authService)
    
    // Wire up routes...
}
```

**Ventaja:** sin frameworks mágicos, fácil de entender, fácil de testear.

### Sentinel Errors

```go
// model/errors.go
var (
    ErrEmailExists         = errors.New("email already exists")
    ErrUserNotFound        = errors.New("user not found")
    ErrInvalidCredentials  = errors.New("invalid credentials")
    ErrTokenExpired        = errors.New("token expired")
    ErrTokenRevoked        = errors.New("token revoked")
    ErrTokenNotFound       = errors.New("token not found")
)
```

**Ventaja:** el handler puede mapear errores a códigos HTTP:
```go
if errors.Is(err, model.ErrEmailExists) {
    return http.StatusConflict // 409
}
if errors.Is(err, model.ErrUserNotFound) {
    return http.StatusNotFound // 404
}
```

### Repository Pattern

```go
// El service no sabe que usa PostgreSQL
type AuthService struct {
    userRepo  repository.UserRepository        // interfaz
    tokenRepo repository.RefreshTokenRepository // interfaz
}
```

**Ventaja:** si mañana cambiás PostgreSQL por MongoDB, solo cambiás la implementación del repository, el service sigue igual.

### DTOs (Data Transfer Objects)

```go
// model/dto.go (lo vamos a agregar en PR 2)
type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type AuthResponse struct {
    AccessToken string `json:"access_token"`
    User        *User  `json:"user"`
}
```

**Ventaja:** separa el modelo de dominio (User) del formato JSON que envía la API.

---

## Patrones de Seguridad

### httpOnly cookies para refresh tokens

- Protege contra XSS (JavaScript no puede leer la cookie)
- SameSite=Strict protege contra CSRF
- Solo el navegador puede enviar la cookie, no el código JS

### Access tokens en memoria (no localStorage)

- Si el usuario refresca la página, se pierde el token
- Pero se recupera con el refresh token automáticamente
- Protege contra XSS (no hay token persistente en el navegador)

### Family-based token rotation

- Cada refresh genera un nuevo token e invalida el anterior
- Si se reusa un token revocado, se invalidan TODOS los de esa familia
- Detecta robo de tokens automáticamente

**Ejemplo:**
```
1. Login → Token A (family: 123)
2. Refresh → Token B (family: 123), Token A revocado
3. Refresh → Token C (family: 123), Token B revocado
4. Alguien usa Token A (robado) → detecta reuso → revoca TODA la familia 123
5. Usuario legítimo no puede usar Token C → debe hacer login de nuevo
```

---

## Frontend (Next.js)

### Estructura

```
frontend/
├── src/
│   ├── app/              # App Router de Next.js
│   │   ├── layout.tsx    # Layout global (envuelve todas las páginas)
│   │   ├── page.tsx      # Página principal (/)
│   │   └── globals.css   # Estilos globales
│   └── components/       # Componentes reutilizables
│       ├── Hello.tsx     # Componente de ejemplo
│       └── Hello.test.tsx
├── public/               # Archivos estáticos (imágenes, favicon)
├── .env.example          # Plantilla de variables de entorno
├── next.config.ts        # Configuración de Next.js
├── vitest.config.ts      # Configuración de Vitest (tests)
├── vitest.setup.ts       # Setup de tests (jest-dom, etc.)
├── package.json          # Dependencias de Node
├── Dockerfile            # Imagen Docker optimizada (multi-stage)
└── README.md             # Docs del frontend
```

### ¿Qué hace cada cosa?

**`src/app/` (App Router)**
- **`layout.tsx`**: Layout global que envuelve todas las páginas. Acá va el `<html>`, `<body>`, navbar, footer, etc.
- **`page.tsx`**: La página principal (`/`). Cada carpeta dentro de `app/` puede tener su propio `page.tsx` para crear rutas.
- **`globals.css`**: Estilos globales (Tailwind, resets, etc.)

**`src/components/`**
- Componentes reutilizables (botones, cards, formularios)
- Cada componente tiene su test al lado: `Component.test.tsx`
- Ejemplo: `Hello.tsx` es un componente de ejemplo

**`public/`**
- Archivos estáticos accesibles desde `/`
- Ejemplo: `public/logo.png` → `http://localhost:3000/logo.png`

**`next.config.ts`**
- Configuración de Next.js
- Acá configurás: rewrites, redirects, env vars, output mode (standalone para Docker)

**`vitest.config.ts` y `vitest.setup.ts`**
- Configuración de Vitest (framework de testing)
- `vitest.setup.ts`: setup global (jest-dom matchers, etc.)

**`Dockerfile`**
- Imagen Docker optimizada con multi-stage build
- 3 etapas: deps → build → runner
- Usa Alpine (~5MB) y standalone output de Next.js
- Tamaño final: ~150-200MB

---

## Infraestructura

### `docker-compose.yml`

Levanta servicios locales para desarrollo:

```yaml
services:
  postgres:               # Base de datos PostgreSQL 16
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: billetera
      POSTGRES_USER: billetera
      POSTGRES_PASSWORD: billetera
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

**Comandos:**
- `docker compose up -d` → levanta PostgreSQL en background
- `docker compose down` → baja los servicios
- `docker compose logs -f postgres` → ver logs

### `Makefile`

Comandos útiles para desarrollo:

```makefile
db-up:                    # Levanta PostgreSQL
db-down:                  # Baja PostgreSQL
migrate-up:               # Aplica migraciones
migrate-down:             # Revierte última migración
```

**Uso:**
```bash
make db-up                # Levanta la DB
make migrate-up           # Aplica migraciones
```

### `.github/workflows/ci.yml`

Pipeline de CI/CD que se ejecuta en cada push/PR:

**Backend (Go):**
- `go vet ./...` → análisis estático
- `go test ./...` → tests
- `go build ./...` → compilación

**Frontend (Node):**
- `npm run lint` → linting
- `npm test` → tests
- `npm run build` → build de producción

---

## Flujo de Desarrollo

### 1. Setup inicial (una sola vez)

```bash
# Clonar repo
git clone <repo>
cd billetera

# Backend
cd backend
cp .env.example .env.local
go mod download

# Frontend
cd ../frontend
cp .env.example .env.local
npm install

# Base de datos
cd ..
make db-up
make migrate-up
```

### 2. Desarrollo diario

```bash
# Terminal 1: Base de datos
make db-up

# Terminal 2: Backend
cd backend
go run cmd/server/main.go

# Terminal 3: Frontend
cd frontend
npm run dev
```

### 3. Testing

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm test
```

### 4. Crear nueva feature

```bash
# Crear branch desde develop
git checkout develop
git pull origin develop
git checkout -b feature/nombre-feature

# Trabajar...
git add .
git commit -m "feat(scope): descripción"

# Push y crear PR
git push origin feature/nombre-feature
# Crear PR en GitHub: feature/nombre-feature → develop
```

---

## Convenciones

### Commits

Formato: `<type>(scope): <description>`

**Types:**
- `feat`: nueva funcionalidad
- `fix`: corrección de bug
- `docs`: cambios en documentación
- `style`: formato, sin cambio de lógica
- `refactor`: refactorización de código
- `test`: agregar o corregir tests
- `chore`: tareas de mantenimiento

**Ejemplos:**
```
feat(backend): add expense creation endpoint
fix(frontend): correct date formatting in dashboard
docs(readme): update setup instructions
test(handler): add tests for user registration
```

### Branches

- `main`: producción estable
- `develop`: integración de features
- `feature/*`: cada feature en su branch

---

## Próximos Pasos

### PR 2: Service + HTTP + Wiring (en progreso)

**Backend:**
- `internal/service/` → AuthService (lógica de autenticación)
  - `jwt.go` → generación y validación de JWT
  - `password.go` → hashing con bcrypt
  - `auth_service.go` → register, login, refresh, logout
- `internal/handler/` → AuthHandler (5 endpoints)
- `internal/middleware/` → Auth middleware (validación de JWT)
- `internal/model/dto.go` → DTOs para request/response

### PR 3: Frontend

**Frontend:**
- `src/app/(auth)/` → páginas de login/registro
- `src/app/(dashboard)/` → páginas protegidas
- `src/lib/` → Axios client con interceptors
- `src/store/` → Zustand auth store
- `src/middleware.ts` → protección de rutas

### PR 4: Documentation

**Documentación pública:**
- README.md actualizado con instrucciones de autenticación
- API documentation (endpoints, request/response examples)
- Este archivo (ARQUITECTURA.md) actualizado

---

## Recursos Útiles

- **Go docs:** https://pkg.go.dev/
- **Next.js docs:** https://nextjs.org/docs
- **PostgreSQL docs:** https://www.postgresql.org/docs/
- **Docker Compose:** https://docs.docker.com/compose/
