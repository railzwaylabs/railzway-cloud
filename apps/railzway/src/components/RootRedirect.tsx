import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

export default function RootRedirect() {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    const checkOrgs = async () => {
      try {
        const res = await axios.get('/user/organizations');
        const orgs = res.data.data || [];
        if (orgs.length > 0) {
          // If user has orgs, go to organizations list
          navigate('/orgs', { replace: true });
        } else {
          // If no orgs, go to onboarding
          navigate('/onboarding', { replace: true });
        }
      } catch (error: any) {
        if (error.response?.status !== 401) {
          // Fallback
          navigate('/onboarding', { replace: true });
        }
      } finally {
        setChecking(false);
      }
    };

    checkOrgs();
  }, [navigate]);

  if (checking) {
    return (
      <div className="h-screen flex items-center justify-center bg-bg-primary">
        {/* Simple spinner or blank */}
      </div>
    );
  }

  return null;
}
