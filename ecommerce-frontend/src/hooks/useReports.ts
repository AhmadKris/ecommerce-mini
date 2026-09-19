import { useQuery } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { DashboardReport } from "../types/report";

/** Fetches the admin dashboard's business metrics (admin, needs
 * report:read) for the last `days` days, compared against the equal-length
 * period before that. */
export function useDashboardReport(days = 30) {
  return useQuery({
    queryKey: ["admin", "reports", "dashboard", { days }],
    queryFn: () =>
      apiClient
        .get<DashboardReport>("/admin/reports/dashboard", { params: { days } })
        .then((response) => response.data),
  });
}
