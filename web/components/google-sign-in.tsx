"use client";

import { useRef, useState } from "react";
import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { firebaseGoogleError, getFirebaseGoogleAuth } from "@/lib/firebase-auth";
import { GoogleAuthProvider, signInWithPopup, signOut } from "firebase/auth";

export function GoogleSignIn({ onCredential, onError, disabled, onBusyChange }: {
  onCredential: (credential: string) => Promise<void>;
  onError: (message: string) => void;
  disabled: boolean;
  onBusyChange: (busy: boolean) => void;
}) {
  const [busy, setBusy] = useState(false);
  const pending = useRef(false);
  const configured = !!(process.env.NEXT_PUBLIC_FIREBASE_API_KEY && process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN && process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID && process.env.NEXT_PUBLIC_FIREBASE_APP_ID);

  const signIn = async () => {
    if (pending.current || disabled || !configured) return;
    pending.current = true;
    setBusy(true);
    onBusyChange(true);
    onError("");
    try {
      const auth = getFirebaseGoogleAuth();
      const provider = new GoogleAuthProvider();
      provider.setCustomParameters({ prompt: "select_account" });
      // Start the popup directly from the click, before any network await.
      const result = await signInWithPopup(auth, provider);
      try {
        await onCredential(await result.user.getIdToken());
      } finally {
        // Firebase only supplies identity; Qayen owns the application session.
        await signOut(auth).catch(() => {});
      }
    } catch (error) {
      const message = firebaseGoogleError(error);
      if (message) onError(message);
    } finally {
      pending.current = false;
      setBusy(false);
      onBusyChange(false);
    }
  };

  return <div className="space-y-2 text-center">
    <Button type="button" variant="outline" className="w-full gap-3" disabled={disabled || busy || !configured} onClick={() => void signIn()}>
      {busy ? <Loader2 className="h-5 w-5 animate-spin" aria-hidden="true" /> : <svg className="h-5 w-5" viewBox="0 0 24 24" aria-hidden="true">
        <path fill="#4285F4" d="M21.6 12.23c0-.71-.06-1.39-.18-2.05H12v3.88h5.38a4.6 4.6 0 0 1-2 3.02v2.51h3.24c1.9-1.75 2.98-4.33 2.98-7.36Z" />
        <path fill="#34A853" d="M12 22c2.7 0 4.96-.9 6.62-2.42l-3.24-2.51c-.9.6-2.05.96-3.38.96-2.6 0-4.8-1.76-5.59-4.12H3.06v2.59A10 10 0 0 0 12 22Z" />
        <path fill="#FBBC05" d="M6.41 13.91A6 6 0 0 1 6.1 12c0-.66.11-1.3.31-1.91V7.5H3.06A10 10 0 0 0 2 12c0 1.61.39 3.14 1.06 4.5l3.35-2.59Z" />
        <path fill="#EA4335" d="M12 5.97c1.47 0 2.79.51 3.82 1.5l2.87-2.87A9.6 9.6 0 0 0 12 2a10 10 0 0 0-8.94 5.5l3.35 2.59C7.2 7.73 9.4 5.97 12 5.97Z" />
      </svg>}
      {busy ? "Signing In..." : "Continue with Google"}
    </Button>
    {!configured && <p className="text-xs text-muted-foreground">Google sign-in is not configured yet.</p>}
  </div>;
}
