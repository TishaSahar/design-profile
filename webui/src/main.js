import { api } from "./api.js";
import { t, tRu, getLang, setLang, applyTranslations } from "./locales.js";
import {
  CATEGORIES, categoryOf, esc, pad, photoHTML, initPhotos, projectCardHTML, projectCaseHTML,
} from "./project-view.js";

// Prices in euros per m², by property type and service.
const PRICING = {
  residential: { layout: 7, drawings: 10, concept: 19 },
  commercial: { layout: 9, drawings: 13, concept: 24 },
};

const state = {
  projects: [],
  projectsError: null,
  contacts: {},
  faq: [],
  filter: "all",
  calcType: "residential",
  area: 75,
  // True once the visitor types an area into the request form: from then on the
  // calculator no longer overwrites it.
  formAreaEdited: false,
  faqOpen: new Set([0]),
  openProjectId: null,
  openProject: null,
  openSample: null,
};

const $ = (id) => document.getElementById(id);

// ── Bootstrap ────────────────────────────────────────────────────────────────
applyTranslations();
initPhotos(document);
initLanguage();
initMenu();
initCalculator();
initForm();
initClicks();
renderStatic();
loadData();
window.addEventListener("hashchange", handleHash);

async function loadData() {
  const [projects, contacts, faq] = await Promise.allSettled([
    api.projects.list(),
    api.contacts.get(),
    api.faq.list(),
  ]);
  if (projects.status === "fulfilled") {
    state.projects = projects.value.data ?? [];
    state.projectsError = null;
  } else {
    state.projectsError = projects.reason;
  }
  if (contacts.status === "fulfilled") state.contacts = contacts.value.data ?? {};
  if (faq.status === "fulfilled") state.faq = faq.value.data ?? [];
  renderData();
  handleHash();
}

// Content that comes from the dictionary only.
function renderStatic() {
  renderDirectionPrices();
  renderServices();
  renderServiceOptions();
  renderProcess();
  renderFacts();
}

// Content that depends on the API.
function renderData() {
  renderFilters();
  renderProjects();
  renderAbout();
  renderFaq();
  renderContacts();
}

// ── Language ─────────────────────────────────────────────────────────────────
function initLanguage() {
  document.querySelectorAll(".lang-option").forEach((btn) => {
    btn.addEventListener("click", () => {
      setLang(btn.dataset.lang);
      applyTranslations();
      renderStatic();
      renderData();
      if (state.openProject) renderCase();
      if (state.openSample) renderSample(state.openSample);
      document.activeElement?.blur();
    });
  });
}

// ── Mobile menu ──────────────────────────────────────────────────────────────
function initMenu() {
  const toggle = $("menu-toggle");
  const menu = $("mobile-menu");
  toggle.addEventListener("click", () => setMenu(menu.hidden));
  menu.addEventListener("click", (e) => {
    if (e.target.closest("a")) setMenu(false);
  });
}

function setMenu(open) {
  $("mobile-menu").hidden = !open;
  $("menu-toggle").setAttribute("aria-expanded", String(open));
}

// ── Clicks (delegated) & keyboard ────────────────────────────────────────────
function initClicks() {
  document.addEventListener("click", (e) => {
    const order = e.target.closest("[data-order]");
    if (order) {
      closeDialog("project-modal");
      orderService(order.dataset.order, order.dataset.orderType);
      return;
    }
    const sample = e.target.closest("[data-sample]");
    if (sample) return openSample(sample.dataset.sample);

    const filterTab = e.target.closest("[data-filter-tab]");
    if (filterTab) return setFilter(filterTab.dataset.filterTab);

    // Direction blocks are links to #projects: filter, then let the browser scroll.
    const direction = e.target.closest("a[data-filter]");
    if (direction) return setFilter(direction.dataset.filter);

    const faq = e.target.closest("[data-faq]");
    if (faq) return toggleFaq(faq);

    const close = e.target.closest("[data-close-dialog]");
    if (close) return closeDialog(close.closest(".overlay").id);
    if (e.target.classList.contains("overlay")) closeDialog(e.target.id);
  });

  document.addEventListener("keydown", (e) => {
    if (e.key !== "Escape") return;
    const open = [...document.querySelectorAll(".overlay:not([hidden])")].pop();
    if (open) closeDialog(open.id);
    setMenu(false);
  });
}

