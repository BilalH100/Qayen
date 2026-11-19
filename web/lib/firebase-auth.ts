"use client";

import { getApps, initializeApp } from "firebase/app";
import { browserPopupRedirectResolver, getAuth, initializeAuth, inMemoryPersistence, type Auth } from "firebase/auth";

let auth: Auth | undefined;

export function getFirebaseGoogleAuth(): Auth {
  if (auth) return auth;
  const config = {
    apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
    authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
    projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
    appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID,
  };
  if (!config.apiKey || !config.authDomain || !config.projectId || !config.appId) throw new Error("Google sign-in is not configured yet.");
  const existing = getApps().find(app => app.name === "qayen-google-sign-in");
  if (existing) {
    auth = getAuth(existing);
  } else {
    const app = initializeApp(config, "qayen-google-sign-in");
    auth = initializeAuth(app, { persistence: inMemoryPersistence, popupRedirectResolver: browserPopupRedirectResolver });
  }
  return auth;
}

export function firebaseGoogleError(error: unknown): string {
  const code = (error as { code?: string })?.code;
  switch (code) {
    case "auth/popup-closed-by-user":
    case "auth/cancelled-popup-request": return "";
    case "auth/popup-blocked": return "Your browser blocked the Google sign-in popup. Please allow popups and try again.";
    case "auth/unauthorized-domain": return "Add this website address to Firebase Authentication's authorized domains.";
    case "auth/operation-not-allowed": return "Enable Google sign-in in Firebase Authentication.";
    case "auth/network-request-failed": return "Google sign-in could not connect. Check your internet connection and try again.";
    case "auth/invalid-api-key":
    case "auth/configuration-not-found": return "Check the Firebase configuration and enable Authentication.";
    default: return "Google sign-in failed. Please try again.";
  }
}
