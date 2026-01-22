import { useEffect, useMemo } from 'react';
import { Activity, Server, ExternalLink, Zap, Play, Square, Pause, CheckCircle, AlertTriangle } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';
import clsx from 'clsx';
import { motion, useReducedMotion } from 'motion/react';
import { useQueryClient } from '@tanstack/react-query';
import { getPageVariants, getStaggerVariants } from '../lib/motion';
import { useOrganizations, useInstanceStatus, useInstanceAction } from '../lib/api';
import type { InstanceStatus } from '../lib/types';

const TIER_CONFIG: Record<string, { name: string; color: string; price: string; features: { label: string; sub: string }[] }> = {
  'FREE_TRIAL': {
    name: 'Free Trial',
    color: 'bg-slate-600',
    price: '$0',
    features: [
      { label: '3 Customers', sub: '3 Subscriptions' },
      { label: '10k Events', sub: '14 days retention' },
      { label: 'Community Support', sub: 'Best effort' },
      { label: 'Shared Infra', sub: 'Standard performance' },
    ]
  },
  'STARTER': {
    name: 'Starter',
    color: 'bg-blue-600',
    price: '$19',
    features: [
      { label: '100 Customers', sub: '200 Subscriptions' },
      { label: '10k Events', sub: '30 days retention' },
      { label: 'Email Support', sub: '24h response' },
      { label: 'Shared Infra', sub: 'Standard performance' },
    ]
  },
  'PRO': {
    name: 'Pro',
    color: 'bg-indigo-600',
    price: '$39',
    features: [
      { label: '1k Customers', sub: '2k Subscriptions' },
      { label: '50k Events', sub: '90 days retention' },
      { label: 'Priority Support', sub: '12h response' },
      { label: 'Dedicated Namespace', sub: 'High performance' },
    ]
  },
  'TEAM': {
    name: 'Team',
    color: 'bg-purple-600',
    price: '$99',
    features: [
      { label: '10k Customers', sub: '25k Subscriptions' },
      { label: '200k Events', sub: '1 year retention' },
      { label: 'Chat Support', sub: '4h response' },
      { label: 'Isolated Infra', sub: 'Guaranteed resources' },
    ]
  },
  'ENTERPRISE': {
    name: 'Enterprise',
    color: 'bg-black',
    price: 'Custom',
    features: [
      { label: 'Unlimited', sub: 'Custom quotas' },
      { label: 'Unlimited', sub: 'Unlimited retention' },
      { label: 'Dedicated Support', sub: '1h response SLA' },
      { label: 'Dedicated Infra', sub: 'Physically isolated' },
    ]
  },
};

