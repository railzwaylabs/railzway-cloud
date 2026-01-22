import { useEffect, useMemo, useState, type ReactNode } from 'react';
import axios from 'axios';
import Login from '../pages/Login';

type AuthStatus = 'checking' | 'authenticated' | 'unauthenticated' | 'error';

type AuthGateProps = {
  children: ReactNode;
};

export default function AuthGate({ children }: AuthGateProps) {
  const [status, setStatus] = useState<AuthStatus>('checking');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const retry = useMemo(
    () => async () => {
      setStatus('checking');
      setErrorMessage(null);
      try {
        await axios.get('/auth/session');
        setStatus('authenticated');
      } catch (error: any) {
        const code = error.response?.status;
        if (code === 401) {
          setStatus('unauthenticated');
          return;
        }
        setStatus('error');
        setErrorMessage(error?.message || 'Unable to verify session.');
      }
    },
    []
  );

  useEffect(() => {
    let isMounted = true;

    const checkSession = async () => {
      try {
        await axios.get('/auth/session');
        if (!isMounted) return;
        setStatus('authenticated');
      } catch (error: any) {
        if (!isMounted) return;
        const code = error.response?.status;
        if (code === 401) {
          setStatus('unauthenticated');
          return;
        }
        setStatus('error');
        setErrorMessage(error?.message || 'Unable to verify session.');
      }
    };

    const interceptor = axios.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error?.response?.status === 401) {
          setStatus('unauthenticated');
        }
        return Promise.reject(error);
      }
    );

    checkSession();

    return () => {
      isMounted = false;
      axios.interceptors.response.eject(interceptor);
    };
  }, []);

  if (status === 'checking') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-bg-primary text-text-primary">
        <div className="flex items-center gap-3 text-sm font-mono text-text-muted">
          <div className="w-2 h-2 bg-accent-primary rounded-full animate-pulse" />
          <span>VERIFYING_SESSION</span>
        </div>
      </div>
    );
  }

  if (status === 'unauthenticated') {
    return <Login />;
  }

  if (status === 'error') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-bg-primary text-text-primary px-6 py-12">
        <div className="surface-card w-full max-w-md p-8 text-center space-y-4">
          <h1 className="text-xl font-semibold">Unable to verify session</h1>
          <p className="text-sm text-text-muted">{errorMessage || 'Please try again.'}</p>
          <button
            type="button"
            onClick={retry}
            className="w-full rounded-lg bg-accent-primary px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-accent-primary/90"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
