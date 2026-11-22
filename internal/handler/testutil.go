package handler

import (
	"github.com/KiberIGOR/tinyurl/internal/audit"
	"github.com/KiberIGOR/tinyurl/internal/facade"
	"github.com/KiberIGOR/tinyurl/internal/service"
)

// NewForTest собирает Handler с фасадом для unit-тестов.
func NewForTest(svc *service.Shortener, pinger Pinger, auditBroker *audit.Broker) *Handler {
	fac := facade.NewShortener(svc, auditBroker)
	return New(fac, svc, pinger, auditBroker)
}