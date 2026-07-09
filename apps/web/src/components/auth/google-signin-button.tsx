"use client";

import { useEffect, useRef } from "react";
import { loginWithGoogle } from "@/lib/api";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";

const CLIENT_ID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;

type GoogleCredentialResponse = {
  credential: string;
};

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string;
            callback: (response: GoogleCredentialResponse) => void;
          }) => void;
          renderButton: (
            parent: HTMLElement,
            options: { theme?: string; size?: string; width?: number; text?: string }
          ) => void;
        };
      };
    };
  }
}

type GoogleSignInButtonProps = {
  onSuccess?: (token: string) => void;
};

export function GoogleSignInButton({ onSuccess }: GoogleSignInButtonProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { signIn } = useAuth();
  const { t } = useLocale();

  useEffect(() => {
    if (!CLIENT_ID || !containerRef.current) return;

    let cancelled = false;
    const mountButton = () => {
      if (cancelled || !containerRef.current || !window.google) return;
      window.google.accounts.id.initialize({
        client_id: CLIENT_ID,
        callback: async (response) => {
          const result = await loginWithGoogle(response.credential);
          if (result.ok) {
            signIn(result.data.token, result.data.user_id);
            onSuccess?.(result.data.token);
          }
        },
      });
      window.google.accounts.id.renderButton(containerRef.current, {
        theme: "outline",
        size: "large",
        width: 320,
        text: "continue_with",
      });
    };

    if (window.google) {
      mountButton();
      return;
    }

    const script = document.createElement("script");
    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;
    script.onload = mountButton;
    document.body.appendChild(script);

    return () => {
      cancelled = true;
      script.remove();
    };
  }, [onSuccess, signIn]);

  if (!CLIENT_ID) {
    return (
      <p className="text-center text-xs text-muted-foreground">{t.profile.googleUnavailable}</p>
    );
  }

  return (
    <div className="space-y-2">
      <div ref={containerRef} className="flex justify-center" />
    </div>
  );
}
