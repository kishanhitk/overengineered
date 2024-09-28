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
Next, we will setup cloudflare tunnel.
Why CF Tunnel? Why not simply map the domain to the vps public IP?
Because the vps device does not have a reserved public IP adrress and it will keep changing. Getting a reserved/fixed IP address is not free, on almost any cloud provider.
So, we use CF tunnel. It runs on our vps and then we generate some config and add those config details in our CF DNS management and now our server without any reserved IP is accessible via our domain.
Install cf tunnel
```
sudo mkdir -p --mode=0755 /usr/share/keyrings
curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared jammy main' | sudo tee /etc/apt/sources.list.d/cloudflared.list
sudo apt update && sudo apt install cloudflared



cloudflared tunnel login

cloudflared tunnel create api-overengineered
```
create a configuration file in /etc/cloudflared/config.yml. This will tell that tunnel is running on 8082 and if request comes for API, send it to port 8081 where our ngnix server is running.
```
url: http://localhost:8082
tunnel: <tunnel-id>
credentials-file: /root/.cloudflared/<your-tunnel-id>.json

ingress:
  - hostname: api.overengineered.kishans.in
    service: http://localhost:8081
  - service: http_status:404
```

Add this CNAME DNS record to our subdomain `<tunnel-id>.cfargotunnel.com`
And now our go API is accessible to the world on  `api.overengineered.kishans.in`
Now, to make cloudflare tunnel service run automatically and keep it running, we will create a systemd service.
this config inside `/etc/systemd/system/cloudflared.service`
```
[Unit]
Description=cloudflared
After=network.target

[Service]
TimeoutStartSec=0
Type=notify
ExecStart=/usr/bin/cloudflared --config /etc/cloudflared/config.yml tunnel run
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

And, then start and enable the cloudflared service:
```
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
```

Now, the basic setup is done. We have a go API running on our server, and is served via ngnix. And we have domain conected to this ngnix port with the cloudlare tunnel, and the API is publically accessible. And, all of the services are set as systemd so that it keeps running and restarts when system restarts.

Now, we don't want to pull my code manually on my server whenever I push some code updates to my API. So, we will setup CI/CD.
Since, we are using GItHub, we can use Gh actions.
So, one every push to the main branch, the gh action will build our go project and then copy the build output file on our server by connecting via ssh. we will create a separate ssh key for gh actions. After copying files, it will restarts all the systemd services so that all the services are updated based on the new code. 

Now, our whole setup for api is ready in the mvp state. Everything works as expected. 
- We code and push to github,
-  gh actions runs on push and builds and copies the new version of build output of the api on the server
-  the server serves new api
Next, thing is to setup our astro app. This will be simple since we have already setup most of the things.
First, we will add the astro app to our github repo. Then we will update our gh action to build and copy our astro app as well.
We want to serve the astro app using bun because it is more performant than Node.js. So, we will have to install bun on our server. 
Now, just like we added systmed, ngnix, and cloudflare tunnel config for our API, we will do this for astro app as well.
astro systemd at `/etc/systemd/system/astro-app.service`
```
[Unit]
Description=Astro App
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/overengineered/website
ExecStart=/root/.bun/bin/bun run /var/www/overengineered/website/dist/server/entry.mjs
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

The astro app will be accessible on `https://api-overengineered.kishans.in`
The ngnix config file at `/etc/nginx/sites-available/overengineered.kishans.in.conf`
```
server {
    listen 127.0.0.1:8082;
    server_name overengineered.kishans.in;

    location / {
        proxy_pass http://localhost:4321;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Add this ngnix config to sites enable va ngnix symlink
`sudo ln -s /etc/nginx/sites-available/overengineered.kishans.in.conf /etc/nginx/sites-enabled/`

And, the updated cloudflare tunnel config at `/etc/cloudflared/config.yml`
```
url: http://localhost:8081
tunnel: 05362019-2c26-446e-ab4d-8b0f491f966c
credentials-file: /root/.cloudflared/05362019-2c26-446e-ab4d-8b0f491f966c.json

ingress:
  - hostname: api-overengineered.kishans.in
    service: http://localhost:8081
  - hostname: overengineered.kishans.in
    service: http://localhost:8082
  - service: http_status:404
```
We will also update the DNS records for this new domain.

Now, we will restart all the processes, including ngnix, cf tunnel, and api and astro services. And voila our whole project, including frontend and backedn is now deployed is now deployed on a vps along with ci/cd.

### Add a SQLite DB
This won't require any change in the infra. We just update our API code to use SQLite and creation of DB will be automatically handled by the go app.
### Adding redis (dicedb)
I am using dicedb instead of redis. It's a redis alternative which is 100% compatible with redis API and claims to be faster than redis.
I am running dicedb  as a docker container, and for that I need to install dokcer on the vps.
once docker is installed, i can run the dicedb image using this command `docker run -p 7379:7379 dicedb/dicedb`
We will update the application code to use redis DB.

Also, our API has evolved now. It stores requests to greetings in DB and it also has a API endpoint to return count.
We use redis to cache this count.

But we have a problem now. The redis DB docker service is not set as systemd and will stop if server restarts.
