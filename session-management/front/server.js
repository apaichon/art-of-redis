const express = require('express');
const bodyParser = require('body-parser');
const axios = require('axios'); 

const app = express();
const PORT = 3000;

app.use(bodyParser.json());
const path = require('path'); // Add path module

app.use(express.static(path.join(__dirname, 'public'))); // Serve static files from the 'public' directory

app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, 'public', 'index.html')); // Serve index.html on root
});

app.post('/api/login', async (req, res) => {
    console.log("Login clicked", req.body);
    const response = await axios.post('http://127.0.0.1:9001/api/login', req.body, {
        headers: { 'Content-Type': 'application/json' }
    });
    res.json(response.data); // Update to use axios response
});

app.post('/api/logout', async (req, res) => {
    const response = await axios.post('http://127.0.0.1:9001/api/logout', req.body, {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': req.headers.authorization
        }
    });
    res.json(response.data); // Update to use axios response
});

app.post('/api/check-auth', async (req, res) => {
    const response = await axios.post('http://localhost:9001/api/check-auth', req.body, {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': req.headers.authorization
        }
    });
    res.json(response.data); // Update to use axios response
});

app.listen(PORT, () => {
    console.log(`Server is running on http://localhost:${PORT}`);
});

// Handle uncaught exceptions
process.on('uncaughtException', (error) => {
    console.error('Uncaught Exception:', error);
    // Optionally, you can implement a graceful shutdown or alerting here
});