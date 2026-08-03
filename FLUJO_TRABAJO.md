# Flujo de Trabajo - QA + Product Owner

> **Nota:** Este archivo es para referencia personal. NO se sube a git (está en .gitignore).

## Roles

- **Desarrollador**: IA (escribe código, crea PRs, arregla bugs)
- **QA + Product Owner**: Vos (verifica que todo funcione, reporta errores, aprueba PRs)

## Flujo General

```
1. IA crea branch y escribe código
   ↓
2. IA crea PR y te da instrucciones de verificación
   ↓
3. Vos corrés los comandos de verificación
   ↓
4. ¿Todo pasa?
   ├─ SÍ → Mergeás el PR
   └─ NO → Reportás el error a la IA, ella arregla, volvés al paso 3
   ↓
5. Después del merge → Probás en develop para verificar integración
```

## Tu Rol

**Qué hacés:**
- Corrés los comandos que te doy
- Verificás que todo funcione (tests pasan, servidor levanta, endpoints responden)
- Si algo falla, me reportás el error (copiás el output y me lo pasás)
- Si todo funciona, aprobás el PR y lo mergeás

**Qué NO hacés:**
- Escribir código
- Arreglar bugs
- Configurar cosas complejas

---

## Instrucciones de Verificación por PR

### PR 1: Foundation + Repository

**Comandos:**

```bash
# 1. Pull de la branch
git checkout feature/auth-pr1-foundation
git pull origin feature/auth-pr1-foundation

# 2. Levantar base de datos
make db-up

# 3. Aplicar migraciones
make migrate-up

# 4. Verificar que las tablas existen
docker compose exec postgres psql -U billetera -d billetera -c "\dt"

# 5. Correr tests
cd backend
go test ./internal/database/... -v
go test ./internal/repository/... -v

# 6. Verificar que compila
go build ./...
```

**Qué deberías ver:**

```bash
# Paso 4: Deberías ver 2 tablas
          List of relations
 Schema |      Name       | Type  |  Owner
--------+-----------------+-------+----------
 public | refresh_tokens  | table | billetera
 public | users           | table | billetera

# Paso 5: Deberías ver tests pasando
=== RUN   TestUserRepository_Create
--- PASS: TestUserRepository_Create (0.05s)
=== RUN   TestUserRepository_FindByEmail
--- PASS: TestUserRepository_FindByEmail (0.03s)
PASS
ok      github.com/gentle-ai/billetera/backend/internal/repository    0.123s

# Paso 6: Sin output (compilación exitosa)
```

**Si algo falla:** Copiás el error completo y me lo pasás.

---

### PR 2: Service + HTTP + Wiring

**Comandos:**

```bash
# 1. Pull
git checkout feature/auth-pr2-service
git pull origin feature/auth-pr2-service

# 2. Levantar DB
make db-up
make migrate-up

# 3. Correr todos los tests
cd backend
go test ./... -v

# 4. Levantar servidor
go run cmd/server/main.go

# 5. En OTRA terminal, probar endpoints
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 6. Verificar que /health sigue funcionando
curl http://localhost:8080/health
```

**Qué deberías ver:**

```bash
# Paso 3: Todos los tests pasando
ok      github.com/gentle-ai/billetera/backend/internal/database
ok      github.com/gentle-ai/billetera/backend/internal/repository
ok      github.com/gentle-ai/billetera/backend/internal/service
ok      github.com/gentle-ai/billetera/backend/internal/handler
ok      github.com/gentle-ai/billetera/backend/internal/middleware

# Paso 4: Servidor levantado
2026/07/27 15:30:45 Server starting on :8080

# Paso 5: Register exitoso
{"data":{"user":{"id":"uuid","email":"test@example.com","created_at":"..."}}}

# Paso 5: Login exitoso
{"data":{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...","user":{"id":"uuid","email":"test@example.com"}}}

# Paso 6: Health check
{"status":"ok"}
```

**Si algo falla:** Copiás el error y me lo pasás.

---

### PR 3: Frontend

