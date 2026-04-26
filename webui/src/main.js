import { api } from "./api.js";

// ── Router ────────────────────────────────────────────────────────────────────
function getRoute() {
  const hash = location.hash.replace("#", "") || "/";
  const match = hash.match(/^\/project\/([a-f0-9-]+)/);
  if (match) return { page: "project", id: match[1] };
  if (hash === "/contacts") return { page: "contacts" };
  return { page: "portfolio" };
}

async function navigate() {
  const route = getRoute();
  const main = document.getElementById("main");
  main.innerHTML = `<div class="loader"></div>`;
  try {
    if (route.page === "project") {
      await renderProject(main, route.id);
    } else if (route.page === "contacts") {
      await renderContacts(main);
    } else {
      await renderPortfolio(main);
    }
  } catch (err) {
    main.innerHTML = `<p class="error">Ошибка загрузки: ${err.message}</p>`;
  }
}

window.addEventListener("hashchange", navigate);

// ── Portfolio page ─────────────────────────────────────────────────────────────
async function renderPortfolio(main) {
  const result = await api.projects.list();
  const projects = result.data ?? [];

  if (projects.length === 0) {
    main.innerHTML = `<p class="empty">Проекты появятся здесь совсем скоро.</p>`;
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

// ── Project detail page ────────────────────────────────────────────────────────
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
      <a href="#/" class="btn btn--ghost">&larr; Назад</a>
      <h1 class="project__title">${esc(p.title)}</h1>
      ${p.description ? `<p class="project__desc">${esc(p.description)}</p>` : ""}
      <div class="gallery">${mediaHTML || "<p class='empty'>Фото будут добавлены.</p>"}</div>
      <div class="project__cta">
        <button class="btn btn--primary" id="open-request">Оставить заявку</button>
      </div>
    </div>`;

  document.getElementById("open-request")?.addEventListener("click", openRequestModal);
}

// ── Contacts page ──────────────────────────────────────────────────────────────
async function renderContacts(main) {
  const result = await api.contacts.get();
  const c = result.data;

  main.innerHTML = `
    <section class="contacts">
      <h1 class="contacts__title">Контакты</h1>
      <ul class="contacts__list">
        ${c.telegram  ? `<li><a href="https://t.me/${c.telegram.replace("@","")}"  target="_blank" rel="noopener" class="contact-link contact-link--telegram">Telegram: ${esc(c.telegram)}</a></li>`  : ""}
        ${c.instagram ? `<li><a href="https://instagram.com/${c.instagram.replace("@","")}" target="_blank" rel="noopener" class="contact-link contact-link--instagram">Instagram: ${esc(c.instagram)}</a></li>` : ""}
        ${c.email     ? `<li><a href="mailto:${c.email}" class="contact-link contact-link--email">Email: ${esc(c.email)}</a></li>` : ""}
      </ul>
      <div class="contacts__cta">
        <button class="btn btn--primary" id="open-request">Оставить заявку на проект</button>
      </div>
    </section>`;

  document.getElementById("open-request")?.addEventListener("click", openRequestModal);
}

// ── Request modal ──────────────────────────────────────────────────────────────
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
      <button id="close-modal" class="modal__close" aria-label="Закрыть">&times;</button>
      <h2 class="modal__title">Заявка на проект</h2>
      <form id="request-form" class="form" enctype="multipart/form-data" novalidate>
        <div class="form__row">
          <label class="form__label" for="rf-first">Имя *</label>
          <input id="rf-first" name="first_name" class="form__input" required>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-last">Фамилия *</label>
          <input id="rf-last" name="last_name" class="form__input" required>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-contact">Контакт (телефон / e-mail / Telegram) *</label>
          <input id="rf-contact" name="contact" class="form__input" required>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-desc">Описание проекта *</label>
          <textarea id="rf-desc" name="description" class="form__textarea" rows="4" required></textarea>
        </div>
        <div class="form__row">
          <label class="form__label" for="rf-files">
            Приложите до 10 фото или 1 PDF с чертежами
          </label>
          <input id="rf-files" name="attachments" type="file"
            accept="image/jpeg,image/png,image/webp,image/gif,application/pdf"
            multiple class="form__file">
        </div>
        <div class="form__row form__row--consent">
          <label class="form__checkbox-label">
            <input type="checkbox" id="rf-consent" name="consented" class="form__checkbox" required>
            Я даю согласие на
            <button type="button" class="btn-link" id="show-consent">обработку персональных данных</button>
          </label>
        </div>
        <div id="consent-text" class="consent-text hidden">
          <p>Нажимая кнопку «Отправить», вы соглашаетесь с тем, что предоставленные вами персональные данные (имя, контактная информация) будут обработаны исключительно в целях ответа на ваш запрос. Данные не передаются третьим лицам и хранятся в соответствии с законодательством о защите персональных данных.</p>
        </div>
        <div id="form-error" class="form__error hidden"></div>
        <button type="submit" class="btn btn--primary btn--full">Отправить заявку</button>
      </form>
    </div>`;
}

async function submitRequest(e) {
  e.preventDefault();
  const form = e.target;
  const errEl = form.querySelector("#form-error");
  errEl.classList.add("hidden");

  if (!form.querySelector("#rf-consent").checked) {
    showFormError(errEl, "Необходимо согласие на обработку персональных данных.");
    return;
  }

  const fd = new FormData(form);
  fd.set("consented", "true");

  try {
    await api.requests.create(fd);
    form.closest(".modal").innerHTML = `
      <div class="modal__success">
        <h2>Заявка отправлена!</h2>
        <p>Мы свяжемся с вами в ближайшее время.</p>
        <button class="btn btn--primary" onclick="document.getElementById('modal-overlay').classList.add('hidden')">Закрыть</button>
      </div>`;
  } catch (err) {
    showFormError(errEl, err.message);
  }
}

function showFormError(el, msg) {
  el.textContent = msg;
  el.classList.remove("hidden");
}

// ── Consent text toggle ────────────────────────────────────────────────────────
document.addEventListener("click", (e) => {
  if (e.target.id === "show-consent") {
    document.getElementById("consent-text")?.classList.toggle("hidden");
  }
});

// ── Utilities ──────────────────────────────────────────────────────────────────
function esc(str) {
  return String(str ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

// Bootstrap
navigate();
