import { api } from "./api.js";

// ── Auth helpers ───────────────────────────────────────────────────────────────
function isLoggedIn() {
  return Boolean(localStorage.getItem("admin_token"));
}

function logout() {
  localStorage.removeItem("admin_token");
  location.hash = "#/login";
}

// ── Router ─────────────────────────────────────────────────────────────────────
function getRoute() {
  const hash = location.hash.replace("#", "") || "/";
  if (hash === "/login") return { page: "login" };
  if (hash === "/requests") return { page: "requests" };
  if (hash.startsWith("/requests/")) return { page: "request", id: hash.split("/")[2] };
  return { page: "projects" };
}

async function navigate() {
  const route = getRoute();

  if (route.page === "login") {
    renderLogin();
    return;
  }

  if (!isLoggedIn()) {
    location.hash = "#/login";
    return;
  }

  renderShell();

  const content = document.getElementById("admin-content");
  content.innerHTML = `<div class="loader"></div>`;

  try {
    if (route.page === "requests") {
      await renderRequests(content);
    } else if (route.page === "request") {
      await renderRequest(content, route.id);
    } else {
      await renderProjects(content);
    }
  } catch (err) {
    content.innerHTML = `<p class="error">Ошибка: ${err.message}</p>`;
  }
}

window.addEventListener("hashchange", navigate);

// ── Shell layout ───────────────────────────────────────────────────────────────
function renderShell() {
  document.getElementById("app").innerHTML = `
    <header class="admin-header">
      <span class="admin-header__logo">Панель управления</span>
      <nav class="admin-nav">
        <a href="#/projects" class="admin-nav__link">Проекты</a>
        <a href="#/requests" class="admin-nav__link">Заявки</a>
        <button class="btn btn--ghost btn--sm" id="logout-btn">Выйти</button>
      </nav>
    </header>
    <main id="admin-content" class="admin-main"></main>`;
  document.getElementById("logout-btn").addEventListener("click", logout);
}

// ── Login page ─────────────────────────────────────────────────────────────────
function renderLogin() {
  document.getElementById("app").innerHTML = `
    <div class="login-wrap">
      <div class="login-card">
        <h1 class="login-card__title">Вход в панель</h1>
        <div id="login-step-email">
          <p class="login-card__hint">Введите email администратора — вам придёт код входа.</p>
          <form id="form-email" class="form" novalidate>
            <div class="form__row">
              <input id="login-email" type="email" name="email" class="form__input"
                placeholder="admin@example.com" required autocomplete="email">
            </div>
            <div id="email-error" class="form__error hidden"></div>
            <button type="submit" class="btn btn--primary btn--full">Получить код</button>
          </form>
        </div>
        <div id="login-step-otp" class="hidden">
          <p class="login-card__hint">Код отправлен. Введите его ниже (действителен 10 минут).</p>
          <form id="form-otp" class="form" novalidate>
            <div class="form__row">
              <input id="login-otp" type="text" name="code" class="form__input form__input--otp"
                placeholder="000000" maxlength="6" inputmode="numeric" required autocomplete="one-time-code">
            </div>
            <div id="otp-error" class="form__error hidden"></div>
            <button type="submit" class="btn btn--primary btn--full">Войти</button>
          </form>
          <button class="btn-link" id="back-to-email">Изменить email</button>
        </div>
      </div>
    </div>`;

  let pendingEmail = "";

  document.getElementById("form-email").addEventListener("submit", async (e) => {
    e.preventDefault();
    const email = document.getElementById("login-email").value.trim();
    const errEl = document.getElementById("email-error");
    errEl.classList.add("hidden");
    try {
      await api.auth.requestOTP(email);
      pendingEmail = email;
      document.getElementById("login-step-email").classList.add("hidden");
      document.getElementById("login-step-otp").classList.remove("hidden");
    } catch (err) {
      errEl.textContent = err.message;
      errEl.classList.remove("hidden");
    }
  });

  document.getElementById("form-otp").addEventListener("submit", async (e) => {
    e.preventDefault();
    const code = document.getElementById("login-otp").value.trim();
    const errEl = document.getElementById("otp-error");
    errEl.classList.add("hidden");
    try {
      const res = await api.auth.verifyOTP(pendingEmail, code);
      localStorage.setItem("admin_token", res.data.token);
      location.hash = "#/projects";
    } catch (err) {
      errEl.textContent = "Неверный или просроченный код.";
      errEl.classList.remove("hidden");
    }
  });

  document.getElementById("back-to-email").addEventListener("click", () => {
    document.getElementById("login-step-otp").classList.add("hidden");
    document.getElementById("login-step-email").classList.remove("hidden");
  });
}

