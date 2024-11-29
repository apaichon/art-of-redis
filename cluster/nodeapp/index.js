const express = require('express');
const { createCluster } = require('redis');

const app = express();
const port = 3000;

// Enable JSON body parsing
app.use(express.json());

// Redis cluster configuration
const redisHosts = [
  { host: 'redis1', port: 7001 },
  { host: 'redis2', port: 7002 },
  { host: 'redis3', port: 7003 },
];

// Create a Redis cluster client
const redisClient =  createCluster({
  rootNodes: [
    { url: 'redis://redis1:7001' },
    { url: 'redis://redis2:7002' },
    { url: 'redis://redis3:7003' }
  ]
});

// Connect to the Redis cluster
redisClient.on('ready', () => {
  console.log('Connected to Redis cluster');
});

redisClient.on('error', (err) => {
  console.error('Error connecting to Redis cluster:', err);
});

// Express API endpoints
app.get('/api/tickets', (req, res) => {
  // Retrieve tickets from Redis cluster
  redisClient.get('tickets', (err, tickets) => {
    if (err) {
      res.status(500).send({ message: 'Error retrieving tickets' });
    } else {
      res.send(tickets);
    }
  });
});

app.post('/api/tickets', (req, res) => {
  // Create a new ticket in Redis cluster
  const ticket = req.body;
  console.log('ticket', ticket);
  redisClient.set(`ticket:${ticket.id}`, JSON.stringify(ticket), (err) => {
    if (err) {
      res.status(500).send({ message: 'Error creating ticket' });
    } else {
      res.send({ message: 'Ticket created successfully' });
    }
  });
});

// Start the Express server
app.listen(port, () => {
  console.log(`Server started on port ${port}`);
});

// Handle uncaught exceptions
process.on('uncaughtException', (err) => {
  console.error('Uncaught Exception:', err);
  process.exit(1);
});


