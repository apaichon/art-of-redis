import React from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';

export const TopConcerts = ({ concerts }) => (
  <Card className="mb-6">
    <CardHeader>
      <CardTitle>Top Performing Concerts</CardTitle>
    </CardHeader>
    <CardContent>
      <div className="h-[300px]">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={concerts} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
            <XAxis dataKey="concertName" />
            <YAxis />
            <Tooltip />
            <Bar dataKey="revenue" fill="#8884d8" name="Revenue" />
            <Bar dataKey="ticketsSold" fill="#82ca9d" name="Tickets Sold" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </CardContent>
  </Card>
);