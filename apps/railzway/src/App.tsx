import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Onboarding from './pages/Onboarding';
import Dashboard from './pages/Dashboard';
import Organizations from './pages/Organizations';
import { Layout } from './components/Layout';

import RootRedirect from './components/RootRedirect';
import AuthGate from './components/AuthGate';

function App() {
  return (
    <AuthGate>
      <Router>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/onboarding" element={<Onboarding />} />
            <Route path="/orgs" element={<Organizations />} />
            <Route path="/orgs/:slug" element={<Dashboard />} />
            <Route path="/" element={<RootRedirect />} />
          </Route>

          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </AuthGate>
  );
}

export default App;
