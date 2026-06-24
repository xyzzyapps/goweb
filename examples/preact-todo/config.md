# Build Configuration

This file defines all build configuration files for the Preact + Bun + Tailwind project.
Each chunk uses a `file:` attribute to specify the output path during tangling.

The following files are generated:
- **`package.json`** — Bun project manifest with `preact`, `tailwindcss`, and `typescript` dependencies. Uses `{{APP_NAME}}` and `{{AUTHOR}}` variables.
- **`tsconfig.json`** — TypeScript configuration with `jsxImportSource` set to `"preact"` for JSX support.
- **`tailwind.config.js`** — Tailwind CSS configuration with custom `primary` color palette and content paths.
- **`index.html`** — Minimal HTML shell with Tailwind CDN, app mount point `<div id="app">`, and module script entry.

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
