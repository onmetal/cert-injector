package api

const (
	EmailAnnotationKey      = "cert.injector.ko/email"
	CaURLAnnotationKey      = "cert.injector.ko/ca-url"
	DomainsAnnotationKey    = "cert.injector.ko/domains"
	InjectAnnotationKey     = "cert.injector.ko/inject"
	AutoInjectAnnotationKey = "cert.injector.ko/auto-inject"
)

const (
	AdmissionWebhookAnnotationInjectKey = "cert.injector.ko/mount"
	AdmissionWebhookAnnotationCertKey   = "cert.injector.ko/cert-name"
)

const AnnotationKeyEnabled = "true"
