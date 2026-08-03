# 📚 Plan de Estudio - Billetera

Este documento mapea cada feature del proyecto con los conceptos que necesitás estudiar antes de implementarla.

## 🔄 Flujo de Trabajo

1. **Antes de cada PR:** Revisá los temas a estudiar en este documento
2. **Estudiá:** Usá los recursos proporcionados (links, videos, artículos)
3. **Confirmá:** Avisame cuando estés listo ("Ya estudié X, vamos con PR2")
4. **Implementamos:** Escribimos el código juntos, vos entendés qué y por qué
5. **Reflexioná:** Después de implementar, anotá qué aprendiste

---

## ✅ PR1: Foundation + Repository

**Estado:** ✅ Completado

### Conceptos estudiados:
- [x] PostgreSQL básico (tablas, queries, constraints)
- [x] Go modules y estructura de proyecto
- [x] Repository Pattern
- [x] Docker básico (compose, volumes)
- [x] Migraciones de base de datos

### Reflexiones post-implementación:
- Aprendí sobre UUIDs vs INTs para IDs
- Entendí por qué usamos interfaces en Go
- Descubrí los sentinel errors para manejo de errores

---

## 🎯 PR2: Autenticación (JWT + bcrypt)

**Estado:** ⏳ Pendiente

### Temas a estudiar antes de implementar:

#### 1. JWT (JSON Web Tokens) - 1 hora
**Qué entender:**
- Estructura: header, payload, signature
- Cómo se firma y verifica un token
- Claims estándar (exp, iat, sub)
- Por qué JWT es stateless

