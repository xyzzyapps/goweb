# The envelope

Before any component can run, the browser needs a page and a way to find
Preact. That is all this file is: a thin envelope. <<index-html>> is the
paper; the import map is the address book; <<package-json>> is a name tag
for humans and git, not a build ritual.

## Files generated

- **`index.html`** — HTML shell, import map (Preact + htm from esm.sh), app mount
- **`package.json`** — Optional metadata (`{{APP_NAME}}`, `{{AUTHOR}}`)

## Goweb features shown

**`{{var}}` substitution.** Placeholders like `{{APP_NAME}}` and `{{AUTHOR}}` are replaced
at tangling/render time via `--var APP_NAME="Todo App"`. `AUTHOR`, `EMAIL`, and `REPO` default from git config (`user.name`, `user.email`, `remote.origin.url`).

The import map is what makes `import { html } from "htm/preact"` work in the browser
without a build step. Scripts must be `.js` so the server sends `text/javascript`.

<<package-json>>= file: package.json tags: config
{
  "name": "{{APP_NAME}}",
  "version": "1.0.0",
  "author": "{{AUTHOR}}",
  "private": true,
  "type": "module",
  "description": "goweb literate example — Preact + htm, no bundler"
}
>>

<<index-html>>= file: index.html tags: config
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{APP_NAME}}</title>
  <meta name="description" content="{{APP_NAME}} — a goweb literate-programming example. Written with DeepSeek and Grok 4.6.">
  <link rel="stylesheet" href="src/style.css">
  <script type="importmap">
  {
    "imports": {
      "preact": "https://esm.sh/preact@10.26.4",
      "preact/hooks": "https://esm.sh/preact@10.26.4/hooks",
      "htm/preact": "https://esm.sh/htm@3.1.1/preact?external=preact"
    }
  }
  </script>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="src/main.js"></script>
</body>
</html>
>>
