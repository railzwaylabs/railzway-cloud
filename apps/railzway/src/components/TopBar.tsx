import { useEffect, useMemo, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ChevronDown, User } from 'lucide-react';
import clsx from 'clsx';
import { useOrganizations, useUserProfile } from '../lib/api';

const isOrgRoute = (path: string) => path === '/orgs' || path.startsWith('/orgs/');

const TopBar = () => {
  const location = useLocation();
  const { data: orgs = [] } = useOrganizations();
  const { data: profile } = useUserProfile();
  const [orgMenuOpen, setOrgMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);

  const activeSlug = useMemo(() => {
    const match = location.pathname.match(/^\/orgs\/([^/]+)/);
    return match ? match[1] : null;
  }, [location.pathname]);

  const activeOrg = useMemo(
    () => (activeSlug ? orgs.find((org) => org.slug === activeSlug) : null),
    [activeSlug, orgs]
  );

  useEffect(() => {
    if (!orgMenuOpen && !userMenuOpen) {
      return;
    }

    const handleClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (!target.closest('[data-topbar-menu="orgs"]')) {
        setOrgMenuOpen(false);
      }
      if (!target.closest('[data-topbar-menu="user"]')) {
        setUserMenuOpen(false);
      }
    };

    document.addEventListener('click', handleClick);
    return () => document.removeEventListener('click', handleClick);
  }, [orgMenuOpen, userMenuOpen]);

  if (location.pathname.startsWith('/onboarding')) {
    return null;
  }

  return (
    <div className="sticky top-0 z-header border-b border-border-subtle bg-bg-primary/90 backdrop-blur">
      <div className="mx-auto flex max-w-6xl flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-10">
        <div className="flex min-w-0 flex-wrap items-center gap-2">
          <Link to="/orgs" className="flex items-center gap-2 font-bold text-xl tracking-tight text-slate-900">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600 text-white shadow-sm">
              R
            </span>
            Railzway
          </Link>
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            {isOrgRoute(location.pathname) && (
              <div className="relative" data-topbar-menu="orgs">
                <button
                  type="button"
                  onClick={() => setOrgMenuOpen((prev) => !prev)}
                  className={clsx(
                    "flex min-w-0 items-center gap-2 rounded-full border px-3 py-1 text-xs font-mono text-text-secondary transition-colors",
                    orgMenuOpen ? "border-border-strong bg-bg-surface" : "border-border-subtle bg-bg-primary"
                  )}
                >
                  <span className="max-w-[52vw] truncate sm:max-w-none">
                    {activeOrg?.name || "Organizations"}
                  </span>
                  <ChevronDown className="h-3.5 w-3.5" />
                </button>
                {orgMenuOpen && (
                  <div className="absolute left-0 mt-2 w-56 rounded-xl border border-border-subtle bg-bg-surface shadow-md">
                    <div className="px-3 py-2 text-xs font-mono uppercase tracking-[0.2em] text-text-muted">
                      Organizations
                    </div>
                    <div className="max-h-64 overflow-auto">
                      {orgs.length === 0 ? (
                        <div className="px-3 py-2 text-xs text-text-muted">No organizations</div>
                      ) : (
                        orgs.map((org) => (
                          <Link
                            key={org.id}
                            to={`/orgs/${org.slug}`}
                            onClick={() => setOrgMenuOpen(false)}
                            className="flex flex-col gap-1 px-3 py-2 text-sm text-text-primary hover:bg-bg-surface-strong"
                          >
                            <span className="font-medium">{org.name}</span>
                            <span className="text-xs font-mono text-text-muted">{org.slug}</span>
                          </Link>
                        ))
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        <div className="flex w-full items-center justify-end gap-2 sm:w-auto">
          <div className="relative" data-topbar-menu="user">
            <button
              type="button"
              onClick={() => setUserMenuOpen((prev) => !prev)}
              className={clsx(
                "flex h-9 w-9 items-center justify-center rounded-full border transition-colors",
                userMenuOpen ? "border-border-strong bg-bg-surface" : "border-border-subtle bg-bg-primary"
              )}
              aria-label="Open user menu"
            >
              <User className="h-4 w-4 text-text-secondary" />
            </button>
            {userMenuOpen && (
              <div className="absolute right-0 mt-2 w-56 overflow-hidden rounded-xl border border-border-subtle bg-bg-surface shadow-md">
                <div className="px-3 py-2 text-sm font-medium text-text-primary">
                  {profile?.email || 'Account'}
                </div>
                <div className="border-t border-border-subtle">
                  <Link
                    to="/settings"
                    onClick={() => setUserMenuOpen(false)}
                    className="flex w-full items-center px-3 py-2 text-sm text-text-primary hover:bg-bg-surface-strong"
                  >
                    Settings
                  </Link>
                  <a
                    href="/auth/logout"
                    className="flex w-full items-center px-3 py-2 text-sm text-text-primary hover:bg-bg-surface-strong"
                  >
                    Logout
                  </a>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default TopBar;
