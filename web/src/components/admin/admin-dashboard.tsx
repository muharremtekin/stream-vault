'use client';

import { StatsCards } from '@/components/admin/stats-cards';
import { RecentJobs } from '@/components/admin/recent-jobs';

export function AdminDashboard() {
  return (
    <div className="space-y-8">
      <StatsCards />
      <RecentJobs />
    </div>
  );
}
