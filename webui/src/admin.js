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
      <span class="admin-list__name">${esc(p.title)}</span>
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
      const p = projects.find((x) => x.id === editId);
      openProjectEditor(p);
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

function openProjectEditor(project) {
  const editor = document.getElementById("project-editor");
  const isNew = !project;
  editor.classList.remove("hidden");
  editor.innerHTML = `
    <div class="editor-card">
      <h3>${isNew ? "Новый проект" : "Редактировать проект"}</h3>
      <form id="project-form" class="form" enctype="multipart/form-data" novalidate>
        <div class="form__row">
          <label class="form__label">Название *</label>
          <input name="title" class="form__input" value="${esc(project?.title ?? "")}" required>
        </div>
        <div class="form__row">
          <label class="form__label">Описание (до 500 символов)</label>
          <textarea name="description" class="form__textarea" rows="3" maxlength="500">${esc(project?.description ?? "")}</textarea>
        </div>
        ${isNew ? `
        <div class="form__row">
          <label class="form__label">Фотографии</label>
          <input type="file" name="media" multiple accept="image/*" class="form__file">
        </div>` : ""}
        <div id="project-error" class="form__error hidden"></div>
        <div class="editor-card__actions">
          <button type="submit" class="btn btn--primary">${isNew ? "Создать" : "Сохранить"}</button>
          <button type="button" class="btn btn--ghost" id="cancel-edit">Отмена</button>
        </div>
      </form>
      ${!isNew ? renderMediaManager(project) : ""}
    </div>`;

  document.getElementById("cancel-edit").addEventListener("click", () => {
    editor.classList.add("hidden");
  });

  document.getElementById("project-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = document.getElementById("project-error");
    errEl.classList.add("hidden");
    const form = e.target;
    try {
      if (isNew) {
        const fd = new FormData(form);
        await api.projects.create(fd);
      } else {
        const title = form.querySelector("[name=title]").value;
        const description = form.querySelector("[name=description]").value;
        await api.projects.update(project.id, { title, description });
      }
      editor.classList.add("hidden");
      const content = document.getElementById("admin-content");
      await renderProjects(content);
    } catch (err) {
      errEl.textContent = err.message;
      errEl.classList.remove("hidden");
    }
  });

  if (!isNew) {
    setupMediaManager(project);
  }
}

function renderMediaManager(project) {
  const media = project.media ?? [];
  return `
    <div class="media-manager">
      <h4>Фотографии</h4>
      <div class="media-grid" id="media-grid-${project.id}">
        ${media.map((m) => `
          <div class="media-thumb" data-media-id="${m.id}">
            <img src="${api.projects.mediaUrl(project.id, m.id)}" alt="${esc(m.filename)}" class="media-thumb__img">
            <div class="media-thumb__actions">
              <button class="btn btn--sm btn--ghost" data-set-cover="${m.id}">Обложка</button>
              <button class="btn btn--sm btn--danger" data-del-media="${m.id}">&times;</button>
            </div>
          </div>`).join("")}
      </div>
      <div class="form__row">
        <label class="form__label">Добавить фото</label>
        <input type="file" id="add-media-input" accept="image/*" class="form__file">
        <button class="btn btn--ghost btn--sm" id="add-media-btn">Загрузить</button>
      </div>
    </div>`;
}

function setupMediaManager(project) {
  const grid = document.getElementById(`media-grid-${project.id}`);
  if (!grid) return;

  grid.addEventListener("click", async (e) => {
    const mediaId = e.target.dataset.delMedia;
    const coverId = e.target.dataset.setCover;

    if (mediaId) {
      if (!confirm("Удалить фото?")) return;
      try {
        await api.projects.deleteMedia(project.id, mediaId);
        const thumb = grid.querySelector(`[data-media-id="${mediaId}"]`);
        thumb?.remove();
      } catch (err) {
        alert("Ошибка: " + err.message);
      }
    }

    if (coverId) {
      try {
        await api.projects.update(project.id, {
          title: project.title,
          description: project.description,
          cover_media_id: coverId,
        });
        alert("Обложка обновлена.");
      } catch (err) {
        alert("Ошибка: " + err.message);
      }
    }
  });

  document.getElementById("add-media-btn").addEventListener("click", async () => {
    const input = document.getElementById("add-media-input");
    if (!input.files.length) return;
    const fd = new FormData();
    fd.append("media", input.files[0]);
    try {
      await api.projects.addMedia(project.id, fd);
      // Reload editor
      const full = await api.projects.get(project.id);
      openProjectEditor(full.data);
    } catch (err) {
      alert("Ошибка загрузки: " + err.message);
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
              <span class="admin-list__meta">${esc(r.contact)} — ${new Date(r.created_at).toLocaleDateString("ru-RU")}</span>
            </div>
            <a href="#/requests/${r.id}" class="btn btn--ghost btn--sm">Открыть</a>
          </div>`).join("")}
      </div>
    </div>`;
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
          <dt>Контакт</dt><dd>${esc(r.contact)}</dd>
          <dt>Дата</dt><dd>${new Date(r.created_at).toLocaleString("ru-RU")}</dd>
          <dt>Согласие на обработку данных</dt><dd>${r.consented ? "Да" : "Нет"}</dd>
        </dl>
        <h3>Описание проекта</h3>
        <p>${esc(r.description)}</p>
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
