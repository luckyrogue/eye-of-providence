# Alpha — instalación de clientes

Documento para participantes de la alpha (v0.1.x). Contacte al maintainer si
necesita acceso a los artefactos (el GH Release sigue en borrador).

## Requisitos previos

1. Cuenta en el panel (https://eop.rysdavletov.org). Registro por
   email/código de acceso.
2. Un cliente compatible instalado (véase abajo).
3. Código de emparejamiento de 6 caracteres del panel (`Settings → Devices →
   Claim device`) — válido 10 minutos; genere uno nuevo al caducar.

## 1. Extensión del navegador (Chrome / Edge / Brave)

**Origen:** `eop-browser-extension.zip` del GH Release (borrador).

```bash
# Descomprimir
unzip eop-browser-extension.zip -d eop-browser-extension/
```

1. Abra `chrome://extensions/`.
2. Active **Modo de desarrollador** (esquina superior derecha).
3. **Cargar descomprimida** → seleccione la carpeta `eop-browser-extension/`.
4. Fije el icono en la barra de herramientas.
5. Pulse el icono → **Start pairing**.
6. Se abre `Settings → Devices` del panel; el código se copia al portapapeles.
7. Pegue el código y pulse `Claim`. La extensión muestra `paired`.

**Comprobación:** abra ChatGPT / Claude / Gemini y escriba algo. Tras ~30 s
pulse el popup — debe aparecer `Pending: 1+`. Los eventos llegan al panel en 1–2 minutos.

## 2. Agente (macOS / Windows / Linux)

**Origen:** `.dmg` / `.msi` / `.AppImage` del GH Release (borrador).

### macOS

```bash
# DMG → arrastre .app a /Applications
open ~/Downloads/eop-agent_*.dmg
```

Primer arranque:

1. **Ajustes del sistema → Privacidad y seguridad → Accesibilidad** → active EoP
   (necesario para ventana activa).
2. Abra EoP → pestaña **Settings** → **Start pairing**.
3. Abra el enlace del panel desde la tarjeta e introduzca el código de 6 caracteres.
4. Tras `paired`, el agente empieza a registrar eventos.

Registros: `Settings → Open logs folder` → `~/Library/Application Support/com.eop.agent/logs/`.

### Windows

1. Ejecute `.msi`, acepte la advertencia de SmartScreen (compilación sin firmar).
2. Tras la instalación arranca desde la bandeja.
3. Abra desde la bandeja → pestaña **Settings** → **Start pairing**.

### Linux

```bash
chmod +x eop-agent_*.AppImage
./eop-agent_*.AppImage
```

`.AppImage` autocontenido, sin instalación. Para inicio automático véase
**Settings → Autostart**.

## 3. Extensión VS Code

**Origen:** `eop-vscode.vsix` del GH Release.

```bash
code --install-extension eop-vscode.vsix
```

o en VS Code: **Extensions** → `…` → **Install from VSIX…**.

Tras instalar:

1. Se abre la ventana de bienvenida.
2. Cmd/Ctrl+Shift+P → **EoP: Pair this editor**.
3. El panel se abre en el navegador; el código se copia al portapapeles.
4. Pegue el código en Devices → Claim.
5. La barra de estado muestra `$(eye) EoP idle` — extensión lista.

**Qué registramos:** entrada manual de caracteres en el editor (`chars_in`),
conteos de pulsaciones, líneas pegadas por IA (diffs del búfer).

**Qué no registramos:** contenido de archivos, nombres de archivos ni código.

## Solución de problemas

| Síntoma | Qué hacer |
| --- | --- |
| `auth required` en popup/barra | Token revocado. Settings → Devices → revocar y crear nuevo. |
| `Pending: N+` no baja | Backend no disponible o token caducado. Revise `Open logs` o salud del panel. |
| Código de emparejamiento caducado | Genere uno nuevo — botón `Claim device` en el panel. |
| macOS: sin eventos | Revise permiso de Accesibilidad. Sin él el agente solo ve bundle.id, no pulsaciones. |
| Linux/AppImage no abre | Instale `libwebkit2gtk-4.1`, `libgtk-3` (véase release notes). |

## Privacidad

- Sin URLs, títulos de página, nombres de archivo ni código — solo
  `app_bundle` (p. ej. `chat.openai.com`), categoría (`ai` / `manual`)
  y contadores (`chars_in`, `duration_ms`).
- El token de emparejamiento va al llavero del SO (macOS Keychain / Windows Credential
  Manager / GNOME keyring), no en texto plano.
- Esquema completo: [/docs/data-model](/docs/data-model).
- Modelo de amenazas: [/security](/security).

## Enviar feedback

GitHub Issues → etiqueta `alpha-feedback`. Indique:

- Cliente + versión (`Settings → About` o `vsce show`).
- SO + versión.
- Pasos para reproducir.
- Contenido de logs (sin datos personales identificables).