// ── Deep links ───────────────────────────────────────────────────────────────
// #/project/<id> opens a case; the hashes of the previous multi-page site
// (#/, #/about, #/contacts) lead to the matching section.
function handleHash() {
  const match = location.hash.match(/^#\/project\/([0-9a-f-]{36})$/i);
  if (match) return openProject(match[1]);

  const legacy = { "#/about": "about", "#/contacts": "contacts" };
  const section = legacy[location.hash];
  if (section) {
    history.replaceState(null, "", `#${section}`);
    $(section).scrollIntoView();
  } else if (location.hash === "#/") {
    history.replaceState(null, "", location.pathname + location.search);
  }
}

// ── Photos ───────────────────────────────────────────────────────────────────
// A photo file that does not exist yet shows a calm coloured block instead of a
// broken image.

// ── Directions & projects ────────────────────────────────────────────────────
function renderDirectionPrices() {
  document.querySelectorAll("[data-price-from]").forEach((el) => {
    const from = Math.min(...Object.values(PRICING[el.dataset.priceFrom]));
    el.textContent = `${t("dir_from")} ${euro(from)} / ${t("unit_m2")}`;
  });
}

function setFilter(filter) {
  state.filter = filter === "all" || CATEGORIES.includes(filter) ? filter : "all";
  renderFilters();
  renderProjects();
}

function renderFilters() {
  const box = $("project-filters");
  if (!state.projects.length) {
    box.innerHTML = "";
    return;
  }
  const count = (f) => (f === "all" ? state.projects : state.projects.filter((p) => categoryOf(p) === f)).length;
  box.innerHTML = ["all", ...CATEGORIES].map((f) => `
    <button class="tab" data-filter-tab="${f}" aria-pressed="${state.filter === f}">
      ${f === "all" ? t("filter_all") : t(`category_${f}`)} (${pad(count(f))})
    </button>`).join("");
}

function renderProjects() {
  const grid = $("projects-grid");
  if (state.projectsError) {
    grid.innerHTML = `<p class="error">${t("err_load")} ${esc(state.projectsError.message)}</p>`;
    return;
  }
  if (!state.projects.length) {
    grid.innerHTML = `<p class="empty">${t("projects_empty")}</p>`;
    return;
  }
  const shown = state.projects.filter((p) => state.filter === "all" || categoryOf(p) === state.filter);
  if (!shown.length) {
    grid.innerHTML = `<p class="empty">${t("projects_empty_filter")}</p>`;
    return;
  }
  grid.innerHTML = shown.map((p, i) => projectCardHTML(p, {
    number: i + 1,
    coverUrl: p.cover_media_id ? api.projects.mediaUrl(p.id, p.cover_media_id) : null,
  })).join("");
  initPhotos(grid);
}

// ── Project case dialog ──────────────────────────────────────────────────────
async function openProject(id) {
  state.openProjectId = id;
  state.openProject = null;
  $("case-meta").textContent = "";
  $("case-body").innerHTML = `<div class="loader"></div>`;
  showDialog("project-modal");
  try {
    const result = await api.projects.get(id);
    if (state.openProjectId !== id) return;
    state.openProject = result.data;
    renderCase();
  } catch (err) {
    if (state.openProjectId !== id) return;
    $("case-body").innerHTML = `<p class="error">${t("err_load")} ${esc(err.message)}</p>`;
  }
}

function renderCase() {
  const p = state.openProject;
  const media = p.media ?? [];
  const coverId = p.cover_media_id ?? media[0]?.id;
  const { meta, body } = projectCaseHTML(p, {
    coverUrl: coverId ? api.projects.mediaUrl(p.id, coverId) : null,
    galleryUrls: media.filter((m) => m.id !== coverId).map((m) => api.projects.mediaUrl(p.id, m.id)),
    photoCount: media.length,
  });
  $("case-meta").textContent = meta;
  $("case-body").innerHTML = body;
  initPhotos($("case-body"));
}

// ── Services & calculator ────────────────────────────────────────────────────
function initCalculator() {
  document.querySelectorAll("[data-calc-type]").forEach((btn) => {
    btn.addEventListener("click", () => {
      state.calcType = btn.dataset.calcType;
      document.querySelectorAll("[data-calc-type]").forEach((b) => {
        b.setAttribute("aria-pressed", String(b === btn));
      });
      updateEstimates();
    });
  });
  $("area-range").addEventListener("input", (e) => {
    state.area = Number(e.target.value) || 75;
    updateEstimates();
    syncFormArea();
  });
}

function renderServices() {
  $("service-cards").innerHTML = t("services").map((s) => {
    const accent = s.id === "concept";
    return `
      <article class="service-card ${accent ? "service-card--accent" : ""}">
        <div>
          <div class="service-card__top"><span>${s.tag}</span></div>
          <h3 class="service-card__title">${s.title}</h3>
          ${s.note ? `<p class="service-card__note">${s.note}</p>` : ""}
          <p class="service-card__text">${s.text}</p>
          <ul class="service-card__items">${s.items.map((item) => `<li>${item}</li>`).join("")}</ul>
        </div>
        <div>
          <div class="service-card__prices">
            <div class="price-row"><span>${t("price_residential")}</span><span>${euro(PRICING.residential[s.id])} / ${t("unit_m2")}</span></div>
            <div class="price-row"><span>${t("price_commercial")}</span><span>${euro(PRICING.commercial[s.id])} / ${t("unit_m2")}</span></div>
            <div class="price-row price-row--total">
              <span>${t("price_for")} <span data-area-echo>${state.area}</span> ${t("unit_m2")}</span>
              <span data-estimate="${s.id}"></span>
            </div>
          </div>
          <div class="service-card__actions">
            <button class="btn btn--dark" data-order="${s.id}">${t("service_choose")}</button>
            <button class="btn btn--ghost" data-sample="${s.id}">${t("service_sample")}</button>
          </div>
        </div>
      </article>`;
  }).join("");
  updateEstimates();
}

function updateEstimates() {
  $("area-display").textContent = state.area;
  document.querySelectorAll("[data-area-echo]").forEach((el) => { el.textContent = state.area; });
  document.querySelectorAll("[data-estimate]").forEach((el) => {
    el.textContent = euro(state.area * PRICING[state.calcType][el.dataset.estimate]);
  });
}

// Pre-fills the request form and takes the visitor to it.
function orderService(serviceId, type = state.calcType) {
  $("rf-service").value = serviceId;
  $("rf-object").value = CATEGORIES.includes(type) ? type : "residential";
  syncFormArea();
  $("contacts").scrollIntoView();
  $("rf-first").focus({ preventScroll: true });
}

// ── Service sample dialog ────────────────────────────────────────────────────
function openSample(type) {
  state.openSample = type;
  renderSample(type);
  showDialog("sample-modal");
}

function renderSample(type) {
  const s = t("samples")[type];
  let body = "";
  if (type === "layout") {
    body = `
      <svg viewBox="0 0 640 300" class="sample-plan" role="img" aria-label="${s.sheet}">
        <rect x="30" y="25" width="580" height="250" fill="none" stroke="#1E1C1A" stroke-width="4"/>
        <line x1="250" y1="25" x2="250" y2="180" stroke="#1E1C1A" stroke-width="2.5"/>
        <line x1="250" y1="225" x2="250" y2="275" stroke="#1E1C1A" stroke-width="2.5"/>
        <rect x="55" y="50" width="140" height="45" fill="#E4E0D8" stroke="#1E1C1A"/>
        <rect x="290" y="50" width="290" height="85" fill="#77785E" fill-opacity="0.2" stroke="#1E1C1A"/>
        <text x="125" y="77" font-size="11" fill="#1E1C1A" text-anchor="middle">${s.bedroom}</text>
        <text x="435" y="96" font-size="11" fill="#1E1C1A" text-anchor="middle">${s.living}</text>
        <text x="320" y="245" font-size="10" fill="#754044" text-anchor="middle" font-family="monospace">${s.sheet}</text>
      </svg>`;
  } else if (type === "drawings") {
    body = `<div class="sample-sheets">${s.sheets.map(([sheet, text]) => `<div><strong>${sheet}</strong> ${text}</div>`).join("")}</div>`;
  } else {
    body = `
      <div class="sample-table-wrap">
        <table class="sample-table">
          <thead><tr>${s.head.map((h) => `<th>${h}</th>`).join("")}</tr></thead>
          <tbody>${s.rows.map((row) => `<tr>${row.map((cell) => `<td>${cell}</td>`).join("")}</tr>`).join("")}</tbody>
        </table>
      </div>`;
  }
  $("sample-body").innerHTML = `
    <h3 class="dialog__title" id="sample-title">${s.title}</h3>
    <p class="dialog__text">${s.text}</p>
    ${body}`;
}

// ── Dialogs ──────────────────────────────────────────────────────────────────
let returnFocusTo = null;

function showDialog(id) {
  const overlay = $(id);
  if (overlay.hidden) returnFocusTo = document.activeElement;
  overlay.hidden = false;
  document.body.classList.add("is-locked");
  overlay.querySelector("[data-close-dialog]").focus();
}

function closeDialog(id) {
  const overlay = $(id);
  if (overlay.hidden) return;
  overlay.hidden = true;
  if (!document.querySelector(".overlay:not([hidden])")) document.body.classList.remove("is-locked");
  if (id === "project-modal") {
    state.openProjectId = null;
    state.openProject = null;
    if (location.hash.startsWith("#/project/")) {
      history.replaceState(null, "", location.pathname + location.search);
    }
  } else if (id === "sample-modal") {
    state.openSample = null;
  }
  returnFocusTo?.focus?.({ preventScroll: true });
}

// ── Process, about, FAQ, contacts ────────────────────────────────────────────
function renderProcess() {
  $("process-steps").innerHTML = t("process").map((step, i) => `
    <li class="process__step">
      <span class="process__stage">${t("process_stage")} ${pad(i + 1)}</span>
      <h3 class="process__title">${step.title}</h3>
      <p class="process__text">${step.text}</p>
    </li>`).join("");
}

function renderFacts() {
  $("about-facts").innerHTML = t("about_facts").map((fact) => `
    <div><dt>${fact.label}</dt><dd>${fact.value}</dd></div>`).join("");
}

// The about text comes from backend/config/about/bio.txt (bio.en.txt in English).
function renderAbout() {
  const c = state.contacts;
  const bio = (getLang() === "en" && c.bio_en) || c.bio || "";
  $("about-bio").innerHTML = paragraphs(bio);
}

// The FAQ comes from backend/config/about/faq.yaml; the section stays hidden
// while the list is empty.
function renderFaq() {
  $("faq").hidden = !state.faq.length;
  $("faq-list").innerHTML = state.faq.map((item, i) => {
    const open = state.faqOpen.has(i);
    return `
      <div class="faq__item">
        <button class="faq__question" data-faq="${i}" aria-expanded="${open}" aria-controls="faq-answer-${i}">
          <span>${esc(localized(item.question))}</span><span class="faq__icon" aria-hidden="true">${open ? "−" : "+"}</span>
        </button>
        <div class="faq__answer" id="faq-answer-${i}" ${open ? "" : "hidden"}>${paragraphs(localized(item.answer))}</div>
      </div>`;
  }).join("");
}

function toggleFaq(button) {
  const i = Number(button.dataset.faq);
  const open = !state.faqOpen.has(i);
  if (open) state.faqOpen.add(i);
  else state.faqOpen.delete(i);
  button.setAttribute("aria-expanded", String(open));
  button.querySelector(".faq__icon").textContent = open ? "−" : "+";
  $(`faq-answer-${i}`).hidden = !open;
}

function renderContacts() {
  const c = state.contacts;
  const handle = (value) => encodeURIComponent(value.replace(/^@/, ""));
  const rows = [
    c.telegram && { label: t("contacts_telegram"), value: c.telegram, href: `https://t.me/${handle(c.telegram)}` },
    c.instagram && { label: t("contacts_instagram"), value: c.instagram, href: `https://instagram.com/${handle(c.instagram)}` },
    c.email && { label: t("contacts_email"), value: c.email, href: `mailto:${c.email}`, local: true },
  ].filter(Boolean);
  const list = $("contact-list");
  list.hidden = !rows.length;
  list.innerHTML = rows.map((row) => `
    <a href="${esc(row.href)}" ${row.local ? "" : `target="_blank" rel="noopener"`}>
      <span class="caps">${row.label}</span><span>${esc(row.value)} →</span>
    </a>`).join("");
}

// ── Request form ─────────────────────────────────────────────────────────────
function renderServiceOptions() {
  const select = $("rf-service");
  const current = select.value || "layout";
  select.innerHTML = [
    ...t("services").map((s) => `<option value="${s.id}">${s.title}</option>`),
    `<option value="several">${t("form_service_several")}</option>`,
  ].join("");
  select.value = current;
}

// The request form takes the area chosen in the calculator, unless the visitor
// has already set a different one in the form.
function syncFormArea() {
  if (!state.formAreaEdited) $("rf-area").value = state.area;
}

function initForm() {
  const form = $("request-form");
  $("rf-area").addEventListener("input", () => { state.formAreaEdited = true; });
  form.addEventListener("input", () => {
    syncPreferredContact();
    $("form-success").hidden = true;
  });
  form.addEventListener("submit", submitRequest);
  $("consent-toggle").addEventListener("click", (e) => {
    const text = $("consent-text");
    text.hidden = !text.hidden;
    e.currentTarget.setAttribute("aria-expanded", String(!text.hidden));
  });
  syncPreferredContact();
}

// Offers only the contacts that are filled in; there is nothing to choose
// from until at least two of them are.
function syncPreferredContact() {
  const form = $("request-form");
  const group = $("rf-preferred");
  const radios = [...group.querySelectorAll('input[name="preferred_contact"]')];
  for (const radio of radios) {
    const filled = form.elements[radio.value].value.trim() !== "";
    radio.disabled = !filled;
    radio.closest("label").hidden = !filled;
  }
  const available = radios.filter((r) => !r.disabled);
  group.hidden = available.length < 2;
  if (available.length && !available.some((r) => r.checked)) available[0].checked = true;
}

// Type, service and area have no columns of their own: they go into the
// description, in Russian, for the admin panel.
function composeDescription() {
  const object = $("rf-object").value === "commercial" ? "form_object_commercial" : "form_object_residential";
  const serviceId = $("rf-service").value;
  const service = serviceId === "several"
    ? tRu("form_service_several")
    : tRu("services").find((s) => s.id === serviceId)?.title ?? serviceId;
  const lines = [`Тип объекта: ${tRu(object)}`, `Услуга: ${service}`];
  const area = Number($("rf-area").value);
  if (area > 0) lines.push(`Площадь: ${area} м²`);
  const comment = $("rf-comment").value.trim();
  return lines.join("\n") + (comment ? `\n\n${comment}` : "");
}

async function submitRequest(e) {
  e.preventDefault();
  const form = e.target;
  const errorBox = $("form-error");
  const fail = (message, field) => {
    errorBox.textContent = message;
    errorBox.hidden = false;
    field?.focus();
  };
  errorBox.hidden = true;

  const first = $("rf-first");
  const last = $("rf-last");
  if (!first.value.trim()) return fail(t("err_name"), first);
  if (!last.value.trim()) return fail(t("err_name"), last);
  const email = $("rf-email");
  if (!email.value.trim() || !email.checkValidity()) return fail(t("err_email"), email);
  if (!$("rf-consent").checked) return fail(t("err_consent"), $("rf-consent"));

  const fd = new FormData();
  fd.set("first_name", first.value.trim());
  fd.set("last_name", last.value.trim());
  fd.set("email", email.value.trim());
  fd.set("phone", $("rf-phone").value.trim());
  fd.set("telegram", $("rf-telegram").value.trim());
  const preferred = form.querySelector('input[name="preferred_contact"]:checked:not(:disabled)');
  if (preferred) fd.set("preferred_contact", preferred.value);
  fd.set("description", composeDescription());
  fd.set("consented", "true");
  for (const file of $("rf-files").files) fd.append("attachments", file);

  const submit = $("form-submit");
  submit.disabled = true;
  submit.textContent = t("form_sending");
  try {
    await api.requests.create(fd);
    form.reset();
    state.formAreaEdited = false;
    syncFormArea();
    syncPreferredContact();
    $("form-success").hidden = false;
  } catch (err) {
    fail(err.message);
  } finally {
    submit.disabled = false;
    submit.textContent = t("form_submit");
  }
}

// ── Utilities ────────────────────────────────────────────────────────────────
// Picks the text for the current language from {ru, en}; English falls back to Russian.
function localized(text) {
  return (getLang() === "en" && text?.en) || text?.ru || "";
}

// Plain text to HTML: a blank line starts a paragraph, a single newline breaks the line.
function paragraphs(text) {
  return String(text ?? "")
    .split(/\n{2,}/)
    .filter((para) => para.trim())
    .map((para) => `<p>${esc(para).replace(/\n/g, "<br>")}</p>`)
    .join("");
}

function euro(value) {
  return "€" + value.toLocaleString(getLang() === "en" ? "en-GB" : "ru-RU");
}


