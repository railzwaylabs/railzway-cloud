import { useState, useMemo, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, ArrowRight, Terminal, Cpu, HardDrive, Database, Shield, AlertCircle, CheckCircle } from 'lucide-react';
import clsx from 'clsx';
import { motion, useReducedMotion } from 'motion/react';
import { toast } from 'sonner';
import { getPageVariants, getStaggerVariants } from '../lib/motion';
import { useOrganizations, useCheckOrgName, useInitializeOrg, usePrices, usePriceAmounts } from '../lib/api';



export default function Onboarding() {
  const [step, setStep] = useState(1);
  const [plan, setPlan] = useState<string>('');
  const [orgName, setOrgName] = useState('');
  const navigate = useNavigate();

  const { data: orgs = [] } = useOrganizations();
  const { data: isAvailable, isFetching: isChecking } = useCheckOrgName(orgName);
  const { mutate: initialize, isPending: loading } = useInitializeOrg();

  // Pricing API Integration
  const { data: prices = [], isLoading: loadingPrices } = usePrices();
  const { data: priceAmounts = [], isLoading: loadingAmounts } = usePriceAmounts();

  const instanceProfiles = useMemo(() => {
    if (!prices.length) return [];

    // Transform API data to UI model
    // Transform API data to UI model
    const profiles = prices
      .filter(p => p.active)
      .map(price => {
        const amountObj = priceAmounts.find(pa => pa.price_id === price.id);
        const amount = amountObj ? Math.floor(amountObj.unit_amount_cents / 100) : 0;
        const meta = price.metadata || {};

        return {
          id: price.id,
          name: price.name, // "Evaluation", "Hobby", etc.
          badge: meta.badge || 'TIER',
          badgeColor: meta.badge_color || 'bg-slate-600',
          price: `$${amount}`,
          priceUnit: '/mo',
          type: meta.type || 'Instance',
          specs: meta.specs || { cpu: '-', ram: '-', storage: '-', isolation: '-' },
          desc: meta.description || '',
          warning: meta.warning || null,
          highlight: meta.highlight || null,
          // Used for sorting if needed, logic for 'free' check
          amountVal: amount
        };
      })
      .sort((a, b) => a.amountVal - b.amountVal);

    return profiles;
  }, [prices, priceAmounts]);

  // Set default plan selection once data is loaded
  useEffect(() => {
    if (!plan && instanceProfiles.length > 0) {
      setPlan(instanceProfiles[0].id);
    }
  }, [instanceProfiles, plan]);

  const hasExistingOrgs = orgs.length > 0;
  const selectedPlan = instanceProfiles.find((profile) => profile.id === plan) ?? instanceProfiles[0];
  const reduceMotion = useReducedMotion();
  const pageVariants = getPageVariants(reduceMotion);
  const { container: staggerContainer, item: staggerItem } = getStaggerVariants(reduceMotion);

  const errorMessage = useMemo(() => {
    if (orgName.length > 0 && orgName.length < 3) return 'Namespace must be at least 3 characters';
    if (isAvailable === false) return 'This namespace is already taken';
    return '';
  }, [orgName, isAvailable]);

  const handleInitialize = () => {
    const planID = selectedPlan?.name ?? '';
    const priceID = plan;
    initialize({ planID, priceID, orgName }, {
      onSuccess: (data) => {
        toast.success("Instance provisioning initiated");
        const org = Array.isArray(data) ? data[0] : data;
        const slug = org?.slug || org?.Slug;
        if (slug) {
          navigate(`/orgs/${slug}`);
        } else {
          navigate('/');
        }
      },
      onError: (err: any) => {
        console.error(err);
        toast.error("Provisioning failed due to an infrastructure error.");
      }
    });
  };

  return (
    <motion.div
      className="min-h-screen px-6 py-12 lg:py-16"
      variants={pageVariants}
      initial="hidden"
      animate="show"
    >
      <motion.div className="mx-auto w-full max-w-6xl space-y-10" variants={staggerContainer}>
        <motion.header className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between" variants={staggerItem}>
          <div className="space-y-3">
            {hasExistingOrgs && (
              <Link
                to="/orgs"
                className="inline-flex items-center gap-2 rounded-full border border-border-subtle px-3 py-1 text-xs font-mono text-text-muted transition-colors hover:border-border-strong hover:text-text-primary"
              >
                <ArrowLeft className="h-3.5 w-3.5" />
                Back to organizations
              </Link>
            )}
            <p className="text-xs font-mono uppercase tracking-[0.32em] text-text-muted">Railzway Cloud Onboarding</p>
            <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-text-primary">
              {step === 1 ? 'Claim your instance namespace' : 'Choose your infrastructure profile'}
            </h1>
            <p className="text-text-secondary max-w-2xl text-base md:text-lg leading-relaxed">
              {step === 1
                ? 'Choose a unique subdomain for your Railzway instance.'
                : 'Pick the isolation profile that matches your workload. You can upgrade at any time as usage grows.'}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3 text-xs font-mono uppercase tracking-[0.2em]">
            <span className={clsx(
              "rounded-full border px-3 py-1 transition-colors",
              step === 1 ? "border-accent-primary bg-accent-primary text-white" : "border-border-subtle text-text-muted"
            )}>
              1 Namespace
            </span>
            <span className={clsx(
              "rounded-full border px-3 py-1 transition-colors",
              step === 2 ? "border-accent-primary bg-accent-primary text-white" : "border-border-subtle text-text-muted"
            )}>
              2 Provision
            </span>
          </div>
        </motion.header>

        <motion.div className="grid gap-8 lg:grid-cols-[0.85fr,1.15fr]" variants={staggerItem}>
          <motion.aside className="surface-card p-6 md:p-8 space-y-6" variants={staggerItem}>
            <div className="flex items-center justify-between">
              <p className="text-xs font-mono uppercase tracking-[0.2em] text-text-muted">Blueprint</p>
              <span className="rounded-full border border-border-subtle bg-bg-surface px-3 py-1 text-xs font-mono text-text-muted">
                Step {step}/2
              </span>
            </div>
            <h2 className="text-2xl font-bold text-text-primary">What you are provisioning</h2>
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <div className="h-9 w-9 rounded-xl bg-accent-primary/10 border border-accent-primary/20 flex items-center justify-center text-accent-primary">
                  <Shield className="w-4 h-4" />
                </div>
                <div>
                  <p className="text-sm font-semibold text-text-primary">Single-tenant security</p>
                  <p className="text-xs text-text-muted">Isolated compute and billing boundaries per org.</p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="h-9 w-9 rounded-xl bg-accent-primary/10 border border-accent-primary/20 flex items-center justify-center text-accent-primary">
                  <Database className="w-4 h-4" />
                </div>
                <div>
                  <p className="text-sm font-semibold text-text-primary">Metrics and logs</p>
                  <p className="text-xs text-text-muted">Usage insights with automated billing guardrails.</p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="h-9 w-9 rounded-xl bg-accent-primary/10 border border-accent-primary/20 flex items-center justify-center text-accent-primary">
                  <Terminal className="w-4 h-4" />
                </div>
                <div>
                  <p className="text-sm font-semibold text-text-primary">Managed upgrades</p>
                  <p className="text-xs text-text-muted">Continuous updates with rollback protection.</p>
                </div>
              </div>
            </div>

            {selectedPlan ? (
              <div className="rounded-xl border border-border-subtle bg-bg-surface-strong/70 p-4 space-y-1">
                <p className="text-xs font-mono uppercase tracking-[0.2em] text-text-muted">Selected plan</p>
                <p className="text-lg font-bold text-text-primary">{selectedPlan.name}</p>
                <p className="text-sm text-text-secondary">{selectedPlan.price}{selectedPlan.priceUnit}</p>
              </div>
            ) : (
              <div className="rounded-xl border border-border-subtle bg-bg-surface-strong/70 p-4 space-y-1 animate-pulse">
                <p className="text-xs font-mono uppercase tracking-[0.2em] text-text-muted">Loading plan...</p>
                <div className="h-6 w-3/4 bg-border-subtle rounded mt-2"></div>
                <div className="h-4 w-1/2 bg-border-subtle rounded mt-1"></div>
              </div>
            )}
          </motion.aside>

          <motion.div className="surface-card p-6 md:p-8" variants={staggerItem}>
            {step === 1 && (
              <div className="space-y-8 animate-slide-up">
                <div className="space-y-3">
                  <label className="block text-xs font-mono font-medium text-text-muted uppercase tracking-[0.2em]">
                    Instance Namespace
                  </label>
                  <div className="flex items-center">
                    <input
                      type="text"
                      value={orgName}
                      onChange={(e) => {
                        const sanitized = e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '');
                        setOrgName(sanitized);
                      }}
                      className={clsx(
                        "flex-1 bg-bg-surface border border-r-0 rounded-l-xl py-3 px-4 text-text-primary font-mono text-sm focus:ring-2 focus:ring-accent-primary/20 outline-none transition-all placeholder:text-text-muted/40",
                        !isAvailable && errorMessage ? "border-red-500 focus:border-red-500" :
                          isAvailable ? "border-status-success focus:border-status-success" :
                            "border-border-strong focus:border-accent-primary"
                      )}
                      placeholder="acme-corp"
                      autoFocus
                      spellCheck={false}
                    />
                    <div className={clsx(
                      "bg-bg-surface-strong/70 border border-l-0 rounded-r-xl px-3 py-3 text-text-muted font-mono text-sm select-none",
                      !isAvailable && errorMessage ? "border-red-500" :
                        isAvailable ? "border-status-success" :
                          "border-border-strong"
                    )}>
                      {import.meta.env.VITE_ROOT_DOMAIN || '.railzway.com'}
                    </div>
                  </div>
                  <div className="text-xs">
                    {isChecking ? (
                      <p className="text-text-muted flex items-center gap-1">
                        <span className="w-3 h-3 border-2 border-text-muted border-t-transparent rounded-full animate-spin"></span>
                        Checking availability...
                      </p>
                    ) : !isAvailable && errorMessage ? (
                      <p className="text-red-500 flex items-center gap-1">
                        <AlertCircle className="w-3 h-3" />
                        {errorMessage}
                      </p>
                    ) : isAvailable ? (
                      <p className="text-status-success flex items-center gap-1">
                        <CheckCircle className="w-3 h-3" />
                        Namespace available
                      </p>
                    ) : (
                      <p className="text-text-muted">
                        Lowercase letters (a-z), numbers (0-9), and hyphens only. Minimum 3 characters.
                      </p>
                    )}
                  </div>
                </div>

                <button
                  disabled={!orgName || !isAvailable || isChecking}
                  onClick={() => setStep(2)}
                  className="w-full rounded-lg bg-accent-primary hover:bg-accent-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-text-inverse font-mono text-sm font-medium py-3 flex items-center justify-center gap-2 transition-all shadow-sm"
                >
                  Continue <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            )}

            {step === 2 && (
              <div className="space-y-8 animate-slide-up">
                {loadingPrices || loadingAmounts ? (
                  <div className="flex flex-col items-center justify-center py-20 text-text-muted">
                    <span className="w-6 h-6 border-2 border-text-muted border-t-transparent rounded-full animate-spin mb-4"></span>
                    <p>Loading pricing options...</p>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {instanceProfiles.map((p) => (
                      <label
                        key={p.id}
                        className={clsx(
                          "relative rounded-2xl border p-6 cursor-pointer transition-all duration-200",
                          "hover:-translate-y-0.5 hover:shadow-md",
                          plan === p.id
                            ? "border-accent-primary shadow-glow bg-bg-surface"
                            : "border-border-subtle bg-bg-surface/70 hover:border-border-strong"
                        )}
                      >
                        <input
                          type="radio"
                          name="plan"
                          value={p.id}
                          checked={plan === p.id}
                          onChange={() => setPlan(p.id)}
                          className="sr-only"
                        />

                        <div className="absolute top-6 left-6">
                          <div className={clsx(
                            "w-5 h-5 rounded-full border-2 flex items-center justify-center transition-all",
                            plan === p.id
                              ? "border-accent-primary bg-accent-primary"
                              : "border-border-strong"
                          )}>
                            {plan === p.id && (
                              <CheckCircle className="w-3 h-3 text-text-inverse" />
                            )}
                          </div>
                        </div>

                        <div className="pl-10 space-y-4">
                          <div className="flex items-start justify-between gap-4">
                            <div>
                              <h3 className="text-lg font-bold text-text-primary">{p.name}</h3>
                              <p className="text-xs text-text-secondary mt-1 leading-tight">{p.desc}</p>
                            </div>
                            <span className={clsx("text-[10px] font-mono px-2 py-1 rounded-full text-white whitespace-nowrap", p.badgeColor)}>
                              {p.badge}
                            </span>
                          </div>

                          <div className="space-y-2 text-sm text-text-secondary">
                            <div className="flex items-center gap-2">
                              <Cpu className="w-4 h-4 text-text-muted" />
                              <span className="font-mono">{p.specs.cpu}</span>
                            </div>
                            <div className="flex items-center gap-2">
                              <HardDrive className="w-4 h-4 text-text-muted" />
                              <span className="font-mono">{p.specs.ram}</span>
                            </div>
                            <div className="flex items-center gap-2">
                              <Database className="w-4 h-4 text-text-muted" />
                              <span className="font-mono">{p.specs.storage}</span>
                            </div>
                            <div className="flex items-center gap-2">
                              <Shield className="w-4 h-4 text-text-muted" />
                              <span className="font-mono">{p.specs.isolation}</span>
                            </div>
                          </div>

                          <div className="flex items-end justify-between pt-2 border-t border-border-subtle">
                            <div>
                              <div className="text-2xl font-bold text-text-primary">
                                {p.price}<span className="text-base font-normal text-text-secondary">{p.priceUnit}</span>
                              </div>
                              {p.warning && (
                                <div className="flex items-center gap-1 text-xs text-status-warning mt-1">
                                  <AlertCircle className="w-3 h-3" />
                                  {p.warning}
                                </div>
                              )}
                              {p.highlight && (
                                <div className="flex items-center gap-1 text-xs text-status-success mt-1">
                                  <CheckCircle className="w-3 h-3" />
                                  {p.highlight}
                                </div>
                              )}
                            </div>
                          </div>
                        </div>
                      </label>
                    ))}

                  </div>
                )}

                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between pt-4 border-t border-border-subtle">
                  <p className="text-xs text-text-muted">
                    Provisioning will create a dedicated Railzway instance bound to <strong>{orgName}</strong>.
                  </p>
                  <div className="flex gap-3">
                    <button
                      onClick={() => setStep(1)}
                      className="px-4 py-2.5 text-text-secondary hover:text-text-primary font-mono text-sm font-medium transition-colors"
                    >
                      Back
                    </button>
                    <button
                      disabled={!plan || loading}
                      onClick={handleInitialize}
                      className="px-6 py-2.5 rounded-lg bg-accent-primary hover:bg-accent-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-text-inverse font-mono text-sm font-medium flex items-center gap-2 transition-all shadow-sm"
                    >
                      {loading ? (
                        <span className="flex items-center gap-2">
                          <Terminal className="w-4 h-4 animate-pulse" />
                          Initializing...
                        </span>
                      ) : (
                        <>
                          <Terminal className="w-4 h-4" />
                          Provision Instance
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            )}
          </motion.div>
        </motion.div>
      </motion.div>
    </motion.div>
  );
}
