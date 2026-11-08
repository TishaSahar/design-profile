import { api } from "./api.js";
import { t, getLang, setLang, applyTranslations } from "./locales.js";

// ── Language switcher ─────────────────────────────────────────────────────────
document.querySelectorAll(".lang-option").forEach((btn) => {
  btn.addEventListener("click", () => {
    setLang(btn.dataset.lang);
    applyTranslations();
    document.activeElement?.blur();
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
  markActiveNav(route.page);
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

function markActiveNav(page) {
  const current = page === "project" ? "portfolio" : page;
  document.querySelectorAll(".site-nav__link").forEach((link) => {
    if (link.dataset.page === current) link.setAttribute("aria-current", "page");
    else link.removeAttribute("aria-current");
  });
}

window.addEventListener("hashchange", () => {
  window.scrollTo({ top: 0, behavior: "instant" });
  navigate();
});

// ── Portfolio page ────────────────────────────────────────────────────────────
async function renderPortfolio(main) {
  const [projectsResult, contactsResult] = await Promise.all([
    api.projects.list(),
    api.contacts.get().catch(() => ({})),
  ]);
  const projects = projectsResult.data ?? [];
  const contacts = contactsResult.data ?? {};
  const heroPhoto = contacts.photo_url
    ? `style="background-image: url('${api.contacts.photoUrl()}')"`
    : "";

  main.innerHTML = `
    <section class="hero">
      <div class="hero__bg" ${heroPhoto}></div>
      <div class="hero__content">
        <h1 class="hero__title">${t("logo")}</h1>
        <p class="hero__subtitle">${t("hero_subtitle")}</p>
        <button class="btn btn--primary btn--lg" id="open-request">${t("hero_cta")}</button>
      </div>
      <button class="hero__scroll" id="scroll-to-portfolio" aria-label="${t("nav_portfolio")}">
        <svg width="38" height="18" viewBox="0 0 38 18" aria-hidden="true">
          <path d="M1 1l18 15L37 1" stroke="currentColor" stroke-width="1.5" fill="none"/>
        </svg>
      </button>
    </section>
    <section class="section" id="portfolio">
      <div class="section__inner">
        <h2 class="section__title">${t("nav_portfolio")}</h2>
        <p class="section__descr">${t("portfolio_descr")}</p>
        ${projects.length
          ? `<div class="grid" id="portfolio-grid"></div>`
          : `<p class="empty">${t("portfolio_empty")}</p>`}
      </div>
    </section>`;

  document.getElementById("open-request").addEventListener("click", openRequestModal);
  document.getElementById("scroll-to-portfolio").addEventListener("click", () => {
    document.getElementById("portfolio").scrollIntoView({ behavior: "smooth" });
  });

  const grid = document.getElementById("portfolio-grid");
  for (const p of projects) {
    const card = document.createElement("article");
    card.className = "card";
    card.innerHTML = `
      <a href="#/project/${p.id}" class="card__link" data-tooltip="${esc(t("project_open_tooltip"))}">
        <div class="card__img-wrap">
          ${p.cover_media_id
            ? `<img src="${api.projects.mediaUrl(p.id, p.cover_media_id)}" alt="${esc(p.title)}" loading="lazy" class="card__img">`
            : `<div class="card__img--placeholder"></div>`}
          <div class="card__overlay">
            <h3 class="card__title">${esc(p.title)}</h3>
            ${p.description ? `<p class="card__desc">${esc(p.description)}</p>` : ""}
          </div>
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
    <section class="section section--first section--fill">
      <div class="section__inner project">
        <a href="#/" class="btn btn--ghost btn--sm">${t("project_back")}</a>
        <h1 class="section__title project__title">${esc(p.title)}</h1>
        ${p.description ? `<p class="section__descr">${esc(p.description)}</p>` : ""}
        <div class="gallery">${mediaHTML || `<p class="empty">${t("project_no_photos")}</p>`}</div>
        <div class="section__cta">
          <button class="btn btn--primary" id="open-request">${t("project_request_btn")}</button>
        </div>
      </div>
    </section>`;

  document.getElementById("open-request").addEventListener("click", openRequestModal);
}

// ── Contact links (shared by the About & Contacts pages) ──────────────────────
const CONTACT_ICONS = {
  instagram: `<svg viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.8">
      <rect x="4" y="4" width="16" height="16" rx="4.5"/><circle cx="12" cy="12" r="3.6"/>
      <circle cx="16.6" cy="7.4" r="1" fill="currentColor" stroke="none"/></svg>`,
  telegram: `<svg viewBox="0 0 24 24" aria-hidden="true" fill="currentColor">
      <path d="M21.3 4.2 2.9 11.3c-.8.3-.8 1.4 0 1.7l4.5 1.5 1.8 5.5c.2.6 1 .8 1.5.4l2.6-2.2 4.6 3.4c.6.4 1.4.1 1.6-.6L22.6 5.4c.2-.8-.6-1.5-1.3-1.2ZM9.6 14.4l-.5 3.3-1.2-3.8 9.9-6.2-8.2 6.7Z"/></svg>`,
  email: `<svg viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round">
      <rect x="3.5" y="5.5" width="17" height="13" rx="2"/><path d="m4 7 8 6 8-6"/></svg>`,
};

function contactLinksHTML(c) {
  const handle = (value) => encodeURIComponent(value.replace(/^@/, ""));
  const links = [
    c.instagram && { key: "instagram", href: `https://instagram.com/${handle(c.instagram)}`, value: c.instagram },
    c.telegram && { key: "telegram", href: `https://t.me/${handle(c.telegram)}`, value: c.telegram },
    c.email && { key: "email", href: `mailto:${c.email}`, value: c.email },
  ].filter(Boolean);

  const items = links.map((link) => `
    <a class="contact-link" href="${esc(link.href)}" aria-label="${t(`contacts_${link.key}`)}: ${esc(link.value)}"
      ${link.key === "email" ? "" : `target="_blank" rel="noopener"`}>
      <span class="contact-link__icon">${CONTACT_ICONS[link.key]}</span>
      <span class="contact-link__value">${esc(link.value)}</span>
    </a>`);
  return `<div class="contact-links">${items.join("")}</div>`;
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
    ${c.photo_url ? `
      <div class="about__hero">
        <img src="${api.contacts.photoUrl()}" alt="" class="about__photo">
      </div>` : ""}
    <section class="section section--muted section--fill ${c.photo_url ? "" : "section--first"}">
      <div class="section__inner">
        <h1 class="section__title">${t("about_title")}</h1>
        ${bioHTML ? `<div class="about__bio">${bioHTML}</div>` : ""}
        ${contactLinksHTML(c)}
        <div class="section__cta">
          <button class="btn btn--primary" id="open-request">${t("about_request_btn")}</button>
        </div>
      </div>
    </section>`;

  document.getElementById("open-request").addEventListener("click", openRequestModal);
}

// ── Contacts page ─────────────────────────────────────────────────────────────
async function renderContacts(main) {
  const result = await api.contacts.get();
  const c = result.data ?? {};

  main.innerHTML = `
    <section class="section section--muted section--first section--fill">
      <div class="section__inner">
        <h1 class="section__title">${t("contacts_title")}</h1>
        <p class="section__descr">${t("contacts_descr")}</p>
        ${contactLinksHTML(c)}
        <div class="section__cta">
          <button class="btn btn--primary" id="open-request">${t("contacts_request_btn")}</button>
        </div>
      </div>
    </section>`;

  document.getElementById("open-request").addEventListener("click", openRequestModal);
}

// ── Request modal ─────────────────────────────────────────────────────────────
function openRequestModal() {
  const overlay = document.getElementById("modal-overlay");
  overlay.classList.remove("hidden");
  overlay.innerHTML = buildRequestForm();
  overlay.querySelector("#close-modal").addEventListener("click", closeModal);
  const form = overlay.querySelector("#request-form");
  form.addEventListener("submit", submitRequest);
  form.addEventListener("input", () => syncPreferredContact(form));
  syncPreferredContact(form);
  overlay.addEventListener("click", (e) => { if (e.target === overlay) closeModal(); });
}

function closeModal() {
  document.getElementById("modal-overlay").classList.add("hidden");
}

// Offers only the contacts that are filled in; there is nothing to choose
// from until at least two of them are.
function syncPreferredContact(form) {
  const group = form.querySelector("#rf-preferred");
  const radios = [...group.querySelectorAll('input[name="preferred_contact"]')];

  for (const radio of radios) {
    const filled = form.elements[radio.value].value.trim() !== "";
    radio.disabled = !filled;
    radio.closest("label").classList.toggle("hidden", !filled);
  }

  const available = radios.filter((radio) => !radio.disabled);
  group.classList.toggle("hidden", available.length < 2);
  if (available.length && !available.some((radio) => radio.checked)) {
    available[0].checked = true;
  }
}

function buildRequestForm() {
  const preferredOptions = [
    ["email", "modal_email_short"],
    ["phone", "modal_phone"],
    ["telegram", "modal_telegram"],
  ].map(([value, label]) => `
    <label class="pref-option">
      <input type="radio" name="preferred_contact" value="${value}">
      <span>${t(label)}</span>
    </label>`).join("");

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
          <label class="form__label" for="rf-email">${t("modal_email")}</label>
          <input id="rf-email" name="email" type="email" autocomplete="email" class="form__input" required>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-phone">${t("modal_phone")}</label>
          <input id="rf-phone" name="phone" type="tel" autocomplete="tel" class="form__input">
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-telegram">${t("modal_telegram")}</label>
          <input id="rf-telegram" name="telegram" placeholder="@username" class="form__input">
        </div>
        <fieldset id="rf-preferred" class="form__fieldset hidden">
          <legend class="form__label">${t("modal_preferred")}</legend>
          <div class="pref-options">${preferredOptions}</div>
        </fieldset>
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

  const email = form.querySelector("#rf-email");
  if (!email.value.trim() || !email.checkValidity()) {
    showFormError(errEl, t("err_email"));
    email.focus();
    return;
  }

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
