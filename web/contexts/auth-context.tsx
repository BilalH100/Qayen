"use client";

import React, { createContext, useContext, useState, useEffect } from "react";
import axios from "axios";
import { toast } from "sonner";
import { BASE_URL } from "@/utils/api";
import { ROLES } from "@/utils/types";
import { rememberAccountEmail } from "@/lib/remembered-accounts";
interface User {
  id: string;
  email: string;
  name: string;
  phone: string;
  role: string;
  managed_pharmacy?: number;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  loginWithGoogle: (credential: string) => Promise<void>;
  updateProfile: (data: { name: string; phone: string }) => Promise<void>;
  register: (
    email: string,
    password: string,
    name: string,
    phone: string,
  ) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const isAuthenticated = !!user;

  useEffect(() => {
    const token = sessionStorage.getItem("token");
    const userData = sessionStorage.getItem("user");

    if (token && userData) {
      try {
        const parsedUser = JSON.parse(userData);
        setUser(parsedUser);
        axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;
      } catch (error) {
        sessionStorage.removeItem("token");
        sessionStorage.removeItem("user");
      }
    }
    setIsLoading(false);
  }, []);

  const completeLogin = (data: { token: string; user: User }) => {
    if (!data.token || !data.user?.email) throw new Error("Invalid response format");
    sessionStorage.setItem("token", data.token);
    sessionStorage.setItem("user", JSON.stringify(data.user));
    axios.defaults.headers.common["Authorization"] = `Bearer ${data.token}`;
    const remembered = rememberAccountEmail(data.user.email);
    setUser(data.user);
    toast.success(`Welcome back, ${data.user.name}!`);
    if (!remembered) toast.message("Signed in, but your browser could not remember this account.");
  };

  const loginWithGoogle = async (credential: string) => {
    try {
      const challenge = await axios.get(`${BASE_URL}/users/google/nonce`, { withCredentials: true });
      const response = await axios.post(`${BASE_URL}/users/google/firebase`, { credential, nonce: challenge.data.nonce }, { withCredentials: true });
      completeLogin(response.data);
    } catch (error: any) {
      throw new Error(error.response?.data?.message || "Google sign-in failed. Please try again.");
    }
  };

  const login = async (email: string, password: string) => {
    try {
      const response = await axios.post(`${BASE_URL}/users/login`, {
        email,
        password,
      });

      completeLogin(response.data);
    } catch (error: any) {
      console.error("Login error:", error);
      throw new Error(error.response?.data?.message || "Login failed");
    }
  };

  const register = async (
    email: string,
    password: string,
    name: string,
    phone: string,
  ) => {
    try {
      const response = await axios.post(`${BASE_URL}/users/register`, {
        email,
        password,
        name,
        phone,
      });

      if (response.data.message) {
        toast.success("Account created successfully!");
        await login(email, password);
      } else {
        throw new Error("Registration failed");
      }
    } catch (error: any) {
      console.error("Registration error:", error);
      throw new Error(error.response?.data?.message || "Registration failed");
    }
  };

  const updateProfile = async (data: { name: string; phone: string }) => {
    try {
      const token = sessionStorage.getItem("token");
      if (!token) throw new Error("Please sign in again to edit your profile.");
      const response = await axios.patch<User>(`${BASE_URL}/users/me`, data, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (sessionStorage.getItem("token") !== token) throw new Error("Please sign in again to edit your profile.");
      sessionStorage.setItem("user", JSON.stringify(response.data));
      setUser(response.data);
      toast.success("Profile updated successfully.");
    } catch (error: any) {
      throw new Error(error.response?.data?.message || error.message || "Could not save your profile. Please try again.");
    }
  };

  const logout = () => {
    sessionStorage.removeItem("token");
    sessionStorage.removeItem("user");
    delete axios.defaults.headers.common["Authorization"];
    setUser(null);
    toast.success("You've been signed out successfully!");
  };

  const value = {
    user,
    isLoading,
    login,
    loginWithGoogle,
    updateProfile,
    register,
    logout,
    isAuthenticated,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
