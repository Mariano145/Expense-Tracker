# QA manual — authentication

> Registro de la pasada manual sobre `feature/authentication`. Formato libre:
> anotá lo que veas, sin filtrar. Al terminar se triagea en una sola pasada y
> lo que corresponde se convierte en issue/historia.
>
> Grabar: qué hicieron falta para reproducirlo + qué salió vs qué esperaba.
> Un "OK" en una tilde también sirve (es la evidencia de que pasó).

## Pasada

- Fecha:
- Branch / commit: `feature/authentication @ ____`
- Backend: `go run ./cmd/server` · Frontend: `npm run dev`
- Usuario de prueba: (email, o "existente")

## Errores encontrados

### E1 — login
- Paso a paso: /login logueo con usuario no existente
- Esperaba: cartel que diga 'usuario no registrado'
- Vi: cartel con 'Request failed with status code 401'
- Severidad sospechada: UI (back envia error y front no formatea el texto)

### E2 — login
- Paso a paso: /login logueo correcto
- Esperaba: el accesToken no aparece en localstorage
- Vi: el accestoken aparece en localstorage
- Severidad sospechada: duda, no estoy seguro de como funciona todavia el token y donde deberia aparecer

### E3 — sign in
- Paso a paso: /register con mail ya registrado
- Esperaba: cartel que diga 'Mail en uso'
- Vi: cartel con codigo 'Request failed with status code 409'
- Severidad sospechada: UI igual que E1

### E4 — sign in
- Paso a paso: /register con password de < 7 digitos
- Esperaba: cartel que diga 'La contraseña debe contener como minimo 8 caracteres'
- Vi: cartel con codigo 'Request failed with status code 400'
- Severidad sospechada: UI igual que E1



## Tildes del checklist (OK)

- [ ] Register crea user + deja logueado
- [ ] Login devuelve access 15m + refresh 7d en cookie httpOnly (sin token en localStorage)
- [ ] `/dashboard` protegido: deslogueado → `/login`
- [ ] `GET /me` muestra el user real (id correcto)
- [ ] Reload con access expirado → refresh + rotación de refresh token, no expulsa
- [ ] Logout invalida refresh (recargar → vuelve al login)
- [ ] Login con password inválida → mensaje
- [ ] Register con email duplicado → mensaje
- [ ] Register con password de 7 chars → mensaje (borde B-004)
- [ ] Password de 8 chars → acepta

## Pendientes de judgment-day a confirmar acá

- **B-003**: la regla del password vive solo en el handler — ¿alguien puede crear un user con password corta por API directa (curl)?
- **B-004**: fronteras 7/8 chars sin test.
