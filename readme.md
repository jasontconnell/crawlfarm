**Usage**

The project contains two executables.

Update server/config.json to include your listen address and the root of the website that you would like to test. If your site is behind basic authentication you can add that to the headers object.

~~~~ 
{
    "listen": "127.0.0.1:12345",
    "root": "http://www.example.com",
    "headers": { "Authorization": "Basic xxxx" },
    "virtualPaths": []
}
~~~~

virtualPaths is used for additional links that may not be accessible via all other links that are accessible via the root url. They are relative urls.

**Server**

~~~~
cd crawlfarm/server
go build
server.exe   // or ./server
~~~~

**Worker**

~~~~
cd crawlfarm/worker
go build
worker.exe   // or ./worker
~~~~

Copy worker to any network accessible computers. Start up the server first, then add workers as you see fit. They can be added as long as there are URLs in the process queue on the server.


**Known Bugs**

When a worker disconnects, the server doesn't recognize it. I will fix soon.

Code suggestions WELCOME!

Enjoy :)