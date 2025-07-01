# replacePathRegex  
Traefik Middleware Plugin – Rewrite Request Path with Regular Expressions
Target is to solve the original replace Path Regex plugin, Can't use the full URL as the Regex, this plugin can use the full URL as the Regex to make more flexible.

## Overview
`replaceRathRegex` is a lightweight middleware plugin for **Traefik v3** that rewrites an incoming request's URL path using a configurable regular-expression pattern and replacement string (back-references `$1`, `$2`, … supported).  
Typical use cases include:

* Mapping per-domain `robots.txt` / `sitemap.xml` requests to a shared S3 bucket  
* Refactoring legacy routes, e.g. `/old/path/...` ➜ `/new/path/...`

The original path is preserved in the `X-Replaced-Path` header for debugging purposes.

## Features
* **Flexible regex & replacement** – any Go `regexp` syntax, back-references `$n` expanded automatically  
* **No external dependencies** – built with Go standard library only  
* **Easy debugging** – original path stored in a custom response/request header

---

## Installation

### 1 – Enable the plugin in `traefik.yml`
```yaml
experimental:
  plugins:
    replacepathregex:
      modName: github.com/Noahnut/replacePathRegex
      version: v0.0.0          # replace with desired tag
```

### 2 – Declare middlewares (dynamic configuration)
```yaml
http:
  middlewares:
    robots-rewrite:
      plugin:
        replacepathregex:
          regex: "^https://([^/]+)/robots\\.txt$"
          replace: "/robots_txt/$1/robots.txt"

    sitemap-rewrite:
      plugin:
        replacepathregex:
          regex: "^https://([^/]+)/([^/]+\\.xml)$"
          replace: "/sitemap/$1/$2"
```

### 3 – Attach middleware to routers
```yaml
http:
  routers:
    robots-router:
      rule: "HostRegexp(`{host:.+}`) && Path(`/robots.txt`)"
      middlewares:
        - robots-rewrite
      service: s3-backend
      tls: {}

    sitemap-router:
      rule: "HostRegexp(`{host:.+}`) && PathRegexp(`/.*\\.xml`)"
      middlewares:
        - sitemap-rewrite
      service: s3-backend
      tls: {}
```

---

## Examples

| Incoming URL                                   | Regex / Replace                                  | Resulting Path                                   |
|------------------------------------------------|--------------------------------------------------|--------------------------------------------------|
| `https://xxxx.test.com/robots.txt`          | `^https://([^/]+)/robots\.txt$` / `/robots_txt/$1/robots.txt` | `/robots_txt/xxxx.test.com/robots.txt` |
| `https://xxxx.test.com/xx.xml`              | `^https://([^/]+)/([^/]+\.xml)$` / `/sitemap/$1/$2`           | `/sitemap/xxxx.test.com/xx.xml`       |
| `https://xxxx.test.com/robots.txt`         | same as robots rule                              | `/robots_txt/xxxx.test.com/robots.txt`|
| `https://xxxx.test.com/xx.xml`             | same as sitemap rule                             | `/sitemap/xxxx.test.com/xx.xml`      |

---

## Configuration Parameters

| Field      | Type   | Description                                                  |
|------------|--------|--------------------------------------------------------------|
| `regex`    | string | Regular expression applied to `req.URL.String()`             |
| `replace`  | string | Replacement template; supports `$1 … $n` back-references     |

Notes:

* If a back-reference has no match, it is replaced by an empty string.  
* After substitution, any leftover `$n` tokens are stripped automatically.  
* The original untouched path is stored in `X-Replaced-Path` header.

---

## Development & Tests

```bash
git clone https://github.com/Noahnut/replace-path-regex.git
cd replace-path-regex
go mod tidy        # download dependencies
go test ./...      # run unit tests
```

The test suite covers various rewrite scenarios and can be extended as needed.

---

## License
MIT © 2024 Noahnut
