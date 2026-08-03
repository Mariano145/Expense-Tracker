# Backup Feature 1: Foundation + Repository

> **Nota:** Este archivo es para referencia personal y aprendizaje. NO se sube a git.

## Resumen de la Feature

**Objetivo:** Establecer la base del backend con migraciones, configuración, conexión a base de datos, modelos de dominio y capa de repositorio.

**Qué se implementó:**
- 2 migraciones (users, refresh_tokens)
- Sistema de configuración (variables de entorno)
- Pool de conexiones a PostgreSQL
- Modelos de dominio (User, RefreshToken)
- Sentinel errors
- 2 repositories (UserRepository, RefreshTokenRepository)
- 13 tests pasando

---

## Base de Datos

### ¿Por qué necesitamos estas tablas?

**Contexto:** Estamos implementando autenticación con JWT. Necesitamos:
1. **Guardar usuarios** (email + password hasheado)
2. **Guardar refresh tokens** (para renovar access tokens sin pedir password de nuevo)

### Tabla: `users`

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

#### Columnas explicadas

**`id UUID PRIMARY KEY DEFAULT gen_random_uuid()`**
- **Tipo:** UUID (Universally Unique Identifier)
- **Por qué UUID y no INT:**
  - No revela cuántos usuarios tenés (INT secuencial sí)
  - Se puede generar en el cliente sin consultar la DB
  - Más seguro para APIs (no podés adivinar IDs)
- **DEFAULT gen_random_uuid():** PostgreSQL genera el UUID automáticamente
- **PRIMARY KEY:** identificador único, no puede repetirse

**`email VARCHAR(255) NOT NULL UNIQUE`**
- **Tipo:** VARCHAR(255) - texto de hasta 255 caracteres
- **NOT NULL:** no puede estar vacío
- **UNIQUE:** no puede haber dos usuarios con el mismo email
- **Por qué 255:** estándar RFC 5321 para emails

**`password_hash VARCHAR(255) NOT NULL`**
- **Tipo:** VARCHAR(255)
- **Por qué "hash" y no "password":** NUNCA guardamos passwords en texto plano
- **Contenido:** hash generado por bcrypt (ej: `$2a$12$LJ3m4ys3Lg0EP7hZw.q.e...`)
- **Por qué 255:** bcrypt genera hashes de ~60 caracteres, pero dejamos espacio por si cambiamos el algoritmo

**`created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`**
- **Tipo:** TIMESTAMPTZ (timestamp con timezone)
- **DEFAULT NOW():** se setea automáticamente al insertar
- **Por qué:** para saber cuándo se creó el usuario (auditoría, debugging)

**`updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`**
- **Tipo:** TIMESTAMPTZ
- **Trigger:** se actualiza automáticamente en cada UPDATE
- **Por qué:** para saber cuándo se modificó por última vez

#### Índice: `idx_users_email`

```sql
CREATE INDEX idx_users_email ON users(email);
```

**Por qué:**
- Las búsquedas por email son muy frecuentes (login, validación de duplicados)
- Sin índice: PostgreSQL escanea toda la tabla (lento)
- Con índice: PostgreSQL va directo al registro (rápido)
- **Analogía:** es como el índice de un libro, te dice en qué página está cada tema

#### Trigger: `update_users_updated_at`

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Qué hace:**
- Cada vez que hacés UPDATE en users, automáticamente setea `updated_at = NOW()`
- No tenés que acordarte de hacerlo manualmente en el código

**Por qué:**
- Consistencia: todos los updates tienen timestamp correcto
- Menos código: no repetís `updated_at = NOW()` en cada query

---

