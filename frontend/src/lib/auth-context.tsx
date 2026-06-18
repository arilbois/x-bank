import { createContext, useContext, useEffect, useState } from "react";
import type { ReactNode } from "react";
import type { AuthResponse, User } from "./types";
import { api, setToken, clearAuth, USER_KEY } from "./api";

interface AuthContextValue {
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<User>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

function readStoredUser(): User | null {
  try {
    const raw = localStorage.getItem(USER_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as User;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(() => readStoredUser());
  const [loading, setLoading] = useState(false);

  // keep multiple tabs in sync
  useEffect(() => {
    function onStorage(e: StorageEvent) {
      if (e.key === USER_KEY) setUser(readStoredUser());
    }
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);

  async function login(username: string, password: string): Promise<User> {
    setLoading(true);
    try {
      const { data } = await api.post<AuthResponse>("/auth/login", {
        username,
        password,
      });
      setToken(data.token);
      localStorage.setItem(USER_KEY, JSON.stringify(data.user));
      setUser(data.user);
      return data.user;
    } catch (err) {
      clearAuth();
      setUser(null);
      throw err;
    } finally {
      setLoading(false);
    }
  }

  function logout() {
    clearAuth();
    setUser(null);
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
