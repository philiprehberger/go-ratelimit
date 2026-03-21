# Changelog

## 0.2.1

- Consolidate README badges onto single line

## 0.2.0

- Add `Limiter.SetRate` for runtime rate and burst reconfiguration
- Add `LimiterStats` and `Limiter.Stats` for allowed/rejected request counters
- Add `KeyedLimiter.OnReject` callback for rejection notifications

## 0.1.1

- Add badges and Development section to README

## 0.1.0

- Initial release
- Token bucket `Limiter` with `Allow` and `Wait`
- Per-key `KeyedLimiter` for multi-tenant rate limiting
- HTTP middleware with 429 responses