**Comandos:**

```bash
# 1. Pull
git checkout feature/auth-pr3-frontend
git pull origin feature/auth-pr3-frontend

# 2. Instalar dependencias
cd frontend
npm install

# 3. Correr tests
npm test

# 4. Levantar backend (en otra terminal)
cd ../backend
make db-up && make migrate-up
go run cmd/server/main.go

# 5. Levantar frontend (en otra terminal)
cd frontend
npm run dev

# 6. Abrir navegador: http://localhost:3000
# 7. Probar:
#    - Ir a /login
#    - Registrarse con email + password
#    - Login
#    - Verificar que redirige a dashboard
#    - Logout
```

**Qué deberías ver:**

```bash
# Paso 3: Tests pasando
 ✓ src/store/authStore.test.ts (5 tests) 123ms
 ✓ src/app/(auth)/login/page.test.tsx (3 tests) 456ms
 ✓ src/app/(auth)/register/page.test.tsx (3 tests) 389ms

 Test Files  3 passed (3)
      Tests  11 passed (11)

# Paso 5: Frontend levantado
  ▲ Next.js 15.0.0
  - Local:        http://localhost:3000
  - Ready in 2.3s

# Paso 6-7: En el navegador
# - Página de login carga correctamente
# - Formulario de registro funciona
# - Después del login, redirige a /dashboard
# - Logout funciona y redirige a /login
```

**Si algo falla:** Screenshot del error + output de la terminal.

---

### PR 4: Documentation

**Comandos:**

```bash
# 1. Pull
git checkout feature/auth-pr4-docs
git pull origin feature/auth-pr4-docs

# 2. Leer README.md (verificar que esté actualizado)
cat README.md

# 3. Leer ARQUITECTURA.md (verificar que esté actualizado)
cat ARQUITECTURA.md

# 4. Verificar que todo sigue funcionando
make db-up && make migrate-up
cd backend && go test ./...
cd ../frontend && npm test
```

**Qué deberías ver:**
- README actualizado con instrucciones de auth
- ARQUITECTURA actualizado con nuevas carpetas
- Todos los tests siguen pasando

---

## Qué Hacer Si Algo Falla

### Bug menor (algo no funciona como esperabas)

**NO hagas rollback.** Reportame el error:

```
"El test TestUserRepository_FindByEmail falló con este error:
[pegás el output completo]"
```

Yo arreglo en la misma branch, pusheo, y vos corrés los comandos de nuevo.

### Bug crítico (todo está roto)

**Opción 1: Revertir el PR**

```bash
# Desde GitHub UI:
# 1. Entrá al PR mergeado
# 2. Click en "Revert" (abajo del PR)
# 3. Eso crea un nuevo PR que revierte todos los commits
```

**Opción 2: Revertir commits individuales**

```bash
git checkout develop
git revert <commit-hash>  # Revierte un commit específico
git push origin develop
```

### Todo roto (último recurso)

```bash
# Reset develop al estado anterior al PR
git checkout develop
git reset --hard <commit-before-pr>
git push origin develop --force  # ⚠️ CUIDADO: esto reescribe historia
```

**⚠️ Advertencia:** `--force` es peligroso si otros están trabajando en el repo. Pero como sos solo vos, está bien.

---

## Resumen

**Para cada PR:**
1. Pull de la branch
2. Corrés los comandos que te doy
3. Verificás que el output sea el esperado
4. Si todo pasa → mergeás
5. Si algo falla → me pasás el error

**No necesitás saber de código**, solo seguir las instrucciones y reportarme si algo no funciona.

---

## Estrategia de PRs (Autenticación)

**4 PRs con stacked-to-main:**

```
PR 1: Foundation + Repository (~410 líneas) → develop
PR 2: Service + HTTP + Wiring (~900 líneas) → develop
PR 3: Frontend (~530 líneas) → develop
PR 4: Documentation (~210 líneas) → develop
```

Cada PR mergea a `develop` en orden. Si algo falla, podés revertir PRs individuales.