**Recursos:**
- [jwt.io/introduction](https://jwt.io/introduction) (30 min) - Leé la intro y probá el debugger
- [Auth0: Introduction to JWT](https://auth0.com/learn/json-web-tokens/) (30 min) - Más detallado

**Ejercicio práctico:**
- Andá a [jwt.io](https://jwt.io), decodificá un token de ejemplo
- Cambiá el payload y fijate cómo cambia la signature

#### 2. Password Hashing con bcrypt - 30 min
**Qué entender:**
- Por qué NO usar MD5/SHA1/SHA256 directamente
- Qué es el "salt" y por qué es importante
- Qué es el "cost factor" (work factor)
- Diferencia entre hashing y encryption

**Recursos:**
- [CrackStation: Hashing Security](https://crackstation.net/hashing-security.htm) (20 min) - Skim, enfocáte en "How Hashing Works" y "Why bcrypt"
- [Auth0: Hashing Passwords](https://auth0.com/blog/hashing-passwords-with-bcrypt/) (10 min) - Más práctico

**Ejercicio práctico:**
- Pensá: si dos usuarios tienen la misma password, ¿sus hashes son iguales? (respuesta: no, por el salt)

#### 3. HTTP Cookies - 30 min
**Qué entender:**
- Diferencia entre cookies y localStorage
- Atributos de seguridad: httpOnly, Secure, SameSite
- Por qué httpOnly previene XSS
- Qué es SameSite=Strict vs Lax

**Recursos:**
- [MDN: HTTP Cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies) (20 min) - Enfocáte en "Security" y "SameSite"
- [Web.dev: SameSite cookies](https://web.dev/samesite-cookies-explained/) (10 min) - Más claro que MDN

**Ejercicio práctico:**
- Abrí DevTools en tu navegador → Application → Cookies
- Fijate qué cookies tienen los sitios que usás (Gmail, GitHub, etc.)

#### 4. Refresh Token Rotation - 30 min
**Qué entender:**
- Por qué access tokens tienen vida corta (15 min)
- Por qué refresh tokens tienen vida larga (7 días)
- Qué es token rotation
- Qué es family-based rotation (detección de robo)

**Recursos:**
- [Auth0: Refresh Tokens](https://auth0.com/blog/refresh-tokens-what-are-they-and-when-to-use-them/) (20 min)
- [Okta: Token Rotation](https://developer.okta.com/docs/guides/refresh-tokens/main/#token-rotation) (10 min)

**Ejercicio práctico:**
- Pensá en el flujo: login → access token (15 min) → expira → refresh token → nuevo access token
- ¿Qué pasa si alguien roba el refresh token? (respuesta: family-based rotation lo detecta)

### Tiempo total estimado: 2.5 horas

### Nivel de profundidad:
- **Entender el "qué" y "por qué"**: sí
- **Entender el "cómo" (implementación)**: no necesariamente, eso lo vemos en el código

---

## 🎯 PR3: Frontend (Next.js + React)

**Estado:** ⏳ Pendiente

### Temas a estudiar antes de implementar:

#### 1. React Hooks básicos - 1 hora
**Qué entender:**
- useState: estado local de componentes
- useEffect: efectos secundarios (fetch data, subscriptions)
- useContext: compartir estado sin prop drilling
- Reglas de los hooks

**Recursos:**
- [React Official: Hooks](https://react.dev/reference/react/hooks) (30 min) - Leé useState y useEffect
- [React Official: Learn Hooks](https://react.dev/learn/hooks) (30 min) - Tutorial interactivo

#### 2. Next.js App Router - 1 hora
**Qué entender:**
- File-based routing (carpetas = rutas)
- Server Components vs Client Components
- Layouts y templates
- Data fetching en App Router

**Recursos:**
- [Next.js Learn: App Router](https://nextjs.org/learn/dashboard-app) (45 min) - Capítulos 1-3
- [Next.js Docs: App Router](https://nextjs.org/docs/app) (15 min) - Skim la intro

#### 3. Zustand (State Management) - 30 min
**Qué entender:**
- Por qué Zustand y no Redux/Context
- Cómo crear un store
- Cómo usar el store en componentes
- Persistencia de estado

**Recursos:**
- [Zustand Docs](https://docs.pmnd.rs/zustand/getting-started/introduction) (20 min) - Leé "Getting Started" y "Persistence"
- [Zustand vs Redux](https://docs.pmnd.rs/zustand/guides/comparison) (10 min) - Para entender por qué lo elegimos

### Tiempo total estimado: 2.5 horas

---

## 🎯 PR4: CRUD de Gastos

**Estado:** ⏳ Pendiente

### Temas a estudiar antes de implementar:

#### 1. REST API Design - 1 hora
**Qué entender:**
- Recursos y endpoints (nouns, no verbs)
- Métodos HTTP: GET, POST, PUT, PATCH, DELETE
- Status codes correctos (200, 201, 204, 400, 404)
- Paginación y filtrado

**Recursos:**
- [REST API Tutorial](https://restfulapi.net/) (45 min) - Enfocáte en "HTTP Methods" y "Status Codes"
- [Microsoft: REST API Design](https://learn.microsoft.com/en-us/azure/architecture/best-practices/api-design) (15 min) - Buenas prácticas

#### 2. Validación de Input - 30 min
**Qué entender:**
- Por qué validar en backend (no solo frontend)
- Tipos de validación: formato, rango, existencia
- Errores de validación: cómo reportarlos

**Recursos:**
- [OWASP: Input Validation](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html) (20 min) - Skim
- [Go Validator Package](https://github.com/go-playground/validator) (10 min) - Si usamos este package

### Tiempo total estimado: 1.5 horas

---

## 🎯 PR5: Dashboard y Reportes

**Estado:** ⏳ Pendiente

### Temas a estudiar antes de implementar:

#### 1. SQL Avanzado (Agregaciones) - 1 hora
**Qué entender:**
- GROUP BY y HAVING
- Funciones de agregación: SUM, COUNT, AVG, MIN, MAX
- Window functions (básico)
- Subqueries vs JOINs

**Recursos:**
- [PostgreSQL Tutorial: GROUP BY](https://www.postgresqltutorial.com/postgresql-group-by/) (20 min)
- [PostgreSQL Tutorial: Aggregate Functions](https://www.postgresqltutorial.com/postgresql-aggregate-functions/) (20 min)
- [PostgreSQL Window Functions](https://www.postgresqltutorial.com/postgresql-window-function/) (20 min) - Opcional

#### 2. Visualización de Datos - 30 min
**Qué entender:**
- Tipos de gráficos: barras, líneas, torta
- Cuándo usar cada tipo
- Librerías de gráficos en React (Recharts, Chart.js)

**Recursos:**
- [Recharts Docs](https://recharts.org/en-US/guide) (20 min) - Si usamos Recharts
- [Data Visualization Best Practices](https://www.tableau.com/learn/articles/data-visualization-best-practices) (10 min)

### Tiempo total estimado: 1.5 horas

---

## 🎯 PR6: Deploy a Producción

**Estado:** ⏳ Pendiente

### Temas a estudiar antes de implementar:

#### 1. Docker para Producción - 1 hora
**Qué entender:**
- Multi-stage builds (optimización de imágenes)
- Docker Compose para múltiples servicios
- Variables de entorno en producción
- Volúmenes y persistencia

**Recursos:**
- [Docker: Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/) (20 min)
- [Docker Compose Production](https://docs.docker.com/compose/production/) (20 min)
- [12 Factor App](https://12factor.net/) (20 min) - Principios de apps modernas

#### 2. Reverse Proxy (Nginx/Caddy) - 1 hora
**Qué entender:**
- Qué es un reverse proxy
- Por qué usarlo (SSL termination, load balancing)
- Configuración básica de Nginx o Caddy
- SSL/TLS y certificados (Let's Encrypt)

**Recursos:**
- [Nginx: Reverse Proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/) (30 min)
- [Caddy Docs](https://caddyserver.com/docs/) (15 min) - Alternativa más simple que Nginx
- [Let's Encrypt](https://letsencrypt.org/how-it-works/) (15 min) - Certificados SSL gratis

#### 3. Monitoreo y Logs - 30 min
**Qué entender:**
- Por qué monitorear (detección temprana de problemas)
- Logs estructurados (JSON)
- Herramientas básicas: logs de Docker, uptime monitoring

**Recursos:**
- [DigitalOcean: Application Monitoring](https://www.digitalocean.com/community/tutorials/an-introduction-to-application-monitoring) (20 min)
- [Structured Logging](https://stackify.com/what-is-structured-logging/) (10 min)

### Tiempo total estimado: 2.5 horas

---

## 📝 Notas Generales

### Cómo usar este documento:

1. **Antes de cada PR:**
   - Leé los temas a estudiar
   - Marcá los checkboxes a medida que completás cada recurso
   - Anotá dudas para preguntarme

2. **Durante la implementación:**
   - Referenciá este documento cuando algo no tenga sentido
   - Preguntá "por qué" si no entendés una decisión

3. **Después de cada PR:**
   - Anotá reflexiones en la sección correspondiente
   - Actualizá el estado (⏳ → ✅)

### Recursos adicionales (opcionales):

**Libros:**
- "Clean Code" de Robert Martin (capítulos 1-5, 10-11)
- "The Pragmatic Programmer" de Hunt & Thomas (selectivo)

**Cursos:**
- [Frontend Masters](https://frontendmasters.com/) (pago, pero excelente)
- [Pluralsight](https://www.pluralsight.com/) (pago, bueno para backend)

**YouTube:**
- [Fireship](https://www.youtube.com/@Fireship) - Videos cortos de 100 segundos
- [The Primeagen](https://www.youtube.com/@ThePrimeagen) - Go, backend, opiniones fuertes
- [Web Dev Simplified](https://www.youtube.com/@WebDevSimplified) - React, JS, CSS

### Progreso general:

- [x] PR1: Foundation + Repository ✅
- [ ] PR2: Autenticación ⏳
- [ ] PR3: Frontend ⏳
- [ ] PR4: CRUD de Gastos ⏳
- [ ] PR5: Dashboard y Reportes ⏳
- [ ] PR6: Deploy a Producción ⏳

**Progreso:** 1/6 features completadas (17%)
