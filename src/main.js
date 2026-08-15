import { render } from "preact";
import { html } from "htm/preact";
import { App } from "./app.js";

const root = document.getElementById("app");
if (root) {
  render(html`<${App} />`, root);
}
