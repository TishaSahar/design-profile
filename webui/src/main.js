import { api } from "./api.js";
import { t, getLang, setLang, applyTranslations } from "./locales.js";

// ── Language switcher ─────────────────────────────────────────────────────────
document.querySelectorAll(".lang-option").forEach((btn) => {
  btn.addEventListener("click", () => {
    setLang(btn.dataset.lang);
    applyTranslations();
    navigate();
  });
});

applyTranslations();

// ── Router ────────────────────────────────────────────────────────────────────
function getRoute() {
  const hash = location.hash.replace("#", "") || "/";
  const match = hash.match(/^\/project\/([a-f0-9-]+)/);
  if (match) return { page: "project", id: match[1] };
  if (hash === "/contacts") return { page: "contacts" };
  if (hash === "/about")    return { page: "about" };
  return { page: "portfolio" };
}

async function navigate() {
  const route = getRoute();
  const main = document.getElementById("main");
  main.innerHTML = `<div class="loader"></div>`;
  try {
    if      (route.page === "project")   await renderProject(main, route.id);
    else if (route.page === "contacts")  await renderContacts(main);
    else if (route.page === "about")     await renderAbout(main);
    else                                 await renderPortfolio(main);
  } catch (err) {
    main.innerHTML = `<p class="error">${t("err_load")} ${err.message}</p>`;
  }
}

window.addEventListener("hashchange", navigate);

// ── Portfolio page ────────────────────────────────────────────────────────────
async function renderPortfolio(main) {
  const result = await api.projects.list();
  const projects = result.data ?? [];

  if (projects.length === 0) {
    main.innerHTML = `<p class="empty">${t("portfolio_empty")}</p>`;
    return;
  }

  main.innerHTML = `<div class="grid" id="portfolio-grid"></div>`;
  const grid = document.getElementById("portfolio-grid");

  for (const p of projects) {
    const card = document.createElement("article");
    card.className = "card";
    card.innerHTML = `
      <a href="#/project/${p.id}" class="card__link">
        <div class="card__img-wrap">
          ${p.cover_media_id
            ? `<img src="${api.projects.mediaUrl(p.id, p.cover_media_id)}" alt="${esc(p.title)}" loading="lazy" class="card__img">`
            : `<div class="card__img card__img--placeholder"></div>`}
        </div>
        <div class="card__body">
          <h2 class="card__title">${esc(p.title)}</h2>
          ${p.description ? `<p class="card__desc">${esc(p.description)}</p>` : ""}
        </div>
      </a>`;
    grid.appendChild(card);
  }
}

// ── Project detail page ───────────────────────────────────────────────────────
async function renderProject(main, id) {
  const result = await api.projects.get(id);
  const p = result.data;

  const mediaHTML = (p.media ?? [])
    .map(
      (m) => `
        <div class="gallery__item">
          <img src="${api.projects.mediaUrl(p.id, m.id)}" alt="${esc(m.filename)}" loading="lazy" class="gallery__img">
        </div>`
    )
    .join("");

  main.innerHTML = `
    <div class="project">
      <a href="#/" class="btn btn--ghost">${t("project_back")}</a>
      <h1 class="project__title">${esc(p.title)}</h1>
      ${p.description ? `<p class="project__desc">${esc(p.description)}</p>` : ""}
      <div class="gallery">${mediaHTML || `<p class="empty">${t("project_no_photos")}</p>`}</div>
      <div class="project__cta">
        <button class="btn btn--primary" id="open-request">${t("project_request_btn")}</button>
      </div>
    </div>`;

  document.getElementById("open-request")?.addEventListener("click", openRequestModal);
}

// ── About page ────────────────────────────────────────────────────────────────
async function renderAbout(main) {
  const result = await api.contacts.get();
  const c = result.data ?? {};

  const bioHTML = (c.bio || "")
    .split(/\n{2,}/)
    .filter(Boolean)
    .map((para) => `<p>${esc(para).replace(/\n/g, "<br>")}</p>`)
    .join("");

  main.innerHTML = `
    <section class="about">
      ${c.photo_url ? `
        <div class="about__hero">
          <img src="${api.contacts.photoUrl()}" alt="" class="about__photo">
        </div>` : ""}
      ${bioHTML ? `<div class="about__bio">${bioHTML}</div>` : ""}
      <h1 class="about__title">${t("about_title")}</h1>
      <div class="about__directory">
        ${c.email ? `
          <div class="dir-row">
            <span class="dir-row__label">${t("about_email")}</span>
            <a href="mailto:${esc(c.email)}" class="dir-row__value">${esc(c.email)}</a>
          </div>` : ""}
        ${c.instagram ? `
          <div class="dir-row">
            <span class="dir-row__label">${t("about_instagram")}</span>
            <a href="https://instagram.com/${c.instagram.replace("@", "")}" target="_blank" rel="noopener" class="dir-row__value">${esc(c.instagram)}</a>
          </div>` : ""}
        ${c.telegram ? `
          <div class="dir-row">
            <span class="dir-row__label">${t("about_telegram")}</span>
            <a href="https://t.me/${c.telegram.replace("@", "")}" target="_blank" rel="noopener" class="dir-row__value">${esc(c.telegram)}</a>
          </div>` : ""}
      </div>
      <div class="about__cta">
        <button class="btn btn--primary" id="open-request">${t("about_request_btn")}</button>
      </div>
    </section>`;

  document.getElementById("open-request")?.addEventListener("click", openRequestModal);
}

