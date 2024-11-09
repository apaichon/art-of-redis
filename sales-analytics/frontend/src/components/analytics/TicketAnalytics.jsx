/* import React from 'react';
import { useWebSocket } from '../../hooks/useWebSocket';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { SummaryCards } from './SummaryCards';
import { HourlySalesChart } from './HourlySalesChart';
import { CategoryDistribution } from './CategoryDistribution';
import { TopConcerts } from './TopConcerts';

const TicketAnalytics = () => {
  const analytics = useWebSocket('ws://localhost:8080/ws');

  if (!analytics) {
    return <LoadingSpinner />;
  }

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <h1 className="text-3xl font-bold mb-6">Concert Ticket Sales Analytics</h1>
      
      <SummaryCards analytics={analytics} />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <HourlySalesChart data={analytics.revenueByHour} />
        <CategoryDistribution salesByCategory={analytics.salesByCategory} />
      </div>

      <TopConcerts concerts={analytics.topConcerts} />
    </div>
  );
};

export default TicketAnalytics;
*/

import React from 'react';
import { useWebSocket } from '../../hooks/useWebSocket';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { SummaryCards } from './SummaryCards';
import { HourlySalesChart } from './HourlySalesChart';
import { CategoryDistribution } from './CategoryDistribution';
import { TopConcerts } from './TopConcerts';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { AlertTriangle } from 'lucide-react';

const TicketAnalytics = () => {
  const { data: analytics, error, isConnecting } = useWebSocket();

  if (error) {
    return (
      <div className="p-6">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    );
  }

  if (isConnecting || !analytics) {
    return <LoadingSpinner />;
  }

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      {isConnecting && (
        <Alert className="mb-4">
          <AlertTitle>Reconnecting...</AlertTitle>
          <AlertDescription>
            Attempting to reconnect to the server...
          </AlertDescription>
        </Alert>
      )}

      <h1 className="text-3xl font-bold mb-6">Concert Ticket Sales Analytics</h1>
      
      <SummaryCards analytics={analytics} />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <HourlySalesChart data={analytics.revenueByHour} />
        <CategoryDistribution salesByCategory={analytics.salesByCategory} />
      </div>

      <TopConcerts concerts={analytics.topConcerts} />
    </div>
  );
};

export default TicketAnalytics;