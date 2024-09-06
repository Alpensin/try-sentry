package zlog

import (
	"time"

	"github.com/rs/zerolog"

	sentry "github.com/getsentry/sentry-go"
)

// The identifier of the zerolog SDK.
const sdkIdentifier = "sentry.go.platform.zlog"
const name = "platform.zlog"

// These default log field keys are used to pass specific metadata in a way that
// Sentry understands. If they are found in the log fields, and the value is of
// the expected datatype, it will be converted from a generic field, into Sentry
// metadata.
//
// These keys may be overridden by calling SetKey on the hook object.
const (
	// FieldRequest holds an *http.Request.
	FieldRequest = "request"
	// FieldUser holds a User or *User value.
	FieldUser = "user"
	// FieldTransaction holds a transaction ID as a string.
	FieldTransaction = "transaction"
	// FieldFingerprint holds a string slice ([]string), used to dictate the
	// grouping of this event.
	FieldFingerprint = "fingerprint"

	// These fields are simply omitted, as they are duplicated by the Sentry SDK.
	FieldGoVersion = "go_version"
	FieldMaxProcs  = "go_maxprocs"
)

// Hook is the zerolog hook for Sentry.
//
// It is not safe to configure the hook while logging is happening. Please
// perform all configuration before using it.
type Hook struct {
	hub    *sentry.Hub
	keys   map[string]string
	levels []zerolog.Level
}

// New initializes a new zerolog hook which sends logs to a new Sentry client
// configured according to opts.
func NewHook(levels []zerolog.Level, opts sentry.ClientOptions) (*Hook, error) {
	client, err := sentry.NewClient(opts)
	if err != nil {
		return nil, err
	}

	client.SetSDKIdentifier(sdkIdentifier)

	return NewHookFromClient(levels, client), nil
}

// NewFromClient initializes a new zerolog hook which sends logs to the provided
// sentry client.
func NewHookFromClient(levels []zerolog.Level, client *sentry.Client) *Hook {
	h := &Hook{
		levels: levels,
		hub:    sentry.NewHub(client, sentry.NewScope()),
		keys:   make(map[string]string),
	}
	return h
}

// AddTags adds tags to the hook's scope.
func (h *Hook) AddTags(tags map[string]string) {
	h.hub.Scope().SetTags(tags)
}

// SetKey sets an alternate field key. Use this if the default values conflict
// with other loggers, for instance. You may pass "" for new, to unset an
// existing alternate.
func (h *Hook) SetKey(oldKey, newKey string) {
	if oldKey == "" {
		return
	}
	if newKey == "" {
		delete(h.keys, oldKey)
		return
	}
	delete(h.keys, newKey)
	h.keys[oldKey] = newKey
}

func (h *Hook) key(key string) string {
	if val := h.keys[key]; val != "" {
		return val
	}
	return key
}

// Run sends entry to Sentry.
func (h *Hook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	event := h.entryToEvent(e, level, message)
	h.hub.CaptureEvent(event)
}

var levelMap = map[zerolog.Level]sentry.Level{
	zerolog.DebugLevel: sentry.LevelDebug,
	zerolog.InfoLevel:  sentry.LevelInfo,
	zerolog.WarnLevel:  sentry.LevelWarning,
	zerolog.ErrorLevel: sentry.LevelError,
	zerolog.FatalLevel: sentry.LevelFatal,
	zerolog.PanicLevel: sentry.LevelFatal,
}

func (h *Hook) entryToEvent(e *zerolog.Event, level zerolog.Level, message string) *sentry.Event {
	// Initialize the Sentry s

	s := &sentry.Event{
		// Breadcrumbs: []*sentry.Breadcrumb{},
		Dist:        "ThatsDist",
		Environment: "ThatsEnvironment",
		// EventID: "12312312",
		// Fingerprint: []string{"Finger", "Print"},
		Level:      levelMap[level],
		Message:    message,
		Release:    "SomeRelease",
		ServerName: "ServerName",
		Logger:     name,
	}
	return s
}

// Flush waits until the underlying Sentry transport sends any buffered events,
// blocking for at most the given timeout. It returns false if the timeout was
// reached, in which case some events may not have been sent.
func (h *Hook) Flush(timeout time.Duration) bool {
	return h.hub.Client().Flush(timeout)
}
