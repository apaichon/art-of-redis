// ตัวอย่างการเชื่อมต่อ Redis Cluster ด้วย Node.js
const Redis = require('ioredis');
const express = require('express');
const app = express();
app.use(express.json()); // เพื่อให้สามารถรับ JSON ได้

const cluster = new Redis.Cluster([
  {
    host: 'redis1',
    port: 7001,
  },
  {
    host: 'redis2',
    port: 7002,
  },
  {
    host: 'redis3',
    port: 7003,
  },
]);

// API สำหรับสร้างตั๋ว
app.post('/api/tickets', (req, res) => {
  const { ticketId, ticketData } = req.body;
  createTicket(ticketId, ticketData);
  res.status(201).send('Ticket created');
});

// API สำหรับอ่านตั๋ว
app.get('/api/tickets/:ticketId', (req, res) => {
  const ticketId = req.params.ticketId;
  readTicket(ticketId, (err, ticket) => {
    if (err) {
      return res.status(500).send('Error reading ticket');
    }
    res.status(200).json(ticket);
  });
});

// API สำหรับอัปเดตตั๋ว
app.put('/api/tickets/:ticketId', (req, res) => {
  const ticketId = req.params.ticketId;
  const ticketData = req.body;
  updateTicket(ticketId, ticketData);
  res.status(200).send('Ticket updated');
});

// API สำหรับลบตั๋ว
app.delete('/api/tickets/:ticketId', (req, res) => {
  const ticketId = req.params.ticketId;
  deleteTicket(ticketId);
  res.status(200).send('Ticket deleted');
});

// เริ่มเซิร์ฟเวอร์
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
});

// ทดสอบการเชื่อมต่อ
/* cluster.set('key', 'value', (err, result) => {
  if (err) {
    console.error('Error setting key:', err);
  } else {
    console.log('Key set result:', result);
  }
});

// ดึงค่าจาก Redis
cluster.get('key', (err, result) => {
  if (err) {
    console.error('Error getting key:', err);
  } else {
    console.log('Value:', result);
  }
});
*/

// ฟังก์ชันสำหรับสร้างตั๋ว
function createTicket(ticketId, ticketData) {
  cluster.set(`ticket:${ticketId}`, JSON.stringify(ticketData), (err) => {
    if (err) {
      console.error('Error creating ticket:', err);
    }
  });
}

// ฟังก์ชันสำหรับอ่านตั๋ว
function readTicket(ticketId, callback) {
  cluster.get(`ticket:${ticketId}`, (err, result) => {
    if (err) {
      return callback(err);
    }
    callback(null, JSON.parse(result));
  });
}

// ฟังก์ชันสำหรับอัปเดตตั๋ว
function updateTicket(ticketId, ticketData) {
  cluster.set(`ticket:${ticketId}`, JSON.stringify(ticketData), (err) => {
    if (err) {
      console.error('Error updating ticket:', err);
    }
  });
}

// ฟังก์ชันสำหรับลบตั๋ว
function deleteTicket(ticketId) {
  console.log(`Deleting ticket: ${ticketId}`);
  cluster.del(`ticket:${ticketId}`, (err) => {
    if (err) {
      console.error('Error deleting ticket:', err);
    }
  });
}

