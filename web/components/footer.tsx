"use client";

import Link from "next/link"
import { useAutoTranslation } from "@/contexts/auto-translation-context"
import { contactCopy } from "@/components/contact-content"
import { Pill, Facebook, Twitter, Instagram, Linkedin } from "lucide-react"

export function Footer() {
  const { language } = useAutoTranslation();
  const t = contactCopy[language];
  const footerAddress = {
    en: ["2, Brahim Roudani St., Apt. 206", "10050 Kebibat, Rabat, Morocco"],
    fr: ["2, rue Brahim Roudani, appt 206", "10050 Kebibat, Rabat, Maroc"],
    ar: ["2، شارع إبراهيم الروداني، الشقة 206", "10050، القبيبات، الرباط، المغرب"],
  }[language];
  const contactLink = "inline-block max-w-full rounded text-slate-300 underline-offset-4 transition-colors hover:text-teal-300 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-teal-400";
  return (
    <footer className="bg-slate-950 text-slate-300">
      <div className="container mx-auto px-4 py-12">
        <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-4">
          <div>
            <Link href="/" className="flex items-center gap-2 mb-4">
              <div className="bg-teal-600 text-white p-1 rounded">
                <Pill className="h-5 w-5" />
              </div>
              <span className="font-bold text-xl text-white">Qayen</span>
            </Link>
            <p className="text-slate-400 mb-4">
              Find medications and pharmacies near you with real-time availability and pricing.
            </p>
            <div className="flex gap-4">
              <Link href="#" className="text-slate-400 hover:text-white transition-colors">
                <Facebook className="h-5 w-5" />
              </Link>
              <Link href="#" className="text-slate-400 hover:text-white transition-colors">
                <Twitter className="h-5 w-5" />
              </Link>
              <Link href="#" className="text-slate-400 hover:text-white transition-colors">
                <Instagram className="h-5 w-5" />
              </Link>
              <Link href="#" className="text-slate-400 hover:text-white transition-colors">
                <Linkedin className="h-5 w-5" />
              </Link>
            </div>
          </div>
          <div>
            <h3 className="font-semibold text-white mb-4">Quick Links</h3>
            <ul className="space-y-2">
              <li>
                <Link href="/medications" className="text-slate-400 hover:text-white transition-colors">
                  Medications
                </Link>
              </li>
              <li>
                <Link href="/pharmacies" className="text-slate-400 hover:text-white transition-colors">
                  Pharmacies
                </Link>
              </li>
              <li>
                <Link href="/categories" className="text-slate-400 hover:text-white transition-colors">
                  Categories
                </Link>
              </li>
              <li>
                <Link href="/blog" className="text-slate-400 hover:text-white transition-colors">
                  Health Blog
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h3 className="font-semibold text-white mb-4">Information</h3>
            <ul className="space-y-2">
              <li>
                <Link href="/about" className="text-slate-400 hover:text-white transition-colors">
                  About Us
                </Link>
              </li>
              <li>
                <Link href="/contact" className="text-slate-400 hover:text-white transition-colors">
                  Contact Us
                </Link>
              </li>
              <li>
                <Link href="/privacy" className="text-slate-400 hover:text-white transition-colors">
                  Privacy Policy
                </Link>
              </li>
              <li>
                <Link href="/terms" className="text-slate-400 hover:text-white transition-colors">
                  Terms of Service
                </Link>
              </li>
            </ul>
          </div>
          <div translate="no" lang={language} dir={language === "ar" ? "rtl" : "ltr"}>
            <h3 className="font-semibold text-white mb-4">{t.title}</h3>
            <address className="not-italic text-sm text-slate-400 space-y-2">
              <div className="leading-6">
                {footerAddress.map(line => <p key={line}>{line}</p>)}
              </div>
              <p>
                <a href="tel:+212688735709" className={contactLink}><bdi dir="ltr">+(212)0688735709</bdi></a>
              </p>
              <p>
                <a href="mailto:QAYENsolutions@outlook.com" className={`${contactLink} break-all`}><bdi dir="ltr">QAYENsolutions@outlook.com</bdi></a>
              </p>
            </address>
          </div>
        </div>
        <div className="border-t border-slate-800 mt-12 pt-6 text-center text-slate-400">
          <p>&copy; {new Date().getFullYear()} Qayen. All rights reserved.</p>
        </div>
      </div>
    </footer>
  )
}
