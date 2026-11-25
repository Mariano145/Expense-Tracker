# State — dónde está cada feature

> Tablero de una mirada. Se actualiza al cerrar/comenzar cada historia grande.
> El detalle de QUÉ construir vive en los issues (milestone = feature); esto es
> solo el estado y el siguiente paso. `go test -p 1 ./...` / `npm test` en verde
> es el estado real de la última verificación.

| Feature (milestone) | Estado | Última verificación | Siguiente paso |
|---|---|---|---|
| `authentication` | 🟡 Mergeada a `develop` (PR #8, `c9967e3`) **con deuda abierta** — milestone sigue abierto | backend 7 paquetes OK · frontend 45/45 (2026-11-06, sobre el código que subió) | cerrar #9 y #10 → QA del flujo refresh/logout → cerrar milestone |
| `expenses` | 🔵 Arrancada: rama `feature/expenses` sobre `develop` | — | `/backlog` del dominio → slice de historias → primera historia |
| `categories` | ⬜ No arrancada | — | después de expenses; el seed de defaults se muda a expenses |
| `dashboard` | ⬜ No arrancada | — | después de categories |
| `profile` | ⬜ No arrancada | — | backlog futuro |

## Deuda de `authentication` — ya tiene issue, no anotar de nuevo

- **#9** — Los formularios muestran el `AxiosError.message` en vez de `response.data.error.message`. Un solo bug raíz para los hallazgos E1 + E3 + E4 del QA. El backend ya manda el mensaje bueno (`handler/auth.go:181` `writeError`, `dto.go:28`).
- **#10** — El access token queda en `localStorage` (`auth-store.ts`: `persist` + `partialize` incluye `accessToken`), cuando `design.md:13` eligió in-memory + sessionStorage y descartó localStorage explícitamente. Desvío de decisión documentada: un XSS sobrevive al cierre del navegador.

## Deuda técnica sin issue todavía

- **B-003 (reclasificado)** — No es verificable desde QA: un curl pasa por el handler y dispara la validación igual. `service.Register` no valida nada; hoy no hay forma de llegar al service sin pasar por el handler. El riesgo es el próximo caller interno (el seed de categorías de expenses, `change-password` en `profile`). La regla del password vive solo en `handler/validation.go`.
- **B-004** — Fronteras de longitud de password (7 rechaza / 8 acepta) sin test. El QA manual tocó "< 7" pero no los bordes exactos.
- **Sospecha confirmada en código, no en runtime** — El email se guarda crudo (`auth_service.go:56`, sin `ToLower`/trim) y se busca exacto (`user_repository.go:66`, `WHERE email = $1`). Register con `Admin@X.com` + login con `admin@x.com` debería fallar; y la misma dirección en dos cajas crear dos cuentas. Confirmar en runtime antes de abrir issue.

## Cómo correr el QA de auth desde otro equipo

El refresh cookie va con `Secure: true` (`backend/internal/handler/auth.go:130`): el navegador solo la acepta en origen seguro (`https` o `localhost`). Entrar por IP de red/Tailscale (`http://100.89.200.14:3000`) hace que el login *parezca* funcionar — el access token viaja en el JSON — pero la cookie de refresh se descarta en silencio: rotación y logout quedan invisibles.

Desde la notebook, tunelizar y abrir **`http://localhost:3000`** (nunca la IP):

```bash
ssh -N -L 3000:localhost:3000 -L 8080:localhost:8080 mariano145@100.89.200.14
```

Con `localhost` como origen la cookie entra, el `CORS_ORIGIN` del `.env` coincide y `api.ts` ya apunta a `localhost:8080` por default.

Tip: para ver la rotación sin esperar 15 min, `JWT_ACCESS_EXPIRY=30s` en `backend/.env` y reiniciar el server. Monitoreo en vivo:

```bash
watch -n2 'docker exec spendly-db psql -U spendly -d spendly -c "select left(family_id::text,8) as fam, revoked, expires_at from refresh_tokens order by created_at desc limit 8"'
```

## QA manual pendiente de auth

Registro en `odd/tasks/qa-authentication.md` (E1–E4 ya triageados arriba). Faltan las tildes de flujo: `/me` con user real, refresh + rotación, logout que revoque la familia, rutas protegidas, bordes 7/8 chars.

