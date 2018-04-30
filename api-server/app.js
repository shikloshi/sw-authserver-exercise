const http = require('http');

const port = process.env.PORT || '8080';

http.createServer((req, res) => {
    console.log("Got new request")
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.write(JSON.stringify({ method: req.method }));
    res.end();
}).listen(port);
