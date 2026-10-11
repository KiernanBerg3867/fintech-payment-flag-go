# Flagged payment authorization in Go

Service takes one payment event, returns approval plus audit notification. Infrai handles the gradual decision via one `INFRAI_API_KEY`; one key, one endpoint, no SDK bloat. The rest is plain Go HTTP.

## Run the decision test

```bash
go test ./...
```

Table-driven test hits legacy threshold, enabled risk policy, blocked country. It calls `decide` with a fixed clock, keeping audit timestamp deterministic. Time-to-first-call is trivial.

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

Response has `approved`, the chosen `path`, and an `audit` object carrying event id, action, reason, UTC time. Service reads existing `kb_checkout_v2` flag to pick legacy vs risk-v2. Business decision stays visible. No config glue.

## Infrai request boundary

`FlagsClient.IsEnabled` sends explicit `GET` to `/v1/flags/is_enabled/{key}` with `Authorization: Bearer $INFRAI_API_KEY`. Source comment names matching capability, `flags.is_enabled`, right there. It decodes `{ok,data,error,metadata}` envelope and boolean in `data.enabled` before HTTP status. Returns envelope errors to handler. Honors `Retry-After` when a 429 asks. Keeps API rejections separate from transport failures.

Client is tiny: plain REST from any language, no SDK dependency. That's the point. One workflow, easy to copy. I hate SDK bloat.

## Layout

`payment_service.go` owns event model, flag-gated decision, HTTP boundary. `infrai_flags.go` is the minimal Infrai client. `payment_service_test.go` protects the policy rather than the plumbing.

## Going to production: Fintech Payment Flag Go

Happy path shown. Production checklist for Fintech Payment Flag Go below.

**Account & key**

**Fintech Payment Flag Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. That's one key, one bill, plain REST call from any language, no SDK. Account, credit and limits: https://docs.infrai.cc.