import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import type { Organization, InstanceStatus, Price, PriceAmount, UserProfile } from './types';
import { toast } from 'sonner';

export const useOrganizations = () => {
  return useQuery<Organization[]>({
    queryKey: ['organizations'],
    queryFn: async () => {
      const res = await axios.get('/user/organizations');
      return res.data.data || [];
    },
  });
};

export const useUserProfile = () => {
  return useQuery<UserProfile>({
    queryKey: ['user-profile'],
    queryFn: async () => {
      const res = await axios.get('/user/profile');
      return res.data.data;
    },
  });
};

export const useUpdateUserProfile = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ first_name, last_name }: { first_name: string; last_name: string }) => {
      const res = await axios.put('/user/profile', { first_name, last_name });
      return res.data.data as UserProfile;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['user-profile'], data);
      toast.success('Profile updated');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to update profile');
    },
  });
};

export const usePrices = () => {
  return useQuery<Price[]>({
    queryKey: ['prices'],
    queryFn: async () => {
      // Direct call to Billing API (proxied via Traefik or similar in real setup)
      // For now we assume the frontend can hit /api/prices
      // NOTE: Adjust base URL if needed. Assuming relative path works via proxy.
      const res = await axios.get('/api/prices');
      return res.data.data || [];
    },
    staleTime: 300000, // 5 minutes
  });
};

export const usePriceAmounts = () => {
  return useQuery<PriceAmount[]>({
    queryKey: ['price-amounts'],
    queryFn: async () => {
      const res = await axios.get('/api/price_amounts');
      return res.data.data || [];
    },
    staleTime: 300000,
  });
};

export const useInstanceStatus = (orgID?: number) => {
  return useQuery<InstanceStatus>({
    queryKey: ['instance-status', orgID],
    queryFn: async () => {
      if (!orgID) throw new Error('Organization ID is required');
      try {
        const res = await axios.get('/user/instance', {
          params: { org_id: orgID }
        });
        return res.data;
      } catch (error: any) {
        if (error.response?.status === 404) {
          return { ...error.response?.data, status: 'missing' };
        }
        throw error;
      }
    },
    enabled: !!orgID,
    staleTime: 30000, // 30 seconds
  });
};

export const useInstanceAction = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ action, orgID, payload }: { action: string; orgID: number; payload?: any }) => {
      await axios.post(`/user/instance/${action}?org_id=${orgID}`, payload);
      return { action, orgID };
    },
    onSuccess: (data) => {
      toast.success(`Instance ${data.action} initiated successfully`);
      queryClient.invalidateQueries({ queryKey: ['instance-status', data.orgID] });
    },
    onError: (error: any, variables) => {
      console.error(`Action ${variables.action} failed`, error, variables);
      toast.error(`Failed to ${variables.action} instance: ${error.response?.data?.message || 'Infrastructure error'}`);
    }
  });
};

export const useCheckOrgName = (namespace: string, rootDomain?: string) => {
  const normalized = namespace.trim().toLowerCase();
  const namespaceValue = normalized
    ? (rootDomain ? `${normalized}.${rootDomain}` : normalized)
    : '';
  return useQuery({
    queryKey: ['check-org-name', namespaceValue],
    queryFn: async () => {
      if (normalized.length < 3) return null;
      const res = await axios.get('/user/onboarding/check-org-name', {
        params: { namespace: namespaceValue }
      });
      return res.data.available as boolean;
    },
    enabled: normalized.length >= 3,
    staleTime: 60000,
  });
};

export const useInitializeOrg = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      planID,
      priceID,
      orgName,
      orgNamespace
    }: {
      planID: string;
      priceID: string;
      orgName: string;
      orgNamespace: string;
    }) => {
      const res = await axios.post('/user/onboarding/initialize', {
        plan_id: planID,
        price_id: priceID,
        org_name: orgName,
        org_namespace: orgNamespace
      });
      return res.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
    }
  });
};
