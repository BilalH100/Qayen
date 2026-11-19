"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import logo from "../public/LogoHomePage.png";
import { Menu, Search, User, ShoppingCart, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
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
  { href: "/pharmacies", label: "Pharmacies" },
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
            <Button variant="ghost" size="icon" className="xl:hidden">
              <Menu className="h-5 w-5" />
              <span className="sr-only">Toggle menu</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-[300px] overflow-y-auto sm:w-[400px]">
            <nav className="mt-8 flex flex-col gap-2">
              <SheetClose asChild>
                <Link href="/" className={navigationLinkClassName}>Home</Link>
              </SheetClose>
              <SheetClose asChild>
                <Link href="/medications" className={navigationLinkClassName}>Medications</Link>
              </SheetClose>
              <details className="group rounded-lg border border-teal-100 dark:border-slate-700">
                <summary className="flex cursor-pointer list-none items-center justify-between rounded-lg bg-teal-50/70 px-3 py-2 text-sm font-medium dark:bg-slate-800/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-teal-500 [&::-webkit-details-marker]:hidden">
                  More <ChevronDown className="h-4 w-4 transition-transform group-open:rotate-180" aria-hidden="true" />
                </summary>
                <div className="flex flex-col gap-1 p-2">
                  {catalogSections.map(({ href, label }) => (
                    <SheetClose asChild key={href}>
                      <Link href={href} className={navigationLinkClassName}>{label}</Link>
                    </SheetClose>
                  ))}
                </div>
              </details>
              <div className="mt-4 flex flex-col gap-2 border-t pt-4">
                <SheetClose asChild>
                  <Link href="/about" className={navigationLinkClassName}>About</Link>
                </SheetClose>
                <SheetClose asChild>
                  <Link href="/contact" className={navigationLinkClassName}>Contact</Link>
                </SheetClose>
                {user?.role === ROLES.PHARMACIST && (
                  <SheetClose asChild>
                    <Link href="/pharmacy" className={navigationLinkClassName}>Pharmacy Dashboard</Link>
                  </SheetClose>
                )}
                {user?.role === ROLES.ADMIN && (
                  <SheetClose asChild>
                    <Link href="/admin" className={navigationLinkClassName}>Admin Dashboard</Link>
                  </SheetClose>
                )}
              </div>

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

        <Link href="/" className="relative me-3 block h-14 w-36 shrink-0 overflow-hidden rounded-md sm:me-4 sm:w-40">
          <Image
            src={logo}
            alt="Qayen"
            fill
            sizes="(min-width: 640px) 160px, 144px"
            priority
            className="object-cover mix-blend-multiply brightness-105"
            style={{ objectPosition: "50% 40%" }}
          />
        </Link>

        <nav className="hidden min-w-0 flex-1 items-center gap-2 py-2 pe-3 xl:flex">
          <div className="flex shrink-0 items-center gap-2">
            <Link href="/medications" className={navigationLinkClassName}>Medications</Link>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button type="button" className={`${navigationLinkClassName} gap-1.5 group`}>
                  More <ChevronDown className="h-4 w-4 transition-transform group-data-[state=open]:rotate-180" aria-hidden="true" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="min-w-56">
                {catalogSections.map(({ href, label }) => (
                  <DropdownMenuItem key={href} asChild className={`${navigationLinkClassName} my-1 w-full cursor-pointer focus:bg-teal-100 focus:text-teal-900 dark:focus:bg-teal-950 dark:focus:text-teal-300`}>
                    <Link href={href}>{label}</Link>
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <div className="flex shrink-0 items-center gap-2 whitespace-nowrap">
            <Link href="/about" className={navigationLinkClassName}>About</Link>
            <Link href="/contact" className={navigationLinkClassName}>Contact</Link>
            {user?.role === ROLES.PHARMACIST && (
              <Link href="/pharmacy" className={navigationLinkClassName}>Pharmacy Dashboard</Link>
            )}
            {user?.role === ROLES.ADMIN && (
              <Link href="/admin" className={navigationLinkClassName}>Admin Dashboard</Link>
            )}
          </div>
        </nav>

        <div className="flex shrink-0 items-center gap-2 ms-auto">
          {isSearchOpen ? (
            <div className="absolute inset-x-0 top-full border-b bg-background p-3 shadow-sm">
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