// ── Projects page ──────────────────────────────────────────────────────────────
async function renderProjects(content) {
  const result = await api.projects.list();
  const projects = result.data ?? [];

  content.innerHTML = `
    <div class="admin-section">
      <div class="admin-section__head">
        <h2>Проекты</h2>
        <button class="btn btn--primary btn--sm" id="btn-add-project">+ Добавить проект</button>
      </div>
      <div id="project-list" class="admin-list">
        ${projects.length === 0 ? `<p class="empty">Проектов пока нет.</p>` : ""}
      </div>
    </div>
    <div id="project-editor" class="admin-editor hidden"></div>`;

  const list = document.getElementById("project-list");
  for (const p of projects) {
    const row = document.createElement("div");
    row.className = "admin-list__row";
    row.innerHTML = `
      <span>
        <span class="admin-list__name">${esc(p.title)}</span>
        <span class="admin-list__meta">${PROJECT_CATEGORIES[p.category] ?? PROJECT_CATEGORIES.residential}</span>
      </span>
      <div class="admin-list__actions">
        <button class="btn btn--ghost btn--sm" data-edit="${p.id}">Редактировать</button>
        <button class="btn btn--danger btn--sm" data-delete="${p.id}">Удалить</button>
      </div>`;
    list.appendChild(row);
  }

  document.getElementById("btn-add-project").addEventListener("click", () => openProjectEditor(null));

  list.addEventListener("click", async (e) => {
    const editId = e.target.dataset.edit;
    const deleteId = e.target.dataset.delete;
    if (editId) {
      // The list has no photos, so load the full project for the editor.
      try {
        const full = await api.projects.get(editId);
        openProjectEditor(full.data);
      } catch (err) {
        alert("Ошибка загрузки проекта: " + err.message);
      }
    }
    if (deleteId) {
      if (!confirm("Удалить проект и все его фотографии?")) return;
      try {
        await api.projects.delete(deleteId);
        await renderProjects(content);
      } catch (err) {
        alert("Ошибка удаления: " + err.message);
      }
    }
  });
}

// Work directions shown as filters on the public site.
const PROJECT_CATEGORIES = {
  residential: "Жилой интерьер (Residential)",
  commercial: "HoReCa",
};

