# Seguridad — Eye of Providence

**Última actualización:** 2026-05-19 · **Versión:** v0.1 alpha

Esta página describe cómo manejamos la seguridad: qué hay implementado,
cómo reportar una vulnerabilidad, y qué esperar. Para el modelo de
amenazas interno de ingeniería, ver
[`docs/threat-model.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/threat-model.md);
esta página es el resumen público.

## Reportar una vulnerabilidad

**Por favor, no abras un issue público en GitHub.** Envía un correo a
**`main@rysdavletov.org`** con:

- Descripción del problema y su impacto
- Pasos de reproducción o PoC
- Versión o SHA del commit afectado
- Tu contacto para seguimiento

Objetivo — **acuse de recibo en 48 horas** y proporcionar timeline de
remediación en **5 días hábiles**.

En alcance: backend API, dashboard, agente (desktop / browser ext /
VS Code), imagen Docker e infraestructura CI.

Fuera de alcance: instancias self-hosted operadas por terceros;
servicios de terceros (ClickHouse Cloud, Resend, Dokploy); pruebas de
DoS (tenemos rate-limit, pero no es una superficie de explotación).

## Versiones soportadas

Los fixes de seguridad aterrizan en `main`. Durante alpha no mantenemos
ramas LTS.

| Versión              | Soportada |
| -------------------- | --------- |
| `main`               | ✅       |
| `v0.1.x-alpha.*`     | ✅ (solo latest) |
| tags rolling pre-alpha | ❌      |

## Postura de seguridad actual

Controles concretos en marcha hoy:

### Backend y datos

- **Auth:** bcrypt (cost 10), JWT HS256 con revocación vía
  `token_version`. Soporte WebAuthn / passkey para segundo factor.
- **Rate limits:** 10 req/min en endpoints de auth, 120 req/min en `/v1/*`.
- **Aislamiento por usuario:** cada consulta analítica se filtra por
  JWT subject; acceso cross-user imposible a nivel SQL.
- **GDPR:** `GET /v1/me/export` devuelve todos tus datos como JSON;
  `DELETE /v1/me/data` los elimina permanentemente.

### Imagen y cadena de suministro

- **Cosign signed:** cada imagen Docker subida a
  `ghcr.io/luckyrogue/eop` está firmada con Sigstore keyless OIDC.
  Verifica con:
  ```bash
  cosign verify ghcr.io/luckyrogue/eop:<sha> \
    --certificate-identity-regexp '^https://github.com/luckyrogue/eye-of-providence/' \
    --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
  ```
- **SLSA Build L3 provenance:** prueba verificable de que la imagen fue
  construida por nuestro CI desde un commit específico.
- **CycloneDX SBOM:** cada imagen publica un Software Bill of Materials
  como attestation.

Recetas completas:
[`.github/SECURITY.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/.github/SECURITY.md).

### Agente (desktop)

- **Búfer local cifrado:** los eventos en SQLite están cifrados con
  AES-256-GCM. La clave vive en el keyring del SO (macOS Keychain,
  Windows Credential Manager, GNOME keyring).
- **Tokens de pairing** permanecen en el keyring, nunca en texto plano.
- **Invariantes de privacidad:** el agente nunca lee contenido de
  archivos, prompts, respuestas de IA, pulsaciones de teclado raw, ni
  texto del portapapeles. Solo conteos, hashes y timestamps salen del
  dispositivo. Mapa de datos completo en [Política de privacidad](/privacy)
  §1.

### CI / desarrollo

- Análisis estático CodeQL (Go + JS/TS) en cada PR.
- Escaneos Trivy + OSV en source e imagen; dependency-review en PR.
- gitleaks escanea cada commit en busca de secretos accidentalmente
  comprometidos.
- Step-Security `harden-runner` audita el egress de los runners.

### Limitaciones conocidas (transparencia)

- **Los instaladores alpha no están firmados.** El Apple Developer ID
  y el cert EV de Windows aún no se han comprado. Ver
  [Why is this unsigned?](/docs/install#почему-installer-не-подписан)
  en la guía de instalación para workarounds. La firma de la imagen
  (Cosign) no se ve afectada — la imagen del backend es totalmente
  verificable.
- **Branch protection** se está activando según
  [`docs/ci-hardening.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/ci-hardening.md)
  como follow-up del alpha-1. Por ahora todos los merges pasan por CI
  pero los required reviewers no están enforced.
- **PITR para Postgres** no configurado; RPO = 24h (dump diario).
  Objetivo RPO más estricto para GA. Detalles en
  [`docs/disaster-recovery.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/disaster-recovery.md).

## Divulgación responsable

Actualmente no ofrecemos bug bounty. Damos crédito a investigadores
(con permiso) en la entrada CHANGELOG correspondiente y en el commit
message del fix.

Pedimos una ventana razonable — normalmente el timeline acordado en el
acuse inicial, default 90 días — antes de la divulgación pública.

## Actualizaciones de esta página

Los cambios materiales en la postura de seguridad se registran en
[`CHANGELOG.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/CHANGELOG.md)
bajo la sub-sección **Security** de cada release.
