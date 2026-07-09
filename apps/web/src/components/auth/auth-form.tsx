"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { LogIn } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { login, register } from "@/lib/api";
import { GoogleSignInButton } from "@/components/auth/google-signin-button";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

type AuthFormProps = {
  onSuccess?: (token: string, meta?: { isRegister?: boolean }) => void;
};

export function AuthForm({ onSuccess }: AuthFormProps) {
  const { signIn } = useAuth();
  const { t } = useLocale();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  return (
    <Card className="tv-card shadow-none">
      <CardHeader>
        <CardTitle className="text-base">{mode === "login" ? t.profile.signIn : t.profile.register}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {mode === "register" && (
          <Input
            placeholder={t.profile.displayName}
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
            autoComplete="name"
          />
        )}
        <Input
          placeholder={t.profile.email}
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          autoComplete="email"
        />
        <Input
          type="password"
          placeholder={t.profile.password}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          autoComplete={mode === "login" ? "current-password" : "new-password"}
        />
        {error ? <p className="text-xs text-destructive">{error}</p> : null}
        <Button
          className="w-full"
          disabled={submitting || !email || !password}
          onClick={async () => {
            setSubmitting(true);
            setError("");
            const response =
              mode === "login"
                ? await login({ email, password })
                : await register({ email, password, display_name: displayName });
            if (response.ok) {
              signIn(response.data.token, response.data.user_id);
              setEmail("");
              setPassword("");
              setDisplayName("");
              onSuccess?.(response.data.token, { isRegister: mode === "register" });
            } else {
              setError(response.error);
            }
            setSubmitting(false);
          }}
        >
          {mode === "login" ? <LogIn className="mr-2 size-4" /> : null}
          {mode === "login" ? t.profile.signIn : t.profile.register}
        </Button>
        <Button variant="ghost" className="w-full" onClick={() => setMode(mode === "login" ? "register" : "login")}>
          {mode === "login" ? t.profile.needAccount : t.profile.haveAccount}
        </Button>
        <div className="space-y-3">
          <div className="flex items-center gap-3">
            <Separator className="flex-1" />
            <span className="text-xs text-muted-foreground">{t.profile.orContinueWith}</span>
            <Separator className="flex-1" />
          </div>
          <GoogleSignInButton onSuccess={(token) => onSuccess?.(token)} />
        </div>
      </CardContent>
    </Card>
  );
}

export function SignInPrompt({ compact = false }: { compact?: boolean }) {
  const { t } = useLocale();
  const pathname = usePathname();
  const loginHref = `/login?redirect=${encodeURIComponent(pathname)}`;

  if (compact) {
    return (
      <Link
        href={loginHref}
        className="inline-flex h-8 items-center justify-center gap-2 rounded-md border border-input bg-background px-3 text-xs font-medium hover:bg-accent hover:text-accent-foreground"
      >
        <LogIn className="size-4" />
        {t.auth.signIn}
      </Link>
    );
  }

  return (
    <Card className="tv-card shadow-none">
      <CardContent className="space-y-3 p-4">
        <div>
          <p className="text-sm font-medium">{t.auth.signInRequired}</p>
          <p className="text-xs text-muted-foreground">{t.auth.signInHint}</p>
        </div>
        <Link
          href={loginHref}
          className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <LogIn className="size-4" />
          {t.auth.signInOrRegister}
        </Link>
      </CardContent>
    </Card>
  );
}