function openProjectEditor(project) {
  const editor = document.getElementById("project-editor");
  const isNew = !project;
  editor.classList.remove("hidden");
  editor.innerHTML = `
    <div class="editor-card">
      <h3>${isNew ? "Новый проект" : "Редактировать проект"}</h3>
      <form id="project-form" class="form" novalidate>
        <div class="form__row">
          <label class="form__label">Название *</label>
          <input name="title" class="form__input" value="${esc(project?.title ?? "")}" required>
        </div>
        <div class="form__row">
          <label class="form__label">Описание (до 500 символов)</label>
          <textarea name="description" class="form__textarea" rows="3" maxlength="500">${esc(project?.description ?? "")}</textarea>
        </div>
        <div class="form__row">
          <label class="form__label">Направление</label>
          <select name="category" class="form__input">
            ${Object.entries(PROJECT_CATEGORIES).map(([value, label]) => `
              <option value="${value}" ${(project?.category ?? "residential") === value ? "selected" : ""}>${label}</option>`).join("")}
          </select>
        </div>
        ${isNew ? `
        <div class="form__row">
          <label class="form__label">Фотографии (можно выбрать несколько, в том числе в несколько заходов)</label>
          <input type="file" id="new-media-input" multiple accept="image/*" class="form__file">
          <p class="media-hint">${PHOTO_HINT}</p>
          <div class="media-grid" id="new-media-list"></div>
        </div>` : ""}
        <div id="project-error" class="form__error hidden"></div>
        <div class="editor-card__actions">
          <button type="submit" class="btn btn--primary">${isNew ? "Создать" : "Сохранить"}</button>
          <button type="button" class="btn btn--ghost" id="cancel-edit">Отмена</button>
        </div>
      </form>
      ${isNew ? "" : `<div class="media-manager" id="media-manager"></div>`}
    </div>`;

  const pending = isNew ? setupPendingPhotos() : null;

  document.getElementById("cancel-edit").addEventListener("click", () => {
    pending?.release();
    editor.classList.add("hidden");
  });

  document.getElementById("project-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = document.getElementById("project-error");
    errEl.classList.add("hidden");
    const form = e.target;
    const title = form.querySelector("[name=title]").value;
    const description = form.querySelector("[name=description]").value;
    const category = form.querySelector("[name=category]").value;
    try {
      if (isNew) {
        const fd = new FormData();
        fd.set("title", title);
        fd.set("description", description);
        fd.set("category", category);
        for (const file of pending.files()) fd.append("media", file);
        if (pending.files().length) fd.set("cover_index", String(pending.coverIndex()));
        await api.projects.create(fd);
        pending.release();
      } else {
        // The API replaces every field, so pass the current cover to keep it.
        await api.projects.update(project.id, {
          title, description, category, cover_media_id: project.cover_media_id ?? null,
        });
      }
      editor.classList.add("hidden");
      const content = document.getElementById("admin-content");
      await renderProjects(content);
    } catch (err) {
      errEl.textContent = err.message;
      errEl.classList.remove("hidden");
    }
  });

  if (!isNew) mountMediaManager(project);
}

const PHOTO_HINT =
  "Порядок фото — это порядок в галерее проекта на сайте; меняйте его стрелками ← →. " +
  "Обложка показывается в карточке проекта и первой в его окне.";

// One photo tile with order and cover controls; `index` is its 0-based position.
function photoTile({ src, alt, index, total, isCover }) {
  return `
    <div class="media-thumb ${isCover ? "media-thumb--cover" : ""}" data-index="${index}">
      <div class="media-thumb__frame">
        <img src="${src}" alt="${esc(alt)}" class="media-thumb__img">
        <span class="media-thumb__pos">${index + 1}</span>
        ${isCover ? `<span class="media-thumb__badge">Обложка</span>` : ""}
      </div>
      <div class="media-thumb__actions">
        <button type="button" class="btn btn--sm btn--ghost" data-move="-1" ${index === 0 ? "disabled" : ""}
          title="Раньше" aria-label="Сдвинуть фото ${index + 1} раньше">←</button>
        <button type="button" class="btn btn--sm btn--ghost" data-move="1" ${index === total - 1 ? "disabled" : ""}
          title="Позже" aria-label="Сдвинуть фото ${index + 1} позже">→</button>
        <button type="button" class="btn btn--sm btn--ghost" data-cover ${isCover ? "disabled" : ""}
          aria-pressed="${isCover}">Обложка</button>
        <button type="button" class="btn btn--sm btn--danger" data-remove
          title="Удалить" aria-label="Удалить фото ${index + 1}">&times;</button>
      </div>
    </div>`;
}

