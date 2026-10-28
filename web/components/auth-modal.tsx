"use client";

import { useEffect, useState } from "react";
import { Eye, EyeOff, Loader2, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useAuth } from "@/contexts/auth-context";
import { forgetAccountEmail, readRememberedAccounts, SAVED_ACCOUNTS_CHANGED, SAVED_ACCOUNTS_KEY } from "@/lib/remembered-accounts";

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialMode?: "login" | "register";
}

export function AuthModal({
  isOpen,
  onClose,
  initialMode = "login",
}: AuthModalProps) {
  const [mode, setMode] = useState<"login" | "register">(initialMode);
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [savedAccounts, setSavedAccounts] = useState<string[]>([]);

  const [formData, setFormData] = useState({
    email: "",
    password: "",
    name: "",
    phone: "",
  });

  const { login, register } = useAuth();

  // Load previously used accounts whenever the login modal is opened
  useEffect(() => {
    if (!isOpen) return;

    setMode(initialMode);
    setError("");
    setShowPassword(false);

    setFormData({
      email: "",
      password: "",
      name: "",
      phone: "",
    });

    if (initialMode === "login") {
      loadSavedAccounts();
    }
  }, [isOpen, initialMode]);

  const loadSavedAccounts = () => {
    setSavedAccounts(readRememberedAccounts());
  };

  useEffect(() => {
    if (!isOpen) return;
    const reload = () => setSavedAccounts(readRememberedAccounts());
    const onStorage = (event: StorageEvent) => {
      if (event.key === SAVED_ACCOUNTS_KEY || event.key === null) reload();
    };
    window.addEventListener("storage", onStorage);
    window.addEventListener(SAVED_ACCOUNTS_CHANGED, reload);
    return () => {
      window.removeEventListener("storage", onStorage);
      window.removeEventListener(SAVED_ACCOUNTS_CHANGED, reload);
    };
  }, [isOpen]);

  const removeSavedAccount = (
    email: string,
    event: React.MouseEvent
  ) => {
    event.stopPropagation();

    if (!forgetAccountEmail(email)) {
      setError("Your browser could not update saved accounts.");
      return;
    }
    loadSavedAccounts();

    // If the removed account is currently selected, clear the email
    if (formData.email.toLowerCase() === email.toLowerCase()) {
      setFormData((prev) => ({
        ...prev,
        email: "",
        password: "",
      }));
    }
  };

  const selectSavedAccount = (email: string) => {
    setFormData((prev) => ({
      ...prev,
      email,
      password: "",
    }));

    setError("");
  };

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const { name, value } = e.target;

    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    setError("");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setIsLoading(true);
    setError("");

    try {
      if (mode === "login") {
        await login(formData.email, formData.password);
      } else {
        await register(
          formData.email,
          formData.password,
          formData.name,
          formData.phone
        );
      }

      setFormData({
        email: "",
        password: "",
        name: "",
        phone: "",
      });

      onClose();
    } catch (err: any) {
      setError(err.message || "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  const switchMode = () => {
    const newMode = mode === "login" ? "register" : "login";

    setMode(newMode);
    setError("");
    setShowPassword(false);

    setFormData({
      email: "",
      password: "",
      name: "",
      phone: "",
    });

    if (newMode === "login") {
      loadSavedAccounts();
    }
  };

  const handleDialogChange = (open: boolean) => {
    if (!open && !isLoading) {
      onClose();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleDialogChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="text-2xl font-bold text-center">
            {mode === "login" ? "Welcome Back" : "Create Account"}
          </DialogTitle>

          <DialogDescription className="text-center">
            {mode === "login"
              ? "Sign in to your account to continue"
              : "Sign up to access personalized features"}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 mt-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Previously used accounts */}
          {mode === "login" && savedAccounts.length > 0 && (
            <div className="space-y-2">
              <Label>Choose an account</Label>

              <div className="max-h-48 space-y-2 overflow-y-auto">
                {savedAccounts.map((email) => (
                  <div
                    key={email}
                    className={`flex items-center justify-between rounded-md border p-3 cursor-pointer transition-colors ${
                      formData.email.toLowerCase() ===
                      email.toLowerCase()
                        ? "border-teal-600 bg-teal-50"
                        : "border-gray-200 hover:bg-gray-50"
                    }`}
                  >
                    <button type="button" className="flex min-w-0 flex-1 items-center gap-3 text-left" onClick={() => selectSavedAccount(email)} disabled={isLoading}>
                      <div className="h-8 w-8 rounded-full bg-teal-100 flex items-center justify-center text-teal-700 font-semibold">
                        {email.charAt(0).toUpperCase()}
                      </div>

                      <span className="text-sm truncate" translate="no">
                        {email}
                      </span>
                    </button>

                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 shrink-0"
                      onClick={(event) =>
                        removeSavedAccount(email, event)
                      }
                      disabled={isLoading}
                      aria-label={`Remove ${email}`}
                    >
                      <X className="h-4 w-4 text-gray-400" />
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {mode === "register" && (
            <>
              <div className="space-y-2">
                <Label htmlFor="name">Full Name</Label>

                <Input
                  id="name"
                  name="name"
                  type="text"
                  placeholder="John Doe"
                  value={formData.name}
                  onChange={handleInputChange}
                  required
                  disabled={isLoading}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="phone">Phone Number</Label>

                <Input
                  id="phone"
                  name="phone"
                  type="tel"
                  placeholder="+1 (555) 123-4567"
                  value={formData.phone}
                  onChange={handleInputChange}
                  required
                  disabled={isLoading}
                />
              </div>
            </>
          )}

          {/* Email */}
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>

            <Input
              id="email"
              name="email"
              type="email"
              autoComplete="username"
              placeholder="john@example.com"
              value={formData.email}
              onChange={handleInputChange}
              required
              disabled={isLoading}
            />
          </div>

          {/* Password */}
          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>

            <div className="relative">
              <Input
                id="password"
                name="password"
                type={showPassword ? "text" : "password"}
                autoComplete={mode === "login" ? "current-password" : "new-password"}
                placeholder="Enter your password"
                value={formData.password}
                onChange={handleInputChange}
                required
                disabled={isLoading}
                className="pr-10"
              />

              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                onClick={() =>
                  setShowPassword(!showPassword)
                }
                disabled={isLoading}
              >
                {showPassword ? (
                  <EyeOff className="h-4 w-4 text-gray-400" />
                ) : (
                  <Eye className="h-4 w-4 text-gray-400" />
                )}
              </Button>
            </div>
          </div>

          {/* Submit */}
          <Button
            type="submit"
            className="w-full bg-teal-600 hover:bg-teal-700"
            disabled={isLoading}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />

                {mode === "login"
                  ? "Signing In..."
                  : "Creating Account..."}
              </>
            ) : mode === "login" ? (
              "Sign In"
            ) : (
              "Create Account"
            )}
          </Button>

          {/* Switch Login/Register */}
          <div className="text-center text-sm">
            <span className="text-gray-600">
              {mode === "login"
                ? "Don't have an account?"
                : "Already have an account?"}
            </span>

            <Button
              type="button"
              variant="link"
              className="p-0 ml-1 h-auto text-teal-600 hover:text-teal-700"
              onClick={switchMode}
              disabled={isLoading}
            >
              {mode === "login" ? "Sign up" : "Sign in"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
