// Markup of a project on the public site: its card in the projects grid and
// its case dialog. The site (main.js) and the admin preview (preview.js) both
// use it, so the preview always matches what visitors see.
import { t } from "./locales.js";

export const CATEGORIES = ["residential", "commercial"];

export function categoryOf(project) {
  return CATEGORIES.includes(project.category) ? project.category : "residential";
}

export function pad(n) {
  return String(n).padStart(2, "0");
}

export function esc(str) {
  return String(str ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

export function photoHTML(src, alt, extraClass = "") {
  if (!src) return `<div class="photo photo--empty ${extraClass}"></div>`;
  return `<div class="photo ${extraClass}"><img src="${src}" alt="${esc(alt)}" loading="lazy"></div>`;
}

// Photos that fail to load show the empty placeholder instead of a broken image.
export function initPhotos(root) {
  root.querySelectorAll(".photo img").forEach((img) => {
    const markEmpty = () => img.closest(".photo")?.classList.add("photo--empty");
    if (img.complete && img.naturalWidth === 0) markEmpty();
    else img.addEventListener("error", markEmpty, { once: true });
  });
}

// Card in the projects grid. `number` is the position shown on the card (1-based).
export function projectCardHTML(p, { number, coverUrl }) {
  return `
    <article class="project-card">
      <a href="#/project/${p.id ?? "preview"}">
        ${photoHTML(coverUrl, p.title, "photo--zoom")}
        <div class="project-card__body">
          <div class="project-card__meta">
            <span>${t(`category_${categoryOf(p)}`)}</span>
            <span>${pad(number)}</span>
          </div>
          <h3 class="project-card__title">${esc(p.title)}</h3>
          ${p.description ? `<p class="project-card__desc">${esc(p.description)}</p>` : ""}
        </div>
        <div class="project-card__open"><span>${t("project_open")}</span><span aria-hidden="true">→</span></div>
      </a>
    </article>`;
}

// Case dialog: `meta` goes to the dialog bar, `body` to the dialog body.
// The cover comes first; `galleryUrls` are the other photos in order.
export function projectCaseHTML(p, { coverUrl, galleryUrls, photoCount }) {
  return {
    meta: `${t(`category_${categoryOf(p)}`)} • ${t("project_photos", photoCount)}`,
    body: `
      <h3 class="dialog__title" id="case-title">${esc(p.title)}</h3>
      ${coverUrl
        ? photoHTML(coverUrl, p.title, "case__cover")
        : `<p class="muted">${t("case_no_photos")}</p>`}
      ${p.description ? `<p class="dialog__text">${esc(p.description)}</p>` : ""}
      ${galleryUrls.length ? `
        <div class="case__gallery">
          ${galleryUrls.map((url, i) => photoHTML(url, `${p.title} — ${i + 2}`)).join("")}
        </div>` : ""}
      <div class="case__cta">
        <span>${t("case_cta_text")}</span>
        <button class="btn btn--dark" data-order="concept" data-order-type="${categoryOf(p)}">${t("case_cta_btn")}</button>
      </div>`,
  };
}
