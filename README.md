## Cert-injector

Cert-injector is an operator for issuing certificates for services which couldn't be pointed behind 
ingress and when impossible to use DNS-challenge.


### Certificate issuer:

Will create secret with issued certificate.

Annotations for service:
```
 "cert.injector.ko/email": < your email address for LE>
 "cert.injector.ko/ca-url": < certificate authority url > // LE-staging by default
 "cert.injector.ko/domains": < slice of domains > // Exampl: domain.com,zzz.domain.com,yyy.domain.com
 "cert.injector.ko/inject": <true / false>
 "cert.injector.ko/auto-inject": < true / false >
```

### Certificate injector:

Will mutate deployment and will add volume and volume mounts.

Certificate path: "/certs/..."

Annotations for deployment:
```
 "cert.injector.ko/mount": < true / false > // Specify deployment which you want to mutate
 "cert.injector.ko/cert-name": < secret name with certs >
```
