import { type ReactNode, createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useAuth } from "../auth/AuthContext";
import { getMe, updateLocale as updateLocaleOnServer } from "../api";
import { type Locale, translations } from "./translations";

const STORAGE_KEY = "locale";
const DEFAULT_LOCALE: Locale = "de";

function readStoredLocale(): Locale {
  const stored = window.localStorage.getItem(STORAGE_KEY);
  return stored === "en" ? "en" : DEFAULT_LOCALE;
}

type LocaleContextValue = {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: string) => string;
};

const LocaleContext = createContext<LocaleContextValue | undefined>(undefined);

export function LocaleProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const [locale, setLocaleState] = useState<Locale>(readStoredLocale);

  const setLocale = useCallback((next: Locale) => {
    setLocaleState(next);
    window.localStorage.setItem(STORAGE_KEY, next);
    if (user) {
      void updateLocaleOnServer(next).catch(() => undefined);
    }
  }, [user]);

  // On sign-in, the backend's saved preference takes precedence over the local default.
  useEffect(() => {
    if (!user) return;
    void getMe().then((me) => {
      if (me.locale === "de" || me.locale === "en") {
        setLocaleState(me.locale);
        window.localStorage.setItem(STORAGE_KEY, me.locale);
      }
    }).catch(() => undefined);
  }, [user]);

  const t = useCallback((key: string) => translations[locale][key] ?? key, [locale]);

  const value = useMemo(() => ({ locale, setLocale, t }), [locale, setLocale, t]);

  return <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>;
}

export function useLocale() {
  const ctx = useContext(LocaleContext);
  if (!ctx) {
    throw new Error("useLocale must be used within a LocaleProvider");
  }
  return ctx;
}