// Photos chosen for a new project: ordered and with a cover before anything is uploaded.
function setupPendingPhotos() {
  const input = document.getElementById("new-media-input");
  const list = document.getElementById("new-media-list");
  let items = []; // { file, url }
  let cover = 0;

  const render = () => {
    list.innerHTML = items.map((item, i) => photoTile({
      src: item.url, alt: item.file.name, index: i, total: items.length, isCover: i === cover,
    })).join("");
  };

  input.addEventListener("change", () => {
    for (const file of input.files) items.push({ file, url: URL.createObjectURL(file) });
    input.value = ""; // allows picking more files, including the same ones again
    render();
  });

  list.addEventListener("click", (e) => {
    const button = e.target.closest("button");
    if (!button) return;
    const i = Number(button.closest(".media-thumb").dataset.index);
    if (button.dataset.move) {
      const j = i + Number(button.dataset.move);
      [items[i], items[j]] = [items[j], items[i]];
      if (cover === i) cover = j;
      else if (cover === j) cover = i;
    } else if ("cover" in button.dataset) {
      cover = i;
    } else if ("remove" in button.dataset) {
      URL.revokeObjectURL(items[i].url);
      items.splice(i, 1);
      if (i < cover || cover >= items.length) cover = Math.max(0, cover - 1);
    }
    render();
  });

  return {
    files: () => items.map((item) => item.file),
    coverIndex: () => cover,
    release: () => items.forEach((item) => URL.revokeObjectURL(item.url)),
  };
}

// Photos of an existing project: every change is saved right away.
function mountMediaManager(project) {
  const box = document.getElementById("media-manager");
  const media = project.media ?? [];
  box.innerHTML = `
    <h4>Фотографии</h4>
    <p class="media-hint">${PHOTO_HINT}</p>
    <div class="media-grid" id="media-grid">
      ${media.length ? media.map((m, i) => photoTile({
        src: api.projects.mediaUrl(project.id, m.id), alt: m.filename,
        index: i, total: media.length, isCover: m.id === project.cover_media_id,
      })).join("") : `<p class="empty">Фото пока нет.</p>`}
    </div>
    <div class="form__row">
      <label class="form__label">Добавить фото (можно выбрать несколько, они встанут в конец)</label>
      <input type="file" id="add-media-input" accept="image/*" multiple class="form__file">
      <div class="upload-actions">
        <button class="btn btn--ghost btn--sm" id="add-media-btn">Загрузить</button>
        <span id="upload-progress" class="upload-progress hidden"></span>
      </div>
    </div>`;

  // Reload the project (keeping the same object, which the form's save uses) and redraw.
  const refresh = async () => {
    const full = await api.projects.get(project.id);
    Object.assign(project, full.data, { media: full.data.media ?? [] });
    mountMediaManager(project);
  };

  const setCover = (coverId) => api.projects.update(project.id, {
    title: project.title,
    description: project.description,
    category: project.category,
    cover_media_id: coverId,
  });

  document.getElementById("media-grid").addEventListener("click", async (e) => {
    const button = e.target.closest("button");
    if (!button) return;
    const i = Number(button.closest(".media-thumb").dataset.index);
    const target = media[i];
    try {
      if (button.dataset.move) {
        const ids = media.map((m) => m.id);
        const j = i + Number(button.dataset.move);
        [ids[i], ids[j]] = [ids[j], ids[i]];
        await api.projects.reorderMedia(project.id, ids);
      } else if ("cover" in button.dataset) {
        await setCover(target.id);
      } else if ("remove" in button.dataset) {
        if (!confirm("Удалить фото?")) return;
        await api.projects.deleteMedia(project.id, target.id);
        // A deleted cover would leave the card without a picture: use the first remaining photo.
        if (target.id === project.cover_media_id) {
          const next = media.find((m) => m.id !== target.id);
          await setCover(next ? next.id : null);
        }
      }
      await refresh();
    } catch (err) {
      alert("Ошибка: " + err.message);
      await refresh();
    }
  });

  document.getElementById("add-media-btn").addEventListener("click", async () => {
    const input = document.getElementById("add-media-input");
    if (!input.files.length) return;

    const files = Array.from(input.files);
    const btn = document.getElementById("add-media-btn");
    const progress = document.getElementById("upload-progress");
    btn.disabled = true;
    progress.classList.remove("hidden");

    const errors = [];
    for (let i = 0; i < files.length; i++) {
      progress.textContent = `${i + 1} / ${files.length}`;
      const fd = new FormData();
      fd.append("media", files[i]);
      try {
        await api.projects.addMedia(project.id, fd);
      } catch (err) {
        errors.push(`${files[i].name}: ${err.message}`);
      }
    }

    // A project without a cover gets the first photo as its cover.
    if (!project.cover_media_id) {
      const full = await api.projects.get(project.id);
      const first = full.data.media?.[0];
      if (first) await setCover(first.id).catch(() => {});
    }
    await refresh();

    if (errors.length) {
      alert("Ошибка при загрузке некоторых файлов:\n" + errors.join("\n"));
    }
  });
}

