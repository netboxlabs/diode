package sentry

import (
	"github.com/getsentry/sentry-go"
)

// CaptureError captures an error and sends it to sentry with tags and context
func CaptureError(err error, tags map[string]string, contextKey string, context map[string]any) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		for k, v := range tags {
			scope.SetTag(k, v)
		}

		if contextKey != "" && context != nil {
			scope.SetContext(contextKey, context)
		}

		sentry.CaptureException(err)
	})
}
