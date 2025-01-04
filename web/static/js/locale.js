
class Locale {
  constructor(url) {
    this.url = url;
    this.locale = null;
    this.translations = null;
    this.attribute = 'data-i18n';
  }

  async setLocale(newLocale) {
    if (newLocale == this.locale) return;
    const newTranslations = await this.fetchTranslationsFor(newLocale);
    this.locale = newLocale;
    this.translations = newTranslations;
    this.translatePage();
  }

  get(key) {
    return this.translations[key];
  }

  async fetchTranslationsFor(locale) {
    const response = await fetch(`${this.url}/${locale}.json`);
    return await response.json();
  }
  
  translatePage() {
    this.translateElementTree(document)
  }
  
  translateElementTree(root) {
    root.querySelectorAll(`[${this.attribute}]`)
      .forEach(this.translateElement.bind(this));

    const fields = ['placeholder', 'value'];
    fields.forEach(this.translatePageField.bind(this));
  }
  
  translateElement(element) {
    const key = element.getAttribute(this.attribute);
    const translation = this.translations[key];
    element.innerHTML = translation;
  } 
  
  translatePageField(field) {
    document.querySelectorAll(`[${this.attribute}-${field}]`)
      .forEach(elem => this.translateElementField(elem, field));
  }
  
  translateElementField(element, field) {
    const attr = `${this.attribute}-${field}`
    const key = element.getAttribute(attr);
    const translation = this.translations[key];
    element.setAttribute(field, translation)
  }
}

class LocaleSwitcher {
  constructor(locale, buttons) {
    this.locale = locale;
    this.buttons = buttons;
    
    for (const btn of buttons) {
      btn.addEventListener('click', () => {
        const locale = btn.getAttribute('locale');
        this.locale.setLocale(locale);
        buttons.forEach(b => b.setAttribute('locale-on', 'false'));
        btn.setAttribute('locale-on', 'true');
      });
    }
  }

  setLocale(locale) {
    for (const btn of this.buttons) {
      const btnLocale = btn.getAttribute('locale');
      if (btnLocale === locale) {
        btn.click();
        return;
      }
    }
    throw new Error('invalid locale ' + locale);
  }
}

let locale;
let localeSwitcher;

document.addEventListener('htmx:afterSwap', (e) => {
  const tryToTranslate = () => {
    if (locale.translations !== null) {
      locale.translateElementTree(e.target);
    } else {
      setTimeout(() => {
        tryToTranslate()
      }, 500);
    }
  }
  tryToTranslate();
});

document.addEventListener("DOMContentLoaded", () => {
  locale = new Locale("/static/lang");
  const buttons = document.querySelectorAll(".locale-switcher .locale-btn");
  localeSwitcher = new LocaleSwitcher(locale, buttons);
  for (const bl of browserLocales()) {
    try {
      localeSwitcher.setLocale(bl);
      return;
    } catch (err) {
      console.log(err);
      continue;
    }
  }
  localeSwitcher.setLocale("en");
});

function browserLocales() {
  return navigator.languages.map((locale) => locale.split('-')[0]);
}