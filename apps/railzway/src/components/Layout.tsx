import { Outlet } from 'react-router-dom';
import TopBar from './TopBar';

export const Layout = () => {
  return (
    <div className="app-shell font-sans text-text-primary">
      <div className="pointer-events-none absolute inset-0 -z-10 overflow-hidden">
        <div className="absolute top-0 left-0 h-full w-full bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-accent-primary/5 via-bg-primary to-bg-primary" />
        <div className="absolute -top-[40%] -right-[20%] h-[80%] w-[80%] rounded-full bg-accent-primary/5 blur-3xl opacity-50" />
        <div className="absolute top-[20%] -left-[20%] h-[60%] w-[60%] rounded-full bg-accent-glow/10 blur-3xl opacity-40" />
      </div>
      <div className="relative z-10">
        <TopBar />
        <Outlet />
      </div>
    </div>
  );
};
