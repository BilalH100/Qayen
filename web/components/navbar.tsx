"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { Menu, Search, User, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PlatformSearchForm } from "@/components/platform-search-form";
import { ThemeToggle } from "@/components/theme-toggle";
import { LanguageSelector } from "@/components/language-selector";
import { Sheet, SheetClose, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { AuthModal } from "@/components/auth-modal";
import { UserMenu } from "@/components/user-menu";
import { useAuth } from "@/contexts/auth-context";
import { ROLES } from "@/utils/types";

const navigationLinkClassName =
  "inline-flex items-center rounded-lg border border-teal-100 bg-teal-50/70 px-3 py-2 text-sm font-medium text-slate-700 transition-colors hover:border-teal-300 hover:bg-teal-100 hover:text-teal-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-teal-500 focus-visible:ring-offset-2 dark:border-slate-700 dark:bg-slate-800/70 dark:text-slate-200 dark:hover:border-teal-700 dark:hover:bg-teal-950 dark:hover:text-teal-300";

const catalogSections = [
  { href: "/parapharmacies", label: "Parapharmacies" },
  { href: "/supplies-accessories", label: "Supplies & Accessories" },
  { href: "/medical-materials", label: "Medical Materials" },
];

export function Navbar() {
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [authMode, setAuthMode] = useState<"login" | "register">("login");

  const { user, isAuthenticated, isLoading } = useAuth();

  const handleAuthClick = (mode: "login" | "register") => {
    setAuthMode(mode);
    setIsAuthModalOpen(true);
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex min-h-16 items-center">
        <Sheet>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="lg:hidden">
              <Menu className="h-5 w-5" />
              <span className="sr-only">Toggle menu</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-[300px] overflow-y-auto sm:w-[400px]">
            <nav className="mt-8 flex flex-col gap-2">
              <Link href="/" className={navigationLinkClassName}>
                Home
              </Link>
              <Link href="/medications" className={navigationLinkClassName}>
                Medications
              </Link>
              <Link href="/pharmacies" className={navigationLinkClassName}>
                Pharmacies
              </Link>
              {user?.role === ROLES.PHARMACIST && (
                <Link href="/pharmacy" className={navigationLinkClassName}>
                  Pharmacy Dashboard
                </Link>
              )}
              {catalogSections.map(({ href, label }) => (
                <SheetClose asChild key={href}>
                  <Link href={href} className={navigationLinkClassName}>
                    {label}
                  </Link>
                </SheetClose>
              ))}
              <Link href="/about" className={navigationLinkClassName}>
                About
              </Link>
              <Link href="/contact" className={navigationLinkClassName}>
                Contact
              </Link>

              <div className="mt-8 pt-4 border-t">
                {isAuthenticated ? (
                  <div className="space-y-2">
                    <p className="text-sm text-gray-600">Signed in as</p>
                    <p className="font-medium">{user?.name}</p>
                    <p className="text-sm text-gray-500">{user?.email}</p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    <Button
                      className="w-full bg-teal-600 hover:bg-teal-700"
                      onClick={() => handleAuthClick("login")}
                    >
                      Sign In
                    </Button>
                    <Button
                      variant="outline"
                      className="w-full"
                      onClick={() => handleAuthClick("register")}
                    >
                      Create Account
                    </Button>
                  </div>
                )}
              </div>
            </nav>
          </SheetContent>
        </Sheet>

        <Link href="/" className="flex shrink-0 items-center gap-2 me-4">
          {/* <div className="bg-teal-600 text-white p-1 rounded">
            <Image src={logo} alt={""} height={50} width={50}/>
          </div> */}
        <Image
          src="/Logo_HomePage.png"
          alt="Qayen"
          width={100}
          height={60}
          priority
          style={{ width: "70px", height: "60px" }}
          className="object-contain"
/>
        </Link>

        <nav className="hidden min-w-0 flex-1 flex-wrap items-center gap-2 py-2 pe-3 lg:flex [&>a]:whitespace-nowrap">
          <Link
            href="/medications"
            className={navigationLinkClassName}
          >
            Medications
          </Link>

          <Link
            href="/pharmacies"
            className={navigationLinkClassName}
          >
            Pharmacies
          </Link>

          {user?.role === ROLES.PHARMACIST && (
            <Link
              href="/pharmacy"
              className={navigationLinkClassName}
            >
              Pharmacy Dashboard
            </Link>
          )}
          {user?.role === ROLES.ADMIN && (
            <Link
              href="/admin"
              className={navigationLinkClassName}
            >
              Admin Dashboard
            </Link>
          )}
          {catalogSections.map(({ href, label }) => (
            <Link
              key={href}
              href={href}
              className={navigationLinkClassName}
            >
              {label}
            </Link>
          ))}
          <Link
            href="/about"
            className={navigationLinkClassName}
          >
            About
          </Link>
          <Link
            href="/contact"
            className={navigationLinkClassName}
          >
            Contact
          </Link>
        </nav>

        <div className="flex shrink-0 items-center gap-2 ms-auto">
          {isSearchOpen ? (
            <div className="absolute inset-x-0 top-full border-b bg-background p-3 shadow-sm sm:static sm:w-[260px] sm:border-0 sm:p-0 sm:shadow-none">
              <PlatformSearchForm autoFocus onClose={() => setIsSearchOpen(false)} />
            </div>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setIsSearchOpen(true)}
            >
              <Search className="h-5 w-5" />
              <span className="sr-only">Search</span>
            </Button>
          )}
          <LanguageSelector />
          <ThemeToggle />
          <Button variant="ghost" size="icon">
            <ShoppingCart className="h-5 w-5" />
            <span className="sr-only">Cart</span>
          </Button>

          {isAuthenticated ? (
            <UserMenu />
          ) : (
            <>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => handleAuthClick("login")}
              >
                <User className="h-5 w-5" />
                <span className="sr-only">Account</span>
              </Button>
              <Button
                className="hidden md:flex bg-teal-600 hover:bg-teal-700"
                onClick={() => handleAuthClick("login")}
              >
                Sign In
              </Button>
            </>
          )}
        </div>
      </div>

      <AuthModal
        isOpen={isAuthModalOpen}
        onClose={() => setIsAuthModalOpen(false)}
        initialMode={authMode}
      />
    </header>
  );
}
