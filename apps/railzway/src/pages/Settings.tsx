import { useEffect, useMemo, useState } from 'react';
import { motion, useReducedMotion } from 'motion/react';
import { getPageVariants, getStaggerVariants } from '../lib/motion';
import { useUpdateUserProfile, useUserProfile } from '../lib/api';

export default function Settings() {
  const reduceMotion = useReducedMotion();
  const pageVariants = getPageVariants(reduceMotion);
  const { container: staggerContainer, item: staggerItem } = getStaggerVariants(reduceMotion);
  const { data: profile, isLoading, isError } = useUserProfile();
  const updateProfile = useUpdateUserProfile();
  const isSaving = updateProfile.isPending;
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');

  useEffect(() => {
    if (!profile) return;
    setFirstName(profile.first_name ?? '');
    setLastName(profile.last_name ?? '');
  }, [profile]);

  const isDirty = useMemo(() => {
    return firstName !== (profile?.first_name ?? '') || lastName !== (profile?.last_name ?? '');
  }, [firstName, lastName, profile]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    updateProfile.mutate({ first_name: firstName, last_name: lastName });
  };

  return (
    <motion.div
      className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-10 lg:px-10"
      variants={pageVariants}
      initial="hidden"
      animate="show"
    >
      <motion.div className="space-y-8" variants={staggerContainer}>
        <motion.header className="space-y-3" variants={staggerItem}>
          <p className="text-xs font-mono uppercase tracking-[0.3em] text-text-muted">Settings</p>
          <h1 className="text-2xl font-bold text-text-primary sm:text-3xl">Profile information</h1>
          <p className="text-sm text-text-secondary">Update the name shown across your workspace.</p>
        </motion.header>

        {isLoading && !profile && (
          <motion.div className="surface-card flex items-center gap-3 p-6 text-text-muted" variants={staggerItem}>
            <span className="text-sm">Loading profile...</span>
          </motion.div>
        )}

        {isError && !profile && (
          <motion.div className="surface-card p-6 text-sm text-text-muted" variants={staggerItem}>
            Unable to load your profile right now. Please try again.
          </motion.div>
        )}

        {profile && (
          <motion.form onSubmit={handleSubmit} className="space-y-6" variants={staggerItem}>
            <div className="surface-card overflow-hidden">
              <div className="flex flex-col gap-2 border-b border-border-subtle px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
                <label htmlFor="first-name" className="text-sm font-semibold text-text-primary">
                  First name
                </label>
                <input
                  id="first-name"
                  type="text"
                  value={firstName}
                  onChange={(event) => setFirstName(event.target.value)}
                  className="w-full rounded-xl border border-border-strong bg-bg-surface px-4 py-2 text-sm text-text-primary shadow-sm transition focus:border-accent-primary focus:outline-none sm:w-72"
                  placeholder="First name"
                  disabled={isSaving}
                />
              </div>
              <div className="flex flex-col gap-2 border-b border-border-subtle px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
                <label htmlFor="last-name" className="text-sm font-semibold text-text-primary">
                  Last name
                </label>
                <input
                  id="last-name"
                  type="text"
                  value={lastName}
                  onChange={(event) => setLastName(event.target.value)}
                  className="w-full rounded-xl border border-border-strong bg-bg-surface px-4 py-2 text-sm text-text-primary shadow-sm transition focus:border-accent-primary focus:outline-none sm:w-72"
                  placeholder="Last name"
                  disabled={isSaving}
                />
              </div>
              <div className="flex items-center justify-end px-6 py-4">
                <button
                  type="submit"
                  className="rounded-lg bg-accent-primary px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-accent-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={!isDirty || isSaving}
                >
                  {isSaving ? 'Saving...' : 'Save'}
                </button>
              </div>
            </div>
          </motion.form>
        )}
      </motion.div>
    </motion.div>
  );
}
