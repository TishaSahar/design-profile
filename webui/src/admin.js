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
      <div id="project-list">
        ${projects.length === 0 ? `<p class="empty">Проектов пока нет.</p>` : ""}
      </div>
    </div>
    <div id="project-editor" class="admin-editor hidden"></div>`;

  // One list per category, in the order the projects appear on the site.
  const byCategory = Object.keys(PROJECT_CATEGORIES).map((category) => ({
    category,
    items: projects.filter((p) => (p.category in PROJECT_CATEGORIES ? p.category : "residential") === category),
  }));

  projectPositions = new Map();
  for (const { items } of byCategory) items.forEach((p, i) => projectPositions.set(p.id, i + 1));

  const list = document.getElementById("project-list");
  if (projects.length) {
    list.innerHTML = byCategory.map(({ category, items }) => `
      <section class="project-group">
        <h3 class="project-group__title">${PROJECT_CATEGORIES[category]} <span class="project-group__count">${items.length}</span></h3>
        <div class="admin-list" data-category="${category}">
          ${items.length ? items.map((p, i) => `
            <div class="admin-list__row ${items.length > 1 ? "admin-list__row--sortable" : ""}" data-id="${p.id}">
              <span class="admin-list__title" data-tooltip="${items.length > 1
                ? "Нажмите, чтобы открыть превью. Перетащите, чтобы изменить порядок на сайте"
                : "Нажмите, чтобы открыть превью"}">
                ${items.length > 1 ? `<button type="button" class="drag-handle"
                  aria-label="Переместить «${esc(p.title)}»: стрелки вверх и вниз">⋮⋮</button>` : ""}
                <span class="admin-list__pos">${i + 1}</span>
                <button type="button" class="admin-list__name admin-list__open">${esc(p.title)}</button>
              </span>
              <div class="admin-list__actions">
                <button class="btn btn--ghost btn--sm" data-edit="${p.id}">Редактировать</button>
                <button class="btn btn--danger btn--sm" data-delete="${p.id}">Удалить</button>
              </div>
            </div>`).join("") : `<p class="empty">Проектов этого направления пока нет.</p>`}
        </div>
      </section>`).join("");
  }

  document.getElementById("btn-add-project").addEventListener("click", () => openProjectEditor(null));

  for (const group of list.querySelectorAll(".admin-list[data-category]")) {
    makeSortable(group, {
      item: ".admin-list__row",
      pos: ".admin-list__pos",
      onChange: async (ids) => {
        try {
          await api.projects.reorder(group.dataset.category, ids);
        } catch (err) {
          alert("Ошибка: " + err.message);
        }
        await renderProjects(content);
      },
    });
  }

  list.addEventListener("click", async (e) => {
    // A click on a row (not on its buttons) opens the preview. After a drag the click
    // lands on the list itself, not on a row, so dragging does not open it.
    const row = e.target.closest(".admin-list__row[data-id]");
    if (row && (e.target.closest(".admin-list__open") || !e.target.closest("button"))) {
      try {
        const full = await api.projects.get(row.dataset.id);
        openProjectPreview(savedProjectPreview(full.data, full.data, projectPositions.get(row.dataset.id) ?? 1));
      } catch (err) {
        alert("Ошибка загрузки проекта: " + err.message);
      }
      return;
    }
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

// Position of each project within its direction (1-based), from the last list render.
let projectPositions = new Map();

// Preview message for a saved project: its photos as published (the cover first, then
// the rest in order) with `fields` (texts and category, see readProjectFields), possibly unsaved edits.
function savedProjectPreview(project, fields, number) {
  const media = project.media ?? [];
  const coverId = project.cover_media_id ?? media[0]?.id;
  return {
    project: fields,
    number,
    coverUrl: coverId ? api.projects.mediaUrl(project.id, coverId) : null,
    galleryUrls: media.filter((m) => m.id !== coverId).map((m) => api.projects.mediaUrl(project.id, m.id)),
  };
}

// Shows the project as visitors will see it: preview.html draws it with the site's
// own markup and styles. `data` is the message preview.js expects.
function openProjectPreview(data) {
  document.getElementById("project-preview")?.remove();
  const overlay = document.createElement("div");
  overlay.id = "project-preview";
  overlay.className = "preview-overlay";
  overlay.innerHTML = `
    <div class="preview-panel" role="dialog" aria-modal="true" aria-label="Превью проекта">
      <div class="preview-panel__bar">
        <strong>Превью: так проект будет выглядеть на сайте</strong>
        <div class="preview-panel__tools">
          <div class="segmented" role="group" aria-label="Ширина экрана">
            <button type="button" class="segmented__btn" data-width="desktop" aria-pressed="true">Компьютер</button>
            <button type="button" class="segmented__btn" data-width="phone" aria-pressed="false">Телефон</button>
          </div>
          <div class="segmented" role="group" aria-label="Язык сайта">
            <button type="button" class="segmented__btn" data-lang="ru" aria-pressed="true">RU</button>
            <button type="button" class="segmented__btn" data-lang="en" aria-pressed="false">EN</button>
          </div>
          <button type="button" class="btn btn--ghost btn--sm" data-close-preview>Закрыть</button>
        </div>
      </div>
      <div class="preview-panel__stage">
        <iframe class="preview-panel__frame" src="./preview.html" title="Превью проекта"></iframe>
      </div>
    </div>`;
  document.body.appendChild(overlay);
  document.body.classList.add("is-locked");

  const frame = overlay.querySelector("iframe");
  const stage = overlay.querySelector(".preview-panel__stage");
  // "Компьютер" lays the page out at a typical desktop width and scales it down to fit,
  // so the admin sees the real 3-column grid even in a narrow window.
  const DESKTOP_WIDTH = 1280;
  const PHONE_WIDTH = 390;
  let phone = false;
  const fit = () => {
    const width = phone ? Math.min(PHONE_WIDTH, stage.clientWidth) : DESKTOP_WIDTH;
    const scale = phone ? 1 : Math.min(1, stage.clientWidth / DESKTOP_WIDTH);
    frame.style.width = `${width}px`;
    frame.style.height = `${stage.clientHeight / scale}px`;
    frame.style.transform = `translateX(-50%) scale(${scale})`;
  };
  fit();
  window.addEventListener("resize", fit);

  let lang = "ru";
  const send = () => frame.contentWindow.postMessage({ type: "project-preview", ...data, lang }, location.origin);
  const onMessage = (e) => {
    if (e.source === frame.contentWindow && e.data?.type === "project-preview-ready") send();
  };
  window.addEventListener("message", onMessage);

  const close = () => {
    window.removeEventListener("message", onMessage);
    window.removeEventListener("resize", fit);
    document.removeEventListener("keydown", onKey);
    document.body.classList.remove("is-locked");
    overlay.remove();
  };
  const onKey = (e) => { if (e.key === "Escape") close(); };
  document.addEventListener("keydown", onKey);

  overlay.addEventListener("click", (e) => {
    if (e.target === overlay || e.target.closest("[data-close-preview]")) close();
    const widthBtn = e.target.closest("[data-width]");
    if (widthBtn) {
      overlay.querySelectorAll("[data-width]").forEach((b) => b.setAttribute("aria-pressed", String(b === widthBtn)));
      phone = widthBtn.dataset.width === "phone";
      frame.classList.toggle("preview-panel__frame--phone", phone);
      fit();
    }
    const langBtn = e.target.closest("[data-lang]");
    if (langBtn) {
      overlay.querySelectorAll("[data-lang]").forEach((b) => b.setAttribute("aria-pressed", String(b === langBtn)));
      lang = langBtn.dataset.lang;
      send();
    }
  });
  overlay.querySelector("[data-close-preview]").focus();
}

// Drag-and-drop ordering of the `item` elements (each with data-id) inside `container`:
// project rows in a list, photo tiles in a wrapping grid.
// A mouse can grab an item anywhere except its buttons; touch uses the ⋮⋮ handle, so the
// page still scrolls; the keyboard moves the focused handle with the arrow keys.
// `pos` is the selector of the position number inside an item, renumbered while dragging.
// `onChange(ids)` gets the new order once a move ends with the order changed.
function makeSortable(container, { item, pos, onChange }) {
  // Items are looked up on every event: the container may be redrawn or still empty.
  const items = () => [...container.querySelectorAll(`${item}[data-id]`)];
  const order = () => items().map((el) => el.dataset.id);
  const renumber = () => items().forEach((el, i) => { el.querySelector(pos).textContent = i + 1; });

  let drag = null; // { el, pointerId, before }

  container.addEventListener("pointerdown", (e) => {
    if (e.button !== 0) return;
    const el = e.target.closest(`${item}[data-id]`);
    if (!el || items().length < 2) return;
    const onHandle = e.target.closest(".drag-handle");
    const plainButton = e.target.closest("button:not(.admin-list__open), a, input");
    if (!onHandle && (e.pointerType !== "mouse" || plainButton)) return;
    e.preventDefault(); // no text selection while dragging
    drag = { el, pointerId: e.pointerId, before: order().join(), x: e.clientX, y: e.clientY, started: false };
  });

  container.addEventListener("pointermove", (e) => {
    if (!drag || e.pointerId !== drag.pointerId) return;
    // A press that barely moves is a click (e.g. opening the preview), not a drag.
    if (!drag.started) {
      if (Math.hypot(e.clientX - drag.x, e.clientY - drag.y) < 5) return;
      drag.started = true;
      container.setPointerCapture(e.pointerId);
      drag.el.classList.add("is-dragging");
      container.classList.add("is-sorting");
    }
    // Over another item: take its place. Moving forward the item goes after it,
    // moving back before it, so the pointer ends up over the dragged item and nothing flickers.
    const target = items().find((el) => {
      if (el === drag.el) return false;
      const box = el.getBoundingClientRect();
      return e.clientX >= box.left && e.clientX <= box.right && e.clientY >= box.top && e.clientY <= box.bottom;
    });
    if (!target) return;
    const list = items();
    if (list.indexOf(drag.el) < list.indexOf(target)) target.after(drag.el);
    else target.before(drag.el);
    renumber();
  });

  const finish = (e) => {
    if (!drag || e.pointerId !== drag.pointerId) return;
    drag.el.classList.remove("is-dragging");
    container.classList.remove("is-sorting");
    const changed = order().join() !== drag.before;
    drag = null;
    if (changed) onChange(order());
  };
  container.addEventListener("pointerup", finish);
  container.addEventListener("pointercancel", finish);

  container.addEventListener("keydown", (e) => {
    const handle = e.target.closest(".drag-handle");
    const back = e.key === "ArrowUp" || e.key === "ArrowLeft";
    const forward = e.key === "ArrowDown" || e.key === "ArrowRight";
    if (!handle || (!back && !forward)) return;
    e.preventDefault();
    const el = handle.closest(item);
    const list = items();
    const sibling = list[list.indexOf(el) + (back ? -1 : 1)];
    if (!sibling) return;
    if (back) sibling.before(el); else sibling.after(el);
    renumber();
    handle.focus();
    onChange(order());
  });
}

function openProjectEditor(project) {
  const editor = document.getElementById("project-editor");
  const isNew = !project;
  editor.classList.remove("hidden");
  editor.innerHTML = `
    <div class="editor-card">
      <h3>${isNew ? "Новый проект" : "Редактировать проект"}</h3>
      <form id="project-form" class="form" novalidate>
        <div class="form__row">
          <label class="form__label" for="project-title">Название *</label>
          <input id="project-title" name="title" class="form__input" value="${esc(project?.title ?? "")}" required>
        </div>
        <div class="form__row">
          <label class="form__label" for="project-title-en">Название на английском</label>
          <input id="project-title-en" name="title_en" class="form__input" lang="en" maxlength="255"
            value="${esc(project?.title_en ?? "")}" aria-describedby="project-en-hint">
        </div>
        <div class="form__row">
          <label class="form__label" for="project-description">Описание (до 1000 символов)</label>
          <textarea id="project-description" name="description" class="form__textarea" rows="3" maxlength="1000">${esc(project?.description ?? "")}</textarea>
        </div>
        <div class="form__row">
          <label class="form__label" for="project-description-en">Описание на английском (до 1000 символов)</label>
          <textarea id="project-description-en" name="description_en" class="form__textarea" rows="3" maxlength="1000"
            lang="en" aria-describedby="project-en-hint">${esc(project?.description_en ?? "")}</textarea>
          <p class="form__hint" id="project-en-hint">Для английской версии сайта. Если оставить поле пустым, там будет показан русский текст.</p>
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
          <div class="media-grid" id="new-media-list"></div>
        </div>` : ""}
        <div id="project-error" class="form__error hidden"></div>
        <div class="editor-card__actions">
          <button type="submit" class="btn btn--primary">${isNew ? "Создать" : "Сохранить"}</button>
          <button type="button" class="btn btn--ghost" id="preview-project">Превью</button>
          <button type="button" class="btn btn--ghost" id="cancel-edit">Отмена</button>
        </div>
      </form>
      ${isNew ? "" : `<div class="media-manager" id="media-manager"></div>`}
    </div>`;

  const pending = isNew ? setupPendingPhotos() : null;

  document.getElementById("preview-project").addEventListener("click", () => {
    const fields = readProjectFields(document.getElementById("project-form"));
    fields.title ||= "Без названия";
    if (isNew) {
      // New projects go to the top of their direction.
      openProjectPreview({ project: fields, number: 1, ...pending.previewUrls() });
      return;
    }
    // A project moved to another direction goes to the top of it.
    const keepsPlace = fields.category === project.category && projectPositions.has(project.id);
    openProjectPreview(savedProjectPreview(project, fields, keepsPlace ? projectPositions.get(project.id) : 1));
  });

  document.getElementById("cancel-edit").addEventListener("click", () => {
    pending?.release();
    editor.classList.add("hidden");
  });

  document.getElementById("project-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = document.getElementById("project-error");
    errEl.classList.add("hidden");
    const fields = readProjectFields(e.target);
    try {
      if (isNew) {
        const fd = new FormData();
        for (const [name, value] of Object.entries(fields)) fd.set(name, value);
        for (const file of pending.files()) fd.append("media", file);
        if (pending.files().length) fd.set("cover_index", String(pending.coverIndex()));
        await api.projects.create(fd);
        pending.release();
      } else {
        // The API replaces every field, so pass the current cover to keep it.
        await api.projects.update(project.id, { ...fields, cover_media_id: project.cover_media_id ?? null });
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

// Texts and category from the project form, as the API takes them.
function readProjectFields(form) {
  const value = (name) => form.querySelector(`[name=${name}]`).value.trim();
  return {
    title: value("title"),
    title_en: value("title_en"),
    description: value("description"),
    description_en: value("description_en"),
    category: value("category"),
  };
}

// Tooltip over each photo (shown on hover, see [data-tooltip] in main.css).
const PHOTO_HINT =
  "Порядок как в галерее проекта на сайте: перетащите фото на нужное место. " +
  "Обложка показывается в карточке проекта и первой в его окне";

// One photo tile with cover and remove controls. `id` identifies it for makeSortable,
// `index` is its 0-based position; photos can be dragged when there are several.
function photoTile({ src, alt, id, index, total, isCover }) {
  const sortable = total > 1;
  return `
    <div class="media-thumb ${isCover ? "media-thumb--cover" : ""} ${sortable ? "media-thumb--sortable" : ""}"
      data-id="${id}" data-index="${index}">
      <div class="media-thumb__frame" ${sortable ? `data-tooltip="${PHOTO_HINT}"` : ""}>
        <img src="${src}" alt="${esc(alt)}" class="media-thumb__img" draggable="false">
        <span class="media-thumb__pos">${index + 1}</span>
        ${sortable ? `<button type="button" class="drag-handle drag-handle--photo"
          aria-label="Переместить фото ${index + 1}: клавиши со стрелками">⋮⋮</button>` : ""}
        ${isCover ? `<span class="media-thumb__badge">Обложка</span>` : ""}
      </div>
      <div class="media-thumb__actions">
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
  let items = []; // { key, file, url }
  let nextKey = 1;
  let coverKey = null; // the cover follows its photo when the order changes
  const cover = () => Math.max(0, items.findIndex((item) => item.key === coverKey));

  const render = () => {
    list.innerHTML = items.map((item, i) => photoTile({
      src: item.url, alt: item.file.name, id: item.key, index: i, total: items.length,
      isCover: item.key === coverKey,
    })).join("");
  };

  // The list element stays while tiles are redrawn, so dragging is set up once.
  makeSortable(list, {
    item: ".media-thumb",
    pos: ".media-thumb__pos",
    onChange: (keys) => {
      items = keys.map((key) => items.find((item) => String(item.key) === key));
      render();
    },
  });

  input.addEventListener("change", () => {
    for (const file of input.files) items.push({ key: nextKey++, file, url: URL.createObjectURL(file) });
    input.value = ""; // allows picking more files, including the same ones again
    if (coverKey === null && items.length) coverKey = items[0].key;
    render();
  });

  list.addEventListener("click", (e) => {
    const button = e.target.closest("button:not(.drag-handle)");
    if (!button) return;
    const i = Number(button.closest(".media-thumb").dataset.index);
    if ("cover" in button.dataset) {
      coverKey = items[i].key;
    } else if ("remove" in button.dataset) {
      URL.revokeObjectURL(items[i].url);
      const [removed] = items.splice(i, 1);
      if (removed.key === coverKey) coverKey = items[0]?.key ?? null;
    }
    render();
  });

  return {
    previewUrls: () => {
      const coverUrl = items[cover()]?.url ?? null;
      return {
        coverUrl,
        galleryUrls: items.filter((_, i) => i !== cover()).map((item) => item.url),
      };
    },
    files: () => items.map((item) => item.file),
    coverIndex: cover,
    release: () => items.forEach((item) => URL.revokeObjectURL(item.url)),
  };
}

// Photos of an existing project: every change is saved right away.
function mountMediaManager(project) {
  const box = document.getElementById("media-manager");
  const media = project.media ?? [];
  box.innerHTML = `
    <h4>Фотографии</h4>
    <div class="media-grid" id="media-grid">
      ${media.length ? media.map((m, i) => photoTile({
        src: api.projects.mediaUrl(project.id, m.id), alt: m.filename, id: m.id,
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

  makeSortable(document.getElementById("media-grid"), {
    item: ".media-thumb",
    pos: ".media-thumb__pos",
    onChange: async (ids) => {
      try {
        await api.projects.reorderMedia(project.id, ids);
      } catch (err) {
        alert("Ошибка: " + err.message);
      }
      await refresh();
    },
  });

  document.getElementById("media-grid").addEventListener("click", async (e) => {
    const button = e.target.closest("button:not(.drag-handle)");
    if (!button) return;
    const i = Number(button.closest(".media-thumb").dataset.index);
    const target = media[i];
    try {
      if ("cover" in button.dataset) {
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