### Tabla: `refresh_tokens`

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    family_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);
```

#### ¿Por qué necesitamos refresh tokens?

**Problema:** Los access tokens son de corta duración (15 min) por seguridad. Pero no queremos que el usuario haga login cada 15 minutos.

**Solución:** Refresh tokens
- **Access token:** 15 min, en memoria (JavaScript)
- **Refresh token:** 7 días, en httpOnly cookie + base de datos
- **Flujo:**
  1. Usuario hace login → recibe access token + refresh token
  2. Access token expira (15 min) → frontend usa refresh token para obtener nuevo access token
  3. Refresh token expira (7 días) → usuario debe hacer login de nuevo

#### Columnas explicadas

**`id UUID PRIMARY KEY DEFAULT gen_random_uuid()`**
- Igual que en users: identificador único

**`user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`**
- **REFERENCES users(id):** Foreign Key - este token pertenece a un usuario específico
- **ON DELETE CASCADE:** si se borra el usuario, se borran todos sus tokens automáticamente
- **Por qué:** integridad referencial - no pueden existir tokens de usuarios que no existen

**`token_hash VARCHAR(255) NOT NULL UNIQUE`**
- **Por qué "hash" y no el token en texto plano:** seguridad
- **UNIQUE:** cada token es único
- **Contenido:** hash SHA-256 del refresh token original
- **Flujo:**
  1. Backend genera refresh token (string aleatorio)
  2. Backend hashea el token con SHA-256
  3. Backend guarda el hash en la DB
  4. Backend envía el token original al cliente (en cookie)
  5. Cuando el cliente usa el token, backend lo hashea y busca el hash en la DB

**`family_id UUID NOT NULL`**
- **Qué es:** identificador de "familia" de tokens
- **Por qué:** para detectar robo de tokens (family-based rotation)
- **Cómo funciona:**
  1. Login → genera Token A con family_id = 123
  2. Refresh → genera Token B con family_id = 123, Token A se marca como revoked
  3. Refresh → genera Token C con family_id = 123, Token B se marca como revoked
  4. Si alguien usa Token A (robado) → detecta que ya está revoked → revoca TODA la familia 123
  5. Usuario legítimo no puede usar Token C → debe hacer login de nuevo

**`expires_at TIMESTAMPTZ NOT NULL`**
- **Cuándo expira el token** (7 días desde creación)
- **NOT NULL:** siempre debe tener fecha de expiración
- **Uso:** el sistema puede borrar tokens expirados periódicamente

**`revoked BOOLEAN NOT NULL DEFAULT FALSE`**
- **FALSE:** token válido, se puede usar
- **TRUE:** token revocado, no se puede usar
- **Cuándo se revoca:**
  - Usuario hace logout
  - Se detecta reuso de token (posible robo)
  - Se revoca toda la familia

**`created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`**
- Cuándo se creó el token

#### Índices

```sql
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);
```

**Por qué estos 3 índices:**

1. **user_id:** buscar todos los tokens de un usuario (ej: logout de todos los dispositivos)
2. **token_hash:** buscar un token específico (cada vez que se usa un refresh token)
3. **family_id:** buscar todos los tokens de una familia (para revocar toda la familia)

---

## Backend: Estructura y Patrones

### Estructura de carpetas

```
backend/internal/
├── config/
│   └── config.go           # Carga variables de entorno
├── database/
│   └── postgres.go         # Pool de conexiones PostgreSQL
├── model/
│   ├── user.go             # Modelo User
│   ├── refresh_token.go    # Modelo RefreshToken
│   └── errors.go           # Sentinel errors
└── repository/
    ├── user_repository.go
    ├── user_repository_test.go
    ├── refresh_token_repository.go
    └── refresh_token_repository_test.go
