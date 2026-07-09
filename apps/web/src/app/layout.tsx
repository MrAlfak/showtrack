import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { AuthProvider } from "@/components/auth/auth-provider";
import { LocaleProvider } from "@/components/locale/locale-provider";
import { BottomNav } from "@/components/layout/bottom-nav";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "ShowTrack — TV Show Tracker",
  description: "Track shows, discover new series, and explore cast — self-hosted TV tracker",
  manifest: "/manifest.json",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} dark h-full antialiased`}
    >
      <body className="min-h-full bg-background text-foreground">
        <LocaleProvider>
          <AuthProvider>
            <main className="mx-auto min-h-screen max-w-lg px-4 pb-24 pt-6">
              {children}
            </main>
            <BottomNav />
          </AuthProvider>
        </LocaleProvider>
      </body>
    </html>
  );
}
