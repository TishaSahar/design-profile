// Admin preview of a project (preview.html inside an iframe of admin.html).
// The admin page sends the project — possibly unsaved — with postMessage:
//   { type: "project-preview", project: {title, title_en, description, description_en, category},
//     number, coverUrl, galleryUrls, lang }
// and the page draws it with the same markup as the public site.
import { setPageLang, applyTranslations } from "./locales.js";
import { initPhotos, initCardGalleries, projectCardHTML, projectCaseHTML } from "./project-view.js";

function render(data) {
  // The language the admin picked in the preview; the site's saved choice stays as it is.
  setPageLang(data.lang ?? "ru");
  applyTranslations();
  const grid = document.getElementById("preview-grid");
  const photoUrls = data.coverUrl ? [data.coverUrl, ...data.galleryUrls] : [];
  grid.innerHTML = projectCardHTML(data.project, { number: data.number, photoUrls });
  const { meta, body } = projectCaseHTML(data.project, data);
  document.getElementById("case-meta").textContent = meta;
  document.getElementById("case-body").innerHTML = body;
  initPhotos(document.body);
  initCardGalleries(grid);
}

// Links lead nowhere in the preview; the card's photo arrows still work.
document.addEventListener("click", (e) => {
  if (e.target.closest("a")) e.preventDefault();
});

window.addEventListener("message", (e) => {
  if (e.origin !== location.origin || e.data?.type !== "project-preview") return;
  render(e.data);
});

// Ask the admin page for the project once this page can draw it.
window.parent.postMessage({ type: "project-preview-ready" }, location.origin);