```

**Por qué esta estructura:**
- **Separación de responsabilidades:** cada carpeta tiene un propósito claro
- **Convención de Go:** `internal/` significa que no es accesible desde fuera del módulo
- **Tests junto al código:** `foo_test.go` al lado de `foo.go`

### Patrón: Repository Pattern

**Qué es:**
- Abstrae el acceso a datos detrás de una interfaz
- El service no sabe si usa PostgreSQL, MongoDB, o un mock

**Implementación:**

```go
// Interfaz (contrato)
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// Implementación concreta (PostgreSQL)
type postgresUserRepository struct {
    pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
    return &postgresUserRepository{pool: pool}
}
```

**Ventajas:**
- **Testeable:** podés mockear el repository en tests del service
- **Flexible:** podés cambiar la DB sin tocar el service
- **Claro:** la interfaz documenta qué operaciones están disponibles

### Patrón: Sentinel Errors

**Qué es:**
- Errores predefinidos que representan condiciones específicas
- Se comparan con `errors.Is()`

**Implementación:**

```go
// model/errors.go
var (
    ErrEmailExists        = errors.New("email already exists")
    ErrUserNotFound       = errors.New("user not found")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrTokenExpired       = errors.New("token expired")
    ErrTokenRevoked       = errors.New("token revoked")
    ErrTokenNotFound      = errors.New("token not found")
)
```

**Uso:**

```go
// En repository
func (r *postgresUserRepository) Create(ctx context.Context, user *model.User) error {
    // ... query ...
    if isDuplicateKeyError(err) {
        return model.ErrEmailExists
    }
    return err
}

// En handler (después)
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    err := h.service.Register(r.Context(), req)
    if errors.Is(err, model.ErrEmailExists) {
        http.Error(w, "email already exists", http.StatusConflict)
        return
    }
}
```

**Ventajas:**
- **Tipado:** el compilador verifica que uses errores válidos
- **Composable:** podés wrappear errores sin perder el tipo original
- **Claro:** el nombre del error documenta qué pasó

---

## Testing

### Estrategia de testing

**Tests de repository:**
- Usan base de datos real (PostgreSQL en Docker)
- Cada test limpia las tablas antes de correr (aislamiento)
- Testean casos de éxito y error

**Ejemplo:**

```go
func TestUserRepository_Create(t *testing.T) {
    pool := testPool(t)  // conexión a DB de test
    repo := NewUserRepository(pool)

    t.Run("success - creates user", func(t *testing.T) {
        user := &model.User{
            Email:        "test@example.com",
            PasswordHash: "hash123",
        }
        err := repo.Create(context.Background(), user)
        require.NoError(t, err)
        assert.NotEmpty(t, user.ID)
    })

    t.Run("duplicate email - returns ErrEmailExists", func(t *testing.T) {
        // Insertar primer usuario
        user1 := &model.User{Email: "dup@example.com", PasswordHash: "hash"}
        repo.Create(context.Background(), user1)

        // Intentar insertar duplicado
        user2 := &model.User{Email: "dup@example.com", PasswordHash: "hash"}
        err := repo.Create(context.Background(), user2)
        require.Error(t, err)
        assert.True(t, errors.Is(err, model.ErrEmailExists))
    })
}
```

### Base de datos de test

**Configuración:**
- Variable de entorno: `TEST_DATABASE_URL`
- Base de datos separada: `billetera_test`
- Migraciones aplicadas antes de correr tests

**Aislamiento:**
```go
func testPool(t *testing.T) *pgxpool.Pool {
    // ... conexión ...
    
    // Limpiar tablas antes de cada test
    pool.Exec(ctx, "DELETE FROM refresh_tokens")
    pool.Exec(ctx, "DELETE FROM users")
    
    return pool
}
```

**Por qué:**
- Cada test empieza con tablas vacías
- Tests no interfieren entre sí
- Podés correr tests en paralelo (después)

---

## Conceptos Aprendidos

### 1. Migraciones con golang-migrate

**Qué es:**
- Sistema para versionar cambios en la base de datos
- Cada migración tiene `.up.sql` (aplicar) y `.down.sql` (revertir)

**Nomenclatura:**
```
000001_baseline.up.sql
000001_baseline.down.sql
000002_create_users.up.sql
000002_create_users.down.sql
```

**Comandos:**
```bash
make migrate-up      # aplicar todas las migraciones pendientes
make migrate-down    # revertir la última migración
```

**Tabla `schema_migrations`:**
- Creada automáticamente por golang-migrate
- Trackea qué migraciones ya se aplicaron
- Columnas: `version` (última migración), `dirty` (si falló a la mitad)

### 2. pgxpool (pool de conexiones)

**Qué es:**
- Pool de conexiones reutilizables a PostgreSQL
- Evita abrir/cerrar conexión en cada query

**Configuración:**
```go
config, _ := pgxpool.ParseConfig(databaseURL)
config.MaxConns = 25              // máximo 25 conexiones
config.MinConns = 5               // mínimo 5 conexiones
config.MaxConnLifetime = time.Hour // cerrar después de 1 hora
config.MaxConnIdleTime = 30 * time.Minute // cerrar si está idle 30 min