export default function Dashboard() {
  const { slug } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const reduceMotion = useReducedMotion();

  const { data: orgs = [], isLoading: orgsLoading } = useOrganizations();
  const { mutate: performAction, isPending: actionLoading } = useInstanceAction();

  const currentOrg = useMemo(() => {
    if (orgsLoading || orgs.length === 0) return null;

    // Redirect logic
    if (!slug) {
      navigate(`/orgs/${orgs[0].slug}`, { replace: true });
      return null;
    }

    const matched = orgs.find(o => o.slug === slug);
    if (!matched) {
      navigate(`/orgs/${orgs[0].slug}`, { replace: true });
      return null;
    }

    return matched;
  }, [orgs, slug, orgsLoading, navigate]);

  const { data: status, isLoading: statusLoading } = useInstanceStatus(currentOrg?.id);

  const pageVariants = getPageVariants(reduceMotion);
  const { container: staggerContainer, item: staggerItem } = getStaggerVariants(reduceMotion);

  // SSE Stream Integration
  useEffect(() => {
    if (!currentOrg) return;

    let isActive = true;
    let source: EventSource | null = null;

    source = new EventSource(`/user/instance/stream?org_id=${currentOrg.id}`, {
      withCredentials: true,
    });

    source.onmessage = (event) => {
      if (!event.data || !isActive) return;
      try {
        const nextStatus = JSON.parse(event.data) as InstanceStatus;
        // Update Query Cache directly from SSE
        queryClient.setQueryData(['instance-status', currentOrg.id], nextStatus);
      } catch (err) {
        console.error('Failed to parse SSE message', err);
      }
    };

    source.onerror = () => {
      if (!isActive) return;
      // SSE error usually means connection lost, Query Query will fall back to polling if we set it up,
      // but here we just wait for standard query refetch or reconnection.
    };

    return () => {
      isActive = false;
      if (source) source.close();
    };
  }, [currentOrg?.id, queryClient]);

  const handleAction = (action: string, payload?: any) => {
    if (!currentOrg) return;
    performAction({ action, orgID: currentOrg.id, payload });
  };

  const loading = orgsLoading || (currentOrg && statusLoading && !status);

  if (loading) {
    return (
      <div className="h-screen flex flex-col items-center justify-center bg-bg-primary text-text-primary font-mono text-sm">
        <div className="flex items-center gap-3">
          <div className="w-2 h-2 bg-accent-primary rounded-full animate-pulse"></div>
          <span className="flex items-center gap-1">
            CONNECTING_TO_CONTROL_PLANE
            <span className="inline-flex">
              <span className="animate-[pulse_1.4s_infinite]">.</span>
              <span className="animate-[pulse_1.4s_infinite_200ms]">.</span>
              <span className="animate-[pulse_1.4s_infinite_400ms]">.</span>
            </span>
          </span>
        </div>
      </div>
    );
  }

  if (!status || !currentOrg) {
    return null;
  }

  // State definitions hoisted for scope visibility
  const isRunning = status.status === 'running' || status.status === 'complete' || status.status === 'active';
  const isStopped = status.status === 'stopped';
  const isProvisioning = status.status === 'init' || status.status === 'provisioning' || status.status === 'pending' || status.status === 'queued';
  const isFailed = status.status === 'provision_failed';
  const isTransitioning = isProvisioning || status.status === 'upgrading' || actionLoading;

  // Handling "Missing" status (404) -> Empty State
  if (status.status === 'missing') {
    return (
      <div className="min-h-[70vh] flex flex-col items-center justify-center p-4">
        <div className="surface-card w-full max-w-md p-8 text-center space-y-6 animate-slide-up">
          <div className="w-16 h-16 rounded-full border border-dashed border-border-strong bg-bg-surface-strong/50 flex items-center justify-center mx-auto">
            <Server className="w-8 h-8 text-text-muted" />
          </div>

          <div className="space-y-2">
            <h2 className="text-2xl font-bold text-text-primary">No instance found</h2>
            <p className="text-text-secondary text-sm leading-relaxed">
              This organization does not have an active billing instance yet.
            </p>
          </div>

          <button
            onClick={() => handleAction('deploy', { version: 'latest' })}
            disabled={!!actionLoading}
            className="w-full rounded-lg bg-accent-primary hover:bg-accent-primary/90 text-white font-semibold py-3 px-4 transition-all shadow-sm text-sm flex items-center justify-center gap-2"
          >
            {actionLoading ? (
              <>
                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Initializing...
              </>
            ) : (
              <>
                <Zap className="w-4 h-4" />
                Provision Instance
              </>
            )}
          </button>
        </div>
      </div>
    );
  }

  // Handling "Provisioning" status -> Loading Overlay
  if (isProvisioning) {
    return (
      <div className="min-h-[70vh] flex flex-col items-center justify-center p-4">
        <div className="surface-card w-full max-w-md p-8 text-center space-y-8 animate-slide-up relative overflow-hidden">
          {/* Progress Bar */}
          <div className="absolute top-0 left-0 w-full h-1 bg-bg-surface-strong">
            <div className="h-full bg-accent-primary animate-[loading_2s_ease-in-out_infinite]" />
          </div>

          <div className="w-20 h-20 rounded-full border border-accent-primary/20 bg-accent-primary/5 flex items-center justify-center mx-auto relative">
            <div className="absolute inset-0 rounded-full border-2 border-accent-primary/30 border-t-transparent animate-spin" />
            <Activity className="w-8 h-8 text-accent-primary" />
          </div>

          <div className="space-y-4">
            <div>
              <p className="text-xs font-mono uppercase tracking-[0.2em] text-accent-primary font-bold mb-2">Deploying</p>
              <h2 className="text-2xl font-bold text-text-primary">
                Setting up environment
              </h2>
            </div>

            <div className="space-y-3 pt-2">
              <div className="flex items-center gap-3 text-sm text-text-secondary">
                <CheckCircle className="w-4 h-4 text-status-success" />
                <span>Validating billing metrics</span>
              </div>
              <div className="flex items-center gap-3 text-sm text-text-secondary">
                <div className="w-4 h-4 flex items-center justify-center">
                  <div className="w-2 h-2 bg-text-muted rounded-full animate-pulse" />
                </div>
                <span>Allocating control plane</span>
              </div>
              <div className="flex items-center gap-3 text-sm text-text-muted/50">
                <div className="w-4 h-4" />
                <span>Configuring ingress routes</span>
              </div>
            </div>

            <p className="text-xs text-text-muted pt-4 border-t border-border-subtle">
              This usually takes less than a minute. You can leave this page.
            </p>
          </div>
        </div>
      </div>
    );
  }

  const currentTier = TIER_CONFIG[status.tier as keyof typeof TIER_CONFIG] || {
    name: status.plan_id || status.tier, // Fallback to plan_id if available, else tier
    color: 'bg-gray-600',
    price: 'Custom',
    features: []
  };
  const launchUrl = status.launch_url?.trim()
    || import.meta.env.VITE_LAUNCH_URL
    || "http://localhost:8080/login/railzway_com";

  return (
    <motion.div
      className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-10 lg:py-12"
      variants={pageVariants}
      initial="hidden"
      animate="show"
    >
      <motion.div className="space-y-10" variants={staggerContainer}>
        <motion.header className="space-y-6" variants={staggerItem}>
          <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex items-start gap-4">
              <div className="h-12 w-12 rounded-2xl border border-accent-primary/20 bg-accent-primary/10 text-accent-primary shadow-sm flex items-center justify-center">
                <Zap className="w-6 h-6" />
              </div>
              <div>
                <div className="flex flex-wrap items-center gap-2 text-xs font-mono uppercase tracking-[0.24em] text-text-muted">
                  <span>Railzway Cloud</span>
                  <span className="h-1 w-1 rounded-full bg-border-strong" />
                  <span>{currentOrg.slug}</span>
                </div>
                <h1 className="text-3xl font-bold text-text-primary leading-tight mt-2">
                  {currentOrg.name}
                  <span className="block text-base font-medium text-text-secondary mt-1">Plan and instance control center</span>
                </h1>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-3">
              <a
                href={launchUrl}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 rounded-lg bg-accent-primary px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-accent-primary/90"
              >
                <ExternalLink className="h-4 w-4" />
                Launch
              </a>
              <div
                className={clsx(
                  "flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-mono font-bold",
                  isRunning
                    ? "bg-status-success/10 border-status-success/20 text-status-success"
                    : isFailed || isStopped
                      ? "bg-status-error/10 border-status-error/20 text-status-error"
                      : "bg-status-warning/10 border-status-warning/20 text-status-warning animate-pulse"
                )}
              >
                <div
                  className={clsx(
                    "h-1.5 w-1.5 rounded-full",
                    isRunning ? "bg-status-success" : isFailed || isStopped ? "bg-status-error" : "bg-status-warning"
                  )}
                />
                {isTransitioning ? "UPDATING" : status.status.toUpperCase()}
              </div>
              <div className="rounded-full border border-border-subtle bg-bg-surface px-3 py-1 text-xs font-mono text-text-muted">
                Version <span className="text-text-primary">{status.version}</span>
              </div>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-3 text-xs text-text-muted">
            <span className="rounded-full border border-border-subtle bg-bg-surface px-3 py-1 font-mono">ORG {currentOrg.id}</span>
            <span className="rounded-full border border-border-subtle bg-bg-surface px-3 py-1 font-mono">Tier: {status.tier}</span>
          </div>
        </motion.header>

        {status.last_error && (
          <motion.div
            className="rounded-2xl border border-red-200/60 bg-red-50/70 p-4 text-sm text-red-700 flex items-start gap-3"
            variants={staggerItem}
          >
            <AlertTriangle className="h-5 w-5 mt-0.5" />
            <div>
              <div className="font-semibold">Provisioning failed</div>
              <div className="mt-1 font-mono text-xs">{status.last_error}</div>
            </div>
          </motion.div>
        )}

        <motion.div className="grid gap-8 lg:grid-cols-3" variants={staggerItem}>
          <div className="lg:col-span-2 space-y-8">
            <section className="surface-card relative overflow-hidden p-6 space-y-6">
              <div className="absolute -right-10 -top-10 h-28 w-28 rounded-full bg-[radial-gradient(circle_at_center,hsl(var(--accent-glow)/0.25),transparent_70%)] blur-2xl" />
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <p className="text-xs font-mono uppercase tracking-[0.2em] text-text-muted">Subscription</p>
                  <h3 className="text-xl font-bold text-text-primary mt-2">Your current plan</h3>
                  <p className="text-sm text-text-secondary mt-1">
                    Billing method: <span className="font-semibold text-text-primary">Railzway OSS Billing</span>
                  </p>
                </div>
                <div className={clsx("px-3 py-1 rounded-full text-white text-xs font-bold uppercase shadow-sm", currentTier.color)}>
                  {currentTier.name}
                </div>
              </div>

              <div className="rounded-xl border border-border-subtle bg-bg-surface-strong/60 p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                  <div className="text-xs text-text-muted uppercase font-mono mb-1">Monthly Cost</div>
                  <div className="text-3xl font-bold text-text-primary">
                    {currentTier.price}
                    <span className="text-base font-normal text-text-muted">/mo</span>
                  </div>
                </div>
                <button className="text-sm font-semibold text-accent-primary hover:text-accent-primary/80 transition-colors">
                  Upgrade plan or view usage limits
                </button>
              </div>

              <div className="grid sm:grid-cols-2 gap-4 pt-4 border-t border-border-subtle">
                {currentTier.features?.map((feature, i) => (
                  <div key={i} className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-status-success mt-0.5" />
                    <div>
                      <p className="text-sm font-medium text-text-primary">{feature.label}</p>
                      <p className="text-xs text-text-muted">{feature.sub}</p>
                    </div>
                  </div>
                ))}
              </div>
            </section>

            <section className="surface-card p-6 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-bold text-text-primary">Usage this month</h3>
                <span className="text-xs font-mono text-text-muted">Live feed</span>
              </div>
              <div className="h-52 rounded-xl border border-border-subtle bg-[linear-gradient(135deg,hsl(var(--bg-surface))_0%,hsl(var(--bg-surface-strong))_100%)] p-4 flex flex-col justify-between">
                <div className="text-xs font-mono text-text-muted">Realtime usage summary coming soon</div>
                <div className="grid grid-cols-12 items-end gap-2">
                  {[38, 22, 48, 30, 54, 36, 62, 28, 40, 50, 26, 58].map((height, index) => (
                    <div
                      key={`usage-bar-${index}`}
                      className="rounded-full bg-accent-primary/30"
                      style={{ height: `${height}%` }}
                    />
                  ))}
                </div>
              </div>
            </section>
          </div>

          <div className="space-y-6">
            <div className="surface-card p-6 space-y-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Server className="w-5 h-5 text-text-secondary" />
                  <h3 className="text-md font-bold text-text-primary">Instance Control</h3>
                </div>
                <span className="text-xs font-mono text-text-muted">Actions</span>
              </div>

              <div className="rounded-xl border border-border-subtle bg-bg-surface-strong/70 p-4 space-y-3">
                <div className="flex justify-between items-center">
                  <span className="text-xs font-mono text-text-muted">STATUS</span>
                  <span className={clsx("text-xs font-bold", isRunning ? "text-status-success" : "text-status-error")}>
                    {isRunning ? "HEALTHY" : isStopped ? "STOPPED" : "DOWN"}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-xs font-mono text-text-muted">VERSION</span>
                  <span className="text-xs font-mono text-text-primary">{status.version}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-xs font-mono text-text-muted">URL</span>
                  <span className="text-xs font-mono text-text-primary truncate max-w-[160px]">
                    {status.launch_url || status.url || "Not available"}
                  </span>
                </div>
              </div>

              {isRunning ? (
                <div className="space-y-3">
                  <button
                    disabled={isTransitioning}
                    onClick={() => handleAction('pause')}
                    className="w-full flex items-center justify-center gap-2 rounded-lg bg-accent-primary hover:bg-accent-primary/90 text-white font-semibold py-2.5 px-4 transition-all shadow-sm text-sm disabled:opacity-60"
                  >
                    <Pause className="w-4 h-4" />
                    {actionLoading ? "Pausing..." : "Pause Instance"}
                  </button>
                  <button
                    disabled={isTransitioning}
                    onClick={() => handleAction('stop')}
                    className="w-full flex items-center justify-center gap-2 rounded-lg border border-red-200 bg-white hover:bg-red-50 text-red-600 font-semibold py-2.5 px-4 transition-all text-sm disabled:opacity-60"
                  >
                    <Square className="w-4 h-4 fill-current" />
                    {actionLoading ? "Stopping..." : "Stop Instance"}
                  </button>
                </div>
              ) : (
                <button
                  disabled={isTransitioning}
                  onClick={() => handleAction('start')}
                  className="w-full flex items-center justify-center gap-2 rounded-lg bg-status-success hover:bg-status-success/90 text-white font-semibold py-2.5 px-4 transition-all shadow-sm text-sm disabled:opacity-60"
                >
                  <Play className="w-4 h-4 fill-current" />
                  {actionLoading ? "Starting..." : "Start Instance"}
                </button>
              )}
            </div>

            <div className="rounded-2xl border border-red-100 bg-red-50/60 p-6 space-y-3">
              <div className="flex items-center gap-2 text-red-600">
                <AlertTriangle className="w-4 h-4" />
                <h3 className="text-sm font-bold">Danger Zone</h3>
              </div>
              <p className="text-xs text-text-secondary">
                To delete this organization and cancel your subscription, please contact support.
              </p>
            </div>
          </div>
        </motion.div>
      </motion.div>
    </motion.div>
  );
}
