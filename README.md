# Overengineered App

### The first version
tech stack: 
ngnix: reverse proxy
go: api
astro: full-stack frontend
cloudflare tunnel: site accessible to domain

There is a go app inside /api and an astro app inside /website.
There is VPS taken from digital ocean.
First, I created a very minimal hello-world api and pushed it to a private git repo.
Then I sshed into the VPS and cloned the git repo there. Downloaded go on the vps. Built and run the app, that gave me the `main` executable file. It seemed to work fine.
Now, I created a systemd configuration for the api service so that the go server can run and restart when the vps restarts. So, that I don't have to manually start the server. It is configured to start the server only after the network is available on the machine.
```
cat /etc/systemd/system/go-api.service
[Unit]
Description=Go API
After=network.target

[Service]
User=root
WorkingDirectory=/var/www/overengineered/api
ExecStart=/var/www/overengineered/api/main
Restart=always
RestartSec=5
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

Then, we run these two commands. The first command starts the `go-api` service right now. The second command sets up the `go-api` service to start automatically whenever the system boots.
```
sudo systemctl start go-api
sudo systemctl enable go-api
```
Now, our go app is running and will be running as long as system is live and will restart when the system restarts.
Now, we will setup ngnix. The API will be available on `api-overengineered.kishans.in`
First, we will instlal ngnix.
And then create a file `/etc/nginx/sites-available/api-overengineered.kishans.in.conf`
```
cat /etc/nginx/sites-available/api-overengineered.kishans.in.conf
server {****
    listen 127.0.0.1:8081;
    server_name api-overengineered.kishans.in;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```
This configuration sets up Nginx to act as a reverse proxy, forwarding requests received on 127.0.0.1:8081 to a service running on localhost:8080.
Next, we will make this configuration active and restarts ngnix.
```
sudo ln -s /etc/nginx/sites-available/api-overengineered.kishans.in.conf /etc/nginx/sites-enabled/
sudo systemctl reload nginx
```
Now, our ngnix is ready.