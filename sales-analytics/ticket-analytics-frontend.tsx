import React, { useState, useEffect, useRef } from 'react';
import { Line, Bar, Pie } from 'recharts';
import { format } from 'date-fns';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { CircleDollarSign, Ticket, TrendingUp, Clock } from 'lucide-react';

const TicketAnalytics = () => {
  const [analytics, setAnalytics] = useState(null);
  const ws = useRef(null);

  useEffect(() => {
    ws.current = new WebSocket('ws://localhost:8080/ws');
    
    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setAnalytics(data);
    };

    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);

  if (!analytics) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-lg">Loading analytics...</div>
      </div>
    );
  }

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <h1 className="text-3xl font-bold mb-6">Concert Ticket Sales Analytics</h1>
      
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
            <CircleDollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${analytics.totalRevenue.toLocaleString()}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Tickets Sold</CardTitle>
            <Ticket className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {analytics.ticketsSold.toLocaleString()}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Average Ticket Price</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(analytics.totalRevenue / analytics.ticketsSold || 0).toFixed(2)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Last Hour Sales</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {analytics.revenueByHour[23]?.saleCount || 0}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Hourly Sales Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Hourly Sales Revenue</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <Line 
                data={analytics.revenueByHour}
                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
              >
                <XAxis dataKey="hour" />
                <YAxis />
                <Line type="monotone" dataKey="revenue" stroke="#8884d8" />
                <Line type="monotone" dataKey="saleCount" stroke="#82ca9d" />
              </Line>
            </div>
          </CardContent>
        </Card>

        {/* Category Distribution */}
        <Card>
          <CardHeader>
            <CardTitle>Sales by Category</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <Pie
                data={Object.entries(analytics.salesByCategory).map(([name, value]) => ({
                  name,
                  value
                }))}
                dataKey="value"
                nameKey="name"
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={80}
                fill="#8884d8"
                label
              />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Top Concerts */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Top Performing Concerts</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[300px]">
            <Bar
              data={analytics.topConcerts}
              margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
            >
              <XAxis dataKey="concertName" />
              <YAxis />
              <Bar dataKey="revenue" fill="#8884d8" />
              <Bar dataKey="ticketsSold" fill="#82ca9d" />
            </Bar>
          </div>
        </CardContent>
      </Card>

      {