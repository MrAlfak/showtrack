import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Vazirmatn } from "next/font/google";
import { AuthProvider } from "@/components/auth/auth-provider";
import { LocaleProvider } from "@/components/locale/locale-provider";
import { AppHeader } from "@/components/layout/app-header";
import { BottomNav } from "@/components/layout/bottom-nav";
import { InstallPrompt } from "@/components/pwa/install-prompt";
import { TmdbAttribution } from "@/components/layout/tmdb-attribution";
import { APP_MAX_WIDTH } from "@/lib/layout";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const vazirmatn = Vazirmatn({
  variable: "--font-vazirmatn",
  subsets: ["arabic", "latin"],
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "ShowTrack — TV Show Tracker",
  description: "Track shows, discover new series, and explore cast — self-hosted TV tracker",
  manifest: "/manifest.json",
  appleWebApp: {
    capable: true,
    statusBarStyle: "black-translucent",
    title: "ShowTrack",
  },
  icons: {
    icon: "/icons/icon.svg",
    apple: "/icons/icon.svg",
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  themeColor: "#ffd60a",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} ${vazirmatn.variable} dark h-full antialiased`}
    >
      <body className="min-h-full">
        <LocaleProvider>
          <AuthProvider>
            <div className="flex min-h-full justify-center md:py-4">
              <div
                className="relative flex min-h-screen w-full flex-col bg-background md:min-h-[calc(100dvh-2rem)] md:overflow-hidden md:rounded-[1.75rem] md:border md:border-white/10 md:shadow-[0_0_80px_rgba(0,0,0,0.65)]"
                style={{ maxWidth: APP_MAX_WIDTH }}
              >
                <div className="flex-1 px-4 pb-24 pt-3">
                  <AppHeader />
                  <InstallPrompt />
                  {children}
                  <TmdbAttribution />
                </div>
                <BottomNav />
              </div>
            </div>
          </AuthProvider>
        </LocaleProvider>
      </body>
    </html>
  );
}
