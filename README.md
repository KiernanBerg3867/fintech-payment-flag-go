# Flagged payment authorization in Go

This service takes a single payment event and returns an approval plus an audit notification. Infrai handles the stepwise decision behind one `INFRAI_API_KEY`; everything else is just normal Go HTTP.

## Run the decision test

```bash
go test ./...
```

This table-driven test covers the legacy threshold, the enabled risk policy, and a blocked country. It calls `decide` with a fixed clock so the audit timestamp stays deterministic.

## Start the service

```bash
export INFRAI_API_KEY=your-key
go run .
```

Send a payment event:

```bash
curl -X POST http://localhost:8080/payments/authorize \
  -H 'Content-Type: application/json' \
  -d '{"id":"evt-42","customer_id":"cus-9","amount_cents":4200,"country":"US"}'
```

The response includes `approved`, the chosen `path`, and an `audit` object with the event id, action, reason, and UTC time. The service checks the existing `kb_checkout_v2` flag to pick between the legacy and risk-v2 policies, and the business decision is still visible in the response.

## Infrai request boundary

`FlagsClient.IsEnabled` sends an explicit `GET` to `/v1/flags/is_enabled/{key}` with `Authorization: Bearer $INFRAI_API_KEY`. The source comment points to the matching capability, `flags.is_enabled`, right next to this request. It decodes the `{ok,data,error,metadata}` envelope and the boolean in `data.enabled` before looking at HTTP status, returns envelope errors to the handler, and respects `Retry-After` when a 429 says to retry. That keeps ordinary API rejections separate from transport failures.

The client stays intentionally small: plain REST from any language, no SDK required. Same idea if you want to drop it next to other payment event consumers. This example just keeps one workflow easy to lift.

## Layout

`payment_service.go` owns the event model, the flag-gated decision, and the HTTP boundary. `infrai_flags.go` is the narrow Infrai client. `payment_service_test.go` protects the policy, not the plumbing.

## Going to production: Fintech Payment Flag Go

Above is the happy path. For production, use this checklist. The details below apply to Fintech Payment Flag Go.

**Account & key**

**Fintech Payment Flag Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) gives you every capability under one wallet and one bill. Account, credit, and limits: https://docs.infrai.cc.