// ── Requests page ──────────────────────────────────────────────────────────────
async function renderRequests(content) {
  const result = await api.requests.list();
  const requests = result.data ?? [];

  content.innerHTML = `
    <div class="admin-section">
      <div class="admin-section__head"><h2>Заявки клиентов</h2></div>
      <div class="admin-list">
        ${requests.length === 0 ? `<p class="empty">Заявок пока нет.</p>` : ""}
        ${requests.map((r) => `
          <div class="admin-list__row">
            <div>
              <strong>${esc(r.first_name)} ${esc(r.last_name)}</strong>
              <span class="admin-list__meta">${esc(r[r.preferred_contact] || r.contact)} — ${new Date(r.created_at).toLocaleDateString("ru-RU")}</span>
            </div>
            <a href="#/requests/${r.id}" class="btn btn--ghost btn--sm">Открыть</a>
          </div>`).join("")}
      </div>
    </div>`;
}

const REQUEST_CONTACT_LABELS = { email: "E-mail", phone: "Телефон", telegram: "Telegram" };

function requestContactRows(r) {
  // Requests created before the separate fields existed only have `contact`.
  if (!r.email) return `<dt>Контакт</dt><dd>${esc(r.contact)}</dd>`;

  return Object.entries(REQUEST_CONTACT_LABELS)
    .filter(([key]) => r[key])
    .map(([key, label]) => `
      <dt>${label}</dt>
      <dd>${esc(r[key])}${r.preferred_contact === key
        ? ` <span class="request-detail__badge">предпочтительный</span>`
        : ""}</dd>`)
    .join("");
}

async function renderRequest(content, id) {
  const result = await api.requests.get(id);
  const r = result.data;

  const attachments = (r.attachments ?? [])
    .map((a) => {
      const url = api.requests.attachmentUrl(r.id, a.id);
      const token = localStorage.getItem("admin_token");
      if (a.content_type === "application/pdf") {
        return `<a href="${url}" target="_blank" class="btn btn--ghost btn--sm">Скачать PDF: ${esc(a.filename)}</a>`;
      }
      // Fetch image with auth header via JS (workaround: use direct URL for now)
      return `<img src="${url}" alt="${esc(a.filename)}" class="attachment-img" style="max-width:200px;margin:4px" onerror="this.style.display='none'">`;
    })
    .join("");

  content.innerHTML = `
    <div class="admin-section">
      <a href="#/requests" class="btn btn--ghost">&larr; Назад к заявкам</a>
      <div class="request-detail">
        <h2>${esc(r.first_name)} ${esc(r.last_name)}</h2>
        <dl class="request-detail__meta">
          ${requestContactRows(r)}
          <dt>Дата</dt><dd>${new Date(r.created_at).toLocaleString("ru-RU")}</dd>
          <dt>Согласие на обработку данных</dt><dd>${r.consented ? "Да" : "Нет"}</dd>
        </dl>
        <h3>Описание проекта</h3>
        <p class="request-detail__description">${esc(r.description)}</p>
        ${attachments ? `<h3>Вложения</h3><div class="attachments">${attachments}</div>` : ""}
      </div>
    </div>`;
}

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
