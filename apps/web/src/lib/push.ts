import { registerDevice } from "@/lib/api";

type FirebaseConfig = {
  apiKey: string;
  projectId: string;
  messagingSenderId: string;
  appId: string;
  vapidKey: string;
};

export function isPushConfigured() {
  return Boolean(
    process.env.NEXT_PUBLIC_FIREBASE_API_KEY &&
      process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID &&
      process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID &&
      process.env.NEXT_PUBLIC_FIREBASE_APP_ID &&
      process.env.NEXT_PUBLIC_FIREBASE_VAPID_KEY
  );
}

function getFirebaseConfig(): FirebaseConfig | null {
  const apiKey = process.env.NEXT_PUBLIC_FIREBASE_API_KEY;
  const projectId = process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID;
  const messagingSenderId = process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID;
  const appId = process.env.NEXT_PUBLIC_FIREBASE_APP_ID;
  const vapidKey = process.env.NEXT_PUBLIC_FIREBASE_VAPID_KEY;
  if (!apiKey || !projectId || !messagingSenderId || !appId || !vapidKey) {
    return null;
  }
  return { apiKey, projectId, messagingSenderId, appId, vapidKey };
}

async function registerServiceWorker() {
  if (!("serviceWorker" in navigator)) {
    return null;
  }
  return navigator.serviceWorker.register("/firebase-messaging-sw.js");
}

export async function enableWebPush(authToken: string): Promise<{ ok: boolean; reason?: string }> {
  if (!isPushConfigured()) {
    return { ok: false, reason: "unsupported" };
  }
  if (!("Notification" in window)) {
    return { ok: false, reason: "unsupported" };
  }

  const permission = await Notification.requestPermission();
  if (permission !== "granted") {
    return { ok: false, reason: "denied" };
  }

  const config = getFirebaseConfig();
  if (!config) {
    return { ok: false, reason: "unsupported" };
  }

  await registerServiceWorker();

  const { initializeApp, getApps } = await import("firebase/app");
  const { getMessaging, getToken, isSupported } = await import("firebase/messaging");

  if (!(await isSupported())) {
    return { ok: false, reason: "unsupported" };
  }

  const app = getApps().length
    ? getApps()[0]
    : initializeApp({
        apiKey: config.apiKey,
        projectId: config.projectId,
        messagingSenderId: config.messagingSenderId,
        appId: config.appId,
      });

  const messaging = getMessaging(app);
  const fcmToken = await getToken(messaging, { vapidKey: config.vapidKey });
  if (!fcmToken) {
    return { ok: false, reason: "token" };
  }

  const response = await registerDevice(authToken, {
    token: fcmToken,
    platform: "web",
    app_version: "web-pwa",
  });

  return response?.ok ? { ok: true } : { ok: false, reason: "register" };
}
