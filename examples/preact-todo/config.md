# Build Configuration

This file defines all build configuration files for the Preact + Bun + Tailwind project.
Each chunk uses a `file:` attribute to specify the output path during tangling.

## Files generated

- **`package.json`** — Bun project manifest with `preact`, `tailwindcss`, and `typescript` dependencies
- **`tsconfig.json`** — TypeScript configuration with `jsxImportSource` set to `"preact"` for JSX support
- **`tailwind.config.js`** — Tailwind CSS configuration with custom `primary` color palette
- **`index.html`** — Minimal HTML shell with Tailwind CDN, app mount point, and module script entry

## Goweb features shown

**`{{var}}` substitution.** Placeholders like `{{APP_NAME}}` and `{{AUTHOR}}` are replaced
at tangling/render time via `--var APP_NAME="Todo App" --var AUTHOR="You"`.
This allows a single template to produce customized output for different projects
without editing the source.

The `{{var}}` syntax supports:
- Any key-value pair passed via `--var key=value`
- Unset variables are left as-is (shown as `{{NAME}}` in output)
- Variables are expanded in both prose and chunk body code

**The `tags:` attribute at scale.** These config chunks are all tagged `tags: config`.
When running `goweb graph --filter config`, only the configuration chunks would be shown,
making it easy to understand the project's non-code dependencies.

**Mixed file types.** This example generates JSON (`package.json`, `tsconfig.json`),
JavaScript (`tailwind.config.js`), and HTML (`index.html`) — all from a single
markdown file. Goweb's tangling is format-agnostic: it simply writes chunk bodies
to the specified `file:` path.

**Language inference for rendering.** During `goweb render`, each chunk's language
is inferred from its file extension:
- `.json` → syntax highlighted as JSON
- `.js`, `.jsx` → syntax highlighted as JavaScript/JSX
- `.html` → syntax highlighted as HTML

This gives the rendered documentation proper syntax highlighting via highlight.js.

<<package-json>>= file: package.json tags: config
{
  "name": "{{APP_NAME}}",
  "version": "1.0.0",
  "author": "{{AUTHOR}}",
  "scripts": {
    "dev": "bun run --hot src/main.tsx",
    "build": "bun build src/main.tsx --outdir dist",
    "preview": "bun run dist/index.html"
  },
  "dependencies": {
    "preact": "^10.26.0"
  },
  "devDependencies": {
    "tailwindcss": "^3.4.0",
    "typescript": "^5.6.0",
    "@types/node": "^22.0.0"
  }
}
>>

<<tsconfig>>= file: tsconfig.json tags: config
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "react-jsx",
    "jsxImportSource": "preact",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "outDir": "dist",
    "rootDir": "src",
    "declaration": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
>>

<<tailwind-config>>= file: tailwind.config.js tags: config
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{tsx,ts}", "./index.html"],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#eff6ff",
          100: "#dbeafe",
          200: "#bfdbfe",
          300: "#93c5fd",
          400: "#60a5fa",
          500: "#3b82f6",
          600: "#2563eb",
          700: "#1d4ed8",
          800: "#1e40af",
          900: "#1e3a8a",
        },
      },
    },
  },
  plugins: [],
};
>>

<<index-html>>= file: index.html tags: config
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{APP_NAME}}</title>
  <link rel="stylesheet" href="src/style.css">
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50 min-h-screen">
  <div id="app"></div>
  <script type="module" src="src/main.tsx"></script>
</body>
</html>
>>
