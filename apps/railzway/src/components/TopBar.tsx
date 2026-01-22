import { useEffect, useMemo, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ChevronDown, Plus } from 'lucide-react';
import clsx from 'clsx';
import primaryLogo from '../assets/primary.svg';
import { useOrganizations } from '../lib/api';

const isOrgRoute = (path: string) => path === '/orgs' || path.startsWith('/orgs/');

const TopBar = () => {
  const location = useLocation();
  const { data: orgs = [] } = useOrganizations();
  const [open, setOpen] = useState(false);

  const activeSlug = useMemo(() => {
    const match = location.pathname.match(/^\/orgs\/([^/]+)/);
    return match ? match[1] : null;
  }, [location.pathname]);

  const activeOrg = useMemo(
    () => (activeSlug ? orgs.find((org) => org.slug === activeSlug) : null),
    [activeSlug, orgs]
  );

  useEffect(() => {
    if (!open) {
      return;
    }

    const handleClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (!target.closest('[data-topbar-menu="orgs"]')) {
        setOpen(false);
      }
    };

    document.addEventListener('click', handleClick);
    return () => document.removeEventListener('click', handleClick);
  }, [open]);

  if (location.pathname.startsWith('/onboarding')) {
    return null;
  }

  return (
    <div className="sticky top-0 z-header border-b border-border-subtle bg-bg-primary/90 backdrop-blur">
      <div className="mx-auto flex max-w-6xl flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-10">
        <div className="flex min-w-0 flex-wrap items-center gap-3 sm:gap-4">
          <img src={primaryLogo} alt="Railzway" className="h-9 sm:h-10" />
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <Link to="/orgs" className="text-base font-semibold text-text-primary sm:text-lg">
              Railzway
            </Link>
            {isOrgRoute(location.pathname) && (
              <div className="relative" data-topbar-menu="orgs">
                <button
                  type="button"
                  onClick={() => setOpen((prev) => !prev)}
                  className={clsx(
                    "flex min-w-0 items-center gap-2 rounded-full border px-3 py-1 text-xs font-mono text-text-secondary transition-colors",
                    open ? "border-border-strong bg-bg-surface" : "border-border-subtle bg-bg-primary"
                  )}
                >
                  <span className="max-w-[52vw] truncate sm:max-w-none">
                    {activeOrg?.name || "Organizations"}
                  </span>
                  <ChevronDown className="h-3.5 w-3.5" />
                </button>
                {open && (
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
                            onClick={() => setOpen(false)}
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
          <Link
            to="/orgs"
            className="hidden rounded-full border border-border-subtle px-3 py-1 text-xs font-mono text-text-secondary transition-colors hover:border-border-strong hover:text-text-primary sm:inline-flex"
          >
            Organizations
          </Link>
          <Link
            to="/onboarding"
            className="inline-flex items-center gap-2 rounded-full bg-accent-primary px-3 py-1 text-xs font-semibold text-white shadow-sm transition-colors hover:bg-accent-primary/90"
          >
            <Plus className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">New org</span>
          </Link>
        </div>
      </div>
    </div>
  );
};

export default TopBar;