pool, _ := pgxpool.NewWithConfig(ctx, config)
```

**Por qué usar pool:**
- **Performance:** reutilizar conexiones es más rápido que abrir nuevas
- **Límites:** evita saturar PostgreSQL con demasiadas conexiones
- **Health checks:** detecta conexiones muertas y las reemplaza

### 3. UUIDs vs INTs

**UUIDs (lo que usamos):**
- ✅ No revela cantidad de registros
- ✅ Se puede generar en el cliente
- ✅ Más seguro para APIs
- ❌ Más largo (36 caracteres vs 1-10 dígitos)
- ❌ Más difícil de debuggear

**INTs (alternativa):**
- ✅ Más corto y legible
- ✅ Más rápido para índices
- ❌ Revela cantidad de registros
- ❌ Se puede adivinar (1, 2, 3...)
- ❌ Necesita auto-increment (depende de la DB)

**Decisión:** UUIDs por seguridad y flexibilidad.

### 4. Foreign Keys y CASCADE

**Qué es:**
- Foreign Key: columna que referencia a otra tabla
- CASCADE: acción automática cuando se borra/actualiza el registro referenciado

**Ejemplo:**
```sql
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
```

**Qué significa:**
- `user_id` debe existir en `users.id` (integridad referencial)
- Si se borra el usuario, se borran todos sus refresh tokens automáticamente

**Otros opciones:**
- `ON DELETE SET NULL`: setea la columna a NULL
- `ON DELETE RESTRICT`: impide borrar el usuario si tiene tokens
- `ON DELETE NO ACTION`: igual que RESTRICT pero se verifica al final de la transacción

### 5. Índices en PostgreSQL

**Qué es:**
- Estructura de datos que acelera búsquedas
- Analogía: índice de un libro

**Tipos:**
- **B-tree (default):** bueno para igualdad y rangos
- **Hash:** solo igualdad
- **GIN:** para arrays, JSONB, full-text search
- **GiST:** para datos geométricos

**Cuándo crear índices:**
- ✅ Columnas que usás en WHERE frecuentemente
- ✅ Foreign Keys (joins)
- ✅ Columnas con UNIQUE constraint
- ❌ Columnas que casi nunca filtrás
- ❌ Tablas muy pequeñas (full scan es más rápido)

**Trade-off:**
- Índices aceleran lecturas (SELECT)
- Índices ralentizan escrituras (INSERT, UPDATE, DELETE)
- Más índices = más espacio en disco

### 6. Triggers en PostgreSQL

**Qué es:**
- Función que se ejecuta automáticamente antes/después de un evento (INSERT, UPDATE, DELETE)

**Ejemplo (el que usamos):**
```sql
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Qué hace:**
- Antes de cada UPDATE en users, ejecuta la función
- La función setea `NEW.updated_at = NOW()`

**Otros usos:**
- Auditoría (log de cambios)
- Validación compleja
- Sincronización entre tablas

### 7. Context en Go

**Qué es:**
- Objeto que lleva deadlines, cancellation signals, y valores a través de la call stack
- Primer parámetro de casi todas las funciones en Go moderno

