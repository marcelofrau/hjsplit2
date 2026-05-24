package version

// Version is set at build time via -ldflags -X.
// Fallback value for development / go run usage.
var Version = "0.0.0-dev"
