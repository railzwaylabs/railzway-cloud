import { useEffect, useMemo, useState } from 'react';
import axios from 'axios';
import { Link } from 'react-router-dom';
import { Building2, ExternalLink, Loader2 } from 'lucide-react';
import clsx from 'clsx';
import { motion, useReducedMotion } from 'motion/react';
import { getPageVariants, getStaggerVariants } from '../lib/motion';

type Organization = {
  id: string;
  name: string;
  slug: string;
};

type StatusMap = Record<string, string>;

const statusTone = (status: string) => {
  switch (status) {
    case 'running':
    case 'active':
    case 'complete':
      return 'text-status-success';
    case 'stopped':
    case 'provision_failed':
      return 'text-status-error';
    case 'provisioning':
    case 'pending':
    case 'queued':
    case 'init':
    case 'upgrading':
      return 'text-status-warning';
    default:
      return 'text-text-muted';
  }
};

export default function Organizations() {
  const [orgs, setOrgs] = useState<Organization[]>([]);
  const [statuses, setStatuses] = useState<StatusMap>({});
  const [loading, setLoading] = useState(true);
  const reduceMotion = useReducedMotion();
  const pageVariants = getPageVariants(reduceMotion);
  const { container: staggerContainer, item: staggerItem } = getStaggerVariants(reduceMotion);

  const loadStatuses = async (organizations: Organization[]) => {
    const entries = await Promise.all(
      organizations.map(async (org) => {
        const orgID = String(org.id);
        try {
          const res = await axios.get('/user/instance', { params: { org_id: orgID } });
          return [orgID, res.data?.status || 'unknown'];
        } catch (error: any) {
          if (error.response?.status === 404) {
            return [orgID, error.response?.data?.status || 'not_deployed'];
          }
          return [orgID, 'unknown'];
        }
      })
    );

    setStatuses(Object.fromEntries(entries));
  };

  useEffect(() => {
    const loadOrganizations = async () => {
      try {
        const res = await axios.get('/user/organizations');
        const data = res.data.data || [];
        setOrgs(data);
        await loadStatuses(data);
      } catch (error: any) {
        // Auth is handled by AuthGate. 401s will flip the gate.
      } finally {
        setLoading(false);
      }
    };

    loadOrganizations();
  }, []);

  const emptyState = useMemo(() => !loading && orgs.length === 0, [loading, orgs.length]);

  return (
    <motion.div
      className="mx-auto w-full max-w-6xl px-4 py-8 sm:px-6 sm:py-10 lg:px-10"
      variants={pageVariants}
      initial="hidden"
      animate="show"
    >
      <motion.div className="space-y-8" variants={staggerContainer}>
        <motion.header className="space-y-3" variants={staggerItem}>
          <p className="text-xs font-mono uppercase tracking-[0.3em] text-text-muted">Organizations</p>
          <h1 className="text-2xl font-bold text-text-primary sm:text-3xl">Your organizations</h1>
          <p className="text-sm text-text-secondary">
            Pick an organization to manage its instance configuration.
          </p>
        </motion.header>

        {loading && (
          <motion.div className="surface-card flex items-center gap-3 p-6 text-text-muted" variants={staggerItem}>
            <Loader2 className="h-4 w-4 animate-spin" />
            <span className="text-sm">Loading organizations...</span>
          </motion.div>
        )}

        {emptyState && (
          <motion.div className="surface-card p-8 text-center space-y-4" variants={staggerItem}>
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-bg-surface-strong text-text-muted">
              <Building2 className="h-6 w-6" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-text-primary">No organizations yet</h2>
              <p className="text-sm text-text-muted">Create your first organization to get started.</p>
            </div>
            <Link
              to="/onboarding"
              className="inline-flex items-center gap-2 rounded-full bg-accent-primary px-4 py-2 text-sm font-semibold text-white shadow-sm"
            >
              Create organization
            </Link>
          </motion.div>
        )}

        {!loading && orgs.length > 0 && (
          <motion.div className="grid gap-4 md:grid-cols-2" variants={staggerItem}>
            {orgs.map((org) => {
              const status = statuses[String(org.id)] || 'unknown';
              return (
                <Link
                  key={org.id}
                  to={`/orgs/${org.slug}`}
                  className="surface-card flex flex-col gap-4 p-5 transition-all hover:border-border-strong hover:shadow-md sm:flex-row sm:items-center sm:justify-between"
                >
                  <div className="flex min-w-0 items-center gap-4">
                    <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-bg-surface-strong text-text-primary">
                      <Building2 className="h-5 w-5" />
                    </div>
                    <div className="min-w-0">
                      <div className="text-base font-semibold text-text-primary">{org.name}</div>
                      <div className="text-xs font-mono text-text-muted truncate">{org.slug}</div>
                    </div>
                  </div>
                  <div className="flex w-full items-center justify-between gap-3 text-xs sm:w-auto sm:justify-end">
                    <span className={clsx("text-xs font-mono uppercase", statusTone(status))}>
                      {status.replace(/_/g, ' ')}
                    </span>
                    <ExternalLink className="h-4 w-4 text-text-muted" />
                  </div>
                </Link>
              );
            })}
          </motion.div>
        )}
      </motion.div>
    </motion.div>
  );
}