// ── Contacts page ─────────────────────────────────────────────────────────────
async function renderContacts(main) {
  const result = await api.contacts.get();
  const c = result.data ?? {};

  main.innerHTML = `
    <section class="contacts">
      <h1 class="contacts__title">${t("contacts_title")}</h1>
      <div class="about__directory">
        ${c.email ? `
          <div class="dir-row">
            <span class="dir-row__label">${t("contacts_email")}</span>
            <a href="mailto:${esc(c.email)}" class="dir-row__value">${esc(c.email)}</a>
          </div>` : ""}
        ${c.instagram ? `
          <div class="dir-row">
            <span class="dir-row__label">${t("contacts_instagram")}</span>
            <a href="https://instagram.com/${c.instagram.replace("@", "")}" target="_blank" rel="noopener" class="dir-row__value">${esc(c.instagram)}</a>
          </div>` : ""}
        ${c.telegram ? `
          <div class="dir-row">
            <span class="dir-row__label">${t("contacts_telegram")}</span>
            <a href="https://t.me/${c.telegram.replace("@", "")}" target="_blank" rel="noopener" class="dir-row__value">${esc(c.telegram)}</a>
          </div>` : ""}
      </div>
      <div class="contacts__cta">
        <button class="btn btn--primary" id="open-request">${t("contacts_request_btn")}</button>
      </div>
    </section>`;

  document.getElementById("open-request")?.addEventListener("click", openRequestModal);
}

// ── Request modal ─────────────────────────────────────────────────────────────
function openRequestModal() {
  const overlay = document.getElementById("modal-overlay");
  overlay.classList.remove("hidden");
  overlay.innerHTML = buildRequestForm();
  overlay.querySelector("#close-modal").addEventListener("click", closeModal);
  overlay.querySelector("#request-form").addEventListener("submit", submitRequest);
  overlay.addEventListener("click", (e) => { if (e.target === overlay) closeModal(); });
}

function closeModal() {
  document.getElementById("modal-overlay").classList.add("hidden");
}

function buildRequestForm() {
  return `
    <div class="modal">
      <button id="close-modal" class="modal__close" aria-label="${t("modal_close")}">&times;</button>
      <h2 class="modal__title">${t("modal_title")}</h2>
      <form id="request-form" class="form" enctype="multipart/form-data" novalidate>
        <div class="form__row">
          <label class="form__label" for="rf-first">${t("modal_first_name")}</label>
          <input id="rf-first" name="first_name" class="form__input" required>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-last">${t("modal_last_name")}</label>
          <input id="rf-last" name="last_name" class="form__input" required>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-contact">${t("modal_contact")}</label>
          <input id="rf-contact" name="contact" class="form__input" required>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-desc">${t("modal_description")}</label>
          <textarea id="rf-desc" name="description" class="form__textarea" rows="4" required></textarea>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-files">${t("modal_files")}</label>
          <input id="rf-files" name="attachments" type="file"
            accept="image/jpeg,image/png,image/webp,image/gif,application/pdf"
            multiple class="form__file">
        </div>
        <div class="form__row form__row--consent">
          <label class="form__checkbox-label">
            <input type="checkbox" id="rf-consent" name="consented" class="form__checkbox" required>
            ${t("modal_consent_label")}
            <button type="button" class="btn-link" id="show-consent">${t("modal_consent_link")}</button>
          </label>
        </div>
        <div id="consent-text" class="consent-text hidden">
          <p>${t("modal_consent_body")}</p>
        </div>
        <div id="form-error" class="form__error hidden"></div>
        <button type="submit" class="btn btn--primary btn--full">${t("modal_submit")}</button>
      </form>
    </div>`;
}

async function submitRequest(e) {
  e.preventDefault();
  const form = e.target;
  const errEl = form.querySelector("#form-error");
  errEl.classList.add("hidden");

  if (!form.querySelector("#rf-consent").checked) {
    showFormError(errEl, t("err_consent"));
    return;
  }

  const fd = new FormData(form);
  fd.set("consented", "true");

  try {
    await api.requests.create(fd);
    form.closest(".modal").innerHTML = `
      <div class="modal__success">
        <h2>${t("modal_success_title")}</h2>
        <p>${t("modal_success_body")}</p>
        <button class="btn btn--primary" onclick="document.getElementById('modal-overlay').classList.add('hidden')">${t("modal_close")}</button>
      </div>`;
  } catch (err) {
    showFormError(errEl, err.message);
  }
}

function showFormError(el, msg) {
  el.textContent = msg;
  el.classList.remove("hidden");
}

// ── Consent text toggle ───────────────────────────────────────────────────────
document.addEventListener("click", (e) => {
  if (e.target.id === "show-consent") {
    document.getElementById("consent-text")?.classList.toggle("hidden");
  }
});

// ── Utilities ─────────────────────────────────────────────────────────────────
function esc(str) {
  return String(str ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

// Bootstrap
navigate();
