# Domain Robots Traefik Plugin

A Traefik plugin that dynamically routes `robots.txt` and `sitemap.xml` requests to S3  based on the requesting domain.

## Overview

This plugin intercepts requests for `robots.txt` and `sitemap.xml` files and redirects them to the appropriate S3 bucket  distribution. The target URL is constructed using the requesting domain name, allowing you to serve different robots.txt and sitemap files for different domains from a single S3 bucket.

## Features

- **Dynamic Domain Routing**: Automatically routes requests based on the requesting domain
- **S3 Integration**: Supports direct S3 bucket access with configurable regions
- **Custom Paths**: Configurable robots.txt and sitemap.xml paths
- **Prefix Support**: Optional S3 prefix path for better organization
- **Port Handling**: Automatically strips port numbers from domain names

## Installation

### Prerequisites

- Traefik v2.10 or later
- Go 1.23.2 or later (for building from source)

### Building from Source

```bash
git clone https://github.com/Noahnut/domain-robots.git
cd domain-robots
go build -buildmode=plugin -o domain-robots.so domain_robots.go
```

### Docker Build

```bash
docker build -t domain-robots-plugin .
```

## Configuration

### Plugin Configuration

```yaml
# traefik.yml
experimental:
  plugins:
    domainrobots:
      modName: github.com/Noahnut/domain-robots
      version: v0.0.0
```

### Middleware Configuration

```yaml
# dynamic/domain-robots.yml
http:
  middlewares:
    domain-robots-s3:
      plugin:
        domainrobots:
          # Required: S3 bucket name
          s3Bucket: "your-static-bucket"
          
          # Required: AWS region
          s3Region: "us-east-1"
          
          # Optional: Custom S3 endpoint (for MinIO, etc.)
          s3Endpoint: ""
          
          # Optional: S3 prefix path for organization
          s3PrefixPath: "assets"
          
          # Optional: Custom robots.txt path, robotsTxtPath and sitemapPath pick one 
          robotsTxtPath: "/robots.txt"
          
          # Optional: Custom sitemap.xml path
          sitemapPath: "/sitemap.xml"
          
          # Optional: Protocol (default: https)
          protocol: "https"
```

### Router Configuration

```yaml
http:
  routers:
    # Robots.txt routing
    robots-router:
      rule: "HostRegexp(`{domain:.+}`) && PathPrefix(`/robots.txt`)"
      middlewares:
        - domain-robots-s3
      service: backend-service
      tls: {}
    
    # Sitemap.xml routing
    sitemap-router:
      rule: "HostRegexp(`{domain:.+}`) && PathPrefix(`/sitemap.xml`)"
      middlewares:
        - domain-robots-s3
      service: backend-service
      tls: {}
    
    # Specific domain routing
    example-com-robots:
      rule: "Host(`example.com`) && PathPrefix(`/robots.txt`)"
      middlewares:
        - domain-robots-s3
      service: backend-service
      tls: {}
```
