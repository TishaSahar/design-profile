const YEAR = new Date().getFullYear();

export const LOCALES = {
  ru: {
    site_title: "Ольга — дизайн интерьеров",
    logo: "Ольга",
    nav_portfolio: "Портфолио",
    nav_about: "Обо мне",
    nav_contacts: "Контакты",
    footer_copy: `© ${YEAR} Ольга. Все права защищены.`,

    hero_subtitle: "Дизайн интерьеров",
    hero_cta: "Заказать проект",

    portfolio_descr: "Проекты, над которыми я работала",
    portfolio_empty: "Проекты появятся здесь совсем скоро.",
    project_open_tooltip: "Перейти к проекту →",

    project_back: "← Назад",
    project_no_photos: "Фото будут добавлены.",
    project_request_btn: "Оставить заявку",

    contacts_title: "Контакты",
    contacts_descr: "Напишите удобным для вас способом — обсудим ваш проект",
    contacts_email: "Электронная почта",
    contacts_instagram: "Instagram",
    contacts_telegram: "Telegram",
    contacts_request_btn: "Оставить заявку на проект",

    about_title: "Обо мне",
    about_request_btn: "Оставить заявку на проект",

    modal_title: "Заявка на проект",
    modal_first_name: "Имя *",
    modal_last_name: "Фамилия *",
    modal_email: "E-mail *",
    modal_email_short: "E-mail",
    modal_phone: "Телефон",
    modal_telegram: "Telegram",
    modal_preferred: "Предпочтительный способ связи",
    modal_description: "Описание проекта *",
    modal_files: "Приложите до 10 фото или 1 PDF с чертежами",
    modal_consent_label: "Я даю согласие на",
    modal_consent_link: "обработку персональных данных",
    modal_consent_body:
      "Нажимая кнопку «Отправить», вы соглашаетесь с тем, что предоставленные вами персональные данные (имя, контактная информация) будут обработаны исключительно в целях ответа на ваш запрос. Данные не передаются третьим лицам и хранятся в соответствии с законодательством о защите персональных данных.",
    modal_submit: "Отправить заявку",
    modal_success_title: "Заявка отправлена!",
    modal_success_body: "Мы свяжемся с вами в ближайшее время.",
    modal_close: "Закрыть",

    err_email: "Укажите корректный e-mail.",
    err_consent: "Необходимо согласие на обработку персональных данных.",
    err_load: "Ошибка загрузки:",
  },

  en: {
    site_title: "Olga — Interior Design",
    logo: "Olga",
    nav_portfolio: "Portfolio",
    nav_about: "About",
    nav_contacts: "Contacts",
    footer_copy: `© ${YEAR} Olga. All rights reserved.`,

    hero_subtitle: "Interior Design",
    hero_cta: "Request a project",

    portfolio_descr: "Projects I have worked on",
    portfolio_empty: "Projects will appear here soon.",
    project_open_tooltip: "View project →",

    project_back: "← Back",
    project_no_photos: "Photos will be added.",
    project_request_btn: "Make a request",

    contacts_title: "Contacts",
    contacts_descr: "Get in touch whichever way suits you — let's talk about your project",
    contacts_email: "Email",
    contacts_instagram: "Instagram",
    contacts_telegram: "Telegram",
    contacts_request_btn: "Submit a project request",

    about_title: "About the designer",
    about_request_btn: "Submit a project request",

    modal_title: "Project Request",
    modal_first_name: "First name *",
    modal_last_name: "Last name *",
    modal_email: "E-mail *",
    modal_email_short: "E-mail",
    modal_phone: "Phone",
    modal_telegram: "Telegram",
    modal_preferred: "Preferred contact method",
    modal_description: "Project description *",
    modal_files: "Attach up to 10 photos or 1 PDF",
    modal_consent_label: "I consent to the",
    modal_consent_link: "processing of personal data",
    modal_consent_body:
      `By clicking "Submit", you agree that the personal data you provide (name, contact information) will be processed solely for the purpose of responding to your inquiry. Data is not shared with third parties and is stored in accordance with applicable data protection law.`,
    modal_submit: "Submit request",
    modal_success_title: "Request submitted!",
    modal_success_body: "We will contact you shortly.",
    modal_close: "Close",

    err_email: "Please enter a valid e-mail address.",
    err_consent: "You must consent to the processing of personal data.",
    err_load: "Loading error:",
  },
};

export function getLang() {
  return localStorage.getItem("lang") || "ru";
}

export function setLang(lang) {
  if (!LOCALES[lang]) return;
  localStorage.setItem("lang", lang);
  document.documentElement.lang = lang;
}

export function t(key) {
  const lang = getLang();
  return LOCALES[lang]?.[key] ?? LOCALES.ru[key] ?? key;
}

export function applyTranslations() {
  document.querySelectorAll("[data-i18n]").forEach((el) => {
    el.textContent = t(el.dataset.i18n);
  });
  const lang = getLang();
  document.querySelectorAll(".lang-option").forEach((btn) => {
    btn.classList.toggle("lang-option--active", btn.dataset.lang === lang);
  });
  const label = document.querySelector(".lang-current-label");
  if (label) label.textContent = lang.toUpperCase();
  document.title = t("site_title");
  document.documentElement.lang = lang;
}
