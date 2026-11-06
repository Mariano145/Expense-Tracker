# State — dónde está cada feature

> Tablero de una mirada. Se actualiza al cerrar/comenzar cada historia grande.
> El detalle de QUÉ construir vive en los issues (milestone = feature); esto es
> solo el estado y el siguiente paso. `go test -p 1 ./...` / `npm test` en verde
> es el estado real de la última verificación.

| Feature (milestone) | Estado | Última verificación | Siguiente paso |
|---|---|---|---|
| `authentication` | ✅ Implementada en `feature/authentication` — falta QA manual + merge a `develop` | backend 7/7 · frontend 45/45 (2026-10-31) | QA del flujo completo → merge → cerrar milestone |
| `expenses` | ⬜ No arrancada | — | `/backlog` del dominio → historias |
| `categories` | ⬜ No arrancada | — | después de expenses |
| `dashboard` | ⬜ No arrancada | — | después de categories |
| `profile` | ⬜ No arrancada | — | backlog futuro |

## Bugs/fixes abiertos (fuera de milestones)

- Formularios login/register muestran error crudo de axios en vez del `error.message` de la API (UX 400) — primera historia candidata.
- Hallazgos judgment-day pendientes de la validación (B-003: regla de password solo en handler; B-004: fronteras 7/8 chars sin test) — ver sesión 2026-10-31.
