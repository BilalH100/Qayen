import Link from "next/link"
import { Pill, Facebook, Twitter, Instagram, Linkedin } from "lucide-react"

export function Footer() {
  return (
    <footer className="bg-slate-950 text-slate-300">
      <div className="container mx-auto px-4 py-12">
        <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-4">
          <div>
            <Link href="/" className="flex items-center gap-2 mb-4">
              <div className="bg-teal-600 text-white p-1 rounded">
                <Pill className="h-5 w-5" />
              </div>
              <span className="font-bold text-xl text-white">Kayena</span>
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
          <div>
            <h3 className="font-semibold text-white mb-4">Contact</h3>
            <address className="not-italic text-slate-400 space-y-2">
              <p>123 Health Street</p>
              <p>New York, NY 10001</p>
              <p>Email: info@Kayena.com</p>
              <p>Phone: (123) 456-7890</p>
            </address>
          </div>
        </div>
        <div className="border-t border-slate-800 mt-12 pt-6 text-center text-slate-400">
          <p>&copy; {new Date().getFullYear()} Kayena. All rights reserved.</p>
        </div>
      </div>
    </footer>
  )
}
