"use client";

import { createContext, useContext, useMemo, useState } from "react";

type AuthContextValue = {
  token: string | null;
  userId: string | null;
  isAuthenticated: boolean;
  ready: boolean;
  signIn: (token: string, userId: string) => void;
  signOut: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

const TOKEN_KEY = "showtrack.token";
const USER_KEY = "showtrack.userId";

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() =>
    typeof window === "undefined" ? null : window.localStorage.getItem(TOKEN_KEY)
  );
  const [userId, setUserId] = useState<string | null>(() =>
    typeof window === "undefined" ? null : window.localStorage.getItem(USER_KEY)
  );
  const ready = true;

  const value = useMemo<AuthContextValue>(
    () => ({
      token,
      userId,
      ready,
      isAuthenticated: Boolean(token),
      signIn(nextToken, nextUserId) {
        setToken(nextToken);
        setUserId(nextUserId);
        window.localStorage.setItem(TOKEN_KEY, nextToken);
        window.localStorage.setItem(USER_KEY, nextUserId);
      },
      signOut() {
        setToken(null);
        setUserId(null);
        window.localStorage.removeItem(TOKEN_KEY);
        window.localStorage.removeItem(USER_KEY);
      },
    }),
    [ready, token, userId]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
