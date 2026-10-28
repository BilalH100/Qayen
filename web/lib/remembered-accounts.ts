// Keep the existing key so accounts saved by earlier versions remain available.
export const SAVED_ACCOUNTS_KEY = "kayena_saved_accounts";
export const SAVED_ACCOUNTS_CHANGED = "qayen:saved-accounts-changed";

function cleanAccounts(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return [...new Set(value.filter((email): email is string => typeof email === "string")
    .map(email => email.trim().toLowerCase())
    .filter(email => /^[^\s@]+@[^\s@]+$/.test(email)))];
}

export function readRememberedAccounts(): string[] {
  try {
    return cleanAccounts(JSON.parse(window.localStorage.getItem(SAVED_ACCOUNTS_KEY) ?? "[]"));
  } catch {
    // A malformed legacy value must not prevent the next successful login
    // from replacing it with a valid list.
    return [];
  }
}

function writeAccounts(accounts: string[]): boolean {
  try {
    window.localStorage.setItem(SAVED_ACCOUNTS_KEY, JSON.stringify(accounts));
    window.dispatchEvent(new Event(SAVED_ACCOUNTS_CHANGED));
    return true;
  } catch {
    return false;
  }
}

export function rememberAccountEmail(email: string): boolean {
  const [account] = cleanAccounts([email]);
  if (!account) return false;
  const existing = readRememberedAccounts().filter(value => value !== account);
  return writeAccounts([account, ...existing]);
}

export function forgetAccountEmail(email: string): boolean {
  const account = email.trim().toLowerCase();
  return writeAccounts(readRememberedAccounts().filter(value => value !== account));
}
