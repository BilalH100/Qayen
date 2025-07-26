"use client"

import { useState } from "react"
import Link from "next/link"
import { Menu, X, Search, User, ShoppingCart, Pill } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ThemeToggle } from "@/components/theme-toggle"
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"
import { AuthModal } from "@/components/auth-modal"
import { UserMenu } from "@/components/user-menu"
import { useAuth } from "@/contexts/auth-context"
import Image from "next/image"
import logo from "../public/logo.png"

export function Navbar() {
  const [isSearchOpen, setIsSearchOpen] = useState(false)
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false)
  const [authMode, setAuthMode] = useState<"login" | "register">("login")
  
  const { user, isAuthenticated, isLoading } = useAuth()

  const handleAuthClick = (mode: "login" | "register") => {
    setAuthMode(mode)
    setIsAuthModalOpen(true)
  }

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center">
        <Sheet>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="md:hidden">
              <Menu className="h-5 w-5" />
              <span className="sr-only">Toggle menu</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-[300px] sm:w-[400px]">
            <nav className="flex flex-col gap-4 mt-8">
              <Link href="/" className="text-lg font-semibold">
                Home
              </Link>
              <Link href="/medications" className="text-lg font-semibold">
                Medications
              </Link>
              <Link href="/pharmacies" className="text-lg font-semibold">
                Pharmacies
              </Link>
              <Link href="/about" className="text-lg font-semibold">
                About
              </Link>
              <Link href="/contact" className="text-lg font-semibold">
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

        <Link href="/" className="flex items-center gap-2 mr-6">
          {/* <div className="bg-teal-600 text-white p-1 rounded">
            <Image src={logo} alt={""} height={50} width={50}/>
          </div> */}
          <span className="font-bold text-xl hidden sm:inline-block p-2">Kayena</span>
        </Link>

        <nav className="hidden md:flex items-center gap-6 text-sm font-medium flex-1">
          <Link href="/medications" className="transition-colors hover:text-foreground/80">
            Medications
          </Link>
          <Link href="/pharmacies" className="transition-colors hover:text-foreground/80">
            Pharmacies
          </Link>
          <Link href="/about" className="transition-colors hover:text-foreground/80">
            About
          </Link>
          <Link href="/contact" className="transition-colors hover:text-foreground/80">
            Contact
          </Link>
        </nav>

        <div className="flex items-center gap-2 ml-auto">
          {isSearchOpen ? (
            <div className="relative flex items-center">
              <Input
                placeholder="Search..."
                className="w-[200px] md:w-[300px] pr-8"
                autoFocus
                onBlur={() => setIsSearchOpen(false)}
              />
              <Button variant="ghost" size="icon" className="absolute right-0" onClick={() => setIsSearchOpen(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
          ) : (
            <Button variant="ghost" size="icon" onClick={() => setIsSearchOpen(true)}>
              <Search className="h-5 w-5" />
              <span className="sr-only">Search</span>
            </Button>
          )}
          <ThemeToggle />
          <Button variant="ghost" size="icon">
            <ShoppingCart className="h-5 w-5" />
            <span className="sr-only">Cart</span>
          </Button>
          
          {isLoading ? (
            <div className="w-8 h-8 animate-pulse bg-gray-200 rounded-full" />
          ) : isAuthenticated ? (
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
  )
}
