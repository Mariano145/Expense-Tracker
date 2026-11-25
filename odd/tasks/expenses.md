# expenses: registrar y revisar gastos propios

**Intent (verbatim, the user's words):** "quiero avanzar con mas en el desarrollo de la app" · "mergea a develop entonces y vamos a crear la rama expenses" · "Sembrar defaults en expenses"
**Branch:** `feature/expenses` (base: `develop` @ `c9967e3`)   **Issues:** S1 a crear · milestone `expenses` (abierto, 0 issues)   **Status:** in-progress

## Problem & scope

Hoy la app autentica pero no registra nada: no existe la tabla de gastos. Esta
feature entrega el ciclo completo: crear un gasto, verlo en una lista, filtrarlo,
editarlo y borrarlo — con categorías reales desde el primer demo.

OUT de scope: dashboard y stats (milestone `dashboard`), CRUD de categorías
(listar/crear/editar/borrar categorías propias → milestone `categories`),
multi-moneda, importación desde planilla, edición en lote.

## Constraints

- **Dinero: BIGINT de centavos.** 123456 = $1.234,56. Decidido 2026-11-06 por el
  PO para que las sumas del dashboard sean exactas. Prohibido float8 y NUMERIC.
- **Fecha: `DATE` sin hora.** Un gasto es un día, no un instante. Evita que
  agrupar por mes dependa de la zona horaria del server. `created_at` sigue
  siendo `timestamptz` para auditoría.
- **DELETE duro.** Sin `deleted_at`: cada query futuro y cada agregación tendría
  que recordar el filtro. Decidido 2026-11-06.
- **Multi-tenant por `user_id` en cada tabla.** Ningún query sin filtro de dueño.
  Es AC explícito del milestone (historia S8).
- **El seed de las 7 categorías vive acá, no en `categories`.** El AC de expenses
  pide `category_id` y sin categorías la primera historia no es demosable sola.
- Pruebas: `go test -p 1 ./...` en backend (el `-p 1` es obligatorio, comparte
  `spendly_test`), `npm test` en frontend. Ver `docs/setup.md`.
- Commits: Conventional, ~400 líneas, test+doc con el código que cubren.

## Stories / tasks

| ID | Story (issue #) | Criterio | Status | Evidence |
|----|-----------------|----------|--------|----------|
| S1 | Cada usuario nuevo recibe sus 7 categorías por defecto | pendiente issue | [ ] | — |
| S2 | Registrar un gasto eligiendo categoría | requiere S1 | [ ] | — |
| S3 | Ver mis gastos del mes | requiere S2 | [ ] | — |
| S4 | Cambiar de mes desde la lista | requiere S3 | [ ] | — |
| S5 | Filtrar mis gastos por categoría | requiere S3 | [ ] | — |
| S6 | Editar un gasto desde la lista | requiere S3 | [ ] | — |
| S7 | Borrar un gasto | requiere S3 | [ ] | — |
| S8 | Nadie ve ni toca gastos de otro usuario | requiere S2, S6, S7 | [ ] | — |
| S9 | Monto legible: miles y decimales | requiere S2 | [ ] | — |
| S10 | La categoría lleva su color en la lista | requiere S1, S3 | [ ] | — |
| S11 | Buscar gastos por descripción | requiere S3 | [ ] | — |
| S12 | Lista paginada cuando crece | requiere S3 | [ ] | — |

Las ACs literales viven en cada issue; esta tabla no las duplica.

## Verification log

- 2026-11-06 · `go vet ./...` limpio · `go test -p 1 ./...` 7 paquetes OK · `npm test` 45/45 en 7 archivos — todo sobre el código de auth que subió al PR #8.
- (vacío para expenses todavía)

## Decisions & changes to intent

- 2026-11-06 — Orden del board: el seed de categorías **sale** del milestone
  `categories` y entra en S1 de expenses. Motivo: el AC de expenses pide
  `category_id`; sin eso S2 no demoa sola. `categories` queda reducido a
  listar/crear/editar/borrar categorías propias.
- 2026-11-06 — Schema: centavos `BIGINT` (descartado float8 y NUMERIC), `DATE`
  (descartado timestamptz para el día del gasto), `DELETE` duro (descartado
  soft delete). Las tres las decidió el PO en la misma pasada.
- 2026-11-06 — `feature/expenses` nace de `develop` con auth ya integrada, no
  stackeada sobre `feature/authentication`.

## Next step

Refinar S1 (`/story`): Given/When/Then + plan + out-of-scope, abrir el issue en
el milestone `expenses`, y arrancar el RED del primer test del repositorio de
categorías. Nada de expenses se toca sin que auth's #10 (token storage) quede
decidido, porque S1 toca `Register` en el service.
