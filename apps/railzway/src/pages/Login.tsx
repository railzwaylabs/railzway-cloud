export default function Login() {
  const handleLogin = () => {
    window.location.href = '/auth/login';
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-primary text-text-primary px-6 py-12">
      <div className="surface-card w-full max-w-md p-8 text-center space-y-6">
        <div className="mx-auto flex items-center justify-center gap-2 font-bold text-xl tracking-tight text-slate-900">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600 text-white shadow-sm">
            R
          </span>
          Railzway
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-bold">Welcome to Railzway Cloud</h1>
          <p className="text-sm text-text-muted">
            Sign in using your Railzway account to continue.
          </p>
        </div>
        <button
          type="button"
          onClick={handleLogin}
          className="w-full rounded-lg bg-accent-primary px-4 py-3 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-accent-primary/90"
        >
          Sign in with Railzway.com
        </button>
        <p className="text-xs text-text-muted">
          This instance only accepts OAuth sign-in.
        </p>
      </div>
    </div>
  );
}
