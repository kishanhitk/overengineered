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


Initially the actions script used to build and deploy both api and the astro app wheneever there was a push to main. But, now I have updated it to deploy conditionally. Build and deploy api only if something changed in api directory, similarly for astro app.

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


## Docker - Week 2

Now it is time to introduce docker. Right now our system had many parts, the astro app, the go api, dicedb, ngnix. And during deploy we used to build and place the build outputs in the server machine.
But now, to make things simpler, and more controlled, we will package each of our services into a separate docker container and will create our docker stack in the docker-compose.yml.
And, now instead of gh actions building and putting the build output in the server, it will build a docker image with the build instructions and will push to ghrc (we are not using docker hub because hosting private docker image is not free there ). 
So, this is how our docker images looks like
```
# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```
This is for the go api. Notice here we are using multistages. The main benefit of multi-stage builds is to create a smaller final image. In this case, we use a larger image (golang:1.23-alpine) with all the build tools to compile the Go application, but the final image is based on a minimal alpine image, containing only the necessary runtime dependencies and the compiled binary.

```
# Use an official Bun runtime as a parent image
FROM oven/bun:1

# Set the working directory in the container
WORKDIR /app

# Copy package.json and bun.lockb (if you're using bun.lockb)
COPY website/package.json website/bun.lockb* ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy the rest of the application code
COPY website/ .

# Build the Astro app
RUN bun run build

# Expose the port the app runs on
EXPOSE 4321

# Run the app
CMD ["bun", "run", "start"]
```

And, this is our astro app docker image.
And finally to combine all of these, we have our docker-compose file
```
services:
  api:
    image: ghcr.io/kishanhitk/overengineered-api:latest
    ports:
      - "8080:8080"
    depends_on:
      - dicedb
    environment:
      - REDIS_URL=dicedb:7379
    restart: always

  astro-app:
    image: ghcr.io/kishanhitk/overengineered-astro:latest
    ports:
      - "4321:4321"
    depends_on:
      - api
    environment:
      - PUBLIC_API_BASE_URL=https://api-overengineered.kishans.in
      - API_BASE_URL_INTERNAL=http://api:8080
    restart: always

  dicedb:
    image: dicedb/dicedb
    ports:
      - "7379:7379"
    restart: always
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - api
      - astro-app
    restart: always
```

Here, we specify all the container we need, along with the images. We have added restart:always so that the container restarts automatically when it crashed or when the host server restarts without explicitly creating a systemd. We have used internal networking for the containers to communicate with each other. We are also paasing some env varibale to astro container. Notice we have two different API base url. This is because if some API call is made on the astro server, then it will go directly to the go api container, but id an API call is made from the astro client on user's browser, it will go the deployed version of the API.


To deploy we will first install docker and docker compose on our server. And when we run docker compose up in the directory where we added our docker-compose file, docker will pull all the images and create containers based on the config we provided. 
We will also add a ngnix file in the same folder that will be copied to the ngnix container.
```
user nginx;
worker_processes auto;
worker_rlimit_nofile 65535;

events {
    worker_connections 1024;
    multi_accept on;
    use epoll;
}

http {
    charset utf-8;
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    server_tokens off;
    log_not_found off;
    types_hash_max_size 2048;
    client_max_body_size 16M;

    # MIME
    include mime.types;
    default_type application/octet-stream;

    # Logging
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    # SSL
    # ssl_protocols TLSv1.2 TLSv1.3;
    # ssl_prefer_server_ciphers on;
    # ssl_session_cache shared:SSL:10m;
    # ssl_session_timeout 10m;

    # Compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml application/json application/javascript application/rss+xml application/atom+xml image/svg+xml;

    # Cache settings
    proxy_cache_path /tmp/nginx_cache levels=1:2 keys_zone=my_cache:10m max_size=10g inactive=60m use_temp_path=off;

    # API Server
    server {
        listen 80;
        server_name api-overengineered.kishans.in;

        location / {
            proxy_pass http://api:8080;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache my_cache;
            proxy_cache_use_stale error timeout http_500 http_502 http_503 http_504;
            proxy_cache_lock on;
            add_header X-Cache-Status $upstream_cache_status;
        }

        # Security headers
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header Referrer-Policy "no-referrer-when-downgrade" always;
        add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    }

    # Astro App Server
    server {
        listen 80;
        server_name overengineered.kishans.in;

        location / {
            proxy_pass http://astro-app:4321;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Static file caching
        location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
            expires 30d;
            add_header Cache-Control "public, no-transform";
        }

        # Security headers
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header Referrer-Policy "no-referrer-when-downgrade" always;
        add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    }
}
```

This time we have added a bunch of configurations to the ngnix config, like gzip enable, caching, etc.

And this is our gh action script
```
name: Build, Push, and Deploy

on:
  push:
    branches: [main, docker]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push-images:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Log in to the Container registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push API image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: api/Dockerfile
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-api:latest

      - name: Build and push Astro image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: website/Dockerfile
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-astro:latest

  deploy:
    needs: build-and-push-images
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Copy docker-compose file to server
        uses: appleboy/scp-action@master
        with:
          host: ${{ secrets.HOST }}
          username: ${{ secrets.USERNAME }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          port: ${{ secrets.PORT }}
          source: "docker-compose.yml"
          target: "/var/www/overengineered"

      - name: Deploy to server
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.HOST }}
          username: ${{ secrets.USERNAME }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          port: ${{ secrets.PORT }}
          script: |
            cd /var/www/overengineered
            docker compose down
            docker compose pull
            docker compose up -d
            docker image prune -f

```

We also removed cloudflare tunnel and added our domain directly to the vps from digital ocean dashboard.
Now, the requests go to DNS which resovled to our VPS IP adderss and on our VPS ngnix handles the requests and sends it to the correct ports running on the containers.
This works perfecly fine.
Althoug we did some load testing and saw the performance has decreased, compared to non-docker version. Maybe this is becuase of some overhead that docker brings. We will see how can we minimise that.

Now, this looks good. But, what if we receive a lot of traffic and our single VPS is unable to handle the load and we need to scale horizontally. Maybe its time to think about load balancing or container management with Kubernetes...

