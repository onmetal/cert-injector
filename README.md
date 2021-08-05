## Cert-injector

Operator for issuing certificates for TCP applications 
which couldn't be hidden behind ingress and when it's impossible to use DNS-challenge.


### Certificate issuer:

Will create secret with issued certificate.

Annotations for service:
```
apiVersion: v1
kind: Service
metadata:
  name: injector
  annotations:
    "cert.injector.ko/issue": "true"
    "cert.injector.ko/ca-url": "https://acme-v02.api.letsencrypt.org/directory"
    "cert.injector.ko/domains": "example.domain.com"
    "cert.injector.ko/email": "your@email.com"
    "cert.injector.ko/auto-inject": "true"
spec:
  type: LoadBalancer
  ports:
    - name: http
      port: 80
      targetPort: http
  selector:
    app: nginx
```

**"cert.injector.ko/issue"** - Specify service you want to issue certificate.

**"cert.injector.ko/ca-url"** - Certificate authority URL. LE-staging use by default.

**"cert.injector.ko/domains"** - Domain list, e.g. "domain.com,zzz.domain.com,yyy.domain.com".

**"cert.injector.ko/email"** - Email address for Let's Encrypt account.

**"cert.injector.ko/auto-inject"** - Will automatically inject annotations to the deployment.

### Certificate injector:

Will mutate deployment and will add volume and volumemounts.

Certificate path: "/certs/..."

Annotations for deployment:
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  annotations:
    cert.injector.ko/mount: "true"
    cert.injector.ko/secret: "injector-tls"
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx

```

**"cert.injector.ko/mount"** - Specify deployment you want to add certificates.

**"cert.injector.ko/cert-name"** - Specify secret name which contains certificates.

### Install
```
helm install cert-injector ./deploy/helm/injector
```