**Uso:**
```go
func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
    query := "SELECT ... FROM users WHERE email = $1"
    err := r.pool.QueryRow(ctx, query, email).Scan(...)
    // ...
}
```

**Por qué:**
- **Timeouts:** si el request HTTP expira, cancela la query
- **Cancellation:** si el cliente cierra la conexión, cancela el trabajo
- **Tracing:** lleva información de tracing (OpenTelemetry)

**Ejemplo de timeout:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := repo.FindByEmail(ctx, "test@example.com")
// Si tarda más de 5 segundos, ctx se cancela y la query falla
```

---

## Decisiones de Diseño

### 1. ¿Por qué bcrypt para passwords?

**Alternativas:**
- **bcrypt:** lento por diseño, resistente a GPU attacks
- **scrypt:** más resistente a ASIC attacks, pero menos adoptado
- **argon2:** ganador de Password Hashing Competition, pero más complejo

**Decisión:** bcrypt con cost factor 12
- **Por qué:** estándar de la industria, bien soportado en Go
- **Cost 12:** ~250ms por hash en hardware moderno (suficientemente lento para ataques)

### 2. ¿Por qué family-based token rotation?

**Alternativas:**
- **Sin rotación:** refresh token válido hasta que expira (inseguro)
- **Rotación simple:** cada refresh genera nuevo token, invalida el anterior
- **Family-based:** detecta robo de tokens (lo que usamos)

**Decisión:** family-based rotation
- **Por qué:** detecta robo de tokens automáticamente
- **Trade-off:** más complejo, pero más seguro

### 3. ¿Por qué httpOnly cookies para refresh tokens?

**Alternativas:**
- **localStorage:** vulnerable a XSS
- **Session storage:** vulnerable a XSS
- **httpOnly cookie:** no accesible desde JavaScript (seguro)

**Decisión:** httpOnly cookie + SameSite=Strict
- **Por qué:** protege contra XSS y CSRF
- **Trade-off:** más complejo de implementar, pero más seguro

### 4. ¿Por qué access tokens en memoria?

**Alternativas:**
- **localStorage:** vulnerable a XSS
- **httpOnly cookie:** vulnerable a CSRF (necesita tokens CSRF)
- **En memoria (JavaScript):** se pierde al refrescar página

**Decisión:** en memoria + Zustand con sessionStorage
- **Por qué:** protege contra XSS
- **Trade-off:** se pierde al refrescar, pero se recupera con refresh token

---

## Próximos Pasos (PR 2)

**Qué vamos a implementar:**
- AuthService (lógica de autenticación)
  - JWT generation/validation
  - Password hashing con bcrypt
  - Register, Login, Refresh, Logout
- AuthHandler (5 endpoints HTTP)
- Auth middleware (validación de JWT)
- DTOs (request/response)
- main.go actualizado (dependency injection)
- router.go actualizado (rutas de auth)
- Integration tests

**Qué vamos a aprender:**
- JWT (JSON Web Tokens): estructura, firma, validación
- bcrypt: hashing, verificación, cost factor
- Middleware en Go: cómo interceptar requests
- Dependency Injection: cómo conectar las capas
- Integration testing: cómo testear todo el flujo

---

## Recursos Adicionales

**PostgreSQL:**
- [UUID vs INT debate](https://www.cockroachlabs.com/blog/what-is-a-uuid/)
- [Indexes explained](https://use-the-index-luke.com/)
- [Foreign Keys](https://www.postgresql.org/docs/current/tutorial-fk.html)

**Go:**
- [Context documentation](https://pkg.go.dev/context)
- [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool)
- [Errors package](https://pkg.go.dev/errors)

**Seguridad:**
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [JWT best practices](https://tools.ietf.org/html/rfc8725)
- [Refresh token rotation](https://www.ietf.org/archive/id/draft-ietf-oauth-refresh-tokens-00.html)
