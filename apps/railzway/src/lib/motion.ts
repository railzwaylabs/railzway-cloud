import type { Variants } from 'motion/react';

const easeStandard: [number, number, number, number] = [0.16, 1, 0.3, 1];

export const getPageVariants = (reducedMotion: boolean | null): Variants => ({
  hidden: { opacity: 0, y: reducedMotion ? 0 : 14 },
  show: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.32, ease: easeStandard },
  },
});

export const getStaggerVariants = (reducedMotion: boolean | null) => {
  const container: Variants = {
    hidden: { opacity: 1 },
    show: {
      opacity: 1,
      transition: reducedMotion
        ? {}
        : { staggerChildren: 0.08, delayChildren: 0.05 },
    },
  };

  const item: Variants = {
    hidden: { opacity: 0, y: reducedMotion ? 0 : 12 },
    show: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.28, ease: easeStandard },
    },
  };

  return { container, item };
};
