const YEAR = new Date().getFullYear();

export const LOCALES = {
  ru: {
    site_title: "OLGA SAKHA — архитектор интерьеров | SAKHA DESIGN",
    lang_label: "Язык / Language",

    nav_projects: "Проекты",
    nav_services: "Услуги",
    nav_about: "Обо мне",
    nav_contacts: "Контакты",
    nav_cta: "Заказать проект",
    nav_menu: "Меню",

    hero_title: "Дизайн интерьеров для жизни и бизнеса",
    hero_subtitle:
      "Продуманные пространства для жизни и бизнеса. От планировки до дизайн-концепции с реальными материалами и мебелью без необходимости заказывать полный дизайн-проект.",
    hero_btn_projects: "Смотреть проекты",
    hero_btn_services: "Выбрать услугу",
    hero_photo_alt: "Интерьер, спроектированный SAKHA DESIGN",

    dir_first: "1 направление",
    dir_second: "2 направление",
    dir_from: "от",
    dir_link: "Проекты →",
    dir_res_vertical: "Residential • Частные интерьеры",
    dir_res_title: "Residential — Квартиры и дома",
    dir_res_text:
      "Эргономичные планировки, кухни-гостиные, спальни, детские комнаты и приватные зоны отдыха, где каждый квадратный метр работает на ваш ежедневный комфорт и остаётся актуальным годами.",
    dir_res_photo_alt: "Жилой интерьер",
    dir_com_kicker: "Направления работы",
    dir_com_title: "HoReCa — рестораны и кафе, мини-отели и апартаменты, магазины и офисы",
    dir_com_text:
      "Кафе, кофейни, рестораны и небольшие коммерческие пространства: продуманная логистика персонала и гостей, оптимальная посадка, барные зоны и выразительная визуальная концепция.",
    dir_com_photo_alt: "Интерьер HoReCa",

    projects_title: "Избранные проекты",
    filter_all: "Все проекты",
    category_residential: "Residential",
    category_commercial: "HoReCa",
    projects_empty: "Проекты появятся здесь совсем скоро.",
    projects_empty_filter: "В этом направлении проектов пока нет.",
    project_open: "Открыть кейс",
    project_photos: (n) => `${n} фото`,
    case_no_photos: "Фотографии проекта будут добавлены.",
    case_cta_text: "Хотите похожее решение для вашего пространства?",
    case_cta_btn: "Обсудить похожий проект",
    modal_close: "Закрыть",

    services_kicker: "Независимые этапы проектирования",
    services_title: "Услуги и цены",
    services_text:
      "Уникальные дизайнерские проекты для вас и вашего бизнеса. Вы можете заказать каждую услугу отдельно без необходимости покупать полный дизайн-проект. Чертежи для ремонта выполняются после согласования планировочного решения.",
    calc_residential: "Residential / Жилой интерьер",
    calc_commercial: "HoReCa / Бизнес",
    calc_area: "Площадь:",
    unit_m2: "м²",
    price_residential: "Жилые интерьеры",
    price_commercial: "HoReCa",
    price_for: "Для",
    service_choose: "Выбрать услугу",
    service_sample: "Пример",
    services: [
      {
        id: "layout",
        tag: "01 • Независимый этап",
        title: "Планировочное решение",
        text: "Продуманная организация пространства с учётом ваших потребностей и образа жизни.",
        items: [
          "3 варианта планировки с расстановкой мебели",
          "Доработка выбранного варианта",
          "2 круга правок",
          "Итоговый план с размерами и расстановкой мебели",
        ],
      },
      {
        id: "drawings",
        tag: "02 • После планировки",
        title: "Чертежи для ремонта",
        note: "После согласования планировочного решения",
        text: "Техническая документация для реализации согласованного планировочного решения.",
        items: [
          "Обмерный план и общие данные проекта",
          "Планы демонтажа и монтажа перегородок",
          "Схема дверных и оконных проёмов",
          "Планы полов и потолков",
          "План размещения сантехнического оборудования",
          "План освещения и управления светом",
          "Сводный план электрических точек",
          "Развёртки стен всех помещений",
          "Ведомость отделочных материалов",
          "1 круг правок",
        ],
      },
      {
        id: "concept",
        tag: "03 • Визуал + артикулы",
        title: "Дизайн-концепция",
        text: "От идеи до готового визуального решения с реальными материалами и предметами интерьера.",
        items: [
          "Концептуальные коллажи по всем помещениям с подбором мебели, материалов, освещения и декора",
          "2 круга правок коллажей",
          "3D-модель интерьера в SketchUp",
          "2 круга правок 3D-модели",
          "Фотореалистичные AI-визуализации (2–3 ракурса на помещение)",
          "Спецификация мебели, освещения, сантехники, материалов и декора с артикулами, ценами и ссылками на покупку",
        ],
      },
    ],

    process_kicker: "Алгоритм взаимодействия",
    process_title: "Процесс работы",
    process_note: "100% дистанционный формат",
    process_stage: "Этап",
    process: [
      { title: "Заявка", text: "Вы оставляете заявку на сайте или в мессенджере, указываете тип объекта, площадь и интересующую услугу. Ответ в течение 24 часов." },
      { title: "Бриф", text: "Проводим онлайн-встречу, заполняем анкету-бриф, изучаем исходную планировку." },
      { title: "Договор и оплата", text: "Фиксируем точный состав выбранной услуги, сроки и прозрачные границы ответственности в договоре перед стартом проектирования." },
      { title: "Разработка и согласование", text: "Поэтапно презентуем решения онлайн и вносим корректировки в рамках согласованного количества кругов правок (1–2 круга на этап)." },
      { title: "Передача материалов", text: "Вы получаете структурированный PDF-альбом, рабочие чертежи, визуализации и таблицу спецификации с артикулами и ссылками на покупку." },
    ],

    about_kicker: "Обо мне • Olga Sakha",
    about_quote: "«Каждый квадратный метр должен работать на ваш комфорт и оставаться актуальным годами»",
    about_portrait_alt: "Ольга Саха — архитектор интерьеров",
    about_interior_alt: "Интерьер, спроектированный Ольгой Саха",
    about_facts: [
      { label: "Образование", value: "ННГАСУ (инженерно-строительное)" },
      { label: "Опыт", value: "С 2018 года" },
      { label: "География", value: "Дистанционно по всему миру" },
    ],

    faq_kicker: "Вопросы и ответы",
    faq_title: "Частые вопросы о работе",
    // The questions and answers themselves live in backend/config/about/faq.yaml.
    faq_text:
      "Прозрачные условия, фиксированный состав каждого этапа и честное обозначение границ ответственности перед началом проекта.",
    contacts_kicker: "Контакты",
    contacts_title: "Обсудить проект",
    contacts_text:
      "Для заказа проекта заполните короткую форму или напишите напрямую в Telegram, Instagram или на E-mail. Мы обсудим вашу задачу и подберём оптимальный формат работы.",
    contacts_email: "E-mail",
    contacts_instagram: "Instagram",
    contacts_telegram: "Telegram",

    form_title: "Заявка на расчёт и консультацию",
    form_first_name: "Имя *",
    form_last_name: "Фамилия *",
    form_email: "E-mail *",
    form_email_short: "E-mail",
    form_phone: "Телефон",
    form_telegram: "Telegram",
    form_preferred: "Предпочтительный способ связи",
    form_object: "Тип объекта",
    form_object_residential: "Квартира / дом",
    form_object_commercial: "Кафе / HoReCa",
    form_service: "Услуга",
    form_service_several: "Несколько услуг",
    form_area: "Площадь (м²)",
    form_comment: "Комментарий или город объекта",
    form_comment_placeholder: "Расскажите коротко о задаче или прикрепите ссылку на план…",
    form_files: "Приложите до 10 фото или 1 PDF с чертежами",
    form_consent_label: "Я даю согласие на",
    form_consent_link: "обработку персональных данных",
    form_consent_body:
      "Нажимая кнопку «Обсудить проект», вы соглашаетесь с тем, что предоставленные вами персональные данные (имя, контактная информация) будут обработаны исключительно в целях ответа на ваш запрос. Данные не передаются третьим лицам и хранятся в соответствии с законодательством о защите персональных данных.",
    form_submit: "Обсудить проект",
    form_sending: "Отправляем…",
    form_success_title: "Заявка отправлена.",
    form_success_body:
      "Спасибо за обращение — Ольга изучит вводные и свяжется с вами выбранным способом в течение рабочего дня.",

    err_name: "Укажите имя и фамилию.",
    err_email: "Укажите корректный e-mail.",
    err_consent: "Необходимо согласие на обработку персональных данных.",
    err_load: "Не удалось загрузить данные:",

    footer_role: "Olga Sakha — Interior Architect",
    footer_copy: `© 2018–${YEAR} SAKHA DESIGN. Все права защищены.`,

    sample_kicker: "Пример результата услуги",
    samples: {
      layout: {
        title: "Пример результата: Планировочное решение",
        text: "Вы получаете 3 принципиально разных варианта зонирования, проработку выбранного сценария (2 круга правок) и финальный монтажно-планировочный лист с габаритами мебели и проходов.",
        bedroom: "Спальня 16.4 м²",
        living: "Кухня-гостиная 34.2 м² (проход 1100 мм)",
        sheet: "ЛИСТ 04 — ИТОГОВОЕ ПЛАНИРОВОЧНОЕ РЕШЕНИЕ С РАЗМЕРАМИ",
      },
      drawings: {
        title: "Пример результата: Альбом чертежей для ремонта",
        text: "Полный комплект рабочей документации (10–18 листов PDF в масштабе) для строительной бригады: электрика, освещение, сантехника, развёртки стен и ведомость отделки.",
        sheets: [
          ["Лист 01–02:", "Обмерный план, демонтаж и монтаж перегородок"],
          ["Лист 03–05:", "Планы полов, потолков и привязка сантехники"],
          ["Лист 06–09:", "Сценарии света, электрика и развёртки всех стен"],
        ],
      },
      concept: {
        title: "Пример результата: Дизайн-концепция и спецификация",
        text: "Коллажи, 3D-модель в SketchUp, AI-визуализации (2–3 ракурса на помещение) и таблица комплектации с реальными артикулами, ценами и ссылками на покупку в вашей стране.",
        head: ["Наименование", "Артикул", "Ориентир цены", "Поставщик"],
        rows: [
          ["Обеденный стол из массива дуба 1800х900", "TB-OK-180", "€890", "Локальный шоурум / ссылка"],
          ["Краска стен тёплый алебастр (матовая)", "NCS S 1002-Y", "€14 / л", "Колеровка по вееру"],
          ["Подвесной светильник над столом", "LT-442-BR", "€310", "Официальный дилер"],
        ],
      },
    },
  },

  en: {
    site_title: "OLGA SAKHA — Interior Architect | SAKHA DESIGN",
    lang_label: "Язык / Language",

    nav_projects: "Projects",
    nav_services: "Services",
    nav_about: "About",
    nav_contacts: "Contacts",
    nav_cta: "Request a project",
    nav_menu: "Menu",

    hero_title: "Interior Design for Living & Business",
    hero_subtitle:
      "Thoughtful spaces for life and business. From space planning to a full design concept with real materials and furniture — without having to order a complete design project.",
    hero_btn_projects: "Explore projects",
    hero_btn_services: "Choose a service",
    hero_photo_alt: "Interior designed by SAKHA DESIGN",

    dir_first: "Direction 1",
    dir_second: "Direction 2",
    dir_from: "from",
    dir_link: "Projects →",
    dir_res_vertical: "Residential • Private interiors",
    dir_res_title: "Residential — Apartments & Houses",
    dir_res_text:
      "Ergonomic layouts, kitchen-living rooms, bedrooms, children's rooms and private retreats where every square metre works for your everyday comfort and stays relevant for years.",
    dir_res_photo_alt: "Residential interior",
    dir_com_kicker: "Areas of work",
    dir_com_title: "HoReCa — Restaurants & Cafés, Mini-Hotels & Apartments, Shops & Offices",
    dir_com_text:
      "Cafés, coffee shops, restaurants and small commercial spaces: well-planned staff and guest flows, efficient seating, bar areas and an expressive visual concept.",
    dir_com_photo_alt: "HoReCa interior",

    projects_title: "Selected projects",
    filter_all: "All projects",
    category_residential: "Residential",
    category_commercial: "HoReCa",
    projects_empty: "Projects will appear here soon.",
    projects_empty_filter: "No projects in this direction yet.",
    project_open: "Open case",
    project_photos: (n) => `${n} ${n === 1 ? "photo" : "photos"}`,
    case_no_photos: "Project photos will be added.",
    case_cta_text: "Would you like a similar solution for your space?",
    case_cta_btn: "Discuss a similar project",
    modal_close: "Close",

    services_kicker: "Independent design stages",
    services_title: "Services & Prices",
    services_text:
      "Bespoke design projects for you and your business. Each service can be ordered on its own — there is no need to buy a complete design project. Renovation drawings are prepared once the space plan is approved.",
    calc_residential: "Residential",
    calc_commercial: "HoReCa / Business",
    calc_area: "Area:",
    unit_m2: "m²",
    price_residential: "Residential",
    price_commercial: "HoReCa",
    price_for: "For",
    service_choose: "Choose service",
    service_sample: "Sample",
    services: [
      {
        id: "layout",
        tag: "01 • Independent stage",
        title: "Space planning",
        text: "A well-thought-out organisation of space that reflects your needs and lifestyle.",
        items: [
          "3 layout options with furniture arrangement",
          "Refinement of the chosen option",
          "2 rounds of revisions",
          "Final plan with dimensions and furniture arrangement",
        ],
      },
      {
        id: "drawings",
        tag: "02 • After the layout",
        title: "Renovation drawings",
        note: "Once the space plan is approved",
        text: "Technical documentation to build the approved space plan.",
        items: [
          "Measured plan and general project data",
          "Demolition and partition installation plans",
          "Door and window opening schedule",
          "Floor and ceiling plans",
          "Plumbing fixture layout",
          "Lighting and lighting control plan",
          "Consolidated electrical outlet plan",
          "Wall elevations for every room",
          "Finishing materials schedule",
          "1 round of revisions",
        ],
      },
      {
        id: "concept",
        tag: "03 • Visuals + product codes",
        title: "Design concept",
        text: "From an idea to a finished visual solution with real materials and furnishings.",
        items: [
          "Concept collages for every room with furniture, materials, lighting and decor",
          "2 rounds of revisions for the collages",
          "3D interior model in SketchUp",
          "2 rounds of revisions for the 3D model",
          "Photorealistic AI renderings (2–3 views per room)",
          "Specification of furniture, lighting, plumbing, materials and decor with product codes, prices and purchase links",
        ],
      },
    ],

    process_kicker: "How we work together",
    process_title: "Work process",
    process_note: "100% remote",
    process_stage: "Stage",
    process: [
      { title: "Request", text: "You leave a request on the website or in a messenger, stating the type of property, its area and the service you are interested in. You get a reply within 24 hours." },
      { title: "Brief", text: "We hold an online meeting, fill in a brief questionnaire and study the original layout." },
      { title: "Contract & payment", text: "Before design work starts, the contract sets out the exact scope of the chosen service, the timeline and clear boundaries of responsibility." },
      { title: "Design & approval", text: "Solutions are presented online stage by stage, and changes are made within the agreed number of revision rounds (1–2 per stage)." },
      { title: "Handover", text: "You receive a structured PDF album, working drawings, renderings and a specification table with product codes and purchase links." },
    ],

    about_kicker: "About me • Olga Sakha",
    about_quote: "“Every square metre should work for your comfort and stay relevant for years”",
    about_portrait_alt: "Olga Sakha — Interior Architect",
    about_interior_alt: "Interior designed by Olga Sakha",
    about_facts: [
      { label: "Education", value: "Degree in civil engineering (NNGASU)" },
      { label: "Experience", value: "Since 2018" },
      { label: "Location", value: "Remote, worldwide" },
    ],

    faq_kicker: "Questions & answers",
    faq_title: "Frequently asked questions",
    faq_text:
      "Transparent terms, a fixed scope for every stage and clearly stated boundaries of responsibility before the project starts.",
    contacts_kicker: "Contacts",
    contacts_title: "Discuss your project",
    contacts_text:
      "To order a project, fill in the short form or write directly via Telegram, Instagram or e-mail. We will discuss your task and find the best way to work together.",
    contacts_email: "E-mail",
    contacts_instagram: "Instagram",
    contacts_telegram: "Telegram",

    form_title: "Request a quote and consultation",
    form_first_name: "First name *",
    form_last_name: "Last name *",
    form_email: "E-mail *",
    form_email_short: "E-mail",
    form_phone: "Phone",
    form_telegram: "Telegram",
    form_preferred: "Preferred contact method",
    form_object: "Property type",
    form_object_residential: "Apartment / house",
    form_object_commercial: "Café / HoReCa",
    form_service: "Service",
    form_service_several: "Several services",
    form_area: "Area (m²)",
    form_comment: "Comment or city of the property",
    form_comment_placeholder: "Briefly describe your task or add a link to the floor plan…",
    form_files: "Attach up to 10 photos or 1 PDF with drawings",
    form_consent_label: "I consent to the",
    form_consent_link: "processing of personal data",
    form_consent_body:
      "By clicking “Discuss your project”, you agree that the personal data you provide (name, contact information) will be processed solely to respond to your request. The data is not shared with third parties and is stored in accordance with data protection law.",
    form_submit: "Discuss your project",
    form_sending: "Sending…",
    form_success_title: "Request sent.",
    form_success_body:
      "Thank you for getting in touch — Olga will review the details and contact you via your preferred method within one working day.",

    err_name: "Please enter your first and last name.",
    err_email: "Please enter a valid e-mail address.",
    err_consent: "You must consent to the processing of personal data.",
    err_load: "Could not load data:",

    footer_role: "Olga Sakha — Interior Architect",
    footer_copy: `© 2018–${YEAR} SAKHA DESIGN. All rights reserved.`,

    sample_kicker: "Service deliverable sample",
    samples: {
      layout: {
        title: "Sample deliverable: Space planning",
        text: "You receive 3 fundamentally different zoning options, development of the chosen scenario (2 rounds of revisions) and a final layout sheet with furniture and passage dimensions.",
        bedroom: "Bedroom 16.4 m²",
        living: "Kitchen-living room 34.2 m² (passage 1100 mm)",
        sheet: "SHEET 04 — FINAL SPACE PLAN WITH DIMENSIONS",
      },
      drawings: {
        title: "Sample deliverable: Renovation drawings album",
        text: "A complete set of working documentation (10–18 scaled PDF sheets) for the construction team: electrics, lighting, plumbing, wall elevations and a finishes schedule.",
        sheets: [
          ["Sheets 01–02:", "Measured plan, demolition and partition installation"],
          ["Sheets 03–05:", "Floor and ceiling plans, plumbing positions"],
          ["Sheets 06–09:", "Lighting scenarios, electrics and all wall elevations"],
        ],
      },
      concept: {
        title: "Sample deliverable: Design concept and specification",
        text: "Collages, a SketchUp 3D model, AI renderings (2–3 views per room) and a furnishing schedule with real product codes, prices and purchase links in your country.",
        head: ["Item", "Product code", "Approx. price", "Supplier"],
        rows: [
          ["Solid oak dining table 1800×900", "TB-OK-180", "€890", "Local showroom / link"],
          ["Warm alabaster wall paint (matt)", "NCS S 1002-Y", "€14 / l", "Tinted from the fan deck"],
          ["Pendant light above the table", "LT-442-BR", "€310", "Official dealer"],
        ],
      },
    },
  },
};

