"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { AuthForm } from "@/components/auth/auth-form";
import { useAuth } from "@/components/auth/auth-provider";
import { OnboardingModal } from "@/components/onboarding/onboarding-modal";
import { useLocale } from "@/components/locale/locale-provider";

function LoginContent() {
  const { isAuthenticated, ready } = useAuth();
  const { t } = useLocale();
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirect = searchParams.get("redirect") || "/";
  const [onboardingToken, setOnboardingToken] = useState<string | null>(null);

  useEffect(() => {
    if (ready && isAuthenticated && !onboardingToken) {
      router.replace(redirect);
    }
  }, [ready, isAuthenticated, redirect, router, onboardingToken]);

  if (!ready) return null;

  if (onboardingToken) {
    return (
      <OnboardingModal open token={onboardingToken} onDone={() => router.replace(redirect)} />
    );
  }

  if (isAuthenticated) return null;

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t.auth.loginTitle}</h1>
        <p className="text-sm text-muted-foreground">{t.auth.loginSubtitle}</p>
      </header>
      <AuthForm
        onSuccess={(token, meta) => {
          if (meta?.isRegister) {
            setOnboardingToken(token);
          } else {
            router.replace(redirect);
          }
        }}
      />
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={null}>
      <LoginContent />
    </Suspense>
  );
}
