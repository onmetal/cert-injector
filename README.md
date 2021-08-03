## Cert-injector

Operator for issuing TCP applications certificates 
which couldn't be hidden behind ingress and when it's impossible to use DNS-challenge.


### Certificate issuer:

Will create secret with issued certificate.

Annotations for service:
```
 "cert.injector.ko/email": < your email address for LE >
 "cert.injector.ko/ca-url": < certificate authority URL > // LE-staging by default
 "cert.injector.ko/domains": < slice of domains > // Example: domain.com,zzz.domain.com,yyy.domain.com
 "cert.injector.ko/inject": <true / false>
 "cert.injector.ko/auto-inject": < true / false >
```

### Certificate injector:

Will mutate deployment and will add volume and volumemounts.

Certificate path: "/certs/..."

Annotations for deployment:
```
 "cert.injector.ko/mount": < true / false >
 "cert.injector.ko/cert-name": < secret name >
```

### Install
```
helm install cert-injector ./deploy/helm/injector
```
