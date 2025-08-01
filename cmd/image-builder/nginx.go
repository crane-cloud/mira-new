package imagebuilder

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteNginxConfig(appPath string, webServerRoot string) error {
	nginxConfig := `worker_processes auto;
daemon off;
pid /tmp/nginx.pid;

error_log stderr warn;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging format
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /dev/stdout main;

    # Basic settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 10M;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1000;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        application/atom+xml
        application/geo+json
        application/javascript
        application/x-javascript
        application/json
        application/ld+json
        application/manifest+json
        application/rdf+xml
        application/rss+xml
        application/xhtml+xml
        application/xml
        font/eot
        font/otf
        font/ttf
        image/svg+xml
        text/css
        text/javascript
        text/plain
        text/xml;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Server configuration
    server {
        listen 8080;
        server_name _;
        root /workspace/` + webServerRoot + `;
        index index.html index.htm;

        # Security - hide nginx version
        server_tokens off;

        # Handle favicon
        location = /favicon.ico {
            log_not_found off;
            access_log off;
        }

        # Handle robots.txt
        location = /robots.txt {
            log_not_found off;
            access_log off;
        }

        # Static assets with caching
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
            add_header X-Frame-Options "SAMEORIGIN" always;
            add_header X-Content-Type-Options "nosniff" always;
        }

        # Main location block for SPA applications
        location / {
            try_files $uri $uri/ /index.html;
            
            # Add security headers to HTML files
            location ~* \.html$ {
                add_header Cache-Control "no-cache, no-store, must-revalidate";
                add_header Pragma "no-cache";
                add_header Expires "0";
                add_header X-Frame-Options "SAMEORIGIN" always;
                add_header X-Content-Type-Options "nosniff" always;
                add_header X-XSS-Protection "1; mode=block" always;
            }
        }

        # API proxy (if needed) - uncomment and modify as needed
        # location /api/ {
        #     proxy_pass http://backend-service:8080/;
        #     proxy_set_header Host $host;
        #     proxy_set_header X-Real-IP $remote_addr;
        #     proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        #     proxy_set_header X-Forwarded-Proto $scheme;
        # }

        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }

        # Error pages
        error_page 404 /index.html;
        error_page 500 502 503 504 /50x.html;
        location = /50x.html {
            root /usr/share/nginx/html;
        }
    }
}`

	// Create the nginx.conf file path
	nginxConfPath := filepath.Join(appPath, "nginx.conf")

	// Write the nginx configuration to the file
	if err := os.WriteFile(nginxConfPath, []byte(nginxConfig), 0644); err != nil {
		return fmt.Errorf("failed to write nginx.conf: %w", err)
	}

	fmt.Printf("nginx.conf written to: %s\n", nginxConfPath)
	return nil
}
