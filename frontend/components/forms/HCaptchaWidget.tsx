"use client";

import { useEffect, useRef, useState } from "react";

import { env } from "@/lib/env";

declare global {
  interface Window {
    hcaptcha?: {
      render: (container: HTMLElement, options: Record<string, unknown>) => number;
      reset: (widgetId?: number) => void;
    };
  }
}

type Props = {
  onTokenChange?: (token: string) => void;
  size?: "normal" | "compact" | "invisible";
};

export function HCaptchaWidget({ onTokenChange, size = "normal" }: Props) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const widgetIdRef = useRef<number | null>(null);
  const [token, setToken] = useState("");
  const [scriptLoaded, setScriptLoaded] = useState(false);
  const siteKey = env.hcaptchaSiteKey;

  useEffect(() => {
    if (typeof window === "undefined" || !siteKey) {
      return;
    }

    if (window.hcaptcha) {
      setScriptLoaded(true);
      return;
    }

    const existing = document.querySelector<HTMLScriptElement>(
      'script[src="https://hcaptcha.com/1/api.js?render=explicit"]',
    );
    const loadListener = () => setScriptLoaded(true);

    if (existing) {
      existing.addEventListener("load", loadListener, { once: true });
      if (window.hcaptcha) {
        setScriptLoaded(true);
      }
      return () => existing.removeEventListener("load", loadListener);
    }

    const script = document.createElement("script");
    script.src = "https://hcaptcha.com/1/api.js?render=explicit";
    script.async = true;
    script.defer = true;
    script.addEventListener("load", loadListener, { once: true });
    document.head.appendChild(script);

    return () => {
      script.removeEventListener("load", loadListener);
    };
  }, [siteKey]);

  useEffect(() => {
    if (!siteKey || !scriptLoaded || !containerRef.current || !window.hcaptcha) {
      return;
    }
    if (widgetIdRef.current !== null) {
      return;
    }

    widgetIdRef.current = window.hcaptcha.render(containerRef.current, {
      sitekey: siteKey,
      size,
      callback: (value: string) => {
        setToken(value);
        onTokenChange?.(value);
      },
      "expired-callback": () => {
        setToken("");
        onTokenChange?.("");
      },
    });
  }, [scriptLoaded, siteKey, size, onTokenChange]);

  useEffect(() => {
    return () => {
      if (widgetIdRef.current !== null && window.hcaptcha) {
        window.hcaptcha.reset(widgetIdRef.current);
      }
    };
  }, []);

  if (!siteKey) {
    return (
      <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-xs text-amber-200">
        hCaptcha site anahtarını tanımlamak için `NEXT_PUBLIC_HCAPTCHA_SITEKEY` ortam değişkenini
        ayarlayın.
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div ref={containerRef} className="flex justify-center" />
      <input type="hidden" name="hcaptcha_token" value={token} readOnly />
    </div>
  );
}
