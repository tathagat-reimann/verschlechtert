import type { ReactNode } from "react";
import { LocalizationProvider } from "@mui/x-date-pickers/LocalizationProvider";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { de, enUS } from "date-fns/locale";
import { useLocale } from "../i18n/LocaleContext";

export function AppDatePickerProvider({ children }: { children: ReactNode }) {
  const { locale } = useLocale();
  return (
    <LocalizationProvider dateAdapter={AdapterDateFns} adapterLocale={locale === "de" ? de : enUS}>
      {children}
    </LocalizationProvider>
  );
}