export function getLang() {
  try {
    return localStorage.getItem("lang") || "ru";
  } catch {
    return "ru";
  }
}

export function setLang(lang) {
  if (!LOCALES[lang]) return;
  try {
    localStorage.setItem("lang", lang);
  } catch {
    // Private mode: keep the choice for this page only.
  }
  document.documentElement.lang = lang;
}

// t returns a string (or a list/object for structured content) in the current language.
export function t(key, ...args) {
  const lang = getLang();
  const value = LOCALES[lang]?.[key] ?? LOCALES.ru[key] ?? key;
  return typeof value === "function" ? value(...args) : value;
}

// tRu returns the Russian text regardless of the interface language (used for
// request descriptions, which the Russian-speaking admin reads).
export function tRu(key) {
  return LOCALES.ru[key] ?? key;
}

export function applyTranslations() {
  document.querySelectorAll("[data-i18n]").forEach((el) => {
    el.textContent = t(el.dataset.i18n);
  });
  for (const attr of ["placeholder", "alt", "aria-label"]) {
    document.querySelectorAll(`[data-i18n-${attr}]`).forEach((el) => {
      el.setAttribute(attr, t(el.getAttribute(`data-i18n-${attr}`)));
    });
  }
  const lang = getLang();
  document.querySelectorAll(".lang-option").forEach((btn) => {
    btn.classList.toggle("lang-option--active", btn.dataset.lang === lang);
  });
  const label = document.querySelector(".lang-current-label");
  if (label) label.textContent = lang.toUpperCase();
  document.title = t("site_title");
  document.documentElement.lang = lang;
}
