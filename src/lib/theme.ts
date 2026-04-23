// Shared theme helpers — single source of truth for persisting & reading
// the user's light/dark preference. The initial application happens in
// index.html (pre-paint) to avoid flash-of-wrong-theme.

export const THEME_STORAGE_KEY = "gitmap-theme";

export type Theme = "light" | "dark";

/** Read the currently applied theme from the <html> element. */
export function getCurrentTheme(): Theme {
  if (typeof document === "undefined") return "light";
  return document.documentElement.classList.contains("dark") ? "dark" : "light";
}

/** Apply a theme to the <html> element AND persist it to localStorage. */
export function setTheme(theme: Theme): void {
  if (typeof document === "undefined") return;
  document.documentElement.classList.toggle("dark", theme === "dark");
  try {
    localStorage.setItem(THEME_STORAGE_KEY, theme);
  } catch {
    /* localStorage may be unavailable (private mode, SSR) — silently ignore */
  }
}

/** Toggle between light and dark, persisting the new value. */
export function toggleTheme(): Theme {
  const next: Theme = getCurrentTheme() === "dark" ? "light" : "dark";
  setTheme(next);
  return next;
}
