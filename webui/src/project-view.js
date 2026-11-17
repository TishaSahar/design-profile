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
// Photos whose loading is postponed (data-src, no src yet) are checked once they load.
export function initPhotos(root) {
  root.querySelectorAll(".photo img[src]").forEach(watchPhoto);
}

function watchPhoto(img) {
  const markEmpty = () => img.closest(".photo")?.classList.add("photo--empty");
  if (img.complete && img.naturalWidth === 0) markEmpty();
  else img.addEventListener("error", markEmpty, { once: true });
}

// Photo URLs of a saved project in gallery order: the cover first, then the rest.
export function projectPhotoUrls(p, mediaUrl) {
  const media = p.media ?? [];
  const coverId = p.cover_media_id ?? media[0]?.id;
  const ordered = coverId ? [coverId, ...media.map((m) => m.id).filter((id) => id !== coverId)] : [];
  return ordered.map((id) => mediaUrl(p.id, id));
}

// Card in the projects grid. `number` is the position shown on the card (1-based);
// `photoUrls` are the project's photos, cover first — with several, the card pages
// through them (see initCardGalleries).
export function projectCardHTML(p, { number, photoUrls }) {
  const href = `#/project/${p.id ?? "preview"}`;
  const many = photoUrls.length > 1;
  return `
    <article class="project-card">
      <div class="card-gallery" ${many ? "data-gallery" : ""}>
        <a class="card-gallery__track" href="${href}" tabindex="-1" aria-hidden="true">
          ${photoUrls.length
            // Only the cover loads with the page; the other photos load when the card is near.
            ? photoUrls.map((url, i) => (i === 0
              ? photoHTML(url, p.title, "photo--zoom")
              : `<div class="photo photo--zoom"><img data-src="${url}" alt="${esc(`${p.title} — ${i + 1}`)}"></div>`)).join("")
            : photoHTML(null, p.title)}
        </a>
        ${many ? `
          <button type="button" class="card-gallery__nav card-gallery__nav--prev" data-step="-1" aria-label="${t("card_prev")}">‹</button>
          <button type="button" class="card-gallery__nav card-gallery__nav--next" data-step="1" aria-label="${t("card_next")}">›</button>
          <div class="card-gallery__dots" aria-hidden="true">
            ${photoUrls.map((_, i) => `<span class="card-gallery__dot ${i === 0 ? "is-active" : ""}"></span>`).join("")}
          </div>
          <span class="visually-hidden" aria-live="polite" data-gallery-status></span>` : ""}
      </div>
      <a class="project-card__link" href="${href}">
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

// ── Card galleries ───────────────────────────────────────────────────────────
// Visitors page through a card's photos with the arrows (mouse) or by swiping (touch).
// The photos also play by themselves: on a computer one second after the pointer
// rests on a card, on a phone one second after scrolling stops on a card.
const AUTOPLAY_DELAY = 1000; // ms before the first change
const AUTOPLAY_STEP = 2200; // ms per photo
const reducedMotion = matchMedia("(prefers-reduced-motion: reduce)");
const canHover = matchMedia("(hover: hover) and (pointer: fine)");

const galleryOf = (el) => el.querySelector(".card-gallery[data-gallery]");
const trackOf = (g) => g.querySelector(".card-gallery__track");
const slidesOf = (g) => [...trackOf(g).children];

// The photo shown — or being slid to: read from the scroll position only once it
// settles, so quick clicks during a slide count from where the slide is going.
function currentIndex(g) {
  return Number(g.dataset.index ?? 0);
}

function settledIndex(g) {
  const track = trackOf(g);
  return Math.round(track.scrollLeft / Math.max(1, track.clientWidth));
}

function loadSlide(g, i) {
  const img = slidesOf(g)[i]?.querySelector("img[data-src]");
  if (!img) return;
  img.addEventListener("error", () => img.closest(".photo")?.classList.add("photo--empty"), { once: true });
  img.src = img.dataset.src;
  img.removeAttribute("data-src");
}

function goTo(g, i) {
  const count = slidesOf(g).length;
  const target = (i + count) % count; // past the last photo comes the first again
  loadSlide(g, target);
  loadSlide(g, (target + 1) % count);
  const track = trackOf(g);
  // Neighbouring photos slide; longer jumps (back to the first photo) cut, so the
  // visitor does not watch every photo rush past backwards.
  const smooth = !reducedMotion.matches && Math.abs(target - currentIndex(g)) === 1;
  g.dataset.index = target;
  syncIndicators(g);
  track.scrollTo({ left: target * track.clientWidth, behavior: smooth ? "smooth" : "auto" });
}

function syncIndicators(g) {
  const i = currentIndex(g);
  const count = slidesOf(g).length;
  g.querySelectorAll(".card-gallery__dot").forEach((dot, n) => dot.classList.toggle("is-active", n === i));
  g.querySelector("[data-gallery-status]").textContent = t("card_photo", i + 1, count);
  loadSlide(g, (i + 1) % count);
  loadSlide(g, (i - 1 + count) % count);
}

// Only one card plays at a time.
let playing = null; // { gallery, timer }

function stopAutoplay({ rewind = false } = {}) {
  if (!playing) return;
  clearTimeout(playing.timer);
  if (rewind && playing.gallery.isConnected) goTo(playing.gallery, 0);
  playing = null;
}

function startAutoplay(g) {
  if (reducedMotion.matches || playing?.gallery === g) return;
  stopAutoplay();
  loadSlide(g, 1);
  const tick = () => {
    // Pause while a dialog covers the page or the card has been redrawn.
    if (!g.isConnected) { playing = null; return; }
    if (!document.body.classList.contains("is-locked")) goTo(g, currentIndex(g) + 1);
    playing.timer = setTimeout(tick, AUTOPLAY_STEP);
  };
  playing = { gallery: g, timer: setTimeout(tick, AUTOPLAY_DELAY) };
}

// Phones: after scrolling stops, play the card that is mostly on screen, nearest the centre.
let scrollTimer = null;
function onPageScroll() {
  if (canHover.matches) return;
  stopAutoplay();
  clearTimeout(scrollTimer);
  scrollTimer = setTimeout(() => {
    const middle = innerHeight / 2;
    let best = null;
    let bestDistance = Infinity;
    for (const g of document.querySelectorAll(".card-gallery[data-gallery]")) {
      const box = g.getBoundingClientRect();
      const visible = Math.min(box.bottom, innerHeight) - Math.max(box.top, 0);
      if (visible < box.height * 0.6) continue;
      const distance = Math.abs(box.top + box.height / 2 - middle);
      if (distance < bestDistance) { best = g; bestDistance = distance; }
    }
    if (best) startAutoplay(best);
  }, 150);
}

let pageListeners = false;

// Sets up the galleries of the cards inside `root` (call after drawing the cards).
export function initCardGalleries(root) {
  if (!pageListeners) {
    pageListeners = true;
    addEventListener("scroll", onPageScroll, { passive: true });
  }
  for (const g of root.querySelectorAll(".card-gallery[data-gallery]")) {
    const card = g.closest(".project-card");
    const track = trackOf(g);

    g.querySelectorAll("[data-step]").forEach((btn) => btn.addEventListener("click", (e) => {
      e.preventDefault();
      stopAutoplay();
      goTo(g, currentIndex(g) + Number(btn.dataset.step));
    }));

    // After a swipe, take the photo it stopped on.
    let settle = 0;
    track.addEventListener("scroll", () => {
      clearTimeout(settle);
      settle = setTimeout(() => {
        if (settledIndex(g) === currentIndex(g)) return;
        g.dataset.index = settledIndex(g);
        syncIndicators(g);
      }, 120);
    }, { passive: true });

    // A visitor swiping takes over from autoplay.
    track.addEventListener("touchstart", () => { if (playing?.gallery === g) stopAutoplay(); }, { passive: true });

    card.addEventListener("mouseenter", () => { if (canHover.matches) startAutoplay(g); });
    card.addEventListener("mouseleave", () => {
      if (!canHover.matches) return;
      if (playing?.gallery === g) stopAutoplay({ rewind: true });
      else if (currentIndex(g) !== 0) goTo(g, 0);
    });
  }
  // A phone visitor may land on a card without scrolling (e.g. after a filter change).
  onPageScroll();
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
