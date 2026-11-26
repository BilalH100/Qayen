"use client";

import { ArrowUpRight, Mail, MapPin, Phone } from "lucide-react";
import { useAutoTranslation } from "@/contexts/auto-translation-context";

const copy = {
  en: {
    title: "Contact Qayen",
    description: "Have a question? Get in touch with our team by phone or email.",
    phone: "Call us", morocco: "Morocco", international: "International (USA)",
    email: "Email us", inquiries: "General inquiries", healthtech: "Qayen HealthTech",
    address: "Our address", maps: "Open in Maps",
    addressLines: ["2, Brahim Roudani Street, Apt. 206", "10050 Kebibat, Rabat", "Kingdom of Morocco"],
  },
  fr: {
    title: "Contactez Qayen",
    description: "Une question ? Contactez notre équipe par téléphone ou par e-mail.",
    phone: "Appelez-nous", morocco: "Maroc", international: "International (États-Unis)",
    email: "Écrivez-nous", inquiries: "Demandes de renseignements", healthtech: "Qayen HealthTech",
    address: "Notre adresse", maps: "Ouvrir la carte",
    addressLines: ["2, rue Brahim Roudani, appt 206", "CP 10050, quartier Kebibat, Rabat", "Royaume du Maroc"],
  },
  ar: {
    title: "تواصل مع قايِن",
    description: "هل لديك سؤال؟ تواصل مع فريقنا عبر الهاتف أو البريد الإلكتروني.",
    phone: "اتصل بنا", morocco: "المغرب", international: "دولي (الولايات المتحدة)",
    email: "راسلنا", inquiries: "الاستفسارات", healthtech: "قايِن للتكنولوجيا الصحية",
    address: "عنواننا", maps: "فتح الخريطة",
    addressLines: ["2، شارع إبراهيم الروداني، الشقة 206", "الرمز البريدي 10050، حي القبيبات، الرباط", "المملكة المغربية"],
  },
};

const card = "rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-700 dark:bg-slate-900";
const contactLink = "inline-block max-w-full rounded text-base font-medium text-teal-700 underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-teal-500 focus-visible:ring-offset-2 dark:text-teal-300";
const iconBox = "mb-5 flex h-12 w-12 items-center justify-center rounded-xl bg-teal-50 text-teal-700 dark:bg-teal-950 dark:text-teal-300";
const mapQuery = "2, rue Brahim Roudani appt 206 CP 10050 Quartier Kebibat Rabat, Maroc";

export function ContactContent() {
  const { language } = useAutoTranslation();
  const t = copy[language];

  return (
    <section translate="no" lang={language} dir={language === "ar" ? "rtl" : "ltr"} className="min-h-[60vh] bg-slate-50 py-12 dark:bg-slate-950 sm:py-16">
      <div className="container mx-auto max-w-6xl px-4">
        <div className="mb-8 rounded-2xl border border-teal-100 bg-gradient-to-br from-teal-50 to-white p-6 dark:border-teal-900 dark:from-teal-950 dark:to-slate-900 sm:p-10">
          <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white sm:text-4xl">{t.title}</h1>
          <p className="mt-3 max-w-2xl text-base leading-relaxed text-slate-600 dark:text-slate-300 sm:text-lg">{t.description}</p>
        </div>

        <div className="grid gap-5 lg:grid-cols-3">
          <article className={card}>
            <div className={iconBox}><Phone className="h-6 w-6" aria-hidden="true" /></div>
            <h2 className="mb-5 text-xl font-semibold text-slate-900 dark:text-white">{t.phone}</h2>
            <div className="space-y-5">
              <div>
                <p className="mb-1 text-sm text-slate-500 dark:text-slate-400">{t.morocco}</p>
                <a href="tel:+212688735709" className={contactLink}><bdi dir="ltr">+(212)0688735709</bdi></a>
              </div>
              <div>
                <p className="mb-1 text-sm text-slate-500 dark:text-slate-400">{t.international}</p>
                <a href="tel:+14437649250" className={contactLink}><bdi dir="ltr">+(1)443-764-9250</bdi></a>
              </div>
            </div>
          </article>

          <article className={card}>
            <div className={iconBox}><Mail className="h-6 w-6" aria-hidden="true" /></div>
            <h2 className="mb-5 text-xl font-semibold text-slate-900 dark:text-white">{t.email}</h2>
            <div className="space-y-5">
              <div>
                <p className="mb-1 text-sm text-slate-500 dark:text-slate-400">{t.inquiries}</p>
                <a href="mailto:QAYENsolutions@outlook.com" className={`${contactLink} break-all`}><bdi dir="ltr">QAYENsolutions@outlook.com</bdi></a>
              </div>
              <div>
                <p className="mb-1 text-sm text-slate-500 dark:text-slate-400">{t.healthtech}</p>
                <a href="mailto:QAYENHealthTech@outlook.com" className={`${contactLink} break-all`}><bdi dir="ltr">QAYENHealthTech@outlook.com</bdi></a>
              </div>
            </div>
          </article>

          <article className={card}>
            <div className={iconBox}><MapPin className="h-6 w-6" aria-hidden="true" /></div>
            <h2 className="mb-5 text-xl font-semibold text-slate-900 dark:text-white">{t.address}</h2>
            <address className="text-base not-italic leading-7 text-slate-600 dark:text-slate-300">
              {t.addressLines.map(line => <p key={line}>{line}</p>)}
            </address>
            <a href={`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(mapQuery)}`} target="_blank" rel="noopener noreferrer" className={`${contactLink} mt-5 inline-flex items-center gap-1.5 text-sm`}>
              {t.maps}<ArrowUpRight className="h-4 w-4 rtl:-rotate-90" aria-hidden="true" />
            </a>
          </article>
        </div>
      </div>
    </section>
  );
}

export { copy as contactCopy };
